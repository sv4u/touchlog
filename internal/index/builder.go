package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/note"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// Builder handles full scan indexing of a vault
type Builder struct {
	vaultRoot string
	cfg       *config.Config
}

// NewBuilder creates a new index builder
func NewBuilder(vaultRoot string, cfg *config.Config) *Builder {
	return &Builder{
		vaultRoot: vaultRoot,
		cfg:       cfg,
	}
}

// Rebuild performs a full atomic rebuild of the index
// It creates a temporary database, populates it, then atomically replaces the existing index
func (b *Builder) Rebuild() error {
	// Create temporary database path
	touchlogDir := filepath.Join(b.vaultRoot, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		return fmt.Errorf("creating .touchlog directory: %w", err)
	}

	tmpDBPath := filepath.Join(touchlogDir, "index.db.tmp")
	finalDBPath := filepath.Join(touchlogDir, "index.db")

	// Remove any existing temp database
	if _, err := os.Stat(tmpDBPath); err == nil {
		if err := os.Remove(tmpDBPath); err != nil {
			return fmt.Errorf("removing existing temp database: %w", err)
		}
	}

	// Create and populate temporary database
	// OpenOrCreateDB expects a vault root, but we need to create a temp DB
	// So we'll use sql.Open directly for the temp database
	db, err := sql.Open("sqlite3", tmpDBPath+"?_foreign_keys=1")
	if err != nil {
		return fmt.Errorf("opening temp database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging temp database: %w", err)
	}

	// Apply migrations
	if err := store.ApplyMigrations(db); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	// Perform two-pass indexing
	// Progress messages are handled by indexAll if a progress callback is provided
	if err := b.indexAll(db); err != nil {
		return fmt.Errorf("indexing notes: %w", err)
	}

	// Sync and close database
	if err := db.Close(); err != nil {
		return fmt.Errorf("closing temp database: %w", err)
	}

	// Atomic rename: temp -> final
	if err := os.Rename(tmpDBPath, finalDBPath); err != nil {
		return fmt.Errorf("renaming temp database to final: %w", err)
	}

	return nil
}

// indexAll performs two-pass indexing:
// Pass 1: Parse all notes, build (type,key) -> id map
// Pass 2: Resolve links to edges
func (b *Builder) indexAll(db *sql.DB) error {
	// Pass 1: Parse all notes and build type/key -> id map
	typeKeyMap := make(map[model.TypeKey]model.NoteID)
	notesByPath := make(map[string]*model.Note)

	// Discover all .Rmd files in type directories
	typeDirs, err := b.discoverTypeDirectories()
	if err != nil {
		return fmt.Errorf("discovering type directories: %w", err)
	}

	for _, dir := range typeDirs {
		rmdFiles, err := b.discoverRmdFiles(dir)
		if err != nil {
			return fmt.Errorf("discovering .Rmd files in %s: %w", dir, err)
		}

		for _, rmdPath := range rmdFiles {
			// Parse note
			content, err := os.ReadFile(rmdPath)
			if err != nil {
				// Skip files we can't read, but continue indexing
				continue
			}

			parsedNote := note.Parse(rmdPath, content)

			// Store note for pass 2
			notesByPath[rmdPath] = parsedNote

			// Build type/key -> id map (only if note has valid frontmatter)
			if parsedNote.FM.ID != "" && parsedNote.FM.Type != "" && parsedNote.FM.Key != "" {
				typeKey := model.TypeKey{
					Type: parsedNote.FM.Type,
					Key:  parsedNote.FM.Key,
				}
				typeKeyMap[typeKey] = parsedNote.FM.ID

				// Get file stats for change detection
				fileInfo, err := os.Stat(rmdPath)
				if err != nil {
					return fmt.Errorf("getting file stats for %s: %w", rmdPath, err)
				}

				// Upsert node in database
				// For Phase 2, we use mtime and size for change detection (hash can be added later)
				mtimeNs := fileInfo.ModTime().UnixNano()
				sizeBytes := fileInfo.Size()
				hash := "" // Hash can be computed later if needed for suspicious changes

				if err := store.UpsertNode(db, parsedNote.FM.ID, parsedNote.FM.Type, parsedNote.FM.Key, parsedNote.FM.Title, parsedNote.FM.State, parsedNote.FM.Created, parsedNote.FM.Updated, rmdPath, mtimeNs, sizeBytes, hash); err != nil {
					return fmt.Errorf("upserting node %s: %w", parsedNote.FM.ID, err)
				}

				// Replace tags for node
				if err := store.ReplaceTagsForNode(db, parsedNote.FM.ID, parsedNote.FM.Tags); err != nil {
					return fmt.Errorf("replacing tags for node %s: %w", parsedNote.FM.ID, err)
				}

				// Insert diagnostics
				if len(parsedNote.Diags) > 0 {
					if err := store.InsertDiagnostics(db, parsedNote.FM.ID, parsedNote.Diags); err != nil {
						return fmt.Errorf("inserting diagnostics for node %s: %w", parsedNote.FM.ID, err)
					}
				}
			}
		}
	}

	// Pass 2: Resolve links to edges
	for _, parsedNote := range notesByPath {
		if parsedNote.FM.ID == "" {
			// Skip notes without valid IDs
			continue
		}

		// Resolve links
		resolvedEdges, diags := b.resolveLinks(parsedNote.RawLinks, typeKeyMap, parsedNote.FM.Type)

		// Replace edges for node
		if err := store.ReplaceEdgesForNode(db, parsedNote.FM.ID, resolvedEdges); err != nil {
			return fmt.Errorf("replacing edges for node %s: %w", parsedNote.FM.ID, err)
		}

		// Add link resolution diagnostics
		if len(diags) > 0 {
			existingDiags := parsedNote.Diags
			allDiags := make([]model.Diagnostic, 0, len(existingDiags)+len(diags))
			allDiags = append(allDiags, existingDiags...)
			allDiags = append(allDiags, diags...)
			if err := store.InsertDiagnostics(db, parsedNote.FM.ID, allDiags); err != nil {
				return fmt.Errorf("inserting link resolution diagnostics for node %s: %w", parsedNote.FM.ID, err)
			}
		}
	}

	return nil
}

// discoverTypeDirectories discovers type directories from config or filesystem
func (b *Builder) discoverTypeDirectories() (map[model.TypeName]string, error) {
	typeDirs := make(map[model.TypeName]string)

	// Use types from config
	for typeName := range b.cfg.Types {
		typeDir := filepath.Join(b.vaultRoot, string(typeName))
		// Check if directory exists
		if info, err := os.Stat(typeDir); err == nil && info.IsDir() {
			typeDirs[typeName] = typeDir
		}
	}

	return typeDirs, nil
}

// discoverRmdFiles discovers all .Rmd files in a directory
func (b *Builder) discoverRmdFiles(dir string) ([]string, error) {
	var rmdFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".Rmd" {
			rmdFiles = append(rmdFiles, path)
		}
		return nil
	})

	return rmdFiles, err
}

// resolveLinks resolves raw links to edges, returning resolved edges and diagnostics
// This implements the link resolution rules:
// - [[type:key]] → direct resolution
// - [[key]] → uniqueness check, diagnostic if ambiguous
// - Unknown targets → diagnostics + unresolved edge rows
func (b *Builder) resolveLinks(rawLinks []model.RawLink, typeKeyMap map[model.TypeKey]model.NoteID, sourceType model.TypeName) ([]model.RawLink, []model.Diagnostic) {
	var resolvedEdges []model.RawLink
	var diags []model.Diagnostic

	for _, link := range rawLinks {
		var targetID *model.NoteID
		var diagnostic *model.Diagnostic

		if link.Target.Type != nil {
			// Qualified link: [[type:key]]
			typeKey := model.TypeKey{
				Type: *link.Target.Type,
				Key:  link.Target.Key,
			}
			if id, ok := typeKeyMap[typeKey]; ok {
				targetID = &id
			} else {
				// Unknown target
				diagnostic = &model.Diagnostic{
					Level:   model.DiagnosticLevelWarn,
					Code:    "UNRESOLVED_LINK",
					Message: fmt.Sprintf("Link target '%s:%s' not found. The target note may not exist or may not have been indexed yet. Use 'touchlog index rebuild' to update the index.", *link.Target.Type, link.Target.Key),
					Span:    link.Span,
				}
			}
		} else {
			// Unqualified link: [[key]]
			// Check for uniqueness across all types
			var matches []model.TypeKey
			for typeKey, id := range typeKeyMap {
				if typeKey.Key == link.Target.Key {
					matches = append(matches, typeKey)
					if targetID == nil {
						targetID = &id
					}
				}
			}

			if len(matches) == 0 {
				// Unknown target
				diagnostic = &model.Diagnostic{
					Level:   model.DiagnosticLevelWarn,
					Code:    "UNRESOLVED_LINK",
					Message: fmt.Sprintf("Link target '%s' not found. The target note may not exist or may not have been indexed yet. Use a qualified link (type:key) or run 'touchlog index rebuild'.", link.Target.Key),
					Span:    link.Span,
				}
			} else if len(matches) > 1 {
				// Ambiguous: multiple matches
				typeNames := make([]string, len(matches))
				for i, match := range matches {
					typeNames[i] = string(match.Type)
				}
				diagnostic = &model.Diagnostic{
					Level:   model.DiagnosticLevelError,
					Code:    "AMBIGUOUS_LINK",
					Message: fmt.Sprintf("Link target '%s' is ambiguous - matches %d types: %s. Use a qualified link (type:key) to specify the target type.", link.Target.Key, len(matches), strings.Join(typeNames, ", ")),
					Span:    link.Span,
				}
				// Don't set targetID for ambiguous links
				targetID = nil
			}
			// If len(matches) == 1, targetID is already set
		}

		// Create resolved edge with resolved ID
		resolvedLink := link
		resolvedLink.ResolvedToID = targetID // May be nil for unresolved links

		resolvedEdges = append(resolvedEdges, resolvedLink)

		if diagnostic != nil {
			diags = append(diags, *diagnostic)
		}
	}

	return resolvedEdges, diags
}

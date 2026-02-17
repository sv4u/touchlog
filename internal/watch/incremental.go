package watch

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/note"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// IncrementalIndexer handles incremental indexing of changed files.
// Each event opens a short-lived DB connection for transaction isolation.
type IncrementalIndexer struct {
	vaultRoot string
	cfg       *config.Config
}

// NewIncrementalIndexer creates a new incremental indexer
func NewIncrementalIndexer(vaultRoot string, cfg *config.Config) *IncrementalIndexer {
	return &IncrementalIndexer{
		vaultRoot: vaultRoot,
		cfg:       cfg,
	}
}

// ProcessEvent processes a single filesystem event
// Uses transaction-per-batch pattern: opens DB, applies updates in transaction, commits and closes
func (ii *IncrementalIndexer) ProcessEvent(event Event) error {
	// Open a new database connection for this transaction
	db, err := store.OpenOrCreateDB(ii.vaultRoot)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Ensure migrations are applied before using the database
	if err := store.ApplyMigrations(db); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Process based on event operation
	switch {
	case event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Create != 0:
		// File created or modified
		return ii.processFileUpdate(tx, event.Path)

	case event.Op&fsnotify.Remove != 0:
		// File deleted
		if err := ii.processFileDeleteTx(tx, event.Path); err != nil {
			return err
		}
		return tx.Commit()

	default:
		// Other operations (Rename, Chmod) - ignore for now
		return nil
	}
}

// processFileUpdate processes a file update (create or modify)
func (ii *IncrementalIndexer) processFileUpdate(tx *sql.Tx, filePath string) error {
	// Check if file exists (might have been deleted between event and processing)
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// File was deleted, treat as deletion
		return ii.processFileDeleteTx(tx, filePath)
	}
	if err != nil {
		return fmt.Errorf("stating file: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Parse note
	parsedNote := note.Parse(filePath, content)

	// Check if note has valid frontmatter
	if parsedNote.FM.ID == "" || parsedNote.FM.Type == "" || parsedNote.FM.Key == "" {
		// Invalid note, skip indexing
		return nil
	}

	// Check if file has changed (mtime + size)
	// Get current file stats
	mtimeNs := fileInfo.ModTime().UnixNano()
	sizeBytes := fileInfo.Size()

	// Check if node exists and if it needs updating
	var existingMtimeNs, existingSizeBytes int64
	var existingPath string
	err = tx.QueryRow(`
		SELECT mtime_ns, size_bytes, path
		FROM nodes
		WHERE id = ?
	`, parsedNote.FM.ID).Scan(&existingMtimeNs, &existingSizeBytes, &existingPath)

	if err == sql.ErrNoRows {
		// New node - insert it
		// First, we need to build the type/key -> id map and last-segment map for link resolution
		// For incremental updates, we'll load them from the database
		typeKeyMap, lastSegmentMap, err := ii.loadTypeKeyMap(tx)
		if err != nil {
			return fmt.Errorf("loading type/key map: %w", err)
		}

		// Upsert node
		if err := ii.upsertNodeTx(tx, parsedNote, mtimeNs, sizeBytes); err != nil {
			return err
		}

		// Replace tags
		if err := ii.replaceTagsTx(tx, parsedNote.FM.ID, parsedNote.FM.Tags); err != nil {
			return err
		}

		// Resolve and replace edges
		resolvedEdges, diags := note.ResolveLinks(parsedNote.RawLinks, typeKeyMap, lastSegmentMap, parsedNote.FM.Type)
		if err := ii.replaceEdgesTx(tx, parsedNote.FM.ID, resolvedEdges); err != nil {
			return err
		}

		// Insert diagnostics
		allDiags := append(parsedNote.Diags, diags...)
		if len(allDiags) > 0 {
			if err := ii.insertDiagnosticsTx(tx, parsedNote.FM.ID, allDiags); err != nil {
				return err
			}
		}
	} else if err != nil {
		return fmt.Errorf("checking existing node: %w", err)
	} else {
		// Existing node - check if it needs updating
		if existingMtimeNs == mtimeNs && existingSizeBytes == sizeBytes && existingPath == filePath {
			// No change detected, skip
			return nil
		}

		// File changed - update it
		typeKeyMap, lastSegmentMap, err := ii.loadTypeKeyMap(tx)
		if err != nil {
			return fmt.Errorf("loading type/key map: %w", err)
		}

		// Upsert node
		if err := ii.upsertNodeTx(tx, parsedNote, mtimeNs, sizeBytes); err != nil {
			return err
		}

		// Replace tags
		if err := ii.replaceTagsTx(tx, parsedNote.FM.ID, parsedNote.FM.Tags); err != nil {
			return err
		}

		// Resolve and replace edges
		resolvedEdges, diags := note.ResolveLinks(parsedNote.RawLinks, typeKeyMap, lastSegmentMap, parsedNote.FM.Type)
		if err := ii.replaceEdgesTx(tx, parsedNote.FM.ID, resolvedEdges); err != nil {
			return err
		}

		// Replace diagnostics
		allDiags := append(parsedNote.Diags, diags...)
		if err := ii.replaceDiagnosticsTx(tx, parsedNote.FM.ID, allDiags); err != nil {
			return err
		}
	}

	// Commit transaction
	return tx.Commit()
}

// processFileDeleteTx processes a file deletion within a transaction
func (ii *IncrementalIndexer) processFileDeleteTx(tx *sql.Tx, filePath string) error {
	// Find node by path
	var nodeID string
	err := tx.QueryRow(`
		SELECT id
		FROM nodes
		WHERE path = ?
	`, filePath).Scan(&nodeID)

	if err == sql.ErrNoRows {
		// Node not found, nothing to delete
		return nil
	}
	if err != nil {
		return fmt.Errorf("finding node by path: %w", err)
	}

	// Delete node (CASCADE will handle edges, tags, diagnostics)
	_, err = tx.Exec("DELETE FROM nodes WHERE id = ?", nodeID)
	if err != nil {
		return fmt.Errorf("deleting node: %w", err)
	}

	return nil
}

// loadTypeKeyMap loads the type/key -> id map and last-segment map from the database
func (ii *IncrementalIndexer) loadTypeKeyMap(tx *sql.Tx) (map[model.TypeKey]model.NoteID, map[string][]model.NoteID, error) {
	rows, err := tx.Query("SELECT id, type, key FROM nodes")
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	typeKeyMap := make(map[model.TypeKey]model.NoteID)
	lastSegmentMap := make(map[string][]model.NoteID)
	for rows.Next() {
		var id, typ, key string
		if err := rows.Scan(&id, &typ, &key); err != nil {
			return nil, nil, err
		}
		noteID := model.NoteID(id)
		typeKeyMap[model.TypeKey{
			Type: model.TypeName(typ),
			Key:  model.Key(key),
		}] = noteID

		// Also index by last segment for unqualified link resolution
		lastSeg := config.LastSegment(key)
		lastSegmentMap[lastSeg] = append(lastSegmentMap[lastSeg], noteID)
	}

	return typeKeyMap, lastSegmentMap, rows.Err()
}

// Helper functions for transaction-based operations

func (ii *IncrementalIndexer) upsertNodeTx(tx *sql.Tx, parsedNote *model.Note, mtimeNs, sizeBytes int64) error {
	_, err := tx.Exec(`
		INSERT INTO nodes (id, type, key, title, state, created, updated, path, mtime_ns, size_bytes, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			key = excluded.key,
			title = excluded.title,
			state = excluded.state,
			created = excluded.created,
			updated = excluded.updated,
			path = excluded.path,
			mtime_ns = excluded.mtime_ns,
			size_bytes = excluded.size_bytes,
			hash = excluded.hash
	`, parsedNote.FM.ID, parsedNote.FM.Type, parsedNote.FM.Key, parsedNote.FM.Title, parsedNote.FM.State,
		parsedNote.FM.Created.Format(time.RFC3339), parsedNote.FM.Updated.Format(time.RFC3339),
		parsedNote.Path, mtimeNs, sizeBytes, "")
	return err
}

func (ii *IncrementalIndexer) replaceTagsTx(tx *sql.Tx, nodeID model.NoteID, tags []string) error {
	// Delete existing tags
	if _, err := tx.Exec("DELETE FROM tags WHERE node_id = ?", nodeID); err != nil {
		return err
	}

	// Insert new tags
	for _, tag := range tags {
		if _, err := tx.Exec("INSERT INTO tags (node_id, tag) VALUES (?, ?)", nodeID, tag); err != nil {
			return err
		}
	}

	return nil
}

func (ii *IncrementalIndexer) replaceEdgesTx(tx *sql.Tx, fromID model.NoteID, edges []model.RawLink) error {
	// Delete existing edges
	if _, err := tx.Exec("DELETE FROM edges WHERE from_id = ?", fromID); err != nil {
		return err
	}

	// Insert new edges (using store's logic)
	// For now, we'll duplicate the logic here
	for _, edge := range edges {
		targetJSON, err := json.Marshal(edge.Target)
		if err != nil {
			return fmt.Errorf("marshaling edge target: %w", err)
		}
		spanJSON, err := json.Marshal(edge.Span)
		if err != nil {
			return fmt.Errorf("marshaling edge span: %w", err)
		}

		var toID *string
		if edge.ResolvedToID != nil {
			toIDStr := string(*edge.ResolvedToID)
			toID = &toIDStr
		}

		if _, err := tx.Exec(`
			INSERT INTO edges (from_id, to_id, edge_type, raw_target, span)
			VALUES (?, ?, ?, ?, ?)
		`, fromID, toID, edge.EdgeType, string(targetJSON), string(spanJSON)); err != nil {
			return err
		}
	}

	return nil
}

func (ii *IncrementalIndexer) insertDiagnosticsTx(tx *sql.Tx, nodeID model.NoteID, diags []model.Diagnostic) error {
	// Delete existing diagnostics
	if _, err := tx.Exec("DELETE FROM diagnostics WHERE node_id = ?", nodeID); err != nil {
		return err
	}

	// Insert new diagnostics
	now := time.Now().UTC().Format(time.RFC3339)
	for _, diag := range diags {
		spanJSON, err := json.Marshal(diag.Span)
		if err != nil {
			return fmt.Errorf("marshaling diagnostic span: %w", err)
		}
		if _, err := tx.Exec(`
			INSERT INTO diagnostics (node_id, level, code, message, span, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, nodeID, diag.Level, diag.Code, diag.Message, string(spanJSON), now); err != nil {
			return err
		}
	}

	return nil
}

func (ii *IncrementalIndexer) replaceDiagnosticsTx(tx *sql.Tx, nodeID model.NoteID, diags []model.Diagnostic) error {
	return ii.insertDiagnosticsTx(tx, nodeID, diags)
}


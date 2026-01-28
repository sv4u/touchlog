package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/store"
	cli3 "github.com/urfave/cli/v3"
)

// BuildEditCommand builds the edit command
func BuildEditCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "edit",
		Usage: "Open an existing note for editing",
		Description: `Opens an interactive fuzzy-search wizard to find and edit notes.

Examples:
  touchlog edit                    # Interactive wizard
  touchlog edit --key note:my-key  # Direct edit by type:key
  touchlog edit --key my-key       # Direct edit by key (unqualified)
  touchlog edit --type note        # Pre-filter by type`,
		Flags: []cli3.Flag{
			&cli3.StringFlag{
				Name:  "key",
				Usage: "Directly open note by type:key or key (skips wizard)",
			},
			&cli3.StringFlag{
				Name:  "type",
				Usage: "Pre-filter notes by type",
			},
			&cli3.StringSliceFlag{
				Name:  "tag",
				Usage: "Pre-filter notes by tag (can be repeated)",
			},
		},
		Action: editAction,
	}
}

// editAction handles the edit command
func editAction(ctx context.Context, cmd *cli3.Command) error {
	vaultRoot, err := GetVaultFromContext(ctx, cmd)
	if err != nil {
		return fmt.Errorf("resolving vault: %w", err)
	}

	// Validate vault exists
	if err := ValidateVault(vaultRoot); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig(vaultRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Check if index exists
	dbPath := filepath.Join(vaultRoot, ".touchlog", "index.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("index not found. Run 'touchlog index' first")
	}

	// Get flags
	keyFlag := cmd.String("key")
	typeFilter := cmd.String("type")
	tagFilters := cmd.StringSlice("tag")

	// If --key is provided, directly open the note
	if keyFlag != "" {
		return editByKey(vaultRoot, cfg, keyFlag)
	}

	// Check if we're in an interactive terminal
	if !isInteractiveTerminal() {
		return fmt.Errorf("not in an interactive terminal. Use --key flag to specify the note directly")
	}

	// Run interactive wizard
	return runEditWizard(vaultRoot, cfg, typeFilter, tagFilters)
}

// editByKey opens a note directly by its key (type:key or just key)
func editByKey(vaultRoot string, cfg *config.Config, identifier string) error {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Resolve the note path
	notePath, err := resolveNotePath(db, vaultRoot, identifier)
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note file not found: %s", notePath)
	}

	// Launch editor
	if err := launchEditor(cfg, notePath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to launch editor: %v\n", err)
		fmt.Printf("Note path: %s\n", notePath)
	}

	return nil
}

// resolveNotePath resolves a note identifier to its file path
func resolveNotePath(db *sql.DB, vaultRoot, identifier string) (string, error) {
	// Parse identifier
	parts := strings.Split(identifier, ":")
	var nodeType *string
	var key string

	if len(parts) == 2 {
		// Qualified: type:key
		nodeTypeStr := parts[0]
		nodeType = &nodeTypeStr
		key = parts[1]
	} else if len(parts) == 1 {
		// Unqualified: key
		key = parts[0]
	} else {
		return "", fmt.Errorf("invalid note identifier format: %s (expected 'type:key' or 'key')", identifier)
	}

	if nodeType != nil {
		// Qualified lookup - exact match on (type, key)
		var path string
		err := db.QueryRow("SELECT path FROM nodes WHERE type = ? AND key = ?", *nodeType, key).Scan(&path)
		if err != nil {
			return "", fmt.Errorf("note not found: %s:%s", *nodeType, key)
		}
		// Path in database is already absolute
		return path, nil
	}

	// Unqualified lookup - first try exact match on full key, then fall back to last-segment matching
	rows, err := db.Query("SELECT type, key, path FROM nodes")
	if err != nil {
		return "", fmt.Errorf("querying nodes: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var exactMatches []struct {
		typ, key, path string
	}
	var lastSegMatches []struct {
		typ, key, path string
	}

	searchLastSeg := config.LastSegment(key)

	for rows.Next() {
		var typ, nodeKey, path string
		if err := rows.Scan(&typ, &nodeKey, &path); err != nil {
			return "", err
		}

		// Check for exact match on full key first
		if nodeKey == key {
			exactMatches = append(exactMatches, struct{ typ, key, path string }{typ, nodeKey, path})
		}

		// Also collect last-segment matches for fallback
		nodeLastSeg := config.LastSegment(nodeKey)
		if nodeLastSeg == searchLastSeg {
			lastSegMatches = append(lastSegMatches, struct{ typ, key, path string }{typ, nodeKey, path})
		}
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	// Priority 1: Exact match on full key
	if len(exactMatches) == 1 {
		// Path in database is already absolute
		return exactMatches[0].path, nil
	}
	if len(exactMatches) > 1 {
		keys := make([]string, len(exactMatches))
		for i, m := range exactMatches {
			keys[i] = fmt.Sprintf("%s:%s", m.typ, m.key)
		}
		return "", fmt.Errorf("ambiguous note identifier '%s': matches %d notes (%s). Use qualified identifier (type:key)", key, len(exactMatches), strings.Join(keys, ", "))
	}

	// Priority 2: Fall back to last-segment matching
	if len(lastSegMatches) == 0 {
		return "", fmt.Errorf("note not found: %s", identifier)
	}
	if len(lastSegMatches) == 1 {
		// Path in database is already absolute
		return lastSegMatches[0].path, nil
	}

	// Multiple last-segment matches
	keys := make([]string, len(lastSegMatches))
	for i, m := range lastSegMatches {
		keys[i] = fmt.Sprintf("%s:%s", m.typ, m.key)
	}
	return "", fmt.Errorf("ambiguous note identifier '%s': matches %d notes (%s). Use qualified identifier (type:full/path/key)", key, len(lastSegMatches), strings.Join(keys, ", "))
}

// runEditWizard runs the interactive edit wizard
func runEditWizard(vaultRoot string, cfg *config.Config, typeFilter string, tagFilters []string) error {
	// Load notes
	notes, err := loadNotesForEdit(vaultRoot, typeFilter, tagFilters)
	if err != nil {
		return fmt.Errorf("loading notes: %w", err)
	}

	if len(notes) == 0 {
		return fmt.Errorf("no notes found. Create one with 'touchlog new'")
	}

	// Create and run the wizard
	model := initialEditModel(notes, vaultRoot)
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("running wizard: %w", err)
	}

	// Check if a note was selected
	if editModel, ok := finalModel.(editWizardModel); ok {
		if editModel.selected != nil {
			// Launch editor for the selected note
			// Path in database is already absolute
			notePath := editModel.selected.path
			if err := launchEditor(cfg, notePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to launch editor: %v\n", err)
				fmt.Printf("Note path: %s\n", notePath)
			}
		}
	}

	return nil
}

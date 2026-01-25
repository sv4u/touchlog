package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
	cli3 "github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

// BuildNewCommand builds the new command
func BuildNewCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "new",
		Usage: "Create a new note",
		Action: func(ctx context.Context, cmd *cli3.Command) error {
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

			return runNewWizard(vaultRoot, cfg)
		},
	}
}

// runNewWizard runs the interactive wizard for creating a new note
func runNewWizard(vaultRoot string, cfg *config.Config) error {
	if len(cfg.Types) == 0 {
		return fmt.Errorf("no types configured in vault (run 'touchlog init' first)")
	}

	// Check if we're in an interactive terminal
	if isInteractiveTerminal() {
		return runInteractiveWizard(vaultRoot, cfg)
	}

	// Non-interactive mode (for tests)
	return runNonInteractiveWizard(vaultRoot, cfg)
}

// runInteractiveWizard runs the bubbletea interactive wizard
func runInteractiveWizard(vaultRoot string, cfg *config.Config) error {
	model := initialModel(vaultRoot, cfg)
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("running wizard: %w", err)
	}

	// Launch editor if note was created successfully
	if wizardModel, ok := finalModel.(wizardModel); ok {
		if wizardModel.notePath != "" {
			if err := launchEditor(cfg, wizardModel.notePath); err != nil {
				// Don't fail the command if editor launch fails, just warn
				fmt.Fprintf(os.Stderr, "Warning: failed to launch editor: %v\n", err)
			}
		}
	}

	return nil
}

// runNonInteractiveWizard runs the wizard in non-interactive mode (for tests)
func runNonInteractiveWizard(vaultRoot string, cfg *config.Config) error {
	// Step 1: Select type
	typeName, err := selectType(cfg)
	if err != nil {
		return fmt.Errorf("selecting type: %w", err)
	}

	typeDef := cfg.Types[typeName]

	// Step 2: Input key
	key, err := inputKey(typeDef, vaultRoot, typeName)
	if err != nil {
		return fmt.Errorf("inputting key: %w", err)
	}

	// Step 3: Input title
	title, err := inputTitle()
	if err != nil {
		return fmt.Errorf("inputting title: %w", err)
	}

	// Step 4: Input tags
	tags, err := inputTags()
	if err != nil {
		return fmt.Errorf("inputting tags: %w", err)
	}

	// Step 5: Select state (default from type)
	state := typeDef.DefaultState

	// Step 6: Input filename
	filename, err := inputFilename(vaultRoot, typeName, key, title)
	if err != nil {
		return fmt.Errorf("inputting filename: %w", err)
	}

	// Step 7: Determine path with branching logic for backward compatibility
	var notePath string
	keyStr := string(key)
	if strings.Contains(keyStr, "/") {
		// Path-based key: file in subfolder
		notePath = filepath.Join(vaultRoot, string(typeName), keyStr, filename+".Rmd")
	} else {
		// Flat key: file at type root (backward compatible)
		notePath = filepath.Join(vaultRoot, string(typeName), filename+".Rmd")
	}
	typeDir := filepath.Dir(notePath)

	// Step 8: Prompt for directory creation if missing
	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		// For Phase 1, we'll create it automatically
		// In a full implementation, we'd prompt: "Create directory? (y/N)"
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return fmt.Errorf("creating type directory: %w", err)
		}
	}

	// Step 9: Generate frontmatter and body
	now := time.Now().UTC()
	noteID := generateNoteID()
	frontmatter := generateFrontmatter(noteID, typeName, key, title, tags, state, now)
	body := generateBody(title, cfg, typeName)

	// Step 10: Write file atomically
	content := formatNote(frontmatter, body)
	if err := AtomicWrite(notePath, content); err != nil {
		return fmt.Errorf("writing note: %w", err)
	}

	fmt.Printf("✓ Created %s\n", notePath)
	fmt.Printf("Note ID: %s\n", noteID)

	// Step 11: Launch editor
	if err := launchEditor(cfg, notePath); err != nil {
		// Don't fail the command if editor launch fails, just warn
		fmt.Fprintf(os.Stderr, "Warning: failed to launch editor: %v\n", err)
		fmt.Println("\nNote created successfully. Edit it with your preferred editor.")
	}

	return nil
}

// selectType prompts user to select a type (for Phase 1, prefers "note" type)
func selectType(cfg *config.Config) (model.TypeName, error) {
	// For Phase 1, prefer "note" type if available, otherwise use first type
	// In a full implementation, we'd show a list and prompt
	if _, ok := cfg.Types["note"]; ok {
		return "note", nil
	}

	// Fallback to first type
	for typeName := range cfg.Types {
		return typeName, nil
	}
	return "", fmt.Errorf("no types available")
}

// inputKey prompts for and validates a key
func inputKey(typeDef config.TypeDef, vaultRoot string, typeName model.TypeName) (model.Key, error) {
	// For Phase 1, we'll use a simple default
	// In a full implementation, we'd prompt and validate interactively
	keyStr := "new-note"

	// Validate key using the new ValidateKey function that supports path-based keys
	if err := config.ValidateKey(keyStr, typeDef.KeyPattern, typeDef.KeyMaxLen); err != nil {
		return "", fmt.Errorf("invalid key: %w", err)
	}

	// Check uniqueness using store if it exists (for indexed notes)
	if err := checkKeyUniqueness(vaultRoot, typeName, keyStr); err != nil {
		return "", err
	}

	return model.Key(keyStr), nil
}

// checkKeyUniqueness checks if a key already exists in the store
func checkKeyUniqueness(vaultRoot string, typeName model.TypeName, keyStr string) error {
	dbPath := filepath.Join(vaultRoot, ".touchlog", "index.db")
	if _, err := os.Stat(dbPath); err == nil {
		db, err := store.OpenOrCreateDB(vaultRoot)
		if err == nil {
			defer func() {
				_ = db.Close()
			}()
			var exists int
			err := db.QueryRow("SELECT 1 FROM nodes WHERE type = ? AND key = ?", typeName, keyStr).Scan(&exists)
			if err == nil && exists == 1 {
				return fmt.Errorf("note with key %q already exists in type %q", keyStr, typeName)
			}
		}
	}
	return nil
}

// inputTitle prompts for title (for Phase 1, uses default)
func inputTitle() (string, error) {
	// For Phase 1, use a default
	// In a full implementation, we'd prompt
	return "New Note", nil
}

// inputTags prompts for tags (for Phase 1, uses empty)
func inputTags() ([]string, error) {
	// For Phase 1, use empty tags
	// In a full implementation, we'd prompt and parse comma-separated values
	return []string{}, nil
}

// inputFilename prompts for output filename
func inputFilename(vaultRoot string, typeName model.TypeName, key model.Key, title string) (string, error) {
	// For Phase 1, we'll use the title as default
	// In a full implementation, we'd prompt interactively
	// For non-interactive mode (tests), use sanitized title as default
	filenameStr := sanitizeTitleForFilename(title)

	// Remove .Rmd extension if provided (we'll add it automatically)
	filenameStr = strings.TrimSuffix(filenameStr, ".Rmd")
	filenameStr = strings.TrimSuffix(filenameStr, ".rmd")

	// Validate filename (basic validation - no path separators)
	if strings.Contains(filenameStr, string(filepath.Separator)) || strings.Contains(filenameStr, "/") || strings.Contains(filenameStr, "\\") {
		return "", fmt.Errorf("filename cannot contain path separators")
	}

	// Check if file already exists using branching path logic
	var notePath string
	keyStr := string(key)
	if strings.Contains(keyStr, "/") {
		// Path-based key: file in subfolder
		notePath = filepath.Join(vaultRoot, string(typeName), keyStr, filenameStr+".Rmd")
	} else {
		// Flat key: file at type root (backward compatible)
		notePath = filepath.Join(vaultRoot, string(typeName), filenameStr+".Rmd")
	}
	if _, err := os.Stat(notePath); err == nil {
		return "", fmt.Errorf("file %q already exists", filenameStr+".Rmd")
	}

	return filenameStr, nil
}

// generateNoteID generates a unique note ID
func generateNoteID() model.NoteID {
	// For Phase 1, use a simple timestamp-based ID
	// In a full implementation, we'd use a more robust ID generator
	timestamp := time.Now().UnixNano()
	return model.NoteID(fmt.Sprintf("note-%d", timestamp))
}

// generateFrontmatter generates the frontmatter for a new note
func generateFrontmatter(id model.NoteID, typeName model.TypeName, key model.Key, title string, tags []string, state string, now time.Time) map[string]any {
	return map[string]any{
		"id":      string(id),
		"type":    string(typeName),
		"key":     string(key),
		"created": now.Format(time.RFC3339),
		"updated": now.Format(time.RFC3339),
		"title":   title,
		"tags":    tags,
		"state":   state,
	}
}

// generateBody generates the body content for a new note
func generateBody(title string, cfg *config.Config, typeName model.TypeName) string {
	// For Phase 1, create a simple body with a heading
	// In a full implementation, we'd check for templates and render them
	return fmt.Sprintf("# %s\n\n", title)
}

// formatNote formats frontmatter and body into a complete .Rmd file
func formatNote(frontmatter map[string]any, body string) []byte {
	var buf strings.Builder

	// Write frontmatter
	buf.WriteString("---\n")
	fmYAML, err := yaml.Marshal(frontmatter)
	if err != nil {
		// Should not happen, but handle gracefully
		panic(fmt.Sprintf("marshaling frontmatter: %v", err))
	}
	buf.Write(fmYAML)
	buf.WriteString("---\n")

	// Write body with newline between frontmatter and body
	buf.WriteString("\n")
	buf.WriteString(body)

	return []byte(buf.String())
}

// sanitizeTitleForFilename converts a title to a filename-safe string
func sanitizeTitleForFilename(title string) string {
	// Convert to lowercase
	filename := strings.ToLower(title)

	// Replace spaces with hyphens
	filename = strings.ReplaceAll(filename, " ", "-")

	// Remove or replace invalid filename characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", ".", ",", "!", "@", "#", "$", "%", "^", "&", "(", ")", "+", "=", "[", "]", "{", "}", ";", "'"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "")
	}

	// Replace multiple consecutive hyphens with a single hyphen
	for strings.Contains(filename, "--") {
		filename = strings.ReplaceAll(filename, "--", "-")
	}

	// Trim leading and trailing hyphens
	filename = strings.Trim(filename, "-")

	// If empty after sanitization, use a default
	if filename == "" {
		filename = "untitled"
	}

	return filename
}

// launchEditor launches the configured editor to open the specified file
func launchEditor(cfg *config.Config, filePath string) error {
	// If no editor is configured, skip
	if cfg.Editor == "" {
		return nil
	}

	// Parse editor command - split by spaces
	// Note: This simple parsing doesn't handle quoted arguments with spaces.
	// For complex cases, users should ensure editor paths don't contain spaces
	// or use the full path without spaces.
	editorCmd := strings.Fields(cfg.Editor)
	if len(editorCmd) == 0 {
		return fmt.Errorf("editor command is empty")
	}

	// Create command with file path as last argument
	// #nosec G204 - filePath is constructed using filepath.Join from validated inputs
	// and passed directly to exec.Command (not shell), so it's safe from command injection
	cmd := exec.Command(editorCmd[0], append(editorCmd[1:], filePath)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the editor (this will block until editor exits)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running editor %q: %w", cfg.Editor, err)
	}

	return nil
}

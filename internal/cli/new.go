package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/model"
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

	// Step 6: Determine path
	notePath := filepath.Join(vaultRoot, string(typeName), string(key)+".Rmd")
	typeDir := filepath.Dir(notePath)

	// Step 7: Prompt for directory creation if missing
	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		// For Phase 1, we'll create it automatically
		// In a full implementation, we'd prompt: "Create directory? (y/N)"
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return fmt.Errorf("creating type directory: %w", err)
		}
	}

	// Step 8: Generate frontmatter and body
	now := time.Now().UTC()
	noteID := generateNoteID()
	frontmatter := generateFrontmatter(noteID, typeName, key, title, tags, state, now)
	body := generateBody(title, cfg, typeName)

	// Step 9: Write file atomically
	content := formatNote(frontmatter, body)
	if err := AtomicWrite(notePath, content); err != nil {
		return fmt.Errorf("writing note: %w", err)
	}

	fmt.Printf("✓ Created %s\n", notePath)
	fmt.Printf("Note ID: %s\n", noteID)

	// Step 10: Launch editor (skip for Phase 1, will be added later)
	// For now, just inform the user
	fmt.Println("\nNote created successfully. Edit it with your preferred editor.")

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

	// Validate pattern
	if typeDef.KeyPattern != nil {
		if !typeDef.KeyPattern.MatchString(keyStr) {
			return "", fmt.Errorf("key %q does not match pattern %q", keyStr, typeDef.KeyPattern.String())
		}
	}

	// Validate length
	if len(keyStr) > typeDef.KeyMaxLen {
		return "", fmt.Errorf("key %q exceeds maximum length of %d", keyStr, typeDef.KeyMaxLen)
	}

	// Check uniqueness within the type directory
	notePath := filepath.Join(vaultRoot, string(typeName), keyStr+".Rmd")
	if _, err := os.Stat(notePath); err == nil {
		return "", fmt.Errorf("note with key %q already exists in type %q", keyStr, typeName)
	}

	return model.Key(keyStr), nil
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

	// Write body
	buf.WriteString(body)

	return []byte(buf.String())
}

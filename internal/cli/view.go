package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/store"
	cli3 "github.com/urfave/cli/v3"
)

// BuildViewCommand builds the view command
func BuildViewCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "view",
		Usage: "Run rmarkdown::run on an Rmd file",
		Description: `Runs rmarkdown::run() on an R Markdown file to render and view it.

Examples:
  touchlog view                    # Interactive wizard to select file
  touchlog view --file note.Rmd    # Direct file path
  touchlog view --key note:my-key  # Direct note by type:key
  touchlog view --key my-key       # Direct note by key (unqualified)
  touchlog view --type note        # Pre-filter by type in wizard`,
		Flags: []cli3.Flag{
			&cli3.StringFlag{
				Name:  "file",
				Usage: "Direct file path to Rmd file (skips wizard)",
			},
			&cli3.StringFlag{
				Name:  "key",
				Usage: "Directly view note by type:key or key (skips wizard)",
			},
			&cli3.StringFlag{
				Name:  "type",
				Usage: "Pre-filter notes by type (for wizard)",
			},
			&cli3.StringSliceFlag{
				Name:  "tag",
				Usage: "Pre-filter notes by tag (can be repeated, for wizard)",
			},
		},
		Action: viewAction,
	}
}

// viewAction handles the view command
func viewAction(ctx context.Context, cmd *cli3.Command) error {
	// 1. Resolve vault
	vaultRoot, err := GetVaultFromContext(ctx, cmd)
	if err != nil {
		return fmt.Errorf("resolving vault: %w", err)
	}

	// 2. Validate vault exists
	if err := ValidateVault(vaultRoot); err != nil {
		return err
	}

	// 3. Check if index exists (required for --key and wizard)
	dbPath := filepath.Join(vaultRoot, ".touchlog", "index.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Only require index if using --key or wizard (not for --file)
		fileFlag := cmd.String("file")
		keyFlag := cmd.String("key")
		if fileFlag == "" && keyFlag == "" {
			return fmt.Errorf("index not found. Run 'touchlog index' first")
		}
		// If --file is provided, we can skip index check
	}

	// 4. Get flags
	fileFlag := cmd.String("file")
	keyFlag := cmd.String("key")
	typeFilter := cmd.String("type")
	tagFilters := cmd.StringSlice("tag")

	// 5. Route to appropriate handler
	if fileFlag != "" {
		// Direct file path
		return viewByFile(fileFlag)
	}

	if keyFlag != "" {
		// Direct key resolution (requires index)
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			return fmt.Errorf("index not found. Run 'touchlog index' first")
		}
		return viewByKey(vaultRoot, keyFlag)
	}

	// 6. Interactive wizard (requires index and interactive terminal)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("index not found. Run 'touchlog index' first")
	}

	if !isInteractiveTerminal() {
		return fmt.Errorf("not in an interactive terminal. Use --file or --key flag to specify the file")
	}

	// Load config for wizard (needed for consistency, though not strictly required)
	cfg, err := config.LoadConfig(vaultRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	return runViewWizard(vaultRoot, cfg, typeFilter, tagFilters)
}

// viewByFile executes rmarkdown::run on a file path
func viewByFile(filePath string) error {
	// Normalize path to absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolving file path: %w", err)
	}

	// Validate file
	if err := validateRmdFile(absPath); err != nil {
		return err
	}

	// Execute rmarkdown
	return runRmarkdown(absPath)
}

// viewByKey executes rmarkdown::run on a note identified by key
func viewByKey(vaultRoot, key string) error {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Resolve the note path
	notePath, err := resolveNotePath(db, vaultRoot, key)
	if err != nil {
		return err
	}

	// Validate resolved file exists
	if err := validateRmdFile(notePath); err != nil {
		return err
	}

	// Execute rmarkdown
	return runRmarkdown(notePath)
}

// runViewWizard runs the interactive wizard to select a note to view
func runViewWizard(vaultRoot string, cfg *config.Config, typeFilter string, tagFilters []string) error {
	// Load notes (reuse existing function)
	notes, err := loadNotesForEdit(vaultRoot, typeFilter, tagFilters)
	if err != nil {
		return fmt.Errorf("loading notes: %w", err)
	}

	if len(notes) == 0 {
		return fmt.Errorf("no notes found. Create one with 'touchlog new'")
	}

	// Create wizard model (reuse edit wizard, just change title)
	model := initialEditModel(notes, vaultRoot)
	model.list.Title = "Select a note to view" // Customize title

	// Run wizard
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("running wizard: %w", err)
	}

	// Extract selected note and run Rmarkdown
	if editModel, ok := finalModel.(editWizardModel); ok {
		if editModel.selected != nil {
			notePath := editModel.selected.path // Already absolute from database
			return runRmarkdown(notePath)
		}
	}

	return nil // User cancelled
}

// runRmarkdown executes rmarkdown::run() on the specified file
// If ctx is provided and gets cancelled, the Rscript process will be killed
func runRmarkdown(filePath string) error {
	return runRmarkdownWithContext(context.Background(), filePath)
}

// runRmarkdownWithContext executes rmarkdown::run() on the specified file with context support
// The context can be used to cancel/kill the Rscript process (useful for tests)
func runRmarkdownWithContext(ctx context.Context, filePath string) error {
	// Normalize to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolving file path: %w", err)
	}

	// Validate file exists and is readable
	if err := validateRmdFile(absPath); err != nil {
		return err
	}

	// Escape single quotes for R string literal
	escapedPath := strings.ReplaceAll(absPath, "'", "\\'")

	// Set working directory to file's directory for relative paths
	workDir := filepath.Dir(absPath)

	// Build R command: Rscript -e "rmarkdown::run('path')"
	// #nosec G204 - absPath is validated and escaped, exec.Command prevents injection
	cmd := exec.CommandContext(ctx, "Rscript", "-e", fmt.Sprintf("rmarkdown::run('%s')", escapedPath))
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute (blocking, like launchEditor)
	// If ctx is cancelled, the process will be killed
	if err := cmd.Run(); err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("rmarkdown execution cancelled: %w", err)
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("rmarkdown execution timeout: %w", err)
		}
		return fmt.Errorf("running rmarkdown::run: %w\nHint: Make sure R and rmarkdown package are installed", err)
	}

	return nil
}

// validateRmdFile validates that a file exists, is readable, and has the correct extension
func validateRmdFile(filePath string) error {
	// Check file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("checking file: %w", err)
	}

	// Check it's a regular file (not directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Check extension (.Rmd or .rmd)
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".rmd" {
		return fmt.Errorf("file must have .Rmd or .rmd extension, got: %s", ext)
	}

	// Check file is readable
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file is not readable: %w", err)
	}
	file.Close()

	return nil
}

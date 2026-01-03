package editor

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sv4u/touchlog/internal/config"
)

// InternalEditor represents the internal Bubble Tea editor
// This is a simplified wrapper around the existing editor model
// Full refactoring with reusable components will be done in Phase 5.3
type InternalEditor struct {
	filePath string
	content  string
	config   *config.Config
}

// NewInternalEditor creates a new internal editor instance
// For Phase 5, this is a simplified version that will be enhanced later
func NewInternalEditor(filePath string, initialContent string, cfg *config.Config) (*InternalEditor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file path: %w", err)
	}

	return &InternalEditor{
		filePath: absPath,
		content:  initialContent,
		config:   cfg,
	}, nil
}

// Run runs the internal editor (blocking until user finishes)
// This shows template selection first, then editing, then saves to the existing file
func (e *InternalEditor) Run() error {
	// For Phase 5, use the existing editor model with file path override
	// TODO: Phase 5.3 - Refactor to extract reusable TUI components

	// Create model with file path override (so it saves to existing file)
	// Also pass the existing content so it can be loaded instead of template
	m, err := NewModel(
		WithFilePathOverride(e.filePath),
		WithInitialContent(e.content),
	)
	if err != nil {
		return fmt.Errorf("failed to create editor model: %w", err)
	}

	// Create and run the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run editor: %w", err)
	}

	return nil
}

// GetContent returns the edited content
// TODO: This needs to be implemented after Run() completes
func (e *InternalEditor) GetContent() string {
	// Read the file that was saved
	content, err := os.ReadFile(e.filePath)
	if err != nil {
		return e.content // Return initial content if read fails
	}
	return string(content)
}


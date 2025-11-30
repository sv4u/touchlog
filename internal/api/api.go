package api

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sv4u/touchlog/internal/editor"
)

// Options represents configuration options for programmatic touchlog usage
type Options struct {
	// OutputDirectory overrides the notes directory specified in the config file
	OutputDirectory string
	// ConfigPath is optional: path to config file (if empty, uses default)
	// This is reserved for future use
	ConfigPath string
}

// Run creates and runs a touchlog instance with the given options
// This function allows external tools (like a Zettelkasten daemon) to call touchlog programmatically
func Run(opts *Options) error {
	var outputDirOverride string
	if opts != nil {
		outputDirOverride = opts.OutputDirectory
	}

	m, err := editor.NewModel(editor.WithOutputDirectory(outputDirOverride))
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}


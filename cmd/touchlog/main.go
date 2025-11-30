package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sv4u/touchlog/internal/editor"
)

var (
	outputDir      = flag.String("output-dir", "", "Output directory for notes (overrides config file)")
	outputDirShort = flag.String("o", "", "Output directory for notes (shorthand for -output-dir)")
)

// main is the entry point of the program
// In Go, the main function in package main is where execution starts
func main() {
	flag.Parse()

	// Determine output directory override from flags
	var outputDirOverride string
	if *outputDir != "" {
		outputDirOverride = *outputDir
	} else if *outputDirShort != "" {
		outputDirOverride = *outputDirShort
	}

	// Create the initial model
	var m tea.Model
	var err error
	if outputDirOverride != "" {
		m, err = editor.NewModel(editor.WithOutputDirectory(outputDirOverride))
	} else {
		m, err = editor.NewModel()
	}
	if err != nil {
		// Print error and exit
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create a new Bubble Tea program
	// tea.NewProgram initializes the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	// This blocks until the program exits
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

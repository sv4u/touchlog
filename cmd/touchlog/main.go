package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sv4u/touchlog/internal/editor"
)

// main is the entry point of the program
// In Go, the main function in package main is where execution starts
func main() {
	// Create the initial model
	m, err := editor.NewModel()
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


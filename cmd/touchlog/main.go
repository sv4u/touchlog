package main

import (
	"fmt"
	"os"

	"github.com/sv4u/touchlog/cmd/touchlog/commands"
	"github.com/sv4u/touchlog/internal/platform"
)

// main is the entry point of the program
// In Go, the main function in package main is where execution starts
func main() {
	// Platform check (first thing - before any other initialization)
	if err := platform.CheckSupported(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute the root command
	// TODO: When commands.Execute() is implemented, it should return an error
	// that should be checked and handled here:
	// if err := commands.Execute(); err != nil {
	//     fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//     os.Exit(1)
	// }
	commands.Execute()
}

package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// LaunchEditor launches an external editor with the specified file path
// This blocks until the editor exits, allowing proper terminal handoff
// The file remains even if the editor fails to launch
func LaunchEditor(editor string, args []string, filePath string) error {
	if editor == "" {
		return fmt.Errorf("editor command cannot be empty")
	}

	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Validate file path exists and is absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Check if file exists (it should, since we just created it)
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file does not exist: %s", absPath)
	}

	// Build command with file path appended to args
	cmdArgs := make([]string, len(args))
	copy(cmdArgs, args)
	cmdArgs = append(cmdArgs, absPath)

	// Create command
	cmd := exec.Command(editor, cmdArgs...)

	// Set up stdin/stdout/stderr to use the terminal
	// This ensures the editor gets proper terminal control
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the editor process (blocking until editor exits)
	// This properly hands off terminal control to the editor
	// and waits for the editor to finish before touchlog exits
	if err := cmd.Run(); err != nil {
		// Editor exited with error, but file remains
		// Return error so caller can handle it if needed
		return fmt.Errorf("editor '%s' exited with error: %w", editor, err)
	}

	return nil
}

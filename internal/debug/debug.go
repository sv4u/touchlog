package debug

import (
	"fmt"
	"os"
)

// Enabled indicates whether debug mode is active
// This is set by the root command when --debug flag is provided
var Enabled bool

// Log writes a debug message to stderr if debug mode is enabled
// Format: "DEBUG: <message>"
func Log(format string, args ...interface{}) {
	if Enabled {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
	}
}

// LogError writes a debug error message to stderr if debug mode is enabled
// Format: "DEBUG: [context] error: <error>"
func LogError(err error, context string) {
	if Enabled && err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: [%s] error: %v\n", context, err)
	}
}

// LogCollision writes a debug message about file collision handling
// Includes original filename, reason, and resolved filename
func LogCollision(originalPath, reason, resolvedPath string) {
	if Enabled {
		fmt.Fprintf(os.Stderr, "DEBUG: File collision detected\n")
		fmt.Fprintf(os.Stderr, "  Original filename: %s\n", originalPath)
		fmt.Fprintf(os.Stderr, "  Reason: %s\n", reason)
		fmt.Fprintf(os.Stderr, "  Resolved filename: %s\n", resolvedPath)
	}
}

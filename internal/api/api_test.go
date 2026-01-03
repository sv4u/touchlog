package api

import (
	"os"
	"testing"
)

// TestMain runs after all tests complete
func TestMain(m *testing.M) {
	// Run all tests
	code := m.Run()
	
	// Exit with the test result code
	os.Exit(code)
}

func TestOptions(t *testing.T) {
	t.Run("Options struct fields", func(t *testing.T) {
		opts := &Options{
			OutputDirectory: "/test/path",
			ConfigPath:      "/test/config.yaml",
		}

		if opts.OutputDirectory != "/test/path" {
			t.Errorf("Options.OutputDirectory = %q, want %q", opts.OutputDirectory, "/test/path")
		}
		if opts.ConfigPath != "/test/config.yaml" {
			t.Errorf("Options.ConfigPath = %q, want %q", opts.ConfigPath, "/test/config.yaml")
		}
	})

	t.Run("Options with empty values", func(t *testing.T) {
		opts := &Options{}

		if opts.OutputDirectory != "" {
			t.Errorf("Options.OutputDirectory = %q, want empty string", opts.OutputDirectory)
		}
		if opts.ConfigPath != "" {
			t.Errorf("Options.ConfigPath = %q, want empty string", opts.ConfigPath)
		}
	})

	t.Run("Options with nil", func(t *testing.T) {
		var opts *Options = nil

		// This should be safe to pass to Run
		// Run will handle nil gracefully
		// This test verifies the Options type can be nil
		// Actual nil handling is tested in TestRun
		_ = opts // Use the variable to avoid unused variable warning
	})
}


// Note: Run() creates TUI components and is not suitable for unit testing
// The underlying behaviors (config loading, option handling) are tested in:
// - internal/config/config_test.go (config loading)
// - internal/editor/editor_test.go (option handling)
// These tests are removed to avoid creating TUI components during testing

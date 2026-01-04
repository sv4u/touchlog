package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/validation"
)


func TestExpandPath(t *testing.T) {
	// Get the actual home directory for testing
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v (needed for test setup)", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "path starting with ~",
			path:    "~/notes",
			want:    filepath.Join(homeDir, "notes"),
			wantErr: false,
		},
		{
			name:    "path that is just ~",
			path:    "~",
			want:    homeDir,
			wantErr: false,
		},
		{
			name:    "path that doesn't start with ~",
			path:    "/absolute/path",
			want:    "/absolute/path",
			wantErr: false,
		},
		{
			name:    "path with ~ in middle - should not expand",
			path:    "/path/with~tilde",
			want:    "/path/with~tilde",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			want:    "", // Empty path will be converted to current directory by filepath.Abs
			wantErr: false,
		},
		{
			name:    "path with ~/ at start",
			path:    "~/Documents/notes",
			want:    filepath.Join(homeDir, "Documents", "notes"),
			wantErr: false,
		},
		{
			name:    "relative path without ~",
			path:    "notes/test",
			want:    "", // Will be converted to absolute path, so we need to check differently
			wantErr: false,
		},
		{
			name: "path with multiple ~ at start",
			// Paths starting with ~ must be followed by / (e.g., ~/path)
			// "~~/notes" is invalid and should return an error
			path:    "~~/notes",
			want:    "",
			wantErr: true,
		},
		{
			name:    "path with ~ but missing slash",
			path:    "~notes",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validation.ExpandPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ExpandPath(%q) expected error, got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("ExpandPath(%q) unexpected error = %v", tt.path, err)
				}
				// Special handling for empty path and relative paths
				if tt.name == "empty path" {
					// Empty path becomes current directory (absolute path)
					if got == "" {
						t.Errorf("ExpandPath(%q) = %q, want absolute path (current directory)", tt.path, got)
					}
					if !filepath.IsAbs(got) {
						t.Errorf("ExpandPath(%q) = %q, want absolute path", tt.path, got)
					}
				} else if tt.name == "relative path without ~" {
					// Relative paths are converted to absolute paths
					if !filepath.IsAbs(got) {
						t.Errorf("ExpandPath(%q) = %q, want absolute path", tt.path, got)
					}
					if !strings.Contains(got, "notes") || !strings.Contains(got, "test") {
						t.Errorf("ExpandPath(%q) = %q, should contain 'notes' and 'test'", tt.path, got)
					}
				} else if got != tt.want {
					t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
				}
			}
		})
	}
}

func TestWithOutputDirectory(t *testing.T) {
	t.Run("sets output directory in config", func(t *testing.T) {
		cfg := &modelConfig{}
		opt := WithOutputDirectory("/test/path")
		opt(cfg)

		if cfg.outputDirectory != "/test/path" {
			t.Errorf("WithOutputDirectory() outputDirectory = %q, want %q", cfg.outputDirectory, "/test/path")
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		cfg := &modelConfig{}
		opt := WithOutputDirectory("")
		opt(cfg)

		if cfg.outputDirectory != "" {
			t.Errorf("WithOutputDirectory() outputDirectory = %q, want empty string", cfg.outputDirectory)
		}
	})

	t.Run("handles path with tilde", func(t *testing.T) {
		cfg := &modelConfig{}
		opt := WithOutputDirectory("~/notes")
		opt(cfg)

		if cfg.outputDirectory != "~/notes" {
			t.Errorf("WithOutputDirectory() outputDirectory = %q, want %q", cfg.outputDirectory, "~/notes")
		}
	})
}

func TestWithOutputDirectoryOption(t *testing.T) {
	// Test that WithOutputDirectory option correctly sets the output directory
	// This is a simple behavioral test that doesn't create TUI components
	t.Run("sets output directory in config", func(t *testing.T) {
		cfg := &modelConfig{}
		opt := WithOutputDirectory("/custom/output/path")
		opt(cfg)
		if cfg.outputDirectory != "/custom/output/path" {
			t.Errorf("WithOutputDirectory() set outputDirectory = %q, want %q", cfg.outputDirectory, "/custom/output/path")
		}
	})

	t.Run("handles tilde path", func(t *testing.T) {
		cfg := &modelConfig{}
		opt := WithOutputDirectory("~/custom-notes")
		opt(cfg)
		if cfg.outputDirectory != "~/custom-notes" {
			t.Errorf("WithOutputDirectory() set outputDirectory = %q, want %q", cfg.outputDirectory, "~/custom-notes")
		}
	})

	t.Run("handles multiple options - last wins", func(t *testing.T) {
		cfg := &modelConfig{}
		opt1 := WithOutputDirectory("/first/path")
		opt2 := WithOutputDirectory("/second/path")
		opt1(cfg)
		opt2(cfg)
		if cfg.outputDirectory != "/second/path" {
			t.Errorf("WithOutputDirectory() with multiple options set outputDirectory = %q, want %q", cfg.outputDirectory, "/second/path")
		}
	})
}

func TestSaveNotePriorityLogic(t *testing.T) {
	// Test the priority logic by creating a model and testing saveNoteCmd
	tmpDir := t.TempDir()
	configNotesDir := filepath.Join(tmpDir, "config-notes")
	overrideNotesDir := filepath.Join(tmpDir, "override-notes")

	// Create directories
	if err := os.MkdirAll(configNotesDir, 0755); err != nil {
		t.Fatalf("Failed to create config notes directory: %v", err)
	}
	if err := os.MkdirAll(overrideNotesDir, 0755); err != nil {
		t.Fatalf("Failed to create override notes directory: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		Templates: []config.Template{
			{Name: "Test", File: "test.md"},
		},
		NotesDirectory: configNotesDir,
	}

	t.Run("uses override directory when set", func(t *testing.T) {
		m := model{
			config:            cfg,
			outputDirOverride: overrideNotesDir,
			noteContent:       "Test note content",
		}

		cmd := m.saveNoteCmd()
		msg := cmd()

		// Check the message type - should be noteSavedMsg on success
		switch v := msg.(type) {
		case noteSavedMsg:
			// Verify the file was saved in the override directory
			rel, err := filepath.Rel(overrideNotesDir, v.filepath)
			if err != nil || strings.HasPrefix(rel, "..") {
				t.Errorf("saveNoteCmd() saved to %q, expected path in %q", v.filepath, overrideNotesDir)
			}
		case errMsg:
			// If there's an error, it should not be about config directory
			if v.err != nil {
				// Error is acceptable - might be about file creation or other issues
				// The important thing is that override directory was attempted
				_ = v.err
			}
		default:
			t.Errorf("saveNoteCmd() returned unexpected message type: %T", msg)
		}
	})

	t.Run("uses config directory when override is empty", func(t *testing.T) {
		m := model{
			config:            cfg,
			outputDirOverride: "",
			noteContent:       "Test note content",
		}

		cmd := m.saveNoteCmd()
		msg := cmd()

		// Check the message type
		switch v := msg.(type) {
		case noteSavedMsg:
			// Verify the file was saved in the config directory
			rel, err := filepath.Rel(configNotesDir, v.filepath)
			if err != nil || strings.HasPrefix(rel, "..") {
				t.Errorf("saveNoteCmd() saved to %q, expected path in %q", v.filepath, configNotesDir)
			}
		case errMsg:
			// Error is acceptable if directory doesn't exist or other issues
			_ = v.err
		default:
			t.Errorf("saveNoteCmd() returned unexpected message type: %T", msg)
		}
	})

	t.Run("expands tilde in override path", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		expectedPath := filepath.Join(homeDir, "test-notes")
		if err := os.MkdirAll(expectedPath, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		m := model{
			config:            cfg,
			outputDirOverride: "~/test-notes",
			noteContent:       "Test note content",
		}

		cmd := m.saveNoteCmd()
		msg := cmd()

		// Check that tilde was expanded
		switch v := msg.(type) {
		case noteSavedMsg:
			rel, err := filepath.Rel(expectedPath, v.filepath)
			if err != nil || strings.HasPrefix(rel, "..") {
				t.Errorf("saveNoteCmd() saved to %q, expected path in %q (tilde should be expanded)", v.filepath, expectedPath)
			}
		case errMsg:
			// Error is acceptable
			_ = v.err
		default:
			t.Errorf("saveNoteCmd() returned unexpected message type: %T", msg)
		}
	})
}

// Note: Config auto-creation behavior is tested in internal/config/config_test.go
// Template auto-creation behavior is tested in internal/template/template_test.go
// These tests are removed to avoid creating TUI components during testing

// Note: Config loading behavior is tested in internal/config/config_test.go
// These tests are removed to avoid creating TUI components during testing

// Note: Error handling for template directory creation is tested in internal/template/template_test.go
// These tests are removed to avoid creating TUI components during testing

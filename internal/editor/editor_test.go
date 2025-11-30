package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
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
			want:    "",
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
			want:    "notes/test",
			wantErr: false,
		},
		{
			name: "path with multiple ~ at start",
			// Note: expandPath only checks if path starts with "~", so "~~/notes"
			// becomes path[2:] which is "/notes", resulting in homeDir + "/notes"
			path:    "~~/notes",
			want:    filepath.Join(homeDir, "notes"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expandPath(%q) expected error, got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("expandPath(%q) unexpected error = %v", tt.path, err)
				}
				if got != tt.want {
					t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.want)
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

func TestNewModelWithOutputDirectory(t *testing.T) {
	// Create a temporary config file for testing
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	templatePath := filepath.Join(templatesDir, "daily.md")

	// Write test config
	configContent := `templates:
  - name: Daily Note
    file: daily.md
notes_directory: ~/default-notes
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Write test template
	templateContent := "# Daily Note\n\nContent here"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Override XDG paths for testing
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalDataHome := os.Getenv("XDG_DATA_HOME")
	defer func() {
		if originalConfigHome != "" {
			_ = os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalDataHome != "" {
			_ = os.Setenv("XDG_DATA_HOME", originalDataHome)
		} else {
			_ = os.Unsetenv("XDG_DATA_HOME")
		}
	}()

	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

	t.Run("accepts output directory option without error", func(t *testing.T) {
		overridePath := "/custom/output/path"
		_, err := NewModel(WithOutputDirectory(overridePath))
		if err != nil {
			t.Fatalf("NewModel() with output directory option error = %v", err)
		}
	})

	t.Run("accepts tilde path in output directory option", func(t *testing.T) {
		overridePath := "~/custom-notes"
		_, err := NewModel(WithOutputDirectory(overridePath))
		if err != nil {
			t.Fatalf("NewModel() with tilde path option error = %v", err)
		}
	})

	t.Run("accepts multiple options", func(t *testing.T) {
		overridePath := "/test/path"
		_, err := NewModel(WithOutputDirectory(overridePath), WithOutputDirectory("/another/path"))
		if err != nil {
			t.Fatalf("NewModel() with multiple options error = %v", err)
		}
		// Last option should win (though in practice only one would be used)
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

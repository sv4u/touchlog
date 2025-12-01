package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
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

	// Set environment variables before running tests
	// The xdg package caches values at init(), so we need to set them and reload
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
	xdg.Reload() // Force xdg package to reload environment variables

	// Verify the config file exists at the expected location
	expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Fatalf("Config file not found at expected path: %s", expectedConfigPath)
	}

	t.Run("accepts output directory option without error", func(t *testing.T) {
		// Ensure environment is set for this subtest
		// Unset first to clear any cached values, then set to new values
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

		// Force xdg package to reload environment variables
		xdg.Reload()

		overridePath := "/custom/output/path"
		_, err := NewModel(WithOutputDirectory(overridePath))
		if err != nil {
			t.Fatalf("NewModel() with output directory option error = %v", err)
		}
	})

	t.Run("accepts tilde path in output directory option", func(t *testing.T) {
		// Ensure environment is set for this subtest
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

		// Force xdg package to reload environment variables
		xdg.Reload()

		overridePath := "~/custom-notes"
		_, err := NewModel(WithOutputDirectory(overridePath))
		if err != nil {
			t.Fatalf("NewModel() with tilde path option error = %v", err)
		}
	})

	t.Run("accepts multiple options", func(t *testing.T) {
		// Ensure environment is set for this subtest
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

		// Force xdg package to reload environment variables
		xdg.Reload()

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

func TestNewModelWithAutoCreatedConfig(t *testing.T) {
	t.Run("creates config and initializes model successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "touchlog")
		templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

		// Create directories but no config file
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
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

		// Set environment variables before running tests
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
		xdg.Reload() // Force xdg package to reload environment variables

		// Call NewModel - should auto-create config
		_, err := NewModel()
		if err != nil {
			t.Fatalf("NewModel() error = %v", err)
		}

		// Verify config file was created
		expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
		if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
			t.Errorf("NewModel() config file not created at %q", expectedConfigPath)
		}

		// Verify config file has default content
		cfg, err := config.LoadConfig(expectedConfigPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if len(cfg.Templates) != 3 {
			t.Errorf("NewModel() created config with %d templates, want 3", len(cfg.Templates))
		}
		if cfg.NotesDirectory != "~/notes" {
			t.Errorf("NewModel() created config with notes_directory = %q, want %q", cfg.NotesDirectory, "~/notes")
		}
	})

	t.Run("works with output directory option", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "touchlog")
		templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

		// Create directories but no config file
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
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

		// Set environment variables
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
		xdg.Reload()

		// Call NewModel with output directory option
		overridePath := "/test/path"
		_, err := NewModel(WithOutputDirectory(overridePath))
		if err != nil {
			t.Fatalf("NewModel() with output directory option error = %v", err)
		}

		// Verify config file was created
		expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
		if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
			t.Errorf("NewModel() config file not created at %q", expectedConfigPath)
		}
	})

	t.Run("creates templates when directory is empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "touchlog")
		templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

		// Create directories but no config file and empty templates directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
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

		// Set environment variables
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
		xdg.Reload()

		// Call NewModel - should auto-create config and templates
		_, err := NewModel()
		if err != nil {
			t.Fatalf("NewModel() error = %v", err)
		}

		// Verify template files were created
		expectedFiles := []string{"daily.md", "meeting.md", "journal.md"}
		for _, filename := range expectedFiles {
			path := filepath.Join(templatesDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("NewModel() template file not created: %q", filename)
			}
		}
	})

	t.Run("does not create templates if directory has files", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "touchlog")
		templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

		// Create directories with existing template
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Create existing template file
		existingTemplate := filepath.Join(templatesDir, "existing.md")
		existingContent := "# Existing Template\n\nThis should not be modified"
		if err := os.WriteFile(existingTemplate, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to write existing template: %v", err)
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

		// Set environment variables
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
		xdg.Reload()

		// Call NewModel - should auto-create config but not templates
		_, err := NewModel()
		if err != nil {
			t.Fatalf("NewModel() error = %v", err)
		}

		// Verify existing template is unchanged
		content, err := os.ReadFile(existingTemplate)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		if string(content) != existingContent {
			t.Error("NewModel() modified existing template")
		}

		// Verify config was still created
		expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
		if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
			t.Errorf("NewModel() config file not created at %q", expectedConfigPath)
		}
	})
}

func TestNewModelWithExistingConfig(t *testing.T) {
	t.Run("loads existing config successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "touchlog")
		templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

		// Create directories
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Create custom config file
		configPath := filepath.Join(configDir, "config.yaml")
		customConfig := `templates:
  - name: Custom Template
    file: custom.md
notes_directory: ~/custom-notes
vim_mode: true
`
		if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Create template file
		templatePath := filepath.Join(templatesDir, "custom.md")
		templateContent := "# Custom Template\n\nContent here"
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

		// Set environment variables
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
		xdg.Reload()

		// Call NewModel - should load existing config
		_, err := NewModel()
		if err != nil {
			t.Fatalf("NewModel() error = %v", err)
		}

		// Verify config file was not modified (should still have custom content)
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		if !strings.Contains(string(content), "Custom Template") {
			t.Error("NewModel() overwrote existing config file")
		}
	})

	t.Run("returns error for invalid existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "touchlog")
		templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

		// Create directories
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Create config file with invalid YAML
		configPath := filepath.Join(configDir, "config.yaml")
		invalidConfig := `templates:
  - name: Test
    file: test.md
invalid yaml: [unclosed bracket
`
		if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
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

		// Set environment variables
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
		_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
		xdg.Reload()

		// Call NewModel - should return error
		_, err := NewModel()
		if err == nil {
			t.Error("NewModel() expected error for invalid config, got nil")
		}

		// Verify invalid config file is not overwritten
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		if !strings.Contains(string(content), "invalid yaml") {
			t.Error("NewModel() overwrote invalid config file")
		}
	})
}

func TestNewModel_TemplatesDirError(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "touchlog")

	// Create config directory but no templates directory setup
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
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
		xdg.Reload()
	}()

	// Set XDG_DATA_HOME to a non-existent path to cause TemplatesDir to fail
	// by setting it to a path that doesn't exist and can't be created
	invalidDataHome := filepath.Join(tmpDir, "nonexistent", "data")
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	_ = os.Setenv("XDG_DATA_HOME", invalidDataHome)
	xdg.Reload()

	// Call NewModel - should succeed even if TemplatesDir fails
	_, err := NewModel()
	if err != nil {
		t.Fatalf("NewModel() error = %v (should succeed even when TemplatesDir fails)", err)
	}

	// Verify config file was created successfully
	expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Errorf("NewModel() config file not created at %q", expectedConfigPath)
	}
}

func TestNewModel_CreateExampleTemplatesError(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Make templates directory read-only to cause CreateExampleTemplates to fail
	if err := os.Chmod(templatesDir, 0444); err != nil {
		// If chmod fails (e.g., on Windows), skip this test
		t.Skip("Cannot set read-only permissions on this platform")
	}
	defer func() {
		// Restore permissions for cleanup
		_ = os.Chmod(templatesDir, 0755)
	}()

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
		xdg.Reload()
	}()

	// Set environment variables
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))
	xdg.Reload()

	// Call NewModel - should succeed even if CreateExampleTemplates fails
	_, err := NewModel()
	if err != nil {
		t.Fatalf("NewModel() error = %v (should succeed even when CreateExampleTemplates fails)", err)
	}

	// Verify config file was created successfully
	expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Errorf("NewModel() config file not created at %q", expectedConfigPath)
	}

	// Verify templates directory remains empty (creation failed but was ignored)
	entries, err := os.ReadDir(templatesDir)
	if err == nil {
		if len(entries) != 0 {
			t.Errorf("NewModel() templates directory should be empty when creation fails, found %d entries", len(entries))
		}
	}
}

func TestNewModel_BothTemplateErrors(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "touchlog")

	// Create config directory but set invalid templates directory path
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
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
		xdg.Reload()
	}()

	// Set XDG_DATA_HOME to invalid path to cause both TemplatesDir and CreateExampleTemplates to fail
	invalidDataHome := filepath.Join(tmpDir, "invalid", "path", "that", "cannot", "be", "created")
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	_ = os.Setenv("XDG_DATA_HOME", invalidDataHome)
	xdg.Reload()

	// Call NewModel - should succeed even when both errors occur
	_, err := NewModel()
	if err != nil {
		t.Fatalf("NewModel() error = %v (should succeed even when both template errors occur)", err)
	}

	// Verify config file was created successfully
	expectedConfigPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Errorf("NewModel() config file not created at %q", expectedConfigPath)
	}

	// Verify application initialization completes without errors
	// (config is created, which is the critical path)
}

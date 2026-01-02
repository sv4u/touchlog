package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// testXDGEnv manages XDG environment variables for testing
type testXDGEnv struct {
	tmpDir             string
	originalConfigHome string
	originalDataHome   string
}

// setupTestXDG creates a temporary directory structure and sets up XDG environment variables
// It returns a cleanup function that should be called with defer
func setupTestXDG(t *testing.T) *testXDGEnv {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Save original environment variables
	env := &testXDGEnv{
		tmpDir:             tmpDir,
		originalConfigHome: os.Getenv("XDG_CONFIG_HOME"),
		originalDataHome:   os.Getenv("XDG_DATA_HOME"),
	}

	// Set new environment variables
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

	// Reload xdg package once (instead of multiple times per test)
	xdg.Reload()

	return env
}

// cleanup restores the original XDG environment variables
func (env *testXDGEnv) cleanup() {
	if env.originalConfigHome != "" {
		_ = os.Setenv("XDG_CONFIG_HOME", env.originalConfigHome)
	} else {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
	}
	if env.originalDataHome != "" {
		_ = os.Setenv("XDG_DATA_HOME", env.originalDataHome)
	} else {
		_ = os.Unsetenv("XDG_DATA_HOME")
	}
	// Reload once after cleanup
	xdg.Reload()
}

// ConfigDir returns the config directory path
func (env *testXDGEnv) ConfigDir() string {
	return filepath.Join(env.tmpDir, ".config", "touchlog")
}

// TemplatesDir returns the templates directory path
func (env *testXDGEnv) TemplatesDir() string {
	return filepath.Join(env.tmpDir, ".local", "share", "touchlog", "templates")
}

// ConfigPath returns the full path to the config file
func (env *testXDGEnv) ConfigPath() string {
	return filepath.Join(env.ConfigDir(), "config.yaml")
}

// TemplatePath returns the full path to a template file
func (env *testXDGEnv) TemplatePath(filename string) string {
	return filepath.Join(env.TemplatesDir(), filename)
}

// WriteConfig writes a config file with the given content
func (env *testXDGEnv) WriteConfig(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile(env.ConfigPath(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}

// WriteTemplate writes a template file with the given content
func (env *testXDGEnv) WriteTemplate(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.WriteFile(env.TemplatePath(filename), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}
}

// WriteDefaultConfig writes a default config file for testing
func (env *testXDGEnv) WriteDefaultConfig(t *testing.T) {
	t.Helper()
	env.WriteConfig(t, `templates:
  - name: Daily Note
    file: daily.md
notes_directory: ~/default-notes
`)
}

// WriteDefaultTemplate writes a default template file for testing
func (env *testXDGEnv) WriteDefaultTemplate(t *testing.T) {
	t.Helper()
	env.WriteTemplate(t, "daily.md", "# Daily Note\n\nContent here")
}

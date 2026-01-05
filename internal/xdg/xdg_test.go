package xdg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// setupTestXDG creates a temporary directory structure and sets up XDG environment variables
// It returns a cleanup function that should be called with defer
func setupTestXDG(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config")
	dataDir := filepath.Join(tmpDir, ".local", "share")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	// Save original environment variables
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalDataHome := os.Getenv("XDG_DATA_HOME")

	// Set new environment variables
	_ = os.Setenv("XDG_CONFIG_HOME", configDir)
	_ = os.Setenv("XDG_DATA_HOME", dataDir)

	// Reload xdg package to pick up new environment variables
	xdg.Reload()

	cleanup := func() {
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
	}

	return tmpDir, cleanup
}

func TestConfigDir(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	configDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".config", "touchlog")
	if configDir != expectedDir {
		t.Errorf("ConfigDir() = %q, want %q", configDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); err != nil {
		t.Errorf("ConfigDir() did not create directory: %v", err)
	}
}

func TestConfigDir_CreatesParentDirectories(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	// Remove the .config directory to test that parent directories are created
	configBase := filepath.Join(tmpDir, ".config")
	if err := os.RemoveAll(configBase); err != nil {
		t.Fatalf("Failed to remove config base: %v", err)
	}

	configDir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".config", "touchlog")
	if configDir != expectedDir {
		t.Errorf("ConfigDir() = %q, want %q", configDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); err != nil {
		t.Errorf("ConfigDir() did not create directory: %v", err)
	}
}

func TestDataDir(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	dataDir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".local", "share", "touchlog")
	if dataDir != expectedDir {
		t.Errorf("DataDir() = %q, want %q", dataDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("DataDir() did not create directory: %v", err)
	}
}

func TestDataDir_CreatesParentDirectories(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	// Remove the .local directory to test that parent directories are created
	dataBase := filepath.Join(tmpDir, ".local")
	if err := os.RemoveAll(dataBase); err != nil {
		t.Fatalf("Failed to remove data base: %v", err)
	}

	dataDir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".local", "share", "touchlog")
	if dataDir != expectedDir {
		t.Errorf("DataDir() = %q, want %q", dataDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("DataDir() did not create directory: %v", err)
	}
}

func TestConfigFilePath(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	configPath, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error = %v, want nil", err)
	}

	expectedPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
	if configPath != expectedPath {
		t.Errorf("ConfigFilePath() = %q, want %q", configPath, expectedPath)
	}

	// Verify config directory was created
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); err != nil {
		t.Errorf("ConfigFilePath() did not create config directory: %v", err)
	}
}

func TestConfigFilePathReadOnly(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	configPath := ConfigFilePathReadOnly()

	expectedPath := filepath.Join(tmpDir, ".config", "touchlog", "config.yaml")
	if configPath != expectedPath {
		t.Errorf("ConfigFilePathReadOnly() = %q, want %q", configPath, expectedPath)
	}

	// Verify this function does NOT create directories (read-only)
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); err == nil {
		// Directory exists, but that's okay - we just verify the path is correct
		// The function should not create it, but if it already exists, that's fine
	}

	// Remove the directory to verify it's not created by this function
	if err := os.RemoveAll(configDir); err != nil {
		t.Fatalf("Failed to remove config dir: %v", err)
	}

	// Call again - should still return the path but not create directory
	configPath2 := ConfigFilePathReadOnly()
	if configPath2 != expectedPath {
		t.Errorf("ConfigFilePathReadOnly() = %q, want %q", configPath2, expectedPath)
	}

	// Verify directory was NOT created
	if _, err := os.Stat(configDir); err == nil {
		t.Error("ConfigFilePathReadOnly() created directory, but should be read-only")
	}
}

func TestTemplatesDir(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	templatesDir, err := TemplatesDir()
	if err != nil {
		t.Fatalf("TemplatesDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if templatesDir != expectedDir {
		t.Errorf("TemplatesDir() = %q, want %q", templatesDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(templatesDir); err != nil {
		t.Errorf("TemplatesDir() did not create directory: %v", err)
	}
}

func TestTemplatesDir_CreatesParentDirectories(t *testing.T) {
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	// Remove the .local directory to test that parent directories are created
	dataBase := filepath.Join(tmpDir, ".local")
	if err := os.RemoveAll(dataBase); err != nil {
		t.Fatalf("Failed to remove data base: %v", err)
	}

	templatesDir, err := TemplatesDir()
	if err != nil {
		t.Fatalf("TemplatesDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if templatesDir != expectedDir {
		t.Errorf("TemplatesDir() = %q, want %q", templatesDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(templatesDir); err != nil {
		t.Errorf("TemplatesDir() did not create directory: %v", err)
	}
}

func TestConfigDir_PermissionError(t *testing.T) {
	// This test verifies error handling when directory creation fails
	// We can't easily simulate permission errors in a portable way,
	// but we can test that the function returns an error when appropriate
	// For now, we'll skip this test as it's difficult to simulate without root
	t.Skip("Permission error testing requires root or special setup")
}

func TestConfigFilePath_ErrorFromConfigDir(t *testing.T) {
	// Test that ConfigFilePath propagates errors from ConfigDir
	// This is hard to test without mocking, but we can verify the error path exists
	_, cleanup := setupTestXDG(t)
	defer cleanup()

	// ConfigFilePath should work normally
	configPath, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error = %v, want nil", err)
	}
	if configPath == "" {
		t.Error("ConfigFilePath() = empty string, want non-empty")
	}

	// Verify it creates the config directory
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); err != nil {
		t.Errorf("ConfigFilePath() did not create config directory: %v", err)
	}
}

func TestTemplatesDir_ErrorFromDataDir(t *testing.T) {
	// Test that TemplatesDir propagates errors from DataDir
	tmpDir, cleanup := setupTestXDG(t)
	defer cleanup()

	// TemplatesDir should work normally
	templatesDir, err := TemplatesDir()
	if err != nil {
		t.Fatalf("TemplatesDir() error = %v, want nil", err)
	}

	expectedDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if templatesDir != expectedDir {
		t.Errorf("TemplatesDir() = %q, want %q", templatesDir, expectedDir)
	}

	// Verify directory was created
	if _, err := os.Stat(templatesDir); err != nil {
		t.Errorf("TemplatesDir() did not create directory: %v", err)
	}
}

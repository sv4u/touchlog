package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
)

func TestInit_CreatesValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	err := runInitWizard(tmpDir)
	if err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Verify config file exists
	configPath := filepath.Join(tmpDir, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}

	// Verify config can be loaded and is valid
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}

	if len(cfg.Types) == 0 {
		t.Error("expected at least one type in config")
	}

	// Verify all types have required fields
	for typeName, typeDef := range cfg.Types {
		if typeDef.DefaultState == "" {
			t.Errorf("type %q: default_state must be non-empty", typeName)
		}
		if typeDef.KeyPattern == nil {
			t.Errorf("type %q: key_pattern must be set", typeName)
		}
		if typeDef.KeyMaxLen <= 0 {
			t.Errorf("type %q: key_max_len must be positive", typeName)
		}
	}
}

func TestInit_RefusesToOverwriteExistingVault(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("failed to create .touchlog dir: %v", err)
	}

	configPath := filepath.Join(touchlogDir, "config.yaml")
	configContent := `version: 1
types: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Try to init again - should fail
	err := runInitWizard(tmpDir)
	if err == nil {
		t.Fatal("expected runInitWizard to fail when vault already exists")
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err := AtomicWrite(testFile, content)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	// Verify content
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("expected content %q, got %q", string(content), string(readContent))
	}
}

func TestAtomicWrite_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "test.txt")
	content := []byte("test content")

	err := AtomicWrite(testFile, content)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("file was not created: %v", err)
	}
}

func TestAtomicWrite_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	initialContent := []byte("initial")
	if err := os.WriteFile(testFile, initialContent, 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	// Overwrite with atomic write
	newContent := []byte("new content")
	err := AtomicWrite(testFile, newContent)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify new content
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(readContent) != string(newContent) {
		t.Errorf("expected content %q, got %q", string(newContent), string(readContent))
	}
}

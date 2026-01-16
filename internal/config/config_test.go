package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/internal/model"
)

func TestLoadConfig_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != model.ConfigSchemaVersion {
		t.Errorf("expected Version to be %d, got %d", model.ConfigSchemaVersion, cfg.Version)
	}
}

func TestLoadConfig_WithRepoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    description: "A note"
    default_state: "draft"
    key_pattern: "^[a-z0-9]+(-[a-z0-9]+)*$"
    key_max_len: 64
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type to be loaded")
	}

	if noteType.Description != "A note" {
		t.Errorf("expected Description to be 'A note', got %q", noteType.Description)
	}

	if noteType.DefaultState != "draft" {
		t.Errorf("expected DefaultState to be 'draft', got %q", noteType.DefaultState)
	}
}

func TestLoadConfig_RejectsUnknownKeys(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
unknown_key: "should fail"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with unknown key")
	}

	if err.Error() != "loading repo config: unknown top-level key: \"unknown_key\"" {
		t.Errorf("expected error about unknown key, got: %v", err)
	}
}

func TestLoadConfig_RejectsMissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `types:
  note:
    default_state: "draft"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with missing version")
	}

	if err.Error() != "loading repo config: missing required field: version" {
		t.Errorf("expected error about missing version, got: %v", err)
	}
}

func TestLoadConfig_RejectsUnsupportedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 999
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with unsupported version")
	}
}

func TestLoadConfig_TypeDefDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: "draft"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type to be loaded")
	}

	if noteType.KeyPattern == nil {
		t.Fatal("expected KeyPattern to have default value")
	}

	if !noteType.KeyPattern.MatchString("test-key") {
		t.Error("expected default KeyPattern to match 'test-key'")
	}

	if noteType.KeyMaxLen != DefaultKeyMaxLen {
		t.Errorf("expected KeyMaxLen to be %d, got %d", DefaultKeyMaxLen, noteType.KeyMaxLen)
	}
}

func TestLoadConfig_RejectsEmptyDefaultState(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: ""
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with empty default_state")
	}
}

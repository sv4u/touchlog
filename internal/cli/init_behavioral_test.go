package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestRunInitWizard_Behavior_CreatesConfig tests runInitWizard creates config file
func TestRunInitWizard_Behavior_CreatesConfig(t *testing.T) {
	tmpDir := t.TempDir()

	err := runInitWizard(tmpDir)
	if err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}

	// Verify config is valid
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if len(cfg.Types) == 0 {
		t.Error("expected config to have types defined")
	}
}

// TestRunInitWizard_Behavior_WithCustomBundle tests runInitWizard with custom bundle
func TestRunInitWizard_Behavior_WithCustomBundle(t *testing.T) {
	tmpDir := t.TempDir()

	// This test verifies the wizard can handle different bundle selections
	// Since the wizard uses default bundle in Phase 1, we test the default behavior
	err := runInitWizard(tmpDir)
	if err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Verify default bundle creates note type
	if cfg.Types[model.TypeName("note")].Description == "" {
		t.Error("expected note type to be configured")
	}
}

// TestGenerateConfig_Behavior_WithBundle tests generateConfig with different bundles
func TestGenerateConfig_Behavior_WithBundle(t *testing.T) {
	// Test with note bundle
	cfg, err := generateConfig([]string{"note"})
	if err != nil {
		t.Fatalf("generateConfig with note bundle failed: %v", err)
	}

	if len(cfg.Types) == 0 {
		t.Error("expected config to have types")
	}

	// Test with multiple bundles
	cfg, err = generateConfig([]string{"note", "task"})
	if err != nil {
		t.Fatalf("generateConfig with multiple bundles failed: %v", err)
	}

	if len(cfg.Types) < 2 {
		t.Errorf("expected at least 2 types, got %d", len(cfg.Types))
	}

	// Test unknown bundle (should return error)
	_, err = generateConfig([]string{"unknown-bundle"})
	if err == nil {
		t.Error("expected error for unknown bundle")
	}
}

// TestMarshalConfig_Behavior_WithComplexConfig tests marshalConfig with complex config
func TestMarshalConfig_Behavior_WithComplexConfig(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyPattern:   nil,
				KeyMaxLen:    64,
			},
			"task": {
				Description:  "A task",
				DefaultState: "todo",
				KeyPattern:   nil,
				KeyMaxLen:    64,
			},
		},
		Tags: config.TagConfig{
			Preferred: []string{"important", "project"},
		},
		Edges: map[model.EdgeType]config.EdgeDef{
			"related-to": {
				Description: "General relationship",
			},
			"depends-on": {
				Description: "Dependency relationship",
			},
		},
	}

	yaml, err := marshalConfig(cfg)
	if err != nil {
		t.Fatalf("marshalConfig failed: %v", err)
	}

	if len(yaml) == 0 {
		t.Error("expected non-empty YAML output")
	}

	// Verify YAML contains expected keys
	yamlStr := string(yaml)
	if !strings.Contains(yamlStr, "version:") {
		t.Error("expected YAML to contain version")
	}
	if !strings.Contains(yamlStr, "types:") {
		t.Error("expected YAML to contain types")
	}
	if !strings.Contains(yamlStr, "note:") {
		t.Error("expected YAML to contain note type")
	}
}

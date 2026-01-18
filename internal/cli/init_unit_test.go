package cli

import (
	"strings"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestGenerateConfig tests generateConfig function behavior
func TestGenerateConfig(t *testing.T) {
	selectedBundles := []string{"note", "task"}

	cfg, err := generateConfig(selectedBundles)
	if err != nil {
		t.Fatalf("generateConfig failed: %v", err)
	}

	if cfg.Version != model.ConfigSchemaVersion {
		t.Errorf("expected version %d, got %d", model.ConfigSchemaVersion, cfg.Version)
	}

	if len(cfg.Types) == 0 {
		t.Error("expected at least one type in config")
	}

	// Verify note type exists
	if _, ok := cfg.Types["note"]; !ok {
		t.Error("expected 'note' type to be in config")
	}

	// Verify default edge type exists
	if _, ok := cfg.Edges[model.DefaultEdgeType]; !ok {
		t.Error("expected default edge type to be in config")
	}
}

// TestGenerateConfig_UnknownBundle tests generateConfig with unknown bundle
func TestGenerateConfig_UnknownBundle(t *testing.T) {
	selectedBundles := []string{"unknown-bundle"}

	_, err := generateConfig(selectedBundles)
	if err == nil {
		t.Error("expected error for unknown bundle")
	}
}

// TestMarshalConfig tests marshalConfig function behavior
func TestMarshalConfig(t *testing.T) {
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
			},
		},
		Tags: config.TagConfig{
			Preferred: []string{},
		},
		Edges: map[model.EdgeType]config.EdgeDef{
			"related-to": {
				Description: "General relationship",
			},
		},
		Templates: config.TemplateConfig{
			Root: "templates",
		},
	}

	yamlBytes, err := marshalConfig(cfg)
	if err != nil {
		t.Fatalf("marshalConfig failed: %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Error("expected non-empty YAML output")
	}

	// Verify YAML contains expected keys
	yamlStr := string(yamlBytes)
	if !strings.Contains(yamlStr, "version") {
		t.Error("expected YAML to contain 'version'")
	}
	if !strings.Contains(yamlStr, "types") {
		t.Error("expected YAML to contain 'types'")
	}
	if !strings.Contains(yamlStr, "note") {
		t.Error("expected YAML to contain 'note' type")
	}
}

// TestMarshalConfig_WithAllowedFromTo tests marshalConfig with edge allowed_from/allowed_to
func TestMarshalConfig_WithAllowedFromTo(t *testing.T) {
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types:   make(map[model.TypeName]config.TypeDef),
		Tags: config.TagConfig{
			Preferred: []string{},
		},
		Edges: map[model.EdgeType]config.EdgeDef{
			"references": {
				Description: "Reference relationship",
				AllowedFrom: []model.TypeName{"note"},
				AllowedTo:   []model.TypeName{"note", "task"},
			},
		},
		Templates: config.TemplateConfig{
			Root: "templates",
		},
	}

	yamlBytes, err := marshalConfig(cfg)
	if err != nil {
		t.Fatalf("marshalConfig failed: %v", err)
	}

	yamlStr := string(yamlBytes)
	if !strings.Contains(yamlStr, "references") {
		t.Error("expected YAML to contain 'references' edge")
	}
}

package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

func TestExport_CreatesJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault and create index
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create minimal config
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyPattern:   nil,
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create a test note
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notePath := filepath.Join(noteDir, "test-note.Rmd")
	noteContent := `---
id: note-1
type: note
key: test-note
title: Test Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Test Note

This is a test note.
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing test note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Export index
	exportPath := filepath.Join(tmpDir, "export.json")
	if err := Export(tmpDir, exportPath); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}

	// Verify export is valid JSON
	exportData, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("reading export file: %v", err)
	}

	var export ExportData
	if err := json.Unmarshal(exportData, &export); err != nil {
		t.Fatalf("parsing export JSON: %v", err)
	}

	// Verify structure
	if export.Version != "1" {
		t.Errorf("expected version '1', got %q", export.Version)
	}

	if len(export.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(export.Nodes))
	}

	if export.Nodes[0].ID != "note-1" {
		t.Errorf("expected node ID 'note-1', got %q", export.Nodes[0].ID)
	}
}

func TestExport_DeterministicOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault and create index with multiple notes
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create minimal config
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyPattern:   nil,
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create multiple test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notes := []struct {
		key   string
		id    string
		title string
	}{
		{"zebra-note", "note-3", "Zebra Note"},
		{"alpha-note", "note-1", "Alpha Note"},
		{"beta-note", "note-2", "Beta Note"},
	}

	for _, n := range notes {
		notePath := filepath.Join(noteDir, n.key+".Rmd")
		noteContent := `---
id: ` + n.id + `
type: note
key: ` + n.key + `
title: ` + n.title + `
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# ` + n.title + `
`
		if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
			t.Fatalf("writing test note %s: %v", n.key, err)
		}
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Export index twice and verify they're identical
	exportPath1 := filepath.Join(tmpDir, "export1.json")
	exportPath2 := filepath.Join(tmpDir, "export2.json")

	if err := Export(tmpDir, exportPath1); err != nil {
		t.Fatalf("first Export failed: %v", err)
	}

	if err := Export(tmpDir, exportPath2); err != nil {
		t.Fatalf("second Export failed: %v", err)
	}

	// Read both exports
	export1, err := os.ReadFile(exportPath1)
	if err != nil {
		t.Fatalf("reading first export: %v", err)
	}

	export2, err := os.ReadFile(exportPath2)
	if err != nil {
		t.Fatalf("reading second export: %v", err)
	}

	// Verify they're identical
	if string(export1) != string(export2) {
		t.Error("exports are not deterministic (two exports differ)")
	}

	// Verify nodes are ordered by (type, key)
	var exportData ExportData
	if err := json.Unmarshal(export1, &exportData); err != nil {
		t.Fatalf("parsing export JSON: %v", err)
	}

	if len(exportData.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(exportData.Nodes))
	}

	// Verify ordering: alpha, beta, zebra
	expectedKeys := []string{"alpha-note", "beta-note", "zebra-note"}
	for i, node := range exportData.Nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("node %d: expected key %q, got %q", i, expectedKeys[i], node.Key)
		}
	}
}

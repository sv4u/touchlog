package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/model"
)

func TestExecuteSearch_NoFilters(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and index
	setupTestVault(t, tmpDir)

	// Execute search with no filters
	q := NewSearchQuery()
	results, err := ExecuteSearch(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteSearch failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestExecuteSearch_FilterByState(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and index
	setupTestVault(t, tmpDir)

	// Execute search with state filter
	q := NewSearchQuery()
	q.States = []string{"draft"}
	results, err := ExecuteSearch(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteSearch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify all results have state "draft"
	for _, result := range results {
		if result.State != "draft" {
			t.Errorf("expected state 'draft', got %q", result.State)
		}
	}
}

func TestExecuteSearch_FilterByType(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and index
	setupTestVault(t, tmpDir)

	// Execute search with type filter
	q := NewSearchQuery()
	q.Types = []string{"note"}
	results, err := ExecuteSearch(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteSearch failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify all results have type "note"
	for _, result := range results {
		if result.Type != "note" {
			t.Errorf("expected type 'note', got %q", result.Type)
		}
	}
}

func TestExecuteSearch_WithLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and index
	setupTestVault(t, tmpDir)

	// Execute search with limit
	q := NewSearchQuery()
	limit := 2
	q.Limit = &limit
	results, err := ExecuteSearch(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteSearch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestExecuteSearch_DeterministicOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and index
	setupTestVault(t, tmpDir)

	// Execute search twice and verify results are in the same order
	q := NewSearchQuery()

	results1, err := ExecuteSearch(tmpDir, q)
	if err != nil {
		t.Fatalf("first ExecuteSearch failed: %v", err)
	}

	results2, err := ExecuteSearch(tmpDir, q)
	if err != nil {
		t.Fatalf("second ExecuteSearch failed: %v", err)
	}

	if len(results1) != len(results2) {
		t.Fatalf("result counts differ: %d vs %d", len(results1), len(results2))
	}

	// Verify ordering is deterministic
	for i := range results1 {
		if results1[i].ID != results2[i].ID {
			t.Errorf("result %d: IDs differ (%q vs %q)", i, results1[i].ID, results2[i].ID)
		}
		if results1[i].Key != results2[i].Key {
			t.Errorf("result %d: Keys differ (%q vs %q)", i, results1[i].Key, results2[i].Key)
		}
	}
}

// setupTestVault creates a test vault with notes and builds the index
func setupTestVault(t *testing.T, tmpDir string) {
	t.Helper()

	// Create .touchlog directory
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create config
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

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notes := []struct {
		key   string
		id    string
		title string
		state string
	}{
		{"alpha-note", "note-1", "Alpha Note", "draft"},
		{"beta-note", "note-2", "Beta Note", "published"},
		{"gamma-note", "note-3", "Gamma Note", "draft"},
	}

	for _, n := range notes {
		notePath := filepath.Join(noteDir, n.key+".Rmd")
		noteContent := `---
id: ` + n.id + `
type: note
key: ` + n.key + `
title: ` + n.title + `
state: ` + n.state + `
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
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}
}

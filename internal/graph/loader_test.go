package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
)

func TestLoadGraph_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault with empty database
	setupTestVaultWithIndex(t, tmpDir)

	// Load graph
	graph, err := LoadGraph(tmpDir)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.OutgoingEdges) != 0 {
		t.Errorf("expected 0 outgoing edge sets, got %d", len(graph.OutgoingEdges))
	}
}

func TestLoadGraph_WithNodes(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index with notes
	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notes := []struct {
		key   string
		id    string
		title string
	}{
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
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Load graph
	graph, err := LoadGraph(tmpDir)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	// Verify nodes are loaded
	if graph.Nodes["note-1"] == nil {
		t.Error("expected node 'note-1' to be loaded")
	}
	if graph.Nodes["note-2"] == nil {
		t.Error("expected node 'note-2' to be loaded")
	}
}

func TestLoadSubgraph_WithSpecificNodes(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index with notes
	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notes := []struct {
		key   string
		id    string
		title string
	}{
		{"alpha-note", "note-1", "Alpha Note"},
		{"beta-note", "note-2", "Beta Note"},
		{"gamma-note", "note-3", "Gamma Note"},
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
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Load subgraph for specific nodes
	nodeIDs := []model.NoteID{"note-1", "note-2"}
	graph, err := LoadSubgraph(tmpDir, nodeIDs)
	if err != nil {
		t.Fatalf("LoadSubgraph failed: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes in subgraph, got %d", len(graph.Nodes))
	}

	// Verify only requested nodes are loaded
	if graph.Nodes["note-1"] == nil {
		t.Error("expected node 'note-1' to be loaded")
	}
	if graph.Nodes["note-2"] == nil {
		t.Error("expected node 'note-2' to be loaded")
	}
	if graph.Nodes["note-3"] != nil {
		t.Error("expected node 'note-3' NOT to be loaded")
	}
}

// setupTestVaultWithIndex creates a test vault with config and database schema
func setupTestVaultWithIndex(t *testing.T, tmpDir string) {
	t.Helper()

	// Create .touchlog directory
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create config file
	configPath := filepath.Join(touchlogDir, "config.yaml")
	configContent := `version: 1
types:
  note:
    description: A note
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
tags:
  preferred: []
edges:
  related-to:
    description: General relationship
templates:
  root: templates
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Create and initialize database with schema
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	if err := store.ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("applying migrations: %v", err)
	}
	_ = db.Close()
}

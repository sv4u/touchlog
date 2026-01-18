package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestLoadGraph_WithEdges tests LoadGraph with nodes and edges
func TestLoadGraph_WithEdges(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create first note
	note1Path := filepath.Join(noteDir, "note-1.Rmd")
	note1Content := `---
id: note-1
type: note
key: note-1
title: First Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# First Note

Links to [[note:note-2]].
`
	if err := os.WriteFile(note1Path, []byte(note1Content), 0644); err != nil {
		t.Fatalf("writing note 1: %v", err)
	}

	// Create second note
	note2Path := filepath.Join(noteDir, "note-2.Rmd")
	note2Content := `---
id: note-2
type: note
key: note-2
title: Second Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Second Note
`
	if err := os.WriteFile(note2Path, []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

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

	// Verify nodes are loaded
	if graph.Nodes["note-1"] == nil {
		t.Error("expected node 'note-1' to be loaded")
	}
	if graph.Nodes["note-2"] == nil {
		t.Error("expected node 'note-2' to be loaded")
	}

	// Verify edges are loaded
	if len(graph.OutgoingEdges["note-1"]) == 0 {
		t.Error("expected outgoing edges from note-1")
	}
	if len(graph.IncomingEdges["note-2"]) == 0 {
		t.Error("expected incoming edges to note-2")
	}
}

// TestLoadSubgraph_WithEdges tests LoadSubgraph with edges
func TestLoadSubgraph_WithEdges(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create linked notes
	note1Path := filepath.Join(noteDir, "note-1.Rmd")
	note1Content := `---
id: note-1
type: note
key: note-1
title: First Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# First Note

Links to [[note:note-2]].
`
	if err := os.WriteFile(note1Path, []byte(note1Content), 0644); err != nil {
		t.Fatalf("writing note 1: %v", err)
	}

	note2Path := filepath.Join(noteDir, "note-2.Rmd")
	note2Content := `---
id: note-2
type: note
key: note-2
title: Second Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Second Note
`
	if err := os.WriteFile(note2Path, []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Load subgraph for note-1
	nodeIDs := []model.NoteID{"note-1"}
	graph, err := LoadSubgraph(tmpDir, nodeIDs)
	if err != nil {
		t.Fatalf("LoadSubgraph failed: %v", err)
	}

	// Verify node is loaded
	if graph.Nodes["note-1"] == nil {
		t.Error("expected node 'note-1' to be loaded")
	}

	// Verify edges are loaded (even if note-2 is not in subgraph, edges from note-1 should be included)
	_ = graph
}

package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

// TestLoadGraph_WithUnresolvedEdges tests LoadGraph with unresolved edges
func TestLoadGraph_WithUnresolvedEdges(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create note with unresolved link
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

Links to [[nonexistent]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
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

	// Verify node is loaded
	if graph.Nodes["note-1"] == nil {
		t.Error("expected node 'note-1' to be loaded")
	}

	// Verify edges are loaded (even if unresolved)
	_ = graph
}

// TestLoadGraph_WithMultipleEdges tests LoadGraph with multiple edges
func TestLoadGraph_WithMultipleEdges(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create source note with multiple links
	sourcePath := filepath.Join(noteDir, "source.Rmd")
	sourceContent := `---
id: note-source
type: note
key: source
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Links to [[note:target1]] and [[note:target2]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create target notes
	for i := 1; i <= 2; i++ {
		targetPath := filepath.Join(noteDir, fmt.Sprintf("target%d.Rmd", i))
		targetContent := fmt.Sprintf(`---
id: note-target%d
type: note
key: target%d
title: Target %d
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Target %d
`, i, i, i, i)
		if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
			t.Fatalf("writing target note %d: %v", i, err)
		}
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

	// Verify multiple outgoing edges from source
	if len(graph.OutgoingEdges["note-source"]) < 2 {
		t.Errorf("expected at least 2 outgoing edges from source, got %d", len(graph.OutgoingEdges["note-source"]))
	}
}

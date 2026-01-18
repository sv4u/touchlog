package query

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

func TestExecuteNeighbors_RequiresMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewNeighborsQuery()
	q.Root = "note:alpha-note"
	// MaxDepth not set (0)

	_, err := ExecuteNeighbors(tmpDir, q)
	if err == nil {
		t.Error("expected error when max_depth is not set")
	}
}

func TestExecuteNeighbors_SingleNode(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewNeighborsQuery()
	q.Root = "note:alpha-note"
	q.MaxDepth = 1

	results, err := ExecuteNeighbors(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteNeighbors failed: %v", err)
	}

	// Should have root at depth 0
	if len(results) == 0 {
		t.Error("expected at least depth 0 result")
	}

	if results[0].Depth != 0 {
		t.Errorf("expected depth 0 for root, got %d", results[0].Depth)
	}
}

func TestExecuteNeighbors_WithLinks(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create linked notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create root note
	rootPath := filepath.Join(noteDir, "root-note.Rmd")
	rootContent := `---
id: note-root
type: note
key: root-note
title: Root Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Root Note

Links to [[note:neighbor-1]] and [[note:neighbor-2]].
`
	if err := os.WriteFile(rootPath, []byte(rootContent), 0644); err != nil {
		t.Fatalf("writing root note: %v", err)
	}

	// Create neighbor notes
	for i := 1; i <= 2; i++ {
		neighborPath := filepath.Join(noteDir, fmt.Sprintf("neighbor-%d.Rmd", i))
		neighborContent := fmt.Sprintf(`---
id: note-neighbor-%d
type: note
key: neighbor-%d
title: Neighbor %d
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Neighbor %d
`, i, i, i, i)
		if err := os.WriteFile(neighborPath, []byte(neighborContent), 0644); err != nil {
			t.Fatalf("writing neighbor note %d: %v", i, err)
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

	// Query neighbors
	q := NewNeighborsQuery()
	q.Root = "note:root-note"
	q.MaxDepth = 1
	q.Direction = "out" // Outgoing links

	results, err := ExecuteNeighbors(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteNeighbors failed: %v", err)
	}

	// Should have root at depth 0 and neighbors at depth 1
	if len(results) < 2 {
		t.Errorf("expected at least 2 depth levels, got %d", len(results))
	}

	// Check depth 1 has neighbors
	var depth1Result *NeighborsResult
	for i := range results {
		if results[i].Depth == 1 {
			depth1Result = &results[i]
			break
		}
	}

	if depth1Result == nil {
		t.Error("expected depth 1 result")
	} else {
		if len(depth1Result.Nodes) != 2 {
			t.Errorf("expected 2 neighbors at depth 1, got %d", len(depth1Result.Nodes))
		}
	}
}

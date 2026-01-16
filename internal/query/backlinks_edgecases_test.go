package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
)

func TestExecuteBacklinks_WithCycles(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes with cycles: A -> B -> A
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Note A links to B
	noteAPath := filepath.Join(noteDir, "note-a.Rmd")
	noteAContent := `---
id: note-a
type: note
key: note-a
title: Note A
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note A

Links to [[note:note-b]].
`
	if err := os.WriteFile(noteAPath, []byte(noteAContent), 0644); err != nil {
		t.Fatalf("writing note A: %v", err)
	}

	// Note B links to A
	noteBPath := filepath.Join(noteDir, "note-b.Rmd")
	noteBContent := `---
id: note-b
type: note
key: note-b
title: Note B
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note B

Links to [[note:note-a]].
`
	if err := os.WriteFile(noteBPath, []byte(noteBContent), 0644); err != nil {
		t.Fatalf("writing note B: %v", err)
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

	// Query backlinks for A (should find B)
	q := NewBacklinksQuery()
	q.Target = "note:note-a"
	q.Direction = "in"

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 backlink, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].HopCount != 1 {
			t.Errorf("expected hop_count 1, got %d", results[0].HopCount)
		}
	}
}

func TestExecuteBacklinks_DirectionBoth(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes: A -> B (A links to B)
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Note A links to B
	noteAPath := filepath.Join(noteDir, "note-a.Rmd")
	noteAContent := `---
id: note-a
type: note
key: note-a
title: Note A
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note A

Links to [[note:note-b]].
`
	if err := os.WriteFile(noteAPath, []byte(noteAContent), 0644); err != nil {
		t.Fatalf("writing note A: %v", err)
	}

	// Note B (no links)
	noteBPath := filepath.Join(noteDir, "note-b.Rmd")
	noteBContent := `---
id: note-b
type: note
key: note-b
title: Note B
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note B
`
	if err := os.WriteFile(noteBPath, []byte(noteBContent), 0644); err != nil {
		t.Fatalf("writing note B: %v", err)
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

	// Query backlinks for B with direction "both"
	// Should find A (incoming) and nothing outgoing
	q := NewBacklinksQuery()
	q.Target = "note:note-b"
	q.Direction = "both"

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 backlink (incoming from A), got %d", len(results))
	}
}

func TestExecuteBacklinks_WithSelfLoop(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create note that links to itself
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notePath := filepath.Join(noteDir, "self-loop.Rmd")
	noteContent := `---
id: note-self
type: note
key: self-loop
title: Self Loop
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Self Loop

Links to [[note:self-loop]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
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

	// Query backlinks for self-loop with direction "both"
	q := NewBacklinksQuery()
	q.Target = "note:self-loop"
	q.Direction = "both"

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	// Should find itself as a backlink (self-loop)
	if len(results) != 1 {
		t.Errorf("expected 1 backlink (self-loop), got %d", len(results))
	}
}

func TestExecuteBacklinks_InvalidDirection(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewBacklinksQuery()
	q.Target = "note:alpha-note"
	q.Direction = "invalid"

	_, err := ExecuteBacklinks(tmpDir, q)
	if err == nil {
		t.Error("expected error for invalid direction")
	}
}

func TestExecuteNeighbors_WithCycles(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes with cycles: A -> B -> A
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Note A links to B
	noteAPath := filepath.Join(noteDir, "note-a.Rmd")
	noteAContent := `---
id: note-a
type: note
key: note-a
title: Note A
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note A

Links to [[note:note-b]].
`
	if err := os.WriteFile(noteAPath, []byte(noteAContent), 0644); err != nil {
		t.Fatalf("writing note A: %v", err)
	}

	// Note B links to A
	noteBPath := filepath.Join(noteDir, "note-b.Rmd")
	noteBContent := `---
id: note-b
type: note
key: note-b
title: Note B
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note B

Links to [[note:note-a]].
`
	if err := os.WriteFile(noteBPath, []byte(noteBContent), 0644); err != nil {
		t.Fatalf("writing note B: %v", err)
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

	// Query neighbors from A with depth 2
	// Should find B at depth 1, and A again at depth 2 (but visited set should prevent infinite loop)
	q := NewNeighborsQuery()
	q.Root = "note:note-a"
	q.MaxDepth = 2
	q.Direction = "out"

	results, err := ExecuteNeighbors(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteNeighbors failed: %v", err)
	}

	// Should have depth 0 (root) and depth 1 (B)
	// Depth 2 should not include A again due to visited set
	if len(results) < 2 {
		t.Errorf("expected at least 2 depth levels, got %d", len(results))
	}

	// Check that we don't have duplicate nodes at different depths
	nodeCounts := make(map[string]int)
	for _, result := range results {
		for _, node := range result.Nodes {
			nodeCounts[string(node.ID)]++
		}
	}

	// Each node should appear at most once (visited set prevents revisiting)
	for nodeID, count := range nodeCounts {
		if count > 1 {
			t.Errorf("node %s appears %d times (visited set should prevent this)", nodeID, count)
		}
	}
}

func TestExecutePaths_NoPathExists(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	// Query paths between two unconnected notes
	q := NewPathsQuery()
	q.Source = "note:alpha-note"
	q.Destinations = []string{"note:beta-note"}
	q.MaxDepth = 5

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should return empty results (no path exists)
	if len(results) != 0 {
		t.Errorf("expected 0 paths for unconnected nodes, got %d", len(results))
	}
}

func TestExecutePaths_MultipleDestinations(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes: A -> B, A -> C
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Note A links to B and C
	noteAPath := filepath.Join(noteDir, "note-a.Rmd")
	noteAContent := `---
id: note-a
type: note
key: note-a
title: Note A
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note A

Links to [[note:note-b]] and [[note:note-c]].
`
	if err := os.WriteFile(noteAPath, []byte(noteAContent), 0644); err != nil {
		t.Fatalf("writing note A: %v", err)
	}

	// Note B
	noteBPath := filepath.Join(noteDir, "note-b.Rmd")
	noteBContent := `---
id: note-b
type: note
key: note-b
title: Note B
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note B
`
	if err := os.WriteFile(noteBPath, []byte(noteBContent), 0644); err != nil {
		t.Fatalf("writing note B: %v", err)
	}

	// Note C
	noteCPath := filepath.Join(noteDir, "note-c.Rmd")
	noteCContent := `---
id: note-c
type: note
key: note-c
title: Note C
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note C
`
	if err := os.WriteFile(noteCPath, []byte(noteCContent), 0644); err != nil {
		t.Fatalf("writing note C: %v", err)
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

	// Query paths from A to both B and C
	q := NewPathsQuery()
	q.Source = "note:note-a"
	q.Destinations = []string{"note:note-b", "note:note-c"}
	q.MaxDepth = 5

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find paths to both destinations
	if len(results) != 2 {
		t.Errorf("expected 2 paths (one to B, one to C), got %d", len(results))
	}

	// Verify both destinations are present
	destinations := make(map[string]bool)
	for _, result := range results {
		destinations[result.Destination] = true
	}

	if !destinations["note:note-b"] {
		t.Error("expected path to note:note-b")
	}
	if !destinations["note:note-c"] {
		t.Error("expected path to note:note-c")
	}
}

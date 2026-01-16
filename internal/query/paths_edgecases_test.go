package query

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
)

func TestExecutePaths_WithCycles(t *testing.T) {
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

	// Query paths from A to B
	// Should find path A -> B (1 hop), not A -> B -> A -> B (3 hops)
	q := NewPathsQuery()
	q.Source = "note:note-a"
	q.Destinations = []string{"note:note-b"}
	q.MaxDepth = 5
	q.MaxPaths = 10

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find at least one path
	if len(results) == 0 {
		t.Error("expected at least 1 path from A to B")
	}

	// Check that we found the shortest path (1 hop)
	if len(results) > 0 {
		shortestHopCount := results[0].HopCount
		if shortestHopCount != 1 {
			t.Errorf("expected shortest path to be 1 hop, got %d", shortestHopCount)
		}

		// Verify no path has a cycle (same node appears twice)
		for _, result := range results {
			nodeIDs := make(map[string]int)
			for _, node := range result.Nodes {
				nodeIDs[string(node.ID)]++
				if nodeIDs[string(node.ID)] > 1 {
					t.Errorf("path contains cycle: node %s appears %d times", node.ID, nodeIDs[string(node.ID)])
				}
			}
		}
	}
}

func TestExecutePaths_MaxPathsLimit(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes: A -> B, A -> C, A -> D (multiple paths from A)
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Note A links to B, C, D
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

Links to [[note:note-b]], [[note:note-c]], and [[note:note-d]].
`
	if err := os.WriteFile(noteAPath, []byte(noteAContent), 0644); err != nil {
		t.Fatalf("writing note A: %v", err)
	}

	// Create B, C, D
	for _, key := range []string{"b", "c", "d"} {
		notePath := filepath.Join(noteDir, fmt.Sprintf("note-%s.Rmd", key))
		noteContent := fmt.Sprintf(`---
id: note-%s
type: note
key: note-%s
title: Note %s
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note %s
`, key, key, key, key)
		if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
			t.Fatalf("writing note %s: %v", key, err)
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

	// Query paths from A to B with max_paths = 2
	// Should find at most 2 paths
	q := NewPathsQuery()
	q.Source = "note:note-a"
	q.Destinations = []string{"note:note-b"}
	q.MaxDepth = 5
	q.MaxPaths = 2

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find at most maxPaths paths
	if len(results) > q.MaxPaths {
		t.Errorf("expected at most %d paths, got %d", q.MaxPaths, len(results))
	}
}

func TestExecutePaths_ExceedsMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes: A -> B -> C -> D (chain of 3 hops)
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

	// Note B links to C
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

Links to [[note:note-c]].
`
	if err := os.WriteFile(noteBPath, []byte(noteBContent), 0644); err != nil {
		t.Fatalf("writing note B: %v", err)
	}

	// Note C links to D
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

Links to [[note:note-d]].
`
	if err := os.WriteFile(noteCPath, []byte(noteCContent), 0644); err != nil {
		t.Fatalf("writing note C: %v", err)
	}

	// Note D (destination)
	noteDPath := filepath.Join(noteDir, "note-d.Rmd")
	noteDContent := `---
id: note-d
type: note
key: note-d
title: Note D
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Note D
`
	if err := os.WriteFile(noteDPath, []byte(noteDContent), 0644); err != nil {
		t.Fatalf("writing note D: %v", err)
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

	// Query paths from A to D with max_depth = 2
	// Path requires 3 hops, so should not be found
	q := NewPathsQuery()
	q.Source = "note:note-a"
	q.Destinations = []string{"note:note-d"}
	q.MaxDepth = 2 // Too small for 3-hop path
	q.MaxPaths = 10

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should not find path (exceeds max_depth)
	if len(results) != 0 {
		t.Errorf("expected 0 paths (exceeds max_depth), got %d", len(results))
	}

	// Now try with max_depth = 3 (should find path)
	q.MaxDepth = 3
	results, err = ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find path
	if len(results) == 0 {
		t.Error("expected path when max_depth is sufficient")
	}

	if len(results) > 0 {
		if results[0].HopCount != 3 {
			t.Errorf("expected hop_count 3, got %d", results[0].HopCount)
		}
	}
}

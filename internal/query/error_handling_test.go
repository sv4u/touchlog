package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

func TestResolveNodeID_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Test invalid format (too many colons)
	_, err := resolveNodeID(tmpDir, "type:key:extra")
	if err == nil {
		t.Error("expected error for invalid node identifier format")
	}
}

func TestExecuteBacklinks_NodeNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewBacklinksQuery()
	q.Target = "note:nonexistent"

	_, err := ExecuteBacklinks(tmpDir, q)
	if err == nil {
		t.Error("expected error when target node not found")
	}
}

func TestExecuteNeighbors_NodeNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewNeighborsQuery()
	q.Root = "note:nonexistent"
	q.MaxDepth = 1

	_, err := ExecuteNeighbors(tmpDir, q)
	if err == nil {
		t.Error("expected error when root node not found")
	}
}

func TestExecutePaths_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewPathsQuery()
	q.Source = "note:nonexistent"
	q.Destinations = []string{"note:alpha-note"}
	q.MaxDepth = 5

	_, err := ExecutePaths(tmpDir, q)
	if err == nil {
		t.Error("expected error when source node not found")
	}
}

func TestExecutePaths_DestinationNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewPathsQuery()
	q.Source = "note:alpha-note"
	q.Destinations = []string{"note:nonexistent"}
	q.MaxDepth = 5

	_, err := ExecutePaths(tmpDir, q)
	if err == nil {
		t.Error("expected error when destination node not found")
	}
}

func TestExecuteBacklinks_EmptyGraph(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create a note but don't build index (empty database)
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

	// Query backlinks (should work even with no links)
	q := NewBacklinksQuery()
	q.Target = "note:test-note"

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	// Should return empty results, not error
	if len(results) != 0 {
		t.Errorf("expected 0 backlinks, got %d", len(results))
	}
}

func TestExecuteNeighbors_MaxDepthZero(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewNeighborsQuery()
	q.Root = "note:alpha-note"
	q.MaxDepth = 0 // Invalid - must be > 0

	_, err := ExecuteNeighbors(tmpDir, q)
	if err == nil {
		t.Error("expected error when max_depth is 0")
	}
}

func TestExecutePaths_MaxDepthZero(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewPathsQuery()
	q.Source = "note:alpha-note"
	q.Destinations = []string{"note:beta-note"}
	q.MaxDepth = 0 // Invalid - must be > 0

	_, err := ExecutePaths(tmpDir, q)
	if err == nil {
		t.Error("expected error when max_depth is 0")
	}
}

func TestExecutePaths_MaxPathsZero(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create linked notes
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

	// Build index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query paths with max_paths = 0
	q := NewPathsQuery()
	q.Source = "note:note-a"
	q.Destinations = []string{"note:note-b"}
	q.MaxDepth = 5
	q.MaxPaths = 0 // No paths allowed

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should return empty (max_paths 0)
	if len(results) != 0 {
		t.Errorf("expected 0 paths (max_paths 0), got %d", len(results))
	}
}

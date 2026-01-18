package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

func TestExecutePaths_RequiresMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewPathsQuery()
	q.Source = "note:alpha-note"
	q.Destinations = []string{"note:beta-note"}
	// MaxDepth not set (0)

	_, err := ExecutePaths(tmpDir, q)
	if err == nil {
		t.Error("expected error when max_depth is not set")
	}
}

func TestExecutePaths_SourceEqualsDestination(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithNotes(t, tmpDir)

	q := NewPathsQuery()
	q.Source = "note:alpha-note"
	q.Destinations = []string{"note:alpha-note"} // Same as source
	q.MaxDepth = 5

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should return zero-hop path immediately
	if len(results) != 1 {
		t.Errorf("expected 1 zero-hop path, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].HopCount != 0 {
			t.Errorf("expected hop_count 0 for src==dst, got %d", results[0].HopCount)
		}
	}
}

func TestExecutePaths_WithLinks(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create linked notes: source -> intermediate -> destination
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create source note
	sourcePath := filepath.Join(noteDir, "source-note.Rmd")
	sourceContent := `---
id: note-source
type: note
key: source-note
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Links to [[note:intermediate-note]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create intermediate note
	intermediatePath := filepath.Join(noteDir, "intermediate-note.Rmd")
	intermediateContent := `---
id: note-intermediate
type: note
key: intermediate-note
title: Intermediate Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Intermediate Note

Links to [[note:destination-note]].
`
	if err := os.WriteFile(intermediatePath, []byte(intermediateContent), 0644); err != nil {
		t.Fatalf("writing intermediate note: %v", err)
	}

	// Create destination note
	destPath := filepath.Join(noteDir, "destination-note.Rmd")
	destContent := `---
id: note-destination
type: note
key: destination-note
title: Destination Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Destination Note
`
	if err := os.WriteFile(destPath, []byte(destContent), 0644); err != nil {
		t.Fatalf("writing destination note: %v", err)
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

	// Query paths
	q := NewPathsQuery()
	q.Source = "note:source-note"
	q.Destinations = []string{"note:destination-note"}
	q.MaxDepth = 5
	q.Direction = "out"

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find a path: source -> intermediate -> destination (2 hops)
	if len(results) == 0 {
		t.Error("expected at least 1 path")
	}

	if len(results) > 0 {
		if results[0].HopCount != 2 {
			t.Errorf("expected hop_count 2, got %d", results[0].HopCount)
		}
		if len(results[0].Nodes) != 3 {
			t.Errorf("expected 3 nodes in path, got %d", len(results[0].Nodes))
		}
	}
}

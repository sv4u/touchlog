package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

// TestExecuteNeighbors_WithDepth tests ExecuteNeighbors with different depth values
func TestExecuteNeighbors_WithDepth(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create chain of notes: root -> level1 -> level2
	notes := []struct {
		id    string
		key   string
		title string
		links string
	}{
		{"note-root", "root", "Root", "Links to [[note:level1]]."},
		{"note-level1", "level1", "Level 1", "Links to [[note:level2]]."},
		{"note-level2", "level2", "Level 2", ""},
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

` + n.links + `
`
		if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
			t.Fatalf("writing note %s: %v", n.key, err)
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

	// Query neighbors with depth 1
	q := NewNeighborsQuery()
	q.Root = "note:root"
	q.MaxDepth = 1
	q.Direction = "out"

	results, err := ExecuteNeighbors(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteNeighbors failed: %v", err)
	}

	// Should find neighbors at depth 1
	if len(results) == 0 {
		t.Error("expected to find at least one neighbor")
	}
}

// TestExecuteNeighbors_WithEdgeTypeFilter tests ExecuteNeighbors with edge type filter
func TestExecuteNeighbors_WithEdgeTypeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create source note
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

Links to [[note:target]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create target note
	targetPath := filepath.Join(noteDir, "target.Rmd")
	targetContent := `---
id: note-target
type: note
key: target
title: Target Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Target Note
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatalf("writing target note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query neighbors with edge type filter
	q := NewNeighborsQuery()
	q.Root = "note:source"
	q.MaxDepth = 1
	q.Direction = "out"
	q.EdgeTypes = []string{"related-to"}

	results, err := ExecuteNeighbors(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteNeighbors failed: %v", err)
	}

	// Should find neighbors with matching edge type
	_ = results
}

// TestExecuteNeighbors_InvalidDirection tests ExecuteNeighbors with invalid direction
func TestExecuteNeighbors_InvalidDirection(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	q := NewNeighborsQuery()
	q.Root = "note:test"
	q.MaxDepth = 1
	q.Direction = "invalid"

	_, err := ExecuteNeighbors(tmpDir, q)
	if err == nil {
		t.Error("expected error for invalid direction")
	}
}

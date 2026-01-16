package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
)

func TestExportDOT_WithEmptyGraph(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Export empty graph
	outputPath := filepath.Join(tmpDir, "graph.dot")
	opts := ExportOptions{
		Depth: 10,
	}

	if err := ExportDOT(tmpDir, outputPath, opts); err != nil {
		t.Fatalf("ExportDOT failed: %v", err)
	}

	// Verify file exists and contains basic structure
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading DOT file: %v", err)
	}

	dotContent := string(content)
	if !strings.Contains(dotContent, "digraph touchlog") {
		t.Error("DOT file should contain 'digraph touchlog'")
	}
}

func TestExportDOT_WithEdges(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes with links
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

	// Export graph
	outputPath := filepath.Join(tmpDir, "graph.dot")
	opts := ExportOptions{
		Depth: 10,
	}

	if err := ExportDOT(tmpDir, outputPath, opts); err != nil {
		t.Fatalf("ExportDOT failed: %v", err)
	}

	// Verify file contains edge
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading DOT file: %v", err)
	}

	dotContent := string(content)
	if !strings.Contains(dotContent, "note-a") {
		t.Error("DOT file should contain 'note-a'")
	}
	if !strings.Contains(dotContent, "note-b") {
		t.Error("DOT file should contain 'note-b'")
	}
	if !strings.Contains(dotContent, "->") {
		t.Error("DOT file should contain edge (->)")
	}
}

func TestExportDOT_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create notes with different states
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Note A (draft)
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
`
	if err := os.WriteFile(noteAPath, []byte(noteAContent), 0644); err != nil {
		t.Fatalf("writing note A: %v", err)
	}

	// Note B (published)
	noteBPath := filepath.Join(noteDir, "note-b.Rmd")
	noteBContent := `---
id: note-b
type: note
key: note-b
title: Note B
state: published
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

	// Export graph with state filter (only published)
	outputPath := filepath.Join(tmpDir, "graph.dot")
	opts := ExportOptions{
		Depth:  10,
		States: []string{"published"},
	}

	if err := ExportDOT(tmpDir, outputPath, opts); err != nil {
		t.Fatalf("ExportDOT failed: %v", err)
	}

	// Verify file contains only published note
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading DOT file: %v", err)
	}

	dotContent := string(content)
	if strings.Contains(dotContent, "note-a") {
		t.Error("DOT file should not contain 'note-a' (draft state filtered out)")
	}
	if !strings.Contains(dotContent, "note-b") {
		t.Error("DOT file should contain 'note-b' (published state)")
	}
}

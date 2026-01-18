package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/graph"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/query"
)

// TestIntegration_QueryBacklinks tests the full workflow for querying backlinks
func TestIntegration_QueryBacklinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create note directory
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

This links to [[note:note-2]].
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

This links back to [[note:note-1]].
`
	if err := os.WriteFile(note2Path, []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

	// Build index
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query backlinks for note-2
	q := query.NewBacklinksQuery()
	q.Target = "note:note-2"
	q.Direction = "in"
	q.Format = "table"

	results, err := query.ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("executing backlinks query: %v", err)
	}

	// Should find at least one backlink (from note-1)
	if len(results) == 0 {
		t.Error("expected to find at least one backlink")
	}
}

// TestIntegration_QueryNeighbors tests the full workflow for querying neighbors
func TestIntegration_QueryNeighbors(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create note directory
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

This links to [[note:note-2]].
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

This links back to [[note:note-1]].
`
	if err := os.WriteFile(note2Path, []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

	// Build index
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query neighbors for note-1
	q := query.NewNeighborsQuery()
	q.Root = "note:note-1"
	q.MaxDepth = 1
	q.Direction = "both"
	q.Format = "table"

	results, err := query.ExecuteNeighbors(tmpDir, q)
	if err != nil {
		t.Fatalf("executing neighbors query: %v", err)
	}

	// Should find at least one neighbor (note-2)
	if len(results) == 0 {
		t.Error("expected to find at least one neighbor")
	}
}

// TestIntegration_QueryPaths tests the full workflow for querying paths
func TestIntegration_QueryPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create note directory
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

This links to [[note:note-2]].
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

This links back to [[note:note-1]].
`
	if err := os.WriteFile(note2Path, []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

	// Build index
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query paths from note-1 to note-2
	q := query.NewPathsQuery()
	q.Source = "note:note-1"
	q.Destinations = []string{"note:note-2"}
	q.MaxDepth = 5
	q.MaxPaths = 10
	q.Direction = "both"
	q.Format = "table"

	results, err := query.ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("executing paths query: %v", err)
	}

	// Should find at least one path
	if len(results) == 0 {
		t.Error("expected to find at least one path")
	}
}

// TestIntegration_QuerySearch_WithFilters tests search query with various filters
func TestIntegration_QuerySearch_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create a note
	notePath := filepath.Join(noteDir, "test-note.Rmd")
	noteContent := `---
id: note-test
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
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query search with type filter
	searchQuery := query.NewSearchQuery()
	searchQuery.Types = []string{"note"}
	searchQuery.States = []string{"draft"}

	results, err := query.ExecuteSearch(tmpDir, searchQuery)
	if err != nil {
		t.Fatalf("executing search query: %v", err)
	}

	// Should find at least one note
	if len(results) == 0 {
		t.Error("expected to find at least one note in search results")
	}
}

// TestIntegration_IndexExport tests index export functionality
func TestIntegration_IndexExport(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create a note
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notePath := filepath.Join(noteDir, "test-note.Rmd")
	noteContent := `---
id: note-test
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
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Export index
	exportPath := filepath.Join(tmpDir, "index-export.json")
	if err := index.Export(tmpDir, exportPath); err != nil {
		t.Fatalf("exporting index: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}
}

// TestIntegration_GraphExport tests graph export functionality
func TestIntegration_GraphExport(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create a note
	notePath := filepath.Join(noteDir, "test-note.Rmd")
	noteContent := `---
id: note-test
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
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Export graph
	exportPath := filepath.Join(tmpDir, "graph.dot")
	opts := graph.ExportOptions{
		Depth: 10,
		Force: true,
	}

	if err := graph.ExportDOT(tmpDir, exportPath, opts); err != nil {
		t.Fatalf("exporting graph: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}
}

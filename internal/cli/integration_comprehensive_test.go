package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/graph"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/query"
)

// TestIntegration_IndexRebuild_WithTags tests index rebuild with notes containing tags
func TestIntegration_IndexRebuild_WithTags(t *testing.T) {
	tmpDir := t.TempDir()

	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

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
tags: [important, todo, project]
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Test Note
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query search with tag filter
	searchQuery := query.NewSearchQuery()
	searchQuery.Tags = []string{"important"}

	results, err := query.ExecuteSearch(tmpDir, searchQuery)
	if err != nil {
		t.Fatalf("executing search: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected to find at least one note with 'important' tag")
	}
}

// TestIntegration_QuerySearch_WithMultipleFilters tests search query with multiple filters
func TestIntegration_QuerySearch_WithMultipleFilters(t *testing.T) {
	tmpDir := t.TempDir()

	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create note with specific state
	notePath := filepath.Join(noteDir, "published-note.Rmd")
	noteContent := `---
id: note-1
type: note
key: published-note
title: Published Note
state: published
tags: [important]
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Published Note
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query search with multiple filters
	searchQuery := query.NewSearchQuery()
	searchQuery.Types = []string{"note"}
	searchQuery.States = []string{"published"}
	searchQuery.Tags = []string{"important"}

	results, err := query.ExecuteSearch(tmpDir, searchQuery)
	if err != nil {
		t.Fatalf("executing search: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected to find at least one note matching all filters")
	}
}

// TestIntegration_IndexExport_WithEdges tests index export with edges
func TestIntegration_IndexExport_WithEdges(t *testing.T) {
	tmpDir := t.TempDir()

	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

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

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Export index
	exportPath := filepath.Join(tmpDir, "index-export.json")
	if err := index.Export(tmpDir, exportPath); err != nil {
		t.Fatalf("exporting index: %v", err)
	}

	// Verify export file exists and contains edges
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}

	// Read export file and verify it contains edges
	exportData, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("reading export file: %v", err)
	}

	exportStr := string(exportData)
	if !strings.Contains(exportStr, "edges") {
		t.Error("expected export to contain 'edges' field")
	}
}

// TestIntegration_GraphExport_WithMultipleNodes tests graph export with multiple nodes
func TestIntegration_GraphExport_WithMultipleNodes(t *testing.T) {
	tmpDir := t.TempDir()

	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create multiple notes
	notes := []struct {
		id    string
		key   string
		title string
	}{
		{"note-1", "alpha", "Alpha"},
		{"note-2", "beta", "Beta"},
		{"note-3", "gamma", "Gamma"},
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
`
		if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
			t.Fatalf("writing note %s: %v", n.key, err)
		}
	}

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

	// Verify file contains all nodes
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("reading export file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "note-1") {
		t.Error("expected export to contain 'note-1'")
	}
	if !strings.Contains(contentStr, "note-2") {
		t.Error("expected export to contain 'note-2'")
	}
	if !strings.Contains(contentStr, "note-3") {
		t.Error("expected export to contain 'note-3'")
	}
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/query"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// TestIntegration_FullWorkflow_InitNewIndexQueryExport tests the complete workflow
func TestIntegration_FullWorkflow_InitNewIndexQueryExport(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Step 2: Create a note
	if err := runNewWizard(tmpDir, cfg); err != nil {
		t.Fatalf("creating note: %v", err)
	}

	// Step 3: Build index
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Step 4: Query search
	searchQuery := query.NewSearchQuery()
	results, err := query.ExecuteSearch(tmpDir, searchQuery)
	if err != nil {
		t.Fatalf("executing search: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected to find at least one note")
	}

	// Step 5: Export index
	exportPath := filepath.Join(tmpDir, "export.json")
	if err := index.Export(tmpDir, exportPath); err != nil {
		t.Fatalf("exporting index: %v", err)
	}

	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}
}

// TestIntegration_QueryBacklinks_WithMultipleLinks tests backlinks query with multiple links
func TestIntegration_QueryBacklinks_WithMultipleLinks(t *testing.T) {
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

	// Create multiple source notes linking to target
	for i := 1; i <= 3; i++ {
		sourcePath := filepath.Join(noteDir, fmt.Sprintf("source-%d.Rmd", i))
		sourceContent := fmt.Sprintf(`---
id: note-source-%d
type: note
key: source-%d
title: Source Note %d
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note %d

Links to [[note:target]].
`, i, i, i, i)
		if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
			t.Fatalf("writing source note %d: %v", i, err)
		}
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query backlinks
	backlinksQuery := query.NewBacklinksQuery()
	backlinksQuery.Target = "note:target"
	backlinksQuery.Direction = "in"

	results, err := query.ExecuteBacklinks(tmpDir, backlinksQuery)
	if err != nil {
		t.Fatalf("executing backlinks query: %v", err)
	}

	// Should find at least 3 backlinks
	if len(results) < 3 {
		t.Errorf("expected at least 3 backlinks, got %d", len(results))
	}
}

// TestIntegration_QueryNeighbors_WithDepth tests neighbors query with depth
func TestIntegration_QueryNeighbors_WithDepth(t *testing.T) {
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
		noteContent := fmt.Sprintf(`---
id: %s
type: note
key: %s
title: %s
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# %s

%s
`, n.id, n.key, n.title, n.title, n.links)
		if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
			t.Fatalf("writing note %s: %v", n.key, err)
		}
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query neighbors with depth 2
	neighborsQuery := query.NewNeighborsQuery()
	neighborsQuery.Root = "note:root"
	neighborsQuery.MaxDepth = 2
	neighborsQuery.Direction = "out"

	results, err := query.ExecuteNeighbors(tmpDir, neighborsQuery)
	if err != nil {
		t.Fatalf("executing neighbors query: %v", err)
	}

	// Should find neighbors at depth 1 and 2
	if len(results) == 0 {
		t.Error("expected to find at least one neighbor")
	}
}

// TestIntegration_IndexRebuild_WithTagsAndEdges tests index rebuild with tags and edges
func TestIntegration_IndexRebuild_WithTagsAndEdges(t *testing.T) {
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

	// Create note with tags and links
	notePath := filepath.Join(noteDir, "test-note.Rmd")
	noteContent := `---
id: note-1
type: note
key: test-note
title: Test Note
state: draft
tags: [important, project, todo]
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Test Note

Links to [[note:target]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
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

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Verify tags were indexed
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var tagCount int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = 'note-1'").Scan(&tagCount)
	if err != nil {
		t.Fatalf("counting tags: %v", err)
	}
	if tagCount != 3 {
		t.Errorf("expected 3 tags, got %d", tagCount)
	}

	// Verify edges were created
	var edgeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = 'note-1'").Scan(&edgeCount)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if edgeCount == 0 {
		t.Error("expected at least one edge")
	}
}

package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// TestBuilder_Rebuild_WithMultipleNotes tests rebuilding index with multiple notes
func TestBuilder_Rebuild_WithMultipleNotes(t *testing.T) {
	tmpDir := t.TempDir()

	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create multiple test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notes := []struct {
		id    string
		key   string
		title string
	}{
		{"note-1", "alpha-note", "Alpha Note"},
		{"note-2", "beta-note", "Beta Note"},
		{"note-3", "gamma-note", "Gamma Note"},
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

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify all nodes were indexed
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	if err != nil {
		t.Fatalf("counting nodes: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 nodes, got %d", count)
	}
}

// TestBuilder_Rebuild_WithLinks tests rebuilding index with linked notes
func TestBuilder_Rebuild_WithLinks(t *testing.T) {
	tmpDir := t.TempDir()

	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create first note with link
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

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify edges were created
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var edgeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = 'note-1'").Scan(&edgeCount)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if edgeCount == 0 {
		t.Error("expected at least one edge from note-1")
	}
}

// TestBuilder_Rebuild_WithUnresolvedLinks tests rebuilding index with unresolved links
func TestBuilder_Rebuild_WithUnresolvedLinks(t *testing.T) {
	tmpDir := t.TempDir()

	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create note with unresolved link
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

Links to [[nonexistent-note]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify diagnostics were created for unresolved link
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var diagCount int
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = 'note-1' AND code = 'UNRESOLVED_LINK'").Scan(&diagCount)
	if err != nil {
		t.Fatalf("counting diagnostics: %v", err)
	}
	if diagCount == 0 {
		t.Error("expected at least one UNRESOLVED_LINK diagnostic")
	}
}

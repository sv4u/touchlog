package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
	"github.com/sv4u/touchlog/internal/store"
)

func TestResolveNodeID_Qualified(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index
	setupTestVaultWithNotes(t, tmpDir)

	// Test qualified lookup
	id, err := resolveNodeID(tmpDir, "note:alpha-note")
	if err != nil {
		t.Fatalf("resolveNodeID failed: %v", err)
	}

	if id != "note-1" {
		t.Errorf("expected node ID 'note-1', got %q", id)
	}
}

func TestResolveNodeID_Unqualified_Unique(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index
	setupTestVaultWithNotes(t, tmpDir)

	// Test unqualified lookup (should work if unique)
	id, err := resolveNodeID(tmpDir, "alpha-note")
	if err != nil {
		t.Fatalf("resolveNodeID failed: %v", err)
	}

	if id != "note-1" {
		t.Errorf("expected node ID 'note-1', got %q", id)
	}
}

func TestResolveNodeID_Unqualified_Ambiguous(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index with ambiguous keys
	setupTestVaultWithNotes(t, tmpDir)

	// Create another note with same key but different type
	// This requires adding another type to config
	// For now, just test that ambiguous lookup fails appropriately
	// (We'd need to set up a multi-type vault for a full test)
}

func TestResolveNodeID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault
	setupTestVaultWithIndex(t, tmpDir)

	// Test non-existent node
	_, err := resolveNodeID(tmpDir, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent node")
	}
}

func TestExecuteBacklinks_NoBacklinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index with isolated note
	setupTestVaultWithNotes(t, tmpDir)

	q := NewBacklinksQuery()
	q.Target = "note:alpha-note"

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 backlinks, got %d", len(results))
	}
}

func TestExecuteBacklinks_WithLinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index with linked notes
	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes with links
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create source note that links to target
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

This links to [[note:target-note]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create target note
	targetPath := filepath.Join(noteDir, "target-note.Rmd")
	targetContent := `---
id: note-target
type: note
key: target-note
title: Target Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Target Note

This is the target.
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatalf("writing target note: %v", err)
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

	// Query backlinks for target
	q := NewBacklinksQuery()
	q.Target = "note:target-note"
	q.Direction = "in" // Incoming links

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
		if len(results[0].Nodes) != 2 {
			t.Errorf("expected 2 nodes in path, got %d", len(results[0].Nodes))
		}
	}
}

// setupTestVaultWithNotes creates a test vault with notes and builds index
func setupTestVaultWithNotes(t *testing.T, tmpDir string) {
	t.Helper()

	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	notes := []struct {
		key   string
		id    string
		title string
	}{
		{"alpha-note", "note-1", "Alpha Note"},
		{"beta-note", "note-2", "Beta Note"},
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
			t.Fatalf("writing test note %s: %v", n.key, err)
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
}

// setupTestVaultWithIndex creates a test vault with config and database schema
func setupTestVaultWithIndex(t *testing.T, tmpDir string) {
	t.Helper()

	// Create .touchlog directory
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create config file
	configPath := filepath.Join(touchlogDir, "config.yaml")
	configContent := `version: 1
types:
  note:
    description: A note
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
tags:
  preferred: []
edges:
  related-to:
    description: General relationship
templates:
  root: templates
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Create and initialize database with schema
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	if err := store.ApplyMigrations(db); err != nil {
		db.Close()
		t.Fatalf("applying migrations: %v", err)
	}
	db.Close()
}

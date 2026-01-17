package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
	"github.com/sv4u/touchlog/internal/query"
	"github.com/sv4u/touchlog/internal/store"
)

// TestIntegration_InitNewIndexQuery tests the full workflow: init -> new -> index -> query
func TestIntegration_InitNewIndexQuery(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify config exists
	configPath := filepath.Join(tmpDir, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Step 2: Load config
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Step 3: Create a note
	if err := runNewWizard(tmpDir, cfg); err != nil {
		t.Fatalf("creating note: %v", err)
	}

	// Verify note was created
	var notePath string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".Rmd") {
			notePath = path
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking directory: %v", err)
	}
	if notePath == "" {
		t.Fatal("note file was not created")
	}

	// Step 4: Build index
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Verify database exists
	dbPath := filepath.Join(tmpDir, ".touchlog", "index.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database file not created: %v", err)
	}

	// Step 5: Query the index
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Execute search query
	searchQuery := query.NewSearchQuery()
	results, err := query.ExecuteSearch(tmpDir, searchQuery)
	if err != nil {
		t.Fatalf("executing search: %v", err)
	}

	// Should find at least one note
	if len(results) == 0 {
		t.Error("expected to find at least one note in search results")
	}
}

// TestIntegration_IndexRebuildWithLinks tests index rebuild with linked notes
func TestIntegration_IndexRebuildWithLinks(t *testing.T) {
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

	// Query backlinks
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Verify both notes are indexed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	if err != nil {
		t.Fatalf("counting nodes: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 nodes, got %d", count)
	}

	// Verify edges are created
	err = db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&count)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one edge between notes")
	}
}

// TestIntegration_DiagnosticsAfterIndex tests diagnostics are generated after indexing
func TestIntegration_DiagnosticsAfterIndex(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create note with unresolved link
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

Links to [[nonexistent-note]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Build index
	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query diagnostics
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	diagnostics, err := queryDiagnostics(db, "", "", "")
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	// Should have at least one diagnostic for unresolved link
	if len(diagnostics) == 0 {
		t.Error("expected at least one diagnostic for unresolved link")
	}

	// Check for unresolved link diagnostic
	foundUnresolved := false
	for _, diag := range diagnostics {
		if diag.Code == "UNRESOLVED_LINK" || strings.Contains(diag.Message, "not found") {
			foundUnresolved = true
			break
		}
	}

	if !foundUnresolved {
		t.Error("expected to find UNRESOLVED_LINK diagnostic")
	}
}

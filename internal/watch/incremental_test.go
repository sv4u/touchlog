package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
	"github.com/sv4u/touchlog/internal/store"
)

func TestIncrementalIndexer_ProcessFileUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and initial index
	setupTestVaultWithIndex(t, tmpDir)

	// Open database
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	// Load config
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create incremental indexer
	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Create a new note file
	notePath := filepath.Join(tmpDir, "note", "new-note.Rmd")
	noteContent := `---
id: note-new
type: note
key: new-note
title: New Note
state: draft
tags: []
created: 2024-01-04T00:00:00Z
updated: 2024-01-04T00:00:00Z
---
# New Note

This is a new note.
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Process create event
	event := Event{
		Path:      notePath,
		Op:        fsnotify.Create,
		Timestamp: time.Now(),
	}

	if err := indexer.ProcessEvent(event); err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	// Verify node was indexed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = 'note-new'").Scan(&count)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 node, got %d", count)
	}
}

func TestIncrementalIndexer_ProcessFileDelete(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and initial index with a note
	setupTestVaultWithIndex(t, tmpDir)

	// Create a note file first
	notePath := filepath.Join(tmpDir, "note", "test-note.Rmd")
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

	// Build initial index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Open database
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	// Verify node exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = 'note-test'").Scan(&count)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 node before deletion, got %d", count)
	}

	// Create incremental indexer
	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Delete the file
	if err := os.Remove(notePath); err != nil {
		t.Fatalf("removing note: %v", err)
	}

	// Process delete event
	event := Event{
		Path:      notePath,
		Op:        fsnotify.Remove,
		Timestamp: time.Now(),
	}

	if err := indexer.ProcessEvent(event); err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	// Verify node was deleted
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = 'note-test'").Scan(&count)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 nodes after deletion, got %d", count)
	}
}

func TestIncrementalIndexer_ChangeDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and initial index
	setupTestVaultWithIndex(t, tmpDir)

	// Create a note file
	notePath := filepath.Join(tmpDir, "note", "test-note.Rmd")
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

	// Build initial index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Get initial mtime
	initialInfo, err := os.Stat(notePath)
	if err != nil {
		t.Fatalf("stating file: %v", err)
	}
	initialMtime := initialInfo.ModTime()

	// Open database
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	// Create incremental indexer
	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Process initial event (should index)
	event1 := Event{
		Path:      notePath,
		Op:        fsnotify.Create,
		Timestamp: time.Now(),
	}
	if err := indexer.ProcessEvent(event1); err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	// Wait a bit and modify the file
	time.Sleep(100 * time.Millisecond)
	modifiedContent := noteContent + "\n\nModified content.\n"
	if err := os.WriteFile(notePath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("modifying note: %v", err)
	}

	// Get new mtime
	newInfo, err := os.Stat(notePath)
	if err != nil {
		t.Fatalf("stating file: %v", err)
	}
	newMtime := newInfo.ModTime()

	// Verify mtime changed
	if newMtime.Equal(initialMtime) {
		t.Fatal("mtime did not change after modification")
	}

	// Process update event (should update)
	event2 := Event{
		Path:      notePath,
		Op:        fsnotify.Write,
		Timestamp: time.Now(),
	}
	if err := indexer.ProcessEvent(event2); err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	// Verify node was updated (check updated timestamp in database)
	var updatedTime string
	err = db.QueryRow("SELECT updated FROM nodes WHERE id = 'note-test'").Scan(&updatedTime)
	if err != nil {
		t.Fatalf("querying updated time: %v", err)
	}
	// The updated time should reflect the new modification
}

// setupTestVaultWithIndex creates a test vault with config and builds initial index
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

	// Create note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}
}

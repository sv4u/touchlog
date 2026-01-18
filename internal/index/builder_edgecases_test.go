package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// setupTestVault creates a test vault with config
func setupTestVault(t *testing.T, tmpDir string) {
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
		_ = db.Close()
		t.Fatalf("applying migrations: %v", err)
	}
	_ = db.Close()
}

// TestBuilder_Rebuild_Behavior_HandlesExistingTempDB tests Rebuild handles existing temp database
func TestBuilder_Rebuild_Behavior_HandlesExistingTempDB(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := NewBuilder(tmpDir, cfg)

	// Create existing temp database
	tmpDBPath := filepath.Join(tmpDir, ".touchlog", "index.db.tmp")
	if err := os.WriteFile(tmpDBPath, []byte("test"), 0644); err != nil {
		t.Fatalf("creating temp DB file: %v", err)
	}

	// Rebuild should remove existing temp DB and succeed
	err = builder.Rebuild()
	if err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify final database exists
	finalDBPath := filepath.Join(tmpDir, ".touchlog", "index.db")
	if _, err := os.Stat(finalDBPath); os.IsNotExist(err) {
		t.Error("expected final database to exist after Rebuild")
	}
}

// TestBuilder_Rebuild_Behavior_AtomicReplace tests Rebuild atomically replaces database
func TestBuilder_Rebuild_Behavior_AtomicReplace(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create initial database with content
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("initial Rebuild failed: %v", err)
	}

	// Verify database exists
	finalDBPath := filepath.Join(tmpDir, ".touchlog", "index.db")
	if _, err := os.Stat(finalDBPath); os.IsNotExist(err) {
		t.Fatal("expected database to exist after initial Rebuild")
	}

	// Rebuild again - should atomically replace
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("second Rebuild failed: %v", err)
	}

	// Verify database still exists (atomic replace succeeded)
	if _, err := os.Stat(finalDBPath); os.IsNotExist(err) {
		t.Error("expected database to exist after second Rebuild")
	}
}

// TestBuilder_IndexAll_Behavior_HandlesUnreadableFiles tests indexAll handles unreadable files
func TestBuilder_IndexAll_Behavior_HandlesUnreadableFiles(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := NewBuilder(tmpDir, cfg)

	// Create note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create a file that will be discovered but may have issues
	// (indexAll should skip unreadable files and continue)
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
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Create database
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := store.ApplyMigrations(db); err != nil {
		t.Fatalf("applying migrations: %v", err)
	}

	// Index all - should handle files gracefully
	err = builder.indexAll(db)
	if err != nil {
		t.Fatalf("indexAll failed: %v", err)
	}
}

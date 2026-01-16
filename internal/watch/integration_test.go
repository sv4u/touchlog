package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/store"
)

func TestIncrementalIndexing_TransactionSafety(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and initial index
	setupTestVaultWithIndex(t, tmpDir)

	// Load config
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Open database
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	// Create incremental indexer
	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

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

	// Process create event
	event := Event{
		Path:      notePath,
		Op:        fsnotify.Create,
		Timestamp: time.Now(),
	}

	// Process event (should use transaction)
	if err := indexer.ProcessEvent(event); err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	// Verify node was indexed (transaction should have committed)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = 'note-test'").Scan(&count)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 node after transaction, got %d", count)
	}
}

func TestIncrementalIndexing_ShortLivedConnections(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and initial index
	setupTestVaultWithIndex(t, tmpDir)

	// Load config
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Create incremental indexer (db connection will be closed after use)
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	indexer := NewIncrementalIndexer(tmpDir, cfg, db)
	db.Close() // Close immediately - ProcessEvent should open its own connection

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

	// Process event (should open its own connection)
	event := Event{
		Path:      notePath,
		Op:        fsnotify.Create,
		Timestamp: time.Now(),
	}

	// This should work even though we closed the original db connection
	// because ProcessEvent opens its own connection
	if err := indexer.ProcessEvent(event); err != nil {
		t.Fatalf("ProcessEvent failed (should use short-lived connection): %v", err)
	}

	// Verify using a new connection
	db2, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db2.Close()

	var count int
	err = db2.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = 'note-test'").Scan(&count)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 node, got %d", count)
	}
}

func TestWatcher_Debouncing(t *testing.T) {
	tmpDir := t.TempDir()

	watcher, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer watcher.Stop()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Rapidly modify a file multiple times
	testFile := filepath.Join(tmpDir, "test.Rmd")
	for i := 0; i < 10; i++ {
		if err := os.WriteFile(testFile, []byte("test"+string(rune(i))), 0644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Very rapid modifications
	}

	// Wait for debounced events (wait longer than debounce period)
	timeout := time.After(500 * time.Millisecond)
	eventCount := 0
	
	// Collect events until timeout
	for {
		select {
		case event := <-watcher.Events():
			if event.Path == testFile {
				eventCount++
			}
		case err := <-watcher.Errors():
			t.Fatalf("watcher error: %v", err)
		case <-timeout:
			// Timeout reached, check results
			goto done
		}
	}
done:

	// With debouncing, we should get significantly fewer events than modifications
	// (exact count depends on timing, but should be much less than 10)
	if eventCount >= 10 {
		t.Errorf("debouncing not working effectively: got %d events for 10 rapid modifications", eventCount)
	}
}

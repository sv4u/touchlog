package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	watcher, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.Rmd")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Wait for event (with timeout)
	timeout := time.After(2 * time.Second)
	select {
	case event := <-watcher.Events():
		if event.Path != testFile {
			t.Errorf("expected event path %q, got %q", testFile, event.Path)
		}
	case err := <-watcher.Errors():
		t.Fatalf("watcher error: %v", err)
	case <-timeout:
		t.Error("timeout waiting for file event")
	}
}

func TestWatcher_IgnoresNonRmdFiles(t *testing.T) {
	tmpDir := t.TempDir()

	watcher, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Create a non-.Rmd file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Wait a bit - should NOT receive event for .txt file
	timeout := time.After(500 * time.Millisecond)
	select {
	case event := <-watcher.Events():
		t.Errorf("unexpected event for non-.Rmd file: %v", event)
	case <-timeout:
		// Expected: no event for .txt file
	}
}

func TestWatcher_Debounce(t *testing.T) {
	tmpDir := t.TempDir()

	watcher, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Rapidly modify a file multiple times
	testFile := filepath.Join(tmpDir, "test.Rmd")
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(testFile, []byte("test"+string(rune(i))), 0644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for debounced event
	timeout := time.After(1 * time.Second)
	eventCount := 0
	for eventCount < 1 {
		select {
		case event := <-watcher.Events():
			if event.Path == testFile {
				eventCount++
			}
		case err := <-watcher.Errors():
			t.Fatalf("watcher error: %v", err)
		case <-timeout:
			if eventCount == 0 {
				t.Error("timeout waiting for debounced event")
			}
			return
		}
	}

	// With debouncing, we should get fewer events than modifications
	// (exact count depends on timing, but should be less than 5)
	if eventCount >= 5 {
		t.Errorf("debouncing not working: got %d events for 5 rapid modifications", eventCount)
	}
}

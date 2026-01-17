package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWrite_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.txt")
	content := []byte("")

	if err := AtomicWrite(filePath, content); err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	// Verify content is empty
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if len(readContent) != 0 {
		t.Errorf("expected empty content, got %q", readContent)
	}
}

func TestAtomicWrite_LargeContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.txt")
	// Create 1MB content
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}

	if err := AtomicWrite(filePath, content); err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists and has correct size
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("statting file: %v", err)
	}

	if info.Size() != int64(len(content)) {
		t.Errorf("expected file size %d, got %d", len(content), info.Size())
	}
}

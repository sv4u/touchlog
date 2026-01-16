package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWrite writes content to a file atomically using temp file + fsync + rename
// This ensures the file is never in a partially-written state
func AtomicWrite(path string, content []byte) error {
	dir := filepath.Dir(path)
	
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	
	// Create temp file in same directory
	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp.*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	
	// Ensure cleanup on error
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()
	
	// Write content
	if _, err = tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing to temp file: %w", err)
	}
	
	// Sync to disk (best effort)
	if err = tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("syncing temp file: %w", err)
	}
	
	// Close temp file
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	
	// Atomic rename
	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}
	
	return nil
}

package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/store"
)

func TestBuildViewCommand(t *testing.T) {
	cmd := BuildViewCommand()

	if cmd.Name != "view" {
		t.Errorf("expected command name 'view', got %q", cmd.Name)
	}

	if cmd.Usage == "" {
		t.Error("expected Usage to be set")
	}

	if cmd.Description == "" {
		t.Error("expected Description to be set")
	}

	// Verify flags exist
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		for _, name := range flag.Names() {
			flagNames[name] = true
		}
	}

	expectedFlags := []string{"file", "key", "type", "tag"}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestValidateRmdFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.Rmd")
	content := []byte("# Test\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	if err := validateRmdFile(filePath); err != nil {
		t.Errorf("expected no error for valid file, got: %v", err)
	}
}

func TestValidateRmdFile_CaseInsensitiveExtension(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []string{"test.Rmd", "test.rmd", "test.RMD"}
	for _, filename := range testCases {
		filePath := filepath.Join(tmpDir, filename)
		content := []byte("# Test\n")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("writing test file %s: %v", filename, err)
		}

		if err := validateRmdFile(filePath); err != nil {
			t.Errorf("expected no error for %s, got: %v", filename, err)
		}

		// Clean up for next iteration
		_ = os.Remove(filePath)
	}
}

func TestValidateRmdFile_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.Rmd")

	err := validateRmdFile(filePath)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if err != nil && !contains(err.Error(), "file not found") {
		t.Errorf("expected 'file not found' error, got: %v", err)
	}
}

func TestValidateRmdFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "test.Rmd")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("creating test directory: %v", err)
	}

	err := validateRmdFile(dirPath)
	if err == nil {
		t.Error("expected error for directory")
	}
	if err != nil && !strings.Contains(err.Error(), "directory") {
		t.Errorf("expected 'directory' error, got: %v", err)
	}
}

func TestValidateRmdFile_WrongExtension(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		filename string
	}{
		{"test.txt"},
		{"test.md"},
		{"test.R"},
		{"test"},
	}

	for _, tc := range testCases {
		filePath := filepath.Join(tmpDir, tc.filename)
		content := []byte("# Test\n")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("writing test file %s: %v", tc.filename, err)
		}

		err := validateRmdFile(filePath)
		if err == nil {
			t.Errorf("expected error for file with extension %q", tc.filename)
		}
		if err != nil && !strings.Contains(err.Error(), ".Rmd or .rmd extension") {
			t.Errorf("expected extension error for %s, got: %v", tc.filename, err)
		}

		// Clean up
		_ = os.Remove(filePath)
	}
}

func TestViewByFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.Rmd")
	content := []byte("# Test\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Skip if Rscript is not available
	skipIfNoRscript(t)

	// Use context with timeout to ensure cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Run in goroutine so we can cancel it
	errCh := make(chan error, 1)
	go func() {
		errCh <- runRmarkdownWithContext(ctx, filePath)
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		// If it completes quickly (validation error), that's fine
		// If it's a context cancellation, that's expected
		if err != nil && !strings.Contains(err.Error(), "file not found") &&
			!strings.Contains(err.Error(), "cancelled") &&
			!strings.Contains(err.Error(), "timeout") {
			// Only fail on unexpected errors
			if !strings.Contains(err.Error(), "rmarkdown") {
				t.Errorf("unexpected error: %v", err)
			}
		}
	case <-ctx.Done():
		// Timeout is expected - rmarkdown::run() starts a server
		// The context cancellation will kill the process
	}
}

func TestViewByFile_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalCwd)
	}()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("changing to temp directory: %v", err)
	}

	filePath := "test.Rmd"
	content := []byte("# Test\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Skip if Rscript is not available
	skipIfNoRscript(t)

	// Use context with timeout to ensure cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Run in goroutine so we can cancel it
	errCh := make(chan error, 1)
	go func() {
		// Relative path should be normalized to absolute
		errCh <- runRmarkdownWithContext(ctx, filePath)
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		// If it completes quickly (validation error), that's fine
		// If it's a context cancellation, that's expected
		if err != nil && strings.Contains(err.Error(), "resolving file path") {
			t.Errorf("unexpected path resolution error: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected - rmarkdown::run() starts a server
		// The context cancellation will kill the process
	}
}

func TestViewByFile_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.Rmd")

	err := viewByFile(filePath)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestViewByKey_NoteFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Create a note manually
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	noteContent := `---
id: test-note-1
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test Note
tags: []
state: draft
---

# Test Note
`
	notePath := filepath.Join(noteDir, "test-key.Rmd")
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Build the index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	// Skip if Rscript is not available
	skipIfNoRscript(t)

	// Use context with timeout to ensure cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Test qualified key - we'll test the resolution separately, just verify it doesn't crash
	// For actual R execution, we need to use the context-aware function
	notePath, resolveErr := resolveNotePathForTesting(tmpDir, "note:test-key")
	if resolveErr != nil {
		t.Fatalf("resolving note path: %v", resolveErr)
	}

	// Run rmarkdown with context
	errCh := make(chan error, 1)
	go func() {
		errCh <- runRmarkdownWithContext(ctx, notePath)
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		// If it completes quickly (validation error), that's fine
		// If it's a context cancellation, that's expected
		if err != nil && strings.Contains(err.Error(), "note not found") {
			t.Errorf("unexpected note resolution error: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected - rmarkdown::run() starts a server
		// The context cancellation will kill the process
	}
}

func TestViewByKey_NoteNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Build empty index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	// Try to view a non-existent note
	err = viewByKey(tmpDir, "nonexistent-key")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
	if err != nil && !strings.Contains(err.Error(), "note not found") {
		t.Errorf("expected 'note not found' error, got: %v", err)
	}
}

func TestViewByKey_AmbiguousKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Create two notes with the same last segment but different paths
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// First note with key "auth"
	note1Content := `---
id: test-note-1
type: note
key: auth
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Auth Note
tags: []
state: draft
---

# Auth Note
`
	if err := os.WriteFile(filepath.Join(noteDir, "auth.Rmd"), []byte(note1Content), 0644); err != nil {
		t.Fatalf("writing note 1: %v", err)
	}

	// Second note with key "projects/auth" (needs subdirectory)
	projectsDir := filepath.Join(noteDir, "projects", "auth")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("creating projects dir: %v", err)
	}

	note2Content := `---
id: test-note-2
type: note
key: projects/auth
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Projects Auth Note
tags: []
state: draft
---

# Projects Auth Note
`
	if err := os.WriteFile(filepath.Join(projectsDir, "auth.Rmd"), []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

	// Build the index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	// Qualified keys should work - test resolution only (don't actually run R)
	notePath1, err := resolveNotePathForTesting(tmpDir, "note:auth")
	if err != nil {
		t.Errorf("qualified key should resolve: %v", err)
	}
	if notePath1 == "" {
		t.Error("expected note path for note:auth")
	}

	notePath2, err := resolveNotePathForTesting(tmpDir, "note:projects/auth")
	if err != nil {
		t.Errorf("qualified key should resolve: %v", err)
	}
	if notePath2 == "" {
		t.Error("expected note path for note:projects/auth")
	}

	// Unqualified "auth" should resolve to exact match (not ambiguous)
	skipIfNoRscript(t)

	// Use context with timeout to ensure cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Resolve the note path first
	notePath, err := resolveNotePathForTesting(tmpDir, "auth")
	if err != nil {
		t.Fatalf("resolving note path: %v", err)
	}

	// Run rmarkdown with context
	errCh := make(chan error, 1)
	go func() {
		errCh <- runRmarkdownWithContext(ctx, notePath)
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		// Should resolve to exact match, not be ambiguous
		if err != nil && strings.Contains(err.Error(), "ambiguous") {
			t.Errorf("unqualified 'auth' should resolve to exact match, not be ambiguous: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected - rmarkdown::run() starts a server
		// The context cancellation will kill the process
	}
}

func TestRunRmarkdown_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.Rmd")

	err := runRmarkdown(filePath)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if err != nil && !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected 'file not found' error, got: %v", err)
	}
}

func TestRunRmarkdown_InvalidExtension(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("# Test\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	err := runRmarkdown(filePath)
	if err == nil {
		t.Error("expected error for file with wrong extension")
	}
	if err != nil && !strings.Contains(err.Error(), ".Rmd or .rmd extension") {
		t.Errorf("expected extension error, got: %v", err)
	}
}

func TestRunViewWizard_EmptyNoteList(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Build empty index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	err = runViewWizard(tmpDir, cfg, "", nil)
	if err == nil {
		t.Error("expected error for empty note list")
	}
	if err != nil && !strings.Contains(err.Error(), "no notes found") {
		t.Errorf("expected 'no notes found' error, got: %v", err)
	}
}

// Helper functions

// hasRscript checks if Rscript is available
func hasRscript() bool {
	_, err := exec.LookPath("Rscript")
	return err == nil
}

// skipIfNoRscript skips test if Rscript is not available
func skipIfNoRscript(t *testing.T) {
	if !hasRscript() {
		t.Skip("Rscript not available, skipping R execution test")
	}
}

// resolveNotePathForTesting is a helper to resolve note paths in tests
// It opens the database and calls resolveNotePath
func resolveNotePathForTesting(vaultRoot, identifier string) (string, error) {
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = db.Close()
	}()
	return resolveNotePath(db, vaultRoot, identifier)
}

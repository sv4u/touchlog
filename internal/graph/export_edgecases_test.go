package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

// TestExportDOT_FileExistsWithoutForce tests export when file exists and force is false
func TestExportDOT_FileExistsWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	// Create an existing file
	exportPath := filepath.Join(tmpDir, "graph.dot")
	if err := os.WriteFile(exportPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	opts := ExportOptions{
		Force: false,
	}

	err := ExportDOT(tmpDir, exportPath, opts)
	if err == nil {
		t.Error("expected error when file exists and force is false")
	}
}

// TestExportDOT_FileExistsWithForce tests export when file exists and force is true
func TestExportDOT_FileExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

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
		t.Fatalf("writing test note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Create an existing file
	exportPath := filepath.Join(tmpDir, "graph.dot")
	if err := os.WriteFile(exportPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	opts := ExportOptions{
		Force: true,
	}

	err = ExportDOT(tmpDir, exportPath, opts)
	if err != nil {
		t.Fatalf("ExportDOT with force=true failed: %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("reading export file: %v", err)
	}

	if string(content) == "existing content" {
		t.Error("expected file to be overwritten")
	}
}

// TestExportDOT_WithFilters tests export with various filters
func TestExportDOT_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

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
		t.Fatalf("writing test note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	exportPath := filepath.Join(tmpDir, "graph.dot")
	opts := ExportOptions{
		Types:     []string{"note"},
		States:    []string{"draft"},
		EdgeTypes: []string{"related-to"},
		Depth:     5,
		Force:     true,
	}

	err = ExportDOT(tmpDir, exportPath, opts)
	if err != nil {
		t.Fatalf("ExportDOT with filters failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}
}

// TestExportDOT_WithRoots tests export with root nodes specified
func TestExportDOT_WithRoots(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	// Create test notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

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
		t.Fatalf("writing test note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	exportPath := filepath.Join(tmpDir, "graph.dot")
	opts := ExportOptions{
		Roots: []string{"note:test-note"},
		Depth: 2,
		Force: true,
	}

	err = ExportDOT(tmpDir, exportPath, opts)
	if err != nil {
		t.Fatalf("ExportDOT with roots failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file was not created: %v", err)
	}
}

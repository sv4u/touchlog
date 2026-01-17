package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
)

func TestExportDOT_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault and create index
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

	// Export graph
	outputPath := filepath.Join(tmpDir, "graph.dot")
	opts := ExportOptions{
		Depth: 10,
	}

	if err := ExportDOT(tmpDir, outputPath, opts); err != nil {
		t.Fatalf("ExportDOT failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("DOT file was not created: %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading DOT file: %v", err)
	}

	dotContent := string(content)
	if !strings.Contains(dotContent, "digraph touchlog") {
		t.Error("DOT file should contain 'digraph touchlog'")
	}

	if !strings.Contains(dotContent, "note-1") {
		t.Error("DOT file should contain node 'note-1'")
	}
}

func TestExportDOT_RefusesOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault
	setupTestVaultWithIndex(t, tmpDir)

	// Create existing file
	outputPath := filepath.Join(tmpDir, "graph.dot")
	if err := os.WriteFile(outputPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	// Try to export without --force
	opts := ExportOptions{
		Depth: 10,
		Force: false,
	}

	err := ExportDOT(tmpDir, outputPath, opts)
	if err == nil {
		t.Error("expected error when overwriting existing file without --force")
	}

	// Try with --force
	opts.Force = true
	if err := ExportDOT(tmpDir, outputPath, opts); err != nil {
		t.Fatalf("ExportDOT with --force failed: %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading DOT file: %v", err)
	}

	if strings.Contains(string(content), "existing content") {
		t.Error("file should have been overwritten")
	}
}

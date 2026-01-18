package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestExportDOT_MatchesFilters tests the matchesFilters function behavior
func TestExportDOT_MatchesFilters(t *testing.T) {
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

	// Load graph
	g, err := LoadGraph(tmpDir)
	if err != nil {
		t.Fatalf("loading graph: %v", err)
	}

	// Test matchesFilters with type filter
	node := g.Nodes["note-1"]
	if node == nil {
		t.Fatal("expected node to exist")
	}

	// Build type set
	typeSet := make(map[model.TypeName]bool)
	typeSet["note"] = true

	stateSet := make(map[string]bool)
	var tags []string

	// Test that node matches type filter
	matches := matchesFilters(node, typeSet, stateSet, tags)
	if !matches {
		t.Error("expected node to match type filter")
	}

	// Test with non-matching type filter
	typeSet = make(map[model.TypeName]bool)
	typeSet["task"] = true
	matches = matchesFilters(node, typeSet, stateSet, tags)
	if matches {
		t.Error("expected node not to match non-matching type filter")
	}

	// Test with state filter
	typeSet = make(map[model.TypeName]bool)
	stateSet = make(map[string]bool)
	stateSet["draft"] = true
	matches = matchesFilters(node, typeSet, stateSet, tags)
	if !matches {
		t.Error("expected node to match state filter")
	}

	// Test with non-matching state filter
	stateSet = make(map[string]bool)
	stateSet["published"] = true
	matches = matchesFilters(node, typeSet, stateSet, tags)
	if matches {
		t.Error("expected node not to match non-matching state filter")
	}
}

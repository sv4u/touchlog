package index

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

func TestBuilder_Rebuild_CreatesIndex(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create minimal config
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyPattern:   nil, // Will use default
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create a test note
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

This is a test note.
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing test note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify index was created
	indexPath := filepath.Join(touchlogDir, "index.db")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("index.db was not created: %v", err)
	}

	// Verify we can open and query the database
	// Use sql.Open directly since we know the path
	db, err := sql.Open("sqlite3", indexPath+"?_foreign_keys=1")
	if err != nil {
		t.Fatalf("opening index: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Verify node was indexed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = 'note-1'").Scan(&count)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 node, got %d", count)
	}
}

func TestBuilder_Rebuild_AtomicReplace(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create existing index (simulate previous index)
	oldIndexPath := filepath.Join(touchlogDir, "index.db")
	oldDB, err := sql.Open("sqlite3", oldIndexPath+"?_foreign_keys=1")
	if err != nil {
		t.Fatalf("creating old index: %v", err)
	}
	_ = oldDB.Close()

	// Create minimal config
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyPattern:   nil,
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	// Create a test note
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

	// Build index (should atomically replace old index)
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify old index was replaced (not corrupted)
	indexPath := filepath.Join(touchlogDir, "index.db")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("index.db was not created: %v", err)
	}

	// Verify temp file was cleaned up
	tmpIndexPath := filepath.Join(touchlogDir, "index.db.tmp")
	if _, err := os.Stat(tmpIndexPath); err == nil {
		t.Error("temp index file should have been removed")
	}
}

func TestBuilder_ResolveLinks_Qualified(t *testing.T) {
	builder := &Builder{}

	typeKeyMap := map[model.TypeKey]model.NoteID{
		{Type: "note", Key: "target-1"}: "note-target-1",
		{Type: "log", Key: "target-1"}:  "log-target-1",
	}

	// Build lastSegmentMap from typeKeyMap
	lastSegmentMap := buildLastSegmentMap(typeKeyMap)

	rawLinks := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "source-1"},
			Target: model.RawTarget{
				Type: func() *model.TypeName { t := model.TypeName("note"); return &t }(),
				Key:  "target-1",
			},
			EdgeType: model.DefaultEdgeType,
			Span:     model.Span{Path: "test.Rmd", StartByte: 0, EndByte: 10},
		},
	}

	resolvedEdges, diags := builder.resolveLinks(rawLinks, typeKeyMap, lastSegmentMap, "note")

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(diags))
	}

	if len(resolvedEdges) != 1 {
		t.Fatalf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}

	if resolvedEdges[0].ResolvedToID == nil {
		t.Error("expected resolved to_id, got nil")
	} else if *resolvedEdges[0].ResolvedToID != "note-target-1" {
		t.Errorf("expected resolved to_id 'note-target-1', got %q", *resolvedEdges[0].ResolvedToID)
	}
}

func TestBuilder_ResolveLinks_Unqualified_Unique(t *testing.T) {
	builder := &Builder{}

	typeKeyMap := map[model.TypeKey]model.NoteID{
		{Type: "note", Key: "target-1"}: "note-target-1",
	}

	// Build lastSegmentMap from typeKeyMap
	lastSegmentMap := buildLastSegmentMap(typeKeyMap)

	rawLinks := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "source-1"},
			Target: model.RawTarget{
				Type: nil, // Unqualified
				Key:  "target-1",
			},
			EdgeType: model.DefaultEdgeType,
			Span:     model.Span{Path: "test.Rmd", StartByte: 0, EndByte: 10},
		},
	}

	resolvedEdges, diags := builder.resolveLinks(rawLinks, typeKeyMap, lastSegmentMap, "note")

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(diags))
	}

	if len(resolvedEdges) != 1 {
		t.Fatalf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}

	if resolvedEdges[0].ResolvedToID == nil {
		t.Error("expected resolved to_id, got nil")
	} else if *resolvedEdges[0].ResolvedToID != "note-target-1" {
		t.Errorf("expected resolved to_id 'note-target-1', got %q", *resolvedEdges[0].ResolvedToID)
	}
}

func TestBuilder_ResolveLinks_Unqualified_Ambiguous(t *testing.T) {
	builder := &Builder{}

	typeKeyMap := map[model.TypeKey]model.NoteID{
		{Type: "note", Key: "target-1"}: "note-target-1",
		{Type: "log", Key: "target-1"}:  "log-target-1",
	}

	// Build lastSegmentMap from typeKeyMap
	lastSegmentMap := buildLastSegmentMap(typeKeyMap)

	rawLinks := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "source-1"},
			Target: model.RawTarget{
				Type: nil, // Unqualified
				Key:  "target-1",
			},
			EdgeType: model.DefaultEdgeType,
			Span:     model.Span{Path: "test.Rmd", StartByte: 0, EndByte: 10},
		},
	}

	resolvedEdges, diags := builder.resolveLinks(rawLinks, typeKeyMap, lastSegmentMap, "note")

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	if diags[0].Code != "AMBIGUOUS_LINK" {
		t.Errorf("expected diagnostic code 'AMBIGUOUS_LINK', got %q", diags[0].Code)
	}

	if diags[0].Level != model.DiagnosticLevelError {
		t.Errorf("expected diagnostic level 'error', got %q", diags[0].Level)
	}

	if len(resolvedEdges) != 1 {
		t.Fatalf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}

	if resolvedEdges[0].ResolvedToID != nil {
		t.Error("expected unresolved to_id for ambiguous link, got resolved")
	}
}

func TestBuilder_ResolveLinks_Unresolved(t *testing.T) {
	builder := &Builder{}

	typeKeyMap := map[model.TypeKey]model.NoteID{}

	// Build lastSegmentMap from typeKeyMap (empty in this case)
	lastSegmentMap := buildLastSegmentMap(typeKeyMap)

	rawLinks := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "source-1"},
			Target: model.RawTarget{
				Type: func() *model.TypeName { t := model.TypeName("note"); return &t }(),
				Key:  "nonexistent",
			},
			EdgeType: model.DefaultEdgeType,
			Span:     model.Span{Path: "test.Rmd", StartByte: 0, EndByte: 10},
		},
	}

	resolvedEdges, diags := builder.resolveLinks(rawLinks, typeKeyMap, lastSegmentMap, "note")

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	if diags[0].Code != "UNRESOLVED_LINK" {
		t.Errorf("expected diagnostic code 'UNRESOLVED_LINK', got %q", diags[0].Code)
	}

	if len(resolvedEdges) != 1 {
		t.Fatalf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}

	if resolvedEdges[0].ResolvedToID != nil {
		t.Error("expected unresolved to_id, got resolved")
	}
}

// buildLastSegmentMap builds a last-segment map from a typeKeyMap for testing
func buildLastSegmentMap(typeKeyMap map[model.TypeKey]model.NoteID) map[string][]model.NoteID {
	lastSegmentMap := make(map[string][]model.NoteID)
	for typeKey, noteID := range typeKeyMap {
		lastSeg := config.LastSegment(string(typeKey.Key))
		lastSegmentMap[lastSeg] = append(lastSegmentMap[lastSeg], noteID)
	}
	return lastSegmentMap
}

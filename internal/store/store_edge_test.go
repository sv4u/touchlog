package store

import (
	"testing"
	"time"

	"github.com/sv4u/touchlog/internal/model"
)

func TestReplaceEdgesForNode_MultipleUnresolved(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	now := time.Now().UTC()
	fromID := model.NoteID("from-id")

	// Insert source node
	err = UpsertNode(db, fromID, "note", "from-key", "From", "draft", now, now, "note/from-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Create multiple unresolved edges (to_id will be NULL)
	typeName := model.TypeName("note")
	edges := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: &typeName,
				Key:  "target1",
			},
			EdgeType: "related-to",
			Span: model.Span{
				Path:      "note/from-key.Rmd",
				StartByte: 100,
				EndByte:   120,
			},
		},
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: nil,
				Key:  "unqualified",
			},
			EdgeType: "depends-on",
			Span: model.Span{
				Path:      "note/from-key.Rmd",
				StartByte: 200,
				EndByte:   220,
			},
		},
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: &typeName,
				Key:  "target2",
			},
			EdgeType: "related-to",
			Span: model.Span{
				Path:      "note/from-key.Rmd",
				StartByte: 300,
				EndByte:   320,
			},
		},
	}

	err = ReplaceEdgesForNode(db, fromID, edges)
	if err != nil {
		t.Fatalf("ReplaceEdgesForNode failed: %v", err)
	}

	// Verify all edges were inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = ?", fromID).Scan(&count)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 edges, got %d", count)
	}

	// Verify all have NULL to_id (unresolved)
	var nullCount int
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = ? AND to_id IS NULL", fromID).Scan(&nullCount)
	if err != nil {
		t.Fatalf("counting NULL edges: %v", err)
	}
	if nullCount != 3 {
		t.Errorf("expected 3 edges with NULL to_id, got %d", nullCount)
	}
}

func TestReplaceEdgesForNode_EmptyEdges(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	now := time.Now().UTC()
	fromID := model.NoteID("from-id")

	// Insert source node
	err = UpsertNode(db, fromID, "note", "from-key", "From", "draft", now, now, "note/from-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// First add some edges
	typeName := model.TypeName("note")
	edges := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: &typeName,
				Key:  "target1",
			},
			EdgeType: "related-to",
			Span: model.Span{
				Path:      "note/from-key.Rmd",
				StartByte: 100,
				EndByte:   120,
			},
		},
	}

	err = ReplaceEdgesForNode(db, fromID, edges)
	if err != nil {
		t.Fatalf("first ReplaceEdgesForNode failed: %v", err)
	}

	// Replace with empty edges (should delete all)
	err = ReplaceEdgesForNode(db, fromID, []model.RawLink{})
	if err != nil {
		t.Fatalf("second ReplaceEdgesForNode failed: %v", err)
	}

	// Verify all edges were deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = ?", fromID).Scan(&count)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 edges after empty replace, got %d", count)
	}
}

func TestReplaceTagsForNode_EmptyTags(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	now := time.Now().UTC()
	nodeID := model.NoteID("test-id")

	// Insert node first
	err = UpsertNode(db, nodeID, "note", "test-key", "Test", "draft", now, now, "note/test-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Add tags
	tags := []string{"tag1", "tag2"}
	err = ReplaceTagsForNode(db, nodeID, tags)
	if err != nil {
		t.Fatalf("first ReplaceTagsForNode failed: %v", err)
	}

	// Replace with empty tags (should delete all)
	err = ReplaceTagsForNode(db, nodeID, []string{})
	if err != nil {
		t.Fatalf("second ReplaceTagsForNode failed: %v", err)
	}

	// Verify all tags were deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("counting tags: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tags after empty replace, got %d", count)
	}
}

func TestInsertDiagnostics_EmptyDiagnostics(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	now := time.Now().UTC()
	nodeID := model.NoteID("test-id")

	// Insert node first
	err = UpsertNode(db, nodeID, "note", "test-key", "Test", "draft", now, now, "note/test-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Add diagnostics
	diags := []model.Diagnostic{
		{
			Level:   model.DiagnosticLevelError,
			Code:    "ERROR",
			Message: "Test error",
			Span: model.Span{
				Path:      "note/test-key.Rmd",
				StartByte: 0,
				EndByte:   10,
			},
		},
	}

	err = InsertDiagnostics(db, nodeID, diags)
	if err != nil {
		t.Fatalf("first InsertDiagnostics failed: %v", err)
	}

	// Replace with empty diagnostics (should delete all)
	err = InsertDiagnostics(db, nodeID, []model.Diagnostic{})
	if err != nil {
		t.Fatalf("second InsertDiagnostics failed: %v", err)
	}

	// Verify all diagnostics were deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("counting diagnostics: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 diagnostics after empty replace, got %d", count)
	}
}

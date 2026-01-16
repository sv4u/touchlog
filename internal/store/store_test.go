package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sv4u/touchlog/internal/model"
)

func TestOpenOrCreateDB(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	dbPath := filepath.Join(tmpDir, ".touchlog", "index.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file was not created: %v", err)
	}
}

func TestApplyMigrations_CreatesSchemaV1(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	// Verify tables exist
	tables := []string{"meta", "nodes", "edges", "tags", "diagnostics"}
	for _, table := range tables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT name FROM sqlite_master 
				WHERE type='table' AND name=?
			)
		`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("checking table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s was not created", table)
		}
	}

	// Verify schema version
	var version int
	err = db.QueryRow("SELECT schema_version FROM meta").Scan(&version)
	if err != nil {
		t.Fatalf("reading schema version: %v", err)
	}
	if version != model.IndexSchemaVersion {
		t.Errorf("expected schema version %d, got %d", model.IndexSchemaVersion, version)
	}
}

func TestApplyMigrations_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer db.Close()

	// Apply migrations twice
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("first ApplyMigrations failed: %v", err)
	}
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("second ApplyMigrations failed: %v", err)
	}

	// Verify schema version is still correct
	var version int
	err = db.QueryRow("SELECT schema_version FROM meta").Scan(&version)
	if err != nil {
		t.Fatalf("reading schema version: %v", err)
	}
	if version != model.IndexSchemaVersion {
		t.Errorf("expected schema version %d, got %d", model.IndexSchemaVersion, version)
	}
}

func TestUpsertNode(t *testing.T) {
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
	nodeType := model.TypeName("note")
	key := model.Key("test-key")

	err = UpsertNode(db, nodeID, nodeType, key, "Test Title", "draft", now, now, "note/test-key.Rmd", 1234567890, 1024, "abc123")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Verify node was inserted
	var retrievedType, retrievedKey, retrievedTitle, retrievedState string
	err = db.QueryRow(`
		SELECT type, key, title, state FROM nodes WHERE id = ?
	`, nodeID).Scan(&retrievedType, &retrievedKey, &retrievedTitle, &retrievedState)
	if err != nil {
		t.Fatalf("reading node: %v", err)
	}

	if retrievedType != string(nodeType) {
		t.Errorf("expected type %q, got %q", nodeType, retrievedType)
	}
	if retrievedKey != string(key) {
		t.Errorf("expected key %q, got %q", key, retrievedKey)
	}
	if retrievedTitle != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", retrievedTitle)
	}
	if retrievedState != "draft" {
		t.Errorf("expected state 'draft', got %q", retrievedState)
	}
}

func TestUpsertNode_Update(t *testing.T) {
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
	nodeType := model.TypeName("note")
	key := model.Key("test-key")

	// Insert initial node
	err = UpsertNode(db, nodeID, nodeType, key, "Original Title", "draft", now, now, "note/test-key.Rmd", 1234567890, 1024, "abc123")
	if err != nil {
		t.Fatalf("first UpsertNode failed: %v", err)
	}

	// Update node
	updated := now.Add(time.Hour)
	err = UpsertNode(db, nodeID, nodeType, key, "Updated Title", "published", updated, updated, "note/test-key.Rmd", 1234567891, 2048, "def456")
	if err != nil {
		t.Fatalf("second UpsertNode failed: %v", err)
	}

	// Verify node was updated
	var retrievedTitle, retrievedState string
	var retrievedSize int64
	err = db.QueryRow(`
		SELECT title, state, size_bytes FROM nodes WHERE id = ?
	`, nodeID).Scan(&retrievedTitle, &retrievedState, &retrievedSize)
	if err != nil {
		t.Fatalf("reading node: %v", err)
	}

	if retrievedTitle != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", retrievedTitle)
	}
	if retrievedState != "published" {
		t.Errorf("expected state 'published', got %q", retrievedState)
	}
	if retrievedSize != 2048 {
		t.Errorf("expected size 2048, got %d", retrievedSize)
	}
}

func TestReplaceTagsForNode(t *testing.T) {
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
	tags := []string{"tag1", "tag2", "tag3"}
	err = ReplaceTagsForNode(db, nodeID, tags)
	if err != nil {
		t.Fatalf("ReplaceTagsForNode failed: %v", err)
	}

	// Verify tags
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("counting tags: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 tags, got %d", count)
	}

	// Replace with different tags
	newTags := []string{"tag4", "tag5"}
	err = ReplaceTagsForNode(db, nodeID, newTags)
	if err != nil {
		t.Fatalf("second ReplaceTagsForNode failed: %v", err)
	}

	// Verify old tags are gone and new ones are present
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("counting tags: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tags after replace, got %d", count)
	}
}

func TestReplaceEdgesForNode(t *testing.T) {
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
	toID := model.NoteID("to-id")

	// Insert nodes first
	err = UpsertNode(db, fromID, "note", "from-key", "From", "draft", now, now, "note/from-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode from failed: %v", err)
	}
	err = UpsertNode(db, toID, "note", "to-key", "To", "draft", now, now, "note/to-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode to failed: %v", err)
	}

	// Create edges
	typeName := model.TypeName("note")
	edges := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: &typeName,
				Key:  "to-key",
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
		t.Fatalf("ReplaceEdgesForNode failed: %v", err)
	}

	// Verify edge was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = ?", fromID).Scan(&count)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 edge, got %d", count)
	}

	// Replace with different edges
	newEdges := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: nil,
				Key:  "unqualified-key",
			},
			EdgeType: "depends-on",
			Span: model.Span{
				Path:      "note/from-key.Rmd",
				StartByte: 200,
				EndByte:   220,
			},
		},
	}

	err = ReplaceEdgesForNode(db, fromID, newEdges)
	if err != nil {
		t.Fatalf("second ReplaceEdgesForNode failed: %v", err)
	}

	// Verify old edge is gone and new one is present
	err = db.QueryRow("SELECT COUNT(*) FROM edges WHERE from_id = ?", fromID).Scan(&count)
	if err != nil {
		t.Fatalf("counting edges: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 edge after replace, got %d", count)
	}
}

func TestInsertDiagnostics(t *testing.T) {
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

	// Insert diagnostics
	diags := []model.Diagnostic{
		{
			Level:   model.DiagnosticLevelError,
			Code:    "PARSE_ERROR",
			Message: "Test error",
			Span: model.Span{
				Path:      "note/test-key.Rmd",
				StartByte: 0,
				EndByte:   10,
			},
		},
		{
			Level:   model.DiagnosticLevelWarn,
			Code:    "WARNING",
			Message: "Test warning",
			Span: model.Span{
				Path:      "note/test-key.Rmd",
				StartByte: 20,
				EndByte:   30,
			},
		},
	}

	err = InsertDiagnostics(db, nodeID, diags)
	if err != nil {
		t.Fatalf("InsertDiagnostics failed: %v", err)
	}

	// Verify diagnostics were inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("counting diagnostics: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 diagnostics, got %d", count)
	}

	// Replace with different diagnostics
	newDiags := []model.Diagnostic{
		{
			Level:   model.DiagnosticLevelInfo,
			Code:    "INFO",
			Message: "Test info",
			Span: model.Span{
				Path:      "note/test-key.Rmd",
				StartByte: 40,
				EndByte:   50,
			},
		},
	}

	err = InsertDiagnostics(db, nodeID, newDiags)
	if err != nil {
		t.Fatalf("second InsertDiagnostics failed: %v", err)
	}

	// Verify old diagnostics are gone and new one is present
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("counting diagnostics: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 diagnostic after replace, got %d", count)
	}
}

func TestUniqueConstraint_TypeKey(t *testing.T) {
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

	// Insert first node
	err = UpsertNode(db, "id1", "note", "test-key", "Title 1", "draft", now, now, "note/test-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("first UpsertNode failed: %v", err)
	}

	// Try to insert second node with same (type, key) - should fail
	err = UpsertNode(db, "id2", "note", "test-key", "Title 2", "draft", now, now, "note/test-key.Rmd", 0, 0, "")
	if err == nil {
		t.Error("expected error when inserting duplicate (type, key)")
	}
}

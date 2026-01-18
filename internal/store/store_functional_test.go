package store

import (
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestUpsertNode_Functional_UpdatesExisting tests UpsertNode updates existing node
func TestUpsertNode_Functional_UpdatesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	if err := ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("ApplyMigrations failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	nodeID := model.NoteID("note-1")
	nodeType := model.TypeName("note")
	key := model.Key("test-note")
	title := "Original Title"
	state := "draft"
	created := time.Now()
	updated := time.Now()
	path := "note/test-note.Rmd"

	// Insert initial node
	err = UpsertNode(db, nodeID, nodeType, key, title, state, created, updated, path, 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Update node
	newTitle := "Updated Title"
	newState := "published"
	newUpdated := time.Now()
	err = UpsertNode(db, nodeID, nodeType, key, newTitle, newState, created, newUpdated, path, 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode update failed: %v", err)
	}

	// Verify update
	var dbTitle, dbState string
	err = db.QueryRow("SELECT title, state FROM nodes WHERE id = ?", string(nodeID)).Scan(&dbTitle, &dbState)
	if err != nil {
		t.Fatalf("querying updated node: %v", err)
	}

	if dbTitle != newTitle {
		t.Errorf("expected title %q, got %q", newTitle, dbTitle)
	}
	if dbState != newState {
		t.Errorf("expected state %q, got %q", newState, dbState)
	}
}

// TestReplaceTagsForNode_Functional_ReplacesExisting tests ReplaceTagsForNode replaces existing tags
func TestReplaceTagsForNode_Functional_ReplacesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	if err := ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("ApplyMigrations failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	nodeID := model.NoteID("note-1")
	nodeType := model.TypeName("note")
	key := model.Key("test-note")
	title := "Test Note"
	state := "draft"
	now := time.Now()
	path := "note/test-note.Rmd"

	// Insert node
	err = UpsertNode(db, nodeID, nodeType, key, title, state, now, now, path, 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Add initial tags
	initialTags := []string{"tag1", "tag2"}
	err = ReplaceTagsForNode(db, nodeID, initialTags)
	if err != nil {
		t.Fatalf("ReplaceTagsForNode failed: %v", err)
	}

	// Replace with new tags
	newTags := []string{"tag3", "tag4", "tag5"}
	err = ReplaceTagsForNode(db, nodeID, newTags)
	if err != nil {
		t.Fatalf("ReplaceTagsForNode replace failed: %v", err)
	}

	// Verify only new tags exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = ?", string(nodeID)).Scan(&count)
	if err != nil {
		t.Fatalf("querying tags: %v", err)
	}

	if count != len(newTags) {
		t.Errorf("expected %d tags, got %d", len(newTags), count)
	}

	// Verify old tags are gone
	var oldTagCount int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = ? AND tag IN (?, ?)", string(nodeID), "tag1", "tag2").Scan(&oldTagCount)
	if err != nil {
		t.Fatalf("querying old tags: %v", err)
	}

	if oldTagCount != 0 {
		t.Errorf("expected 0 old tags, got %d", oldTagCount)
	}
}

// TestReplaceTagsForNode_Functional_HandlesEmptyTags tests ReplaceTagsForNode handles empty tags
func TestReplaceTagsForNode_Functional_HandlesEmptyTags(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	if err := ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("ApplyMigrations failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	nodeID := model.NoteID("note-1")
	nodeType := model.TypeName("note")
	key := model.Key("test-note")
	title := "Test Note"
	state := "draft"
	now := time.Now()
	path := "note/test-note.Rmd"

	// Insert node
	err = UpsertNode(db, nodeID, nodeType, key, title, state, now, now, path, 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Add tags
	err = ReplaceTagsForNode(db, nodeID, []string{"tag1", "tag2"})
	if err != nil {
		t.Fatalf("ReplaceTagsForNode failed: %v", err)
	}

	// Replace with empty tags
	err = ReplaceTagsForNode(db, nodeID, []string{})
	if err != nil {
		t.Fatalf("ReplaceTagsForNode with empty tags failed: %v", err)
	}

	// Verify all tags removed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE node_id = ?", string(nodeID)).Scan(&count)
	if err != nil {
		t.Fatalf("querying tags: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 tags after empty replace, got %d", count)
	}
}

// TestInsertDiagnostics_Functional_ReplacesDiagnostics tests InsertDiagnostics replaces diagnostics
func TestInsertDiagnostics_Functional_ReplacesDiagnostics(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	if err := ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("ApplyMigrations failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	nodeID := model.NoteID("note-1")
	nodeType := model.TypeName("note")
	key := model.Key("test-note")
	title := "Test Note"
	state := "draft"
	now := time.Now()
	path := "note/test-note.Rmd"

	// Insert node
	err = UpsertNode(db, nodeID, nodeType, key, title, state, now, now, path, 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Insert first set of diagnostics
	diags1 := []model.Diagnostic{
		{Level: model.DiagnosticLevelError, Code: "ERR001", Message: "Error 1"},
		{Level: model.DiagnosticLevelWarn, Code: "WARN001", Message: "Warning 1"},
	}
	err = InsertDiagnostics(db, nodeID, diags1)
	if err != nil {
		t.Fatalf("InsertDiagnostics failed: %v", err)
	}

	// Verify first set exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = ?", string(nodeID)).Scan(&count)
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}
	if count != len(diags1) {
		t.Errorf("expected %d diagnostics after first insert, got %d", len(diags1), count)
	}

	// Insert second set of diagnostics (should replace)
	diags2 := []model.Diagnostic{
		{Level: model.DiagnosticLevelInfo, Code: "INFO001", Message: "Info 1"},
	}
	err = InsertDiagnostics(db, nodeID, diags2)
	if err != nil {
		t.Fatalf("InsertDiagnostics second call failed: %v", err)
	}

	// Verify only second set exists (replaced)
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = ?", string(nodeID)).Scan(&count)
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	if count != len(diags2) {
		t.Errorf("expected %d diagnostics after replace, got %d", len(diags2), count)
	}
}

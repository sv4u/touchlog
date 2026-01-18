package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestReplaceEdgesForNode_UnresolvedLink tests ReplaceEdgesForNode with unresolved links
func TestReplaceEdgesForNode_UnresolvedLink(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	now := time.Now().UTC()
	fromID := model.NoteID("from-id")

	// Insert source node only (target doesn't exist - unresolved link)
	err = UpsertNode(db, fromID, "note", "from-key", "From", "draft", now, now, "note/from-key.Rmd", 0, 0, "")
	if err != nil {
		t.Fatalf("UpsertNode failed: %v", err)
	}

	// Create edge with unresolved link (ResolvedToID is nil)
	typeName := model.TypeName("note")
	edges := []model.RawLink{
		{
			Source: model.TypeKey{Type: "note", Key: "from-key"},
			Target: model.RawTarget{
				Type: &typeName,
				Key:  "nonexistent-key",
			},
			EdgeType:     "related-to",
			ResolvedToID: nil, // Unresolved
			Span: model.Span{
				Path:      "note/from-key.Rmd",
				StartByte: 100,
				EndByte:   120,
			},
		},
	}

	err = ReplaceEdgesForNode(db, fromID, edges)
	if err != nil {
		t.Fatalf("ReplaceEdgesForNode with unresolved link failed: %v", err)
	}

	// Verify edge was inserted with NULL to_id
	var toID sql.NullString
	err = db.QueryRow("SELECT to_id FROM edges WHERE from_id = ?", fromID).Scan(&toID)
	if err != nil {
		t.Fatalf("reading edge: %v", err)
	}
	if toID.Valid {
		t.Error("expected to_id to be NULL for unresolved link")
	}
}

// TestGetCurrentSchemaVersion_NoMetaTable tests getCurrentSchemaVersion when meta table doesn't exist
func TestGetCurrentSchemaVersion_NoMetaTable(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Don't apply migrations - meta table won't exist
	version, err := getCurrentSchemaVersion(db)
	if err != nil {
		t.Fatalf("getCurrentSchemaVersion failed: %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0 when meta table doesn't exist, got %d", version)
	}
}

// TestGetCurrentSchemaVersion_WithMetaTable tests getCurrentSchemaVersion when meta table exists
func TestGetCurrentSchemaVersion_WithMetaTable(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Apply migrations to create meta table
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	version, err := getCurrentSchemaVersion(db)
	if err != nil {
		t.Fatalf("getCurrentSchemaVersion failed: %v", err)
	}
	if version != model.IndexSchemaVersion {
		t.Errorf("expected version %d, got %d", model.IndexSchemaVersion, version)
	}
}

// TestOpenOrCreateDB_CreatesDirectory tests that OpenOrCreateDB creates the directory
func TestOpenOrCreateDB_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	vaultRoot := filepath.Join(tmpDir, "vault")

	// Directory doesn't exist yet
	if _, err := os.Stat(vaultRoot); err == nil {
		t.Fatal("vault directory should not exist")
	}

	db, err := OpenOrCreateDB(vaultRoot)
	if err != nil {
		t.Fatalf("OpenOrCreateDB failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Verify directory was created
	touchlogDir := filepath.Join(vaultRoot, ".touchlog")
	if _, err := os.Stat(touchlogDir); err != nil {
		t.Fatalf(".touchlog directory was not created: %v", err)
	}
}

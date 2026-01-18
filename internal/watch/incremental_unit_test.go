package watch

import (
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// TestIncrementalIndexer_ReplaceDiagnosticsTx tests replaceDiagnosticsTx unit behavior
func TestIncrementalIndexer_ReplaceDiagnosticsTx(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("beginning transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// First, insert a node to have a valid node_id
	nodeID := model.NoteID("note-test")
	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO nodes (id, type, key, title, state, created, updated, path, mtime_ns, size_bytes, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, nodeID, "note", "test", "Test Note", "draft", "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z", "note/test.Rmd", now.UnixNano(), 100, "")
	if err != nil {
		t.Fatalf("inserting node: %v", err)
	}

	// Test replaceDiagnosticsTx with empty diagnostics
	diags := []model.Diagnostic{}
	err = indexer.replaceDiagnosticsTx(tx, nodeID, diags)
	if err != nil {
		t.Fatalf("replaceDiagnosticsTx with empty diagnostics failed: %v", err)
	}

	// Test replaceDiagnosticsTx with diagnostics
	diags = []model.Diagnostic{
		{
			Level:   model.DiagnosticLevelWarn,
			Code:    "TEST_CODE",
			Message: "Test diagnostic message",
			Span:    model.Span{},
		},
	}
	err = indexer.replaceDiagnosticsTx(tx, nodeID, diags)
	if err != nil {
		t.Fatalf("replaceDiagnosticsTx with diagnostics failed: %v", err)
	}

	// Verify diagnostics were inserted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = ?", nodeID).Scan(&count)
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 diagnostic, got %d", count)
	}
}

// TestIncrementalIndexer_ResolveLinks_QualifiedLink tests resolveLinks with qualified links
func TestIncrementalIndexer_ResolveLinks_QualifiedLink(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Create type-key map with a target
	typeKeyMap := map[model.TypeKey]model.NoteID{
		{Type: "note", Key: "target-note"}: "note-target",
	}

	// Test qualified link that resolves
	noteType := model.TypeName("note")
	rawLinks := []model.RawLink{
		{
			Target: model.RawTarget{
				Type: &noteType,
				Key:  "target-note",
			},
			EdgeType: "related-to",
			Span:     model.Span{},
		},
	}

	resolvedEdges, diags := indexer.resolveLinks(rawLinks, typeKeyMap, noteType)

	if len(resolvedEdges) != 1 {
		t.Errorf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}
	if resolvedEdges[0].ResolvedToID == nil {
		t.Error("expected resolved edge to have ResolvedToID")
	}
	if *resolvedEdges[0].ResolvedToID != "note-target" {
		t.Errorf("expected ResolvedToID to be 'note-target', got %q", *resolvedEdges[0].ResolvedToID)
	}
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for resolved link, got %d", len(diags))
	}
}

// TestIncrementalIndexer_ResolveLinks_UnresolvedQualifiedLink tests resolveLinks with unresolved qualified link
func TestIncrementalIndexer_ResolveLinks_UnresolvedQualifiedLink(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Empty type-key map (no targets)
	typeKeyMap := map[model.TypeKey]model.NoteID{}

	// Test qualified link that doesn't resolve
	noteType := model.TypeName("note")
	rawLinks := []model.RawLink{
		{
			Target: model.RawTarget{
				Type: &noteType,
				Key:  "nonexistent",
			},
			EdgeType: "related-to",
			Span:     model.Span{},
		},
	}

	resolvedEdges, diags := indexer.resolveLinks(rawLinks, typeKeyMap, noteType)

	if len(resolvedEdges) != 1 {
		t.Errorf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}
	if resolvedEdges[0].ResolvedToID != nil {
		t.Error("expected unresolved edge to have nil ResolvedToID")
	}
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic for unresolved link, got %d", len(diags))
	}
	if diags[0].Code != "UNRESOLVED_LINK" {
		t.Errorf("expected diagnostic code 'UNRESOLVED_LINK', got %q", diags[0].Code)
	}
}

// TestIncrementalIndexer_ResolveLinks_UnqualifiedLink tests resolveLinks with unqualified link
func TestIncrementalIndexer_ResolveLinks_UnqualifiedLink(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Create type-key map with a target
	typeKeyMap := map[model.TypeKey]model.NoteID{
		{Type: "note", Key: "target-note"}: "note-target",
	}

	// Test unqualified link that resolves
	noteType := model.TypeName("note")
	rawLinks := []model.RawLink{
		{
			Target: model.RawTarget{
				Type: nil, // Unqualified
				Key:  "target-note",
			},
			EdgeType: "related-to",
			Span:     model.Span{},
		},
	}

	resolvedEdges, diags := indexer.resolveLinks(rawLinks, typeKeyMap, noteType)

	if len(resolvedEdges) != 1 {
		t.Errorf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}
	if resolvedEdges[0].ResolvedToID == nil {
		t.Error("expected resolved edge to have ResolvedToID")
	}
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for resolved link, got %d", len(diags))
	}
}

// TestIncrementalIndexer_ResolveLinks_AmbiguousLink tests resolveLinks with ambiguous unqualified link
func TestIncrementalIndexer_ResolveLinks_AmbiguousLink(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	indexer := NewIncrementalIndexer(tmpDir, cfg, db)

	// Create type-key map with multiple targets with same key
	typeKeyMap := map[model.TypeKey]model.NoteID{
		{Type: "note", Key: "ambiguous"}:     "note-ambiguous",
		{Type: "task", Key: "ambiguous"}:     "task-ambiguous",
		{Type: "decision", Key: "ambiguous"}: "decision-ambiguous",
	}

	// Test unqualified link that matches multiple types
	noteType := model.TypeName("note")
	rawLinks := []model.RawLink{
		{
			Target: model.RawTarget{
				Type: nil, // Unqualified
				Key:  "ambiguous",
			},
			EdgeType: "related-to",
			Span:     model.Span{},
		},
	}

	resolvedEdges, diags := indexer.resolveLinks(rawLinks, typeKeyMap, noteType)

	if len(resolvedEdges) != 1 {
		t.Errorf("expected 1 resolved edge, got %d", len(resolvedEdges))
	}
	// Ambiguous link should have nil ResolvedToID
	if resolvedEdges[0].ResolvedToID != nil {
		t.Error("expected ambiguous link to have nil ResolvedToID")
	}
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic for ambiguous link, got %d", len(diags))
	}
	if diags[0].Code != "AMBIGUOUS_LINK" {
		t.Errorf("expected diagnostic code 'AMBIGUOUS_LINK', got %q", diags[0].Code)
	}
}

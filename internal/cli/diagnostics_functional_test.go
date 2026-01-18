package cli

import (
	"testing"

	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// TestQueryDiagnostics_WithFilters tests queryDiagnostics with various filters
func TestQueryDiagnostics_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := store.ApplyMigrations(db); err != nil {
		t.Fatalf("applying migrations: %v", err)
	}

	// Insert test node
	nodeID := model.NoteID("note-1")
	_, err = db.Exec(`
		INSERT INTO nodes (id, type, key, title, state, created, updated, path, mtime_ns, size_bytes, hash)
		VALUES (?, 'note', 'test-note', 'Test Note', 'draft', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z', 'note/test-note.Rmd', 0, 0, '')
	`, nodeID)
	if err != nil {
		t.Fatalf("inserting node: %v", err)
	}

	// Insert test diagnostics
	diagnostics := []struct {
		nodeID  string
		level   string
		code    string
		message string
		span    string
	}{
		{"note-1", "error", "ERR001", "Test error", `{"path":"note/test-note.Rmd","start_line":1,"start_col":1}`},
		{"note-1", "warn", "WARN001", "Test warning", `{"path":"note/test-note.Rmd","start_line":2,"start_col":1}`},
		{"note-1", "info", "INFO001", "Test info", `{"path":"note/test-note.Rmd","start_line":3,"start_col":1}`},
	}

	for _, diag := range diagnostics {
		_, err = db.Exec(`
			INSERT INTO diagnostics (node_id, level, code, message, span, created_at)
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`, diag.nodeID, diag.level, diag.code, diag.message, diag.span)
		if err != nil {
			t.Fatalf("inserting diagnostic: %v", err)
		}
	}

	// Test level filter
	results, err := queryDiagnostics(db, "error", "", "")
	if err != nil {
		t.Fatalf("queryDiagnostics with level filter failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 error diagnostic, got %d", len(results))
	}

	// Test code filter
	results, err = queryDiagnostics(db, "", "", "WARN001")
	if err != nil {
		t.Fatalf("queryDiagnostics with code filter failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 warning diagnostic, got %d", len(results))
	}

	// Test node filter (qualified)
	results, err = queryDiagnostics(db, "", "note:test-note", "")
	if err != nil {
		t.Fatalf("queryDiagnostics with qualified node filter failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 diagnostics for node, got %d", len(results))
	}

	// Test node filter (unqualified)
	results, err = queryDiagnostics(db, "", "test-note", "")
	if err != nil {
		t.Fatalf("queryDiagnostics with unqualified node filter failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 diagnostics for node, got %d", len(results))
	}

	// Test all filters combined
	results, err = queryDiagnostics(db, "warn", "note:test-note", "WARN001")
	if err != nil {
		t.Fatalf("queryDiagnostics with all filters failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 diagnostic with all filters, got %d", len(results))
	}
}

// TestQueryDiagnostics_EmptyResult tests queryDiagnostics with no matching diagnostics
func TestQueryDiagnostics_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := store.ApplyMigrations(db); err != nil {
		t.Fatalf("applying migrations: %v", err)
	}

	// Query with no diagnostics in database
	results, err := queryDiagnostics(db, "", "", "")
	if err != nil {
		t.Fatalf("queryDiagnostics failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(results))
	}
}

// TestRenderDiagnostics_InvalidFormat tests renderDiagnostics with invalid format
func TestRenderDiagnostics_InvalidFormat(t *testing.T) {
	diagnostics := []DiagnosticResult{}
	err := renderDiagnostics(diagnostics, "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
	"github.com/sv4u/touchlog/internal/model"
	"github.com/sv4u/touchlog/internal/store"
)

func TestDiagnosticsCommand_ListDiagnostics(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault with index
	setupTestVaultWithIndex(t, tmpDir)

	// Create note with unresolved link
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

Links to [[nonexistent]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
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

	// Query diagnostics
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	diagnostics, err := queryDiagnostics(db, "", "", "")
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	// Should have at least one diagnostic
	if len(diagnostics) == 0 {
		t.Error("expected at least one diagnostic for unresolved link")
	}

	// Check diagnostic content
	foundUnresolved := false
	for _, diag := range diagnostics {
		if diag.Code == "UNRESOLVED_LINK" {
			foundUnresolved = true
			if diag.Level != model.DiagnosticLevelWarn {
				t.Errorf("expected UNRESOLVED_LINK to be warn level, got %s", diag.Level)
			}
			if !contains(diag.Message, "not found") {
				t.Errorf("expected diagnostic message to mention 'not found', got: %s", diag.Message)
			}
		}
	}

	if !foundUnresolved {
		t.Error("expected to find UNRESOLVED_LINK diagnostic")
	}
}

func TestDiagnosticsCommand_FilterByLevel(t *testing.T) {
	tmpDir := t.TempDir()

	setupTestVaultWithIndex(t, tmpDir)

	// Create note with unresolved link (warning)
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

Links to [[nonexistent]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
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

	// Query diagnostics filtered by level
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Filter by warn level
	warnDiags, err := queryDiagnostics(db, "warn", "", "")
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	if len(warnDiags) == 0 {
		t.Error("expected at least one warning diagnostic")
	}

	// All should be warnings
	for _, diag := range warnDiags {
		if diag.Level != "warn" {
			t.Errorf("expected all diagnostics to be warn level, got %s", diag.Level)
		}
	}

	// Filter by error level (should be empty)
	errorDiags, err := queryDiagnostics(db, "error", "", "")
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	if len(errorDiags) > 0 {
		t.Errorf("expected no error diagnostics, got %d", len(errorDiags))
	}
}

// TestDiagnosticsCommand_RenderTable tests table rendering behavior
func TestDiagnosticsCommand_RenderTable(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	// Create notes with various diagnostics
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create note with error
	errorNotePath := filepath.Join(noteDir, "error-note.Rmd")
	errorNoteContent := `---
id: note-error
type: note
key: error-note
title: Error Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Error Note
Invalid content.
`
	if err := os.WriteFile(errorNotePath, []byte(errorNoteContent), 0644); err != nil {
		t.Fatalf("writing error note: %v", err)
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

	// Query diagnostics
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	diagnostics, err := queryDiagnostics(db, "", "", "")
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	// Test table rendering
	if err := renderDiagnostics(diagnostics, "table"); err != nil {
		t.Fatalf("rendering diagnostics table: %v", err)
	}
}

// TestDiagnosticsCommand_RenderJSON tests JSON rendering behavior
func TestDiagnosticsCommand_RenderJSON(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	// Create note with diagnostic
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
Links to [[nonexistent]].
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
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

	// Query diagnostics
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	diagnostics, err := queryDiagnostics(db, "", "", "")
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}

	// Test JSON rendering
	if err := renderDiagnostics(diagnostics, "json"); err != nil {
		t.Fatalf("rendering diagnostics JSON: %v", err)
	}
}

// TestDiagnosticsCommand_RenderEmpty tests rendering with no diagnostics
func TestDiagnosticsCommand_RenderEmpty(t *testing.T) {
	// Test table rendering with empty diagnostics
	if err := renderDiagnosticsTable([]DiagnosticResult{}); err != nil {
		t.Fatalf("rendering empty diagnostics table: %v", err)
	}

	// Test JSON rendering with empty diagnostics
	if err := renderDiagnosticsJSON([]DiagnosticResult{}); err != nil {
		t.Fatalf("rendering empty diagnostics JSON: %v", err)
	}
}

// TestRenderDiagnosticsTable_WithMultipleLevels tests table rendering with multiple diagnostic levels
func TestRenderDiagnosticsTable_WithMultipleLevels(t *testing.T) {
	diagnostics := []DiagnosticResult{
		{
			NodeID:    "node-1",
			NodeType:  "note",
			NodeKey:   "test-note",
			Level:     "error",
			Code:      "PARSE_ERROR",
			Message:   "Failed to parse frontmatter",
			Span:      model.Span{Path: "note/test-note.Rmd", StartLine: 5, StartCol: 10},
			CreatedAt: "2024-01-01T00:00:00Z",
		},
		{
			NodeID:    "node-2",
			NodeType:  "note",
			NodeKey:   "another-note",
			Level:     "warn",
			Code:      "UNRESOLVED_LINK",
			Message:   "Link not found",
			Span:      model.Span{Path: "note/another-note.Rmd", StartLine: 10},
			CreatedAt: "2024-01-01T00:00:01Z",
		},
		{
			NodeID:    "node-3",
			NodeType:  "note",
			NodeKey:   "info-note",
			Level:     "info",
			Code:      "INFO",
			Message:   "Informational message",
			Span:      model.Span{Path: "note/info-note.Rmd"},
			CreatedAt: "2024-01-01T00:00:02Z",
		},
	}

	if err := renderDiagnosticsTable(diagnostics); err != nil {
		t.Fatalf("rendering diagnostics table: %v", err)
	}
}

// TestRenderDiagnosticsTable_WithSpanDetails tests table rendering with span details
func TestRenderDiagnosticsTable_WithSpanDetails(t *testing.T) {
	diagnostics := []DiagnosticResult{
		{
			NodeID:   "node-1",
			NodeType: "note",
			NodeKey:  "test-note",
			Level:    "error",
			Code:     "PARSE_ERROR",
			Message:  "Test error",
			Span: model.Span{
				Path:      "note/test-note.Rmd",
				StartLine: 5,
				StartCol:  10,
			},
		},
		{
			NodeID:   "node-2",
			NodeType: "note",
			NodeKey:  "another-note",
			Level:    "warn",
			Code:     "WARNING",
			Message:  "Test warning",
			Span: model.Span{
				Path: "note/another-note.Rmd",
			},
		},
	}

	if err := renderDiagnosticsTable(diagnostics); err != nil {
		t.Fatalf("rendering diagnostics table: %v", err)
	}
}

// TestRenderDiagnosticsJSON_WithData tests JSON rendering with actual data
func TestRenderDiagnosticsJSON_WithData(t *testing.T) {
	diagnostics := []DiagnosticResult{
		{
			NodeID:    "node-1",
			NodeType:  "note",
			NodeKey:   "test-note",
			Level:     "error",
			Code:      "PARSE_ERROR",
			Message:   "Test error",
			Span:      model.Span{Path: "note/test-note.Rmd"},
			CreatedAt: "2024-01-01T00:00:00Z",
		},
		{
			NodeID:    "node-2",
			NodeType:  "note",
			NodeKey:   "another-note",
			Level:     "warn",
			Code:      "WARNING",
			Message:   "Test warning",
			Span:      model.Span{Path: "note/another-note.Rmd"},
			CreatedAt: "2024-01-01T00:00:01Z",
		},
	}

	if err := renderDiagnosticsJSON(diagnostics); err != nil {
		t.Fatalf("rendering diagnostics JSON: %v", err)
	}
}

// TestDiagnosticsCommand_RenderInvalidFormat tests invalid format handling
func TestDiagnosticsCommand_RenderInvalidFormat(t *testing.T) {
	diagnostics := []DiagnosticResult{
		{
			NodeID:  "test-id",
			Level:   "warn",
			Code:    "TEST",
			Message: "Test message",
		},
	}

	err := renderDiagnostics(diagnostics, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	if !contains(err.Error(), "unsupported format") {
		t.Errorf("expected error about unsupported format, got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// setupTestVaultWithIndex creates a test vault with config and database schema
func setupTestVaultWithIndex(t *testing.T, tmpDir string) {
	t.Helper()

	// Create .touchlog directory
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create config file
	configPath := filepath.Join(touchlogDir, "config.yaml")
	configContent := `version: 1
types:
  note:
    description: A note
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
tags:
  preferred: []
edges:
  related-to:
    description: General relationship
templates:
  root: templates
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Create and initialize database with schema
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	if err := store.ApplyMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("applying migrations: %v", err)
	}
	_ = db.Close()
}

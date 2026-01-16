package model

import (
	"testing"
	"time"
)

func TestTypeKey(t *testing.T) {
	tk := TypeKey{
		Type: "note",
		Key:  "test-key",
	}

	if tk.Type != "note" {
		t.Errorf("expected Type to be 'note', got %q", tk.Type)
	}
	if tk.Key != "test-key" {
		t.Errorf("expected Key to be 'test-key', got %q", tk.Key)
	}
}

func TestFrontmatter(t *testing.T) {
	now := time.Now()
	fm := Frontmatter{
		ID:      "test-id",
		Type:    "note",
		Key:     "test-key",
		Created: now,
		Updated: now,
		Title:   "Test Title",
		Tags:    []string{"tag1", "tag2"},
		State:   "draft",
		Extra:   make(map[string]any),
	}

	if fm.ID != "test-id" {
		t.Errorf("expected ID to be 'test-id', got %q", fm.ID)
	}
	if len(fm.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(fm.Tags))
	}
}

func TestRawTarget(t *testing.T) {
	// Qualified target
	typeName := TypeName("note")
	target1 := RawTarget{
		Type: &typeName,
		Key:  "test-key",
	}

	if target1.Type == nil {
		t.Error("expected Type to be set for qualified target")
	}
	if *target1.Type != "note" {
		t.Errorf("expected Type to be 'note', got %q", *target1.Type)
	}

	// Unqualified target
	target2 := RawTarget{
		Type: nil,
		Key:  "test-key",
	}

	if target2.Type != nil {
		t.Error("expected Type to be nil for unqualified target")
	}
}

func TestDiagnostic(t *testing.T) {
	diag := Diagnostic{
		Level:   DiagnosticLevelError,
		Code:    "PARSE_ERROR",
		Message: "Test error",
		Span: Span{
			Path:      "test.md",
			StartByte: 0,
			EndByte:   10,
		},
	}

	if diag.Level != DiagnosticLevelError {
		t.Errorf("expected Level to be %q, got %q", DiagnosticLevelError, diag.Level)
	}
	if diag.Span.Path != "test.md" {
		t.Errorf("expected Span.Path to be 'test.md', got %q", diag.Span.Path)
	}
}

func TestConstants(t *testing.T) {
	if DefaultEdgeType != "related-to" {
		t.Errorf("expected DefaultEdgeType to be 'related-to', got %q", DefaultEdgeType)
	}

	if ConfigSchemaVersion != 1 {
		t.Errorf("expected ConfigSchemaVersion to be 1, got %d", ConfigSchemaVersion)
	}

	if IndexSchemaVersion != 1 {
		t.Errorf("expected IndexSchemaVersion to be 1, got %d", IndexSchemaVersion)
	}

	if ProtocolVersion != 1 {
		t.Errorf("expected ProtocolVersion to be 1, got %d", ProtocolVersion)
	}
}

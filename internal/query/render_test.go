package query

import (
	"strings"
	"testing"
)

// TestRenderResults_UnsupportedFormat tests rendering with unsupported format
func TestRenderResults_UnsupportedFormat(t *testing.T) {
	results := []SearchResult{
		{
			ID:    "note-1",
			Type:  "note",
			Key:   "my-note",
			Title: "My Note",
			State: "published",
		},
	}

	err := RenderResults(results, "xml")
	if err == nil {
		t.Error("RenderResults() should return error for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("error message should mention 'unsupported format', got: %v", err)
	}
}

// TestRenderResults_TableFormat tests that table format renders without error
func TestRenderResults_TableFormat(t *testing.T) {
	results := []SearchResult{
		{
			ID:    "note-1",
			Type:  "note",
			Key:   "my-note",
			Title: "My First Note",
			State: "published",
		},
		{
			ID:    "note-2",
			Type:  "note",
			Key:   "another-note",
			Title: "Another Note",
			State: "draft",
		},
	}

	err := RenderResults(results, "table")
	if err != nil {
		t.Fatalf("RenderResults() error = %v", err)
	}
}

// TestRenderResults_TableFormat_EmptyResults tests that empty results render without error
func TestRenderResults_TableFormat_EmptyResults(t *testing.T) {
	results := []SearchResult{}

	err := RenderResults(results, "table")
	if err != nil {
		t.Fatalf("RenderResults() error = %v", err)
	}
}

// TestRenderResults_TableFormat_LongTitle tests that long titles are handled
func TestRenderResults_TableFormat_LongTitle(t *testing.T) {
	results := []SearchResult{
		{
			ID:    "note-1",
			Type:  "note",
			Key:   "my-note",
			Title: "This is a very long title that should be truncated to fit in the table format",
			State: "published",
		},
	}

	err := RenderResults(results, "table")
	if err != nil {
		t.Fatalf("RenderResults() error = %v", err)
	}
}

// TestRenderResults_JSONFormat tests that JSON format renders without error
func TestRenderResults_JSONFormat(t *testing.T) {
	results := []SearchResult{
		{
			ID:    "note-1",
			Type:  "note",
			Key:   "my-note",
			Title: "My First Note",
			State: "published",
		},
		{
			ID:    "note-2",
			Type:  "note",
			Key:   "another-note",
			Title: "Another Note",
			State: "draft",
		},
	}

	err := RenderResults(results, "json")
	if err != nil {
		t.Fatalf("RenderResults() error = %v", err)
	}
}

// TestRenderResults_JSONFormat_EmptyResults tests that empty results render as JSON without error
func TestRenderResults_JSONFormat_EmptyResults(t *testing.T) {
	results := []SearchResult{}

	err := RenderResults(results, "json")
	if err != nil {
		t.Fatalf("RenderResults() error = %v", err)
	}
}

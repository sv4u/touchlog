package query

import (
	"testing"

	"github.com/sv4u/touchlog/v2/internal/graph"
)

func TestRenderBacklinks_EmptyResults(t *testing.T) {
	// Test table format with empty results - just verify it doesn't crash
	err := RenderBacklinks([]BacklinksResult{}, "note:test", "table")
	if err != nil {
		t.Fatalf("RenderBacklinks failed: %v", err)
	}
	// Function should complete without error
}

func TestRenderBacklinks_JSONEmptyResults(t *testing.T) {
	// Test JSON format with empty results - verify it doesn't crash
	err := RenderBacklinks([]BacklinksResult{}, "note:test", "json")
	if err != nil {
		t.Fatalf("RenderBacklinks failed: %v", err)
	}
	// Function should complete without error and produce valid JSON
}

func TestRenderNeighbors_EmptyResults(t *testing.T) {
	// Test table format with empty results - verify it doesn't crash
	err := RenderNeighbors([]NeighborsResult{}, "note:test", "table")
	if err != nil {
		t.Fatalf("RenderNeighbors failed: %v", err)
	}
	// Function should complete without error
}

func TestRenderPaths_EmptyResults(t *testing.T) {
	// Test table format with empty results - verify it doesn't crash
	err := RenderPaths([]PathResult{}, "note:source", "table")
	if err != nil {
		t.Fatalf("RenderPaths failed: %v", err)
	}
	// Function should complete without error
}

func TestRenderBacklinks_InvalidFormat(t *testing.T) {
	err := RenderBacklinks([]BacklinksResult{}, "note:test", "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestRenderNeighbors_InvalidFormat(t *testing.T) {
	err := RenderNeighbors([]NeighborsResult{}, "note:test", "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestRenderPaths_InvalidFormat(t *testing.T) {
	err := RenderPaths([]PathResult{}, "note:source", "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestRenderBacklinks_WithSpecialCharacters(t *testing.T) {
	// Test that special characters in node titles are handled correctly
	results := []BacklinksResult{
		{
			HopCount: 1,
			Nodes: []graph.Node{
				{
					ID:    "note-1",
					Type:  "note",
					Key:   "test-note",
					Title: "Test & Note \"with\" quotes",
				},
				{
					ID:    "note-2",
					Type:  "note",
					Key:   "target-note",
					Title: "Target Note",
				},
			},
			Edges: []graph.Edge{
				{
					EdgeType: "related-to",
				},
			},
		},
	}

	err := RenderBacklinks(results, "note:target-note", "table")
	if err != nil {
		t.Fatalf("RenderBacklinks failed: %v", err)
	}
	// Should not crash with special characters
}

func TestRenderNeighbors_JSONStructure(t *testing.T) {
	// Test JSON structure for neighbors - verify it doesn't crash
	results := []NeighborsResult{
		{
			Depth: 0,
			Nodes: []graph.Node{
				{
					ID:    "note-1",
					Type:  "note",
					Key:   "root-note",
					Title: "Root Note",
				},
			},
		},
		{
			Depth: 1,
			Nodes: []graph.Node{
				{
					ID:    "note-2",
					Type:  "note",
					Key:   "neighbor-note",
					Title: "Neighbor Note",
				},
			},
		},
	}

	err := RenderNeighbors(results, "note:root-note", "json")
	if err != nil {
		t.Fatalf("RenderNeighbors failed: %v", err)
	}
	// Should produce valid JSON structure
}

func TestRenderPaths_JSONStructure(t *testing.T) {
	// Test JSON structure for paths - verify it doesn't crash
	results := []PathResult{
		{
			Source:      "note:source",
			Destination: "note:dest",
			HopCount:    2,
			Nodes: []graph.Node{
				{ID: "note-source", Type: "note", Key: "source"},
				{ID: "note-intermediate", Type: "note", Key: "intermediate"},
				{ID: "note-dest", Type: "note", Key: "dest"},
			},
			Edges: []graph.Edge{
				{EdgeType: "related-to"},
				{EdgeType: "related-to"},
			},
		},
	}

	err := RenderPaths(results, "note:source", "json")
	if err != nil {
		t.Fatalf("RenderPaths failed: %v", err)
	}
	// Should produce valid JSON structure
}

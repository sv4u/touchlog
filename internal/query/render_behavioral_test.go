package query

import (
	"testing"

	"github.com/sv4u/touchlog/v2/internal/graph"
)

// TestRenderNeighbors_TableFormat tests rendering neighbors in table format
func TestRenderNeighbors_TableFormat(t *testing.T) {
	results := []NeighborsResult{
		{
			Depth: 1,
			Nodes: []graph.Node{
				{
					ID:    "note-1",
					Type:  "note",
					Key:   "neighbor-1",
					Title: "Neighbor One",
					State: "published",
				},
			},
		},
	}

	err := RenderNeighbors(results, "note:root", "table")
	if err != nil {
		t.Fatalf("RenderNeighbors() error = %v", err)
	}
}

// TestRenderNeighbors_TableFormat_EmptyResults tests rendering empty neighbors
func TestRenderNeighbors_TableFormat_EmptyResults(t *testing.T) {
	results := []NeighborsResult{}

	err := RenderNeighbors(results, "note:root", "table")
	if err != nil {
		t.Fatalf("RenderNeighbors() error = %v", err)
	}
}

// TestRenderNeighbors_JSONFormat tests rendering neighbors in JSON format
func TestRenderNeighbors_JSONFormat(t *testing.T) {
	results := []NeighborsResult{
		{
			Depth: 1,
			Nodes: []graph.Node{
				{
					ID:    "note-1",
					Type:  "note",
					Key:   "neighbor-1",
					Title: "Neighbor One",
					State: "published",
				},
			},
		},
	}

	err := RenderNeighbors(results, "note:root", "json")
	if err != nil {
		t.Fatalf("RenderNeighbors() error = %v", err)
	}
}

// TestRenderPaths_TableFormat tests rendering paths in table format
func TestRenderPaths_TableFormat(t *testing.T) {
	results := []PathResult{
		{
			Destination: "note:dest",
			HopCount:    2,
			Nodes: []graph.Node{
				{ID: "note-1", Type: "note", Key: "source"},
				{ID: "note-2", Type: "note", Key: "intermediate"},
				{ID: "note-3", Type: "note", Key: "dest"},
			},
			Edges: []graph.Edge{
				{EdgeType: "related-to"},
				{EdgeType: "related-to"},
			},
		},
	}

	err := RenderPaths(results, "note:source", "table")
	if err != nil {
		t.Fatalf("RenderPaths() error = %v", err)
	}
}

// TestRenderPaths_TableFormat_EmptyResults tests rendering empty paths
func TestRenderPaths_TableFormat_EmptyResults(t *testing.T) {
	results := []PathResult{}

	err := RenderPaths(results, "note:source", "table")
	if err != nil {
		t.Fatalf("RenderPaths() error = %v", err)
	}
}

// TestRenderPaths_JSONFormat tests rendering paths in JSON format
func TestRenderPaths_JSONFormat(t *testing.T) {
	results := []PathResult{
		{
			Destination: "note:dest",
			HopCount:    1,
			Nodes: []graph.Node{
				{ID: "note-1", Type: "note", Key: "source"},
				{ID: "note-2", Type: "note", Key: "dest"},
			},
			Edges: []graph.Edge{
				{EdgeType: "related-to"},
			},
		},
	}

	err := RenderPaths(results, "note:source", "json")
	if err != nil {
		t.Fatalf("RenderPaths() error = %v", err)
	}
}

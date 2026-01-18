package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sv4u/touchlog/v2/internal/store"
)

// ExportData represents the complete index export structure
type ExportData struct {
	Version string       `json:"version"`
	Nodes   []NodeExport `json:"nodes"`
	Edges   []EdgeExport `json:"edges"`
	Tags    []TagExport  `json:"tags"`
}

// NodeExport represents a node in the export
type NodeExport struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Key     string `json:"key"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Path    string `json:"path"`
}

// EdgeExport represents an edge in the export
type EdgeExport struct {
	FromID    string  `json:"from_id"`
	ToID      *string `json:"to_id,omitempty"` // nil for unresolved links
	EdgeType  string  `json:"edge_type"`
	RawTarget string  `json:"raw_target"` // JSON string of RawTarget
	Span      string  `json:"span"`       // JSON string of Span
}

// TagExport represents a tag in the export
type TagExport struct {
	NodeID string `json:"node_id"`
	Tag    string `json:"tag"`
}

// Export exports the index to JSON with deterministic ordering
func Export(vaultRoot string, outputPath string) error {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Query all nodes (ordered deterministically)
	nodes, err := queryNodes(db)
	if err != nil {
		return fmt.Errorf("querying nodes: %w", err)
	}

	// Query all edges (ordered deterministically)
	edges, err := queryEdges(db)
	if err != nil {
		return fmt.Errorf("querying edges: %w", err)
	}

	// Query all tags (ordered deterministically)
	tags, err := queryTags(db)
	if err != nil {
		return fmt.Errorf("querying tags: %w", err)
	}

	// Build export data
	exportData := ExportData{
		Version: "1",
		Nodes:   nodes,
		Edges:   edges,
		Tags:    tags,
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("writing export file: %w", err)
	}

	return nil
}

// queryNodes queries all nodes ordered by (type, key) for deterministic output
func queryNodes(db *sql.DB) ([]NodeExport, error) {
	rows, err := db.Query(`
		SELECT id, type, key, title, state, created, updated, path
		FROM nodes
		ORDER BY type, key
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var nodes []NodeExport
	for rows.Next() {
		var node NodeExport
		if err := rows.Scan(&node.ID, &node.Type, &node.Key, &node.Title, &node.State, &node.Created, &node.Updated, &node.Path); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

// queryEdges queries all edges ordered by (from_id, edge_type, to_id, raw_target) for deterministic output
func queryEdges(db *sql.DB) ([]EdgeExport, error) {
	rows, err := db.Query(`
		SELECT from_id, to_id, edge_type, raw_target, span
		FROM edges
		ORDER BY from_id, edge_type, to_id, raw_target
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var edges []EdgeExport
	for rows.Next() {
		var edge EdgeExport
		var toID sql.NullString
		if err := rows.Scan(&edge.FromID, &toID, &edge.EdgeType, &edge.RawTarget, &edge.Span); err != nil {
			return nil, err
		}
		if toID.Valid {
			edge.ToID = &toID.String
		}
		edges = append(edges, edge)
	}

	return edges, rows.Err()
}

// queryTags queries all tags ordered by (node_id, tag) for deterministic output
func queryTags(db *sql.DB) ([]TagExport, error) {
	rows, err := db.Query(`
		SELECT node_id, tag
		FROM tags
		ORDER BY node_id, tag
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var tags []TagExport
	for rows.Next() {
		var tag TagExport
		if err := rows.Scan(&tag.NodeID, &tag.Tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

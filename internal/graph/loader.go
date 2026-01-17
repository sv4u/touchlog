package graph

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/sv4u/touchlog/internal/model"
	"github.com/sv4u/touchlog/internal/store"
)

// Edge represents a graph edge in memory
type Edge struct {
	FromID    model.NoteID
	ToID      *model.NoteID // nil for unresolved links
	EdgeType  model.EdgeType
	RawTarget model.RawTarget
	Span      model.Span
}

// Node represents a graph node in memory
type Node struct {
	ID      model.NoteID
	Type    model.TypeName
	Key     model.Key
	Title   string
	State   string
	Created string
	Updated string
	Path    string
}

// Graph represents a loaded subgraph
type Graph struct {
	Nodes         map[model.NoteID]*Node
	OutgoingEdges map[model.NoteID][]Edge // fromID -> []Edge
	IncomingEdges map[model.NoteID][]Edge // toID -> []Edge
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		Nodes:         make(map[model.NoteID]*Node),
		OutgoingEdges: make(map[model.NoteID][]Edge),
		IncomingEdges: make(map[model.NoteID][]Edge),
	}
}

// LoadGraph loads the complete graph from SQLite
func LoadGraph(vaultRoot string) (*Graph, error) {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	graph := NewGraph()

	// Load all nodes
	if err := loadNodes(db, graph); err != nil {
		return nil, fmt.Errorf("loading nodes: %w", err)
	}

	// Load all edges
	if err := loadEdges(db, graph); err != nil {
		return nil, fmt.Errorf("loading edges: %w", err)
	}

	return graph, nil
}

// LoadSubgraph loads a subgraph for specific node IDs
func LoadSubgraph(vaultRoot string, nodeIDs []model.NoteID) (*Graph, error) {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	graph := NewGraph()

	// Load specific nodes
	if err := loadNodesByIDs(db, graph, nodeIDs); err != nil {
		return nil, fmt.Errorf("loading nodes: %w", err)
	}

	// Load edges where at least one endpoint is in the subgraph
	if err := loadEdgesForNodes(db, graph, nodeIDs); err != nil {
		return nil, fmt.Errorf("loading edges: %w", err)
	}

	return graph, nil
}

// loadNodes loads all nodes from the database
func loadNodes(db *sql.DB, graph *Graph) error {
	rows, err := db.Query(`
		SELECT id, type, key, title, state, created, updated, path
		FROM nodes
		ORDER BY type, key
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var node Node
		if err := rows.Scan(&node.ID, &node.Type, &node.Key, &node.Title, &node.State, &node.Created, &node.Updated, &node.Path); err != nil {
			return err
		}
		graph.Nodes[node.ID] = &node
	}

	return rows.Err()
}

// loadNodesByIDs loads specific nodes by ID
func loadNodesByIDs(db *sql.DB, graph *Graph, nodeIDs []model.NoteID) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// Build query with placeholders (optimized for large node sets)
	placeholders := make([]string, len(nodeIDs))
	args := make([]interface{}, len(nodeIDs))
	for i, id := range nodeIDs {
		placeholders[i] = "?"
		args[i] = string(id)
	}

	// Use efficient IN clause with proper indexing
	query := fmt.Sprintf(`
		SELECT id, type, key, title, state, created, updated, path
		FROM nodes
		WHERE id IN (%s)
		ORDER BY type, key
	`, joinPlaceholders(placeholders))

	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var node Node
		if err := rows.Scan(&node.ID, &node.Type, &node.Key, &node.Title, &node.State, &node.Created, &node.Updated, &node.Path); err != nil {
			return err
		}
		graph.Nodes[node.ID] = &node
	}

	return rows.Err()
}

// joinPlaceholders joins placeholders with commas
func joinPlaceholders(placeholders []string) string {
	result := ""
	for i, p := range placeholders {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// loadEdges loads all edges from the database
func loadEdges(db *sql.DB, graph *Graph) error {
	rows, err := db.Query(`
		SELECT from_id, to_id, edge_type, raw_target, span
		FROM edges
		ORDER BY from_id, edge_type, to_id, raw_target
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var edge Edge
		var toID sql.NullString
		var rawTargetJSON, spanJSON string

		if err := rows.Scan(&edge.FromID, &toID, &edge.EdgeType, &rawTargetJSON, &spanJSON); err != nil {
			return err
		}

		if toID.Valid {
			toIDVal := model.NoteID(toID.String)
			edge.ToID = &toIDVal
		}

		// Parse raw target and span from JSON
		if err := json.Unmarshal([]byte(rawTargetJSON), &edge.RawTarget); err != nil {
			// Skip edges with invalid JSON
			continue
		}
		if err := json.Unmarshal([]byte(spanJSON), &edge.Span); err != nil {
			// Skip edges with invalid JSON
			continue
		}

		// Add to outgoing edges
		graph.OutgoingEdges[edge.FromID] = append(graph.OutgoingEdges[edge.FromID], edge)

		// Add to incoming edges if resolved
		if edge.ToID != nil {
			graph.IncomingEdges[*edge.ToID] = append(graph.IncomingEdges[*edge.ToID], edge)
		}
	}

	// Sort edges deterministically
	sortEdges(graph.OutgoingEdges)
	sortEdges(graph.IncomingEdges)

	return rows.Err()
}

// loadEdgesForNodes loads edges where at least one endpoint is in the node set
func loadEdgesForNodes(db *sql.DB, graph *Graph, nodeIDs []model.NoteID) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// Build query
	placeholders := make([]string, len(nodeIDs))
	args := make([]interface{}, len(nodeIDs)*2) // for from_id and to_id
	for i, id := range nodeIDs {
		placeholders[i] = "?"
		args[i] = string(id)
		args[i+len(nodeIDs)] = string(id)
	}

	query := fmt.Sprintf(`
		SELECT from_id, to_id, edge_type, raw_target, span
		FROM edges
		WHERE from_id IN (%s) OR to_id IN (%s)
		ORDER BY from_id, edge_type, to_id, raw_target
	`, joinPlaceholders(placeholders), joinPlaceholders(placeholders))

	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var edge Edge
		var toID sql.NullString
		var rawTargetJSON, spanJSON string

		if err := rows.Scan(&edge.FromID, &toID, &edge.EdgeType, &rawTargetJSON, &spanJSON); err != nil {
			return err
		}

		if toID.Valid {
			toIDVal := model.NoteID(toID.String)
			edge.ToID = &toIDVal
		}

		// Parse raw target and span from JSON
		if err := json.Unmarshal([]byte(rawTargetJSON), &edge.RawTarget); err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(spanJSON), &edge.Span); err != nil {
			continue
		}

		// Add to outgoing edges
		graph.OutgoingEdges[edge.FromID] = append(graph.OutgoingEdges[edge.FromID], edge)

		// Add to incoming edges if resolved
		if edge.ToID != nil {
			graph.IncomingEdges[*edge.ToID] = append(graph.IncomingEdges[*edge.ToID], edge)
		}
	}

	// Sort edges deterministically
	sortEdges(graph.OutgoingEdges)
	sortEdges(graph.IncomingEdges)

	return rows.Err()
}

// sortEdges sorts edges deterministically by (edge_type, to_id, raw_target, span)
func sortEdges(edgeMap map[model.NoteID][]Edge) {
	for nodeID := range edgeMap {
		edges := edgeMap[nodeID]
		sort.Slice(edges, func(i, j int) bool {
			// Compare by edge_type
			if edges[i].EdgeType != edges[j].EdgeType {
				return edges[i].EdgeType < edges[j].EdgeType
			}

			// Compare by to_id (nil comes last)
			if edges[i].ToID == nil && edges[j].ToID != nil {
				return false
			}
			if edges[i].ToID != nil && edges[j].ToID == nil {
				return true
			}
			if edges[i].ToID != nil && edges[j].ToID != nil {
				if *edges[i].ToID != *edges[j].ToID {
					return *edges[i].ToID < *edges[j].ToID
				}
			}

			// Compare by raw_target (as JSON string for simplicity)
			targetI, _ := json.Marshal(edges[i].RawTarget)
			targetJ, _ := json.Marshal(edges[j].RawTarget)
			if string(targetI) != string(targetJ) {
				return string(targetI) < string(targetJ)
			}

			// Compare by span (as JSON string)
			spanI, _ := json.Marshal(edges[i].Span)
			spanJ, _ := json.Marshal(edges[j].Span)
			return string(spanI) < string(spanJ)
		})
		edgeMap[nodeID] = edges
	}
}

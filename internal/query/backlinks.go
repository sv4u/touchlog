package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/graph"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// BacklinksResult represents a single backlink path
type BacklinksResult struct {
	HopCount int          `json:"hop_count"`
	Nodes    []graph.Node `json:"nodes"`
	Edges    []graph.Edge `json:"edges"`
}

// ExecuteBacklinks executes a backlinks query
func ExecuteBacklinks(vaultRoot string, q *BacklinksQuery) ([]BacklinksResult, error) {
	// Resolve target node
	targetID, err := resolveNodeID(vaultRoot, q.Target)
	if err != nil {
		return nil, fmt.Errorf("resolving target node: %w", err)
	}

	// Load graph
	g, err := graph.LoadGraph(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("loading graph: %w", err)
	}

	// Get edges based on direction
	var edges []graph.Edge
	switch q.Direction {
	case "in":
		edges = g.IncomingEdges[targetID]
	case "out":
		edges = g.OutgoingEdges[targetID]
	case "both":
		edges = append(g.IncomingEdges[targetID], g.OutgoingEdges[targetID]...)
	default:
		return nil, fmt.Errorf("invalid direction: %s (must be 'in', 'out', or 'both')", q.Direction)
	}

	// Filter by edge types if specified
	if len(q.EdgeTypes) > 0 {
		edgeTypeSet := make(map[model.EdgeType]bool)
		for _, et := range q.EdgeTypes {
			edgeTypeSet[model.EdgeType(et)] = true
		}

		filtered := make([]graph.Edge, 0, len(edges))
		for _, edge := range edges {
			if edgeTypeSet[edge.EdgeType] {
				filtered = append(filtered, edge)
			}
		}
		edges = filtered
	}

	// Build results (one hop only - no traversal)
	results := make([]BacklinksResult, 0, len(edges))
	seenPaths := make(map[string]bool) // Track unique paths

	for _, edge := range edges {
		// Determine source and target based on direction
		var sourceID, targetIDInPath model.NoteID
		if q.Direction == "in" || (q.Direction == "both" && edge.ToID != nil && *edge.ToID == targetID) {
			// Incoming edge
			sourceID = edge.FromID
			targetIDInPath = targetID
		} else {
			// Outgoing edge
			sourceID = targetID
			if edge.ToID == nil {
				continue // Skip unresolved links for outgoing
			}
			targetIDInPath = *edge.ToID
		}

		// Create path key for deduplication
		pathKey := fmt.Sprintf("%s->%s", sourceID, targetIDInPath)
		if seenPaths[pathKey] {
			continue
		}
		seenPaths[pathKey] = true

		// Get source node
		sourceNode := g.Nodes[sourceID]
		if sourceNode == nil {
			continue // Skip if source node not found
		}

		// Get target node
		targetNode := g.Nodes[targetIDInPath]
		if targetNode == nil {
			continue // Skip if target node not found
		}

		// Build result
		result := BacklinksResult{
			HopCount: 1, // Backlinks are always one hop
			Nodes:    []graph.Node{*sourceNode, *targetNode},
			Edges:    []graph.Edge{edge},
		}

		results = append(results, result)
	}

	// Sort results lexicographically by (type, key) of source node
	sort.Slice(results, func(i, j int) bool {
		if len(results[i].Nodes) == 0 || len(results[j].Nodes) == 0 {
			return false
		}
		sourceI := results[i].Nodes[0]
		sourceJ := results[j].Nodes[0]

		// Compare by type
		if sourceI.Type != sourceJ.Type {
			return sourceI.Type < sourceJ.Type
		}

		// Compare by key
		return sourceI.Key < sourceJ.Key
	})

	return results, nil
}

// resolveNodeID resolves a node identifier (type:key or just key) to a NoteID
// For unqualified identifiers, first tries exact match on full key, then falls back
// to last-segment matching for path-based key support
func resolveNodeID(vaultRoot string, identifier string) (model.NoteID, error) {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return "", fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Parse identifier
	parts := strings.Split(identifier, ":")
	var nodeType *string
	var key string

	if len(parts) == 2 {
		// Qualified: type:key
		nodeTypeStr := parts[0]
		nodeType = &nodeTypeStr
		key = parts[1]
	} else if len(parts) == 1 {
		// Unqualified: key
		key = parts[0]
	} else {
		return "", fmt.Errorf("invalid node identifier format: %s (expected 'type:key' or 'key')", identifier)
	}

	if nodeType != nil {
		// Qualified lookup - exact match on (type, key)
		var id string
		err := db.QueryRow("SELECT id FROM nodes WHERE type = ? AND key = ?", *nodeType, key).Scan(&id)
		if err != nil {
			return "", fmt.Errorf("node not found: %s:%s", *nodeType, key)
		}
		return model.NoteID(id), nil
	}

	// Unqualified lookup - first try exact match on full key, then fall back to last-segment matching
	rows, err := db.Query("SELECT id, type, key FROM nodes")
	if err != nil {
		return "", fmt.Errorf("querying nodes: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var exactMatchIDs []model.NoteID
	var exactMatchKeys []string
	var lastSegMatchIDs []model.NoteID
	var lastSegMatchKeys []string

	searchLastSeg := config.LastSegment(key)

	for rows.Next() {
		var id, typ, nodeKey string
		if err := rows.Scan(&id, &typ, &nodeKey); err != nil {
			return "", err
		}

		// Check for exact match on full key first
		if nodeKey == key {
			exactMatchIDs = append(exactMatchIDs, model.NoteID(id))
			exactMatchKeys = append(exactMatchKeys, fmt.Sprintf("%s:%s", typ, nodeKey))
		}

		// Also collect last-segment matches for fallback
		nodeLastSeg := config.LastSegment(nodeKey)
		if nodeLastSeg == searchLastSeg {
			lastSegMatchIDs = append(lastSegMatchIDs, model.NoteID(id))
			lastSegMatchKeys = append(lastSegMatchKeys, fmt.Sprintf("%s:%s", typ, nodeKey))
		}
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	// Priority 1: Exact match on full key
	if len(exactMatchIDs) == 1 {
		return exactMatchIDs[0], nil
	}
	if len(exactMatchIDs) > 1 {
		// Multiple exact matches (same key in different types)
		return "", fmt.Errorf("ambiguous node identifier '%s': matches %d notes (%s). Use qualified identifier (type:key)", key, len(exactMatchIDs), strings.Join(exactMatchKeys, ", "))
	}

	// Priority 2: Fall back to last-segment matching
	if len(lastSegMatchIDs) == 0 {
		return "", fmt.Errorf("node not found: %s", identifier)
	}
	if len(lastSegMatchIDs) == 1 {
		return lastSegMatchIDs[0], nil
	}

	// Multiple last-segment matches
	return "", fmt.Errorf("ambiguous node identifier '%s': matches %d notes (%s). Use qualified identifier (type:full/path/key)", key, len(lastSegMatchIDs), strings.Join(lastSegMatchKeys, ", "))
}

package query

import (
	"fmt"
	"sort"

	"github.com/sv4u/touchlog/v2/internal/graph"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// PathResult represents a path between source and destination
type PathResult struct {
	Source      string       `json:"source"`
	Destination string       `json:"destination"`
	HopCount    int          `json:"hop_count"`
	Nodes       []graph.Node `json:"nodes"`
	Edges       []graph.Edge `json:"edges"`
}

// ExecutePaths executes a paths query using BFS shortest-path
func ExecutePaths(vaultRoot string, q *PathsQuery) ([]PathResult, error) {
	// Validate max_depth is set
	if q.MaxDepth <= 0 {
		return nil, fmt.Errorf("max_depth is required and must be > 0")
	}

	// Resolve source node
	sourceID, err := resolveNodeID(vaultRoot, q.Source)
	if err != nil {
		return nil, fmt.Errorf("resolving source node: %w", err)
	}

	// Resolve destination nodes
	destIDs := make([]model.NoteID, 0, len(q.Destinations))
	for _, dest := range q.Destinations {
		destID, err := resolveNodeID(vaultRoot, dest)
		if err != nil {
			return nil, fmt.Errorf("resolving destination node %s: %w", dest, err)
		}
		destIDs = append(destIDs, destID)
	}

	// Handle src == dst case (zero-hop path)
	results := make([]PathResult, 0)
	for _, destID := range destIDs {
		if sourceID == destID {
			// Zero-hop path
			g, err := graph.LoadGraph(vaultRoot)
			if err != nil {
				return nil, fmt.Errorf("loading graph: %w", err)
			}
			node := g.Nodes[sourceID]
			if node != nil {
				results = append(results, PathResult{
					Source:      q.Source,
					Destination: q.Destinations[0], // Use first destination string
					HopCount:    0,
					Nodes:       []graph.Node{*node},
					Edges:       []graph.Edge{},
				})
			}
			continue
		}
	}

	// Load graph
	g, err := graph.LoadGraph(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("loading graph: %w", err)
	}

	// Find paths to each destination
	for i, destID := range destIDs {
		if sourceID == destID {
			continue // Already handled zero-hop case
		}

		paths := bfsShortestPaths(g, sourceID, destID, q.MaxDepth, q.MaxPaths, q.Direction, q.EdgeTypes)

		// Convert to PathResult
		for _, path := range paths {
			results = append(results, PathResult{
				Source:      q.Source,
				Destination: q.Destinations[i],
				HopCount:    len(path.Edges),
				Nodes:       path.Nodes,
				Edges:       path.Edges,
			})
		}
	}

	// Sort results by destination, then by path (BFS discovery order preserved)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Destination != results[j].Destination {
			return results[i].Destination < results[j].Destination
		}
		// Same destination - preserve BFS order (already sorted by discovery)
		return i < j
	})

	return results, nil
}

// Path represents a path through the graph
type Path struct {
	Nodes []graph.Node
	Edges []graph.Edge
}

// bfsShortestPaths finds shortest paths from source to destination using BFS
func bfsShortestPaths(g *graph.Graph, sourceID, destID model.NoteID, maxDepth, maxPaths int, direction string, edgeTypes []string) []Path {
	// Build edge type set for filtering
	edgeTypeSet := make(map[model.EdgeType]bool)
	for _, et := range edgeTypes {
		edgeTypeSet[model.EdgeType(et)] = true
	}

	// Visited set for cycle detection (mandatory)
	visited := make(map[model.NoteID]bool)
	paths := make([]Path, 0)

	// Initialize queue with source path
	type queueItem struct {
		nodeID model.NoteID
		path   Path
		depth  int
	}

	queue := []queueItem{{
		nodeID: sourceID,
		path: Path{
			Nodes: []graph.Node{*g.Nodes[sourceID]},
			Edges: []graph.Edge{},
		},
		depth: 0,
	}}
	visited[sourceID] = true

	// BFS traversal
	for len(queue) > 0 && len(paths) < maxPaths {
		item := queue[0]
		queue = queue[1:]

		// Check if we reached destination (before depth check, so we can include paths at maxDepth)
		if item.nodeID == destID {
			paths = append(paths, item.path)
			continue // Continue to find more paths if maxPaths not reached
		}

		if item.depth >= maxDepth {
			continue // Don't traverse beyond max depth
		}

		// Get neighbors based on direction
		var edges []graph.Edge
		switch direction {
		case "in":
			edges = g.IncomingEdges[item.nodeID]
		case "out":
			edges = g.OutgoingEdges[item.nodeID]
		case "both":
			edges = append(g.IncomingEdges[item.nodeID], g.OutgoingEdges[item.nodeID]...)
		default:
			continue
		}

		// Process edges
		for _, edge := range edges {
			// Filter by edge type if specified
			if len(edgeTypeSet) > 0 && !edgeTypeSet[edge.EdgeType] {
				continue
			}

			// Determine neighbor ID
			var neighborID model.NoteID
			if direction == "in" || (direction == "both" && edge.ToID != nil && *edge.ToID == item.nodeID) {
				// Incoming edge - neighbor is the source
				neighborID = edge.FromID
			} else {
				// Outgoing edge - neighbor is the target
				if edge.ToID == nil {
					continue // Skip unresolved links
				}
				neighborID = *edge.ToID
			}

			// Skip if already visited in this path (cycle detection)
			// For shortest paths, we allow revisiting nodes in different paths
			// but not in the same path
			nodeInPath := false
			for _, node := range item.path.Nodes {
				if node.ID == neighborID {
					nodeInPath = true
					break
				}
			}
			if nodeInPath {
				continue // Cycle detected in this path
			}

			// Get neighbor node
			neighborNode := g.Nodes[neighborID]
			if neighborNode == nil {
				continue
			}

			// Create new path
			newPath := Path{
				Nodes: make([]graph.Node, len(item.path.Nodes)+1),
				Edges: make([]graph.Edge, len(item.path.Edges)+1),
			}
			copy(newPath.Nodes, item.path.Nodes)
			newPath.Nodes[len(newPath.Nodes)-1] = *neighborNode
			copy(newPath.Edges, item.path.Edges)
			newPath.Edges[len(newPath.Edges)-1] = edge

			// Add to queue
			queue = append(queue, queueItem{
				nodeID: neighborID,
				path:   newPath,
				depth:  item.depth + 1,
			})
		}
	}

	// Sort paths lexicographically (canonical ordering)
	sort.Slice(paths, func(i, j int) bool {
		// Compare by node sequence
		for k := 0; k < len(paths[i].Nodes) && k < len(paths[j].Nodes); k++ {
			if paths[i].Nodes[k].ID != paths[j].Nodes[k].ID {
				return paths[i].Nodes[k].ID < paths[j].Nodes[k].ID
			}
		}
		return len(paths[i].Nodes) < len(paths[j].Nodes)
	})

	return paths
}

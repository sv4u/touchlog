package query

import (
	"fmt"
	"sort"

	"github.com/sv4u/touchlog/internal/graph"
	"github.com/sv4u/touchlog/internal/model"
)

// NeighborsResult represents neighbors at a specific depth
type NeighborsResult struct {
	Depth int          `json:"depth"`
	Nodes []graph.Node `json:"nodes"`
}

// ExecuteNeighbors executes a neighbors query using BFS
func ExecuteNeighbors(vaultRoot string, q *NeighborsQuery) ([]NeighborsResult, error) {
	// Validate max_depth is set
	if q.MaxDepth <= 0 {
		return nil, fmt.Errorf("max_depth is required and must be > 0")
	}

	// Resolve root node
	rootID, err := resolveNodeID(vaultRoot, q.Root)
	if err != nil {
		return nil, fmt.Errorf("resolving root node: %w", err)
	}

	// Load graph
	g, err := graph.LoadGraph(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("loading graph: %w", err)
	}

	// Verify root node exists
	if g.Nodes[rootID] == nil {
		return nil, fmt.Errorf("root node not found: %s", q.Root)
	}

	// BFS traversal with depth tracking
	results := bfsNeighbors(g, rootID, q.MaxDepth, q.Direction, q.EdgeTypes)

	// Apply node filters if specified
	if q.NodeFilters != nil {
		results = filterNeighborsByNodeFilters(vaultRoot, results, q.NodeFilters)
	}

	return results, nil
}

// bfsNeighbors performs BFS traversal to find neighbors
func bfsNeighbors(g *graph.Graph, rootID model.NoteID, maxDepth int, direction string, edgeTypes []string) []NeighborsResult {
	// Build edge type set for filtering
	edgeTypeSet := make(map[model.EdgeType]bool)
	for _, et := range edgeTypes {
		edgeTypeSet[model.EdgeType(et)] = true
	}

	// Visited set for cycle detection (mandatory)
	visited := make(map[model.NoteID]int) // nodeID -> depth first seen
	results := make([]NeighborsResult, 0, maxDepth+1)

	// Initialize queue with root at depth 0
	type queueItem struct {
		nodeID model.NoteID
		depth  int
	}
	queue := []queueItem{{nodeID: rootID, depth: 0}}
	visited[rootID] = 0

	// Track nodes by depth
	nodesByDepth := make(map[int][]graph.Node)

	// Add root node at depth 0
	rootNode := g.Nodes[rootID]
	if rootNode != nil {
		nodesByDepth[0] = []graph.Node{*rootNode}
	}

	// BFS traversal
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

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

			// Skip if already visited at this or earlier depth
			if seenDepth, seen := visited[neighborID]; seen && seenDepth <= item.depth+1 {
				continue
			}

			// Mark as visited
			visited[neighborID] = item.depth + 1

			// Get neighbor node
			neighborNode := g.Nodes[neighborID]
			if neighborNode == nil {
				continue
			}

			// Add to depth group
			depth := item.depth + 1
			if nodesByDepth[depth] == nil {
				nodesByDepth[depth] = make([]graph.Node, 0)
			}
			nodesByDepth[depth] = append(nodesByDepth[depth], *neighborNode)

			// Add to queue for further traversal
			queue = append(queue, queueItem{nodeID: neighborID, depth: depth})
		}
	}

	// Build results sorted by depth
	for depth := 0; depth <= maxDepth; depth++ {
		if nodes, ok := nodesByDepth[depth]; ok && len(nodes) > 0 {
			// Sort nodes lexicographically by (type, key)
			sortNodesByTypeKey(nodes)
			results = append(results, NeighborsResult{
				Depth: depth,
				Nodes: nodes,
			})
		}
	}

	return results
}

// sortNodesByTypeKey sorts nodes lexicographically by (type, key)
func sortNodesByTypeKey(nodes []graph.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		// Compare by type
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type < nodes[j].Type
		}
		// Compare by key
		return nodes[i].Key < nodes[j].Key
	})
}

// filterNeighborsByNodeFilters applies node filters to neighbor results
func filterNeighborsByNodeFilters(vaultRoot string, results []NeighborsResult, filters *SearchQuery) []NeighborsResult {
	// For Phase 4, we'll implement basic filtering
	// In a full implementation, we'd apply all filter criteria
	filtered := make([]NeighborsResult, 0, len(results))

	for _, result := range results {
		filteredNodes := make([]graph.Node, 0)

		for _, node := range result.Nodes {
			// Apply type filter
			if len(filters.Types) > 0 {
				typeMatch := false
				for _, filterType := range filters.Types {
					if string(node.Type) == filterType {
						typeMatch = true
						break
					}
				}
				if !typeMatch {
					continue
				}
			}

			// Apply state filter
			if len(filters.States) > 0 {
				stateMatch := false
				for _, filterState := range filters.States {
					if node.State == filterState {
						stateMatch = true
						break
					}
				}
				if !stateMatch {
					continue
				}
			}

			filteredNodes = append(filteredNodes, node)
		}

		if len(filteredNodes) > 0 {
			filtered = append(filtered, NeighborsResult{
				Depth: result.Depth,
				Nodes: filteredNodes,
			})
		}
	}

	return filtered
}

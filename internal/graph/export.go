package graph

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sv4u/touchlog/internal/model"
)

// ExportOptions represents options for graph export
type ExportOptions struct {
	Roots      []string // Root nodes (type:key or key) - empty means all nodes
	Types      []string // Filter by types
	Tags       []string // Filter by tags
	States     []string // Filter by states
	EdgeTypes  []string // Filter by edge types
	Depth      int      // Maximum depth (default: 10)
	Force      bool     // Overwrite existing file
}

// ExportDOT exports the graph to DOT format
func ExportDOT(vaultRoot string, outputPath string, opts ExportOptions) error {
	// Check if file exists and force is not set
	if !opts.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", outputPath)
		}
	}

	// Load graph
	g, err := LoadGraph(vaultRoot)
	if err != nil {
		return fmt.Errorf("loading graph: %w", err)
	}

	// Determine depth (default: 10)
	depth := opts.Depth
	if depth == 0 {
		depth = 10
	}

	// Determine which nodes to include
	includedNodes := determineIncludedNodes(g, opts)

	// Build DOT output
	var dot strings.Builder
	dot.WriteString("digraph touchlog {\n")
	dot.WriteString("  rankdir=LR;\n")
	dot.WriteString("  node [shape=box];\n\n")

	// Output nodes
	nodeIDs := make([]model.NoteID, 0, len(includedNodes))
	for id := range includedNodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Slice(nodeIDs, func(i, j int) bool {
		return nodeIDs[i] < nodeIDs[j]
	})

	for _, id := range nodeIDs {
		node := g.Nodes[id]
		if node == nil {
			continue
		}

		// Escape node label for DOT
		label := escapeDOTString(fmt.Sprintf("%s:%s\\n%s", node.Type, node.Key, node.Title))
		dot.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", id, label))
	}

	dot.WriteString("\n")

	// Output edges (only where at least one endpoint is included)
	edgeSet := make(map[string]bool) // Track unique edges
	for _, id := range nodeIDs {
		// Outgoing edges
		for _, edge := range g.OutgoingEdges[id] {
			if edge.ToID == nil {
				continue // Skip unresolved links
			}

			// Include edge if at least one endpoint is included
			if includedNodes[id] || includedNodes[*edge.ToID] {
				edgeKey := fmt.Sprintf("%s->%s", id, *edge.ToID)
				if !edgeSet[edgeKey] {
					edgeSet[edgeKey] = true
					label := escapeDOTString(string(edge.EdgeType))
					dot.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", id, *edge.ToID, label))
				}
			}
		}

		// Incoming edges (to avoid duplicates, we only process outgoing above)
		// But we need to handle cases where only the target is included
		for _, edge := range g.IncomingEdges[id] {
			if !includedNodes[edge.FromID] {
				continue // Source not included
			}

			edgeKey := fmt.Sprintf("%s->%s", edge.FromID, id)
			if !edgeSet[edgeKey] {
				edgeSet[edgeKey] = true
				label := escapeDOTString(string(edge.EdgeType))
				dot.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", edge.FromID, id, label))
			}
		}
	}

	dot.WriteString("}\n")

	// Write to file
	if err := os.WriteFile(outputPath, []byte(dot.String()), 0644); err != nil {
		return fmt.Errorf("writing DOT file: %w", err)
	}

	return nil
}

// determineIncludedNodes determines which nodes should be included in the export
func determineIncludedNodes(g *Graph, opts ExportOptions) map[model.NoteID]bool {
	included := make(map[model.NoteID]bool)

	// Build filter sets
	typeSet := make(map[model.TypeName]bool)
	for _, t := range opts.Types {
		typeSet[model.TypeName(t)] = true
	}

	stateSet := make(map[string]bool)
	for _, s := range opts.States {
		stateSet[s] = true
	}

	// If roots specified, start from roots and traverse
	if len(opts.Roots) > 0 {
		// For Phase 4, we'll include all nodes that match filters
		// In a full implementation, we'd do BFS from roots
		for id, node := range g.Nodes {
			if matchesFilters(node, typeSet, stateSet, opts.Tags) {
				included[id] = true
			}
		}
	} else {
		// No roots - include all nodes that match filters
		for id, node := range g.Nodes {
			if matchesFilters(node, typeSet, stateSet, opts.Tags) {
				included[id] = true
			}
		}
	}

	return included
}

// matchesFilters checks if a node matches the filters
func matchesFilters(node *Node, typeSet map[model.TypeName]bool, stateSet map[string]bool, tags []string) bool {
	// Type filter
	if len(typeSet) > 0 && !typeSet[node.Type] {
		return false
	}

	// State filter
	if len(stateSet) > 0 && !stateSet[node.State] {
		return false
	}

	// Tag filter (would require loading tags from database)
	// For Phase 4, we'll skip tag filtering for now
	_ = tags

	return true
}

// escapeDOTString escapes special characters for DOT format
func escapeDOTString(s string) string {
	// Escape quotes and backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

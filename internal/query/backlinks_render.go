package query

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RenderBacklinks renders backlinks results in the specified format
func RenderBacklinks(results []BacklinksResult, target string, format string) error {
	switch format {
	case "table":
		return renderBacklinksTable(results, target)
	case "json":
		return renderBacklinksJSON(results, target)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderBacklinksTable renders backlinks as a table
func renderBacklinksTable(results []BacklinksResult, target string) error {
	if len(results) == 0 {
		fmt.Printf("No backlinks found for %s\n", target)
		return nil
	}

	// Print header
	fmt.Printf("Backlinks for %s:\n\n", target)
	fmt.Printf("%-30s %s\n", "Path", "Edge Type")
	fmt.Println(strings.Repeat("-", 60))

	// Print rows
	for _, result := range results {
		if len(result.Nodes) < 2 || len(result.Edges) == 0 {
			continue
		}

		sourceNode := result.Nodes[0]
		targetNode := result.Nodes[1]
		edge := result.Edges[0]

		// Format path with inline edge type
		path := fmt.Sprintf("%s:%s -[%s]-> %s:%s",
			sourceNode.Type, sourceNode.Key,
			edge.EdgeType,
			targetNode.Type, targetNode.Key)

		fmt.Printf("%-30s %s\n", path, string(edge.EdgeType))
	}

	fmt.Printf("\nFound %d backlink(s)\n", len(results))
	return nil
}

// renderBacklinksJSON renders backlinks as JSON
func renderBacklinksJSON(results []BacklinksResult, target string) error {
	output := map[string]interface{}{
		"schema_version":   1,
		"touchlog_version": "0.0.0",
		"query": map[string]interface{}{
			"target":    target,
			"direction": "in", // Default, will be normalized
		},
		"target":  target,
		"results": results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

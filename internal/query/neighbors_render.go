package query

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RenderNeighbors renders neighbors results in the specified format
func RenderNeighbors(results []NeighborsResult, root string, format string) error {
	switch format {
	case "table":
		return renderNeighborsTable(results, root)
	case "json":
		return renderNeighborsJSON(results, root)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderNeighborsTable renders neighbors as a table grouped by depth
func renderNeighborsTable(results []NeighborsResult, root string) error {
	if len(results) == 0 {
		fmt.Printf("No neighbors found for %s\n", root)
		return nil
	}

	fmt.Printf("Neighbors of %s:\n\n", root)

	// Group by depth
	for _, result := range results {
		if result.Depth == 0 {
			// Skip root node in output
			continue
		}

		fmt.Printf("Depth: %d\n", result.Depth)
		fmt.Printf("%-20s %-15s %-30s\n", "ID", "Type", "Key")
		fmt.Println(strings.Repeat("-", 65))

		for _, node := range result.Nodes {
			displayTitle := node.Title
			if len(displayTitle) > 28 {
				displayTitle = displayTitle[:25] + "..."
			}
			_ = displayTitle // Reserved for future use in table output
			fmt.Printf("%-20s %-15s %-30s\n", node.ID, node.Type, node.Key)
		}

		fmt.Println()
	}

	totalNodes := 0
	for _, result := range results {
		if result.Depth > 0 {
			totalNodes += len(result.Nodes)
		}
	}

	fmt.Printf("Found %d neighbor(s) across %d depth level(s)\n", totalNodes, len(results)-1)
	return nil
}

// renderNeighborsJSON renders neighbors as JSON
func renderNeighborsJSON(results []NeighborsResult, root string) error {
	// Build layers structure
	layers := make([]map[string]interface{}, 0, len(results))
	for _, result := range results {
		layer := map[string]interface{}{
			"depth": result.Depth,
			"nodes": result.Nodes,
		}
		layers = append(layers, layer)
	}

	output := map[string]interface{}{
		"schema_version":   1,
		"touchlog_version": "0.0.0",
		"query": map[string]interface{}{
			"root":      root,
			"direction": "both", // Default, will be normalized
			"max_depth": 0,      // Will be normalized
		},
		"root": root,
		"results": map[string]interface{}{
			"layers": layers,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

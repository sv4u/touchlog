package query

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// RenderPaths renders paths results in the specified format
func RenderPaths(results []PathResult, source string, format string) error {
	switch format {
	case "table":
		return renderPathsTable(results, source)
	case "json":
		return renderPathsJSON(results, source)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderPathsTable renders paths as a table grouped by destination
func renderPathsTable(results []PathResult, source string) error {
	if len(results) == 0 {
		fmt.Printf("No paths found from %s\n", source)
		return nil
	}

	// Group by destination
	byDestination := make(map[string][]PathResult)
	for _, result := range results {
		byDestination[result.Destination] = append(byDestination[result.Destination], result)
	}

	// Sort destinations
	destinations := make([]string, 0, len(byDestination))
	for dest := range byDestination {
		destinations = append(destinations, dest)
	}
	sort.Strings(destinations)

	// Print grouped by destination
	for _, dest := range destinations {
		paths := byDestination[dest]
		fmt.Printf("Destination: %s\n", dest)
		fmt.Printf("%-5s %s\n", "Hops", "Path")
		fmt.Println(strings.Repeat("-", 80))

		for _, path := range paths {
			// Build path string with inline edge types
			pathParts := make([]string, 0, len(path.Nodes))
			for i, node := range path.Nodes {
				pathParts = append(pathParts, fmt.Sprintf("%s:%s", node.Type, node.Key))
				if i < len(path.Edges) {
					pathParts = append(pathParts, fmt.Sprintf("-[%s]->", path.Edges[i].EdgeType))
				}
			}
			pathStr := strings.Join(pathParts, " ")

			fmt.Printf("%-5d %s\n", path.HopCount, pathStr)
		}

		fmt.Println()
	}

	return nil
}

// renderPathsJSON renders paths as JSON
func renderPathsJSON(results []PathResult, source string) error {
	// Group by destination
	byDestination := make(map[string][]PathResult)
	for _, result := range results {
		byDestination[result.Destination] = append(byDestination[result.Destination], result)
	}

	// Build results array
	resultsArray := make([]map[string]interface{}, 0, len(byDestination))
	destinations := make([]string, 0, len(byDestination))
	for dest := range byDestination {
		destinations = append(destinations, dest)
	}
	sort.Strings(destinations)

	for _, dest := range destinations {
		paths := byDestination[dest]
		pathObjects := make([]map[string]interface{}, 0, len(paths))
		for _, path := range paths {
			pathObjects = append(pathObjects, map[string]interface{}{
				"hop_count": path.HopCount,
				"nodes":     path.Nodes,
				"edges":     path.Edges,
			})
		}

		resultsArray = append(resultsArray, map[string]interface{}{
			"dst":   dest,
			"paths": pathObjects,
		})
	}

	output := map[string]interface{}{
		"schema_version":   1,
		"touchlog_version": "0.0.0",
		"query": map[string]interface{}{
			"source":    source,
			"direction": "both", // Default, will be normalized
			"max_depth": 0,      // Will be normalized
		},
		"source":  source,
		"results": resultsArray,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

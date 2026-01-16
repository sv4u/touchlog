package query

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RenderResults renders search results in the specified format
func RenderResults(results []SearchResult, format string) error {
	switch format {
	case "table":
		return renderTable(results)
	case "json":
		return renderJSON(results)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderTable renders results as a table
func renderTable(results []SearchResult) error {
	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	// Print header
	fmt.Printf("%-20s %-15s %-20s %-30s %-10s\n", "ID", "Type", "Key", "Title", "State")
	fmt.Println(strings.Repeat("-", 95))

	// Print rows
	for _, result := range results {
		title := result.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}
		fmt.Printf("%-20s %-15s %-20s %-30s %-10s\n",
			result.ID, result.Type, result.Key, title, result.State)
	}

	fmt.Printf("\nFound %d result(s)\n", len(results))
	return nil
}

// renderJSON renders results as JSON
func renderJSON(results []SearchResult) error {
	output := map[string]interface{}{
		"results": results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

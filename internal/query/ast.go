package query

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSearchQuery parses a query string into a SearchQuery.
// Supported key:value pairs (space-separated):
//
//	type:note           - filter by type (comma-separated for multiple)
//	state:published     - filter by state (comma-separated for multiple)
//	tag:important       - filter by tag (comma-separated for multiple)
//	match-any-tag:true  - match any tag instead of all
//	limit:10            - limit number of results
//	offset:5            - offset for pagination
//	format:json         - output format (table or json)
//
// An empty query string returns a default SearchQuery (no filters).
func ParseSearchQuery(queryStr string) (*SearchQuery, error) {
	q := NewSearchQuery()
	queryStr = strings.TrimSpace(queryStr)
	if queryStr == "" {
		return q, nil
	}

	tokens := strings.Fields(queryStr)
	for _, token := range tokens {
		parts := strings.SplitN(token, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid query token %q: expected key:value format", token)
		}

		key, value := parts[0], parts[1]
		switch key {
		case "type":
			q.Types = splitCSV(value)
		case "state":
			q.States = splitCSV(value)
		case "tag":
			q.Tags = splitCSV(value)
		case "match-any-tag":
			q.MatchAnyTag = value == "true" || value == "1"
		case "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid limit %q: %w", value, err)
			}
			q.Limit = &n
		case "offset":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid offset %q: %w", value, err)
			}
			q.Offset = n
		case "format":
			if value != "table" && value != "json" {
				return nil, fmt.Errorf("invalid format %q: must be 'table' or 'json'", value)
			}
			q.Format = value
		default:
			return nil, fmt.Errorf("unknown query key %q", key)
		}
	}

	return q, nil
}

// splitCSV splits a comma-separated string into trimmed, non-empty parts.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// SearchQuery represents a search query AST
type SearchQuery struct {
	Types       []string // Filter by types (CSV)
	States      []string // Filter by states (CSV)
	Tags        []string // Filter by tags (CSV)
	MatchAnyTag bool     // If true, match any tag; if false, match all tags
	Limit       *int     // Limit number of results
	Offset      int      // Offset for pagination
	Format      string   // Output format: "table" or "json"
}

// NewSearchQuery creates a new search query with defaults
func NewSearchQuery() *SearchQuery {
	return &SearchQuery{
		Format:      "table",
		MatchAnyTag: false, // Default: match all tags
		Offset:      0,
	}
}

// BacklinksQuery represents a backlinks query
type BacklinksQuery struct {
	Target    string   // Target node (type:key or just key)
	Direction string   // "in", "out", or "both" (default: "in")
	EdgeTypes []string // Filter by edge types (inclusive)
	Format    string   // Output format: "table" or "json"
}

// NewBacklinksQuery creates a new backlinks query with defaults
func NewBacklinksQuery() *BacklinksQuery {
	return &BacklinksQuery{
		Direction: "in", // Default: incoming links only
		Format:    "table",
	}
}

// NeighborsQuery represents a neighbors query
type NeighborsQuery struct {
	Root        string       // Root node (type:key or just key)
	Direction   string       // "in", "out", or "both" (default: "both")
	MaxDepth    int          // Maximum depth (required, no default)
	EdgeTypes   []string     // Filter by edge types (inclusive)
	NodeFilters *SearchQuery // Node filters for result set
	Format      string       // Output format: "table" or "json"
}

// NewNeighborsQuery creates a new neighbors query with defaults
func NewNeighborsQuery() *NeighborsQuery {
	return &NeighborsQuery{
		Direction: "both", // Default: both directions
		Format:    "table",
	}
}

// PathsQuery represents a paths query
type PathsQuery struct {
	Source       string       // Source node (type:key or just key)
	Destinations []string     // Destination nodes (type:key or just key) - supports multiple
	Direction    string       // "in", "out", or "both" (default: "both")
	MaxDepth     int          // Maximum depth (required, no default)
	MaxPaths     int          // Maximum number of paths per destination (default: 10)
	EdgeTypes    []string     // Filter by edge types (inclusive)
	NodeFilters  *SearchQuery // Node filters for result set
	Format       string       // Output format: "table" or "json"
}

// NewPathsQuery creates a new paths query with defaults
func NewPathsQuery() *PathsQuery {
	return &PathsQuery{
		Direction: "both", // Default: both directions
		MaxPaths:  10,     // Default: 10 paths per destination
		Format:    "table",
	}
}

package query

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

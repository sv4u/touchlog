package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/sv4u/touchlog/internal/query"
	cli3 "github.com/urfave/cli/v3"
)

// BuildQueryCommand builds the query command
func BuildQueryCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "query",
		Usage: "Query the note index",
		Commands: []*cli3.Command{
			{
				Name:  "backlinks",
				Usage: "Find backlinks to a node",
				Description: "Find all nodes that link to the specified target node.\n\n" +
					"Examples:\n" +
					"  touchlog query backlinks --target note:my-note\n" +
					"  touchlog query backlinks --target my-note --direction both\n" +
					"  touchlog query backlinks --target note:article --edge-type references --format json",
				Flags: []cli3.Flag{
					&cli3.StringFlag{
						Name:     "target",
						Usage:    "Target node (type:key or key)",
						Required: true,
					},
					&cli3.StringFlag{
						Name:  "direction",
						Usage: "Link direction (in|out|both)",
						Value: "in",
					},
					&cli3.StringFlag{
						Name:  "edge-type",
						Usage: "Filter by edge types (comma-separated)",
					},
					&cli3.StringFlag{
						Name:  "format",
						Usage: "Output format (table|json)",
						Value: "table",
					},
				},
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					// Validate vault exists
					if err := ValidateVault(vaultRoot); err != nil {
						return err
					}

					// Build query
					q := query.NewBacklinksQuery()
					q.Target = cmd.String("target")
					q.Direction = cmd.String("direction")
					q.Format = cmd.String("format")

					if edgeTypeFlag := cmd.String("edge-type"); edgeTypeFlag != "" {
						q.EdgeTypes = parseCSV(edgeTypeFlag)
					}

					// Execute query
					results, err := query.ExecuteBacklinks(vaultRoot, q)
					if err != nil {
						return fmt.Errorf("failed to find backlinks for %q: %w\nHint: Make sure the target node exists and the vault index is up to date (run 'touchlog index rebuild')", q.Target, err)
					}

					// Render results
					if err := query.RenderBacklinks(results, q.Target, q.Format); err != nil {
						return fmt.Errorf("rendering results: %w", err)
					}

					return nil
				},
			},
			{
				Name:  "search",
				Usage: "Search for notes",
				Description: "Search the note index with filters for type, state, tags, and pagination.\n\n" +
					"Examples:\n" +
					"  touchlog query search --type note\n" +
					"  touchlog query search --state published --tag important\n" +
					"  touchlog query search --tag todo --match-any-tag --limit 10 --format json",
				Flags: []cli3.Flag{
					&cli3.StringFlag{
						Name:  "type",
						Usage: "Filter by types (comma-separated)",
					},
					&cli3.StringFlag{
						Name:  "state",
						Usage: "Filter by states (comma-separated)",
					},
					&cli3.StringFlag{
						Name:  "tag",
						Usage: "Filter by tags (comma-separated)",
					},
					&cli3.BoolFlag{
						Name:  "match-any-tag",
						Usage: "Match any tag (default: match all tags)",
					},
					&cli3.IntFlag{
						Name:  "limit",
						Usage: "Limit number of results",
					},
					&cli3.IntFlag{
						Name:  "offset",
						Usage: "Offset for pagination",
						Value: 0,
					},
					&cli3.StringFlag{
						Name:  "format",
						Usage: "Output format (table|json)",
						Value: "table",
					},
				},
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					// Validate vault exists
					if err := ValidateVault(vaultRoot); err != nil {
						return err
					}

					// Build query from flags
					q := query.NewSearchQuery()

					if typeFlag := cmd.String("type"); typeFlag != "" {
						q.Types = parseCSV(typeFlag)
					}

					if stateFlag := cmd.String("state"); stateFlag != "" {
						q.States = parseCSV(stateFlag)
					}

					if tagFlag := cmd.String("tag"); tagFlag != "" {
						q.Tags = parseCSV(tagFlag)
					}

					if cmd.Bool("match-any-tag") {
						q.MatchAnyTag = true
					}

					if limitFlag := cmd.Int("limit"); limitFlag > 0 {
						q.Limit = &limitFlag
					}

					q.Offset = cmd.Int("offset")
					q.Format = cmd.String("format")

					// Execute search
					results, err := query.ExecuteSearch(vaultRoot, q)
					if err != nil {
						return fmt.Errorf("search failed: %w\nHint: Make sure the vault index is up to date (run 'touchlog index rebuild')", err)
					}

					// Render results
					if err := query.RenderResults(results, q.Format); err != nil {
						return fmt.Errorf("rendering results: %w", err)
					}

					return nil
				},
			},
			{
				Name:  "neighbors",
				Usage: "Find neighbors of a node",
				Description: "Find all nodes within a specified depth from the root node using breadth-first search.\n\n" +
					"Examples:\n" +
					"  touchlog query neighbors --root note:my-note --max-depth 2\n" +
					"  touchlog query neighbors --root my-note --max-depth 3 --direction out\n" +
					"  touchlog query neighbors --root note:article --max-depth 1 --edge-type references",
				Flags: []cli3.Flag{
					&cli3.StringFlag{
						Name:     "root",
						Usage:    "Root node (type:key or key)",
						Required: true,
					},
					&cli3.IntFlag{
						Name:     "max-depth",
						Usage:    "Maximum depth (required)",
						Required: true,
					},
					&cli3.StringFlag{
						Name:  "direction",
						Usage: "Link direction (in|out|both)",
						Value: "both",
					},
					&cli3.StringFlag{
						Name:  "edge-type",
						Usage: "Filter by edge types (comma-separated)",
					},
					&cli3.StringFlag{
						Name:  "format",
						Usage: "Output format (table|json)",
						Value: "table",
					},
				},
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					// Validate vault exists
					if err := ValidateVault(vaultRoot); err != nil {
						return err
					}

					// Build query
					q := query.NewNeighborsQuery()
					q.Root = cmd.String("root")
					q.MaxDepth = cmd.Int("max-depth")
					q.Direction = cmd.String("direction")
					q.Format = cmd.String("format")

					if edgeTypeFlag := cmd.String("edge-type"); edgeTypeFlag != "" {
						q.EdgeTypes = parseCSV(edgeTypeFlag)
					}

					// Execute query
					results, err := query.ExecuteNeighbors(vaultRoot, q)
					if err != nil {
						return fmt.Errorf("failed to find neighbors for %q: %w\nHint: Make sure the root node exists and max-depth is greater than 0", q.Root, err)
					}

					// Render results
					if err := query.RenderNeighbors(results, q.Root, q.Format); err != nil {
						return fmt.Errorf("rendering results: %w", err)
					}

					return nil
				},
			},
			{
				Name:  "paths",
				Usage: "Find paths between nodes",
				Description: "Find shortest paths from a source node to one or more destination nodes.\n\n" +
					"Examples:\n" +
					"  touchlog query paths --source note:start --destination note:end --max-depth 5\n" +
					"  touchlog query paths --source start --destination end1 --destination end2 --max-depth 3\n" +
					"  touchlog query paths --source note:article --destination note:reference --max-depth 10 --max-paths 5",
				Flags: []cli3.Flag{
					&cli3.StringFlag{
						Name:     "source",
						Usage:    "Source node (type:key or key)",
						Required: true,
					},
					&cli3.StringSliceFlag{
						Name:     "destination",
						Usage:    "Destination node(s) (type:key or key) - can be specified multiple times",
						Required: true,
					},
					&cli3.IntFlag{
						Name:     "max-depth",
						Usage:    "Maximum depth (required)",
						Required: true,
					},
					&cli3.IntFlag{
						Name:  "max-paths",
						Usage: "Maximum number of paths per destination",
						Value: 10,
					},
					&cli3.StringFlag{
						Name:  "direction",
						Usage: "Link direction (in|out|both)",
						Value: "both",
					},
					&cli3.StringFlag{
						Name:  "edge-type",
						Usage: "Filter by edge types (comma-separated)",
					},
					&cli3.StringFlag{
						Name:  "format",
						Usage: "Output format (table|json)",
						Value: "table",
					},
				},
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					// Validate vault exists
					if err := ValidateVault(vaultRoot); err != nil {
						return err
					}

					// Build query
					q := query.NewPathsQuery()
					q.Source = cmd.String("source")
					q.Destinations = cmd.StringSlice("destination")
					q.MaxDepth = cmd.Int("max-depth")
					q.MaxPaths = cmd.Int("max-paths")
					q.Direction = cmd.String("direction")
					q.Format = cmd.String("format")

					if edgeTypeFlag := cmd.String("edge-type"); edgeTypeFlag != "" {
						q.EdgeTypes = parseCSV(edgeTypeFlag)
					}

					// Execute query
					results, err := query.ExecutePaths(vaultRoot, q)
					if err != nil {
						return fmt.Errorf("failed to find paths from %q: %w\nHint: Make sure source and destination nodes exist, and max-depth is sufficient", q.Source, err)
					}

					// Render results
					if err := query.RenderPaths(results, q.Source, q.Format); err != nil {
						return fmt.Errorf("rendering results: %w", err)
					}

					return nil
				},
			},
		},
	}
}

// parseCSV parses a comma-separated value string into a slice
func parseCSV(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

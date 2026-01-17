package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/sv4u/touchlog/internal/graph"
	cli3 "github.com/urfave/cli/v3"
)

// BuildGraphCommand builds the graph command
func BuildGraphCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "graph",
		Usage: "Graph operations",
		Commands: []*cli3.Command{
			{
				Name:  "export",
				Usage: "Export graph",
				Commands: []*cli3.Command{
					{
						Name:  "dot",
						Usage: "Export graph as DOT format",
						Description: "Export the knowledge graph to Graphviz DOT format for visualization.\n\n" +
							"The exported graph can be rendered using tools like:\n" +
							"  - dot -Tpng graph.dot -o graph.png\n" +
							"  - dot -Tsvg graph.dot -o graph.svg\n\n" +
							"Examples:\n" +
							"  touchlog graph export dot --out graph.dot\n" +
							"  touchlog graph export dot --out graph.dot --root note:start --depth 3\n" +
							"  touchlog graph export dot --out graph.dot --type note --state published",
						Flags: []cli3.Flag{
							&cli3.StringFlag{
								Name:     "out",
								Usage:    "Output file path (required)",
								Required: true,
							},
							&cli3.StringSliceFlag{
								Name:  "root",
								Usage: "Root node(s) (type:key or key) - can be specified multiple times",
							},
							&cli3.StringFlag{
								Name:  "type",
								Usage: "Filter by types (comma-separated)",
							},
							&cli3.StringFlag{
								Name:  "tag",
								Usage: "Filter by tags (comma-separated)",
							},
							&cli3.StringFlag{
								Name:  "state",
								Usage: "Filter by states (comma-separated)",
							},
							&cli3.StringFlag{
								Name:  "edge-type",
								Usage: "Filter by edge types (comma-separated)",
							},
							&cli3.IntFlag{
								Name:  "depth",
								Usage: "Maximum depth",
								Value: 10,
							},
							&cli3.BoolFlag{
								Name:  "force",
								Usage: "Overwrite existing file",
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

							// Build options
							opts := graph.ExportOptions{
								Roots: cmd.StringSlice("root"),
								Depth: cmd.Int("depth"),
								Force: cmd.Bool("force"),
							}

							if typeFlag := cmd.String("type"); typeFlag != "" {
								opts.Types = parseCSV(typeFlag)
							}

							if tagFlag := cmd.String("tag"); tagFlag != "" {
								opts.Tags = parseCSV(tagFlag)
							}

							if stateFlag := cmd.String("state"); stateFlag != "" {
								opts.States = parseCSV(stateFlag)
							}

							if edgeTypeFlag := cmd.String("edge-type"); edgeTypeFlag != "" {
								opts.EdgeTypes = parseCSV(edgeTypeFlag)
							}

							outputPath := cmd.String("out")
							if outputPath == "" {
								return fmt.Errorf("--out is required")
							}

							// Export graph
							fmt.Fprintf(os.Stderr, "Exporting graph to %s...\n", outputPath)
							if err := graph.ExportDOT(vaultRoot, outputPath, opts); err != nil {
								return fmt.Errorf("failed to export graph: %w\nHint: Make sure the output directory exists and is writable, or use --force to overwrite existing files", err)
							}

							fmt.Fprintf(os.Stderr, "Graph exported to %s\n", outputPath)
							return nil
						},
					},
				},
			},
		},
	}
}

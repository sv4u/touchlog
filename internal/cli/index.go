package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	cli3 "github.com/urfave/cli/v3"
)

// BuildIndexCommand builds the index command
func BuildIndexCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "index",
		Usage: "Manage the note index",
		Commands: []*cli3.Command{
			{
				Name:  "rebuild",
				Usage: "Rebuild the index from scratch",
				Description: "Rebuild the entire note index by scanning all .Rmd files in the vault.\n\n" +
					"This operation:\n" +
					"  - Scans all type directories for .Rmd files\n" +
					"  - Parses frontmatter and extracts links\n" +
					"  - Resolves all links between notes\n" +
					"  - Atomically replaces the existing index\n\n" +
					"Examples:\n" +
					"  touchlog index rebuild\n" +
					"  touchlog --vault /path/to/vault index rebuild",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					// Validate vault exists
					if err := ValidateVault(vaultRoot); err != nil {
						return err
					}

					// Load config
					cfg, err := config.LoadConfig(vaultRoot)
					if err != nil {
						return fmt.Errorf("failed to load config: %w\nHint: Make sure you're in a touchlog vault directory or specify --vault", err)
					}

					// Build index with progress
					fmt.Fprintf(os.Stderr, "Rebuilding index...\n")
					builder := index.NewBuilder(vaultRoot, cfg)
					if err := builder.Rebuild(); err != nil {
						return fmt.Errorf("failed to rebuild index: %w\nHint: Check that all .Rmd files are valid and the vault structure is correct", err)
					}

					fmt.Fprintf(os.Stderr, "Index rebuilt successfully\n")
					return nil
				},
			},
			{
				Name:  "export",
				Usage: "Export the index to JSON",
				Description: "Export the entire index to a JSON file for backup or analysis.\n\n" +
					"Examples:\n" +
					"  touchlog index export --out index.json\n" +
					"  touchlog index export --out backup/index-$(date +%Y%m%d).json",
				Flags: []cli3.Flag{
					&cli3.StringFlag{
						Name:     "out",
						Usage:    "Output file path (required)",
						Required: true,
					},
					&cli3.StringFlag{
						Name:  "format",
						Usage: "Export format (json)",
						Value: "json",
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

					format := cmd.String("format")
					if format != "json" {
						return fmt.Errorf("unsupported format: %s (only 'json' is supported)", format)
					}

					outputPath := cmd.String("out")
					if outputPath == "" {
						return fmt.Errorf("--out is required")
					}

					// Export index
					fmt.Fprintf(os.Stderr, "Exporting index...\n")
					if err := index.Export(vaultRoot, outputPath); err != nil {
						return fmt.Errorf("failed to export index: %w\nHint: Make sure the output directory exists and is writable", err)
					}

					fmt.Fprintf(os.Stderr, "Index exported to %s\n", outputPath)
					return nil
				},
			},
		},
	}
}

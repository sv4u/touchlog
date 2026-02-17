package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
	"github.com/sv4u/touchlog/v2/internal/version"
	cli3 "github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

// BuildDiagnosticsCommand builds the diagnostics command
func BuildDiagnosticsCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "diagnostics",
		Usage: "View and manage diagnostics",
		Description: "View parse errors, warnings, and informational messages for notes in the vault.\n\n" +
			"Diagnostics are generated during note parsing and link resolution.\n" +
			"They help identify issues like missing frontmatter, invalid links, or parsing errors.\n\n" +
			"Examples:\n" +
			"  touchlog diagnostics list\n" +
			"  touchlog diagnostics list --level error\n" +
			"  touchlog diagnostics list --node note:my-note\n" +
			"  touchlog diagnostics list --format json",
		Commands: []*cli3.Command{
			{
				Name:  "list",
				Usage: "List diagnostics",
				Flags: []cli3.Flag{
					&cli3.StringFlag{
						Name:  "level",
						Usage: "Filter by level (info|warn|error)",
					},
					&cli3.StringFlag{
						Name:  "node",
						Usage: "Filter by node (type:key or key)",
					},
					&cli3.StringFlag{
						Name:  "code",
						Usage: "Filter by diagnostic code",
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

					// Open database
					db, err := store.OpenOrCreateDB(vaultRoot)
					if err != nil {
						return fmt.Errorf("opening database: %w", err)
					}
					defer func() {
						_ = db.Close()
					}()

					// Build query
					levelFilter := cmd.String("level")
					nodeFilter := cmd.String("node")
					codeFilter := cmd.String("code")
					format := cmd.String("format")

					// Query diagnostics
					diagnostics, err := queryDiagnostics(db, levelFilter, nodeFilter, codeFilter)
					if err != nil {
						return fmt.Errorf("querying diagnostics: %w", err)
					}

					// Render results
					if err := renderDiagnostics(diagnostics, format); err != nil {
						return fmt.Errorf("rendering diagnostics: %w", err)
					}

					return nil
				},
			},
		},
	}
}

// DiagnosticResult represents a diagnostic with node information
type DiagnosticResult struct {
	NodeID    string     `json:"node_id"`
	NodeType  string     `json:"node_type,omitempty"`
	NodeKey   string     `json:"node_key,omitempty"`
	Level     string     `json:"level"`
	Code      string     `json:"code"`
	Message   string     `json:"message"`
	Span      model.Span `json:"span"`
	CreatedAt string     `json:"created_at,omitempty"`
}

// queryDiagnostics queries diagnostics from the database
func queryDiagnostics(db *sql.DB, levelFilter, nodeFilter, codeFilter string) ([]DiagnosticResult, error) {
	// Build query
	query := `
		SELECT d.node_id, d.level, d.code, d.message, d.span, d.created_at,
		       n.type, n.key
		FROM diagnostics d
		LEFT JOIN nodes n ON d.node_id = n.id
		WHERE 1=1
	`
	args := []interface{}{}

	if levelFilter != "" {
		query += " AND d.level = ?"
		args = append(args, levelFilter)
	}

	if codeFilter != "" {
		query += " AND d.code = ?"
		args = append(args, codeFilter)
	}

	if nodeFilter != "" {
		// Parse node filter (type:key or key)
		parts := strings.Split(nodeFilter, ":")
		if len(parts) == 2 {
			// Qualified: type:key
			query += " AND n.type = ? AND n.key = ?"
			args = append(args, parts[0], parts[1])
		} else {
			// Unqualified: key
			query += " AND n.key = ?"
			args = append(args, parts[0])
		}
	}

	query += " ORDER BY d.level DESC, d.created_at DESC, d.node_id"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []DiagnosticResult
	for rows.Next() {
		var result DiagnosticResult
		var spanJSON string
		var nodeType, nodeKey sql.NullString

		if err := rows.Scan(&result.NodeID, &result.Level, &result.Code, &result.Message, &spanJSON, &result.CreatedAt, &nodeType, &nodeKey); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if nodeType.Valid {
			result.NodeType = nodeType.String
		}
		if nodeKey.Valid {
			result.NodeKey = nodeKey.String
		}

		// Parse span JSON
		if err := json.Unmarshal([]byte(spanJSON), &result.Span); err != nil {
			// Continue even if span parsing fails
			result.Span = model.Span{Path: "unknown"}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return results, nil
}

// renderDiagnostics renders diagnostics in the specified format
func renderDiagnostics(diagnostics []DiagnosticResult, format string) error {
	switch format {
	case "table":
		return renderDiagnosticsTable(diagnostics)
	case "json":
		return renderDiagnosticsJSON(diagnostics)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderDiagnosticsTable renders diagnostics as a table
func renderDiagnosticsTable(diagnostics []DiagnosticResult) error {
	if len(diagnostics) == 0 {
		fmt.Println("No diagnostics found.")
		return nil
	}

	// Group by level
	byLevel := make(map[string][]DiagnosticResult)
	for _, diag := range diagnostics {
		byLevel[diag.Level] = append(byLevel[diag.Level], diag)
	}

	// Sort levels: error, warn, info
	levels := []string{"error", "warn", "info"}
	for _, level := range levels {
		if diags, ok := byLevel[level]; ok {
			// Sort by node_id for consistency
			sort.Slice(diags, func(i, j int) bool {
				return diags[i].NodeID < diags[j].NodeID
			})

			fmt.Printf("\n%s (%d):\n", strings.ToUpper(level), len(diags))
			fmt.Println(strings.Repeat("-", 80))

			for _, diag := range diags {
				nodeRef := diag.NodeID
				if diag.NodeType != "" && diag.NodeKey != "" {
					nodeRef = fmt.Sprintf("%s:%s", diag.NodeType, diag.NodeKey)
				}

				fmt.Printf("  [%s] %s\n", diag.Code, diag.Message)
				fmt.Printf("    Node: %s\n", nodeRef)
				if diag.Span.Path != "" {
					fmt.Printf("    Path: %s", diag.Span.Path)
					if diag.Span.StartLine > 0 {
						fmt.Printf(":%d", diag.Span.StartLine)
						if diag.Span.StartCol > 0 {
							fmt.Printf(":%d", diag.Span.StartCol)
						}
					}
					fmt.Println()
				}
				fmt.Println()
			}
		}
	}

	// Summary
	total := len(diagnostics)
	errorCount := len(byLevel["error"])
	warnCount := len(byLevel["warn"])
	infoCount := len(byLevel["info"])

	fmt.Printf("\nSummary: %d total (%d errors, %d warnings, %d info)\n", total, errorCount, warnCount, infoCount)
	return nil
}

// renderDiagnosticsJSON renders diagnostics as JSON
func renderDiagnosticsJSON(diagnostics []DiagnosticResult) error {
	output := map[string]interface{}{
		"schema_version":   1,
		"touchlog_version": version.GetVersion(),
		"diagnostics":      diagnostics,
		"count":            len(diagnostics),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

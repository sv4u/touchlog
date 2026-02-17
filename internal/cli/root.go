package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/sv4u/touchlog/v2/internal/version"
	cli3 "github.com/urfave/cli/v3"
)

// BuildRootCommand builds the root CLI command
func BuildRootCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "touchlog",
		Usage: "A knowledge graph note-taking system",
		Flags: []cli3.Flag{
			&cli3.StringFlag{
				Name:  "vault",
				Usage: "Path to vault root (default: auto-detect)",
			},
		},
		Commands: []*cli3.Command{
			buildVersionCommand(),
			buildCompletionCommand(),
			BuildInitCommand(),
			BuildNewCommand(),
			BuildEditCommand(),
			BuildIndexCommand(),
			BuildQueryCommand(),
			BuildViewCommand(),
			BuildGraphCommand(),
			BuildDiagnosticsCommand(),
			BuildDaemonCommand(),
		},
		EnableShellCompletion: true,
		Suggest:               true, // Enable command suggestions
	}
}

// buildVersionCommand builds the version command
func buildVersionCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(ctx context.Context, cmd *cli3.Command) error {
			fmt.Printf("touchlog version %s\n", version.GetVersion())
			return nil
		},
	}
}

// buildCompletionCommand builds the shell completion command
func buildCompletionCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "completion",
		Usage: "Generate shell completion scripts",
		Commands: []*cli3.Command{
			{
				Name:  "bash",
				Usage: "Generate bash completion script",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					root := cmd.Root()
					var names []string
					for _, sub := range root.Commands {
						names = append(names, sub.Name)
					}
					fmt.Println("# touchlog bash completion")
					fmt.Println("_touchlog_completion() {")
					fmt.Printf("  COMPREPLY=($(compgen -W \"%s\" -- \"${COMP_WORDS[COMP_CWORD]}\"))\n", strings.Join(names, " "))
					fmt.Println("}")
					fmt.Println("complete -F _touchlog_completion touchlog")
					return nil
				},
			},
			{
				Name:  "zsh",
				Usage: "Generate zsh completion script",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					root := cmd.Root()
					fmt.Println("# touchlog zsh completion")
					fmt.Println("_touchlog() {")
					fmt.Println("  local -a commands")
					fmt.Println("  commands=(")
					for _, sub := range root.Commands {
						fmt.Printf("    '%s:%s'\n", sub.Name, sub.Usage)
					}
					fmt.Println("  )")
					fmt.Println("  _describe 'commands' commands")
					fmt.Println("}")
					fmt.Println("compdef _touchlog touchlog")
					return nil
				},
			},
			{
				Name:  "fish",
				Usage: "Generate fish completion script",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					root := cmd.Root()
					fmt.Println("# touchlog fish completion")
					for _, sub := range root.Commands {
						fmt.Printf("complete -c touchlog -f -a %s -d '%s'\n", sub.Name, sub.Usage)
					}
					return nil
				},
			},
		},
	}
}

// GetVaultFromContext extracts the vault path from the CLI context
// The vault flag is on the root command, so we need to check the root
func GetVaultFromContext(ctx context.Context, cmd *cli3.Command) (string, error) {
	// Get the root command to access global flags
	root := cmd.Root()
	vaultFlag := root.String("vault")
	return ResolveVault(vaultFlag)
}

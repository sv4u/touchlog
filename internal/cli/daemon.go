package cli

import (
	"context"
	"fmt"

	"github.com/sv4u/touchlog/v2/internal/daemon"
	cli3 "github.com/urfave/cli/v3"
)

// BuildDaemonCommand builds the daemon command
func BuildDaemonCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "daemon",
		Usage: "Manage the touchlog daemon",
		Commands: []*cli3.Command{
			{
				Name:  "start",
				Usage: "Start the daemon",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					d := daemon.NewDaemon(vaultRoot)
					if err := d.Start(); err != nil {
						return fmt.Errorf("starting daemon: %w", err)
					}

					fmt.Println("Daemon started successfully")
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the daemon",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					d := daemon.NewDaemon(vaultRoot)
					if err := d.Stop(); err != nil {
						return fmt.Errorf("stopping daemon: %w", err)
					}

					fmt.Println("Daemon stopped successfully")
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check daemon status",
				Action: func(ctx context.Context, cmd *cli3.Command) error {
					vaultRoot, err := GetVaultFromContext(ctx, cmd)
					if err != nil {
						return fmt.Errorf("resolving vault: %w", err)
					}

					d := daemon.NewDaemon(vaultRoot)
					running, pid, err := d.Status()
					if err != nil {
						return fmt.Errorf("checking daemon status: %w", err)
					}

					if running {
						fmt.Printf("Daemon is running (PID: %d)\n", pid)
					} else {
						fmt.Println("Daemon is not running")
					}

					return nil
				},
			},
		},
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sv4u/touchlog/internal/cli"
)

func main() {
	cmd := cli.BuildRootCommand()

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

package cli

import (
	"testing"

	cli3 "github.com/urfave/cli/v3"
)

// TestBuildDaemonCommand_Structure tests the structure of daemon command
func TestBuildDaemonCommand_Structure(t *testing.T) {
	daemonCmd := BuildDaemonCommand()

	if daemonCmd.Name != "daemon" {
		t.Errorf("expected command name 'daemon', got %q", daemonCmd.Name)
	}

	if len(daemonCmd.Commands) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(daemonCmd.Commands))
	}

	// Verify subcommands exist
	subcommandNames := []string{"start", "stop", "status"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range daemonCmd.Commands {
			if cmd.Name == name {
				found = true
				if cmd.Action == nil {
					t.Errorf("subcommand %q should have an action", name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

// TestBuildDiagnosticsCommand_Structure tests the structure of diagnostics command
func TestBuildDiagnosticsCommand_Structure(t *testing.T) {
	diagnosticsCmd := BuildDiagnosticsCommand()

	if diagnosticsCmd.Name != "diagnostics" {
		t.Errorf("expected command name 'diagnostics', got %q", diagnosticsCmd.Name)
	}

	if len(diagnosticsCmd.Commands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(diagnosticsCmd.Commands))
	}

	listCmd := diagnosticsCmd.Commands[0]
	if listCmd.Name != "list" {
		t.Errorf("expected subcommand name 'list', got %q", listCmd.Name)
	}

	if listCmd.Action == nil {
		t.Fatal("list subcommand should have an action")
	}

	// Verify flags exist
	foundLevelFlag := false
	foundNodeFlag := false
	foundCodeFlag := false
	foundFormatFlag := false

	for _, flag := range listCmd.Flags {
		if f, ok := flag.(*cli3.StringFlag); ok {
			switch f.Name {
			case "level":
				foundLevelFlag = true
			case "node":
				foundNodeFlag = true
			case "code":
				foundCodeFlag = true
			case "format":
				foundFormatFlag = true
			}
		}
	}

	if !foundLevelFlag {
		t.Error("expected --level flag to exist")
	}
	if !foundNodeFlag {
		t.Error("expected --node flag to exist")
	}
	if !foundCodeFlag {
		t.Error("expected --code flag to exist")
	}
	if !foundFormatFlag {
		t.Error("expected --format flag to exist")
	}
}

// TestBuildGraphCommand_Structure tests the structure of graph command
func TestBuildGraphCommand_Structure(t *testing.T) {
	graphCmd := BuildGraphCommand()

	if graphCmd.Name != "graph" {
		t.Errorf("expected command name 'graph', got %q", graphCmd.Name)
	}

	if len(graphCmd.Commands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(graphCmd.Commands))
	}

	exportCmd := graphCmd.Commands[0]
	if exportCmd.Name != "export" {
		t.Errorf("expected subcommand name 'export', got %q", exportCmd.Name)
	}

	if len(exportCmd.Commands) != 1 {
		t.Errorf("expected 1 export subcommand, got %d", len(exportCmd.Commands))
	}

	dotCmd := exportCmd.Commands[0]
	if dotCmd.Name != "dot" {
		t.Errorf("expected export subcommand name 'dot', got %q", dotCmd.Name)
	}

	if dotCmd.Action == nil {
		t.Fatal("dot export subcommand should have an action")
	}
}

// TestBuildIndexCommand_Structure tests the structure of index command
func TestBuildIndexCommand_Structure(t *testing.T) {
	indexCmd := BuildIndexCommand()

	if indexCmd.Name != "index" {
		t.Errorf("expected command name 'index', got %q", indexCmd.Name)
	}

	if len(indexCmd.Commands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(indexCmd.Commands))
	}

	// Verify subcommands exist
	subcommandNames := []string{"rebuild", "export"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range indexCmd.Commands {
			if cmd.Name == name {
				found = true
				if cmd.Action == nil {
					t.Errorf("subcommand %q should have an action", name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

// TestBuildQueryCommand_Structure tests the structure of query command
func TestBuildQueryCommand_Structure(t *testing.T) {
	queryCmd := BuildQueryCommand()

	if queryCmd.Name != "query" {
		t.Errorf("expected command name 'query', got %q", queryCmd.Name)
	}

	if len(queryCmd.Commands) < 2 {
		t.Errorf("expected at least 2 subcommands, got %d", len(queryCmd.Commands))
	}

	// Verify subcommands exist
	subcommandNames := []string{"backlinks", "search"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range queryCmd.Commands {
			if cmd.Name == name {
				found = true
				if cmd.Action == nil {
					t.Errorf("subcommand %q should have an action", name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

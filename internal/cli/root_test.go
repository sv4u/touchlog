package cli

import (
	"context"
	"testing"

	cli3 "github.com/urfave/cli/v3"
)

// TestBuildRootCommand tests root command structure
func TestBuildRootCommand(t *testing.T) {
	rootCmd := BuildRootCommand()

	if rootCmd.Name != "touchlog" {
		t.Errorf("expected command name 'touchlog', got %q", rootCmd.Name)
	}

	if len(rootCmd.Commands) == 0 {
		t.Error("expected root command to have subcommands")
	}

	// Check for expected subcommands
	expectedCommands := []string{"version", "completion", "init", "new", "index", "query", "graph", "diagnostics", "daemon"}
	commandMap := make(map[string]bool)
	for _, cmd := range rootCmd.Commands {
		commandMap[cmd.Name] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

// TestBuildVersionCommand tests version command behavior
func TestBuildVersionCommand(t *testing.T) {
	versionCmd := buildVersionCommand()

	if versionCmd.Name != "version" {
		t.Errorf("expected command name 'version', got %q", versionCmd.Name)
	}

	if versionCmd.Action == nil {
		t.Error("expected version command to have action")
	}
}

// TestBuildCompletionCommand tests completion command structure
func TestBuildCompletionCommand(t *testing.T) {
	completionCmd := buildCompletionCommand()

	if completionCmd.Name != "completion" {
		t.Errorf("expected command name 'completion', got %q", completionCmd.Name)
	}

	if len(completionCmd.Commands) == 0 {
		t.Error("expected completion command to have subcommands")
	}

	// Check for expected shell completions
	expectedShells := []string{"bash", "zsh", "fish"}
	shellMap := make(map[string]bool)
	for _, cmd := range completionCmd.Commands {
		shellMap[cmd.Name] = true
	}

	for _, expected := range expectedShells {
		if !shellMap[expected] {
			t.Errorf("expected completion subcommand %q not found", expected)
		}
	}
}

// TestBuildDaemonCommand tests daemon command structure
func TestBuildDaemonCommand(t *testing.T) {
	daemonCmd := BuildDaemonCommand()

	if daemonCmd.Name != "daemon" {
		t.Errorf("expected command name 'daemon', got %q", daemonCmd.Name)
	}

	if len(daemonCmd.Commands) == 0 {
		t.Error("expected daemon command to have subcommands")
	}

	// Check for expected subcommands
	expectedCommands := []string{"start", "stop", "status"}
	commandMap := make(map[string]bool)
	for _, cmd := range daemonCmd.Commands {
		commandMap[cmd.Name] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("expected daemon subcommand %q not found", expected)
		}
	}
}

// TestBuildGraphCommand tests graph command structure
func TestBuildGraphCommand(t *testing.T) {
	graphCmd := BuildGraphCommand()

	if graphCmd.Name != "graph" {
		t.Errorf("expected command name 'graph', got %q", graphCmd.Name)
	}

	if len(graphCmd.Commands) == 0 {
		t.Error("expected graph command to have subcommands")
	}
}

// TestBuildIndexCommand tests index command structure
func TestBuildIndexCommand(t *testing.T) {
	indexCmd := BuildIndexCommand()

	if indexCmd.Name != "index" {
		t.Errorf("expected command name 'index', got %q", indexCmd.Name)
	}

	if len(indexCmd.Commands) == 0 {
		t.Error("expected index command to have subcommands")
	}
}

// TestBuildDiagnosticsCommand tests diagnostics command structure
func TestBuildDiagnosticsCommand(t *testing.T) {
	diagCmd := BuildDiagnosticsCommand()

	if diagCmd.Name != "diagnostics" {
		t.Errorf("expected command name 'diagnostics', got %q", diagCmd.Name)
	}

	if len(diagCmd.Commands) == 0 {
		t.Error("expected diagnostics command to have subcommands")
	}
}

// TestVersionCommand_Action tests the version command action behavior
func TestVersionCommand_Action(t *testing.T) {
	versionCmd := buildVersionCommand()
	if versionCmd.Action == nil {
		t.Fatal("version command should have an action")
	}

	// Test that action executes without error
	ctx := context.Background()
	cmd := &cli3.Command{}
	err := versionCmd.Action(ctx, cmd)
	if err != nil {
		t.Fatalf("version command action failed: %v", err)
	}
}

// TestCompletionCommand_Actions tests that completion command actions execute without error
func TestCompletionCommand_Actions(t *testing.T) {
	completionCmd := buildCompletionCommand()

	ctx := context.Background()
	testCmd := &cli3.Command{}

	// Test bash completion
	bashCmd := completionCmd.Commands[0]
	if bashCmd.Action == nil {
		t.Fatal("bash completion command should have an action")
	}
	if err := bashCmd.Action(ctx, testCmd); err != nil {
		t.Errorf("bash completion action failed: %v", err)
	}

	// Test zsh completion
	zshCmd := completionCmd.Commands[1]
	if zshCmd.Action == nil {
		t.Fatal("zsh completion command should have an action")
	}
	if err := zshCmd.Action(ctx, testCmd); err != nil {
		t.Errorf("zsh completion action failed: %v", err)
	}

	// Test fish completion
	fishCmd := completionCmd.Commands[2]
	if fishCmd.Action == nil {
		t.Fatal("fish completion command should have an action")
	}
	if err := fishCmd.Action(ctx, testCmd); err != nil {
		t.Errorf("fish completion action failed: %v", err)
	}
}

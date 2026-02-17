package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
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

// runCompletionCommand runs a completion subcommand through the full root command
// tree so that cmd.Root() returns the real root with all subcommands populated.
func runCompletionCommand(t *testing.T, shell string) string {
	t.Helper()

	rootCmd := BuildRootCommand()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	runErr := rootCmd.Run(context.Background(), []string{"touchlog", "completion", shell})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if runErr != nil {
		t.Fatalf("%s completion failed: %v", shell, runErr)
	}

	return buf.String()
}

// TestCompletionCommand_Actions tests that completion actions produce output
// containing all registered root commands. The test runs through the full
// command tree so that cmd.Root() returns the real root.
func TestCompletionCommand_Actions(t *testing.T) {
	// Expected commands that must appear in completion output
	expectedCommands := []string{"version", "init", "new", "index", "query", "graph", "diagnostics", "daemon"}

	t.Run("bash", func(t *testing.T) {
		output := runCompletionCommand(t, "bash")
		if !strings.Contains(output, "_touchlog_completion") {
			t.Error("bash completion should contain function definition")
		}
		for _, cmd := range expectedCommands {
			if !strings.Contains(output, cmd) {
				t.Errorf("bash completion missing command %q", cmd)
			}
		}
	})

	t.Run("zsh", func(t *testing.T) {
		output := runCompletionCommand(t, "zsh")
		if !strings.Contains(output, "compdef _touchlog") {
			t.Error("zsh completion should contain compdef directive")
		}
		for _, cmd := range expectedCommands {
			if !strings.Contains(output, cmd) {
				t.Errorf("zsh completion missing command %q", cmd)
			}
		}
	})

	t.Run("fish", func(t *testing.T) {
		output := runCompletionCommand(t, "fish")
		if !strings.Contains(output, "complete -c touchlog") {
			t.Error("fish completion should contain complete directives")
		}
		for _, cmd := range expectedCommands {
			if !strings.Contains(output, cmd) {
				t.Errorf("fish completion missing command %q", cmd)
			}
		}
	})
}

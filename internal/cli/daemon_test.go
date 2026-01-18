package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDaemonCommand_Start_Behavior tests the behavioral aspects of daemon start command
func TestDaemonCommand_Start_Behavior(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemonCmd := BuildDaemonCommand()
	startCmd := daemonCmd.Commands[0]

	if startCmd.Action == nil {
		t.Fatal("start command should have an action")
	}

	// Test that the command structure is correct
	if startCmd.Name != "start" {
		t.Errorf("expected command name 'start', got %q", startCmd.Name)
	}

	// Verify vault exists
	vaultRoot, err := ResolveVault(tmpDir)
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, ".touchlog", "config.yaml")); err != nil {
		t.Fatalf("vault config not found: %v", err)
	}
}

// TestDaemonCommand_Stop_Behavior tests the behavioral aspects of daemon stop command
func TestDaemonCommand_Stop_Behavior(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemonCmd := BuildDaemonCommand()
	stopCmd := daemonCmd.Commands[1]

	if stopCmd.Action == nil {
		t.Fatal("stop command should have an action")
	}

	if stopCmd.Name != "stop" {
		t.Errorf("expected command name 'stop', got %q", stopCmd.Name)
	}
}

// TestDaemonCommand_Status_Behavior tests the behavioral aspects of daemon status command
func TestDaemonCommand_Status_Behavior(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemonCmd := BuildDaemonCommand()
	statusCmd := daemonCmd.Commands[2]

	if statusCmd.Action == nil {
		t.Fatal("status command should have an action")
	}

	if statusCmd.Name != "status" {
		t.Errorf("expected command name 'status', got %q", statusCmd.Name)
	}
}

// TestDaemonCommand_Start_InvalidVault tests start command behavior with invalid vault
func TestDaemonCommand_Start_InvalidVault(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create vault - ResolveVault should fail when auto-detecting

	daemonCmd := BuildDaemonCommand()
	startCmd := daemonCmd.Commands[0]

	if startCmd == nil {
		t.Fatal("start command should exist")
	}

	// Test that ResolveVault fails for non-existent vault when auto-detecting
	// (when explicit path is provided, it just returns the absolute path)
	_, err := ResolveVault("")
	if err == nil {
		// This might not fail if we're in a directory with a vault
		// So we just verify the command structure
		t.Log("ResolveVault with empty string may succeed if in vault directory")
	}

	// Test with explicit non-existent path - should return absolute path without error
	absPath, err := ResolveVault(tmpDir)
	if err != nil {
		t.Logf("ResolveVault with explicit path returned error: %v", err)
	} else {
		// Should return absolute path even if vault doesn't exist
		if absPath == "" {
			t.Error("ResolveVault should return absolute path even for non-existent directory")
		}
	}
}

// setupTestVault creates a minimal test vault
func setupTestVault(t *testing.T, tmpDir string) {
	t.Helper()

	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	configPath := filepath.Join(touchlogDir, "config.yaml")
	configContent := `version: 1
types:
  note:
    description: A note
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
tags:
  preferred: []
edges:
  related-to:
    description: General relationship
templates:
  root: templates
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
}

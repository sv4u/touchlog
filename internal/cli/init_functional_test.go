package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildInitCommand_Action tests the init command action behavior
func TestBuildInitCommand_Action(t *testing.T) {
	tmpDir := t.TempDir()

	initCmd := BuildInitCommand()
	if initCmd.Action == nil {
		t.Fatal("init command should have an action")
	}

	// Test that runInitWizard works (which is what the action calls)
	err := runInitWizard(tmpDir)
	if err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Verify config was created
	configPath := filepath.Join(tmpDir, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
}

// TestBuildInitCommand_Action_AlreadyInitialized tests init command when vault already exists
func TestBuildInitCommand_Action_AlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	// Try to initialize again - should fail
	err := runInitWizard(tmpDir)
	if err == nil {
		t.Error("expected error when vault is already initialized")
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Errorf("expected error about vault already initialized, got: %v", err)
	}
}

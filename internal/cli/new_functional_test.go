package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestBuildNewCommand_Action tests the new command action behavior
func TestBuildNewCommand_Action(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	newCmd := BuildNewCommand()
	if newCmd.Action == nil {
		t.Fatal("new command should have an action")
	}

	// Test that runNewWizard works (which is what the action calls)
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	err = runNewWizard(tmpDir, cfg)
	if err != nil {
		t.Fatalf("runNewWizard failed: %v", err)
	}

	// Verify note was created
	var notePath string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".Rmd") {
			notePath = path
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking directory: %v", err)
	}
	if notePath == "" {
		t.Fatal("note file was not created")
	}
}

// TestBuildNewCommand_Action_InvalidVault tests new command with invalid vault
func TestBuildNewCommand_Action_InvalidVault(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create vault

	cfg := &config.Config{
		Types: map[model.TypeName]config.TypeDef{},
	}

	// Test that runNewWizard fails when no types are configured
	err := runNewWizard(tmpDir, cfg)
	if err == nil {
		t.Error("expected error when no types are configured")
	}
	if !strings.Contains(err.Error(), "no types") {
		t.Errorf("expected error about no types, got: %v", err)
	}
}

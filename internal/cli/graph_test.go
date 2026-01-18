package cli

import (
	"os"
	"path/filepath"
	"testing"

	cli3 "github.com/urfave/cli/v3"
)

// TestGraphCommand_Export_Structure tests the structure of graph export command
func TestGraphCommand_Export_Structure(t *testing.T) {
	graphCmd := BuildGraphCommand()
	exportCmd := graphCmd.Commands[0]
	dotCmd := exportCmd.Commands[0]

	if dotCmd.Name != "dot" {
		t.Errorf("expected command name 'dot', got %q", dotCmd.Name)
	}

	if dotCmd.Action == nil {
		t.Fatal("dot export command should have an action")
	}

	// Verify required flags exist
	foundOutFlag := false
	for _, flag := range dotCmd.Flags {
		if f, ok := flag.(*cli3.StringFlag); ok && f.Name == "out" {
			foundOutFlag = true
			if !f.Required {
				t.Error("--out flag should be required")
			}
			break
		}
	}
	if !foundOutFlag {
		t.Error("expected --out flag to exist")
	}
}

// TestGraphCommand_Export_InvalidVault tests export command with invalid vault path
func TestGraphCommand_Export_InvalidVault(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create vault

	// ResolveVault with explicit path returns absolute path without checking if vault exists
	vaultRoot, err := ResolveVault(tmpDir)
	if err != nil {
		t.Fatalf("ResolveVault should return absolute path even for non-existent directory: %v", err)
	}
	if vaultRoot == "" {
		t.Error("ResolveVault should return absolute path")
	}
	
	// ValidateVault should fail for non-existent vault
	err = ValidateVault(vaultRoot)
	if err == nil {
		t.Error("expected ValidateVault to fail when vault doesn't exist")
	}
}

// TestGraphCommand_Export_ValidVault tests export command structure with valid vault
func TestGraphCommand_Export_ValidVault(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	graphCmd := BuildGraphCommand()
	exportCmd := graphCmd.Commands[0]
	dotCmd := exportCmd.Commands[0]

	// Verify command structure
	if exportCmd.Name != "export" {
		t.Errorf("expected command name 'export', got %q", exportCmd.Name)
	}

	if len(dotCmd.Flags) == 0 {
		t.Error("expected dot command to have flags")
	}

	// Verify vault exists
	vaultRoot, err := ResolveVault(tmpDir)
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, ".touchlog", "config.yaml")); err != nil {
		t.Fatalf("vault config not found: %v", err)
	}
	
	_ = dotCmd
}

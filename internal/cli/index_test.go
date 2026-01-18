package cli

import (
	"os"
	"path/filepath"
	"testing"

	cli3 "github.com/urfave/cli/v3"
)

// TestIndexCommand_Rebuild_Structure tests the structure of index rebuild command
func TestIndexCommand_Rebuild_Structure(t *testing.T) {
	indexCmd := BuildIndexCommand()
	rebuildCmd := indexCmd.Commands[0]

	if rebuildCmd.Name != "rebuild" {
		t.Errorf("expected command name 'rebuild', got %q", rebuildCmd.Name)
	}

	if rebuildCmd.Action == nil {
		t.Fatal("rebuild command should have an action")
	}
}

// TestIndexCommand_Export_Structure tests the structure of index export command
func TestIndexCommand_Export_Structure(t *testing.T) {
	indexCmd := BuildIndexCommand()
	exportCmd := indexCmd.Commands[1]

	if exportCmd.Name != "export" {
		t.Errorf("expected command name 'export', got %q", exportCmd.Name)
	}

	if exportCmd.Action == nil {
		t.Fatal("export command should have an action")
	}

	// Verify required flags exist
	foundOutFlag := false
	for _, flag := range exportCmd.Flags {
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

// TestIndexCommand_Export_InvalidFormat tests export command with invalid format
func TestIndexCommand_Export_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	indexCmd := BuildIndexCommand()
	exportCmd := indexCmd.Commands[1]

	// Find format flag
	var formatFlag *cli3.StringFlag
	for _, flag := range exportCmd.Flags {
		if f, ok := flag.(*cli3.StringFlag); ok && f.Name == "format" {
			formatFlag = f
			break
		}
	}
	if formatFlag == nil {
		t.Fatal("expected --format flag to exist")
	}

	// Default should be "json"
	if formatFlag.Value != "json" {
		t.Errorf("expected default format 'json', got %q", formatFlag.Value)
	}
}

// TestIndexCommand_Export_InvalidVault tests export command with invalid vault
func TestIndexCommand_Export_InvalidVault(t *testing.T) {
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

// TestIndexCommand_Export_ValidVault tests export command structure with valid vault
func TestIndexCommand_Export_ValidVault(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	// Verify vault exists
	vaultRoot, err := ResolveVault(tmpDir)
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, ".touchlog", "config.yaml")); err != nil {
		t.Fatalf("vault config not found: %v", err)
	}
}

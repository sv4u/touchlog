package cli

import (
	"os"
	"path/filepath"
	"testing"

	cli3 "github.com/urfave/cli/v3"
)

// TestParseCSV tests the parseCSV function behavior
func TestParseCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single value",
			input:    "value1",
			expected: []string{"value1"},
		},
		{
			name:     "multiple values",
			input:    "value1,value2,value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "values with spaces",
			input:    "value1, value2 , value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "empty values filtered",
			input:    "value1,,value2, ,value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "only spaces",
			input:    " , , ",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "value1,value2,",
			expected: []string{"value1", "value2"},
		},
		{
			name:     "leading comma",
			input:    ",value1,value2",
			expected: []string{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCSV(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d values, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected value[%d] = %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestQueryCommand_Backlinks_Structure tests the structure of backlinks command
func TestQueryCommand_Backlinks_Structure(t *testing.T) {
	queryCmd := BuildQueryCommand()
	backlinksCmd := queryCmd.Commands[0]

	if backlinksCmd.Name != "backlinks" {
		t.Errorf("expected command name 'backlinks', got %q", backlinksCmd.Name)
	}

	if backlinksCmd.Action == nil {
		t.Fatal("backlinks command should have an action")
	}

	// Verify required flags exist
	foundTargetFlag := false
	for _, flag := range backlinksCmd.Flags {
		if f, ok := flag.(*cli3.StringFlag); ok && f.Name == "target" {
			foundTargetFlag = true
			if !f.Required {
				t.Error("--target flag should be required")
			}
			break
		}
	}
	if !foundTargetFlag {
		t.Error("expected --target flag to exist")
	}
}

// TestQueryCommand_Search_Structure tests the structure of search command
func TestQueryCommand_Search_Structure(t *testing.T) {
	queryCmd := BuildQueryCommand()
	searchCmd := queryCmd.Commands[1]

	if searchCmd.Name != "search" {
		t.Errorf("expected command name 'search', got %q", searchCmd.Name)
	}

	if searchCmd.Action == nil {
		t.Fatal("search command should have an action")
	}

	// Verify flags exist
	if len(searchCmd.Flags) == 0 {
		t.Error("expected search command to have flags")
	}
}

// TestQueryCommand_InvalidVault tests query commands with invalid vault
func TestQueryCommand_InvalidVault(t *testing.T) {
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

// TestQueryCommand_ValidVault tests query commands structure with valid vault
func TestQueryCommand_ValidVault(t *testing.T) {
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

// TestQueryCommand_Neighbors_Structure tests the structure of neighbors command
func TestQueryCommand_Neighbors_Structure(t *testing.T) {
	queryCmd := BuildQueryCommand()
	neighborsCmd := queryCmd.Commands[2]

	if neighborsCmd.Name != "neighbors" {
		t.Errorf("expected command name 'neighbors', got %q", neighborsCmd.Name)
	}

	if neighborsCmd.Action == nil {
		t.Fatal("neighbors command should have an action")
	}

	// Verify required flags exist
	foundRootFlag := false
	for _, flag := range neighborsCmd.Flags {
		if f, ok := flag.(*cli3.StringFlag); ok && f.Name == "root" {
			foundRootFlag = true
			if !f.Required {
				t.Error("--root flag should be required")
			}
			break
		}
	}
	if !foundRootFlag {
		t.Error("expected --root flag to exist")
	}
}

// TestQueryCommand_Paths_Structure tests the structure of paths command
func TestQueryCommand_Paths_Structure(t *testing.T) {
	queryCmd := BuildQueryCommand()
	pathsCmd := queryCmd.Commands[3]

	if pathsCmd.Name != "paths" {
		t.Errorf("expected command name 'paths', got %q", pathsCmd.Name)
	}

	if pathsCmd.Action == nil {
		t.Fatal("paths command should have an action")
	}

	// Verify required flags exist
	foundSourceFlag := false
	foundDestFlag := false
	for _, flag := range pathsCmd.Flags {
		if f, ok := flag.(*cli3.StringFlag); ok && f.Name == "source" {
			foundSourceFlag = true
			if !f.Required {
				t.Error("--source flag should be required")
			}
		}
		if f, ok := flag.(*cli3.StringSliceFlag); ok && f.Name == "destination" {
			foundDestFlag = true
			if !f.Required {
				t.Error("--destination flag should be required")
			}
		}
	}
	if !foundSourceFlag {
		t.Error("expected --source flag to exist")
	}
	if !foundDestFlag {
		t.Error("expected --destination flag to exist")
	}
}

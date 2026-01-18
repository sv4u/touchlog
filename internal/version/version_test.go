package version

import (
	"testing"
)

// TestGetVersion_WithCommit tests GetVersion when Commit is set
func TestGetVersion_WithCommit(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit

	// Restore after test
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	// Test with both version and commit
	Version = "1.2.3"
	Commit = "abc123"
	result := GetVersion()
	expected := "1.2.3-abc123"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}
}

// TestGetVersion_WithoutCommit tests GetVersion when Commit is empty
func TestGetVersion_WithoutCommit(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit

	// Restore after test
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	// Test with only version (no commit)
	Version = "1.2.3"
	Commit = ""
	result := GetVersion()
	expected := "1.2.3"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}
}

// TestGetVersion_DefaultValues tests GetVersion with default values
func TestGetVersion_DefaultValues(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit

	// Restore after test
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	// Test with default values
	Version = "dev"
	Commit = ""
	result := GetVersion()
	expected := "dev"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}
}

// TestGetVersion_EmptyVersion tests GetVersion with empty version
func TestGetVersion_EmptyVersion(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit

	// Restore after test
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	// Test with empty version
	Version = ""
	Commit = "abc123"
	result := GetVersion()
	expected := "-abc123"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}
}

package version

import (
	"bytes"
	"io"
	"os"
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
	originalStderr := os.Stderr

	// Restore after test
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		os.Stderr = originalStderr
	}()

	// Capture stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Test with default values
	Version = "dev"
	Commit = ""

	// Call GetVersion in goroutine to avoid blocking
	done := make(chan string)
	go func() {
		result := GetVersion()
		w.Close()
		done <- result
	}()

	// Read captured stderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	result := <-done

	// Verify version output
	expected := "dev"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}

	// Verify warning message
	expectedWarning := "warning: version information not injected at build time (built without ldflags)\n"
	if output != expectedWarning {
		t.Errorf("expected stderr to be %q, got %q", expectedWarning, output)
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

// TestGetVersion_WarningWhenDevAndNoCommit tests that a warning is logged to stderr
// when Version is "dev" and Commit is empty
func TestGetVersion_WarningWhenDevAndNoCommit(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalStderr := os.Stderr
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		os.Stderr = originalStderr
	}()

	Version = "dev"
	Commit = ""

	// Capture stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Call GetVersion in goroutine to avoid blocking
	done := make(chan string)
	go func() {
		result := GetVersion()
		w.Close()
		done <- result
	}()

	// Read captured stderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	result := <-done

	// Verify version output
	if result != "dev" {
		t.Errorf("GetVersion() = %q, want %q", result, "dev")
	}

	// Verify warning message
	expectedWarning := "warning: version information not injected at build time (built without ldflags)\n"
	if output != expectedWarning {
		t.Errorf("expected stderr to be %q, got %q", expectedWarning, output)
	}
}

// TestGetVersion_NoWarningWhenCommitSet tests that no warning is logged when Commit is set
func TestGetVersion_NoWarningWhenCommitSet(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalStderr := os.Stderr
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		os.Stderr = originalStderr
	}()

	Version = "dev"
	Commit = "abc123"

	// Capture stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Call GetVersion in goroutine to avoid blocking
	done := make(chan string)
	go func() {
		result := GetVersion()
		w.Close()
		done <- result
	}()

	// Read captured stderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	result := <-done

	// Verify version output
	expected := "dev-abc123"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}

	// Verify no warning message
	if output != "" {
		t.Errorf("expected no stderr output, got %q", output)
	}
}

// TestGetVersion_NoWarningWhenVersionNotDev tests that no warning is logged when Version is not "dev"
func TestGetVersion_NoWarningWhenVersionNotDev(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalStderr := os.Stderr
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		os.Stderr = originalStderr
	}()

	Version = "1.2.3"
	Commit = ""

	// Capture stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Call GetVersion in goroutine to avoid blocking
	done := make(chan string)
	go func() {
		result := GetVersion()
		w.Close()
		done <- result
	}()

	// Read captured stderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	result := <-done

	// Verify version output
	expected := "1.2.3"
	if result != expected {
		t.Errorf("GetVersion() = %q, want %q", result, expected)
	}

	// Verify no warning message
	if output != "" {
		t.Errorf("expected no stderr output, got %q", output)
	}
}

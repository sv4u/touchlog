package metadata

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetGitContext_NotInRepo(t *testing.T) {
	// Create a temporary directory that's not a git repo
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	gitCtx, err := GetGitContext(tmpDir)
	if err == nil {
		t.Error("GetGitContext() in non-git directory error = nil, want error")
	}
	if gitCtx != nil {
		t.Error("GetGitContext() in non-git directory returned non-nil, want nil")
	}
}

func TestGetGitContext_InRepo(t *testing.T) {
	// Create a temporary directory and initialize it as a git repo
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git not available or failed to init: %v", err)
	}
	
	// Create a dummy file and commit to have a valid branch
	dummyFile := filepath.Join(tmpDir, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	
	cmd = exec.Command("git", "add", "dummy.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git add failed: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "test commit")
	cmd.Dir = tmpDir
	// Set git config to avoid prompts
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com")
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git commit failed: %v", err)
	}
	
	// Now test GetGitContext
	gitCtx, err := GetGitContext(tmpDir)
	if err != nil {
		t.Fatalf("GetGitContext() in git repo error = %v", err)
	}
	if gitCtx == nil {
		t.Fatal("GetGitContext() in git repo returned nil")
	}
	
	// Verify branch and commit are populated
	if gitCtx.Branch == "" {
		t.Error("GetGitContext() branch is empty")
	}
	if gitCtx.Commit == "" {
		t.Error("GetGitContext() commit is empty")
	}
}

func TestGetGitContext_EmptyDir(t *testing.T) {
	// Test with empty string (should use current directory)
	gitCtx, err := GetGitContext("")
	
	// This may or may not succeed depending on whether we're in a git repo
	// Just verify it doesn't panic
	if err == nil && gitCtx != nil {
		// If we got a result, verify it's valid
		if gitCtx.Branch == "" && gitCtx.Commit == "" {
			t.Error("GetGitContext() returned non-nil but with empty values")
		}
	}
}

func TestGetGitContext_WithTildePath(t *testing.T) {
	// Test with tilde path expansion
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git not available or failed to init: %v", err)
	}
	
	// Create a dummy file and commit
	dummyFile := filepath.Join(tmpDir, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	
	cmd = exec.Command("git", "add", "dummy.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git add failed: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "test commit")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com")
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git commit failed: %v", err)
	}
	
	// Test with tilde path (if we're in home directory)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Skipping test: cannot get home directory: %v", err)
	}
	
	// Only test if tmpDir is in home directory
	if strings.HasPrefix(tmpDir, homeDir) {
		relativePath, err := filepath.Rel(homeDir, tmpDir)
		if err != nil {
			t.Skipf("Skipping test: cannot get relative path: %v", err)
		}
		
		tildePath := "~/" + relativePath
		gitCtx, err := GetGitContext(tildePath)
		if err != nil {
			t.Fatalf("GetGitContext() with tilde path error = %v", err)
		}
		if gitCtx == nil {
			t.Fatal("GetGitContext() with tilde path returned nil")
		}
		if gitCtx.Branch == "" {
			t.Error("GetGitContext() branch is empty")
		}
	}
}

func TestGetGitContext_RelativePath(t *testing.T) {
	// Test with relative path
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git not available or failed to init: %v", err)
	}
	
	// Create a dummy file and commit
	dummyFile := filepath.Join(tmpDir, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	
	cmd = exec.Command("git", "add", "dummy.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git add failed: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "test commit")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com")
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test: git commit failed: %v", err)
	}
	
	// Change to a subdirectory and test with relative path
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Test with relative path ".." from subdir
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	gitCtx, err := GetGitContext("..")
	if err != nil {
		t.Fatalf("GetGitContext() with relative path error = %v", err)
	}
	if gitCtx == nil {
		t.Fatal("GetGitContext() with relative path returned nil")
	}
	if gitCtx.Branch == "" {
		t.Error("GetGitContext() branch is empty")
	}
}

func TestGetGitContext_InvalidPath(t *testing.T) {
	// Test with invalid path (should fail on expansion)
	_, err := GetGitContext("~invalid/path")
	if err == nil {
		t.Error("GetGitContext() with invalid path error = nil, want error")
	}
}


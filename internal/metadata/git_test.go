package metadata

import (
	"os"
	"os/exec"
	"path/filepath"
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


package metadata

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sv4u/touchlog/internal/entry"
)

// GetGitContext detects git context (branch and commit) from the specified directory
// Returns nil if not in a git repository or if git is not available
// The directory parameter should be the output directory where the file will be created
func GetGitContext(directory string) (*entry.GitContext, error) {
	// If directory is empty, use current working directory
	checkDir := directory
	if checkDir == "" {
		var err error
		checkDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Expand path (handle ~ and relative paths)
	expandedDir, err := expandPath(checkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand directory path: %w", err)
	}

	// Find git repository root by walking up the directory tree
	gitRoot, err := findGitRoot(expandedDir)
	if err != nil {
		return nil, err
	}

	// Get branch name
	branch, err := getGitBranch(gitRoot)
	if err != nil {
		// If we can't get branch, continue without it
		branch = ""
	}

	// Get commit hash
	commit, err := getGitCommit(gitRoot)
	if err != nil {
		// If we can't get commit, continue without it
		commit = ""
	}

	// If both branch and commit are empty, return nil (no git context)
	if branch == "" && commit == "" {
		return nil, nil
	}

	return &entry.GitContext{
		Branch: branch,
		Commit: commit,
	}, nil
}

// findGitRoot walks up the directory tree to find the .git directory
func findGitRoot(startDir string) (string, error) {
	dir := startDir
	for {
		gitDir := filepath.Join(dir, ".git")
		info, err := os.Stat(gitDir)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		// Check if .git is a file (for git worktrees)
		if err == nil && !info.IsDir() {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory, no git repo found
			return "", fmt.Errorf("not in a git repository")
		}
		dir = parent
	}
}

// getGitBranch gets the current branch name from the git repository
func getGitBranch(gitRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = gitRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git branch: %w", err)
	}
	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// getGitCommit gets the current commit hash (short format) from the git repository
func getGitCommit(gitRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = gitRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit: %w", err)
	}
	commit := strings.TrimSpace(string(output))
	return commit, nil
}

// expandPath expands a path, handling ~ and relative paths
func expandPath(path string) (string, error) {
	// Handle ~ expansion
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if path == "~" {
			return homeDir, nil
		}
		// Validate that paths starting with ~ must be ~/ (not ~something)
		if !strings.HasPrefix(path, "~/") {
			return "", fmt.Errorf("invalid path: paths starting with ~ must be followed by / (e.g., ~/path), got: %s", path)
		}
		// Skip the leading ~/
		remaining := path[2:]
		path = filepath.Join(homeDir, remaining)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}

	return absPath, nil
}

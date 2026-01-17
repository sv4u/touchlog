package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVault_ExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()

	vaultRoot, err := ResolveVault(tmpDir)
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}

	absPath, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Abs failed: %v", err)
	}

	if vaultRoot != absPath {
		t.Errorf("expected vault root %q, got %q", absPath, vaultRoot)
	}
}

func TestResolveVault_AutoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	touchlogDir := filepath.Join(tmpDir, ".touchlog")

	// Create .touchlog directory
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("failed to create .touchlog dir: %v", err)
	}

	// Change to a subdirectory
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	vaultRoot, err := ResolveVault("")
	if err != nil {
		t.Fatalf("ResolveVault failed: %v", err)
	}

	absPath, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Abs failed: %v", err)
	}

	// Normalize paths for comparison (handles /var vs /private/var on macOS)
	vaultRootClean := filepath.Clean(vaultRoot)
	absPathClean := filepath.Clean(absPath)

	// Use EvalSymlinks to handle symlink differences
	vaultRootEval, _ := filepath.EvalSymlinks(vaultRootClean)
	absPathEval, _ := filepath.EvalSymlinks(absPathClean)

	if vaultRootEval != absPathEval && vaultRootClean != absPathClean {
		t.Errorf("expected vault root %q (eval: %q), got %q (eval: %q)", absPathClean, absPathEval, vaultRootClean, vaultRootEval)
	}
}

func TestResolveVault_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	_, err = ResolveVault("")
	if err == nil {
		t.Fatal("expected ResolveVault to fail when no vault found")
	}
}

func TestValidateVault_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	configPath := filepath.Join(touchlogDir, "config.yaml")

	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("failed to create .touchlog dir: %v", err)
	}

	// Create a minimal config file
	configContent := `version: 1
types: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if err := ValidateVault(tmpDir); err != nil {
		t.Errorf("ValidateVault failed for valid vault: %v", err)
	}
}

func TestValidateVault_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	touchlogDir := filepath.Join(tmpDir, ".touchlog")

	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("failed to create .touchlog dir: %v", err)
	}

	err := ValidateVault(tmpDir)
	if err == nil {
		t.Fatal("expected ValidateVault to fail when config is missing")
	}

	if err.Error() == "" {
		t.Error("expected error message to suggest 'touchlog init'")
	}
}

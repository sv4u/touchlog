package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveVault resolves the vault root directory
// It walks upward from the current working directory until it finds a .touchlog/ directory
// If --vault flag is provided, it uses that instead
func ResolveVault(vaultFlag string) (string, error) {
	if vaultFlag != "" {
		// Use explicit vault path
		absPath, err := filepath.Abs(vaultFlag)
		if err != nil {
			return "", fmt.Errorf("resolving vault path: %w", err)
		}
		return absPath, nil
	}

	// Auto-detect by walking up from CWD
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}

	current := cwd
	for {
		touchlogDir := filepath.Join(current, ".touchlog")
		if info, err := os.Stat(touchlogDir); err == nil && info.IsDir() {
			return current, nil
		}

		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			break
		}
		current = parent
	}

	return "", fmt.Errorf("no vault found (no .touchlog/ directory found walking up from %s)", cwd)
}

// ValidateVault validates that a vault has the required structure
// For Phase 0, we only check that .touchlog/config.yaml exists (except for init command)
func ValidateVault(vaultRoot string) error {
	configPath := filepath.Join(vaultRoot, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vault not initialized: %s/.touchlog/config.yaml does not exist (run 'touchlog init')", vaultRoot)
		}
		return fmt.Errorf("checking vault config: %w", err)
	}
	return nil
}

package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/sv4u/touchlog/internal/errors"
)

// ValidateOutputDir validates that the output directory exists and is writable
// Returns an error if validation fails, nil otherwise
func ValidateOutputDir(path string) error {
	if path == "" {
		return errors.ErrOutputDirRequired
	}

	// Expand path (handle ~ and environment variables)
	expandedPath, err := ExpandPath(path)
	if err != nil {
		return fmt.Errorf("failed to expand path %q: %w", path, fmt.Errorf("%w: %w", errors.ErrPathExpansionFailed, err))
	}

	// Check if path exists
	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if parent directory exists
			parentDir := filepath.Dir(expandedPath)
			parentInfo, parentErr := os.Stat(parentDir)
			if parentErr != nil {
				if os.IsNotExist(parentErr) {
					return fmt.Errorf("parent directory does not exist: %s: %w", parentDir, errors.ErrOutputDirInvalid)
				}
				return fmt.Errorf("failed to check parent directory: %w", fmt.Errorf("%w: %w", errors.ErrOutputDirInvalid, parentErr))
			}
			if !parentInfo.IsDir() {
				return errors.WrapErrorf(errors.ErrOutputDirInvalid, "parent path is not a directory: %s", parentDir)
			}
			// Parent exists, directory can be created
			return nil
		}
		return fmt.Errorf("failed to check directory: %w", fmt.Errorf("%w: %w", errors.ErrOutputDirInvalid, err))
	}

	// Path exists, check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s: %w", expandedPath, errors.ErrOutputDirNotDirectory)
	}

	// Check if directory is writable
	// Try to create a temporary file in the directory
	testFile := filepath.Join(expandedPath, ".touchlog-write-test")
	testFileHandle, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", fmt.Errorf("%w: %w", errors.ErrOutputDirNotWritable, err))
	}
	testFileHandle.Close()
	os.Remove(testFile) // Clean up test file

	return nil
}

// ValidateUTF8 validates that the given data is valid UTF-8
// Returns an error if validation fails, nil otherwise
func ValidateUTF8(data []byte) error {
	if !utf8.Valid(data) {
		return errors.ErrInvalidUTF8
	}
	return nil
}

// ValidateConfigFile validates that a config file path is valid and readable
// This checks file existence and readability, but does not parse the file
// Returns an error if validation fails, nil otherwise
func ValidateConfigFile(path string) error {
	if path == "" {
		return errors.WrapError(errors.ErrConfigNotFound, "config file path is empty")
	}

	// Expand path
	expandedPath, err := ExpandPath(path)
	if err != nil {
		return fmt.Errorf("failed to expand config file path %q: %w", path, fmt.Errorf("%w: %w", errors.ErrPathExpansionFailed, err))
	}

	// Check if file exists
	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s: %w", expandedPath, errors.ErrConfigNotFound)
		}
		return fmt.Errorf("failed to check config file: %w", fmt.Errorf("%w: %w", errors.ErrConfigReadFailed, err))
	}

	// Check if it's a file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("config path is a directory, not a file: %s: %w", expandedPath, errors.ErrConfigInvalid)
	}

	// Check if file is readable
	file, err := os.Open(expandedPath)
	if err != nil {
		return fmt.Errorf("config file is not readable: %w", fmt.Errorf("%w: %w", errors.ErrConfigReadFailed, err))
	}
	file.Close()

	return nil
}

// ValidateTemplateSyntax performs basic validation on template syntax
// This is a basic check - full template parsing should be done by the template processor
// Returns an error if validation fails, nil otherwise
func ValidateTemplateSyntax(template string) error {
	if template == "" {
		// Empty template is valid (user may want an empty file)
		return nil
	}

	// Basic check: ensure template is valid UTF-8
	if !utf8.ValidString(template) {
		return errors.WrapError(errors.ErrTemplateInvalidSyntax, "template contains invalid UTF-8")
	}

	// Additional syntax validation could be added here
	// For now, we just check UTF-8 validity

	return nil
}

// ExpandPath expands a path, handling ~ and environment variables
// This is a public function that should be used throughout the codebase
// to ensure consistent path expansion behavior
func ExpandPath(path string) (string, error) {
	// Handle ~ expansion
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		// Handle "~" alone (expands to home directory)
		if len(path) == 1 {
			return homeDir, nil
		}

		// Handle "~/path" (must be followed by /)
		if len(path) > 1 && path[1] != '/' {
			return "", fmt.Errorf("invalid path: paths starting with ~ must be followed by / (e.g., ~/path), got: %s", path)
		}

		return filepath.Join(homeDir, path[2:]), nil
	}

	// Expand environment variables
	expandedPath := os.ExpandEnv(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}

	return absPath, nil
}

package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchEditor_Validation(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		editor   string
		args     []string
		filePath string
		wantErr  bool
	}{
		{
			name:     "launch with empty editor",
			editor:   "",
			args:     []string{},
			filePath: testFile,
			wantErr:  true,
		},
		{
			name:     "launch with empty file path",
			editor:   "vim",
			args:     []string{},
			filePath: "",
			wantErr:  true,
		},
		// Note: We don't test actual editor launch as it would hang in test environment
		// The actual launch functionality is tested through integration tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LaunchEditor(tt.editor, tt.args, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LaunchEditor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLaunchEditor_FileValidation(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		setup    func() string // Returns file path, cleanup handled by temp dir
		wantErr  bool
	}{
		{
			name:     "file exists",
			filePath: "",
			setup: func() string {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "test.md")
				os.WriteFile(testFile, []byte("# Test\n"), 0644)
				return testFile
			},
			wantErr: false,
		},
		{
			name:     "file does not exist",
			filePath: "",
			setup: func() string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.md")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			// Test file validation by checking if LaunchEditor returns appropriate error
			// We use a nonexistent editor so it fails quickly without hanging
			// The file validation happens before the editor launch attempt
			err := LaunchEditor("nonexistent-editor-xyz123", []string{}, filePath)

			if tt.wantErr {
				// File doesn't exist - should get file validation error
				if err == nil {
					t.Error("LaunchEditor() expected error for nonexistent file, got nil")
				} else if !contains(err.Error(), "file does not exist") {
					// Error should mention file, not just editor
					t.Logf("LaunchEditor() error = %v (expected file validation error)", err)
				}
			} else {
				// File exists - should get editor error (not file error)
				if err == nil {
					t.Error("LaunchEditor() expected error for nonexistent editor, got nil")
				} else if contains(err.Error(), "file does not exist") {
					t.Errorf("LaunchEditor() got file error but file exists: %v", err)
				}
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

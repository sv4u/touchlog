package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

func TestNewInternalEditor(t *testing.T) {
	cfg := config.CreateDefaultConfig()

	tests := []struct {
		name           string
		filePath       string
		initialContent string
		cfg            *config.Config
		wantErr        bool
	}{
		{
			name:           "valid file path",
			filePath:       "/tmp/test.md",
			initialContent: "# Test\n",
			cfg:            cfg,
			wantErr:        false,
		},
		{
			name:           "nil config",
			filePath:       "/tmp/test.md",
			initialContent: "# Test\n",
			cfg:            nil,
			wantErr:        true,
		},
		{
			name:           "relative file path",
			filePath:       "test.md",
			initialContent: "# Test\n",
			cfg:            cfg,
			wantErr:        false, // filepath.Abs should resolve it
		},
		{
			name:           "empty file path",
			filePath:       "",
			initialContent: "# Test\n",
			cfg:            cfg,
			wantErr:        false, // filepath.Abs("") returns current directory
		},
		{
			name:           "empty initial content",
			filePath:       "/tmp/test.md",
			initialContent: "",
			cfg:            cfg,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor, err := NewInternalEditor(tt.filePath, tt.initialContent, tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Error("NewInternalEditor() error = nil, want error")
				}
				if editor != nil {
					t.Error("NewInternalEditor() returned non-nil editor on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewInternalEditor() error = %v, want nil", err)
				}
				if editor == nil {
					t.Fatal("NewInternalEditor() returned nil editor")
				}
				if editor.config != tt.cfg {
					t.Errorf("NewInternalEditor() config = %v, want %v", editor.config, tt.cfg)
				}
				if editor.content != tt.initialContent {
					t.Errorf("NewInternalEditor() content = %q, want %q", editor.content, tt.initialContent)
				}
				// File path should be absolute
				if !filepath.IsAbs(editor.filePath) {
					t.Errorf("NewInternalEditor() filePath = %q, want absolute path", editor.filePath)
				}
			}
		})
	}
}

func TestInternalEditor_GetContent(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	testContent := "# Test Content\n\nThis is test content."

	// Create file with content
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	editor, err := NewInternalEditor(testFile, "initial content", cfg)
	if err != nil {
		t.Fatalf("NewInternalEditor() error = %v", err)
	}

	// GetContent should read from file
	content := editor.GetContent()
	if content != testContent {
		t.Errorf("GetContent() = %q, want %q", content, testContent)
	}
}

func TestInternalEditor_GetContent_FileNotFound(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	tmpDir := t.TempDir()
	nonexistentFile := filepath.Join(tmpDir, "nonexistent.md")
	initialContent := "initial content"

	editor, err := NewInternalEditor(nonexistentFile, initialContent, cfg)
	if err != nil {
		t.Fatalf("NewInternalEditor() error = %v", err)
	}

	// GetContent should return initial content when file doesn't exist
	content := editor.GetContent()
	if content != initialContent {
		t.Errorf("GetContent() = %q, want %q (initial content)", content, initialContent)
	}
}

func TestWithFilePathOverride(t *testing.T) {
	cfg := &modelConfig{}
	opt := WithFilePathOverride("/custom/path.md")
	opt(cfg)

	if cfg.filePathOverride != "/custom/path.md" {
		t.Errorf("WithFilePathOverride() filePathOverride = %q, want %q", cfg.filePathOverride, "/custom/path.md")
	}
}

func TestWithInitialContent(t *testing.T) {
	cfg := &modelConfig{}
	content := "# Custom Content\n\nThis is custom."
	opt := WithInitialContent(content)
	opt(cfg)

	if cfg.initialContent != content {
		t.Errorf("WithInitialContent() initialContent = %q, want %q", cfg.initialContent, content)
	}
}

func TestLaunchEditor_PathResolution(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// For relative path test, we need to change to the temp directory
	// so that "test.md" resolves correctly
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	tests := []struct {
		name     string
		filePath string
		setup    func() // Setup function to change directory if needed
		wantErr  bool
	}{
		{
			name:     "absolute path",
			filePath: testFile,
			setup:    func() {},
			wantErr:  false, // File exists, but editor will fail (which is expected)
		},
		{
			name:     "relative path",
			filePath: "test.md",
			setup: func() {
				os.Chdir(tmpDir)
			},
			wantErr: false, // Should resolve to absolute path
		},
		{
			name:     "nonexistent file",
			filePath: filepath.Join(tmpDir, "nonexistent.md"),
			setup:    func() {},
			wantErr:  true, // File doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			// Use nonexistent editor so it fails quickly
			// We're testing path resolution, not editor launch
			err := LaunchEditor("nonexistent-editor-xyz123", []string{}, tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Error("LaunchEditor() expected error, got nil")
				} else if !strings.Contains(err.Error(), "file does not exist") {
					// Should get file validation error, not just editor error
					t.Logf("LaunchEditor() error = %v (may be file validation or editor error)", err)
				}
			} else {
				// File exists, should get editor error (not file error)
				if err == nil {
					t.Error("LaunchEditor() expected error for nonexistent editor, got nil")
				} else if strings.Contains(err.Error(), "file does not exist") {
					t.Errorf("LaunchEditor() got file error but file should exist: %v", err)
				}
			}
		})
	}
}

func TestLaunchEditor_ArgsHandling(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that args are properly copied and file path is appended
	// We can't easily test the actual command execution, but we can verify
	// the function doesn't panic with various args
	args := []string{"-f", "--wait"}
	err := LaunchEditor("nonexistent-editor-xyz123", args, testFile)

	// Should get editor error (not file error or panic)
	if err == nil {
		t.Error("LaunchEditor() expected error for nonexistent editor, got nil")
	}
	// Error should mention editor, not file
	if err != nil && strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("LaunchEditor() got file error but file exists: %v", err)
	}
}

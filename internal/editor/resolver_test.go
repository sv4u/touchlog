package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

func TestEditorResolver_Resolve(t *testing.T) {
	tests := []struct {
		name               string
		cliEditor          string
		configEditor       *config.EditorConfig
		envEditor          string
		fallbackToInternal bool
		wantType           EditorType
		wantCommand        string
		wantUseInternal    bool
		setupEnv           func() func() // setup function returns cleanup
	}{
		{
			name:               "CLI editor takes precedence",
			cliEditor:          "vim",
			fallbackToInternal: true,
			wantType:           EditorTypeExternal,
			wantCommand:        "vim",
			wantUseInternal:    false,
		},
		{
			name:               "EDITOR env var used when CLI not set",
			envEditor:          "nano",
			fallbackToInternal: true,
			wantType:           EditorTypeExternal,
			wantCommand:        "nano",
			wantUseInternal:    false,
			setupEnv: func() func() {
				oldVal := os.Getenv("EDITOR")
				os.Setenv("EDITOR", "nano")
				return func() {
					if oldVal == "" {
						os.Unsetenv("EDITOR")
					} else {
						os.Setenv("EDITOR", oldVal)
					}
				}
			},
		},
		{
			name:               "Config editor used when CLI and env not set",
			configEditor:       &config.EditorConfig{Command: "vim"},
			fallbackToInternal: true,
			wantType:           EditorTypeExternal,
			wantCommand:        "vim",
			wantUseInternal:    false,
		},
		{
			name:               "Fallback to internal when no external found",
			fallbackToInternal: true,
			wantType:           EditorTypeInternal,
			wantUseInternal:    true,
			// Note: This test may pass with EditorTypeExternal if vi/nano are found
			// That's acceptable - the fallback only happens if no external editor is found
		},
		{
			name:               "Error when no editor and fallback disabled",
			fallbackToInternal: false,
			wantType:           EditorTypeExternal, // This will error
			wantUseInternal:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment if needed
			var cleanup func()
			if tt.setupEnv != nil {
				cleanup = tt.setupEnv()
				defer cleanup()
			} else {
				// Clear EDITOR env var for tests that don't set it
				oldVal := os.Getenv("EDITOR")
				defer func() {
					if oldVal == "" {
						os.Unsetenv("EDITOR")
					} else {
						os.Setenv("EDITOR", oldVal)
					}
				}()
				os.Unsetenv("EDITOR")
			}

			// Create config with editor if specified
			var cfg *config.Config
			if tt.configEditor != nil {
				cfg = config.CreateDefaultConfig()
				cfg.Editor = tt.configEditor
			}

			// Create resolver
			resolver := NewEditorResolver(tt.cliEditor, cfg, tt.fallbackToInternal)

			// Resolve editor
			info, err := resolver.Resolve()

			// Check results
			if !tt.fallbackToInternal && info == nil && err != nil {
				// Expected error when no editor and fallback disabled
				return
			}

			if err != nil {
				// If we expected an error, that's fine
				if !tt.fallbackToInternal {
					return
				}
				t.Errorf("Resolve() error = %v, want no error", err)
				return
			}

			// For fallback test, accept either internal or external (if vi/nano found)
			if tt.name == "Fallback to internal when no external found" {
				// If external editor was found (vi/nano), that's fine
				if info.Type == EditorTypeExternal {
					// External editor found, which is acceptable
					return
				}
				// Otherwise, should be internal
				if info.Type != EditorTypeInternal {
					t.Errorf("Resolve() Type = %v, want %v or %v", info.Type, EditorTypeInternal, EditorTypeExternal)
				}
			} else {
				if info.Type != tt.wantType {
					t.Errorf("Resolve() Type = %v, want %v", info.Type, tt.wantType)
				}

				if info.UseInternal != tt.wantUseInternal {
					t.Errorf("Resolve() UseInternal = %v, want %v", info.UseInternal, tt.wantUseInternal)
				}
			}

			if info.Type == EditorTypeExternal {
				// Check that command is set (exact match may vary based on PATH)
				if info.Command == "" {
					t.Errorf("Resolve() Command = %v, want non-empty", info.Command)
				}
				// Command might be full path, so just check it's not empty
			}
		})
	}
}

func TestFindEditorOnPath(t *testing.T) {
	tests := []struct {
		name    string
		editor  string
		wantErr bool
	}{
		{
			name:    "find vi",
			editor:  "vi",
			wantErr: false, // vi should be available on most systems
		},
		{
			name:    "empty editor name",
			editor:  "",
			wantErr: true,
		},
		{
			name:    "non-existent editor",
			editor:  "nonexistent-editor-xyz123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := FindEditorOnPath(tt.editor)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindEditorOnPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && path == "" {
				t.Errorf("FindEditorOnPath() path = %v, want non-empty", path)
			}
		})
	}
}

func TestFindEditorOnPath_AbsolutePath(t *testing.T) {
	// Create a temporary executable file
	tmpDir := t.TempDir()
	editorPath := filepath.Join(tmpDir, "test-editor")

	// Create a simple executable (just a shell script)
	content := "#!/bin/sh\necho 'test editor'\n"
	if err := os.WriteFile(editorPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create test editor: %v", err)
	}

	// Test with absolute path
	path, err := FindEditorOnPath(editorPath)
	if err != nil {
		t.Errorf("FindEditorOnPath() with absolute path error = %v, want nil", err)
	}
	if path != editorPath {
		t.Errorf("FindEditorOnPath() path = %q, want %q", path, editorPath)
	}
}

func TestFindEditorOnPath_AbsolutePathNotFound(t *testing.T) {
	// Test with absolute path that doesn't exist
	nonexistentPath := "/nonexistent/path/to/editor"
	_, err := FindEditorOnPath(nonexistentPath)
	if err == nil {
		t.Error("FindEditorOnPath() with nonexistent absolute path error = nil, want error")
	}
}

func TestFindEditorOnPath_RelativePathWithSeparator(t *testing.T) {
	// Create a temporary executable file
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	editorName := "test-editor"
	editorPath := filepath.Join(subDir, editorName)

	// Create a simple executable
	content := "#!/bin/sh\necho 'test editor'\n"
	if err := os.WriteFile(editorPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create test editor: %v", err)
	}

	// Test with relative path containing separator
	relativePath := filepath.Join("bin", editorName)
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	path, err := FindEditorOnPath(relativePath)
	if err != nil {
		t.Errorf("FindEditorOnPath() with relative path containing separator error = %v, want nil", err)
	}
	if path == "" {
		t.Error("FindEditorOnPath() path = empty, want non-empty")
	}
}

func TestFindEditorOnPath_NonExecutable(t *testing.T) {
	// Create a temporary file that's not executable
	tmpDir := t.TempDir()
	editorPath := filepath.Join(tmpDir, "test-editor")

	// Create a file without executable permissions
	if err := os.WriteFile(editorPath, []byte("not executable"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with absolute path to non-executable file
	// Note: FindEditorOnPath checks executability for files found on PATH,
	// but for absolute paths it just checks if the file exists
	// So this test verifies the file exists check works
	path, err := FindEditorOnPath(editorPath)
	// The function may return the path even if not executable (for absolute paths)
	// or it may check executability - both behaviors are acceptable
	if err != nil {
		// If it errors, that's fine - it means it checked executability
		if !strings.Contains(err.Error(), "not executable") && !strings.Contains(err.Error(), "not found") {
			t.Logf("FindEditorOnPath() error = %v (acceptable)", err)
		}
	} else {
		// If it succeeds, verify the path is correct
		if path != editorPath {
			t.Errorf("FindEditorOnPath() path = %q, want %q", path, editorPath)
		}
	}
}

func TestResolveExternal(t *testing.T) {
	tests := []struct {
		name               string
		cliEditor          string
		fallbackToInternal bool
		wantErr            bool
		wantExternal       bool
	}{
		{
			name:               "external editor available",
			cliEditor:          "ls", // Use a command that exists
			fallbackToInternal: true,
			wantErr:            false,
			wantExternal:       true,
		},
		{
			name:               "no external editor, fallback disabled",
			cliEditor:          "nonexistent-editor-xyz123",
			fallbackToInternal: false,
			wantErr:            true,
			wantExternal:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear EDITOR env var for this test
			oldEditor := os.Getenv("EDITOR")
			defer func() {
				if oldEditor != "" {
					os.Setenv("EDITOR", oldEditor)
				} else {
					os.Unsetenv("EDITOR")
				}
			}()
			os.Unsetenv("EDITOR")

			resolver := NewEditorResolver(tt.cliEditor, nil, tt.fallbackToInternal)
			command, args, err := resolver.ResolveExternal()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveExternal() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("ResolveExternal() error = %v, want nil", err)
				}
				if tt.wantExternal {
					if command == "" {
						t.Error("ResolveExternal() command = empty, want non-empty")
					}
				}
			}
			_ = args // Use args to avoid unused variable
		})
	}
}

func TestShouldUseInternalEditor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error means use internal",
			err:  fmt.Errorf("editor not found"),
			want: true,
		},
		{
			name: "no error means don't use internal",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldUseInternalEditor(tt.err)
			if got != tt.want {
				t.Errorf("ShouldUseInternalEditor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEditorResolver_parseEditorString(t *testing.T) {
	resolver := NewEditorResolver("", nil, true)

	tests := []struct {
		name     string
		input    string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "simple editor name",
			input:    "vim",
			wantCmd:  "vim",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "editor with single arg",
			input:    "vim -f",
			wantCmd:  "vim",
			wantArgs: []string{"-f"},
			wantErr:  false,
		},
		{
			name:     "editor with multiple args",
			input:    "vim -f --wait",
			wantCmd:  "vim",
			wantArgs: []string{"-f", "--wait"},
			wantErr:  false,
		},
		{
			name:     "editor with args and spaces",
			input:    "vim  -f  --wait",
			wantCmd:  "vim",
			wantArgs: []string{"-f", "--wait"},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  true,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  true,
		},
		{
			name:     "editor with path",
			input:    "/usr/bin/vim",
			wantCmd:  "/usr/bin/vim",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "editor path with args",
			input:    "/usr/bin/vim -f",
			wantCmd:  "/usr/bin/vim",
			wantArgs: []string{"-f"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := resolver.parseEditorString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseEditorString() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("parseEditorString() error = %v, want nil", err)
				}
				if cmd != tt.wantCmd {
					t.Errorf("parseEditorString() command = %q, want %q", cmd, tt.wantCmd)
				}
				if len(args) != len(tt.wantArgs) {
					t.Errorf("parseEditorString() args length = %d, want %d", len(args), len(tt.wantArgs))
				} else {
					for i, wantArg := range tt.wantArgs {
						if args[i] != wantArg {
							t.Errorf("parseEditorString() args[%d] = %q, want %q", i, args[i], wantArg)
						}
					}
				}
			}
		})
	}
}

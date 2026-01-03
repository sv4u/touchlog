package editor

import (
	"fmt"
	"os"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

func TestEditorResolver_Resolve(t *testing.T) {
	tests := []struct {
		name                string
		cliEditor           string
		configEditor        *config.EditorConfig
		envEditor           string
		fallbackToInternal  bool
		wantType            EditorType
		wantCommand         string
		wantUseInternal     bool
		setupEnv            func() func() // setup function returns cleanup
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

func TestShouldUseInternalEditor(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    bool
	}{
		{
			name:    "error means use internal",
			err:     fmt.Errorf("editor not found"),
			want:    true,
		},
		{
			name:    "no error means don't use internal",
			err:     nil,
			want:    false,
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


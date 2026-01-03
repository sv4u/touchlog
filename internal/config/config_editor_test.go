package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEditorConfig_UnmarshalYAML_String(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantCmd string
		wantArgs []string
	}{
		{
			name:    "simple editor name",
			yaml:    `editor: "vim"`,
			wantCmd: "vim",
			wantArgs: []string{},
		},
		{
			name:    "editor with arguments",
			yaml:    `editor: "vim -f"`,
			wantCmd: "vim",
			wantArgs: []string{"-f"},
		},
		{
			name:    "editor with multiple arguments",
			yaml:    `editor: "code --wait"`,
			wantCmd: "code",
			wantArgs: []string{"--wait"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := yaml.Unmarshal([]byte(tt.yaml), &cfg)
			if err != nil {
				t.Fatalf("UnmarshalYAML() error = %v", err)
			}

			if cfg.Editor == nil {
				t.Fatal("Editor config is nil")
			}

			if cfg.Editor.Command != tt.wantCmd {
				t.Errorf("Editor.Command = %v, want %v", cfg.Editor.Command, tt.wantCmd)
			}

			if len(cfg.Editor.Args) != len(tt.wantArgs) {
				t.Errorf("Editor.Args length = %v, want %v", len(cfg.Editor.Args), len(tt.wantArgs))
			}

			for i, arg := range tt.wantArgs {
				if i >= len(cfg.Editor.Args) {
					t.Errorf("Editor.Args[%d] missing, want %v", i, arg)
					continue
				}
				if cfg.Editor.Args[i] != arg {
					t.Errorf("Editor.Args[%d] = %v, want %v", i, cfg.Editor.Args[i], arg)
				}
			}
		})
	}
}

func TestEditorConfig_UnmarshalYAML_Object(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantCmd string
		wantArgs []string
	}{
		{
			name:    "editor object with command only",
			yaml:    "editor:\n  command: vim",
			wantCmd: "vim",
			wantArgs: []string{},
		},
		{
			name:    "editor object with command and args",
			yaml:    "editor:\n  command: vim\n  args:\n    - -f",
			wantCmd: "vim",
			wantArgs: []string{"-f"},
		},
		{
			name:    "editor object with multiple args",
			yaml:    "editor:\n  command: code\n  args:\n    - --wait\n    - -n",
			wantCmd: "code",
			wantArgs: []string{"--wait", "-n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := yaml.Unmarshal([]byte(tt.yaml), &cfg)
			if err != nil {
				t.Fatalf("UnmarshalYAML() error = %v", err)
			}

			if cfg.Editor == nil {
				t.Fatal("Editor config is nil")
			}

			if cfg.Editor.Command != tt.wantCmd {
				t.Errorf("Editor.Command = %v, want %v", cfg.Editor.Command, tt.wantCmd)
			}

			if len(cfg.Editor.Args) != len(tt.wantArgs) {
				t.Errorf("Editor.Args length = %v, want %v", len(cfg.Editor.Args), len(tt.wantArgs))
			}

			for i, arg := range tt.wantArgs {
				if i >= len(cfg.Editor.Args) {
					t.Errorf("Editor.Args[%d] missing, want %v", i, arg)
					continue
				}
				if cfg.Editor.Args[i] != arg {
					t.Errorf("Editor.Args[%d] = %v, want %v", i, cfg.Editor.Args[i], arg)
				}
			}
		})
	}
}

func TestConfig_GetEditor(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		wantNil  bool
		wantCmd  string
		wantArgs []string
	}{
		{
			name:    "editor not set",
			config:  CreateDefaultConfig(),
			wantNil: true,
		},
		{
			name: "editor set",
			config: &Config{
				Editor: &EditorConfig{
					Command: "vim",
					Args:    []string{"-f"},
				},
			},
			wantNil:  false,
			wantCmd:  "vim",
			wantArgs: []string{"-f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := tt.config.GetEditor()
			if (editor == nil) != tt.wantNil {
				t.Errorf("GetEditor() = %v, want nil = %v", editor, tt.wantNil)
			}
			if !tt.wantNil {
				if editor.Command != tt.wantCmd {
					t.Errorf("GetEditor().Command = %v, want %v", editor.Command, tt.wantCmd)
				}
				if len(editor.Args) != len(tt.wantArgs) {
					t.Errorf("GetEditor().Args length = %v, want %v", len(editor.Args), len(tt.wantArgs))
				}
			}
		})
	}
}


package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sv4u/touchlog/internal/config"
)

// EditorType represents the type of editor
type EditorType int

const (
	// EditorTypeExternal represents an external editor
	EditorTypeExternal EditorType = iota
	// EditorTypeInternal represents the internal Bubble Tea editor
	EditorTypeInternal
)

// EditorInfo contains information about the resolved editor
type EditorInfo struct {
	Type     EditorType
	Command  string
	Args     []string
	UseInternal bool
}

// EditorResolver handles editor resolution with precedence chain
type EditorResolver struct {
	cliEditor         string
	configEditor      *config.EditorConfig
	envEditor         string
	fallbackToInternal bool
}

// NewEditorResolver creates a new editor resolver
func NewEditorResolver(cliEditor string, cfg *config.Config, fallbackToInternal bool) *EditorResolver {
	var configEditor *config.EditorConfig
	if cfg != nil {
		configEditor = cfg.GetEditor()
	}

	envEditor := os.Getenv("EDITOR")

	return &EditorResolver{
		cliEditor:          cliEditor,
		configEditor:       configEditor,
		envEditor:          envEditor,
		fallbackToInternal: fallbackToInternal,
	}
}

// Resolve resolves the editor using the precedence chain:
// 1. CLI --editor flag
// 2. EDITOR environment variable
// 3. Config editor setting
// 4. Try vi on PATH
// 5. Try nano on PATH
// 6. Fallback to internal editor (if fallbackToInternal is true)
func (r *EditorResolver) Resolve() (*EditorInfo, error) {
	// 1. Try CLI --editor flag
	if r.cliEditor != "" {
		command, args, err := r.parseEditorString(r.cliEditor)
		if err != nil {
			return nil, fmt.Errorf("invalid CLI editor: %w", err)
		}
		path, err := FindEditorOnPath(command)
		if err == nil {
			return &EditorInfo{
				Type:    EditorTypeExternal,
				Command: path,
				Args:    args,
			}, nil
		}
		// CLI editor specified but not found - return error (don't fallback)
		return nil, fmt.Errorf("editor '%s' not found on PATH", command)
	}

	// 2. Try EDITOR environment variable
	if r.envEditor != "" {
		command, args, err := r.parseEditorString(r.envEditor)
		if err != nil {
			return nil, fmt.Errorf("invalid EDITOR environment variable: %w", err)
		}
		path, err := FindEditorOnPath(command)
		if err == nil {
			return &EditorInfo{
				Type:    EditorTypeExternal,
				Command: path,
				Args:    args,
			}, nil
		}
		// EDITOR env var set but not found - continue to next option
	}

	// 3. Try Config editor setting
	if r.configEditor != nil && r.configEditor.Command != "" {
		path, err := FindEditorOnPath(r.configEditor.Command)
		if err == nil {
			return &EditorInfo{
				Type:    EditorTypeExternal,
				Command: path,
				Args:    r.configEditor.Args,
			}, nil
		}
		// Config editor set but not found - continue to next option
	}

	// 4. Try vi on PATH
	if path, err := FindEditorOnPath("vi"); err == nil {
		return &EditorInfo{
			Type:    EditorTypeExternal,
			Command: path,
			Args:    []string{},
		}, nil
	}

	// 5. Try nano on PATH
	if path, err := FindEditorOnPath("nano"); err == nil {
		return &EditorInfo{
			Type:    EditorTypeExternal,
			Command: path,
			Args:    []string{},
		}, nil
	}

	// 6. Fallback to internal editor
	if r.fallbackToInternal {
		return &EditorInfo{
			Type:        EditorTypeInternal,
			UseInternal: true,
		}, nil
	}

	// No editor found and fallback disabled
	return nil, fmt.Errorf("no editor found and fallback to internal editor is disabled")
}

// ResolveExternal resolves only external editors (does not fallback to internal)
func (r *EditorResolver) ResolveExternal() (string, []string, error) {
	info, err := r.Resolve()
	if err != nil {
		return "", nil, err
	}
	if info.Type == EditorTypeInternal {
		return "", nil, fmt.Errorf("no external editor available")
	}
	return info.Command, info.Args, nil
}

// parseEditorString parses a string like "vim" or "vim -f" into command and args
func (r *EditorResolver) parseEditorString(s string) (string, []string, error) {
	if s == "" {
		return "", nil, fmt.Errorf("empty editor string")
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty editor string")
	}
	command := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}
	return command, args, nil
}

// FindEditorOnPath finds an editor executable on PATH
func FindEditorOnPath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("editor name cannot be empty")
	}

	// If name contains a path separator, treat it as a full path
	if strings.Contains(name, string(filepath.Separator)) {
		// Check if it's an absolute path
		if filepath.IsAbs(name) {
			if _, err := os.Stat(name); err == nil {
				return name, nil
			}
			return "", fmt.Errorf("editor not found at path: %s", name)
		}
		// Relative path - resolve relative to current directory
		absPath, err := filepath.Abs(name)
		if err != nil {
			return "", fmt.Errorf("failed to resolve editor path: %w", err)
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
		return "", fmt.Errorf("editor not found at path: %s", absPath)
	}

	// Search on PATH
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("editor '%s' not found on PATH", name)
	}

	// Verify the file is executable
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat editor: %w", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		return "", fmt.Errorf("editor '%s' is not executable", path)
	}

	return path, nil
}

// ShouldUseInternalEditor determines if we should use the internal editor based on error
func ShouldUseInternalEditor(externalErr error) bool {
	// If external editor resolution failed, we can fallback to internal
	return externalErr != nil
}


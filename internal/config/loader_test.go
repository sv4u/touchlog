package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/xdg"
)

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("explicit path exists", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "explicit.yaml")
		cfg := CreateDefaultConfig()
		if err := SaveConfig(cfg, configPath); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		found, err := FindConfigFile(configPath)
		if err != nil {
			t.Errorf("FindConfigFile() error = %v, want nil", err)
		}
		if found != configPath {
			t.Errorf("FindConfigFile() = %q, want %q", found, configPath)
		}
	})

	t.Run("explicit path does not exist", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "nonexistent.yaml")
		found, err := FindConfigFile(configPath)
		if err == nil {
			t.Error("FindConfigFile() expected error for nonexistent explicit path, got nil")
		}
		if found != "" {
			t.Errorf("FindConfigFile() = %q, want empty string", found)
		}
	})

	t.Run("current directory - touchlog.yaml", func(t *testing.T) {
		// Change to temp directory
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd() error = %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWd)
		}()

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("os.Chdir() error = %v", err)
		}

		configPath := filepath.Join(tmpDir, "touchlog.yaml")
		cfg := CreateDefaultConfig()
		if err := SaveConfig(cfg, configPath); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		found, err := FindConfigFile("")
		if err != nil {
			t.Errorf("FindConfigFile() error = %v, want nil", err)
		}
		// Verify the found file exists and is the one we created
		if found == "" {
			t.Error("FindConfigFile() returned empty string, want config path")
		}
		if _, err := os.Stat(found); err != nil {
			t.Errorf("FindConfigFile() returned path that doesn't exist: %q", found)
		}
		// Verify it's the touchlog.yaml file we created
		if filepath.Base(found) != "touchlog.yaml" {
			t.Errorf("FindConfigFile() returned wrong file: %q, want touchlog.yaml", filepath.Base(found))
		}
	})

	t.Run("current directory - touchlog.yml", func(t *testing.T) {
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd() error = %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWd)
		}()

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("os.Chdir() error = %v", err)
		}

		// Remove touchlog.yaml if it exists
		_ = os.Remove(filepath.Join(tmpDir, "touchlog.yaml"))

		configPath := filepath.Join(tmpDir, "touchlog.yml")
		cfg := CreateDefaultConfig()
		if err := SaveConfig(cfg, configPath); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		found, err := FindConfigFile("")
		if err != nil {
			t.Errorf("FindConfigFile() error = %v, want nil", err)
		}
		// Verify the found file exists and is the one we created
		if found == "" {
			t.Error("FindConfigFile() returned empty string, want config path")
		}
		if _, err := os.Stat(found); err != nil {
			t.Errorf("FindConfigFile() returned path that doesn't exist: %q", found)
		}
		// Verify it's the touchlog.yml file we created
		if filepath.Base(found) != "touchlog.yml" {
			t.Errorf("FindConfigFile() returned wrong file: %q, want touchlog.yml", filepath.Base(found))
		}
	})

	t.Run("no config file found", func(t *testing.T) {
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd() error = %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWd)
		}()

		emptyDir := t.TempDir()
		if err := os.Chdir(emptyDir); err != nil {
			t.Fatalf("os.Chdir() error = %v", err)
		}

		// This test may find a config file in XDG directory if one exists
		// So we just verify it doesn't error and returns a path (empty or not)
		found, err := FindConfigFile("")
		if err != nil {
			t.Errorf("FindConfigFile() error = %v, want nil", err)
		}
		// Note: If XDG config exists, it will be returned, which is valid behavior
		// We just verify no error occurred
		_ = found // Accept any result (empty or XDG path)
	})

	t.Run("FindConfigFile does not create XDG directories (Bug 2)", func(t *testing.T) {
		// This test verifies that FindConfigFile is read-only and doesn't create directories
		// as a side effect when searching for config files
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd() error = %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWd)
		}()

		emptyDir := t.TempDir()
		if err := os.Chdir(emptyDir); err != nil {
			t.Fatalf("os.Chdir() error = %v", err)
		}

		// Get the expected XDG config directory path (without creating it)
		// We'll use the xdg package's read-only function to get the path
		xdgConfigPath := xdg.ConfigFilePathReadOnly()
		xdgConfigDir := filepath.Dir(xdgConfigPath)

		// Check if the XDG config directory already exists before calling FindConfigFile
		// If it exists, we can't reliably test this (we can't distinguish between
		// "was created by FindConfigFile" and "already existed")
		// In that case, we skip the test to avoid modifying user data
		if _, err := os.Stat(xdgConfigDir); err == nil {
			// Directory already exists - skip this test to avoid modifying user data
			// This is safe because if the directory exists, FindConfigFile won't create it anyway
			t.Skipf("XDG config directory %q already exists - skipping test to avoid modifying user data", xdgConfigDir)
		}

		// Directory doesn't exist - verify FindConfigFile doesn't create it
		// Call FindConfigFile - this should NOT create any directories
		found, err := FindConfigFile("")
		if err != nil {
			t.Errorf("FindConfigFile() error = %v, want nil", err)
		}

		// Verify the XDG config directory was NOT created
		if _, err := os.Stat(xdgConfigDir); err == nil {
			t.Errorf("FindConfigFile() created XDG config directory %q as side effect (Bug 2)", xdgConfigDir)
		}

		// Verify FindConfigFile returned empty (no config found, which is expected)
		// since we're in an empty temp directory
		if found != "" {
			// If a config file was found, it means one exists in XDG directory
			// That's okay, but we should verify the directory wasn't created by this call
			// The key is that the directory should not exist if it didn't before
		}
	})
}

func TestDetectConfigFormat(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		want     ConfigFormat
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "yaml extension",
			path:    "config.yaml",
			want:    FormatYAML,
			wantErr: false,
		},
		{
			name:    "yml extension",
			path:    "config.yml",
			want:    FormatYAML,
			wantErr: false,
		},
		{
			name:    "toml extension",
			path:    "config.toml",
			want:    FormatTOML,
			wantErr: false,
		},
		{
			name:    "unknown extension",
			path:    "config.json",
			want:    FormatUnknown,
			wantErr: true,
			errMsg:  "unknown config format",
		},
		{
			name:    "no extension",
			path:    "config",
			want:    FormatUnknown,
			wantErr: true,
			errMsg:  "unknown config format",
		},
		{
			name:    "empty path",
			path:    "",
			want:    FormatUnknown,
			wantErr: true,
			errMsg:  "empty path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectConfigFormat(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectConfigFormat() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DetectConfigFormat() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("DetectConfigFormat() error = %v, want nil", err)
				}
				if got != tt.want {
					t.Errorf("DetectConfigFormat() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestLoadConfigFromPath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("load YAML config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "config.yaml")
		cfg := CreateDefaultConfig()
		if err := SaveConfig(cfg, configPath); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		loaded, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadConfigFromPath() error = %v", err)
		}
		if loaded == nil {
			t.Fatal("LoadConfigFromPath() returned nil config")
		}
		if len(loaded.Templates) != len(cfg.Templates) {
			t.Errorf("LoadConfigFromPath() templates length = %d, want %d", len(loaded.Templates), len(cfg.Templates))
		}
	})

	t.Run("empty path returns default config", func(t *testing.T) {
		loaded, err := LoadConfigFromPath("")
		if err != nil {
			t.Fatalf("LoadConfigFromPath() error = %v", err)
		}
		if loaded == nil {
			t.Fatal("LoadConfigFromPath() returned nil config")
		}
		// Should return default config
		if len(loaded.Templates) != 3 {
			t.Errorf("LoadConfigFromPath() templates length = %d, want 3 (default)", len(loaded.Templates))
		}
	})

	t.Run("TOML format returns error", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "config.toml")
		// Create empty file to test format detection
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := LoadConfigFromPath(configPath)
		if err == nil {
			t.Error("LoadConfigFromPath() expected error for TOML format, got nil")
		}
		if !strings.Contains(err.Error(), "TOML format not yet supported") {
			t.Errorf("LoadConfigFromPath() error = %v, want error containing 'TOML format not yet supported'", err)
		}
	})
}



package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sv4u/touchlog/internal/xdg"
)

// ConfigFormat represents the format of a configuration file
type ConfigFormat string

const (
	// FormatYAML represents YAML format
	FormatYAML ConfigFormat = "yaml"
	// FormatTOML represents TOML format (future)
	FormatTOML ConfigFormat = "toml"
	// FormatUnknown represents an unknown format
	FormatUnknown ConfigFormat = "unknown"
)

// FindConfigFile searches for a configuration file in the following order:
// 1. Explicit path (if provided)
// 2. Current directory: ./touchlog.yaml or ./touchlog.toml
// 3. XDG config directory: $XDG_CONFIG_HOME/touchlog/config.yaml
// Returns the path if found, empty string if not found, and an error if there's a problem
func FindConfigFile(explicitPath string) (string, error) {
	// 1. Check explicit path first
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		// If explicit path is provided but doesn't exist, return error
		return "", fmt.Errorf("config file not found at explicit path: %s", explicitPath)
	}

	// 2. Check current directory for touchlog.yaml or touchlog.toml
	wd, err := os.Getwd()
	if err == nil {
		yamlPath := filepath.Join(wd, "touchlog.yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			return yamlPath, nil
		}

		ymlPath := filepath.Join(wd, "touchlog.yml")
		if _, err := os.Stat(ymlPath); err == nil {
			return ymlPath, nil
		}

		tomlPath := filepath.Join(wd, "touchlog.toml")
		if _, err := os.Stat(tomlPath); err == nil {
			return tomlPath, nil
		}
	}

	// 3. Check XDG config directory (read-only, don't create directories)
	configPath := xdg.ConfigFilePathReadOnly()
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// No config file found - return empty string (not an error, defaults will be used)
	return "", nil
}

// DetectConfigFormat detects the format of a configuration file based on its extension
func DetectConfigFormat(path string) (ConfigFormat, error) {
	if path == "" {
		return FormatUnknown, fmt.Errorf("empty path provided")
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return FormatYAML, nil
	case ".toml":
		return FormatTOML, nil
	default:
		return FormatUnknown, fmt.Errorf("unknown config format for file: %s", path)
	}
}

// LoadConfigFromPath loads a configuration file from the specified path
// It automatically detects the format and loads accordingly
func LoadConfigFromPath(path string) (*Config, error) {
	if path == "" {
		// No config file - return default config
		return CreateDefaultConfig(), nil
	}

	format, err := DetectConfigFormat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect config format: %w", err)
	}

	switch format {
	case FormatYAML:
		return LoadConfig(path)
	case FormatTOML:
		// TOML support will be added later (Phase 2.2)
		return nil, fmt.Errorf("TOML format not yet supported")
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}
}

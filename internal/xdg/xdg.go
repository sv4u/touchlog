package xdg

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// ConfigDir returns the configuration directory path
// In Go, functions can return multiple values
// Here we return a string (path) and an error
func ConfigDir() (string, error) {
	// xdg.ConfigHome returns something like ~/.config
	// We append "touchlog" to get ~/.config/touchlog
	configPath := filepath.Join(xdg.ConfigHome, "touchlog")

	// Ensure the directory exists
	// os.MkdirAll creates all parent directories if needed
	// 0755 is the permission mode (read/write/execute for owner, read/execute for others)
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		// If creation fails, return the error
		return "", err
	}

	// Return the path and nil (no error)
	return configPath, nil
}

// DataDir returns the data directory path
// Similar structure to ConfigDir
func DataDir() (string, error) {
	dataPath := filepath.Join(xdg.DataHome, "touchlog")
	err := os.MkdirAll(dataPath, 0755)
	if err != nil {
		return "", err
	}
	return dataPath, nil
}

// ConfigFilePath returns the full path to the config file
// This function creates the config directory if it doesn't exist
func ConfigFilePath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	// filepath.Join safely joins path components
	return filepath.Join(configDir, "config.yaml"), nil
}

// ConfigFilePathReadOnly returns the full path to the config file without creating directories
// This is a read-only operation that should be used for searching/checking if a config file exists
func ConfigFilePathReadOnly() string {
	// xdg.ConfigHome returns something like ~/.config
	// We append "touchlog" to get ~/.config/touchlog
	configPath := filepath.Join(xdg.ConfigHome, "touchlog", "config.yaml")
	return configPath
}

// TemplatesDir returns the templates directory path
func TemplatesDir() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	templatesPath := filepath.Join(dataDir, "templates")
	// Ensure the templates directory exists
	err = os.MkdirAll(templatesPath, 0755)
	if err != nil {
		return "", err
	}
	return templatesPath, nil
}

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DateTimeVarConfig represents configuration for a date/time variable
type DateTimeVarConfig struct {
	Enabled bool   `yaml:"enabled"` // Whether this variable is enabled
	Format  string `yaml:"format"`  // Go time format string
}

// DateTimeVarsConfig represents configuration for date/time/datetime variables
type DateTimeVarsConfig struct {
	Date     DateTimeVarConfig `yaml:"date"`     // Date variable configuration
	Time     DateTimeVarConfig `yaml:"time"`     // Time variable configuration
	DateTime DateTimeVarConfig `yaml:"datetime"` // DateTime variable configuration
}

// Config represents the application configuration
// Struct fields with uppercase names are exported (public)
type Config struct {
	Templates      []Template         `yaml:"templates"`       // YAML tag maps to "templates" key
	NotesDirectory string             `yaml:"notes_directory"` // YAML tag maps to "notes_directory" key
	DateTimeVars   DateTimeVarsConfig `yaml:"datetime_vars"`   // Date/time/datetime variable configuration
	Variables      map[string]string  `yaml:"variables"`       // Custom static variables
	VimMode        bool               `yaml:"vim_mode"`        // Enable vim keymap support
}

// Template represents a single template definition
type Template struct {
	Name string `yaml:"name"` // Template display name
	File string `yaml:"file"` // Template filename
}

// LoadConfig reads and parses the YAML configuration file
// It returns a pointer to Config (*Config) and an error
func LoadConfig(path string) (*Config, error) {
	// Read the file
	// os.ReadFile returns file contents as []byte and an error
	data, err := os.ReadFile(path)
	if err != nil {
		// Wrap error with context using fmt.Errorf
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create an empty Config struct
	var cfg Config

	// Unmarshal YAML data into the struct
	// yaml.Unmarshal parses YAML and populates the struct fields
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate custom variable names
	if cfg.Variables != nil {
		for name := range cfg.Variables {
			if err := ValidateVariableName(name); err != nil {
				return nil, fmt.Errorf("invalid variable name: %w", err)
			}
		}
	}

	// Return pointer to config and nil error
	// &cfg gets the address (pointer) of the cfg variable
	return &cfg, nil
}

// GetTemplates returns the list of available templates
// This is a method on Config (receiver: c *Config)
func (c *Config) GetTemplates() []Template {
	return c.Templates
}

// GetNotesDirectory returns the configured notes directory
func (c *Config) GetNotesDirectory() string {
	return c.NotesDirectory
}

// GetDateTimeVars returns the date/time/datetime variable configuration
func (c *Config) GetDateTimeVars() DateTimeVarsConfig {
	return c.DateTimeVars
}

// GetVariables returns the custom variables map
func (c *Config) GetVariables() map[string]string {
	if c.Variables == nil {
		return make(map[string]string)
	}
	return c.Variables
}

// GetVimMode returns whether vim mode is enabled
func (c *Config) GetVimMode() bool {
	return c.VimMode
}

// ValidateVariableName checks if a variable name conflicts with reserved names
// Reserved names are: "date", "time", "datetime"
func ValidateVariableName(name string) error {
	reserved := map[string]bool{
		"date":     true,
		"time":     true,
		"datetime": true,
	}
	if reserved[name] {
		return fmt.Errorf("variable name '%s' is reserved and cannot be used", name)
	}
	return nil
}

// ValidateTimeFormat validates a Go time format string by attempting to format the current time
// Returns true if the format is valid, false otherwise
func ValidateTimeFormat(format string) bool {
	if format == "" {
		return false
	}
	// Try to format the current time with the given format
	// If it panics, the format is invalid
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	testTime := time.Now()
	_ = testTime.Format(format)
	return !panicked
}

// CreateDefaultConfig creates a default configuration with sensible defaults
// The returned config includes three common templates (Daily Note, Meeting Notes, Journal),
// a default notes directory, all date/time variables enabled with standard formats,
// an empty custom variables map, and vim mode disabled.
func CreateDefaultConfig() *Config {
	return &Config{
		Templates: []Template{
			{Name: "Daily Note", File: "daily.md"},
			{Name: "Meeting Notes", File: "meeting.md"},
			{Name: "Journal", File: "journal.md"},
		},
		NotesDirectory: "~/notes",
		DateTimeVars: DateTimeVarsConfig{
			Date:     DateTimeVarConfig{Enabled: true, Format: "2006-01-02"},
			Time:     DateTimeVarConfig{Enabled: true, Format: "15:04:05"},
			DateTime: DateTimeVarConfig{Enabled: true, Format: "2006-01-02 15:04:05"},
		},
		Variables: make(map[string]string),
		VimMode:   false,
	}
}

// SaveConfig writes the configuration to a YAML file at the specified path.
// It creates the parent directory if it doesn't exist and writes the file with
// standard permissions (0644). Returns an error if marshaling or writing fails.
func SaveConfig(cfg *Config, path string) error {
	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent directory exists
	// Note: xdg.ConfigDir() already creates the directory, but we ensure it here
	// as a safety measure in case the path is different
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file with standard permissions (read/write for owner, read for others)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadOrCreateConfig loads the configuration file from the specified path.
// If the file doesn't exist, it creates a default configuration file and returns it.
// For other errors (permission denied, invalid YAML, etc.), it returns the error unchanged.
// This function is intended for first-run scenarios where no config file exists yet.
func LoadOrCreateConfig(path string) (*Config, error) {
	// Try to load existing config
	cfg, err := LoadConfig(path)
	if err != nil {
		// Check if error is because file doesn't exist
		// LoadConfig wraps os.ReadFile errors, so we need to unwrap and check
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist) {
			// Config doesn't exist, create default
			cfg = CreateDefaultConfig()
			if err := SaveConfig(cfg, path); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			// Return the newly created config
			return cfg, nil
		}
		// Other error (permission denied, invalid YAML, etc.) - return as-is
		return nil, err
	}
	// Config loaded successfully
	return cfg, nil
}

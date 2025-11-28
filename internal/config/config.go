package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DateTimeVarConfig represents configuration for a date/time variable
type DateTimeVarConfig struct {
	Enabled bool   `yaml:"enabled"` // Whether this variable is enabled
	Format  string `yaml:"format"`   // Go time format string
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
	Templates      []Template        `yaml:"templates"`       // YAML tag maps to "templates" key
	NotesDirectory string            `yaml:"notes_directory"` // YAML tag maps to "notes_directory" key
	DateTimeVars   DateTimeVarsConfig `yaml:"datetime_vars"` // Date/time/datetime variable configuration
	Variables      map[string]string `yaml:"variables"`      // Custom static variables
	VimMode        bool              `yaml:"vim_mode"`       // Enable vim keymap support
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


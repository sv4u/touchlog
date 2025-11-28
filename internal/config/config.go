package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
// Struct fields with uppercase names are exported (public)
type Config struct {
	Templates      []Template `yaml:"templates"`       // YAML tag maps to "templates" key
	NotesDirectory string     `yaml:"notes_directory"` // YAML tag maps to "notes_directory" key
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


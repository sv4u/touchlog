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
	Enabled *bool  `yaml:"enabled"` // Whether this variable is enabled (nil = not specified, use default)
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
	// File-based templates (legacy format)
	Templates []Template `yaml:"templates"` // YAML tag maps to "templates" key
	// Inline templates (new format) - map of template name to template content
	InlineTemplates map[string]string `yaml:"inline_templates"` // Inline template definitions
	// Default template name to use
	DefaultTemplate string `yaml:"default_template"` // Default template name
	// Template configuration (for backward compatibility)
	Template struct {
		Name string `yaml:"name"` // Template name (alternative to default_template)
	} `yaml:"template"`
	NotesDirectory string             `yaml:"notes_directory"` // YAML tag maps to "notes_directory" key
	DateTimeVars   DateTimeVarsConfig `yaml:"datetime_vars"`   // Date/time/datetime variable configuration
	Variables      map[string]string  `yaml:"variables"`       // Custom static variables
	VimMode        bool               `yaml:"vim_mode"`        // Enable vim keymap support
	Timezone       string             `yaml:"timezone"`        // IANA timezone (e.g., "America/Denver", "UTC")
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

// GetInlineTemplates returns the inline templates map
func (c *Config) GetInlineTemplates() map[string]string {
	if c.InlineTemplates == nil {
		return make(map[string]string)
	}
	return c.InlineTemplates
}

// GetDefaultTemplate returns the default template name
// Checks both default_template and template.name fields for backward compatibility
func (c *Config) GetDefaultTemplate() string {
	if c.DefaultTemplate != "" {
		return c.DefaultTemplate
	}
	return c.Template.Name
}

// GetTimezone returns the configured timezone, or empty string if not set
func (c *Config) GetTimezone() string {
	return c.Timezone
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
	enabledTrue := true
	return &Config{
		Templates: []Template{
			{Name: "Daily Note", File: "daily.md"},
			{Name: "Meeting Notes", File: "meeting.md"},
			{Name: "Journal", File: "journal.md"},
		},
		NotesDirectory: "~/notes",
		DateTimeVars: DateTimeVarsConfig{
			Date:     DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
			Time:     DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
			DateTime: DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
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

// CLIFlags represents command-line flags that can override config values
type CLIFlags struct {
	OutputDir string
	Template  string
	Editor    string
	Timezone  string
}

// LoadWithPrecedence loads configuration with precedence: CLI flags > Config file > Defaults
// Precedence order: CLI flags override config values, config values override defaults
// No merging - first non-empty value wins
func LoadWithPrecedence(configPath string, cliFlags *CLIFlags) (*Config, error) {
	// Start with defaults
	cfg := CreateDefaultConfig()

	// Load config file if path is provided
	if configPath != "" {
		fileCfg, err := LoadConfigFromPath(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		// Apply config file values (override defaults)
		applyConfigOverrides(cfg, fileCfg)
	}

	// Apply CLI flags (override config and defaults)
	if cliFlags != nil {
		applyCLIOverrides(cfg, cliFlags)
	}

	return cfg, nil
}

// applyConfigOverrides applies values from fileCfg to cfg (file config overrides defaults)
func applyConfigOverrides(cfg, fileCfg *Config) {
	if fileCfg.NotesDirectory != "" {
		cfg.NotesDirectory = fileCfg.NotesDirectory
	}
	if len(fileCfg.Templates) > 0 {
		cfg.Templates = fileCfg.Templates
	}
	if len(fileCfg.InlineTemplates) > 0 {
		if cfg.InlineTemplates == nil {
			cfg.InlineTemplates = make(map[string]string)
		}
		for k, v := range fileCfg.InlineTemplates {
			cfg.InlineTemplates[k] = v
		}
	}
	if fileCfg.DefaultTemplate != "" {
		cfg.DefaultTemplate = fileCfg.DefaultTemplate
	}
	if fileCfg.Template.Name != "" {
		cfg.Template.Name = fileCfg.Template.Name
	}
	if fileCfg.Timezone != "" {
		cfg.Timezone = fileCfg.Timezone
	}
	if len(fileCfg.Variables) > 0 {
		if cfg.Variables == nil {
			cfg.Variables = make(map[string]string)
		}
		for k, v := range fileCfg.Variables {
			cfg.Variables[k] = v
		}
	}
	cfg.VimMode = fileCfg.VimMode
	// Merge DateTimeVars field by field to preserve defaults when config doesn't specify them
	// Check if datetime_vars section was present (indicated by at least one format string or explicit Enabled)
	// If the section wasn't present, preserve defaults
	hasDateTimeVars := fileCfg.DateTimeVars.Date.Format != "" ||
		fileCfg.DateTimeVars.Time.Format != "" ||
		fileCfg.DateTimeVars.DateTime.Format != "" ||
		fileCfg.DateTimeVars.Date.Enabled != nil ||
		fileCfg.DateTimeVars.Time.Enabled != nil ||
		fileCfg.DateTimeVars.DateTime.Enabled != nil
	if hasDateTimeVars {
		// datetime_vars section was present, merge field by field
		// Per Option B: Only process Enabled when Format is also specified
		if fileCfg.DateTimeVars.Date.Format != "" {
			cfg.DateTimeVars.Date.Format = fileCfg.DateTimeVars.Date.Format
			// If enabled is explicitly set (not nil), use that value
			// If enabled is nil (not specified), default to true when format is specified
			if fileCfg.DateTimeVars.Date.Enabled != nil {
				cfg.DateTimeVars.Date.Enabled = fileCfg.DateTimeVars.Date.Enabled
			} else {
				enabledTrue := true
				cfg.DateTimeVars.Date.Enabled = &enabledTrue
			}
		}
		if fileCfg.DateTimeVars.Time.Format != "" {
			cfg.DateTimeVars.Time.Format = fileCfg.DateTimeVars.Time.Format
			if fileCfg.DateTimeVars.Time.Enabled != nil {
				cfg.DateTimeVars.Time.Enabled = fileCfg.DateTimeVars.Time.Enabled
			} else {
				enabledTrue := true
				cfg.DateTimeVars.Time.Enabled = &enabledTrue
			}
		}
		if fileCfg.DateTimeVars.DateTime.Format != "" {
			cfg.DateTimeVars.DateTime.Format = fileCfg.DateTimeVars.DateTime.Format
			if fileCfg.DateTimeVars.DateTime.Enabled != nil {
				cfg.DateTimeVars.DateTime.Enabled = fileCfg.DateTimeVars.DateTime.Enabled
			} else {
				enabledTrue := true
				cfg.DateTimeVars.DateTime.Enabled = &enabledTrue
			}
		}
		// Note: If Enabled is set without Format, it's ignored per Option B
		// (only respect enabled: false when format is also specified)
	}
}

// applyCLIOverrides applies CLI flag values to cfg (CLI flags override config)
func applyCLIOverrides(cfg *Config, flags *CLIFlags) {
	if flags.OutputDir != "" {
		cfg.NotesDirectory = flags.OutputDir
	}
	if flags.Template != "" {
		cfg.DefaultTemplate = flags.Template
		cfg.Template.Name = flags.Template
	}
	if flags.Timezone != "" {
		cfg.Timezone = flags.Timezone
	}
	// Editor flag will be handled in Phase 5
}

// KnownConfigKeys is a list of all known configuration keys for strict mode validation
var KnownConfigKeys = []string{
	"templates",
	"inline_templates",
	"default_template",
	"template",
	"notes_directory",
	"datetime_vars",
	"variables",
	"vim_mode",
	"timezone",
}

// ValidateStrict validates that the config file doesn't contain unknown keys
// Returns an error if unknown keys are found, listing all unknown keys
func ValidateStrict(cfg *Config, configData map[string]interface{}) error {
	if configData == nil {
		return nil
	}

	var unknownKeys []string
	for key := range configData {
		if !isKnownKey(key) {
			unknownKeys = append(unknownKeys, key)
		}
	}

	if len(unknownKeys) > 0 {
		return fmt.Errorf("unknown config keys found: %v", unknownKeys)
	}

	return nil
}

// isKnownKey checks if a key is in the list of known configuration keys
func isKnownKey(key string) bool {
	for _, knownKey := range KnownConfigKeys {
		if key == knownKey {
			return true
		}
	}
	return false
}

// ValidateStrictFromYAML validates a YAML config file for unknown keys
// This is a convenience function that unmarshals YAML and validates it
func ValidateStrictFromYAML(data []byte) error {
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	var unknownKeys []string
	for key := range rawConfig {
		if !isKnownKey(key) {
			unknownKeys = append(unknownKeys, key)
		}
	}

	if len(unknownKeys) > 0 {
		return fmt.Errorf("unknown config keys found: %v", unknownKeys)
	}

	return nil
}

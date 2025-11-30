package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sv4u/touchlog/internal/config"
)

// LoadTemplate reads a template file from the templates directory
func LoadTemplate(templatesDir, filename string) (string, error) {
	templatePath := filepath.Join(templatesDir, filename)

	// Read file contents
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", filename, err)
	}

	// Convert []byte to string
	return string(content), nil
}

// ProcessTemplate replaces {{variable}} placeholders with actual values
func ProcessTemplate(templateContent string, vars map[string]string) string {
	result := templateContent

	// Iterate over the variables map
	// range over a map gives key and value
	for key, value := range vars {
		// Replace {{key}} with value
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ExtractVariables finds all {{variable}} placeholders in the template
// Returns a slice (array) of variable names
func ExtractVariables(content string) []string {
	// Compile a regular expression
	// \{\{(\w+)\}\} matches {{ followed by word characters, followed by }}
	// The parentheses create a capture group
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)

	// Find all matches
	matches := re.FindAllStringSubmatch(content, -1)

	// Extract variable names from capture groups
	variables := make([]string, 0) // Make an empty slice
	for _, match := range matches {
		if len(match) > 1 {
			// match[0] is the full match, match[1] is the first capture group
			variables = append(variables, match[1])
		}
	}

	return variables
}

// GetDefaultVariables returns a map of default variable values based on configuration
// If cfg is nil, uses default behavior (all variables enabled with default formats)
func GetDefaultVariables(cfg *config.Config) map[string]string {
	now := time.Now()
	vars := make(map[string]string)

	// Default formats
	defaultDateFormat := "2006-01-02"
	defaultTimeFormat := "15:04:05"
	defaultDateTimeFormat := "2006-01-02 15:04:05"

	// If config is nil, use default behavior (all enabled with default formats)
	if cfg == nil {
		return map[string]string{
			"date":     now.Format(defaultDateFormat),
			"time":     now.Format(defaultTimeFormat),
			"datetime": now.Format(defaultDateTimeFormat),
		}
	}

	// Get date/time configuration from config
	dtVars := cfg.GetDateTimeVars()

	// Date variable
	if dtVars.Date.Enabled {
		format := dtVars.Date.Format
		if format == "" {
			format = defaultDateFormat
		} else if !config.ValidateTimeFormat(format) {
			// Fallback to default format if validation fails
			format = defaultDateFormat
		}
		vars["date"] = now.Format(format)
	}

	// Time variable
	if dtVars.Time.Enabled {
		format := dtVars.Time.Format
		if format == "" {
			format = defaultTimeFormat
		} else if !config.ValidateTimeFormat(format) {
			// Fallback to default format if validation fails
			format = defaultTimeFormat
		}
		vars["time"] = now.Format(format)
	}

	// DateTime variable
	if dtVars.DateTime.Enabled {
		format := dtVars.DateTime.Format
		if format == "" {
			format = defaultDateTimeFormat
		} else if !config.ValidateTimeFormat(format) {
			// Fallback to default format if validation fails
			format = defaultDateTimeFormat
		}
		vars["datetime"] = now.Format(format)
	}

	// If no date/time variables are enabled, default to all enabled with default formats
	// This maintains backward compatibility
	if len(vars) == 0 {
		vars = map[string]string{
			"date":     now.Format(defaultDateFormat),
			"time":     now.Format(defaultTimeFormat),
			"datetime": now.Format(defaultDateTimeFormat),
		}
	}

	// Merge custom variables from config
	customVars := cfg.GetVariables()
	for key, value := range customVars {
		// Custom variables can override default variables
		vars[key] = value
	}

	return vars
}

// CreateExampleTemplates creates minimal inline template files in the specified directory
// if the directory is empty. It only creates templates that are referenced in the default
// config (daily.md, meeting.md, journal.md). If the directory already contains files,
// this function does nothing (non-destructive).
func CreateExampleTemplates(templatesDir string) error {
	// Check if directory exists and is empty
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		// Directory doesn't exist - xdg.TemplatesDir() should have created it,
		// but handle the case where it doesn't exist
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read templates directory: %w", err)
		}
		// Directory doesn't exist, create it
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			return fmt.Errorf("failed to create templates directory: %w", err)
		}
	} else if len(entries) > 0 {
		// Directory exists and is not empty - don't overwrite existing templates
		return nil
	}

	// Define minimal inline templates matching default config
	examples := map[string]string{
		"daily.md": `# Daily Note - {{date}}

## Events
- 

## Thoughts
- 

## Tasks
- [ ] 

## Notes
- 
`,
		"meeting.md": `# Meeting Notes - {{date}} {{time}}

## Attendees
- 

## Agenda
- 

## Notes
- 

## Action Items
- [ ] 
`,
		"journal.md": `# Journal Entry - {{date}}

## Today's Highlights
- 

## Reflections
- 

## Tomorrow's Focus
- 
`,
	}

	// Create each template file
	for filename, content := range examples {
		path := filepath.Join(templatesDir, filename)
		// Check if file already exists (shouldn't happen if directory was empty, but be safe)
		if _, err := os.Stat(path); err == nil {
			// File exists, skip it (non-destructive)
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create template %s: %w", filename, err)
		}
	}

	return nil
}

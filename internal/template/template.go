package template

import (
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/errors"
	"github.com/sv4u/touchlog/internal/validation"
)

// ErrTemplateNotFound is re-exported from the errors package for backward compatibility
var ErrTemplateNotFound = errors.ErrTemplateNotFound

// TemplateSource is an interface for template sources (inline or file-based)
type TemplateSource interface {
	GetTemplate(name string) (string, error)
}

// InlineTemplateSource provides templates from inline config definitions
type InlineTemplateSource struct {
	templates map[string]string
}

// GetTemplate retrieves a template by name from inline templates
func (s *InlineTemplateSource) GetTemplate(name string) (string, error) {
	if s.templates == nil {
		return "", ErrTemplateNotFound
	}
	content, ok := s.templates[name]
	if !ok {
		return "", ErrTemplateNotFound
	}
	return content, nil
}

// FileTemplateSource provides templates from file-based template files
type FileTemplateSource struct {
	templatesDir string
	templates    []config.Template // List of available templates from config
}

// GetTemplate retrieves a template by name from file-based templates
// Matches by filename (without extension) - e.g., "daily" matches "daily.md"
func (s *FileTemplateSource) GetTemplate(name string) (string, error) {
	// Find the template in the config list by matching filename (without extension)
	var templateFile string
	for _, tmpl := range s.templates {
		// Remove extension from filename for comparison
		baseName := strings.TrimSuffix(tmpl.File, filepath.Ext(tmpl.File))
		if baseName == name {
			templateFile = tmpl.File
			break
		}
	}

	if templateFile == "" {
		return "", ErrTemplateNotFound
	}

	// Load the template file
	templatePath := filepath.Join(s.templatesDir, templateFile)
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templateFile, err)
	}

	return string(content), nil
}

// ResolveTemplate resolves a template by name, checking inline templates first,
// then falling back to file-based templates. Returns the template content or an error.
func ResolveTemplate(name string, cfg *config.Config, templatesDir string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("template name cannot be empty")
	}

	// Step 1: Check inline templates first
	inlineTemplates := cfg.GetInlineTemplates()
	if len(inlineTemplates) > 0 {
		inlineSource := &InlineTemplateSource{templates: inlineTemplates}
		content, err := inlineSource.GetTemplate(name)
		if err == nil {
			// Validate template syntax before returning
			if err := validation.ValidateTemplateSyntax(content); err != nil {
				return "", fmt.Errorf("invalid inline template syntax for '%s': %w", name, err)
			}
			return content, nil
		}
		// If not found in inline, continue to file-based
	}

	// Step 2: Fall back to file-based templates
	fileTemplates := cfg.GetTemplates()
	if len(fileTemplates) > 0 {
		fileSource := &FileTemplateSource{
			templatesDir: templatesDir,
			templates:    fileTemplates,
		}
		content, err := fileSource.GetTemplate(name)
		if err == nil {
			// Validate template syntax before returning
			if err := validation.ValidateTemplateSyntax(content); err != nil {
				return "", fmt.Errorf("invalid file-based template syntax for '%s': %w", name, err)
			}
			return content, nil
		}
		// If the error is not "template not found", it's a real file read error
		// (e.g., permission denied, I/O error) - propagate it immediately
		if !stderrors.Is(err, errors.ErrTemplateNotFound) {
			return "", fmt.Errorf("failed to resolve file-based template '%s': %w", name, err)
		}
		// If it's ErrTemplateNotFound, continue to fallback below
	}

	// Step 3: Template not found in either source
	return "", fmt.Errorf("template '%s' not found: %w", name, errors.ErrTemplateNotFound)
}

// LoadTemplate reads a template file from the templates directory
// This function is kept for backward compatibility but may be deprecated in favor of ResolveTemplate
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

// ProcessTemplate replaces {{variable}} placeholders with actual values.
// User-provided variables (title, message, tags, custom vars) are automatically escaped
// to prevent template injection. System variables (date, time, datetime, metadata) are trusted.
func ProcessTemplate(templateContent string, vars map[string]string) string {
	result := templateContent

	// Iterate over the variables map
	// range over a map gives key and value
	for key, value := range vars {
		// Escape user-provided variables to prevent template injection
		// System variables (date, time, datetime, metadata) are trusted and not escaped
		if ShouldEscapeVariable(key) {
			value = EscapeUserInput(value)
		}

		// Replace {{key}} with value
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Unescape any remaining escaped braces in the result
	// This restores literal {{ and }} that were in the original template or user input
	result = UnescapeUserInput(result)

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
// Returns an error if the configured timezone is invalid.
// If a timestamp is provided, it will be used instead of time.Now() to ensure consistency
// across filename generation and template variable substitution.
func GetDefaultVariables(cfg *config.Config, timestamp ...time.Time) (map[string]string, error) {
	return GetDefaultVariablesWithMetadata(cfg, nil, timestamp...)
}

// MetadataValues represents metadata values for template processing
// This avoids import cycle with entry package
type MetadataValues struct {
	User   string
	Host   string
	Branch string
	Commit string
}

// GetDefaultVariablesWithMetadata returns a map of default variable values including metadata
// Metadata can be nil or a *MetadataValues struct
func GetDefaultVariablesWithMetadata(cfg *config.Config, metadata *MetadataValues, timestamp ...time.Time) (map[string]string, error) {
	var now time.Time
	if len(timestamp) > 0 && !timestamp[0].IsZero() {
		now = timestamp[0]
	} else {
		now = time.Now()
	}
	
	// Apply timezone conversion if configured
	if cfg != nil {
		tz := cfg.GetTimezone()
		if tz != "" {
			location, err := time.LoadLocation(tz)
			if err != nil {
				// Invalid timezone - return error
				return nil, fmt.Errorf("invalid timezone '%s': %w", tz, err)
			}
			// Convert to the specified timezone
			now = now.In(location)
		}
	}
	
	vars := make(map[string]string)

	// Default formats
	defaultDateFormat := "2006-01-02"
	defaultTimeFormat := "15:04:05"
	defaultDateTimeFormat := "2006-01-02 15:04:05"

	// If config is nil, use default behavior (all enabled with default formats)
	if cfg == nil {
		vars = map[string]string{
			"date":     now.Format(defaultDateFormat),
			"time":     now.Format(defaultTimeFormat),
			"datetime": now.Format(defaultDateTimeFormat),
		}
	} else {
		// Get date/time configuration from config
		dtVars := cfg.GetDateTimeVars()

		// Date variable
		// Enabled is nil = not specified, default to true; otherwise use the value
		if dtVars.Date.Enabled == nil || *dtVars.Date.Enabled {
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
		if dtVars.Time.Enabled == nil || *dtVars.Time.Enabled {
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
		if dtVars.DateTime.Enabled == nil || *dtVars.DateTime.Enabled {
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
	}

	// Add metadata variables if provided
	if metadata != nil {
		if metadata.User != "" {
			vars["user"] = metadata.User
		}
		if metadata.Host != "" {
			vars["host"] = metadata.Host
		}
		if metadata.Branch != "" {
			vars["branch"] = metadata.Branch
		}
		if metadata.Commit != "" {
			vars["commit"] = metadata.Commit
		}
	}

	return vars, nil
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

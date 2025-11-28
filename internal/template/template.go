package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

// GetDefaultVariables returns a map of default variable values
func GetDefaultVariables() map[string]string {
	now := time.Now()

	// Go map literal syntax
	return map[string]string{
		"date":     now.Format("2006-01-02"),          // Go's reference time
		"time":     now.Format("15:04:05"),            // HH:MM:SS
		"datetime": now.Format("2006-01-02 15:04:05"), // Date and time
	}
}


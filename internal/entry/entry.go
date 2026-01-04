package entry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/template"
	"github.com/sv4u/touchlog/internal/validation"
	"github.com/sv4u/touchlog/internal/xdg"
)

// Metadata represents metadata about an entry (for Phase 7)
// For now, this is a placeholder that can be nil
type Metadata struct {
	User string
	Host string
	Git  *GitContext
}

// GitContext represents git context information (for Phase 7)
type GitContext struct {
	Branch string
	Commit string
}

// Entry represents a log entry to be created
type Entry struct {
	Title    string
	Message  string
	Tags     []string
	Metadata *Metadata // Can be nil for Phase 4
	Date     time.Time
}

// CreateEntry creates a new log entry file with the given entry data
// It applies the template, generates the filename, and writes the file
// Returns the path to the created file
func CreateEntry(entry *Entry, cfg *config.Config, outputDir string, overwrite bool) (string, error) {
	// Validate output directory early (before any operations)
	if err := validation.ValidateOutputDir(outputDir); err != nil {
		return "", fmt.Errorf("invalid output directory: %w", err)
	}

	// Expand output directory (handle ~, environment variables, and relative paths)
	expandedDir, err := validation.ExpandPath(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to expand output directory: %w", err)
	}

	// Ensure output directory exists (validation already checked it can be created)
	if err := os.MkdirAll(expandedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get timezone from config
	var tz *time.Location
	if cfg != nil {
		tzStr := cfg.GetTimezone()
		if tzStr != "" {
			location, err := time.LoadLocation(tzStr)
			if err != nil {
				return "", fmt.Errorf("invalid timezone '%s': %w", tzStr, err)
			}
			tz = location
		}
	}
	if tz == nil {
		tz = time.Now().Location() // Use system timezone
	}

	// Generate base filename (without collision handling)
	baseFilename := FormatDate(entry.Date, tz) + "_" + GenerateSlug(entry.Title, entry.Message) + ".md"
	basePath := filepath.Join(expandedDir, baseFilename)

	// Check if base file exists and handle overwrite
	if !overwrite {
		// Check if base file exists
		if _, err := os.Stat(basePath); err == nil {
			return "", fmt.Errorf("file already exists: %s (use --overwrite to overwrite)", basePath)
		}
		// If base doesn't exist, generate filename with collision handling
		filename, err := GenerateFilename(expandedDir, entry.Title, entry.Message, entry.Date, tz)
		if err != nil {
			return "", fmt.Errorf("failed to generate filename: %w", err)
		}
		basePath = filename
	}

	// Use base path for overwrite, or generated path for new file
	filename := basePath

	// Get template name (use default if not specified)
	templateName := ""
	if cfg != nil {
		templateName = cfg.GetDefaultTemplate()
	}
	// If still empty, use "daily" as fallback
	if templateName == "" {
		templateName = "daily"
	}

	// Get templates directory
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		return "", fmt.Errorf("failed to get templates directory: %w", err)
	}

	// Resolve template
	templateContent, err := template.ResolveTemplate(templateName, cfg, templatesDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve template '%s': %w", templateName, err)
	}

	// Get default variables (date, time, datetime, custom vars, metadata)
	// Use entry.Date to ensure consistency between filename and template variables
	// Convert entry.Metadata to template.MetadataValues
	var metaValues *template.MetadataValues
	if entry.Metadata != nil {
		metaValues = &template.MetadataValues{
			User:   entry.Metadata.User,
			Host:   entry.Metadata.Host,
			Branch: "",
			Commit: "",
		}
		if entry.Metadata.Git != nil {
			metaValues.Branch = entry.Metadata.Git.Branch
			metaValues.Commit = entry.Metadata.Git.Commit
		}
	}
	defaultVars, err := template.GetDefaultVariablesWithMetadata(cfg, metaValues, entry.Date)
	if err != nil {
		return "", fmt.Errorf("failed to get default variables: %w", err)
	}

	// Build variables map for template processing
	vars := make(map[string]string)
	// Add default variables
	for k, v := range defaultVars {
		vars[k] = v
	}

	// Add entry-specific variables
	vars["title"] = entry.Title
	vars["message"] = entry.Message

	// Format tags as comma-separated string
	if len(entry.Tags) > 0 {
		vars["tags"] = strings.Join(entry.Tags, ", ")
	} else {
		vars["tags"] = "" // Empty string for empty tags
	}

	// Process template
	content := template.ProcessTemplate(templateContent, vars)

	// Write file atomically (write to temp file, then rename)
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Rename temp file to final filename (atomic operation)
	if err := os.Rename(tempFile, filename); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tempFile)
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	return filename, nil
}



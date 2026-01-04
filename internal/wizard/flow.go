package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/editor"
	"github.com/sv4u/touchlog/internal/entry"
	"github.com/sv4u/touchlog/internal/metadata"
	"github.com/sv4u/touchlog/internal/template"
	"github.com/sv4u/touchlog/internal/xdg"
)

// CreateTempFile creates a temporary file with the entry content
// Returns the path to the temporary file
func (w *Wizard) CreateTempFile() error {
	// Get template name (use selected template or default)
	templateName := w.templateName
	if templateName == "" {
		templateName = w.config.GetDefaultTemplate()
		if templateName == "" {
			templateName = "daily" // Fallback
		}
	}

	// Get templates directory
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	// Resolve template
	templateContent, err := template.ResolveTemplate(templateName, w.config, templatesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve template '%s': %w", templateName, err)
	}

	// Collect metadata if not already collected
	var metaValues *template.MetadataValues
	if w.metadata == nil {
		meta, err := metadata.CollectMetadata(w.config, w.includeGit, w.outputDir)
		if err != nil {
			// If metadata collection fails, continue without it (don't fail)
			w.metadata = nil
		} else {
			w.metadata = meta
		}
	}

	// Convert entry.Metadata to template.MetadataValues
	if w.metadata != nil {
		metaValues = &template.MetadataValues{
			User:   w.metadata.User,
			Host:   w.metadata.Host,
			Branch: "",
			Commit: "",
		}
		if w.metadata.Git != nil {
			metaValues.Branch = w.metadata.Git.Branch
			metaValues.Commit = w.metadata.Git.Commit
		}
	}

	// Get default variables (date, time, datetime, custom vars, metadata)
	defaultVars, err := template.GetDefaultVariablesWithMetadata(w.config, metaValues, w.timestamp)
	if err != nil {
		return fmt.Errorf("failed to get default variables: %w", err)
	}

	// Build variables map for template processing
	vars := make(map[string]string)
	// Add default variables
	for k, v := range defaultVars {
		vars[k] = v
	}

	// Add entry-specific variables
	vars["title"] = w.title
	vars["message"] = w.message

	// Format tags as comma-separated string
	if len(w.tags) > 0 {
		vars["tags"] = strings.Join(w.tags, ", ")
	} else {
		vars["tags"] = "" // Empty string for empty tags
	}

	// Process template
	content := template.ProcessTemplate(templateContent, vars)
	w.fileContent = content

	// Create temporary file in system temp directory
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "touchlog-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Write content to temporary file
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	w.tempFilePath = tmpPath
	return nil
}

// LaunchEditor launches the editor (external or internal) with the temp file
func (w *Wizard) LaunchEditor() error {
	if w.tempFilePath == "" {
		return fmt.Errorf("no temporary file to edit")
	}

	// Resolve editor using Phase 5 resolver
	resolver := editor.NewEditorResolver("", w.config, true) // fallbackToInternal = true
	editorInfo, err := resolver.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve editor: %w", err)
	}

	// Launch editor based on type
	if editorInfo.Type == editor.EditorTypeExternal {
		// Launch external editor
		if err := editor.LaunchEditor(editorInfo.Command, editorInfo.Args, w.tempFilePath); err != nil {
			return fmt.Errorf("failed to launch external editor: %w", err)
		}
	} else {
		// Launch internal editor
		// Read current content from temp file
		content, err := os.ReadFile(w.tempFilePath)
		if err != nil {
			return fmt.Errorf("failed to read temp file: %w", err)
		}

		internalEditor, err := editor.NewInternalEditor(w.tempFilePath, string(content), w.config)
		if err != nil {
			return fmt.Errorf("failed to create internal editor: %w", err)
		}

		if err := internalEditor.Run(); err != nil {
			return fmt.Errorf("failed to run internal editor: %w", err)
		}

		// Update file content from edited file
		updatedContent := internalEditor.GetContent()
		w.fileContent = updatedContent
	}

	// Re-read file content after editor exits (in case external editor modified it)
	content, err := os.ReadFile(w.tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file after editor: %w", err)
	}
	w.fileContent = string(content)

	return nil
}

// Confirm moves the temporary file to the final location
func (w *Wizard) Confirm() error {
	if w.tempFilePath == "" {
		return fmt.Errorf("no temporary file to confirm")
	}

	if w.outputDir == "" {
		return fmt.Errorf("output directory not set")
	}

	// Expand output directory (handle ~ and relative paths)
	expandedDir, err := expandPath(w.outputDir)
	if err != nil {
		return fmt.Errorf("failed to expand output directory: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(expandedDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get timezone from config
	var tz *time.Location
	if w.config != nil {
		tzStr := w.config.GetTimezone()
		if tzStr != "" {
			location, err := time.LoadLocation(tzStr)
			if err != nil {
				return fmt.Errorf("invalid timezone '%s': %w", tzStr, err)
			}
			tz = location
		}
	}
	if tz == nil {
		tz = time.Now().Location() // Use system timezone
	}

	// Generate final filename
	finalPath, err := entry.GenerateFilename(expandedDir, w.title, w.message, w.timestamp, tz)
	if err != nil {
		return fmt.Errorf("failed to generate filename: %w", err)
	}

	// Read current content from temp file (in case it was modified)
	content, err := os.ReadFile(w.tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to read temp file: %w", err)
	}

	// Write to final location atomically
	tempFinalFile := finalPath + ".tmp"
	if err := os.WriteFile(tempFinalFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write final file: %w", err)
	}

	// Rename temp file to final filename (atomic operation)
	if err := os.Rename(tempFinalFile, finalPath); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tempFinalFile)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Clean up temporary file
	_ = os.Remove(w.tempFilePath)

	w.finalFilePath = finalPath
	w.fileContent = string(content)
	return nil
}

// Cancel deletes the temporary file
func (w *Wizard) Cancel() error {
	if w.tempFilePath == "" {
		return nil // Nothing to cancel
	}

	// Delete temporary file
	if err := os.Remove(w.tempFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete temporary file: %w", err)
	}

	w.tempFilePath = ""
	return nil
}

// ValidateOutputDir validates the output directory
func (w *Wizard) ValidateOutputDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Expand path (handle ~ and relative paths)
	expandedDir, err := expandPath(dir)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if directory exists or can be created
	info, err := os.Stat(expandedDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist - check if parent is writable
			parentDir := filepath.Dir(expandedDir)
			parentInfo, err := os.Stat(parentDir)
			if err != nil {
				return fmt.Errorf("parent directory does not exist: %w", err)
			}
			if !parentInfo.IsDir() {
				return fmt.Errorf("parent path is not a directory")
			}
			// Check if we can create the directory (check parent write permissions)
			// We can't actually test write permissions without trying, so we'll just validate the path
			return nil // Path is valid, directory will be created later
		}
		return fmt.Errorf("failed to check directory: %w", err)
	}

	// Directory exists - check if it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory")
	}

	// Check write permissions (try to create a test file)
	testFile := filepath.Join(expandedDir, ".touchlog-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	// Clean up test file
	_ = os.Remove(testFile)

	return nil
}

// ParseTags parses a comma-separated tag string into a slice of tags
func ParseTags(tagString string) []string {
	if tagString == "" {
		return nil
	}

	// Split by comma and trim whitespace
	parts := strings.Split(tagString, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			tags = append(tags, trimmed)
		}
	}

	return tags
}

// GetAvailableTemplates returns a list of all available templates (inline + file-based)
// Returns a slice of template names, with inline templates first
func GetAvailableTemplates(cfg *config.Config) []string {
	templates := make([]string, 0)

	// Add inline templates first (they take precedence)
	inlineTemplates := cfg.GetInlineTemplates()
	for name := range inlineTemplates {
		templates = append(templates, name)
	}

	// Add file-based templates
	fileTemplates := cfg.GetTemplates()
	for _, tmpl := range fileTemplates {
		// Extract base name from filename (without extension)
		baseName := strings.TrimSuffix(tmpl.File, filepath.Ext(tmpl.File))
		// Only add if not already in list (inline templates take precedence)
		found := false
		for _, existing := range templates {
			if existing == baseName {
				found = true
				break
			}
		}
		if !found {
			templates = append(templates, baseName)
		}
	}

	return templates
}

// expandPath expands a path, handling ~ and relative paths
func expandPath(path string) (string, error) {
	// Handle ~ expansion
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if path == "~" {
			return homeDir, nil
		}
		// Validate that paths starting with ~ must be ~/ (not ~something)
		if !strings.HasPrefix(path, "~/") {
			return "", fmt.Errorf("invalid path: paths starting with ~ must be followed by / (e.g., ~/path), got: %s", path)
		}
		// Skip the leading ~/
		remaining := path[2:]
		path = filepath.Join(homeDir, remaining)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}

	return absPath, nil
}


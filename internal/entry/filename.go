package entry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateFilename generates a filename in the format YYYY-MM-DD_slug.md
// It handles collisions by appending numeric suffixes (_1, _2, etc.)
// Returns the full path to the file
func GenerateFilename(outputDir string, title string, message string, timestamp time.Time, timezone *time.Location) (string, error) {
	// Format date in the specified timezone
	dateStr := FormatDate(timestamp, timezone)

	// Generate slug
	slug := GenerateSlug(title, message)

	// Construct base filename: YYYY-MM-DD_slug.md
	baseFilename := fmt.Sprintf("%s_%s.md", dateStr, slug)
	basePath := filepath.Join(outputDir, baseFilename)

	// Find available filename (handles collisions)
	finalPath, err := FindAvailableFilename(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to find available filename: %w", err)
	}

	return finalPath, nil
}

// FormatDate formats a time.Time in the specified timezone as YYYY-MM-DD
func FormatDate(t time.Time, tz *time.Location) string {
	if tz != nil {
		t = t.In(tz)
	}
	return t.Format("2006-01-02")
}

// FindAvailableFilename finds an available filename by appending numeric suffixes
// if the base path already exists. Returns the first available path.
// Example: if "2025-01-01_test.md" exists, returns "2025-01-01_test_1.md"
func FindAvailableFilename(basePath string) (string, error) {
	// If base path doesn't exist, use it
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath, nil
	}

	// Extract directory, base name, and extension
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(basePath)
	baseName := strings.TrimSuffix(filepath.Base(basePath), ext)

	// Try numeric suffixes starting from 1
	for i := 1; i < 10000; i++ {
		candidateName := fmt.Sprintf("%s_%d%s", baseName, i, ext)
		candidatePath := filepath.Join(dir, candidateName)

		if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
			return candidatePath, nil
		}
	}

	// If we've exhausted all reasonable suffixes, return an error
	return "", fmt.Errorf("unable to find available filename after 9999 attempts")
}

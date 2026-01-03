package entry

import (
	"regexp"
	"strings"
)

const (
	// MaxSlugLength is the maximum length for a slug
	MaxSlugLength = 50
	// DefaultSlug is used when slug generation results in an empty string
	DefaultSlug = "untitled"
)

// GenerateSlug generates a URL-friendly slug from title and message
// Algorithm:
// 1. Prefer title, fallback to first line of message
// 2. Convert to lowercase
// 3. Replace spaces and special chars with hyphens
// 4. Remove consecutive hyphens
// 5. Trim hyphens from start/end
// 6. Limit to 50 characters (truncate if needed)
// 7. If resulting slug is empty, use "untitled"
func GenerateSlug(title string, message string) string {
	// Step 1: Prefer title, fallback to first line of message
	source := title
	if source == "" {
		lines := strings.Split(message, "\n")
		if len(lines) > 0 {
			source = strings.TrimSpace(lines[0])
		}
	}

	// If source is still empty, return default
	if source == "" {
		return DefaultSlug
	}

	// Step 2: Convert to lowercase
	source = strings.ToLower(source)

	// Step 3: Replace spaces and special chars with hyphens
	// Replace spaces with hyphens
	source = strings.ReplaceAll(source, " ", "-")

	// Replace special characters with hyphens
	// Allow alphanumeric and hyphens only
	re := regexp.MustCompile(`[^a-z0-9-]`)
	source = re.ReplaceAllString(source, "-")

	// Step 4: Remove consecutive hyphens
	reConsecutive := regexp.MustCompile(`-+`)
	source = reConsecutive.ReplaceAllString(source, "-")

	// Step 5: Trim hyphens from start/end
	source = strings.Trim(source, "-")

	// Step 6: Limit to 50 characters (truncate if needed)
	if len(source) > MaxSlugLength {
		source = source[:MaxSlugLength]
		// Trim any trailing hyphen after truncation
		source = strings.TrimSuffix(source, "-")
	}

	// Step 7: If resulting slug is empty, use "untitled"
	if source == "" {
		return DefaultSlug
	}

	return source
}


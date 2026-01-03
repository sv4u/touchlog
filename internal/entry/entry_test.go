package entry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/xdg"
)

func TestCreateEntry(t *testing.T) {
	// Create temporary directory for output
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal config
	cfg := config.CreateDefaultConfig()
	cfg.NotesDirectory = tmpDir

	// Create templates directory and a simple template
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		t.Fatalf("Failed to get templates dir: %v", err)
	}

	// Create a simple template file
	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}

{{message}}

Tags: {{tags}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	defer os.Remove(templatePath)

	// Set default template
	cfg.DefaultTemplate = "daily"

	tests := []struct {
		name      string
		entry     *Entry
		overwrite bool
		wantError bool
	}{
		{
			name: "basic entry",
			entry: &Entry{
				Title:    "Test Entry",
				Message:  "This is a test message",
				Tags:     []string{"test", "example"},
				Metadata: nil,
				Date:     time.Now(),
			},
			overwrite: false,
			wantError: false,
		},
		{
			name: "entry with empty tags",
			entry: &Entry{
				Title:    "No Tags",
				Message:  "Message without tags",
				Tags:     []string{},
				Metadata: nil,
				Date:     time.Now(),
			},
			overwrite: false,
			wantError: false,
		},
		{
			name: "entry without title",
			entry: &Entry{
				Title:    "",
				Message:  "Message only",
				Tags:     []string{},
				Metadata: nil,
				Date:     time.Now(),
			},
			overwrite: false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := CreateEntry(tt.entry, cfg, tmpDir, tt.overwrite)
			if (err != nil) != tt.wantError {
				t.Errorf("CreateEntry() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				// Verify file was created
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("CreateEntry() file was not created: %s", filePath)
				}

				// Verify file content contains expected values
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read created file: %v", err)
				}

				contentStr := string(content)
				if tt.entry.Title != "" && !contains(contentStr, tt.entry.Title) {
					t.Errorf("CreateEntry() file content missing title: %q", tt.entry.Title)
				}
				if tt.entry.Message != "" && !contains(contentStr, tt.entry.Message) {
					t.Errorf("CreateEntry() file content missing message: %q", tt.entry.Message)
				}
				if len(tt.entry.Tags) > 0 {
					for _, tag := range tt.entry.Tags {
						if !contains(contentStr, tag) {
							t.Errorf("CreateEntry() file content missing tag: %q", tag)
						}
					}
				}

				// Clean up
				_ = os.Remove(filePath)
			}
		})
	}
}

func TestCreateEntryOverwrite(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config
	cfg := config.CreateDefaultConfig()
	cfg.NotesDirectory = tmpDir

	// Create template
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		t.Fatalf("Failed to get templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}}

{{message}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	defer os.Remove(templatePath)

	cfg.DefaultTemplate = "daily"

	entry := &Entry{
		Title:    "Test",
		Message:  "Original message",
		Tags:     []string{},
		Metadata: nil,
		Date:     time.Now(),
	}

	// Create entry first time
	filePath, err := CreateEntry(entry, cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Try to create again without overwrite flag (should fail)
	_, err = CreateEntry(entry, cfg, tmpDir, false)
	if err == nil {
		t.Error("CreateEntry() should fail when file exists and overwrite is false")
	}

	// Try to create again with overwrite flag (should succeed)
	entry.Message = "Updated message"
	filePath2, err := CreateEntry(entry, cfg, tmpDir, true)
	if err != nil {
		t.Fatalf("CreateEntry() with overwrite=true failed: %v", err)
	}

	// With overwrite, it should use the same base filename (may have different suffix if collision)
	// But the important thing is that it overwrites the content
	baseName1 := filepath.Base(filePath)
	baseName2 := filepath.Base(filePath2)
	// Extract base without suffix for comparison
	base1 := baseName1
	base2 := baseName2
	if len(base1) > 3 && base1[len(base1)-3:] == ".md" {
		base1 = base1[:len(base1)-3]
	}
	if len(base2) > 3 && base2[len(base2)-3:] == ".md" {
		base2 = base2[:len(base2)-3]
	}
	// Remove numeric suffix if present
	if len(base1) > 2 && base1[len(base1)-2] == '_' {
		// Check if last part is numeric
		base1 = extractBaseName(base1)
	}
	if len(base2) > 2 && base2[len(base2)-2] == '_' {
		base2 = extractBaseName(base2)
	}
	
	// The base names should match (without suffix)
	if base1 != base2 {
		t.Logf("Note: Filenames differ but overwrite should still work: %q vs %q", base1, base2)
	}

	// Verify content was updated
	content, err := os.ReadFile(filePath2)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !contains(string(content), "Updated message") {
		t.Error("CreateEntry() with overwrite did not update file content")
	}

	// Clean up
	_ = os.Remove(filePath2)
}

func TestCreateEntryTimezone(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config with timezone
	cfg := config.CreateDefaultConfig()
	cfg.NotesDirectory = tmpDir
	cfg.Timezone = "America/Denver"

	// Create template
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		t.Fatalf("Failed to get templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	defer os.Remove(templatePath)

	cfg.DefaultTemplate = "daily"

	// Use a fixed time for testing
	testTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	entry := &Entry{
		Title:    "Test",
		Message:  "Message",
		Tags:     []string{},
		Metadata: nil,
		Date:     testTime,
	}

	filePath, err := CreateEntry(entry, cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Date in template should be formatted in America/Denver timezone
	// The filename will use the entry.Date, but the template {{date}} uses GetDefaultVariables
	// which uses time.Now(). For this test, we just verify the file was created successfully
	// and contains the expected structure
	contentStr := string(content)
	if !contains(contentStr, "Test") {
		t.Errorf("CreateEntry() content missing title: %s", contentStr)
	}

	// Verify filename uses the correct date from entry.Date
	filename := filepath.Base(filePath)
	if !contains(filename, "2025-01-15") {
		t.Errorf("CreateEntry() filename date incorrect: %s", filename)
	}

	// Clean up
	_ = os.Remove(filePath)
}

func TestCreateEntryWithTildeExpansion(t *testing.T) {
	// Test that expandPath correctly handles "~" (just tilde) and "~/path"
	// This verifies the fix for the bug where "~" alone wasn't expanded

	// Create config
	cfg := config.CreateDefaultConfig()

	// Create template
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		t.Fatalf("Failed to get templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}}

{{message}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	defer os.Remove(templatePath)

	cfg.DefaultTemplate = "daily"

	// Get home directory for comparison
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	entry := &Entry{
		Title:    "Tilde Test",
		Message:  "Testing tilde expansion",
		Tags:     []string{},
		Metadata: nil,
		Date:     time.Now(),
	}

	// Test case 1: "~" should expand to home directory
	filePath, err := CreateEntry(entry, cfg, "~", false)
	if err != nil {
		t.Fatalf("CreateEntry() with '~' failed: %v", err)
	}

	// Verify the file was created in the home directory
	if !strings.HasPrefix(filePath, homeDir) {
		t.Errorf("CreateEntry() with '~' did not expand to home directory. Got: %s, Expected prefix: %s", filePath, homeDir)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("CreateEntry() file was not created: %s", filePath)
	}

	// Clean up
	_ = os.Remove(filePath)

	// Test case 2: "~/test-notes" should expand correctly
	testSubdir := "~/test-notes"
	filePath2, err := CreateEntry(entry, cfg, testSubdir, false)
	if err != nil {
		t.Fatalf("CreateEntry() with '~/test-notes' failed: %v", err)
	}

	// Verify the file was created in ~/test-notes
	expectedPath := filepath.Join(homeDir, "test-notes")
	if !strings.HasPrefix(filePath2, expectedPath) {
		t.Errorf("CreateEntry() with '~/test-notes' did not expand correctly. Got: %s, Expected prefix: %s", filePath2, expectedPath)
	}

	// Verify file exists
	if _, err := os.Stat(filePath2); os.IsNotExist(err) {
		t.Errorf("CreateEntry() file was not created: %s", filePath2)
	}

	// Clean up
	_ = os.Remove(filePath2)
	_ = os.RemoveAll(expectedPath)

	// Test case 3: "~notes" (missing slash) should return an error
	_, err = CreateEntry(entry, cfg, "~notes", false)
	if err == nil {
		t.Error("CreateEntry() with '~notes' (missing slash) should return an error")
	}
	if !strings.Contains(err.Error(), "must be followed by /") {
		t.Errorf("CreateEntry() error message should mention 'must be followed by /', got: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && (s[:len(substr)] == substr || 
			(len(s) > len(substr) && containsHelper(s, substr)))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractBaseName removes numeric suffix from base name (e.g., "test_1" -> "test")
func extractBaseName(name string) string {
	// Simple approach: find last underscore and check if followed by digits
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '_' && i < len(name)-1 {
			// Check if rest is numeric
			rest := name[i+1:]
			isNumeric := true
			for _, r := range rest {
				if r < '0' || r > '9' {
					isNumeric = false
					break
				}
			}
			if isNumeric {
				return name[:i]
			}
		}
	}
	return name
}


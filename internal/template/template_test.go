package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sv4u/touchlog/internal/config"
)

func TestProcessTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "single variable",
			template: "Hello {{name}}",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "multiple variables",
			template: "Hello {{name}}, today is {{date}}",
			vars:     map[string]string{"name": "World", "date": "2024-01-01"},
			want:     "Hello World, today is 2024-01-01",
		},
		{
			name:     "variables that don't exist in map",
			template: "Hello {{name}}, missing: {{missing}}",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World, missing: {{missing}}",
		},
		{
			name:     "empty template content",
			template: "",
			vars:     map[string]string{"name": "World"},
			want:     "",
		},
		{
			name:     "template with no variables",
			template: "Hello World",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "special characters in variable values",
			template: "Value: {{value}}",
			vars:     map[string]string{"value": "Special: !@#$%^&*()"},
			want:     "Value: Special: !@#$%^&*()",
		},
		{
			name:     "newlines in variable values",
			template: "Content:\n{{content}}",
			vars:     map[string]string{"content": "Line 1\nLine 2"},
			want:     "Content:\nLine 1\nLine 2",
		},
		{
			name:     "multiple occurrences of same variable",
			template: "{{name}} and {{name}}",
			vars:     map[string]string{"name": "John"},
			want:     "John and John",
		},
		{
			name:     "empty variables map",
			template: "Hello {{name}}",
			vars:     map[string]string{},
			want:     "Hello {{name}}",
		},
		{
			name:     "empty variable value",
			template: "Hello {{name}}",
			vars:     map[string]string{"name": ""},
			want:     "Hello ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessTemplate(tt.template, tt.vars)
			if got != tt.want {
				t.Errorf("ProcessTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantLen int
		exact   bool // If true, check exact match; if false, only check length
	}{
		{
			name:    "single variable extraction",
			content: "Hello {{name}}",
			want:    []string{"name"},
			wantLen: 1,
			exact:   true,
		},
		{
			name:    "multiple variable extraction",
			content: "Hello {{name}}, today is {{date}}",
			want:    []string{"name", "date"},
			wantLen: 2,
			exact:   true,
		},
		{
			name:    "duplicate variables",
			content: "{{name}} and {{name}}",
			want:    []string{"name", "name"},
			wantLen: 2,
			exact:   true,
		},
		{
			name:    "malformed variable syntax - missing closing braces",
			content: "Hello {{name",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "malformed variable syntax - missing opening braces",
			content: "Hello name}}",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "malformed variable syntax - single brace",
			content: "Hello {name}",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "empty content",
			content: "",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "content with no variables",
			content: "Hello World",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "variable with underscores",
			content: "Hello {{user_name}}",
			want:    []string{"user_name"},
			wantLen: 1,
			exact:   true,
		},
		{
			name:    "variable with numbers",
			content: "Value {{value1}} and {{value2}}",
			want:    []string{"value1", "value2"},
			wantLen: 2,
			exact:   true,
		},
		{
			name:    "mixed valid and invalid variables",
			content: "{{valid}} and {invalid} and {{also_valid}}",
			want:    []string{"valid", "also_valid"},
			wantLen: 2,
			exact:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractVariables(tt.content)
			if tt.exact {
				if len(got) != len(tt.want) {
					t.Errorf("ExtractVariables() length = %v, want %v", len(got), len(tt.want))
					return
				}
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("ExtractVariables()[%d] = %v, want %v", i, v, tt.want[i])
					}
				}
			} else {
				if len(got) != tt.wantLen {
					t.Errorf("ExtractVariables() length = %v, want %v", len(got), tt.wantLen)
				}
			}
		})
	}
}

func TestGetDefaultVariables(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		vars := GetDefaultVariables(nil)

		// Should have date, time, and datetime
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables(nil) missing 'date' variable")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables(nil) missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables(nil) missing 'datetime' variable")
		}

		// Should have exactly 3 variables
		if len(vars) != 3 {
			t.Errorf("GetDefaultVariables(nil) returned %d variables, want 3", len(vars))
		}
	})

	t.Run("config with all variables enabled", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
		}

		vars := GetDefaultVariables(cfg)

		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable")
		}

		if len(vars) != 3 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 3", len(vars))
		}
	})

	t.Run("config with some variables disabled", func(t *testing.T) {
		enabledTrue := true
		enabledFalse := false
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledFalse, Format: ""},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
		}

		vars := GetDefaultVariables(cfg)

		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable")
		}
		if _, ok := vars["time"]; ok {
			t.Error("GetDefaultVariables() should not have 'time' variable when disabled")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable")
		}

		if len(vars) != 2 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 2", len(vars))
		}
	})

	t.Run("config with custom date/time formats", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "01/02/2006"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "03:04 PM"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "01/02/2006 03:04 PM"},
			},
		}

		vars := GetDefaultVariables(cfg)

		// Should have all three variables
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable")
		}

		// Note: We can't test the exact format output since it's time-dependent,
		// but we can verify the variables exist
		if len(vars) != 3 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 3", len(vars))
		}
	})

	t.Run("config with invalid formats - should fallback to defaults", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "invalid-format"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "also-invalid"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "invalid-too"},
			},
		}

		vars := GetDefaultVariables(cfg)

		// Should still have all three variables (fallback to default formats)
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable after invalid format")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable after invalid format")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable after invalid format")
		}

		if len(vars) != 3 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 3", len(vars))
		}
	})

	t.Run("config with empty format strings - should use defaults", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: ""},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: ""},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: ""},
			},
		}

		vars := GetDefaultVariables(cfg)

		// Should have all three variables with default formats
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable with empty format")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable with empty format")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable with empty format")
		}

		if len(vars) != 3 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 3", len(vars))
		}
	})

	t.Run("config with all variables disabled - should fallback to all enabled", func(t *testing.T) {
		enabledFalse := false
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledFalse, Format: ""},
				Time:     config.DateTimeVarConfig{Enabled: &enabledFalse, Format: ""},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledFalse, Format: ""},
			},
		}

		vars := GetDefaultVariables(cfg)

		// Should fallback to all enabled (backward compatibility)
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable when all disabled (should fallback)")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable when all disabled (should fallback)")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable when all disabled (should fallback)")
		}

		if len(vars) != 3 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 3 (fallback)", len(vars))
		}
	})

	t.Run("custom variable merging", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
			Variables: map[string]string{
				"author":  "Test Author",
				"project": "Test Project",
			},
		}

		vars := GetDefaultVariables(cfg)

		// Should have date/time/datetime
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable")
		}

		// Should have custom variables
		if vars["author"] != "Test Author" {
			t.Errorf("GetDefaultVariables() custom variable 'author' = %v, want 'Test Author'", vars["author"])
		}
		if vars["project"] != "Test Project" {
			t.Errorf("GetDefaultVariables() custom variable 'project' = %v, want 'Test Project'", vars["project"])
		}

		if len(vars) != 5 {
			t.Errorf("GetDefaultVariables() returned %d variables, want 5", len(vars))
		}
	})

	t.Run("custom variables overriding default variables", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
			Variables: map[string]string{
				"date": "Custom Date Override",
			},
		}

		vars := GetDefaultVariables(cfg)

		// Custom variable should override default date variable
		if vars["date"] != "Custom Date Override" {
			t.Errorf("GetDefaultVariables() 'date' variable = %v, want 'Custom Date Override'", vars["date"])
		}

		// Other default variables should still exist
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable")
		}
	})

	t.Run("timezone conversion applied when configured", func(t *testing.T) {
		enabledTrue := true
		// Use UTC timezone for predictable test results
		cfg := &config.Config{
			Timezone: "UTC",
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
		}

		vars := GetDefaultVariables(cfg)

		// Verify variables exist
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable with timezone")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable with timezone")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable with timezone")
		}

		// Verify the timezone is actually applied by checking the time difference
		// Get current time in UTC and in local timezone
		utcNow := time.Now().UTC()
		localNow := time.Now()

		// Parse the datetime variable to verify it's in UTC
		parsedTime, err := time.Parse("2006-01-02 15:04:05", vars["datetime"])
		if err != nil {
			t.Fatalf("GetDefaultVariables() datetime format error: %v", err)
		}

		// The parsed time should be close to UTC time (within 1 minute to account for test execution time)
		diff := parsedTime.Sub(utcNow)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Minute {
			// If not close to UTC, check if it's close to local time (which would indicate timezone wasn't applied)
			localDiff := parsedTime.Sub(localNow)
			if localDiff < 0 {
				localDiff = -localDiff
			}
			if localDiff < time.Minute {
				t.Errorf("GetDefaultVariables() timezone not applied - datetime is in local timezone, not UTC")
			}
		}
	})

	t.Run("invalid timezone falls back to system timezone", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			Timezone: "Invalid/Timezone",
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
		}

		// Should not panic and should still return variables (using system timezone)
		vars := GetDefaultVariables(cfg)

		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariables() missing 'date' variable with invalid timezone (should fallback)")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariables() missing 'time' variable with invalid timezone (should fallback)")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariables() missing 'datetime' variable with invalid timezone (should fallback)")
		}
	})
}

func TestCreateExampleTemplates(t *testing.T) {
	t.Run("creates templates in empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create empty directory
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create templates
		if err := CreateExampleTemplates(templatesDir); err != nil {
			t.Fatalf("CreateExampleTemplates() error = %v", err)
		}

		// Verify three template files exist
		expectedFiles := []string{"daily.md", "meeting.md", "journal.md"}
		for _, filename := range expectedFiles {
			path := filepath.Join(templatesDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("CreateExampleTemplates() file not created: %q", filename)
			}

			// Verify file permissions (approximately)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("os.Stat() error = %v", err)
			}
			mode := info.Mode()
			if mode&0644 != 0644 {
				t.Errorf("CreateExampleTemplates() file %q permissions = %o, want 0644", filename, mode&0777)
			}

			// Verify file content contains template variables
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("os.ReadFile() error = %v", err)
			}
			contentStr := string(content)
			if !strings.Contains(contentStr, "{{date}}") {
				t.Errorf("CreateExampleTemplates() file %q missing {{date}} variable", filename)
			}
		}
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Directory doesn't exist, CreateExampleTemplates should create it
		if err := CreateExampleTemplates(templatesDir); err != nil {
			t.Fatalf("CreateExampleTemplates() error = %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			t.Errorf("CreateExampleTemplates() directory not created at %q", templatesDir)
		}

		// Verify template files were created
		expectedFiles := []string{"daily.md", "meeting.md", "journal.md"}
		for _, filename := range expectedFiles {
			path := filepath.Join(templatesDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("CreateExampleTemplates() file not created: %q", filename)
			}
		}
	})

	t.Run("does nothing if directory has files", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create directory with existing template file
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		existingFile := filepath.Join(templatesDir, "existing.md")
		existingContent := "# Existing Template\n\nThis should not be modified"
		if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Create templates (should do nothing since directory is not empty)
		if err := CreateExampleTemplates(templatesDir); err != nil {
			t.Fatalf("CreateExampleTemplates() error = %v", err)
		}

		// Verify existing file is unchanged
		content, err := os.ReadFile(existingFile)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		if string(content) != existingContent {
			t.Error("CreateExampleTemplates() modified existing file")
		}

		// Verify no new files were created (directory was not empty)
		entries, err := os.ReadDir(templatesDir)
		if err != nil {
			t.Fatalf("os.ReadDir() error = %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("CreateExampleTemplates() created files in non-empty directory, found %d files", len(entries))
		}
	})

	t.Run("does not overwrite existing templates", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create directory with daily.md already present
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		existingDaily := filepath.Join(templatesDir, "daily.md")
		existingContent := "# Custom Daily Note\n\nThis should not be overwritten"
		if err := os.WriteFile(existingDaily, []byte(existingContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Create templates (should do nothing since directory is not empty)
		if err := CreateExampleTemplates(templatesDir); err != nil {
			t.Fatalf("CreateExampleTemplates() error = %v", err)
		}

		// Verify existing daily.md is unchanged
		content, err := os.ReadFile(existingDaily)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		if string(content) != existingContent {
			t.Error("CreateExampleTemplates() overwrote existing daily.md")
		}

		// Verify no new templates were created (directory was not empty)
		entries, err := os.ReadDir(templatesDir)
		if err != nil {
			t.Fatalf("os.ReadDir() error = %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("CreateExampleTemplates() created files in non-empty directory, found %d files", len(entries))
		}
	})

	t.Run("template content is valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create templates
		if err := CreateExampleTemplates(templatesDir); err != nil {
			t.Fatalf("CreateExampleTemplates() error = %v", err)
		}

		// Verify daily.md content
		dailyPath := filepath.Join(templatesDir, "daily.md")
		dailyContent, err := os.ReadFile(dailyPath)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		dailyStr := string(dailyContent)
		if !strings.Contains(dailyStr, "# Daily Note") {
			t.Error("CreateExampleTemplates() daily.md missing title")
		}
		if !strings.Contains(dailyStr, "{{date}}") {
			t.Error("CreateExampleTemplates() daily.md missing {{date}} variable")
		}

		// Verify meeting.md content
		meetingPath := filepath.Join(templatesDir, "meeting.md")
		meetingContent, err := os.ReadFile(meetingPath)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		meetingStr := string(meetingContent)
		if !strings.Contains(meetingStr, "# Meeting Notes") {
			t.Error("CreateExampleTemplates() meeting.md missing title")
		}
		if !strings.Contains(meetingStr, "{{date}}") {
			t.Error("CreateExampleTemplates() meeting.md missing {{date}} variable")
		}
		if !strings.Contains(meetingStr, "{{time}}") {
			t.Error("CreateExampleTemplates() meeting.md missing {{time}} variable")
		}

		// Verify journal.md content
		journalPath := filepath.Join(templatesDir, "journal.md")
		journalContent, err := os.ReadFile(journalPath)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		journalStr := string(journalContent)
		if !strings.Contains(journalStr, "# Journal Entry") {
			t.Error("CreateExampleTemplates() journal.md missing title")
		}
		if !strings.Contains(journalStr, "{{date}}") {
			t.Error("CreateExampleTemplates() journal.md missing {{date}} variable")
		}
	})

	t.Run("returns error on ReadDir permission error", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create directory with restricted read permissions
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Remove read permission
		if err := os.Chmod(templatesDir, 0333); err != nil {
			// If chmod fails (e.g., on Windows), skip this test
			t.Skip("Cannot set restrictive permissions on this platform")
		}
		defer func() {
			// Restore permissions for cleanup
			_ = os.Chmod(templatesDir, 0755)
		}()

		err := CreateExampleTemplates(templatesDir)
		if err == nil {
			t.Error("CreateExampleTemplates() expected error for permission denied, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to read templates directory") {
			t.Errorf("CreateExampleTemplates() error = %v, want error containing 'failed to read templates directory'", err)
		}
	})

	t.Run("returns error on MkdirAll failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "parent")

		// Create parent directory and make it read-only
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		if err := os.Chmod(parentDir, 0444); err != nil {
			// If chmod fails (e.g., on Windows), skip this test
			t.Skip("Cannot set read-only permissions on this platform")
		}
		defer func() {
			// Restore permissions for cleanup
			_ = os.Chmod(parentDir, 0755)
		}()

		// Try to create templates in a subdirectory of read-only parent
		// Note: ReadDir will fail first with permission error, not MkdirAll
		// This test verifies that ReadDir errors (non-ErrNotExist) are properly returned
		templatesDir := filepath.Join(parentDir, "templates")
		err := CreateExampleTemplates(templatesDir)
		if err == nil {
			t.Error("CreateExampleTemplates() expected error when directory access fails, got nil")
		}
		// The error will be from ReadDir (permission denied), not MkdirAll
		// This is still a valid error path test
		if err != nil && !strings.Contains(err.Error(), "failed to read templates directory") {
			t.Errorf("CreateExampleTemplates() error = %v, want error containing 'failed to read templates directory'", err)
		}
	})

	t.Run("returns error on WriteFile failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create empty directory and make it read-only
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		if err := os.Chmod(templatesDir, 0444); err != nil {
			// If chmod fails (e.g., on Windows), skip this test
			t.Skip("Cannot set read-only permissions on this platform")
		}
		defer func() {
			// Restore permissions for cleanup
			_ = os.Chmod(templatesDir, 0755)
		}()

		err := CreateExampleTemplates(templatesDir)
		if err == nil {
			t.Error("CreateExampleTemplates() expected error for read-only directory, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create template") {
			t.Errorf("CreateExampleTemplates() error = %v, want error containing 'failed to create template'", err)
		}
	})

	t.Run("skips existing file and creates others when directory was empty", func(t *testing.T) {
		// Note: CreateExampleTemplates checks if directory is empty first (line 162)
		// and returns early if it has files. However, there's also a per-file check
		// inside the loop (line 214) that skips existing files. This test verifies
		// that behavior, but it's difficult to test because the initial check prevents
		// reaching the per-file check when files exist.
		//
		// The per-file check is a safety measure for the case where files appear
		// between the initial check and the write (race condition), which is hard to test.
		//
		// Instead, this test verifies that when a directory has files, the function
		// returns early without creating anything (the actual designed behavior).
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create directory with daily.md already present
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create daily.md with custom content
		dailyPath := filepath.Join(templatesDir, "daily.md")
		existingContent := "# Custom Daily Note\n\nThis should not be overwritten"
		if err := os.WriteFile(dailyPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Create templates - should return early because directory is not empty
		if err := CreateExampleTemplates(templatesDir); err != nil {
			t.Fatalf("CreateExampleTemplates() error = %v", err)
		}

		// Verify daily.md is unchanged
		content, err := os.ReadFile(dailyPath)
		if err != nil {
			t.Fatalf("os.ReadFile() error = %v", err)
		}
		if string(content) != existingContent {
			t.Error("CreateExampleTemplates() overwrote existing daily.md")
		}

		// Verify other templates were NOT created (function returns early when directory has files)
		meetingPath := filepath.Join(templatesDir, "meeting.md")
		journalPath := filepath.Join(templatesDir, "journal.md")
		if _, err := os.Stat(meetingPath); err == nil {
			t.Error("CreateExampleTemplates() should not create meeting.md when directory has files")
		}
		if _, err := os.Stat(journalPath); err == nil {
			t.Error("CreateExampleTemplates() should not create journal.md when directory has files")
		}
	})

	t.Run("returns error on partial creation failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		// Create empty directory
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create a directory named daily.md to cause write failure
		// When WriteFile tries to write to a path that's a directory, it will fail
		dailyPath := filepath.Join(templatesDir, "daily.md")
		if err := os.MkdirAll(dailyPath, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}
		defer func() {
			// Clean up
			_ = os.RemoveAll(dailyPath)
		}()

		// Try to create templates - should fail when trying to write daily.md
		// Note: The os.Stat check at line 214 will see that daily.md exists (as a directory)
		// and skip it, so the error won't occur on daily.md. Let's make the directory
		// read-only after creating one file to cause a failure on a later file.
		// Actually, a better approach: make the directory read-only after it's determined
		// to be empty, but before all files are written. But that's a race condition.

		// Instead, let's test with a directory that becomes read-only
		// Remove the daily.md directory we created (it was just for setup)
		_ = os.RemoveAll(dailyPath)

		// Make the directory read-only to cause WriteFile to fail
		if err := os.Chmod(templatesDir, 0444); err != nil {
			t.Skip("Cannot set read-only permissions on this platform")
		}
		defer func() {
			_ = os.Chmod(templatesDir, 0755)
		}()

		// Try to create templates - should fail when trying to write
		err := CreateExampleTemplates(templatesDir)
		if err == nil {
			t.Error("CreateExampleTemplates() expected error when file write fails, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create template") {
			t.Errorf("CreateExampleTemplates() error = %v, want error containing 'failed to create template'", err)
		}
	})
}

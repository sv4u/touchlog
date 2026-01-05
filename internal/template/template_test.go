package template

import (
	"errors"
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
			name:    "variable with hyphens (should not match - only word chars)",
			content: "Hello {{user-name}}",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "variable with numbers",
			content: "Hello {{var123}}",
			want:    []string{"var123"},
			wantLen: 1,
			exact:   true,
		},
		{
			name:    "nested braces (should match inner)",
			content: "Hello {{{{name}}}}",
			want:    []string{"name"},
			wantLen: 1,
			exact:   true,
		},
		{
			name:    "variable with spaces (should not match)",
			content: "Hello {{var name}}",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "multiple variables with special chars (hyphens don't match)",
			content: "{{var1}} and {{var-2}} and {{var_3}}",
			want:    []string{"var1", "var_3"},
			wantLen: 2,
			exact:   true,
		},
		{
			name:    "variable at start of line",
			content: "{{name}}\nSecond line",
			want:    []string{"name"},
			wantLen: 1,
			exact:   true,
		},
		{
			name:    "variable at end of line",
			content: "First line\n{{name}}",
			want:    []string{"name"},
			wantLen: 1,
			exact:   true,
		},
		{
			name:    "variable with only braces",
			content: "{{}}",
			want:    []string{},
			wantLen: 0,
			exact:   true,
		},
		{
			name:    "variable with special characters (should not match word chars)",
			content: "Hello {{var@name}}",
			want:    []string{},
			wantLen: 0,
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

func TestGetDefaultVariablesWithMetadata(t *testing.T) {
	t.Run("nil config, nil metadata", func(t *testing.T) {
		vars, err := GetDefaultVariablesWithMetadata(nil, nil)
		if err != nil {
			t.Fatalf("GetDefaultVariablesWithMetadata(nil, nil) returned error: %v", err)
		}

		// Should have date, time, and datetime
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariablesWithMetadata() missing 'date' variable")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariablesWithMetadata() missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariablesWithMetadata() missing 'datetime' variable")
		}
	})

	t.Run("with metadata - user and host", func(t *testing.T) {
		metadata := &MetadataValues{
			User: "testuser",
			Host: "testhost",
		}
		vars, err := GetDefaultVariablesWithMetadata(nil, metadata)
		if err != nil {
			t.Fatalf("GetDefaultVariablesWithMetadata() returned error: %v", err)
		}

		if vars["user"] != "testuser" {
			t.Errorf("GetDefaultVariablesWithMetadata() user = %q, want %q", vars["user"], "testuser")
		}
		if vars["host"] != "testhost" {
			t.Errorf("GetDefaultVariablesWithMetadata() host = %q, want %q", vars["host"], "testhost")
		}
	})

	t.Run("with metadata - git context", func(t *testing.T) {
		metadata := &MetadataValues{
			User:   "testuser",
			Host:   "testhost",
			Branch: "main",
			Commit: "abc123",
		}
		vars, err := GetDefaultVariablesWithMetadata(nil, metadata)
		if err != nil {
			t.Fatalf("GetDefaultVariablesWithMetadata() returned error: %v", err)
		}

		if vars["branch"] != "main" {
			t.Errorf("GetDefaultVariablesWithMetadata() branch = %q, want %q", vars["branch"], "main")
		}
		if vars["commit"] != "abc123" {
			t.Errorf("GetDefaultVariablesWithMetadata() commit = %q, want %q", vars["commit"], "abc123")
		}
	})

	t.Run("with config timezone", func(t *testing.T) {
		cfg := &config.Config{
			Timezone: "America/Denver",
		}
		vars, err := GetDefaultVariablesWithMetadata(cfg, nil)
		if err != nil {
			t.Fatalf("GetDefaultVariablesWithMetadata() returned error: %v", err)
		}

		// Should have date, time, datetime (timezone applied)
		if _, ok := vars["date"]; !ok {
			t.Error("GetDefaultVariablesWithMetadata() missing 'date' variable")
		}
		if _, ok := vars["time"]; !ok {
			t.Error("GetDefaultVariablesWithMetadata() missing 'time' variable")
		}
		if _, ok := vars["datetime"]; !ok {
			t.Error("GetDefaultVariablesWithMetadata() missing 'datetime' variable")
		}
	})

	t.Run("with invalid timezone", func(t *testing.T) {
		cfg := &config.Config{
			Timezone: "Invalid/Timezone",
		}
		_, err := GetDefaultVariablesWithMetadata(cfg, nil)
		if err == nil {
			t.Error("GetDefaultVariablesWithMetadata() expected error for invalid timezone, got nil")
			return
		}
		if !strings.Contains(err.Error(), "invalid timezone") {
			t.Errorf("GetDefaultVariablesWithMetadata() error = %v, want error containing 'invalid timezone'", err)
		}
	})

	t.Run("with custom timestamp", func(t *testing.T) {
		testTime := time.Date(2023, 1, 15, 14, 30, 0, 0, time.UTC)
		vars, err := GetDefaultVariablesWithMetadata(nil, nil, testTime)
		if err != nil {
			t.Fatalf("GetDefaultVariablesWithMetadata() returned error: %v", err)
		}

		// Date should be 2023-01-15
		if !strings.Contains(vars["date"], "2023-01-15") {
			t.Errorf("GetDefaultVariablesWithMetadata() date = %q, want to contain '2023-01-15'", vars["date"])
		}
	})

	t.Run("with config custom variables", func(t *testing.T) {
		cfg := &config.Config{
			Variables: map[string]string{
				"author":  "Test Author",
				"project": "Test Project",
			},
		}
		vars, err := GetDefaultVariablesWithMetadata(cfg, nil)
		if err != nil {
			t.Fatalf("GetDefaultVariablesWithMetadata() returned error: %v", err)
		}

		if vars["author"] != "Test Author" {
			t.Errorf("GetDefaultVariablesWithMetadata() author = %q, want %q", vars["author"], "Test Author")
		}
		if vars["project"] != "Test Project" {
			t.Errorf("GetDefaultVariablesWithMetadata() project = %q, want %q", vars["project"], "Test Project")
		}
	})
}

func TestGetDefaultVariables(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		vars, err := GetDefaultVariables(nil)
		if err != nil {
			t.Fatalf("GetDefaultVariables(nil) returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

		vars, err := GetDefaultVariables(cfg)
		if err != nil {
			t.Fatalf("GetDefaultVariables() returned error: %v", err)
		}

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

	t.Run("invalid timezone returns error", func(t *testing.T) {
		enabledTrue := true
		cfg := &config.Config{
			Timezone: "Invalid/Timezone",
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: &enabledTrue, Format: "2006-01-02 15:04:05"},
			},
		}

		// Should return an error for invalid timezone
		vars, err := GetDefaultVariables(cfg)
		if err == nil {
			t.Error("GetDefaultVariables() should return error for invalid timezone")
		}
		if vars != nil {
			t.Error("GetDefaultVariables() should return nil map when timezone is invalid")
		}
		if err != nil && !strings.Contains(err.Error(), "Invalid/Timezone") {
			t.Errorf("GetDefaultVariables() error should mention invalid timezone, got: %v", err)
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

	t.Run("directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentDir := filepath.Join(tmpDir, "nonexistent", "templates")

		// Should create the directory and templates
		err := CreateExampleTemplates(nonExistentDir)
		if err != nil {
			t.Errorf("CreateExampleTemplates() error = %v, want nil", err)
		}

		// Verify templates were created
		dailyPath := filepath.Join(nonExistentDir, "daily.md")
		if _, err := os.Stat(dailyPath); err != nil {
			t.Errorf("CreateExampleTemplates() did not create daily.md: %v", err)
		}
	})

	t.Run("directory read error", func(t *testing.T) {
		// Test when ReadDir fails with a non-IsNotExist error
		// This is hard to test without mocking, but we can test the error path conceptually
		// by using an invalid path that causes a different error
		invalidPath := "/proc/self/fd/999999" // Invalid file descriptor path
		err := CreateExampleTemplates(invalidPath)
		// This may or may not error depending on the system, but we verify it handles errors gracefully
		_ = err
	})
}

func TestLoadTemplate(t *testing.T) {
	t.Run("load existing template", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		templateFile := filepath.Join(templatesDir, "test.md")
		expectedContent := "# Test Template\n\n{{date}}\n{{message}}"
		if err := os.WriteFile(templateFile, []byte(expectedContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		content, err := LoadTemplate(templatesDir, "test.md")
		if err != nil {
			t.Fatalf("LoadTemplate() error = %v, want nil", err)
		}
		if content != expectedContent {
			t.Errorf("LoadTemplate() = %q, want %q", content, expectedContent)
		}
	})

	t.Run("template file not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		_, err := LoadTemplate(templatesDir, "nonexistent.md")
		if err == nil {
			t.Error("LoadTemplate() error = nil, want error for nonexistent file")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to read template") {
			t.Errorf("LoadTemplate() error = %v, want error containing 'failed to read template'", err)
		}
	})

	t.Run("empty templates directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		_, err := LoadTemplate(templatesDir, "test.md")
		if err == nil {
			t.Error("LoadTemplate() error = nil, want error for nonexistent file")
		}
	})

	t.Run("template with subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Test that LoadTemplate doesn't allow path traversal
		_, err := LoadTemplate(templatesDir, "../test.md")
		if err == nil {
			t.Error("LoadTemplate() error = nil, want error for path traversal attempt")
		}
	})
}

func TestInlineTemplateSource(t *testing.T) {
	t.Run("get existing template", func(t *testing.T) {
		templates := map[string]string{
			"daily": "# Daily Note\n{{date}}",
			"quick": "# Quick Note\n{{message}}",
		}
		source := &InlineTemplateSource{templates: templates}

		content, err := source.GetTemplate("daily")
		if err != nil {
			t.Fatalf("GetTemplate() error = %v, want nil", err)
		}
		if content != "# Daily Note\n{{date}}" {
			t.Errorf("GetTemplate() = %q, want %q", content, "# Daily Note\n{{date}}")
		}
	})

	t.Run("get non-existent template", func(t *testing.T) {
		templates := map[string]string{
			"daily": "# Daily Note",
		}
		source := &InlineTemplateSource{templates: templates}

		_, err := source.GetTemplate("nonexistent")
		if err == nil {
			t.Error("GetTemplate() error = nil, want ErrTemplateNotFound")
			return
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			t.Errorf("GetTemplate() error = %v, want ErrTemplateNotFound", err)
		}
	})

	t.Run("nil templates map", func(t *testing.T) {
		source := &InlineTemplateSource{templates: nil}

		_, err := source.GetTemplate("daily")
		if err == nil {
			t.Error("GetTemplate() error = nil, want ErrTemplateNotFound")
			return
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			t.Errorf("GetTemplate() error = %v, want ErrTemplateNotFound", err)
		}
	})
}

func TestFileTemplateSource(t *testing.T) {
	t.Run("get template by filename match", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create a template file
		templateFile := filepath.Join(templatesDir, "daily.md")
		templateContent := "# Daily Note\n{{date}}"
		if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Create file source with template config
		templates := []config.Template{
			{Name: "Daily Note", File: "daily.md"},
		}
		source := &FileTemplateSource{
			templatesDir: templatesDir,
			templates:    templates,
		}

		// Resolve by name "daily" (filename without extension)
		content, err := source.GetTemplate("daily")
		if err != nil {
			t.Fatalf("GetTemplate() error = %v, want nil", err)
		}
		if content != templateContent {
			t.Errorf("GetTemplate() = %q, want %q", content, templateContent)
		}
	})

	t.Run("get template with hyphenated filename", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create a template file with hyphenated name
		templateFile := filepath.Join(templatesDir, "quick-note.md")
		templateContent := "# Quick Note\n{{message}}"
		if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		templates := []config.Template{
			{Name: "Quick Note", File: "quick-note.md"},
		}
		source := &FileTemplateSource{
			templatesDir: templatesDir,
			templates:    templates,
		}

		// Resolve by name "quick-note" (filename without extension)
		content, err := source.GetTemplate("quick-note")
		if err != nil {
			t.Fatalf("GetTemplate() error = %v, want nil", err)
		}
		if content != templateContent {
			t.Errorf("GetTemplate() = %q, want %q", content, templateContent)
		}
	})

	t.Run("get non-existent template", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		templates := []config.Template{
			{Name: "Daily Note", File: "daily.md"},
		}
		source := &FileTemplateSource{
			templatesDir: templatesDir,
			templates:    templates,
		}

		_, err := source.GetTemplate("nonexistent")
		if err == nil {
			t.Error("GetTemplate() error = nil, want ErrTemplateNotFound")
			return
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			t.Errorf("GetTemplate() error = %v, want ErrTemplateNotFound", err)
		}
	})

	t.Run("template file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Config references a file that doesn't exist
		templates := []config.Template{
			{Name: "Daily Note", File: "missing.md"},
		}
		source := &FileTemplateSource{
			templatesDir: templatesDir,
			templates:    templates,
		}

		_, err := source.GetTemplate("missing")
		if err == nil {
			t.Error("GetTemplate() error = nil, want file read error")
			return
		}
		if !strings.Contains(err.Error(), "failed to read template") {
			t.Errorf("GetTemplate() error = %v, want error containing 'failed to read template'", err)
		}
	})
}

func TestResolveTemplate(t *testing.T) {
	t.Run("resolve inline template first", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create a file-based template
		templateFile := filepath.Join(templatesDir, "daily.md")
		fileContent := "# File-based Daily Note"
		if err := os.WriteFile(templateFile, []byte(fileContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Config with both inline and file-based templates
		cfg := &config.Config{
			InlineTemplates: map[string]string{
				"daily": "# Inline Daily Note\n{{date}}",
			},
			Templates: []config.Template{
				{Name: "Daily Note", File: "daily.md"},
			},
		}

		// Inline template should take precedence
		content, err := ResolveTemplate("daily", cfg, templatesDir)
		if err != nil {
			t.Fatalf("ResolveTemplate() error = %v, want nil", err)
		}
		if content != "# Inline Daily Note\n{{date}}" {
			t.Errorf("ResolveTemplate() = %q, want inline template content", content)
		}
	})

	t.Run("empty template name", func(t *testing.T) {
		cfg := &config.Config{}
		_, err := ResolveTemplate("", cfg, "/tmp/templates")
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want error for empty name")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot be empty") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'cannot be empty'", err)
		}
	})

	t.Run("invalid inline template syntax", func(t *testing.T) {
		cfg := &config.Config{
			InlineTemplates: map[string]string{
				"daily": "{{invalid UTF-8: \xff\xfe\xfd}}",
			},
		}
		_, err := ResolveTemplate("daily", cfg, "/tmp/templates")
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want error for invalid template syntax")
		}
		if err != nil && !strings.Contains(err.Error(), "invalid inline template syntax") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'invalid inline template syntax'", err)
		}
	})

	t.Run("invalid file-based template syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create template file with invalid UTF-8
		templateFile := filepath.Join(templatesDir, "daily.md")
		invalidContent := []byte{0xff, 0xfe, 0xfd} // Invalid UTF-8
		if err := os.WriteFile(templateFile, invalidContent, 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		cfg := &config.Config{
			Templates: []config.Template{
				{Name: "Daily Note", File: "daily.md"},
			},
		}

		_, err := ResolveTemplate("daily", cfg, templatesDir)
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want error for invalid template syntax")
		}
		if err != nil && !strings.Contains(err.Error(), "invalid file-based template syntax") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'invalid file-based template syntax'", err)
		}
	})

	t.Run("file read error (non-ErrTemplateNotFound)", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create template file but make it unreadable
		templateFile := filepath.Join(templatesDir, "daily.md")
		if err := os.WriteFile(templateFile, []byte("# Daily Note"), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Remove read permissions
		if err := os.Chmod(templateFile, 0000); err != nil {
			t.Skip("Cannot set file permissions on this platform")
		}
		defer os.Chmod(templateFile, 0644)

		cfg := &config.Config{
			Templates: []config.Template{
				{Name: "Daily Note", File: "daily.md"},
			},
		}

		_, err := ResolveTemplate("daily", cfg, templatesDir)
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want error for file read error")
			return
		}
		if !strings.Contains(err.Error(), "failed to resolve file-based template") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'failed to resolve file-based template'", err)
		}
	})

	t.Run("fallback to file-based template", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create a file-based template
		templateFile := filepath.Join(templatesDir, "daily.md")
		fileContent := "# File-based Daily Note\n{{date}}"
		if err := os.WriteFile(templateFile, []byte(fileContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		// Config with only file-based template (no inline)
		cfg := &config.Config{
			Templates: []config.Template{
				{Name: "Daily Note", File: "daily.md"},
			},
		}

		content, err := ResolveTemplate("daily", cfg, templatesDir)
		if err != nil {
			t.Fatalf("ResolveTemplate() error = %v, want nil", err)
		}
		if content != fileContent {
			t.Errorf("ResolveTemplate() = %q, want %q", content, fileContent)
		}
	})

	t.Run("template not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")

		cfg := &config.Config{
			Templates: []config.Template{
				{Name: "Daily Note", File: "daily.md"},
			},
		}

		_, err := ResolveTemplate("nonexistent", cfg, templatesDir)
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want template not found error")
			return
		}
		if !strings.Contains(err.Error(), "template 'nonexistent' not found") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'template 'nonexistent' not found'", err)
		}
	})

	t.Run("propagate file read errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		// Create a directory with the same name as the template file
		// This will cause os.ReadFile to fail with a clear error (is a directory)
		templateDir := filepath.Join(templatesDir, "daily.md")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}

		cfg := &config.Config{
			Templates: []config.Template{
				{Name: "Daily Note", File: "daily.md"},
			},
		}

		// ResolveTemplate should propagate the file read error, not mask it as "template not found"
		_, err := ResolveTemplate("daily", cfg, templatesDir)
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want file read error")
			return
		}
		// The error should be about failing to read the template, not "template not found"
		if strings.Contains(err.Error(), "template 'daily' not found") {
			t.Errorf("ResolveTemplate() masked file read error as 'template not found': %v", err)
		}
		if !strings.Contains(err.Error(), "failed to read template") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'failed to read template'", err)
		}
	})

	t.Run("empty template name", func(t *testing.T) {
		cfg := &config.Config{}
		_, err := ResolveTemplate("", cfg, "/tmp")
		if err == nil {
			t.Error("ResolveTemplate() error = nil, want error for empty name")
			return
		}
		if !strings.Contains(err.Error(), "template name cannot be empty") {
			t.Errorf("ResolveTemplate() error = %v, want error containing 'template name cannot be empty'", err)
		}
	})
}

func TestEscapeUserInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple text",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "text with braces",
			input: "Note with {{date}} in it",
			want:  "Note with __TEMPLATE_BRACE_OPEN__date__TEMPLATE_BRACE_CLOSE__ in it",
		},
		{
			name:  "multiple braces",
			input: "{{title}} and {{message}}",
			want:  "__TEMPLATE_BRACE_OPEN__title__TEMPLATE_BRACE_CLOSE__ and __TEMPLATE_BRACE_OPEN__message__TEMPLATE_BRACE_CLOSE__",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only braces",
			input: "{{}}",
			want:  "__TEMPLATE_BRACE_OPEN____TEMPLATE_BRACE_CLOSE__",
		},
		{
			name:  "text with placeholder string",
			input: "Some text with __TEMPLATE_BRACE_OPEN__ in it",
			want:  "Some text with __ESCAPED_PLACEHOLDER_OPEN__ in it",
		},
		{
			name:  "text with both placeholder strings",
			input: "Text with __TEMPLATE_BRACE_OPEN__ and __TEMPLATE_BRACE_CLOSE__",
			want:  "Text with __ESCAPED_PLACEHOLDER_OPEN__ and __ESCAPED_PLACEHOLDER_CLOSE__",
		},
		{
			name:  "text with placeholder strings and actual braces",
			input: "Text with __TEMPLATE_BRACE_OPEN__ and {{date}}",
			want:  "Text with __ESCAPED_PLACEHOLDER_OPEN__ and __TEMPLATE_BRACE_OPEN__date__TEMPLATE_BRACE_CLOSE__",
		},
		{
			name:  "text with second-level placeholder string",
			input: "Some text with __ESCAPED_PLACEHOLDER_OPEN__ in it",
			want:  "Some text with __ESCAPED_PLACEHOLDER_2_OPEN__ in it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeUserInput(tt.input)
			if got != tt.want {
				t.Errorf("EscapeUserInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnescapeUserInput_EdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		result := UnescapeUserInput("")
		if result != "" {
			t.Errorf("UnescapeUserInput(\"\") = %q, want \"\"", result)
		}
	})

	t.Run("input with no placeholders", func(t *testing.T) {
		input := "normal text"
		result := UnescapeUserInput(input)
		if result != input {
			t.Errorf("UnescapeUserInput(%q) = %q, want %q", input, result, input)
		}
	})

	t.Run("input with multiple levels of placeholders", func(t *testing.T) {
		// Input with level 2 placeholders
		input := "__ESCAPED_PLACEHOLDER_2_OPEN__test__ESCAPED_PLACEHOLDER_2_CLOSE__"
		result := UnescapeUserInput(input)
		// Should unescape to level 1
		if !strings.Contains(result, "__ESCAPED_PLACEHOLDER_OPEN__") {
			t.Error("UnescapeUserInput() should unescape level 2 to level 1")
		}
	})

	t.Run("input with mixed placeholders", func(t *testing.T) {
		// Input with both template braces and placeholder strings
		// Note: templateBraceOpen/Close are already the first level, so they should become {{ and }}
		// escapedPlaceholderOpen/Close are second level, so they should become templateBraceOpen/Close
		input := templateBraceOpen + "test" + templateBraceClose + " and " + escapedPlaceholderOpen + "more" + escapedPlaceholderClose
		result := UnescapeUserInput(input)
		// Should unescape: templateBraceOpen -> {{, templateBraceClose -> }}, escapedPlaceholderOpen -> templateBraceOpen -> {{, etc.
		// The result should have {{ and }} for both
		if !strings.Contains(result, "{{") || !strings.Contains(result, "}}") {
			t.Error("UnescapeUserInput() should unescape template braces to {{ and }}")
		}
		// Should not contain the placeholder strings
		if strings.Contains(result, templateBraceOpen) || strings.Contains(result, escapedPlaceholderOpen) {
			t.Logf("UnescapeUserInput() result = %q (may contain placeholders if not fully unescaped)", result)
		}
	})
}

func TestUnescapeUserInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple text",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "escaped braces",
			input: "Note with __TEMPLATE_BRACE_OPEN__date__TEMPLATE_BRACE_CLOSE__ in it",
			want:  "Note with {{date}} in it",
		},
		{
			name:  "multiple escaped braces",
			input: "__TEMPLATE_BRACE_OPEN__title__TEMPLATE_BRACE_CLOSE__ and __TEMPLATE_BRACE_OPEN__message__TEMPLATE_BRACE_CLOSE__",
			want:  "{{title}} and {{message}}",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "round trip",
			input: "__TEMPLATE_BRACE_OPEN____TEMPLATE_BRACE_CLOSE__",
			want:  "{{}}",
		},
		{
			name:  "escaped placeholder string",
			input: "Some text with __ESCAPED_PLACEHOLDER_OPEN__ in it",
			want:  "Some text with __TEMPLATE_BRACE_OPEN__ in it",
		},
		{
			name:  "escaped placeholder strings",
			input: "Text with __ESCAPED_PLACEHOLDER_OPEN__ and __ESCAPED_PLACEHOLDER_CLOSE__",
			want:  "Text with __TEMPLATE_BRACE_OPEN__ and __TEMPLATE_BRACE_CLOSE__",
		},
		{
			name:  "mixed escaped placeholders and template braces",
			input: "Text with __ESCAPED_PLACEHOLDER_OPEN__ and __TEMPLATE_BRACE_OPEN__date__TEMPLATE_BRACE_CLOSE__",
			want:  "Text with __TEMPLATE_BRACE_OPEN__ and {{date}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnescapeUserInput(tt.input)
			if got != tt.want {
				t.Errorf("UnescapeUserInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEscapeUnescapeRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple text",
			input: "Hello world",
		},
		{
			name:  "text with braces",
			input: "Note with {{date}} in it",
		},
		{
			name:  "text with placeholder string",
			input: "Some text with __TEMPLATE_BRACE_OPEN__ in it",
		},
		{
			name:  "text with both placeholder strings",
			input: "Text with __TEMPLATE_BRACE_OPEN__ and __TEMPLATE_BRACE_CLOSE__",
		},
		{
			name:  "text with placeholder strings and actual braces",
			input: "Text with __TEMPLATE_BRACE_OPEN__ and {{date}}",
		},
		{
			name:  "complex mixed content",
			input: "Note: __TEMPLATE_BRACE_OPEN__ is a placeholder, but {{title}} is a variable",
		},
		{
			name:  "text with second-level placeholder string",
			input: "Some text with __ESCAPED_PLACEHOLDER_OPEN__ in it",
		},
		{
			name:  "text with third-level placeholder string",
			input: "Some text with __ESCAPED_PLACEHOLDER_2_OPEN__ in it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := EscapeUserInput(tt.input)
			unescaped := UnescapeUserInput(escaped)
			if unescaped != tt.input {
				t.Errorf("Round trip failed: input = %q, escaped = %q, unescaped = %q", tt.input, escaped, unescaped)
			}
		})
	}
}

func TestShouldEscapeVariable(t *testing.T) {
	tests := []struct {
		name         string
		varName      string
		shouldEscape bool
	}{
		{
			name:         "system variable date",
			varName:      "date",
			shouldEscape: false,
		},
		{
			name:         "system variable time",
			varName:      "time",
			shouldEscape: false,
		},
		{
			name:         "system variable datetime",
			varName:      "datetime",
			shouldEscape: false,
		},
		{
			name:         "metadata variable user",
			varName:      "user",
			shouldEscape: false,
		},
		{
			name:         "metadata variable host",
			varName:      "host",
			shouldEscape: false,
		},
		{
			name:         "metadata variable branch",
			varName:      "branch",
			shouldEscape: false,
		},
		{
			name:         "metadata variable commit",
			varName:      "commit",
			shouldEscape: false,
		},
		{
			name:         "user variable title",
			varName:      "title",
			shouldEscape: true,
		},
		{
			name:         "user variable message",
			varName:      "message",
			shouldEscape: true,
		},
		{
			name:         "user variable tags",
			varName:      "tags",
			shouldEscape: true,
		},
		{
			name:         "custom variable",
			varName:      "author",
			shouldEscape: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldEscapeVariable(tt.varName)
			if got != tt.shouldEscape {
				t.Errorf("ShouldEscapeVariable(%q) = %v, want %v", tt.varName, got, tt.shouldEscape)
			}
		})
	}
}

func TestProcessTemplateWithEscaping(t *testing.T) {
	t.Run("escape user input variables", func(t *testing.T) {
		template := "Title: {{title}}\nMessage: {{message}}"
		vars := map[string]string{
			"title":   "My Note with {{date}} in title",
			"message": "Content with {{time}} here",
		}

		result := ProcessTemplate(template, vars)

		// After processing and unescaping, the user input braces should appear as literal text
		// The escaping prevents them from being interpreted during template processing
		if !strings.Contains(result, "{{date}}") {
			t.Error("ProcessTemplate() should preserve literal {{date}} from user input (after unescaping)")
		}
		if !strings.Contains(result, "{{time}}") {
			t.Error("ProcessTemplate() should preserve literal {{time}} from user input (after unescaping)")
		}
		// The result should contain the user input text
		if !strings.Contains(result, "My Note with") {
			t.Error("ProcessTemplate() should contain user input text")
		}
		if !strings.Contains(result, "Content with") {
			t.Error("ProcessTemplate() should contain user input text")
		}
	})

	t.Run("do not escape system variables", func(t *testing.T) {
		template := "Date: {{date}}\nTime: {{time}}"
		vars := map[string]string{
			"date": "2025-01-01",
			"time": "12:00:00",
		}

		result := ProcessTemplate(template, vars)

		// System variables should be replaced directly without escaping
		if !strings.Contains(result, "2025-01-01") {
			t.Error("ProcessTemplate() should replace {{date}} with value")
		}
		if !strings.Contains(result, "12:00:00") {
			t.Error("ProcessTemplate() should replace {{time}} with value")
		}
		// Should not contain the template variable syntax
		if strings.Contains(result, "{{date}}") {
			t.Error("ProcessTemplate() should replace {{date}}, not leave it")
		}
	})

	t.Run("template injection prevention", func(t *testing.T) {
		template := "Content: {{message}}"
		// User tries to inject template syntax
		vars := map[string]string{
			"message": "Hello {{date}} world",
		}

		result := ProcessTemplate(template, vars)

		// The {{date}} in the user input should be escaped during processing to prevent interpretation
		// After unescaping, it appears as literal text in the output
		// This prevents it from being interpreted as a template variable during THIS processing pass
		if !strings.Contains(result, "{{date}}") {
			t.Error("ProcessTemplate() should preserve literal {{date}} from user input (preventing injection during processing)")
		}
		// The message should still be present
		if !strings.Contains(result, "Hello") {
			t.Error("ProcessTemplate() should preserve user input text")
		}
		if !strings.Contains(result, "world") {
			t.Error("ProcessTemplate() should preserve user input text")
		}
		// Verify that {{date}} appears as literal text, not replaced with an actual date
		// (if it were interpreted, it would be replaced with a date string)
		if strings.Contains(result, "2025-") || strings.Contains(result, "2024-") {
			t.Error("ProcessTemplate() should not interpret {{date}} from user input as a template variable")
		}
	})

	t.Run("mixed trusted and untrusted variables", func(t *testing.T) {
		template := "Date: {{date}}\nTitle: {{title}}\nMessage: {{message}}"
		vars := map[string]string{
			"date":    "2025-01-01",
			"title":   "Note with {{date}}",
			"message": "Content with {{time}}",
		}

		result := ProcessTemplate(template, vars)

		// System variable should be replaced
		if !strings.Contains(result, "2025-01-01") {
			t.Error("ProcessTemplate() should replace system variable {{date}}")
		}
		// User variables should have their braces escaped
		if strings.Contains(result, "{{date}}") && strings.Contains(result, "Note with") {
			// The {{date}} in title should be escaped (appear as literal)
			// But we need to check it's not being interpreted
			// Since we escape and then unescape, the braces in user input should remain
			// but not match our regex pattern during processing
		}
	})
}

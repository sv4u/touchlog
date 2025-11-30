package template

import (
	"testing"

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
		name     string
		content  string
		want     []string
		wantLen  int
		exact    bool // If true, check exact match; if false, only check length
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: true, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02 15:04:05"},
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: false, Format: ""},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02 15:04:05"},
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: "01/02/2006"},
				Time:     config.DateTimeVarConfig{Enabled: true, Format: "03:04 PM"},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: "01/02/2006 03:04 PM"},
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: "invalid-format"},
				Time:     config.DateTimeVarConfig{Enabled: true, Format: "also-invalid"},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: "invalid-too"},
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: ""},
				Time:     config.DateTimeVarConfig{Enabled: true, Format: ""},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: ""},
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: false, Format: ""},
				Time:     config.DateTimeVarConfig{Enabled: false, Format: ""},
				DateTime: config.DateTimeVarConfig{Enabled: false, Format: ""},
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: true, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02 15:04:05"},
			},
			Variables: map[string]string{
				"author": "Test Author",
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
		cfg := &config.Config{
			DateTimeVars: config.DateTimeVarsConfig{
				Date:     config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02"},
				Time:     config.DateTimeVarConfig{Enabled: true, Format: "15:04:05"},
				DateTime: config.DateTimeVarConfig{Enabled: true, Format: "2006-01-02 15:04:05"},
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
}


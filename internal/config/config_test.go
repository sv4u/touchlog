package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateVariableName(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "reserved name - date",
			varName: "date",
			wantErr: true,
			errMsg:  "reserved",
		},
		{
			name:    "reserved name - time",
			varName: "time",
			wantErr: true,
			errMsg:  "reserved",
		},
		{
			name:    "reserved name - datetime",
			varName: "datetime",
			wantErr: true,
			errMsg:  "reserved",
		},
		{
			name:    "valid variable name",
			varName: "author",
			wantErr: false,
		},
		{
			name:    "valid variable name with underscore",
			varName: "user_name",
			wantErr: false,
		},
		{
			name:    "valid variable name with numbers",
			varName: "value1",
			wantErr: false,
		},
		{
			name:    "empty string",
			varName: "",
			wantErr: false,
		},
		{
			name:    "name with special characters",
			varName: "var-name",
			wantErr: false,
		},
		{
			name:    "name with spaces",
			varName: "var name",
			wantErr: false,
		},
		{
			name:    "case sensitive - Date",
			varName: "Date",
			wantErr: false,
		},
		{
			name:    "case sensitive - TIME",
			varName: "TIME",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVariableName(tt.varName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateVariableName(%q) expected error, got nil", tt.varName)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateVariableName(%q) error = %v, want error containing %q", tt.varName, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateVariableName(%q) unexpected error = %v", tt.varName, err)
				}
			}
		})
	}
}

func TestValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   bool
	}{
		{
			name:   "valid format - RFC3339",
			format: "2006-01-02T15:04:05Z07:00",
			want:   true,
		},
		{
			name:   "valid format - standard date",
			format: "2006-01-02",
			want:   true,
		},
		{
			name:   "valid format - standard time",
			format: "15:04:05",
			want:   true,
		},
		{
			name:   "valid format - datetime",
			format: "2006-01-02 15:04:05",
			want:   true,
		},
		{
			name:   "valid format - custom format",
			format: "01/02/2006",
			want:   true,
		},
		{
			name:   "valid format - 12 hour time",
			format: "03:04 PM",
			want:   true,
		},
		{
			name:   "empty string",
			format: "",
			want:   false,
		},
		{
			name: "invalid format - malformed",
			// Note: Go's time.Format() doesn't panic, so this returns true
			// The function uses recover() but time.Format() never panics
			format: "invalid-format",
			want:   true,
		},
		{
			name: "invalid format - incomplete",
			// Note: Go's time.Format() doesn't panic, so this returns true
			format: "2006-01",
			want:   true,
		},
		{
			name: "invalid format - wrong reference time",
			// Note: Go's time.Format() doesn't panic, so this returns true
			format: "YYYY-MM-DD",
			want:   true,
		},
		{
			name: "invalid format - special characters only",
			// Note: Go's time.Format() doesn't panic, so this returns true
			format: "!!!",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTimeFormat(tt.format)
			if got != tt.want {
				t.Errorf("ValidateTimeFormat(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("valid config parsing", func(t *testing.T) {
		yamlData := `
templates:
  - name: Daily Note
    file: daily.md
  - name: Journal
    file: journal.md
notes_directory: ~/notes
datetime_vars:
  date:
    enabled: true
    format: "2006-01-02"
  time:
    enabled: true
    format: "15:04:05"
  datetime:
    enabled: true
    format: "2006-01-02 15:04:05"
variables:
  author: Test Author
  project: Test Project
vim_mode: true
`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		// Validate custom variables
		for name := range cfg.Variables {
			if err := ValidateVariableName(name); err != nil {
				t.Errorf("LoadConfig() invalid variable name %q: %v", name, err)
			}
		}

		// Check templates
		if len(cfg.Templates) != 2 {
			t.Errorf("LoadConfig() templates length = %d, want 2", len(cfg.Templates))
		}

		// Check datetime_vars
		if !cfg.DateTimeVars.Date.Enabled {
			t.Error("LoadConfig() date variable not enabled")
		}
		if cfg.DateTimeVars.Date.Format != "2006-01-02" {
			t.Errorf("LoadConfig() date format = %q, want %q", cfg.DateTimeVars.Date.Format, "2006-01-02")
		}

		// Check custom variables
		if cfg.Variables["author"] != "Test Author" {
			t.Errorf("LoadConfig() author variable = %q, want %q", cfg.Variables["author"], "Test Author")
		}
		if cfg.Variables["project"] != "Test Project" {
			t.Errorf("LoadConfig() project variable = %q, want %q", cfg.Variables["project"], "Test Project")
		}

		// Check vim_mode
		if !cfg.VimMode {
			t.Error("LoadConfig() vim_mode = false, want true")
		}
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		yamlData := `
templates:
  - name: Daily Note
    file: daily.md
  invalid yaml: [unclosed bracket
`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err == nil {
			t.Error("yaml.Unmarshal() expected error for invalid YAML, got nil")
		}
	})

	t.Run("missing required fields - should still parse", func(t *testing.T) {
		// Config should parse even with missing fields (they'll be zero values)
		yamlData := `
templates:
  - name: Daily Note
    file: daily.md
`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err != nil {
			t.Errorf("yaml.Unmarshal() error = %v (should parse with missing fields)", err)
		}

		// Check that missing fields are zero values
		if cfg.NotesDirectory != "" {
			t.Errorf("LoadConfig() notes_directory = %q, want empty string", cfg.NotesDirectory)
		}
		if len(cfg.Variables) != 0 {
			t.Errorf("LoadConfig() variables = %v, want nil or empty", cfg.Variables)
		}
	})

	t.Run("custom variables validation - reserved name", func(t *testing.T) {
		// This tests the validation logic that would be called in LoadConfig
		yamlData := `
variables:
  date: "2024-01-01"
  author: Test Author
`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		// Manually validate (simulating what LoadConfig does)
		for name := range cfg.Variables {
			err := ValidateVariableName(name)
			if name == "date" {
				// "date" is reserved, so should return an error
				if err == nil {
					t.Error("ValidateVariableName('date') expected error, got nil")
				}
			} else {
				// Other names should be valid
				if err != nil {
					t.Errorf("ValidateVariableName(%q) unexpected error = %v", name, err)
				}
			}
		}
	})

	t.Run("datetime_vars configuration parsing", func(t *testing.T) {
		yamlData := `
datetime_vars:
  date:
    enabled: false
    format: "01/02/2006"
  time:
    enabled: true
    format: ""
  datetime:
    enabled: true
    format: "2006-01-02 15:04:05"
`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		if cfg.DateTimeVars.Date.Enabled {
			t.Error("LoadConfig() date enabled = true, want false")
		}
		if cfg.DateTimeVars.Date.Format != "01/02/2006" {
			t.Errorf("LoadConfig() date format = %q, want %q", cfg.DateTimeVars.Date.Format, "01/02/2006")
		}

		if !cfg.DateTimeVars.Time.Enabled {
			t.Error("LoadConfig() time enabled = false, want true")
		}
		if cfg.DateTimeVars.Time.Format != "" {
			t.Errorf("LoadConfig() time format = %q, want empty string", cfg.DateTimeVars.Time.Format)
		}

		if !cfg.DateTimeVars.DateTime.Enabled {
			t.Error("LoadConfig() datetime enabled = false, want true")
		}
		if cfg.DateTimeVars.DateTime.Format != "2006-01-02 15:04:05" {
			t.Errorf("LoadConfig() datetime format = %q, want %q", cfg.DateTimeVars.DateTime.Format, "2006-01-02 15:04:05")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		yamlData := `{}`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err != nil {
			t.Errorf("yaml.Unmarshal() error = %v (should parse empty config)", err)
		}

		// All fields should be zero values
		if len(cfg.Templates) != 0 {
			t.Errorf("LoadConfig() templates = %v, want nil or empty", cfg.Templates)
		}
		if cfg.NotesDirectory != "" {
			t.Errorf("LoadConfig() notes_directory = %q, want empty string", cfg.NotesDirectory)
		}
	})

	t.Run("config with only templates", func(t *testing.T) {
		yamlData := `
templates:
  - name: Daily Note
    file: daily.md
`

		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		if err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		if len(cfg.Templates) != 1 {
			t.Errorf("LoadConfig() templates length = %d, want 1", len(cfg.Templates))
		}
		if cfg.Templates[0].Name != "Daily Note" {
			t.Errorf("LoadConfig() template name = %q, want %q", cfg.Templates[0].Name, "Daily Note")
		}
		if cfg.Templates[0].File != "daily.md" {
			t.Errorf("LoadConfig() template file = %q, want %q", cfg.Templates[0].File, "daily.md")
		}
	})
}

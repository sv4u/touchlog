package entry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	// Use a fixed time for testing
	testTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		t        time.Time
		tz       *time.Location
		expected string
	}{
		{
			name:     "UTC timezone",
			t:        testTime,
			tz:       time.UTC,
			expected: "2025-01-15",
		},
		{
			name:     "nil timezone (system default)",
			t:        testTime,
			tz:       nil,
			expected: "2025-01-15", // UTC time formatted without conversion
		},
		{
			name:     "America/Denver timezone",
			t:        testTime,
			tz:       mustLoadLocation("America/Denver"),
			expected: "2025-01-15", // Same date, different time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.t, tt.tz)
			if got != tt.expected {
				t.Errorf("FormatDate() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFindAvailableFilename(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		basePath  string
		create    []string // Files to create before calling FindAvailableFilename
		want      string
		wantError bool
	}{
		{
			name:     "file does not exist",
			basePath: filepath.Join(tmpDir, "test.md"),
			create:   []string{},
			want:     filepath.Join(tmpDir, "test.md"),
		},
		{
			name:     "file exists, should use _1 suffix",
			basePath: filepath.Join(tmpDir, "test.md"),
			create:   []string{"test.md"},
			want:     filepath.Join(tmpDir, "test_1.md"),
		},
		{
			name:     "file and _1 exist, should use _2 suffix",
			basePath: filepath.Join(tmpDir, "test.md"),
			create:   []string{"test.md", "test_1.md"},
			want:     filepath.Join(tmpDir, "test_2.md"),
		},
		{
			name:     "multiple existing files",
			basePath: filepath.Join(tmpDir, "test.md"),
			create:   []string{"test.md", "test_1.md", "test_2.md"},
			want:     filepath.Join(tmpDir, "test_3.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create existing files
			for _, filename := range tt.create {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filePath, err)
				}
			}

			got, err := FindAvailableFilename(tt.basePath)
			if (err != nil) != tt.wantError {
				t.Errorf("FindAvailableFilename() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("FindAvailableFilename() = %v, want %v", got, tt.want)
			}

			// Clean up created files
			for _, filename := range tt.create {
				filePath := filepath.Join(tmpDir, filename)
				_ = os.Remove(filePath)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	tz := time.UTC

	tests := []struct {
		name     string
		title    string
		message  string
		expected string // Expected filename (without directory)
	}{
		{
			name:     "simple title",
			title:    "Standup Notes",
			message:  "Content here",
			expected: "2025-01-15_standup-notes.md",
		},
		{
			name:     "empty title, use message",
			title:    "",
			message:  "Quick note",
			expected: "2025-01-15_quick-note.md",
		},
		{
			name:     "empty title and message",
			title:    "",
			message:  "",
			expected: "2025-01-15_untitled.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFilename(tmpDir, tt.title, tt.message, testTime, tz)
			if err != nil {
				t.Fatalf("GenerateFilename() error = %v", err)
			}

			expectedPath := filepath.Join(tmpDir, tt.expected)
			if got != expectedPath {
				t.Errorf("GenerateFilename() = %v, want %v", got, expectedPath)
			}
		})
	}
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}


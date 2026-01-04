package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     []string
	}{
		{"empty string", "", nil},
		{"single tag", "work", []string{"work"}},
		{"multiple tags", "work, meeting, important", []string{"work", "meeting", "important"}},
		{"tags with spaces", "work , meeting , important", []string{"work", "meeting", "important"}},
		{"tags with extra commas", "work,,meeting", []string{"work", "meeting"}},
		{"single tag with spaces", "  work  ", []string{"work"}},
		{"empty tags filtered", "work, , meeting", []string{"work", "meeting"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTags(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("ParseTags(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("ParseTags(%q)[%d] = %q, want %q", tt.input, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestGetAvailableTemplates(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	
	// Add some inline templates
	cfg.InlineTemplates = map[string]string{
		"inline1": "# Template 1\n{{message}}",
		"inline2": "# Template 2\n{{title}}",
	}
	
	// Add some file-based templates
	cfg.Templates = []config.Template{
		{Name: "File Template 1", File: "file1.md"},
		{Name: "File Template 2", File: "file2.md"},
	}
	
	templates := GetAvailableTemplates(cfg)
	
	// Should have inline templates first, then file-based
	if len(templates) < 2 {
		t.Fatalf("GetAvailableTemplates() returned %d templates, want at least 2", len(templates))
	}
	
	// Check that inline templates are included
	hasInline1 := false
	hasInline2 := false
	for _, name := range templates {
		if name == "inline1" {
			hasInline1 = true
		}
		if name == "inline2" {
			hasInline2 = true
		}
	}
	
	if !hasInline1 || !hasInline2 {
		t.Errorf("GetAvailableTemplates() missing inline templates, got: %v", templates)
	}
}

func TestWizard_ValidateOutputDir(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{"valid existing directory", tmpDir, false},
		{"valid non-existent directory (parent exists)", filepath.Join(tmpDir, "newdir"), false},
		{"empty directory", "", true},
		{"invalid path (parent doesn't exist)", filepath.Join(tmpDir, "nonexistent", "subdir"), true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := w.ValidateOutputDir(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Wizard.ValidateOutputDir(%q) error = %v, wantErr %v", tt.dir, err, tt.wantErr)
			}
		})
	}
}

func TestWizard_ValidateOutputDir_HomeExpansion(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}
	
	// Test ~ expansion
	testDir := "~/test-notes"
	err = w.ValidateOutputDir(testDir)
	if err != nil {
		// If validation fails, it might be because the directory doesn't exist
		// Check if it's a path expansion issue
		expandedPath := filepath.Join(homeDir, "test-notes")
		if !strings.Contains(err.Error(), expandedPath) && !strings.Contains(err.Error(), "parent directory") {
			t.Errorf("Wizard.ValidateOutputDir(%q) error = %v, should handle ~ expansion", testDir, err)
		}
	}
}

func TestWizard_CreateTempFile(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	
	// Set up a template name (use default)
	cfg.DefaultTemplate = "daily"
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\nTitle: {{title}}\n\n{{message}}\n\nTags: {{tags}}",
	}
	
	w, _ := NewWizard(cfg, false)
	
	// Set entry data
	w.SetOutputDir(t.TempDir())
	w.SetTitle("Test Title")
	w.SetMessage("Test message")
	w.SetTags([]string{"test", "wizard"})
	
	// Create temp file
	err := w.CreateTempFile()
	if err != nil {
		t.Fatalf("Wizard.CreateTempFile() error = %v", err)
	}
	
	// Check that temp file was created
	tempPath := w.GetTempFilePath()
	if tempPath == "" {
		t.Error("Wizard.CreateTempFile() tempFilePath is empty")
	}
	
	// Check that file exists
	if _, err := os.Stat(tempPath); err != nil {
		t.Errorf("Temp file does not exist: %v", err)
	}
	
	// Check file content
	content := w.GetFileContent()
	if content == "" {
		t.Error("Wizard.CreateTempFile() fileContent is empty")
	}
	
	// Check that template variables were replaced
	if !strings.Contains(content, "Test Title") {
		t.Errorf("File content does not contain title. Content: %q", content)
	}
	if !strings.Contains(content, "Test message") {
		t.Errorf("File content does not contain message. Content: %q", content)
	}
	if !strings.Contains(content, "test, wizard") {
		t.Errorf("File content does not contain tags. Content: %q", content)
	}
}

func TestWizard_Cancel(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\n{{message}}",
	}
	
	w, _ := NewWizard(cfg, false)
	w.SetTitle("Test")
	w.SetMessage("Test message")
	
	// Create temp file
	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	tempPath := w.GetTempFilePath()
	if tempPath == "" {
		t.Fatal("Temp file path is empty")
	}
	
	// Verify file exists
	if _, err := os.Stat(tempPath); err != nil {
		t.Fatalf("Temp file does not exist: %v", err)
	}
	
	// Cancel (should delete temp file)
	if err := w.Cancel(); err != nil {
		t.Fatalf("Wizard.Cancel() error = %v", err)
	}
	
	// Verify file was deleted
	if _, err := os.Stat(tempPath); err == nil {
		t.Error("Temp file was not deleted after Cancel()")
	}
	
	// Cancel again should not error (file already deleted)
	if err := w.Cancel(); err != nil {
		t.Errorf("Wizard.Cancel() on already deleted file error = %v, want nil", err)
	}
}

func TestWizard_Confirm(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\n{{message}}",
	}
	
	outputDir := t.TempDir()
	w, _ := NewWizard(cfg, false)
	
	w.SetOutputDir(outputDir)
	w.SetTitle("Test Entry")
	w.SetMessage("Test message content")
	
	// Create temp file
	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// Modify temp file content to simulate editor changes
	tempPath := w.GetTempFilePath()
	modifiedContent := "# Modified Content\n\nThis was edited in the editor."
	if err := os.WriteFile(tempPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify temp file: %v", err)
	}
	
	// Confirm (should move temp file to final location)
	if err := w.Confirm(); err != nil {
		t.Fatalf("Wizard.Confirm() error = %v", err)
	}
	
	// Check that temp file was deleted
	if _, err := os.Stat(tempPath); err == nil {
		t.Error("Temp file was not deleted after Confirm()")
	}
	
	// Check that final file was created
	finalPath := w.GetFinalFilePath()
	if finalPath == "" {
		t.Error("Final file path is empty")
	}
	
	// Verify final file exists and has correct content
	if _, err := os.Stat(finalPath); err != nil {
		t.Errorf("Final file does not exist: %v", err)
	}
	
	content, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}
	
	if string(content) != modifiedContent {
		t.Errorf("Final file content = %q, want %q", string(content), modifiedContent)
	}
}

func TestWizard_Confirm_NoTempFile(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	w.SetOutputDir(t.TempDir())
	
	// Try to confirm without creating temp file
	err := w.Confirm()
	if err == nil {
		t.Error("Wizard.Confirm() without temp file error = nil, want error")
	}
}

func TestWizard_Confirm_NoOutputDir(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\n{{message}}",
	}
	
	w, _ := NewWizard(cfg, false)
	w.SetTitle("Test")
	w.SetMessage("Test message")
	
	// Create temp file
	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// Try to confirm without output dir
	err := w.Confirm()
	if err == nil {
		t.Error("Wizard.Confirm() without output dir error = nil, want error")
	}
}


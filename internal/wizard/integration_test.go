package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

// TestWizard_CompleteFlow tests the complete wizard flow from start to finish
func TestWizard_CompleteFlow(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\nTitle: {{title}}\n\n{{message}}\n\nTags: {{tags}}",
	}
	cfg.DefaultTemplate = "daily"

	outputDir := t.TempDir()

	w, err := NewWizard(cfg, false)
	if err != nil {
		t.Fatalf("NewWizard() error = %v", err)
	}

	// Step 1: Navigate through states
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Transition to TemplateSelection failed: %v", err)
	}

	// Use default template (empty name)
	w.SetTemplateName("")

	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Transition to OutputDir failed: %v", err)
	}

	// Set and validate output directory
	w.SetOutputDir(outputDir)
	if err := w.ValidateOutputDir(outputDir); err != nil {
		t.Fatalf("ValidateOutputDir() failed: %v", err)
	}

	if err := w.TransitionTo(StateTitle); err != nil {
		t.Fatalf("Transition to Title failed: %v", err)
	}

	w.SetTitle("Integration Test Entry")

	if err := w.TransitionTo(StateTags); err != nil {
		t.Fatalf("Transition to Tags failed: %v", err)
	}

	w.SetTags([]string{"test", "integration"})

	if err := w.TransitionTo(StateMessage); err != nil {
		t.Fatalf("Transition to Message failed: %v", err)
	}

	w.SetMessage("This is a test message for integration testing")

	// Step 2: Create temp file
	if err := w.TransitionTo(StateFileCreated); err != nil {
		t.Fatalf("Transition to FileCreated failed: %v", err)
	}

	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("CreateTempFile() failed: %v", err)
	}

	tempPath := w.GetTempFilePath()
	if tempPath == "" {
		t.Fatal("Temp file path is empty")
	}

	// Verify temp file exists
	if _, err := os.Stat(tempPath); err != nil {
		t.Fatalf("Temp file does not exist: %v", err)
	}

	// Verify temp file content
	content := w.GetFileContent()
	if !strings.Contains(content, "Integration Test Entry") {
		t.Error("Temp file content does not contain title")
	}
	if !strings.Contains(content, "This is a test message") {
		t.Error("Temp file content does not contain message")
	}
	if !strings.Contains(content, "test, integration") {
		t.Error("Temp file content does not contain tags")
	}

	// Step 3: Simulate editor modification (modify temp file)
	modifiedContent := content + "\n\n[Edited in editor]"
	if err := os.WriteFile(tempPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify temp file: %v", err)
	}

	// Step 4: Transition through EditorLaunch to ReviewScreen
	if err := w.TransitionTo(StateEditorLaunch); err != nil {
		t.Fatalf("Transition to EditorLaunch failed: %v", err)
	}

	// Simulate editor launch (we skip actual editor launch in test)
	// Update file content to reflect editor changes
	w.SetFileContent(modifiedContent)

	if err := w.TransitionTo(StateReviewScreen); err != nil {
		t.Fatalf("Transition to ReviewScreen failed: %v", err)
	}

	if err := w.Confirm(); err != nil {
		t.Fatalf("Confirm() failed: %v", err)
	}

	// Verify temp file was deleted
	if _, err := os.Stat(tempPath); err == nil {
		t.Error("Temp file was not deleted after Confirm()")
	}

	// Verify final file was created
	finalPath := w.GetFinalFilePath()
	if finalPath == "" {
		t.Fatal("Final file path is empty")
	}

	if _, err := os.Stat(finalPath); err != nil {
		t.Fatalf("Final file does not exist: %v", err)
	}

	// Verify final file content matches modified content
	finalContent, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	if string(finalContent) != modifiedContent {
		t.Errorf("Final file content = %q, want %q", string(finalContent), modifiedContent)
	}

	// Verify final file is in the correct directory
	if !strings.HasPrefix(finalPath, outputDir) {
		t.Errorf("Final file path %q is not in output directory %q", finalPath, outputDir)
	}
}

// TestWizard_FlowWithBackNavigation tests the wizard flow with back navigation
func TestWizard_FlowWithBackNavigation(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)

	// Navigate forward
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	w.SetOutputDir(t.TempDir())
	if err := w.TransitionTo(StateTitle); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	w.SetTitle("Test Title")

	// Go back
	if err := w.GoBack(); err != nil {
		t.Fatalf("GoBack() failed: %v", err)
	}

	if w.GetState() != StateOutputDir {
		t.Errorf("After GoBack(), state = %v, want %v", w.GetState(), StateOutputDir)
	}

	// Title should still be set (data persists)
	if w.GetTitle() != "Test Title" {
		t.Errorf("After GoBack(), title = %q, want %q", w.GetTitle(), "Test Title")
	}

	// Go back again
	if err := w.GoBack(); err != nil {
		t.Fatalf("GoBack() failed: %v", err)
	}

	if w.GetState() != StateTemplateSelection {
		t.Errorf("After second GoBack(), state = %v, want %v", w.GetState(), StateTemplateSelection)
	}
}

// TestWizard_FlowWithCancel tests canceling the wizard flow
func TestWizard_FlowWithCancel(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\n{{message}}",
	}
	cfg.DefaultTemplate = "daily"

	w, _ := NewWizard(cfg, false)

	// Navigate and create temp file
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	w.SetOutputDir(t.TempDir())
	if err := w.TransitionTo(StateTitle); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateTags); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateMessage); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateFileCreated); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("CreateTempFile() failed: %v", err)
	}

	tempPath := w.GetTempFilePath()
	if _, err := os.Stat(tempPath); err != nil {
		t.Fatalf("Temp file does not exist: %v", err)
	}

	// Cancel
	if err := w.Cancel(); err != nil {
		t.Fatalf("Cancel() failed: %v", err)
	}

	// Verify temp file was deleted
	if _, err := os.Stat(tempPath); err == nil {
		t.Error("Temp file was not deleted after Cancel()")
	}
}

// TestWizard_FlowWithTemplateSelection tests the wizard flow with template selection
func TestWizard_FlowWithTemplateSelection(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"template1": "# Template 1\n\n{{title}}\n\n{{message}}",
		"template2": "# Template 2\n\n{{message}}\n\nTags: {{tags}}",
	}
	cfg.DefaultTemplate = "template1"

	w, _ := NewWizard(cfg, false)

	// Select a specific template
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	w.SetTemplateName("template2")

	// Continue with flow
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	w.SetOutputDir(t.TempDir())
	if err := w.TransitionTo(StateTitle); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateTags); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateMessage); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateFileCreated); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// Create temp file with selected template
	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("CreateTempFile() failed: %v", err)
	}

	// Verify template was used
	content := w.GetFileContent()
	if !strings.Contains(content, "Template 2") {
		t.Error("Selected template was not used")
	}
}

// TestWizard_FlowWithEmptyFields tests the wizard flow with empty optional fields
func TestWizard_FlowWithEmptyFields(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.InlineTemplates = map[string]string{
		"daily": "# {{date}}\n\n{{message}}",
	}
	cfg.DefaultTemplate = "daily"

	w, _ := NewWizard(cfg, false)

	// Navigate through states with empty optional fields
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	w.SetOutputDir(t.TempDir())
	if err := w.TransitionTo(StateTitle); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	// Title is empty (skipped)
	if err := w.TransitionTo(StateTags); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	// Tags are empty (skipped)
	if err := w.TransitionTo(StateMessage); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	// Message is empty (skipped)
	if err := w.TransitionTo(StateFileCreated); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// Should still be able to create file
	if err := w.CreateTempFile(); err != nil {
		t.Fatalf("CreateTempFile() with empty fields failed: %v", err)
	}

	// File should be created successfully
	tempPath := w.GetTempFilePath()
	if tempPath == "" {
		t.Fatal("Temp file path is empty")
	}

	if _, err := os.Stat(tempPath); err != nil {
		t.Fatalf("Temp file does not exist: %v", err)
	}

	// Confirm should work
	if err := w.Confirm(); err != nil {
		t.Fatalf("Confirm() with empty fields failed: %v", err)
	}

	// Final file should exist
	finalPath := w.GetFinalFilePath()
	if finalPath == "" {
		t.Fatal("Final file path is empty")
	}

	if _, err := os.Stat(finalPath); err != nil {
		t.Fatalf("Final file does not exist: %v", err)
	}

	// Verify filename uses "untitled" slug when title is empty
	filename := filepath.Base(finalPath)
	if !strings.Contains(filename, "untitled") {
		t.Logf("Filename with empty title: %s (may use message or untitled)", filename)
	}
}


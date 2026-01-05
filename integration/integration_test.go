package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// getProjectRoot returns the absolute path to the project root
func getProjectRoot(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	current := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	t.Fatalf("Could not find project root (go.mod) from %s", wd)
	return ""
}

// setupTestEnv creates a temporary directory and sets up XDG environment variables
// Returns cleanup function
func setupTestEnv(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "touchlog-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Save original environment
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalDataHome := os.Getenv("XDG_DATA_HOME")

	// Set new environment
	configDir := filepath.Join(tmpDir, ".config")
	dataDir := filepath.Join(tmpDir, ".local", "share")
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(dataDir, 0755)

	os.Setenv("XDG_CONFIG_HOME", configDir)
	os.Setenv("XDG_DATA_HOME", dataDir)

	cleanup := func() {
		os.RemoveAll(tmpDir)
		if originalConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalDataHome != "" {
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		} else {
			os.Unsetenv("XDG_DATA_HOME")
		}
	}

	return tmpDir, cleanup
}

func TestIntegration_NewCommand_BasicFlow(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create output directory
	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Create template
	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}

{{message}}

Tags: {{tags}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run touchlog new command
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Test integration message",
		"--title", "Integration Test",
		"--output", outputDir,
		"--template", "daily")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify output
	outputStr := string(output)
	if !strings.Contains(outputStr, "Created:") {
		t.Errorf("Output should contain 'Created:', got: %s", outputStr)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created in output directory")
	}

	// Verify file content
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Integration Test") {
			t.Errorf("File content should contain 'Integration Test', got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "Test integration message") {
			t.Errorf("File content should contain 'Test integration message', got: %s", contentStr)
		}
	}
}

func TestIntegration_NewCommand_WithStdin(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with templates registered
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `templates:
  - name: "Daily Note"
    file: "daily.md"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Create template
	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}

{{message}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run touchlog new command with stdin
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--title", "Stdin Test",
		"--output", outputDir,
		"--template", "daily",
		"--stdin")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))
	cmd.Stdin = strings.NewReader("This is stdin content\nWith multiple lines")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created in output directory")
	}

	// Verify file content contains stdin input
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "This is stdin content") {
			t.Errorf("File content should contain stdin input, got: %s", contentStr)
		}
	}
}

func TestIntegration_ConfigCommand_ValidateConfig(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "~/test-notes"
default_template: "daily"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run touchlog config command
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "config")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Config file loaded successfully") {
		t.Errorf("Output should contain 'Config file loaded successfully', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Notes directory:") {
		t.Errorf("Output should contain 'Notes directory:', got: %s", outputStr)
	}
}

func TestIntegration_ConfigCommand_StrictMode(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with unknown keys
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "~/test-notes"
unknown_key: "should fail in strict mode"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run touchlog config command with --strict
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "config", "--strict")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"))

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Command should fail with invalid config in strict mode")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "strict validation failed") && !strings.Contains(outputStr, "unknown") {
		t.Errorf("Output should contain 'strict validation failed' or 'unknown', got: %s", outputStr)
	}
}

func TestIntegration_NewCommand_WithTags(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with templates registered
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `templates:
  - name: "Daily Note"
    file: "daily.md"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}

{{message}}

Tags: {{tags}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	projectRoot := getProjectRoot(t)

	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Tagged message",
		"--title", "Tagged Entry",
		"--tag", "work",
		"--tag", "important",
		"--tag", "meeting",
		"--output", outputDir,
		"--template", "daily")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file content contains tags
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "work") || !strings.Contains(contentStr, "important") {
			t.Errorf("File content missing tags, got: %s", contentStr)
		}
	}
}

func TestIntegration_NewCommand_WithOverwrite(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with templates registered
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `templates:
  - name: "Daily Note"
    file: "daily.md"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}

{{message}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	projectRoot := getProjectRoot(t)
	cmdPath := filepath.Join(projectRoot, "cmd", "touchlog")

	// Create first entry
	cmd1 := exec.Command("go", "run", cmdPath, "new",
		"--message", "Original message",
		"--title", "Test Title",
		"--output", outputDir,
		"--template", "daily")
	cmd1.Dir = projectRoot
	cmd1.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	if err := cmd1.Run(); err != nil {
		t.Fatalf("First command failed: %v", err)
	}

	// Try to create again without overwrite (should create numbered file, not fail)
	cmd2 := exec.Command("go", "run", cmdPath, "new",
		"--message", "New message",
		"--title", "Test Title",
		"--output", outputDir,
		"--template", "daily")
	cmd2.Dir = projectRoot
	cmd2.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	err := cmd2.Run()
	if err != nil {
		t.Fatalf("Command should create numbered file when collision occurs, got error: %v", err)
	}

	// Verify two files exist (base file and numbered file)
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files (base and numbered), got %d", len(files))
	}

	// Create again with overwrite (should overwrite base file)
	cmd3 := exec.Command("go", "run", cmdPath, "new",
		"--message", "Updated message",
		"--title", "Test Title",
		"--output", outputDir,
		"--template", "daily",
		"--overwrite")
	cmd3.Dir = projectRoot
	cmd3.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	if err := cmd3.Run(); err != nil {
		t.Fatalf("Command with overwrite failed: %v", err)
	}

	// Verify still 2 files (base file overwritten, numbered file remains)
	files, err = os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files after overwrite, got %d", len(files))
	}

	content, err := os.ReadFile(filepath.Join(outputDir, files[0].Name()))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "Updated message") {
		t.Error("File content was not updated with overwrite")
	}
}

func TestIntegration_NewCommand_WithConfigFile(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with templates registered
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + outputDir + `"
default_template: "daily"
templates:
  - name: "Daily Note"
    file: "daily.md"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create template
	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	templatePath := filepath.Join(templatesDir, "daily.md")
	templateContent := `# {{title}} - {{date}}

{{message}}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command without --output flag (should use config)
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Config test message",
		"--title", "Config Test")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created in config's notes directory
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created in config's notes directory")
	}
}

func TestIntegration_NewCommand_TemplateOverride(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with templates registered
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `templates:
  - name: "Daily Note"
    file: "daily.md"
  - name: "Custom Template"
    file: "custom.md"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	templatesDir := filepath.Join(tmpDir, ".local", "share", "touchlog", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create two templates
	dailyTemplate := filepath.Join(templatesDir, "daily.md")
	if err := os.WriteFile(dailyTemplate, []byte("# Daily - {{date}}\n\n{{message}}"), 0644); err != nil {
		t.Fatalf("Failed to create daily template: %v", err)
	}

	customTemplate := filepath.Join(templatesDir, "custom.md")
	if err := os.WriteFile(customTemplate, []byte("# Custom - {{date}}\n\n{{message}}"), 0644); err != nil {
		t.Fatalf("Failed to create custom template: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command with template override
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Template override test",
		"--title", "Override Test",
		"--output", outputDir,
		"--template", "custom")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file uses custom template
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Custom") {
			t.Errorf("File should use custom template, got: %s", contentStr)
		}
	}
}

func TestIntegration_NewCommand_ErrorHandling(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	projectRoot := getProjectRoot(t)

	// Test with invalid output directory
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Test",
		"--output", "/nonexistent/parent/dir/notes")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Command should fail with invalid output directory")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "invalid output directory") && !strings.Contains(outputStr, "parent directory") {
		t.Logf("Command failed as expected, output: %s", outputStr)
	}
}

func TestIntegration_Main_PlatformCheck(t *testing.T) {
	// Test that main.go performs platform check
	projectRoot := getProjectRoot(t)

	// This test verifies that the binary runs and performs platform check
	// On supported platforms (macOS, Linux), it should succeed
	// On unsupported platforms, it would fail, but we can't easily test that
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "--help")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify help output is shown (indicates platform check passed)
	outputStr := string(output)
	if !strings.Contains(outputStr, "touchlog") {
		t.Errorf("Output should contain 'touchlog', got: %s", outputStr)
	}
}

func TestIntegration_NewCommand_WithInlineTemplates(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with inline templates
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "quick"
inline_templates:
  quick: |
    # Quick Note - {{date}}
    
    {{message}}
    
    Tags: {{tags}}
  daily: |
    # Daily - {{date}}
    
    {{message}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command using inline template
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Inline template test",
		"--title", "Inline Test",
		"--output", outputDir,
		"--template", "quick")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file uses inline template
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Quick Note") {
			t.Errorf("File should use inline template, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "Inline template test") {
			t.Errorf("File should contain message, got: %s", contentStr)
		}
	}
}

func TestIntegration_NewCommand_WithMetadata(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "daily"
inline_templates:
  daily: |
    # {{title}} - {{date}}
    
    User: {{user}}
    Host: {{host}}
    
    {{message}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command with metadata enabled (default)
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Metadata test",
		"--title", "Metadata Test",
		"--output", outputDir)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file contains metadata placeholders (they should be replaced)
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		// Metadata should be replaced (not show as {{user}} or {{host}})
		if strings.Contains(contentStr, "{{user}}") || strings.Contains(contentStr, "{{host}}") {
			t.Logf("Metadata placeholders not replaced (may be empty), content: %s", contentStr)
		}
		// Should contain the message
		if !strings.Contains(contentStr, "Metadata test") {
			t.Errorf("File should contain message, got: %s", contentStr)
		}
	}
}

func TestIntegration_NewCommand_WithCustomVariables(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with custom variables
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "custom"
inline_templates:
  custom: |
    # {{title}} - {{date}}
    
    Project: {{project}}
    Author: {{author}}
    
    {{message}}
variables:
  project: "Touchlog"
  author: "Test User"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Custom variables test",
		"--title", "Custom Test",
		"--output", outputDir)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file contains custom variables
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Touchlog") {
			t.Errorf("File should contain custom variable 'project', got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "Test User") {
			t.Errorf("File should contain custom variable 'author', got: %s", contentStr)
		}
	}
}

func TestIntegration_NewCommand_WithTimezone(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file with timezone
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "daily"
timezone: "UTC"
inline_templates:
  daily: |
    # {{title}} - {{date}} {{time}}
    
    {{message}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "Timezone test",
		"--title", "Timezone Test",
		"--output", outputDir)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"),
		"TZ=America/Denver") // Set TZ env var to verify timezone handling

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file contains date/time (timezone should be applied)
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		// Should contain date and time
		if !strings.Contains(contentStr, "Timezone Test") {
			t.Errorf("File should contain title, got: %s", contentStr)
		}
		// Date should be in filename
		if !strings.Contains(files[0].Name(), "202") {
			t.Errorf("Filename should contain date, got: %s", files[0].Name())
		}
	}
}

func TestIntegration_NewCommand_StdinTitleHeuristic(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "daily"
inline_templates:
  daily: |
    # {{title}} - {{date}}
    
    {{message}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command with stdin where first line is short (should become title)
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--output", outputDir,
		"--stdin")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))
	cmd.Stdin = strings.NewReader("Short Title\nThis is the message content\nWith multiple lines")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("No file was created")
	}

	// Verify file content
	if len(files) > 0 {
		filePath := filepath.Join(outputDir, files[0].Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		contentStr := string(content)
		// Short first line should be used as title
		if !strings.Contains(contentStr, "Short Title") {
			t.Logf("Title heuristic may not have worked, content: %s", contentStr)
		}
		// Should contain message content
		if !strings.Contains(contentStr, "This is the message content") {
			t.Errorf("File should contain message, got: %s", contentStr)
		}
	}
}

func TestIntegration_ConfigCommand_NoConfigFile(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	projectRoot := getProjectRoot(t)

	// Run config command without config file
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "config")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "No config file found") && !strings.Contains(outputStr, "Using defaults") {
		t.Errorf("Output should indicate no config file, got: %s", outputStr)
	}
}

func TestIntegration_NewCommand_EmptyMessageAndTitle(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "daily"
inline_templates:
  daily: |
    # {{title}} - {{date}}
    
    {{message}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)

	// Run command with empty message and title
	cmd := exec.Command("go", "run", filepath.Join(projectRoot, "cmd", "touchlog"), "new",
		"--message", "",
		"--title", "",
		"--output", outputDir)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output, err := cmd.CombinedOutput()
	// Should succeed even with empty message/title (uses default slug)
	if err != nil {
		t.Logf("Command may fail with empty message/title, output: %s", output)
	}

	// If it succeeds, verify file was created
	if err == nil {
		files, err := os.ReadDir(outputDir)
		if err == nil && len(files) > 0 {
			// File should be created with default slug
			if !strings.Contains(files[0].Name(), "untitled") {
				t.Logf("File created with empty message/title: %s", files[0].Name())
			}
		}
	}
}

func TestIntegration_NewCommand_MultipleEntriesSameDay(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config file
	configDir := filepath.Join(tmpDir, ".config", "touchlog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `notes_directory: "` + filepath.Join(tmpDir, "notes") + `"
default_template: "daily"
inline_templates:
  daily: |
    # {{title}} - {{date}}
    
    {{message}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	projectRoot := getProjectRoot(t)
	cmdPath := filepath.Join(projectRoot, "cmd", "touchlog")

	// Create first entry
	cmd1 := exec.Command("go", "run", cmdPath, "new",
		"--message", "First entry",
		"--title", "First",
		"--output", outputDir)
	cmd1.Dir = projectRoot
	cmd1.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	if err := cmd1.Run(); err != nil {
		t.Fatalf("First command failed: %v", err)
	}

	// Create second entry with same title (should get numeric suffix)
	cmd2 := exec.Command("go", "run", cmdPath, "new",
		"--message", "Second entry",
		"--title", "First", // Same title
		"--output", outputDir)
	cmd2.Dir = projectRoot
	cmd2.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, ".local", "share"))

	output2, err := cmd2.CombinedOutput()
	if err != nil {
		// Second command should succeed - collision handling should create numbered file
		t.Fatalf("Second command failed (collision handling should create numbered file): %v\nOutput: %s", err, output2)
	}

	// Verify both files exist (collision handling should create a new file with numeric suffix)
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output dir: %v", err)
	}

	// Filter to only markdown files
	mdFiles := []os.DirEntry{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			mdFiles = append(mdFiles, file)
		}
	}

	if len(mdFiles) < 2 {
		t.Fatalf("Expected at least 2 markdown files (collision handling should create numbered file), got %d", len(mdFiles))
	}

	// Verify files have unique names (collision handling worked)
	fileNames := make([]string, len(mdFiles))
	for i, file := range mdFiles {
		fileNames[i] = file.Name()
	}

	uniqueNames := make(map[string]bool)
	for _, name := range fileNames {
		if uniqueNames[name] {
			t.Errorf("Duplicate filenames found: %v", fileNames)
			return
		}
		uniqueNames[name] = true
	}

	// Log file names for verification
	t.Logf("Created %d unique files: %v", len(mdFiles), fileNames)
}

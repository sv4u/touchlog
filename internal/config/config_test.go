package config

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/model"
)

func TestLoadConfig_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != model.ConfigSchemaVersion {
		t.Errorf("expected Version to be %d, got %d", model.ConfigSchemaVersion, cfg.Version)
	}
}

func TestLoadConfig_WithRepoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    description: "A note"
    default_state: "draft"
    key_pattern: "^[a-z0-9]+(-[a-z0-9]+)*$"
    key_max_len: 64
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type to be loaded")
	}

	if noteType.Description != "A note" {
		t.Errorf("expected Description to be 'A note', got %q", noteType.Description)
	}

	if noteType.DefaultState != "draft" {
		t.Errorf("expected DefaultState to be 'draft', got %q", noteType.DefaultState)
	}
}

func TestLoadConfig_RejectsUnknownKeys(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
unknown_key: "should fail"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with unknown key")
	}

	if err.Error() != "loading repo config: unknown top-level key: \"unknown_key\"" {
		t.Errorf("expected error about unknown key, got: %v", err)
	}
}

func TestLoadConfig_RejectsMissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `types:
  note:
    default_state: "draft"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with missing version")
	}

	if err.Error() != "loading repo config: missing required field: version" {
		t.Errorf("expected error about missing version, got: %v", err)
	}
}

func TestLoadConfig_RejectsUnsupportedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 999
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with unsupported version")
	}
}

func TestLoadConfig_TypeDefDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: "draft"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type to be loaded")
	}

	if noteType.KeyPattern == nil {
		t.Fatal("expected KeyPattern to have default value")
	}

	if !noteType.KeyPattern.MatchString("test-key") {
		t.Error("expected default KeyPattern to match 'test-key'")
	}

	if noteType.KeyMaxLen != DefaultKeyMaxLen {
		t.Errorf("expected KeyMaxLen to be %d, got %d", DefaultKeyMaxLen, noteType.KeyMaxLen)
	}
}

func TestLoadConfig_RejectsEmptyDefaultState(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: ""
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with empty default_state")
	}
}

// TestLoadConfig_WithEdges tests edge type configuration
func TestLoadConfig_WithEdges(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: draft
edges:
  references:
    description: Reference relationship
    allowed_from: [note]
    allowed_to: [note, article]
  related-to:
    description: General relationship
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	refEdge, ok := cfg.Edges["references"]
	if !ok {
		t.Fatal("expected 'references' edge to be loaded")
	}

	if refEdge.Description != "Reference relationship" {
		t.Errorf("expected Description to be 'Reference relationship', got %q", refEdge.Description)
	}

	if len(refEdge.AllowedFrom) != 1 || refEdge.AllowedFrom[0] != "note" {
		t.Errorf("expected AllowedFrom to be [note], got %v", refEdge.AllowedFrom)
	}

	if len(refEdge.AllowedTo) != 2 {
		t.Errorf("expected AllowedTo to have 2 items, got %d", len(refEdge.AllowedTo))
	}
}

// TestLoadConfig_WithTags tests tag configuration
func TestLoadConfig_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: draft
tags:
  preferred: [important, todo, reference]
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Tags.Preferred) != 3 {
		t.Errorf("expected 3 preferred tags, got %d", len(cfg.Tags.Preferred))
	}

	expectedTags := []string{"important", "todo", "reference"}
	for i, tag := range cfg.Tags.Preferred {
		if tag != expectedTags[i] {
			t.Errorf("expected tag[%d] = %q, got %q", i, expectedTags[i], tag)
		}
	}
}

// TestLoadConfig_WithTemplates tests template configuration
func TestLoadConfig_WithTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: draft
templates:
  root: custom-templates
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Templates.Root != "custom-templates" {
		t.Errorf("expected Templates.Root to be 'custom-templates', got %q", cfg.Templates.Root)
	}
}

// TestLoadConfig_WithRequiredFields tests required fields configuration
func TestLoadConfig_WithRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: draft
    required_fields: [title, state]
    recommended_fields: [tags]
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type to be loaded")
	}

	if len(noteType.RequiredFields) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(noteType.RequiredFields))
	}

	if noteType.RequiredFields[0] != "title" || noteType.RequiredFields[1] != "state" {
		t.Errorf("expected RequiredFields to be [title, state], got %v", noteType.RequiredFields)
	}

	if len(noteType.RecommendedFields) != 1 || noteType.RecommendedFields[0] != "tags" {
		t.Errorf("expected RecommendedFields to be [tags], got %v", noteType.RecommendedFields)
	}
}

// TestLoadConfig_WithCustomKeyPattern tests custom key pattern
func TestLoadConfig_WithCustomKeyPattern(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: draft
    key_pattern: "^[A-Z][a-z]+$"
    key_max_len: 32
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type to be loaded")
	}

	// Test that the pattern matches a valid key (single capital followed by lowercase)
	if !noteType.KeyPattern.MatchString("Test") {
		t.Error("expected custom key pattern to match 'Test'")
	}

	// Test that the pattern doesn't match invalid keys
	if noteType.KeyPattern.MatchString("TestKey") {
		t.Error("expected custom key pattern not to match 'TestKey' (has two capitals)")
	}

	if noteType.KeyMaxLen != 32 {
		t.Errorf("expected KeyMaxLen to be 32, got %d", noteType.KeyMaxLen)
	}
}

// TestLoadConfig_RejectsInvalidKeyPattern tests invalid regex pattern
func TestLoadConfig_RejectsInvalidKeyPattern(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("failed to create repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    default_state: draft
    key_pattern: "[invalid regex"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("failed to write repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with invalid key_pattern")
	}

	if !contains(err.Error(), "key_pattern") {
		t.Errorf("expected error about key_pattern, got: %v", err)
	}
}

// TestLoadConfig_WithGlobalConfig tests config merging with global config
func TestLoadConfig_WithGlobalConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	// Create temp directory for global config
	tmpHome := t.TempDir()
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("setting HOME: %v", err)
	}

	// Create global config directory
	globalConfigDir := filepath.Join(tmpHome, ".config", "touchlog")
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("creating global config dir: %v", err)
	}

	globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
	globalConfig := `version: 1
types:
  article:
    description: "Global article type"
    default_state: "draft"
tags:
  preferred: [global-tag]
`

	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("writing global config: %v", err)
	}

	// Create repo config
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("creating repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    description: "Repo note type"
    default_state: "draft"
tags:
  preferred: [repo-tag]
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("writing repo config: %v", err)
	}

	// Load config (should merge global and repo)
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Should have both types (global article + repo note)
	if _, ok := cfg.Types["article"]; !ok {
		t.Error("expected 'article' type from global config")
	}
	if _, ok := cfg.Types["note"]; !ok {
		t.Error("expected 'note' type from repo config")
	}

	// Tags should be merged (repo overrides global)
	if len(cfg.Tags.Preferred) == 0 {
		t.Error("expected preferred tags to be loaded")
	}
}

// TestLoadConfig_RepoOverridesGlobal tests that repo config overrides global
func TestLoadConfig_RepoOverridesGlobal(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	// Create temp directory for global config
	tmpHome := t.TempDir()
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("setting HOME: %v", err)
	}

	// Create global config
	globalConfigDir := filepath.Join(tmpHome, ".config", "touchlog")
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("creating global config dir: %v", err)
	}

	globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
	globalConfig := `version: 1
types:
  note:
    description: "Global note"
    default_state: "published"
`

	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("writing global config: %v", err)
	}

	// Create repo config with same type (should override)
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("creating repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	repoConfig := `version: 1
types:
  note:
    description: "Repo note"
    default_state: "draft"
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("writing repo config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Repo config should override global
	noteType, ok := cfg.Types["note"]
	if !ok {
		t.Fatal("expected 'note' type")
	}

	if noteType.Description != "Repo note" {
		t.Errorf("expected description 'Repo note' (from repo), got %q", noteType.Description)
	}

	if noteType.DefaultState != "draft" {
		t.Errorf("expected default_state 'draft' (from repo), got %q", noteType.DefaultState)
	}
}

// TestLoadConfig_ValidatesAfterMerge tests that validation happens after merging
func TestLoadConfig_ValidatesAfterMerge(t *testing.T) {
	tmpDir := t.TempDir()
	repoConfigDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("creating repo config dir: %v", err)
	}

	repoConfigPath := filepath.Join(repoConfigDir, "config.yaml")
	// Config with invalid key_max_len (must be positive)
	repoConfig := `version: 1
types:
  note:
    default_state: "draft"
    key_max_len: -1
`

	if err := os.WriteFile(repoConfigPath, []byte(repoConfig), 0644); err != nil {
		t.Fatalf("writing repo config: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("expected LoadConfig to fail with invalid key_max_len")
	}

	if !contains(err.Error(), "key_max_len") {
		t.Errorf("expected error about key_max_len, got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestLastSegment tests the LastSegment helper function
func TestLastSegment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple-key", "simple-key"},
		{"a/b/c", "c"},
		{"projects/web/auth", "auth"},
		{"single", "single"},
		{"a/b", "b"},
		{"", ""},
	}

	for _, tt := range tests {
		result := LastSegment(tt.input)
		if result != tt.expected {
			t.Errorf("LastSegment(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// TestValidateKey tests the ValidateKey function
func TestValidateKey(t *testing.T) {
	defaultPattern := DefaultKeyPattern
	pattern := compilePattern(defaultPattern)
	maxLen := DefaultKeyMaxLen

	tests := []struct {
		key       string
		expectErr bool
		errMsg    string
	}{
		// Valid flat keys
		{"simple-key", false, ""},
		{"test", false, ""},
		{"a-b-c", false, ""},

		// Valid path-based keys
		{"projects/web", false, ""},
		{"a/b/c", false, ""},
		{"projects/web/auth", false, ""},

		// Invalid: empty
		{"", true, "cannot be empty"},

		// Invalid: leading slash
		{"/a/b", true, "cannot start or end with /"},

		// Invalid: trailing slash
		{"a/b/", true, "cannot start or end with /"},

		// Invalid: empty segment (consecutive slashes)
		{"a//b", true, "empty segments"},

		// Invalid: segment doesn't match pattern
		{"A/b/c", true, "does not match pattern"},
		{"projects/Web/auth", true, "does not match pattern"},

		// Invalid: key too long
		{createLongKey(100), true, "exceeds maximum length"},
	}

	for _, tt := range tests {
		err := ValidateKey(tt.key, pattern, maxLen)
		if tt.expectErr {
			if err == nil {
				t.Errorf("ValidateKey(%q) expected error containing %q, got nil", tt.key, tt.errMsg)
			} else if !contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateKey(%q) expected error containing %q, got %q", tt.key, tt.errMsg, err.Error())
			}
		} else {
			if err != nil {
				t.Errorf("ValidateKey(%q) expected no error, got %v", tt.key, err)
			}
		}
	}
}

// TestValidateKey_AllSegmentsMustMatch tests that all segments must match the pattern
func TestValidateKey_AllSegmentsMustMatch(t *testing.T) {
	pattern := compilePattern(DefaultKeyPattern)

	// All segments valid
	err := ValidateKey("projects/web/auth", pattern, 100)
	if err != nil {
		t.Errorf("expected valid key, got error: %v", err)
	}

	// First segment invalid
	err = ValidateKey("Projects/web/auth", pattern, 100)
	if err == nil {
		t.Error("expected error for invalid first segment")
	}

	// Middle segment invalid
	err = ValidateKey("projects/Web/auth", pattern, 100)
	if err == nil {
		t.Error("expected error for invalid middle segment")
	}

	// Last segment invalid
	err = ValidateKey("projects/web/Auth", pattern, 100)
	if err == nil {
		t.Error("expected error for invalid last segment")
	}
}

// Helper to compile pattern
func compilePattern(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

// Helper to create a long key
func createLongKey(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
}

package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
)

// TestInputKey_Behavior_ValidKey tests inputKey with valid typeDef
func TestInputKey_Behavior_ValidKey(t *testing.T) {
	tmpDir := t.TempDir()
	typeDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		t.Fatalf("creating type dir: %v", err)
	}

	typeDef := config.TypeDef{
		KeyPattern: regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`),
		KeyMaxLen:  64,
	}

	key, err := inputKey(typeDef, tmpDir, "note")
	if err != nil {
		t.Fatalf("inputKey failed: %v", err)
	}

	if key == "" {
		t.Error("expected non-empty key")
	}
}

// TestInputKey_Behavior_InvalidPattern tests inputKey with invalid pattern
func TestInputKey_Behavior_InvalidPattern(t *testing.T) {
	tmpDir := t.TempDir()
	typeDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		t.Fatalf("creating type dir: %v", err)
	}

	// Create existing note to trigger pattern validation
	existingPath := filepath.Join(typeDir, "new-note.Rmd")
	if err := os.WriteFile(existingPath, []byte("test"), 0644); err != nil {
		t.Fatalf("creating existing note: %v", err)
	}

	typeDef := config.TypeDef{
		KeyPattern: regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`),
		KeyMaxLen:  64,
	}

	_, err := inputKey(typeDef, tmpDir, "note")
	if err == nil {
		t.Error("expected error when key already exists")
	}
}

// TestInputKey_Behavior_ExceedsMaxLength tests inputKey with key exceeding max length
func TestInputKey_Behavior_ExceedsMaxLength(t *testing.T) {
	tmpDir := t.TempDir()
	typeDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		t.Fatalf("creating type dir: %v", err)
	}

	// Set very short max length
	typeDef := config.TypeDef{
		KeyPattern: regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`),
		KeyMaxLen:  5, // "new-note" is 8 characters
	}

	_, err := inputKey(typeDef, tmpDir, "note")
	if err == nil {
		t.Error("expected error when key exceeds max length")
	}
}

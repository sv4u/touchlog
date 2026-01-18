package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestSelectType tests selectType function behavior
func TestSelectType(t *testing.T) {
	cfg := &config.Config{
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
			},
		},
	}

	typeName, err := selectType(cfg)
	if err != nil {
		t.Fatalf("selectType failed: %v", err)
	}

	if typeName != "note" {
		t.Errorf("expected type 'note', got %q", typeName)
	}
}

// TestSelectType_NoNoteType tests selectType when note type is not available
func TestSelectType_NoNoteType(t *testing.T) {
	cfg := &config.Config{
		Types: map[model.TypeName]config.TypeDef{
			"task": {
				Description:  "A task",
				DefaultState: "todo",
			},
		},
	}

	typeName, err := selectType(cfg)
	if err != nil {
		t.Fatalf("selectType failed: %v", err)
	}

	if typeName != "task" {
		t.Errorf("expected type 'task', got %q", typeName)
	}
}

// TestSelectType_NoTypes tests selectType when no types are available
func TestSelectType_NoTypes(t *testing.T) {
	cfg := &config.Config{
		Types: map[model.TypeName]config.TypeDef{},
	}

	_, err := selectType(cfg)
	if err == nil {
		t.Error("expected error when no types are available")
	}
}

// TestInputTitle tests inputTitle function behavior
func TestInputTitle(t *testing.T) {
	title, err := inputTitle()
	if err != nil {
		t.Fatalf("inputTitle failed: %v", err)
	}

	if title == "" {
		t.Error("expected non-empty title")
	}
}

// TestInputTags tests inputTags function behavior
func TestInputTags(t *testing.T) {
	tags, err := inputTags()
	if err != nil {
		t.Fatalf("inputTags failed: %v", err)
	}

	if tags == nil {
		t.Error("expected tags to be non-nil (empty slice)")
	}
}

// TestGenerateNoteID tests generateNoteID function behavior
func TestGenerateNoteID(t *testing.T) {
	id1 := generateNoteID()

	if id1 == "" {
		t.Error("expected non-empty note ID")
	}

	// Verify ID format starts with "note-"
	if !strings.HasPrefix(string(id1), "note-") {
		t.Errorf("expected ID to start with 'note-', got %q", id1)
	}

	// Verify ID format has reasonable length
	if len(id1) < 10 {
		t.Error("expected ID to have reasonable length")
	}
}

// TestGenerateFrontmatter tests generateFrontmatter function behavior
func TestGenerateFrontmatter(t *testing.T) {
	id := model.NoteID("note-1")
	typeName := model.TypeName("note")
	key := model.Key("test-note")
	title := "Test Note"
	tags := []string{"test", "example"}
	state := "draft"
	now := time.Now()

	fm := generateFrontmatter(id, typeName, key, title, tags, state, now)

	if fm["id"] != string(id) {
		t.Errorf("expected id %q, got %q", id, fm["id"])
	}
	if fm["type"] != string(typeName) {
		t.Errorf("expected type %q, got %q", typeName, fm["type"])
	}
	if fm["key"] != string(key) {
		t.Errorf("expected key %q, got %q", key, fm["key"])
	}
	if fm["title"] != title {
		t.Errorf("expected title %q, got %q", title, fm["title"])
	}
	if fm["state"] != state {
		t.Errorf("expected state %q, got %q", state, fm["state"])
	}
}

// TestGenerateBody tests generateBody function behavior
func TestGenerateBody(t *testing.T) {
	title := "Test Note"
	cfg := &config.Config{}
	typeName := model.TypeName("note")

	body := generateBody(title, cfg, typeName)

	if body == "" {
		t.Error("expected non-empty body")
	}

	// Verify body contains title
	if !strings.Contains(body, title) {
		t.Errorf("expected body to contain title %q", title)
	}
}

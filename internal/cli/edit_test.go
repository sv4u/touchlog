package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
	"github.com/sv4u/touchlog/v2/internal/model"
	"github.com/sv4u/touchlog/v2/internal/store"
)

func TestBuildEditCommand(t *testing.T) {
	cmd := BuildEditCommand()

	if cmd.Name != "edit" {
		t.Errorf("expected command name 'edit', got %q", cmd.Name)
	}

	if cmd.Usage == "" {
		t.Error("expected Usage to be set")
	}

	// Verify flags exist
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		for _, name := range flag.Names() {
			flagNames[name] = true
		}
	}

	expectedFlags := []string{"key", "type", "tag"}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestEditByKey_NoteFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Create a note manually
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	noteContent := `---
id: test-note-1
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test Note
tags: []
state: draft
---

# Test Note
`
	notePath := filepath.Join(noteDir, "test-key.Rmd")
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Build the index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	// Verify resolveNotePath works for qualified key
	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	path, err := resolveNotePath(db, tmpDir, "note:test-key")
	if err != nil {
		t.Fatalf("resolveNotePath failed for qualified key: %v", err)
	}
	if path != notePath {
		t.Errorf("expected path %q, got %q", notePath, path)
	}

	// Verify resolveNotePath works for unqualified key
	path, err = resolveNotePath(db, tmpDir, "test-key")
	if err != nil {
		t.Fatalf("resolveNotePath failed for unqualified key: %v", err)
	}
	if path != notePath {
		t.Errorf("expected path %q, got %q", notePath, path)
	}
}

func TestEditByKey_NoteNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Build empty index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Try to resolve a non-existent note
	_, err = resolveNotePath(db, tmpDir, "nonexistent-key")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

func TestEditByKey_AmbiguousKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Create two notes with the same last segment but different paths
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// First note with key "auth"
	note1Content := `---
id: test-note-1
type: note
key: auth
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Auth Note
tags: []
state: draft
---

# Auth Note
`
	if err := os.WriteFile(filepath.Join(noteDir, "auth.Rmd"), []byte(note1Content), 0644); err != nil {
		t.Fatalf("writing note 1: %v", err)
	}

	// Second note with key "projects/auth" (needs subdirectory)
	projectsDir := filepath.Join(noteDir, "projects", "auth")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("creating projects dir: %v", err)
	}

	note2Content := `---
id: test-note-2
type: note
key: projects/auth
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Projects Auth Note
tags: []
state: draft
---

# Projects Auth Note
`
	if err := os.WriteFile(filepath.Join(projectsDir, "auth.Rmd"), []byte(note2Content), 0644); err != nil {
		t.Fatalf("writing note 2: %v", err)
	}

	// Build the index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Qualified keys should work
	_, err = resolveNotePath(db, tmpDir, "note:auth")
	if err != nil {
		t.Errorf("expected note:auth to resolve: %v", err)
	}

	_, err = resolveNotePath(db, tmpDir, "note:projects/auth")
	if err != nil {
		t.Errorf("expected note:projects/auth to resolve: %v", err)
	}

	// Unqualified "auth" matches exactly "auth" (not ambiguous - exact match takes priority)
	path, err := resolveNotePath(db, tmpDir, "auth")
	if err != nil {
		t.Errorf("expected 'auth' to resolve to exact match: %v", err)
	}
	if path != filepath.Join(noteDir, "auth.Rmd") {
		t.Errorf("expected exact match to auth.Rmd, got %q", path)
	}
}

func TestLoadNotesForEdit_EmptyVault(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Build empty index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	notes, err := loadNotesForEdit(tmpDir, "", nil)
	if err != nil {
		t.Fatalf("loadNotesForEdit failed: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(notes))
	}
}

func TestLoadNotesForEdit_WithNotes(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Create notes
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	noteContent := `---
id: test-note-1
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test Note
tags: [important, draft]
state: draft
---

# Test Note

This is the body of the test note.
`
	if err := os.WriteFile(filepath.Join(noteDir, "test-key.Rmd"), []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	// Build index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	notes, err := loadNotesForEdit(tmpDir, "", nil)
	if err != nil {
		t.Fatalf("loadNotesForEdit failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	note := notes[0]
	if note.key != "test-key" {
		t.Errorf("expected key 'test-key', got %q", note.key)
	}
	if note.title != "Test Note" {
		t.Errorf("expected title 'Test Note', got %q", note.title)
	}
	if note.typ != "note" {
		t.Errorf("expected type 'note', got %q", note.typ)
	}
}

func TestLoadNotesForEdit_WithTypeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault with multiple types
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	configContent := `version: 1
types:
  note:
    description: General notes
    default_state: draft
  task:
    description: Tasks
    default_state: todo
`
	if err := os.WriteFile(filepath.Join(touchlogDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Create notes of different types
	noteDir := filepath.Join(tmpDir, "note")
	taskDir := filepath.Join(tmpDir, "task")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("creating task dir: %v", err)
	}

	noteContent := `---
id: note-1
type: note
key: my-note
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: My Note
tags: []
state: draft
---
`
	if err := os.WriteFile(filepath.Join(noteDir, "my-note.Rmd"), []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	taskContent := `---
id: task-1
type: task
key: my-task
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: My Task
tags: []
state: todo
---
`
	if err := os.WriteFile(filepath.Join(taskDir, "my-task.Rmd"), []byte(taskContent), 0644); err != nil {
		t.Fatalf("writing task: %v", err)
	}

	// Build index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	// Load all notes
	allNotes, err := loadNotesForEdit(tmpDir, "", nil)
	if err != nil {
		t.Fatalf("loadNotesForEdit failed: %v", err)
	}
	if len(allNotes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(allNotes))
	}

	// Load only notes
	notes, err := loadNotesForEdit(tmpDir, "note", nil)
	if err != nil {
		t.Fatalf("loadNotesForEdit with type filter failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].typ != "note" {
		t.Errorf("expected type 'note', got %q", notes[0].typ)
	}
}

func TestResolveNotePath_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	if err := runInitWizard(tmpDir); err != nil {
		t.Fatalf("runInitWizard failed: %v", err)
	}

	// Build empty index
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if err := index.NewBuilder(tmpDir, cfg).Rebuild(); err != nil {
		t.Fatalf("building index: %v", err)
	}

	db, err := store.OpenOrCreateDB(tmpDir)
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Test invalid format with too many colons
	_, err = resolveNotePath(db, tmpDir, "type:key:extra")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestNoteItem_Methods(t *testing.T) {
	item := noteItem{
		id:    "test-id",
		typ:   "note",
		key:   "test-key",
		title: "Test Title",
		tags:  []string{"tag1", "tag2"},
		path:  "note/test-key.Rmd",
		body:  "This is the body content",
	}

	// Test FilterValue
	filterVal := item.FilterValue()
	if filterVal == "" {
		t.Error("expected FilterValue to return non-empty string")
	}
	if !containsAll(filterVal, "Test Title", "test-key", "note", "tag1", "tag2", "body content") {
		t.Errorf("FilterValue should contain all searchable fields, got: %q", filterVal)
	}

	// Test Title
	title := item.Title()
	if title != "note:test-key" {
		t.Errorf("expected Title 'note:test-key', got %q", title)
	}

	// Test Description
	desc := item.Description()
	if desc != "Test Title [tag1, tag2]" {
		t.Errorf("expected Description 'Test Title [tag1, tag2]', got %q", desc)
	}

	// Test Description without tags
	item.tags = nil
	desc = item.Description()
	if desc != "Test Title" {
		t.Errorf("expected Description 'Test Title', got %q", desc)
	}
}

func TestNoteItem_DescriptionWithEmptyTags(t *testing.T) {
	item := noteItem{
		id:    "test-id",
		typ:   "note",
		key:   "test-key",
		title: "Test Title",
		tags:  []string{},
		path:  "note/test-key.Rmd",
	}

	desc := item.Description()
	if desc != "Test Title" {
		t.Errorf("expected Description 'Test Title', got %q", desc)
	}
}

// Helper function to check if a string contains all substrings
func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !containsIgnoreCase(s, sub) {
			return false
		}
	}
	return true
}

func containsIgnoreCase(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && containsIgnoreCaseHelper(s, sub)))
}

func containsIgnoreCaseHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if equalFold(s[i:i+len(sub)], sub) {
			return true
		}
	}
	return false
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		c1 := s[i]
		c2 := t[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}

func TestParseTagsFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty array",
			input:    "[]",
			expected: []string{},
		},
		{
			name:     "single tag",
			input:    `["tag1"]`,
			expected: []string{"tag1"},
		},
		{
			name:     "multiple tags",
			input:    `["tag1","tag2","tag3"]`,
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "tag with comma",
			input:    `["tag,with,comma"]`,
			expected: []string{"tag,with,comma"},
		},
		{
			name:     "mixed tags including comma",
			input:    `["normal","has,comma","another"]`,
			expected: []string{"normal", "has,comma", "another"},
		},
		{
			name:     "tag with special characters",
			input:    `["tag:colon","tag\"quote","tag\\backslash"]`,
			expected: []string{"tag:colon", "tag\"quote", "tag\\backslash"},
		},
		{
			name:     "unicode tags",
			input:    `["日本語","emoji🎉","café"]`,
			expected: []string{"日本語", "emoji🎉", "café"},
		},
		{
			name:     "invalid json",
			input:    "not valid json",
			expected: []string{},
		},
		// Bug fix: SQLite json_group_array with LEFT JOIN null handling
		{
			name:     "null from LEFT JOIN produces empty slice",
			input:    `[null]`,
			expected: []string{},
		},
		{
			name:     "mixed null and valid tags filters nulls",
			input:    `["valid",null,"another"]`,
			expected: []string{"valid", "another"},
		},
		{
			name:     "empty string filtered out",
			input:    `["","valid",""]`,
			expected: []string{"valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTagsFromJSON(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("expected tag[%d] = %q, got %q", i, tt.expected[i], tag)
				}
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short ASCII", "short", 10, "short"},
		{"exactly eleven", "exactly ten", 11, "exactly ten"},
		{"truncate ASCII", "this is a long string", 10, "this is a "},
		{"empty string", "", 10, ""},
		// UTF-8 multi-byte character tests
		{"CJK characters - no truncate", "日本語", 5, "日本語"},
		{"CJK characters - truncate", "日本語テスト", 3, "日本語"},
		{"emoji - no truncate", "hello🎉world", 15, "hello🎉world"},
		{"emoji - truncate at emoji", "hello🎉world", 6, "hello🎉"},
		{"emoji - truncate before emoji", "hello🎉world", 5, "hello"},
		{"mixed UTF-8", "café résumé", 5, "café "},
		{"accented - preserve character", "naïve", 3, "naï"},
		// Verify rune count not byte count
		{"CJK 3 runes truncated to 2", "你好世界", 2, "你好"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
			// Also verify the result is valid UTF-8
			if !isValidUTF8(result) {
				t.Errorf("truncateString(%q, %d) produced invalid UTF-8: %q", tt.input, tt.maxLen, result)
			}
		})
	}
}

// isValidUTF8 checks if a string is valid UTF-8
func isValidUTF8(s string) bool {
	for i := 0; i < len(s); {
		r, size := rune(s[i]), 1
		if r >= 0x80 {
			r, size = decodeRuneInString(s[i:])
			if r == 0xFFFD && size == 1 {
				return false // Invalid UTF-8
			}
		}
		i += size
	}
	return true
}

// decodeRuneInString decodes a single UTF-8 rune from the string
func decodeRuneInString(s string) (rune, int) {
	if len(s) == 0 {
		return 0xFFFD, 0
	}
	b := s[0]
	if b < 0x80 {
		return rune(b), 1
	}
	var size int
	var min rune
	switch {
	case b&0xE0 == 0xC0:
		size, min = 2, 0x80
	case b&0xF0 == 0xE0:
		size, min = 3, 0x800
	case b&0xF8 == 0xF0:
		size, min = 4, 0x10000
	default:
		return 0xFFFD, 1
	}
	if len(s) < size {
		return 0xFFFD, 1
	}
	var r rune
	switch size {
	case 2:
		r = rune(b&0x1F)<<6 | rune(s[1]&0x3F)
	case 3:
		r = rune(b&0x0F)<<12 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F)
	case 4:
		r = rune(b&0x07)<<18 | rune(s[1]&0x3F)<<12 | rune(s[2]&0x3F)<<6 | rune(s[3]&0x3F)
	}
	if r < min {
		return 0xFFFD, 1
	}
	return r, size
}

func TestEditWizardModel_Init(t *testing.T) {
	notes := []noteItem{
		{id: "1", typ: "note", key: "key1", title: "Title 1"},
	}

	model := initialEditModel(notes, "/tmp/vault")

	// Test Init returns nil
	cmd := model.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}

	// Test list is properly configured
	if model.list.Title != "Select a note to edit" {
		t.Errorf("expected list title 'Select a note to edit', got %q", model.list.Title)
	}
}

// Helper to create a test note in the vault
func createTestNote(t *testing.T, vaultRoot string, noteID model.NoteID, typeName model.TypeName, key model.Key, title string, tags []string) string {
	t.Helper()

	noteDir := filepath.Join(vaultRoot, string(typeName))
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tagsStr := "[]"
	if len(tags) > 0 {
		tagsStr = "[" + joinQuoted(tags) + "]"
	}

	content := fmt.Sprintf(`---
id: %s
type: %s
key: %s
created: %s
updated: %s
title: %s
tags: %s
state: draft
---

# %s
`, noteID, typeName, key, now, now, title, tagsStr, title)

	notePath := filepath.Join(noteDir, string(key)+".Rmd")
	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		t.Fatalf("writing note: %v", err)
	}

	return notePath
}

func joinQuoted(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := `"` + strs[0] + `"`
	for i := 1; i < len(strs); i++ {
		result += `, "` + strs[i] + `"`
	}
	return result
}

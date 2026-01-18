package note

import (
	"strings"
	"testing"
	"time"

	"github.com/sv4u/touchlog/v2/internal/model"
)

func TestParse_ValidFrontmatter(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test Title
tags:
  - tag1
  - tag2
state: draft
---
This is the body.
`

	note := Parse("test.md", []byte(content))

	if note.FM.ID != "test-id" {
		t.Errorf("expected ID to be 'test-id', got %q", note.FM.ID)
	}
	if note.FM.Type != "note" {
		t.Errorf("expected Type to be 'note', got %q", note.FM.Type)
	}
	if note.FM.Key != "test-key" {
		t.Errorf("expected Key to be 'test-key', got %q", note.FM.Key)
	}
	if note.FM.Title != "Test Title" {
		t.Errorf("expected Title to be 'Test Title', got %q", note.FM.Title)
	}
	if len(note.FM.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(note.FM.Tags))
	}
	if note.FM.State != "draft" {
		t.Errorf("expected State to be 'draft', got %q", note.FM.State)
	}
	if !strings.Contains(note.Body, "This is the body") {
		t.Errorf("expected body to contain 'This is the body', got %q", note.Body)
	}
	if len(note.Diags) > 0 {
		t.Errorf("expected no diagnostics, got %d", len(note.Diags))
	}
}

func TestParse_MissingFrontmatter(t *testing.T) {
	content := `This is just body text with no frontmatter.`

	note := Parse("test.md", []byte(content))

	if len(note.Diags) == 0 {
		t.Fatal("expected diagnostics for missing frontmatter")
	}

	diag := note.Diags[0]
	if diag.Level != model.DiagnosticLevelError {
		t.Errorf("expected diagnostic level to be 'error', got %q", diag.Level)
	}
	if diag.Code != "FRONTMATTER_MISSING" {
		t.Errorf("expected diagnostic code to be 'FRONTMATTER_MISSING', got %q", diag.Code)
	}
	if note.Body != content {
		t.Errorf("expected body to be preserved even with missing frontmatter")
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	content := `---
id: test-id
type: note
invalid: [unclosed bracket
---
Body
`

	note := Parse("test.md", []byte(content))

	if len(note.Diags) == 0 {
		t.Fatal("expected diagnostics for invalid YAML")
	}

	diag := note.Diags[0]
	if diag.Level != model.DiagnosticLevelError {
		t.Errorf("expected diagnostic level to be 'error', got %q", diag.Level)
	}
	if diag.Code != "FRONTMATTER_PARSE_ERROR" {
		t.Errorf("expected diagnostic code to be 'FRONTMATTER_PARSE_ERROR', got %q", diag.Code)
	}
}

func TestParse_ExtraFields(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test Title
tags: []
state: draft
custom_field: custom_value
another_field: 42
---
Body
`

	note := Parse("test.md", []byte(content))

	if note.FM.Extra["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field to be preserved in Extra")
	}
	if note.FM.Extra["another_field"] != 42 {
		t.Errorf("expected another_field to be preserved in Extra")
	}
}

func TestParse_LinkExtraction_TypeKey(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test
tags: []
state: draft
---
This links to [[note:other-key]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}

	link := note.RawLinks[0]
	if link.Target.Type == nil {
		t.Error("expected link target to have Type")
	}
	if *link.Target.Type != "note" {
		t.Errorf("expected link target Type to be 'note', got %q", *link.Target.Type)
	}
	if link.Target.Key != "other-key" {
		t.Errorf("expected link target Key to be 'other-key', got %q", link.Target.Key)
	}
	if link.EdgeType != model.DefaultEdgeType {
		t.Errorf("expected EdgeType to be default, got %q", link.EdgeType)
	}
}

func TestParse_LinkExtraction_UnqualifiedKey(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test
tags: []
state: draft
---
This links to [[other-key]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}

	link := note.RawLinks[0]
	if link.Target.Type != nil {
		t.Error("expected link target Type to be nil for unqualified key")
	}
	if link.Target.Key != "other-key" {
		t.Errorf("expected link target Key to be 'other-key', got %q", link.Target.Key)
	}
}

func TestParse_LinkExtraction_WithEdgeType(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test
tags: []
state: draft
---
This links to [[note:other-key|depends-on]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}

	link := note.RawLinks[0]
	if link.EdgeType != "depends-on" {
		t.Errorf("expected EdgeType to be 'depends-on', got %q", link.EdgeType)
	}
}

func TestParse_LinkExtraction_MultipleLinks(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test
tags: []
state: draft
---
This links to [[note:first]] and [[second]] and [[log:third|related-to]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 3 {
		t.Fatalf("expected 3 links, got %d", len(note.RawLinks))
	}

	// Check first link
	if note.RawLinks[0].Target.Key != "first" {
		t.Errorf("expected first link key to be 'first', got %q", note.RawLinks[0].Target.Key)
	}

	// Check second link (unqualified)
	if note.RawLinks[1].Target.Type != nil {
		t.Error("expected second link to be unqualified")
	}
	if note.RawLinks[1].Target.Key != "second" {
		t.Errorf("expected second link key to be 'second', got %q", note.RawLinks[1].Target.Key)
	}

	// Check third link
	if note.RawLinks[2].Target.Key != "third" {
		t.Errorf("expected third link key to be 'third', got %q", note.RawLinks[2].Target.Key)
	}
}

func TestParse_LinkExtraction_UnqualifiedWithEdgeType(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test
tags: []
state: draft
---
This links to [[other-key|depends-on]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}

	link := note.RawLinks[0]
	if link.Target.Type != nil {
		t.Error("expected link target Type to be nil for unqualified key")
	}
	if link.Target.Key != "other-key" {
		t.Errorf("expected link target Key to be 'other-key', got %q", link.Target.Key)
	}
	if link.EdgeType != "depends-on" {
		t.Errorf("expected EdgeType to be 'depends-on', got %q", link.EdgeType)
	}
}

func TestParse_TimeParsing(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-15T10:30:00Z
updated: 2024-01-15T11:45:00Z
title: Test
tags: []
state: draft
---
Body
`

	note := Parse("test.md", []byte(content))

	expectedCreated := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !note.FM.Created.Equal(expectedCreated) {
		t.Errorf("expected Created to be %v, got %v", expectedCreated, note.FM.Created)
	}

	expectedUpdated := time.Date(2024, 1, 15, 11, 45, 0, 0, time.UTC)
	if !note.FM.Updated.Equal(expectedUpdated) {
		t.Errorf("expected Updated to be %v, got %v", expectedUpdated, note.FM.Updated)
	}
}

func TestParse_LinkSpanTracking(t *testing.T) {
	content := `---
id: test-id
type: note
key: test-key
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
title: Test
tags: []
state: draft
---
This links to [[note:other-key]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}

	link := note.RawLinks[0]
	if link.Span.Path != "test.md" {
		t.Errorf("expected Span.Path to be 'test.md', got %q", link.Span.Path)
	}
	if link.Span.StartByte >= link.Span.EndByte {
		t.Errorf("expected StartByte < EndByte, got %d >= %d", link.Span.StartByte, link.Span.EndByte)
	}
}

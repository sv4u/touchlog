package note

import (
	"strings"
	"testing"
)

func TestParse_EmptyFrontmatter(t *testing.T) {
	content := `---
---
Body text
`

	note := Parse("test.md", []byte(content))

	if len(note.Diags) > 0 {
		// Empty frontmatter is valid YAML, should not error
		t.Logf("Note: empty frontmatter produced diagnostics: %v", note.Diags)
	}
	if !contains(note.Body, "Body text") {
		t.Errorf("expected body to be preserved")
	}
}

func TestParse_NoBody(t *testing.T) {
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
`

	note := Parse("test.md", []byte(content))

	// Body may contain just a newline after closing ---
	bodyTrimmed := strings.TrimSpace(note.Body)
	if bodyTrimmed != "" {
		t.Errorf("expected empty body, got %q", note.Body)
	}
}

func TestParse_LinkAtStartOfBody(t *testing.T) {
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
[[note:first]] is at the start.
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}
	if note.RawLinks[0].Target.Key != "first" {
		t.Errorf("expected link key to be 'first'")
	}
}

func TestParse_LinkAtEndOfBody(t *testing.T) {
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
Text ends with [[note:last]].
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}
	if note.RawLinks[0].Target.Key != "last" {
		t.Errorf("expected link key to be 'last'")
	}
}

func TestParse_MalformedLink_Unclosed(t *testing.T) {
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
This has [[unclosed link.
`

	note := Parse("test.md", []byte(content))

	// Unclosed links should not be extracted
	if len(note.RawLinks) != 0 {
		t.Errorf("expected 0 links for unclosed bracket, got %d", len(note.RawLinks))
	}
}

func TestParse_MalformedLink_Empty(t *testing.T) {
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
This has [[]].
`

	note := Parse("test.md", []byte(content))

	// Empty links should not be extracted (parseLinkContent will return nil)
	if len(note.RawLinks) != 0 {
		t.Errorf("expected 0 links for empty brackets, got %d", len(note.RawLinks))
	}
}

func TestParse_MalformedLink_OnlyColon(t *testing.T) {
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
This has [[:]].
`

	note := Parse("test.md", []byte(content))

	// Malformed link should not be extracted
	if len(note.RawLinks) != 0 {
		t.Errorf("expected 0 links for malformed link, got %d", len(note.RawLinks))
	}
}

func TestParse_LinkWithWhitespace(t *testing.T) {
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
This links to [[ note : key-with-spaces | edge-type ]] .
`

	note := Parse("test.md", []byte(content))

	if len(note.RawLinks) != 1 {
		t.Fatalf("expected 1 link, got %d", len(note.RawLinks))
	}

	link := note.RawLinks[0]
	if link.Target.Key != "key-with-spaces" {
		t.Errorf("expected trimmed key 'key-with-spaces', got %q", link.Target.Key)
	}
	if link.EdgeType != "edge-type" {
		t.Errorf("expected trimmed edge type 'edge-type', got %q", link.EdgeType)
	}
}

func TestParse_MultipleFrontmatterSections(t *testing.T) {
	// This tests that we only parse the first frontmatter section
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
Body text
---
This should be part of the body, not a second frontmatter.
`

	note := Parse("test.md", []byte(content))

	if note.FM.ID != "test-id" {
		t.Errorf("expected ID from first frontmatter")
	}
	if !contains(note.Body, "This should be part of the body") {
		t.Errorf("expected second --- section to be in body")
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

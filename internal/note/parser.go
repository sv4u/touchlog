package note

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sv4u/touchlog/v2/internal/model"
	"gopkg.in/yaml.v3"
)

// Parse parses a .Rmd file and returns a Note with frontmatter, body, and extracted links
// Never crashes - always produces diagnostics for errors
func Parse(path string, content []byte) *model.Note {
	note := &model.Note{
		Path:     path,
		Body:     "",
		RawLinks: []model.RawLink{},
		Diags:    []model.Diagnostic{},
		FM: model.Frontmatter{
			Extra: make(map[string]any),
		},
	}

	// Find frontmatter boundaries
	fmStart, fmEnd, err := findFrontmatter(content)
	if err != nil {
		note.Diags = append(note.Diags, model.Diagnostic{
			Level:   model.DiagnosticLevelError,
			Code:    "FRONTMATTER_MISSING",
			Message: fmt.Sprintf("Frontmatter is missing or invalid: %s. Notes must start with '---' on the first line, followed by YAML frontmatter, and end with '---'.", err.Error()),
			Span: model.Span{
				Path:      path,
				StartByte: 0,
				EndByte:   0,
			},
		})
		// Continue parsing body even if frontmatter is missing
		note.Body = string(content)
		return note
	}

	// Parse frontmatter
	fmContent := content[fmStart:fmEnd]
	if err := parseFrontmatter(fmContent, note); err != nil {
		note.Diags = append(note.Diags, model.Diagnostic{
			Level:   model.DiagnosticLevelError,
			Code:    "FRONTMATTER_PARSE_ERROR",
			Message: fmt.Sprintf("Failed to parse YAML frontmatter: %s. Check that the frontmatter is valid YAML and contains required fields (id, type, key, title, created, updated, state, tags).", err.Error()),
			Span: model.Span{
				Path:      path,
				StartByte: fmStart,
				EndByte:   fmEnd,
			},
		})
	}

	// Extract body (everything after frontmatter)
	bodyStart := fmEnd
	if bodyStart < len(content) {
		// Skip the closing --- and newline
		if bodyStart+3 < len(content) && string(content[bodyStart:bodyStart+3]) == "---" {
			bodyStart += 3
			// Skip newline if present
			if bodyStart < len(content) && content[bodyStart] == '\n' {
				bodyStart++
			}
		}
		note.Body = string(content[bodyStart:])
	}

	// Extract links from body
	note.RawLinks = extractLinks(path, note.Body, bodyStart)

	return note
}

// findFrontmatter finds the start and end positions of frontmatter
// Returns (start, end, error)
func findFrontmatter(content []byte) (int, int, error) {
	if len(content) < 3 {
		return 0, 0, fmt.Errorf("file too short to contain frontmatter")
	}

	// Frontmatter must start at byte 0 with "---"
	if string(content[0:3]) != "---" {
		return 0, 0, fmt.Errorf("frontmatter must start with '---' at byte 0")
	}

	// Find the closing "---"
	// Start searching after the opening "---"
	searchStart := 3
	// Skip the newline after opening ---
	if searchStart < len(content) && content[searchStart] == '\n' {
		searchStart++
	}

	for i := searchStart; i < len(content)-2; i++ {
		if content[i] == '-' && content[i+1] == '-' && content[i+2] == '-' {
			// Check if it's at the start of a line (preceded by newline or at start)
			if i == searchStart || content[i-1] == '\n' {
				return 0, i + 3, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("frontmatter closing '---' not found")
}

// parseFrontmatter parses YAML frontmatter into the note's Frontmatter struct
func parseFrontmatter(content []byte, note *model.Note) error {
	var raw map[string]any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return fmt.Errorf("YAML parse error: %w", err)
	}

	// Extract known fields
	if id, ok := raw["id"].(string); ok {
		note.FM.ID = model.NoteID(id)
	}

	if typ, ok := raw["type"].(string); ok {
		note.FM.Type = model.TypeName(typ)
	}

	if key, ok := raw["key"].(string); ok {
		note.FM.Key = model.Key(key)
	}

	// Handle created time (can be string or time.Time from YAML)
	if created, ok := raw["created"]; ok {
		switch v := created.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				note.FM.Created = t
			}
		case time.Time:
			note.FM.Created = v
		}
	}

	// Handle updated time (can be string or time.Time from YAML)
	if updated, ok := raw["updated"]; ok {
		switch v := updated.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				note.FM.Updated = t
			}
		case time.Time:
			note.FM.Updated = v
		}
	}

	if title, ok := raw["title"].(string); ok {
		note.FM.Title = title
	}

	if tags, ok := raw["tags"].([]any); ok {
		note.FM.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				note.FM.Tags = append(note.FM.Tags, s)
			}
		}
	}

	if state, ok := raw["state"].(string); ok {
		note.FM.State = state
	}

	// Store unknown fields in Extra
	knownFields := map[string]bool{
		"id":      true,
		"type":    true,
		"key":     true,
		"created": true,
		"updated": true,
		"title":   true,
		"tags":    true,
		"state":   true,
	}

	for k, v := range raw {
		if !knownFields[k] {
			note.FM.Extra[k] = v
		}
	}

	return nil
}

// extractLinks extracts wiki-links from the body text
// Recognizes: [[type:key]], [[key]], [[type:key|edge-type]]
func extractLinks(path, body string, bodyStartOffset int) []model.RawLink {
	var links []model.RawLink

	// Pattern to match wiki-links: [[...]]
	// This matches: [[type:key]], [[key]], [[type:key|edge-type]]
	linkPattern := regexp.MustCompile(`\[\[([^\]]+)\]\]`)

	matches := linkPattern.FindAllStringSubmatchIndex(body, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		// match[0], match[1] = full match (including brackets)
		// match[2], match[3] = content inside brackets
		contentStart := match[2]
		contentEnd := match[3]
		linkContent := body[contentStart:contentEnd]

		// Parse the link content
		rawLink := parseLinkContent(path, linkContent, bodyStartOffset+match[0], bodyStartOffset+match[1])
		if rawLink != nil {
			links = append(links, *rawLink)
		}
	}

	return links
}

// parseLinkContent parses the content inside [[...]] brackets
func parseLinkContent(path, content string, startByte, endByte int) *model.RawLink {
	// Determine source (we don't have it in Phase 0, will be set during indexing)
	source := model.TypeKey{}

	// Parse target and optional edge type
	var target model.RawTarget
	edgeType := model.DefaultEdgeType

	// Check for edge type separator: |
	parts := strings.Split(content, "|")
	if len(parts) == 2 {
		// Has edge type: [[type:key|edge-type]] or [[key|edge-type]]
		edgeType = model.EdgeType(strings.TrimSpace(parts[1]))
		content = parts[0]
	}

	// Parse target (type:key or just key)
	if strings.Contains(content, ":") {
		// Qualified: type:key
		typeKeyParts := strings.SplitN(content, ":", 2)
		if len(typeKeyParts) == 2 {
			typeNameStr := strings.TrimSpace(typeKeyParts[0])
			keyStr := strings.TrimSpace(typeKeyParts[1])
			// Reject if type or key is empty
			if typeNameStr == "" || keyStr == "" {
				return nil
			}
			typeName := model.TypeName(typeNameStr)
			key := model.Key(keyStr)
			target = model.RawTarget{
				Type: &typeName,
				Key:  key,
			}
		} else {
			// Malformed, skip
			return nil
		}
	} else {
		// Unqualified: key
		keyStr := strings.TrimSpace(content)
		// Reject if key is empty
		if keyStr == "" {
			return nil
		}
		target = model.RawTarget{
			Type: nil,
			Key:  model.Key(keyStr),
		}
	}

	// Calculate line and column (approximate)
	// For now, we'll set StartLine and StartCol to 0 (optional in v0)
	span := model.Span{
		Path:      path,
		StartByte: startByte,
		EndByte:   endByte,
		StartLine: 0, // Will be calculated properly in later phases
		StartCol:  0,
	}

	return &model.RawLink{
		Source:   source,
		Target:   target,
		EdgeType: edgeType,
		Span:     span,
	}
}

package index

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
	_ "modernc.org/sqlite"
)

// TestBuilder_PathBasedKeys_IndexesSubfolders tests that notes in subfolders are indexed correctly
func TestBuilder_PathBasedKeys_IndexesSubfolders(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    100,
			},
		},
		Edges: make(map[model.EdgeType]config.EdgeDef),
	}

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "note", "projects", "web")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("creating nested dir: %v", err)
	}

	// Create note with path-based key in subfolder
	notePath := filepath.Join(nestedDir, "auth.Rmd")
	noteContent := `---
id: note-1
type: note
key: projects/web/auth
title: Auth System
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Auth System
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatalf("writing test note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify index
	indexPath := filepath.Join(touchlogDir, "index.db")
	db, err := sql.Open("sqlite", indexPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("opening index: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Verify node was indexed with full path key
	var key string
	err = db.QueryRow("SELECT key FROM nodes WHERE id = 'note-1'").Scan(&key)
	if err != nil {
		t.Fatalf("querying nodes: %v", err)
	}
	if key != "projects/web/auth" {
		t.Errorf("expected key 'projects/web/auth', got %q", key)
	}
}

// TestBuilder_LinkResolution_ByLastSegment tests unqualified link resolution by last segment
func TestBuilder_LinkResolution_ByLastSegment(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    100,
			},
		},
		Edges: make(map[model.EdgeType]config.EdgeDef),
	}

	// Create target note in subfolder
	nestedDir := filepath.Join(tmpDir, "note", "projects", "web")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("creating nested dir: %v", err)
	}

	targetPath := filepath.Join(nestedDir, "auth.Rmd")
	targetContent := `---
id: target-note
type: note
key: projects/web/auth
title: Auth System
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Auth System
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatalf("writing target note: %v", err)
	}

	// Create source note with unqualified link using last segment
	noteDir := filepath.Join(tmpDir, "note")
	sourcePath := filepath.Join(noteDir, "source.Rmd")
	sourceContent := `---
id: source-note
type: note
key: source
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Link to [[auth]] using last segment.
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify the link was resolved
	indexPath := filepath.Join(touchlogDir, "index.db")
	db, err := sql.Open("sqlite", indexPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("opening index: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Check that the edge was resolved to target-note
	var toID string
	err = db.QueryRow("SELECT to_id FROM edges WHERE from_id = 'source-note'").Scan(&toID)
	if err != nil {
		t.Fatalf("querying edges: %v", err)
	}
	if toID != "target-note" {
		t.Errorf("expected link to resolve to 'target-note', got %q", toID)
	}
}

// TestBuilder_LinkResolution_Ambiguous tests that ambiguous links generate diagnostics
func TestBuilder_LinkResolution_Ambiguous(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    100,
			},
		},
		Edges: make(map[model.EdgeType]config.EdgeDef),
	}

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create first target in subfolder (path key "projects/auth")
	// Note: No flat "auth" key exists, only path-based keys with same last segment
	projectsDir := filepath.Join(noteDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("creating projects dir: %v", err)
	}

	target1Path := filepath.Join(projectsDir, "auth.Rmd")
	target1Content := `---
id: target-1
type: note
key: projects/auth
title: Auth 1
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
`
	if err := os.WriteFile(target1Path, []byte(target1Content), 0644); err != nil {
		t.Fatalf("writing target 1: %v", err)
	}

	// Create second target in different subfolder (path key "users/auth")
	usersDir := filepath.Join(noteDir, "users")
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		t.Fatalf("creating users dir: %v", err)
	}

	target2Path := filepath.Join(usersDir, "auth.Rmd")
	target2Content := `---
id: target-2
type: note
key: users/auth
title: Auth 2
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
`
	if err := os.WriteFile(target2Path, []byte(target2Content), 0644); err != nil {
		t.Fatalf("writing target 2: %v", err)
	}

	// Create source note with ambiguous link
	// Since "auth" doesn't match any key exactly, it falls back to last-segment matching
	// and finds both "projects/auth" and "users/auth" → ambiguous
	sourcePath := filepath.Join(noteDir, "source.Rmd")
	sourceContent := `---
id: source-note
type: note
key: source
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Link to [[auth]] which is ambiguous (matches projects/auth and users/auth by last segment).
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify the link was NOT resolved (ambiguous)
	indexPath := filepath.Join(touchlogDir, "index.db")
	db, err := sql.Open("sqlite", indexPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("opening index: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Check that the edge has NULL to_id (ambiguous)
	var toID sql.NullString
	err = db.QueryRow("SELECT to_id FROM edges WHERE from_id = 'source-note'").Scan(&toID)
	if err != nil {
		t.Fatalf("querying edges: %v", err)
	}
	if toID.Valid {
		t.Errorf("expected NULL to_id for ambiguous link, got %q", toID.String)
	}

	// Check that a diagnostic was created
	var diagCount int
	err = db.QueryRow("SELECT COUNT(*) FROM diagnostics WHERE node_id = 'source-note' AND code = 'AMBIGUOUS_LINK'").Scan(&diagCount)
	if err != nil {
		t.Fatalf("querying diagnostics: %v", err)
	}
	if diagCount != 1 {
		t.Errorf("expected 1 AMBIGUOUS_LINK diagnostic, got %d", diagCount)
	}
}

// TestBuilder_QualifiedLink_WithPathKey tests qualified links with full path keys
func TestBuilder_QualifiedLink_WithPathKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize vault
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    100,
			},
		},
		Edges: make(map[model.EdgeType]config.EdgeDef),
	}

	noteDir := filepath.Join(tmpDir, "note")
	projectsDir := filepath.Join(noteDir, "projects", "web")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("creating projects dir: %v", err)
	}

	// Create target with path key
	targetPath := filepath.Join(projectsDir, "auth.Rmd")
	targetContent := `---
id: target-note
type: note
key: projects/web/auth
title: Auth System
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatalf("writing target note: %v", err)
	}

	// Create source note with qualified link using full path
	sourcePath := filepath.Join(noteDir, "source.Rmd")
	sourceContent := `---
id: source-note
type: note
key: source
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Link to [[note:projects/web/auth]] using full qualified path.
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Build index
	builder := NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify the link was resolved
	indexPath := filepath.Join(touchlogDir, "index.db")
	db, err := sql.Open("sqlite", indexPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("opening index: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Check that the edge was resolved correctly
	var toID string
	err = db.QueryRow("SELECT to_id FROM edges WHERE from_id = 'source-note'").Scan(&toID)
	if err != nil {
		t.Fatalf("querying edges: %v", err)
	}
	if toID != "target-note" {
		t.Errorf("expected link to resolve to 'target-note', got %q", toID)
	}
}

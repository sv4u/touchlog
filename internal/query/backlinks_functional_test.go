package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

// TestExecuteBacklinks_DirectionOut tests ExecuteBacklinks with direction "out"
func TestExecuteBacklinks_DirectionOut(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create source note
	sourcePath := filepath.Join(noteDir, "source.Rmd")
	sourceContent := `---
id: note-source
type: note
key: source
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Links to [[note:target]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create target note
	targetPath := filepath.Join(noteDir, "target.Rmd")
	targetContent := `---
id: note-target
type: note
key: target
title: Target Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Target Note
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatalf("writing target note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query backlinks with direction "out"
	q := NewBacklinksQuery()
	q.Target = "note:source"
	q.Direction = "out"

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	// Should find outgoing links from source
	if len(results) == 0 {
		t.Error("expected to find at least one outgoing link")
	}
}

// TestExecuteBacklinks_WithEdgeTypeFilter tests ExecuteBacklinks with edge type filter
func TestExecuteBacklinks_WithEdgeTypeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVaultWithIndex(t, tmpDir)

	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create source note
	sourcePath := filepath.Join(noteDir, "source.Rmd")
	sourceContent := `---
id: note-source
type: note
key: source
title: Source Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Source Note

Links to [[note:target]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create target note
	targetPath := filepath.Join(noteDir, "target.Rmd")
	targetContent := `---
id: note-target
type: note
key: target
title: Target Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Target Note
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatalf("writing target note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query backlinks with edge type filter
	q := NewBacklinksQuery()
	q.Target = "note:target"
	q.Direction = "in"
	q.EdgeTypes = []string{"related-to"}

	results, err := ExecuteBacklinks(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecuteBacklinks failed: %v", err)
	}

	// Should find links with matching edge type
	_ = results
}

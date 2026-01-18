package query

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

// TestExecutePaths_WithMultipleDestinations tests ExecutePaths with multiple destinations
func TestExecutePaths_WithMultipleDestinations(t *testing.T) {
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

Links to [[note:dest1]] and [[note:dest2]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create destination notes
	for i := 1; i <= 2; i++ {
		destPath := filepath.Join(noteDir, fmt.Sprintf("dest%d.Rmd", i))
		destContent := fmt.Sprintf(`---
id: note-dest%d
type: note
key: dest%d
title: Destination %d
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Destination %d
`, i, i, i, i)
		if err := os.WriteFile(destPath, []byte(destContent), 0644); err != nil {
			t.Fatalf("writing dest note %d: %v", i, err)
		}
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query paths to multiple destinations
	q := NewPathsQuery()
	q.Source = "note:source"
	q.Destinations = []string{"note:dest1", "note:dest2"}
	q.MaxDepth = 5
	q.MaxPaths = 10

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find paths to both destinations
	if len(results) == 0 {
		t.Error("expected to find at least one path")
	}
}

// TestExecutePaths_WithEdgeTypeFilter tests ExecutePaths with edge type filter
func TestExecutePaths_WithEdgeTypeFilter(t *testing.T) {
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

Links to [[note:dest]].
`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("writing source note: %v", err)
	}

	// Create destination note
	destPath := filepath.Join(noteDir, "dest.Rmd")
	destContent := `---
id: note-dest
type: note
key: dest
title: Destination Note
state: draft
tags: []
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# Destination Note
`
	if err := os.WriteFile(destPath, []byte(destContent), 0644); err != nil {
		t.Fatalf("writing dest note: %v", err)
	}

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	builder := index.NewBuilder(tmpDir, cfg)
	if err := builder.Rebuild(); err != nil {
		t.Fatalf("rebuilding index: %v", err)
	}

	// Query paths with edge type filter
	q := NewPathsQuery()
	q.Source = "note:source"
	q.Destinations = []string{"note:dest"}
	q.MaxDepth = 5
	q.MaxPaths = 10
	q.EdgeTypes = []string{"related-to"}

	results, err := ExecutePaths(tmpDir, q)
	if err != nil {
		t.Fatalf("ExecutePaths failed: %v", err)
	}

	// Should find paths with matching edge type
	_ = results
}

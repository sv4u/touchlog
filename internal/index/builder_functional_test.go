package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// TestBuilder_DiscoverRmdFiles tests discoverRmdFiles function behavior
func TestBuilder_DiscoverRmdFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	builder := NewBuilder(tmpDir, cfg)

	// Create note directory with multiple .Rmd files
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Create multiple .Rmd files
	files := []string{"note1.Rmd", "note2.Rmd", "note3.Rmd"}
	for _, f := range files {
		filePath := filepath.Join(noteDir, f)
		if err := os.WriteFile(filePath, []byte("# Test"), 0644); err != nil {
			t.Fatalf("writing file %s: %v", f, err)
		}
	}

	// Create a non-.Rmd file (should be ignored)
	nonRmdPath := filepath.Join(noteDir, "readme.txt")
	if err := os.WriteFile(nonRmdPath, []byte("readme"), 0644); err != nil {
		t.Fatalf("writing non-Rmd file: %v", err)
	}

	// Discover .Rmd files
	rmdFiles, err := builder.discoverRmdFiles(noteDir)
	if err != nil {
		t.Fatalf("discoverRmdFiles failed: %v", err)
	}

	if len(rmdFiles) != 3 {
		t.Errorf("expected 3 .Rmd files, got %d", len(rmdFiles))
	}

	// Verify all files are .Rmd
	for _, file := range rmdFiles {
		if filepath.Ext(file) != ".Rmd" {
			t.Errorf("expected .Rmd file, got %q", file)
		}
	}
}

// TestBuilder_DiscoverRmdFiles_EmptyDirectory tests discoverRmdFiles with empty directory
func TestBuilder_DiscoverRmdFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	builder := NewBuilder(tmpDir, cfg)

	// Create empty note directory
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	// Discover .Rmd files
	rmdFiles, err := builder.discoverRmdFiles(noteDir)
	if err != nil {
		t.Fatalf("discoverRmdFiles failed: %v", err)
	}

	if len(rmdFiles) != 0 {
		t.Errorf("expected 0 .Rmd files, got %d", len(rmdFiles))
	}
}

// TestBuilder_DiscoverTypeDirectories tests discoverTypeDirectories function behavior
func TestBuilder_DiscoverTypeDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
			"task": {
				Description:  "A task",
				DefaultState: "todo",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	builder := NewBuilder(tmpDir, cfg)

	// Create type directories
	noteDir := filepath.Join(tmpDir, "note")
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		t.Fatalf("creating note dir: %v", err)
	}

	taskDir := filepath.Join(tmpDir, "task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("creating task dir: %v", err)
	}

	// Discover type directories
	typeDirs, err := builder.discoverTypeDirectories()
	if err != nil {
		t.Fatalf("discoverTypeDirectories failed: %v", err)
	}

	if len(typeDirs) != 2 {
		t.Errorf("expected 2 type directories, got %d", len(typeDirs))
	}

	// Verify both types are found
	if _, ok := typeDirs["note"]; !ok {
		t.Error("expected 'note' type directory to be found")
	}
	if _, ok := typeDirs["task"]; !ok {
		t.Error("expected 'task' type directory to be found")
	}
}

// TestBuilder_DiscoverTypeDirectories_NonExistent tests discoverTypeDirectories when directories don't exist
func TestBuilder_DiscoverTypeDirectories_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types: map[model.TypeName]config.TypeDef{
			"note": {
				Description:  "A note",
				DefaultState: "draft",
				KeyMaxLen:    64,
			},
		},
		Tags:      config.TagConfig{Preferred: []string{}},
		Edges:     make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{Root: "templates"},
	}

	builder := NewBuilder(tmpDir, cfg)

	// Don't create note directory

	// Discover type directories - should still work (returns empty map or creates dirs)
	typeDirs, err := builder.discoverTypeDirectories()
	if err != nil {
		t.Fatalf("discoverTypeDirectories failed: %v", err)
	}

	// Should return empty map or create directories
	_ = typeDirs
}

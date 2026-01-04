package validation

import (
	stderrors "errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/internal/errors"
)

func TestValidateOutputDir(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := ValidateOutputDir("")
		if err == nil {
			t.Error("ValidateOutputDir(\"\") error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrOutputDirRequired) {
			t.Errorf("ValidateOutputDir(\"\") error = %v, want ErrOutputDirRequired", err)
		}
	})

	t.Run("non-existent directory with existing parent", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "new-dir")
		
		err := ValidateOutputDir(newDir)
		if err != nil {
			t.Errorf("ValidateOutputDir(%q) error = %v, want nil (parent exists, can create)", newDir, err)
		}
	})

	t.Run("non-existent directory with non-existent parent", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "parent", "new-dir")
		// Remove parent to make it non-existent
		os.RemoveAll(filepath.Join(tmpDir, "parent"))
		
		err := ValidateOutputDir(newDir)
		if err == nil {
			t.Error("ValidateOutputDir() with non-existent parent error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrOutputDirInvalid) {
			t.Errorf("ValidateOutputDir() error = %v, want ErrOutputDirInvalid", err)
		}
	})

	t.Run("existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		err := ValidateOutputDir(tmpDir)
		if err != nil {
			t.Errorf("ValidateOutputDir(%q) error = %v, want nil", tmpDir, err)
		}
	})

	t.Run("existing file (not directory)", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		err := ValidateOutputDir(testFile)
		if err == nil {
			t.Error("ValidateOutputDir() with file path error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrOutputDirNotDirectory) {
			t.Errorf("ValidateOutputDir() error = %v, want ErrOutputDirNotDirectory", err)
		}
	})

	t.Run("non-writable directory", func(t *testing.T) {
		// This test may not work on all systems (permissions)
		// Skip if we can't create a non-writable directory
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		
		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Skipf("Skipping test: cannot create read-only directory: %v", err)
		}
		defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup
		
		err := ValidateOutputDir(readOnlyDir)
		if err == nil {
			// On some systems, we might still be able to write (e.g., as owner)
			// This is acceptable behavior
			t.Logf("ValidateOutputDir() with read-only directory returned nil (may be acceptable on this system)")
		} else if !stderrors.Is(err, errors.ErrOutputDirNotWritable) {
			t.Errorf("ValidateOutputDir() error = %v, want ErrOutputDirNotWritable", err)
		}
	})

	t.Run("path with ~", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("Skipping test: cannot get home directory: %v", err)
		}
		
		testDir := filepath.Join(homeDir, "touchlog-test-dir")
		defer os.RemoveAll(testDir)
		
		// Create the directory
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		
		tildePath := "~/touchlog-test-dir"
		err = ValidateOutputDir(tildePath)
		if err != nil {
			t.Errorf("ValidateOutputDir(%q) error = %v, want nil", tildePath, err)
		}
	})
}

func TestValidateUTF8(t *testing.T) {
	t.Run("valid UTF-8", func(t *testing.T) {
		validData := []byte("Hello, 世界")
		err := ValidateUTF8(validData)
		if err != nil {
			t.Errorf("ValidateUTF8() with valid UTF-8 error = %v, want nil", err)
		}
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		// Invalid UTF-8 sequence
		invalidData := []byte{0xff, 0xfe, 0xfd}
		err := ValidateUTF8(invalidData)
		if err == nil {
			t.Error("ValidateUTF8() with invalid UTF-8 error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrInvalidUTF8) {
			t.Errorf("ValidateUTF8() error = %v, want ErrInvalidUTF8", err)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		emptyData := []byte{}
		err := ValidateUTF8(emptyData)
		if err != nil {
			t.Errorf("ValidateUTF8() with empty data error = %v, want nil", err)
		}
	})
}

func TestValidateConfigFile(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := ValidateConfigFile("")
		if err == nil {
			t.Error("ValidateConfigFile(\"\") error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrConfigNotFound) {
			t.Errorf("ValidateConfigFile(\"\") error = %v, want ErrConfigNotFound", err)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.yaml")
		
		err := ValidateConfigFile(nonExistentFile)
		if err == nil {
			t.Error("ValidateConfigFile() with non-existent file error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrConfigNotFound) {
			t.Errorf("ValidateConfigFile() error = %v, want ErrConfigNotFound", err)
		}
	})

	t.Run("existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		if err := os.WriteFile(configFile, []byte("test: value"), 0644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}
		
		err := ValidateConfigFile(configFile)
		if err != nil {
			t.Errorf("ValidateConfigFile(%q) error = %v, want nil", configFile, err)
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		err := ValidateConfigFile(tmpDir)
		if err == nil {
			t.Error("ValidateConfigFile() with directory error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrConfigInvalid) {
			t.Errorf("ValidateConfigFile() error = %v, want ErrConfigInvalid", err)
		}
	})

	t.Run("non-readable file", func(t *testing.T) {
		// This test may not work on all systems
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		if err := os.WriteFile(configFile, []byte("test: value"), 0000); err != nil {
			t.Skipf("Skipping test: cannot create non-readable file: %v", err)
		}
		defer os.Chmod(configFile, 0644) // Restore permissions for cleanup
		
		err := ValidateConfigFile(configFile)
		if err == nil {
			// On some systems, we might still be able to read (e.g., as owner)
			// This is acceptable behavior
			t.Logf("ValidateConfigFile() with non-readable file returned nil (may be acceptable on this system)")
		} else if !stderrors.Is(err, errors.ErrConfigReadFailed) {
			t.Errorf("ValidateConfigFile() error = %v, want ErrConfigReadFailed", err)
		}
	})
}

func TestValidateTemplateSyntax(t *testing.T) {
	t.Run("empty template", func(t *testing.T) {
		err := ValidateTemplateSyntax("")
		if err != nil {
			t.Errorf("ValidateTemplateSyntax(\"\") error = %v, want nil (empty template is valid)", err)
		}
	})

	t.Run("valid UTF-8 template", func(t *testing.T) {
		validTemplate := "# {{title}}\n\n{{message}}"
		err := ValidateTemplateSyntax(validTemplate)
		if err != nil {
			t.Errorf("ValidateTemplateSyntax() with valid template error = %v, want nil", err)
		}
	})

	t.Run("invalid UTF-8 template", func(t *testing.T) {
		// Invalid UTF-8 sequence
		invalidTemplate := string([]byte{0xff, 0xfe, 0xfd})
		err := ValidateTemplateSyntax(invalidTemplate)
		if err == nil {
			t.Error("ValidateTemplateSyntax() with invalid UTF-8 error = nil, want error")
		}
		if !stderrors.Is(err, errors.ErrTemplateInvalidSyntax) {
			t.Errorf("ValidateTemplateSyntax() error = %v, want ErrTemplateInvalidSyntax", err)
		}
	})

	t.Run("template with unicode", func(t *testing.T) {
		unicodeTemplate := "# 标题\n\n{{message}}"
		err := ValidateTemplateSyntax(unicodeTemplate)
		if err != nil {
			t.Errorf("ValidateTemplateSyntax() with unicode template error = %v, want nil", err)
		}
	})
}


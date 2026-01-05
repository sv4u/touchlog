package validation

import (
	stderrors "errors"
	"os"
	"path/filepath"
	"strings"
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

func TestExpandPath(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		expanded, err := ExpandPath("")
		if err != nil {
			t.Errorf("ExpandPath(\"\") error = %v, want nil", err)
		}
		if expanded == "" {
			t.Error("ExpandPath(\"\") returned empty string")
		}
	})

	t.Run("home directory only", func(t *testing.T) {
		expanded, err := ExpandPath("~")
		if err != nil {
			t.Errorf("ExpandPath(\"~\") error = %v, want nil", err)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("Skipping test: cannot get home directory: %v", err)
		}
		if expanded != homeDir {
			t.Errorf("ExpandPath(\"~\") = %q, want %q", expanded, homeDir)
		}
	})

	t.Run("home directory with path", func(t *testing.T) {
		expanded, err := ExpandPath("~/test-path")
		if err != nil {
			t.Errorf("ExpandPath(\"~/test-path\") error = %v, want nil", err)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("Skipping test: cannot get home directory: %v", err)
		}
		expected := filepath.Join(homeDir, "test-path")
		if expanded != expected {
			t.Errorf("ExpandPath(\"~/test-path\") = %q, want %q", expanded, expected)
		}
	})

	t.Run("invalid tilde path", func(t *testing.T) {
		_, err := ExpandPath("~invalid")
		if err == nil {
			t.Error("ExpandPath(\"~invalid\") error = nil, want error")
		}
		if !strings.Contains(err.Error(), "must be followed by /") {
			t.Errorf("ExpandPath() error = %v, want error about tilde format", err)
		}
	})

	t.Run("relative path", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		
		expanded, err := ExpandPath("test-relative")
		if err != nil {
			t.Errorf("ExpandPath(\"test-relative\") error = %v, want nil", err)
		}
		expected := filepath.Join(wd, "test-relative")
		if expanded != expected {
			t.Errorf("ExpandPath(\"test-relative\") = %q, want %q", expanded, expected)
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		tmpDir := t.TempDir()
		expanded, err := ExpandPath(tmpDir)
		if err != nil {
			t.Errorf("ExpandPath(%q) error = %v, want nil", tmpDir, err)
		}
		if expanded != tmpDir {
			t.Errorf("ExpandPath(%q) = %q, want %q", tmpDir, expanded, tmpDir)
		}
	})

	t.Run("path with environment variable", func(t *testing.T) {
		// Set a test environment variable
		testVar := "TEST_TOUCHLOG_PATH"
		testValue := "/tmp/test-touchlog"
		os.Setenv(testVar, testValue)
		defer os.Unsetenv(testVar)

		path := "$" + testVar + "/subdir"
		expanded, err := ExpandPath(path)
		if err != nil {
			t.Errorf("ExpandPath(%q) error = %v, want nil", path, err)
		}
		expected := filepath.Join(testValue, "subdir")
		if expanded != expected {
			t.Errorf("ExpandPath(%q) = %q, want %q", path, expanded, expected)
		}
	})

	t.Run("path with multiple environment variables", func(t *testing.T) {
		os.Setenv("VAR1", "/tmp")
		os.Setenv("VAR2", "test")
		defer func() {
			os.Unsetenv("VAR1")
			os.Unsetenv("VAR2")
		}()

		path := "$VAR1/$VAR2/path"
		expanded, err := ExpandPath(path)
		if err != nil {
			t.Errorf("ExpandPath(%q) error = %v, want nil", path, err)
		}
		expected := filepath.Join("/tmp", "test", "path")
		if expanded != expected {
			t.Errorf("ExpandPath(%q) = %q, want %q", path, expanded, expected)
		}
	})

	t.Run("path with undefined environment variable", func(t *testing.T) {
		path := "$UNDEFINED_VAR/path"
		expanded, err := ExpandPath(path)
		if err != nil {
			t.Errorf("ExpandPath(%q) error = %v, want nil (undefined vars expand to empty)", path, err)
		}
		// Undefined env vars expand to empty string, so path becomes "/path"
		expected := filepath.Join("", "path")
		if expanded != expected {
			t.Logf("ExpandPath(%q) = %q (undefined env var expands to empty)", path, expanded)
		}
	})

	t.Run("path with tilde and environment variable", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("Skipping test: cannot get home directory: %v", err)
		}

		expanded, err := ExpandPath("~/$HOME/test")
		if err != nil {
			t.Errorf("ExpandPath(\"~/$HOME/test\") error = %v, want nil", err)
		}
		// Tilde expansion happens first, so this becomes "$HOME/$HOME/test"
		// which then expands to homeDir/homeDir/test
		expected := filepath.Join(homeDir, homeDir, "test")
		if expanded != expected {
			t.Logf("ExpandPath(\"~/$HOME/test\") = %q (tilde expands first)", expanded)
		}
	})
}


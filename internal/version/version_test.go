package version

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout captures stdout output and returns it as a string
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	// Capture output in a goroutine
	var buf bytes.Buffer
	done := make(chan bool)
	go func() {
		io.Copy(&buf, r)
		done <- true
	}()

	fn()
	w.Close()
	<-done

	return buf.String()
}

func TestString(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalGitCommit := GitCommit
	defer func() {
		Version = originalVersion
		GitCommit = originalGitCommit
	}()

	tests := []struct {
		name     string
		version  string
		commit   string
		want     string
		contains []string
	}{
		{
			name:     "dev version",
			version:  "dev",
			commit:   "abc1234",
			want:     "touchlog dev",
			contains: []string{"dev", "abc1234"},
		},
		{
			name:     "release version",
			version:  "v1.2.3",
			commit:   "abc1234",
			want:     "touchlog v1.2.3",
			contains: []string{"v1.2.3"},
		},
		{
			name:     "prerelease version",
			version:  "v1.2.3-alpha.1",
			commit:   "abc1234",
			want:     "touchlog v1.2.3-alpha.1",
			contains: []string{"v1.2.3-alpha.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			GitCommit = tt.commit
			got := String()

			if !strings.Contains(got, tt.want) {
				t.Errorf("String() = %v, want to contain %v", got, tt.want)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("String() = %v, want to contain %v", got, substr)
				}
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	originalStdout := os.Stdout
	defer func() {
		Version = originalVersion
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
		os.Stdout = originalStdout
	}()

	t.Run("release version", func(t *testing.T) {
		Version = "v1.2.3"
		GitCommit = "abc1234"
		BuildDate = "2025-01-01T00:00:00Z"

		// Capture stdout to prevent terminal interference
		output := captureStdout(t, PrintVersion)

		// Verify output contains expected content
		if !strings.Contains(output, "touchlog v1.2.3") {
			t.Errorf("PrintVersion() output = %q, want to contain 'touchlog v1.2.3'", output)
		}
		if !strings.Contains(output, "Build date: 2025-01-01T00:00:00Z") {
			t.Errorf("PrintVersion() output = %q, want to contain 'Build date: 2025-01-01T00:00:00Z'", output)
		}
		if !strings.Contains(output, "Git commit: abc1234") {
			t.Errorf("PrintVersion() output = %q, want to contain 'Git commit: abc1234'", output)
		}
	})

	t.Run("dev version", func(t *testing.T) {
		Version = "dev"
		GitCommit = "abc1234"
		BuildDate = "2025-01-01T00:00:00Z"

		// Capture stdout to prevent terminal interference
		output := captureStdout(t, PrintVersion)

		// Verify output contains expected content
		if !strings.Contains(output, "touchlog dev (commit abc1234)") {
			t.Errorf("PrintVersion() output = %q, want to contain 'touchlog dev (commit abc1234)'", output)
		}
		if !strings.Contains(output, "Build date: 2025-01-01T00:00:00Z") {
			t.Errorf("PrintVersion() output = %q, want to contain 'Build date: 2025-01-01T00:00:00Z'", output)
		}
		// Dev version should NOT print Git commit separately (it's in the main string)
		if strings.Contains(output, "Git commit:") {
			t.Errorf("PrintVersion() output = %q, should NOT contain separate 'Git commit:' line for dev version", output)
		}
	})

	t.Run("dev version with unknown commit", func(t *testing.T) {
		Version = "dev"
		GitCommit = "unknown"
		BuildDate = "unknown"

		// Capture stdout to prevent terminal interference
		output := captureStdout(t, PrintVersion)

		// Verify output contains expected content
		if !strings.Contains(output, "touchlog dev (commit unknown)") {
			t.Errorf("PrintVersion() output = %q, want to contain 'touchlog dev (commit unknown)'", output)
		}
		// Should not print Build date or Git commit when unknown
		if strings.Contains(output, "Build date:") {
			t.Errorf("PrintVersion() output = %q, should NOT contain 'Build date:' when BuildDate is unknown", output)
		}
		if strings.Contains(output, "Git commit:") {
			t.Errorf("PrintVersion() output = %q, should NOT contain 'Git commit:' when GitCommit is unknown", output)
		}
	})
}

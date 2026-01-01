package version

import (
	"strings"
	"testing"
)

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
	defer func() {
		Version = originalVersion
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
	}()

	t.Run("release version", func(t *testing.T) {
		Version = "v1.2.3"
		GitCommit = "abc1234"
		BuildDate = "2025-01-01T00:00:00Z"

		// For release versions, Git commit should be printed separately
		// We can't easily test stdout output, but we can verify it doesn't panic
		PrintVersion()
	})

	t.Run("dev version", func(t *testing.T) {
		Version = "dev"
		GitCommit = "abc1234"
		BuildDate = "2025-01-01T00:00:00Z"

		// For dev versions, Git commit should NOT be printed separately
		// (it's already included in String() output)
		// We can't easily test stdout output, but we can verify it doesn't panic
		PrintVersion()
	})

	t.Run("dev version with unknown commit", func(t *testing.T) {
		Version = "dev"
		GitCommit = "unknown"
		BuildDate = "unknown"

		// Dev version with unknown commit should still work
		PrintVersion()
	})
}


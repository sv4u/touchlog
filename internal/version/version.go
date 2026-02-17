package version

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

// Version is the version of the application.
// This variable is set at build time using ldflags.
// Default value is "dev" if not set during build.
var Version = "dev"

// Commit is the git commit hash of the build.
// This variable is set at build time using ldflags.
// Default value is empty string if not set during build.
var Commit = ""

// getVersionFromBuildInfo attempts to extract version information from
// runtime/debug.BuildInfo when ldflags are not set.
// This allows go install to show proper version information.
func getVersionFromBuildInfo() (string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}

	// Try to get version from Main module first
	var version string
	mainVersion := info.Main.Version
	if mainVersion != "" && mainVersion != "(devel)" {
		version = strings.TrimPrefix(mainVersion, "v")
	}

	// If Main doesn't have version, check dependencies for our module
	// This handles cases where the module is a dependency of the build
	if version == "" {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/sv4u/touchlog/v2" {
				if dep.Version != "" && dep.Version != "(devel)" {
					version = strings.TrimPrefix(dep.Version, "v")
					break
				}
			}
		}
	}

	if version == "" {
		return "", ""
	}

	// Try to extract commit hash from BuildInfo settings
	// Go stores VCS revision in BuildInfo.Settings (full 40-char SHA-1)
	var commit string
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			commit = setting.Value
			// Use full commit hash (40 chars) to match GoReleaser's FullCommit format
			// If it's shorter, use what we have
			if len(commit) > 40 {
				commit = commit[:40]
			}
			break
		}
	}

	// Handle pseudo-version format: "2.1.1-0.20240101120000-abc123def456"
	// The last segment is an embedded commit hash that must be stripped to
	// avoid duplication when vcs.revision is also present.
	parts := strings.Split(version, "-")
	if len(parts) >= 3 {
		lastPart := parts[len(parts)-1]
		if len(lastPart) >= 7 && isPseudoVersionCommit(lastPart) {
			// Strip the embedded commit from the version
			cleanVersion := strings.Join(parts[:len(parts)-1], "-")
			if commit == "" {
				// Use the embedded commit if we don't have one from vcs.revision
				if len(lastPart) >= 12 {
					commit = lastPart[:12]
				} else {
					commit = lastPart
				}
			}
			version = cleanVersion
		}
	}

	return version, commit
}

// isPseudoVersionCommit checks if a string looks like a hex commit hash
// (as embedded in Go pseudo-versions).
func isPseudoVersionCommit(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return len(s) > 0
}

// GetVersion returns the version string.
// If Version is "dev" and Commit is empty, attempts to read version from
// runtime/debug.BuildInfo (for go install builds).
// If Commit is set, it appends the commit hash to the version.
func GetVersion() string {
	// If version info was injected via ldflags, use it
	if Version != "dev" || Commit != "" {
		if Commit != "" {
			return Version + "-" + Commit
		}
		return Version
	}

	// Try to get version from BuildInfo (for go install builds)
	buildInfoVersion, buildInfoCommit := getVersionFromBuildInfo()
	if buildInfoVersion != "" {
		if buildInfoCommit != "" {
			return buildInfoVersion + "-" + buildInfoCommit
		}
		return buildInfoVersion
	}

	// No version information available
	fmt.Fprintf(os.Stderr, "warning: version information not injected at build time (built without ldflags)\n")
	return "dev"
}

package version

import (
	"fmt"
	"os"
)

// Version is the version string, injected at build time via -ldflags
// Default value is "dev" for development builds
var Version = "dev"

// GitCommit is the git commit hash, injected at build time via -ldflags
// Default value is "unknown" for development builds
var GitCommit = "unknown"

// BuildDate is the build date in RFC3339 format, injected at build time via -ldflags
// Default value is "unknown" for development builds
var BuildDate = "unknown"

// String returns a formatted version string
// Format: "touchlog v1.2.3" for releases, "touchlog dev (commit abc1234)" for dev builds
func String() string {
	if Version == "dev" {
		return "touchlog dev (commit " + GitCommit + ")"
	}
	return "touchlog " + Version
}

// PrintVersion prints the version information to stdout
func PrintVersion() {
	fmt.Fprintf(os.Stdout, "%s\n", String())
	if BuildDate != "unknown" {
		fmt.Fprintf(os.Stdout, "Build date: %s\n", BuildDate)
	}
	// Only print Git commit separately if Version is not "dev"
	// (when Version is "dev", the commit is already included in String() output)
	if GitCommit != "unknown" && Version != "dev" {
		fmt.Fprintf(os.Stdout, "Git commit: %s\n", GitCommit)
	}
}

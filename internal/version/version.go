package version

// Version is the version of the application.
// This variable is set at build time using ldflags.
// Default value is "dev" if not set during build.
var Version = "dev"

// Commit is the git commit hash of the build.
// This variable is set at build time using ldflags.
// Default value is empty string if not set during build.
var Commit = ""

// GetVersion returns the version string.
// If Commit is set, it appends the commit hash to the version.
func GetVersion() string {
	if Commit != "" {
		return Version + "-" + Commit
	}
	return Version
}

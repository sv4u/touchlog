package metadata

import (
	"os"
	"os/user"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/entry"
)

// CollectMetadata collects metadata based on configuration and flags
// Returns a Metadata struct with user, host, and optionally git context
func CollectMetadata(cfg *config.Config, includeGit bool, outputDir string) (*entry.Metadata, error) {
	meta := &entry.Metadata{}

	// Determine if we should include user and host
	includeUser := true // Default to true
	includeHost := true // Default to true

	if cfg != nil {
		includeUser = cfg.GetIncludeUser()
		includeHost = cfg.GetIncludeHost()
	}

	// Collect user if enabled
	if includeUser {
		username, err := GetUsername()
		if err != nil {
			// If we can't get username, continue without it (don't fail)
			meta.User = ""
		} else {
			meta.User = username
		}
	}

	// Collect host if enabled
	if includeHost {
		hostname, err := GetHostname()
		if err != nil {
			// If we can't get hostname, continue without it (don't fail)
			meta.Host = ""
		} else {
			meta.Host = hostname
		}
	}

	// Collect git context if requested
	if includeGit {
		gitCtx, err := GetGitContext(outputDir)
		if err != nil {
			// If we can't get git context, continue without it (don't fail)
			meta.Git = nil
		} else {
			meta.Git = gitCtx
		}
	}

	return meta, nil
}

// GetUsername returns the current username
// Returns an error if unable to determine username
func GetUsername() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Username, nil
}

// GetHostname returns the hostname of the current machine
// Returns an error if unable to determine hostname
func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}


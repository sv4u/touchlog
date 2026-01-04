package metadata

import (
	"os"
	"os/user"
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

func TestGetUsername(t *testing.T) {
	username, err := GetUsername()
	if err != nil {
		t.Fatalf("GetUsername() error = %v", err)
	}
	if username == "" {
		t.Error("GetUsername() returned empty string")
	}
	// Verify it matches current user
	currentUser, err := user.Current()
	if err == nil {
		if username != currentUser.Username {
			t.Errorf("GetUsername() = %v, want %v", username, currentUser.Username)
		}
	}
}

func TestGetHostname(t *testing.T) {
	hostname, err := GetHostname()
	if err != nil {
		t.Fatalf("GetHostname() error = %v", err)
	}
	if hostname == "" {
		t.Error("GetHostname() returned empty string")
	}
	// Verify it matches system hostname
	systemHostname, err := os.Hostname()
	if err == nil {
		if hostname != systemHostname {
			t.Errorf("GetHostname() = %v, want %v", hostname, systemHostname)
		}
	}
}

func TestCollectMetadata_DefaultEnabled(t *testing.T) {
	cfg := config.CreateDefaultConfig()

	meta, err := CollectMetadata(cfg, false, "")
	if err != nil {
		t.Fatalf("CollectMetadata() error = %v", err)
	}
	if meta == nil {
		t.Fatal("CollectMetadata() returned nil metadata")
	}

	// User and host should be enabled by default
	if meta.User == "" {
		t.Error("CollectMetadata() user is empty, should be populated by default")
	}
	if meta.Host == "" {
		t.Error("CollectMetadata() host is empty, should be populated by default")
	}

	// Git should be nil when includeGit is false
	if meta.Git != nil {
		t.Error("CollectMetadata() git should be nil when includeGit is false")
	}
}

func TestCollectMetadata_IncludeGit(t *testing.T) {
	cfg := config.CreateDefaultConfig()

	// Test with includeGit = true
	// This will only work if we're in a git repo
	meta, err := CollectMetadata(cfg, true, "")
	if err != nil {
		t.Fatalf("CollectMetadata() error = %v", err)
	}

	// CollectMetadata always returns nil error (handles errors gracefully)
	// Verify metadata is not nil
	if meta == nil {
		t.Fatal("CollectMetadata() returned nil metadata")
	}

	// Git context may or may not be populated depending on whether we're in a repo
	// If git context is present, verify it has valid values
	if meta.Git != nil {
		if meta.Git.Branch == "" && meta.Git.Commit == "" {
			t.Error("CollectMetadata() git context is not nil but has no values")
		}
	}
	// If git context is nil, that's okay - we might not be in a git repo
}

func TestCollectMetadata_ConfigDisabled(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	includeUser := false
	includeHost := false
	cfg.IncludeUser = &includeUser
	cfg.IncludeHost = &includeHost

	meta, err := CollectMetadata(cfg, false, "")
	if err != nil {
		t.Fatalf("CollectMetadata() error = %v", err)
	}
	if meta == nil {
		t.Fatal("CollectMetadata() returned nil metadata")
	}

	// User and host should be empty when disabled
	if meta.User != "" {
		t.Errorf("CollectMetadata() user = %v, want empty when disabled", meta.User)
	}
	if meta.Host != "" {
		t.Errorf("CollectMetadata() host = %v, want empty when disabled", meta.Host)
	}
}

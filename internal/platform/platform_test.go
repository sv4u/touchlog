package platform

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	platform, err := Detect()

	// Should not error on supported platforms
	if err != nil && platform != PlatformWindows {
		t.Errorf("Detect() returned error on supported platform: %v", err)
	}

	// Should return appropriate platform
	switch runtime.GOOS {
	case "darwin":
		if platform != PlatformDarwin {
			t.Errorf("Detect() = %v, want %v", platform, PlatformDarwin)
		}
	case "linux":
		if platform != PlatformLinux && platform != PlatformWSL {
			t.Errorf("Detect() = %v, want %v or %v", platform, PlatformLinux, PlatformWSL)
		}
	case "windows":
		if err == nil {
			t.Error("Detect() should return error on Windows")
		}
		if platform != PlatformWindows {
			t.Errorf("Detect() = %v, want %v", platform, PlatformWindows)
		}
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     bool
	}{
		{"darwin", PlatformDarwin, true},
		{"linux", PlatformLinux, true},
		{"wsl", PlatformWSL, true},
		{"windows", PlatformWindows, false},
		{"unknown", Platform("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupported(tt.platform); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsWSL(t *testing.T) {
	// Save original environment
	originalWSLDistro := os.Getenv("WSL_DISTRO_NAME")
	defer func() {
		if originalWSLDistro != "" {
			os.Setenv("WSL_DISTRO_NAME", originalWSLDistro)
		} else {
			os.Unsetenv("WSL_DISTRO_NAME")
		}
	}()

	// Test with WSL_DISTRO_NAME set
	os.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	if !IsWSL() {
		t.Error("IsWSL() = false, want true when WSL_DISTRO_NAME is set")
	}

	// Test without WSL_DISTRO_NAME
	os.Unsetenv("WSL_DISTRO_NAME")
	// On non-Linux systems, IsWSL should return false
	// On Linux systems, it depends on /proc/version or /proc/sys/kernel/osrelease
	// We can't reliably test this without mocking file reads, so we just verify it doesn't panic
	_ = IsWSL()
}

func TestCheckSupported(t *testing.T) {
	// This test depends on the actual platform
	// On supported platforms, should not error
	// On Windows, should error with helpful message
	err := CheckSupported()

	switch runtime.GOOS {
	case "darwin", "linux":
		if err != nil {
			t.Errorf("CheckSupported() returned error on supported platform: %v", err)
		}
	case "windows":
		if err == nil {
			t.Error("CheckSupported() should return error on Windows")
		}
		if err != nil && !strings.Contains(err.Error(), "WSL") {
			t.Errorf("CheckSupported() error should mention WSL, got: %v", err)
		}
	default:
		// For unknown platforms (FreeBSD, OpenBSD, etc.), should return generic error
		// without Windows-specific advice
		if err == nil {
			t.Error("CheckSupported() should return error on unsupported platform")
		}
		if err != nil && strings.Contains(err.Error(), "WSL") {
			t.Errorf("CheckSupported() error should not mention WSL for non-Windows platforms, got: %v", err)
		}
		if err != nil && !strings.Contains(err.Error(), "unsupported platform") {
			t.Errorf("CheckSupported() error should mention unsupported platform, got: %v", err)
		}
	}
}

// TestDetectUnknownPlatform verifies that unknown platforms return empty Platform and generic error
func TestDetectUnknownPlatform(t *testing.T) {
	// We can't easily test this on the actual runtime platform,
	// but we can verify the logic: for unknown platforms, Detect should return
	// an empty Platform string and ErrUnsupportedPlatform
	// This test verifies the default case behavior conceptually

	// Test that the default case returns empty platform for unknown OS
	// Since we can't change runtime.GOOS, we test the error message logic
	// by checking that CheckSupported doesn't give Windows-specific advice
	// for non-Windows platforms

	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		// We're on an unknown platform - verify we get generic error
		platform, err := Detect()
		if err == nil {
			t.Error("Detect() should return error on unknown platform")
		}
		if platform != Platform("") {
			t.Errorf("Detect() should return empty Platform for unknown OS, got: %v", platform)
		}
		if err != ErrUnsupportedPlatform {
			t.Errorf("Detect() should return ErrUnsupportedPlatform for unknown OS, got: %v", err)
		}
	}
}

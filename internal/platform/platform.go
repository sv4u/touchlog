package platform

import (
	stderrors "errors"
	"os"
	"runtime"
	"strings"

	"github.com/sv4u/touchlog/internal/errors"
)

// Platform represents the detected platform type
type Platform string

const (
	// PlatformDarwin represents macOS
	PlatformDarwin Platform = "darwin"
	// PlatformLinux represents Linux
	PlatformLinux Platform = "linux"
	// PlatformWSL represents Windows Subsystem for Linux
	PlatformWSL Platform = "wsl"
	// PlatformWindows represents Windows native (unsupported)
	PlatformWindows Platform = "windows"
)

// ErrUnsupportedPlatform is re-exported from the errors package for backward compatibility
var ErrUnsupportedPlatform = errors.ErrPlatformUnsupported

// Detect detects the current platform and returns the Platform type
func Detect() (Platform, error) {
	goos := runtime.GOOS

	switch goos {
	case "darwin":
		return PlatformDarwin, nil
	case "linux":
		// Check if running in WSL
		if IsWSL() {
			return PlatformWSL, nil
		}
		return PlatformLinux, nil
	case "windows":
		return PlatformWindows, errors.ErrPlatformUnsupported
	default:
		// Unknown platform (FreeBSD, OpenBSD, NetBSD, Solaris, etc.)
		// Return a generic error without platform-specific advice
		return Platform(""), errors.ErrPlatformUnsupported
	}
}

// IsSupported checks if the given platform is supported
func IsSupported(p Platform) bool {
	return p == PlatformDarwin || p == PlatformLinux || p == PlatformWSL
}

// IsWSL detects if the current environment is WSL (Windows Subsystem for Linux)
// It checks multiple methods to reliably detect WSL
// Returns false immediately if not running on Linux (defensive check)
func IsWSL() bool {
	// Defensive check: WSL only exists on Linux
	if runtime.GOOS != "linux" {
		return false
	}

	// Method 1: Check WSL_DISTRO_NAME environment variable
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	// Method 2: Check /proc/version for WSL indicators
	if version, err := os.ReadFile("/proc/version"); err == nil {
		versionStr := strings.ToLower(string(version))
		if strings.Contains(versionStr, "microsoft") || strings.Contains(versionStr, "wsl") {
			return true
		}
	}

	// Method 3: Check /proc/sys/kernel/osrelease for WSL indicators
	if osrelease, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		osreleaseStr := strings.ToLower(string(osrelease))
		if strings.Contains(osreleaseStr, "microsoft") || strings.Contains(osreleaseStr, "wsl") {
			return true
		}
	}

	return false
}

// CheckSupported is a convenience function that checks if the current platform is supported
// It returns an error with a helpful message if the platform is not supported
func CheckSupported() error {
	platform, err := Detect()
	if err != nil {
		// Check if it's actually Windows (not just an unknown platform)
		if runtime.GOOS == "windows" && platform == PlatformWindows {
			return stderrors.New("unsupported platform: touchlog only supports macOS, Linux, and WSL. Windows native is not supported. Please use WSL")
		}
		// For unknown platforms, return the generic error
		return err
	}

	if !IsSupported(platform) {
		return errors.ErrPlatformUnsupported
	}

	return nil
}

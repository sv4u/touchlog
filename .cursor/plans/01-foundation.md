# Phase 1: Foundation & Platform Detection

**Goal**: Establish platform support and basic CLI structure

**Duration**: 1-2 weeks  
**Complexity**: Medium  
**Status**: Not Started

## Prerequisites

- None (this is the foundational phase)

## Dependencies on Other Phases

- None (this phase enables all other work)

## Overview

This phase establishes the foundational infrastructure for the refactored touchlog:
- Platform detection to ensure macOS/Linux/WSL support
- Version management with build-time injection
- CLI framework setup using Cobra
- Basic command structure

## Tasks

### 1.1 Platform Detection Module

**Location**: `internal/platform/platform.go`

**Requirements**:

- Detect operating system (darwin, linux, windows)
- Detect WSL environment on Linux (check `/proc/version` or `WSL_DISTRO_NAME` env var)
- Reject Windows native with clear error message
- Support macOS, Linux, and WSL
- Early exit on unsupported platforms (before any other initialization)

**WSL Detection Methods**:

1. Check `WSL_DISTRO_NAME` environment variable
2. Check `/proc/version` for "Microsoft" or "WSL"
3. Check `/proc/sys/kernel/osrelease` for WSL indicators

**Implementation**:

```go
package platform

import (
    "errors"
    "os"
    "runtime"
    "strings"
)

type Platform string

const (
    PlatformDarwin  Platform = "darwin"
    PlatformLinux   Platform = "linux"
    PlatformWSL     Platform = "wsl"
    PlatformWindows Platform = "windows"
)

var (
    ErrUnsupportedPlatform = errors.New("unsupported platform: touchlog only supports macOS, Linux, and WSL")
)

func Detect() (Platform, error)
func IsSupported(p Platform) bool
func IsWSL() bool
func CheckSupported() error  // Convenience function for main()
```

**Error Messages**:

- Windows: `"unsupported platform: touchlog only supports macOS, Linux, and WSL. Windows native is not supported. Please use WSL."`
- Include helpful suggestion for Windows users

**Files to Create**:

- `internal/platform/platform.go`
- `internal/platform/platform_test.go`

**Tests**:

- Test platform detection on different OS (use build tags: `//go:build !windows`)
- Test WSL detection (mock `/proc/version` or env vars)
- Test Windows rejection (use build tag `//go:build windows`)
- Test error messages are clear and helpful

**See Also**: [shared-concerns.md](./shared-concerns.md) for security and error handling guidelines

---

### 1.2 Version Management

**Location**: `internal/version/version.go`

**Requirements**:

- Build-time version injection via `-ldflags`
- Semantic versioning format: `vX.Y.Z[-prerelease]`
- Accessible via `--version` flag

**Implementation**:

```go
package version

var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)

func String() string
func PrintVersion()
```

**Build Integration**:

- Update CI/CD workflows (`.github/workflows/release.yml`) to inject version during build
- Use git tags for version source
- Format: `touchlog v1.2.3` or `touchlog v1.2.3-alpha`
- Build command example:
  ```bash
  go build -ldflags "-X github.com/sv4u/touchlog/internal/version.Version=$(git describe --tags --always) -X github.com/sv4u/touchlog/internal/version.GitCommit=$(git rev-parse --short HEAD) -X github.com/sv4u/touchlog/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/touchlog
  ```
- For development builds (no tags): use `dev` version
- Update `.goreleaser.yml` to include version injection in build hooks

**Version String Format**:

- Release: `touchlog v1.2.3`
- Pre-release: `touchlog v1.2.3-alpha.1`
- Development: `touchlog dev (commit abc1234)`

**Files to Create**:

- `internal/version/version.go`
- `internal/version/version_test.go`

**Files to Modify**:

- `.github/workflows/release.yml` (add version injection)
- `.goreleaser.yml` (add build hooks for version)
- `Makefile` (if exists, or create one for local builds)

**See Also**: [shared-concerns.md](./shared-concerns.md) for CI/CD integration details

---

### 1.3 CLI Framework Setup

**Location**: `cmd/touchlog/main.go` (refactor)

**Requirements**:

- Replace `flag` package with proper CLI framework (cobra/spf13)
- Support subcommands: `new`, `config`, `list`, `search`
- Support global flags: `--help`, `--version`, `--config`
- Default behavior: launch REPL wizard when no subcommand

**CLI Library Choice**:

- **Option A**: `github.com/spf13/cobra` (recommended)
  - Industry standard
  - Rich features (subcommands, flags, help)
  - Good documentation
- **Option B**: `github.com/urfave/cli/v2`
  - Simpler API
  - Less features

**Recommendation**: Use `cobra` for better subcommand support

**Implementation Structure**:

```go
// cmd/touchlog/main.go
func main() {
    // Platform check (first thing)
    if err := platform.CheckSupported(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // CLI setup
    rootCmd := &cobra.Command{
        Use:   "touchlog",
        Short: "A terminal-based note editor",
        // ...
    }
    
    rootCmd.AddCommand(newCmd)
    rootCmd.AddCommand(configCmd)
    // ... other commands
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

**Files to Modify**:

- `cmd/touchlog/main.go` (complete rewrite)

**Files to Create**:

- `cmd/touchlog/commands/root.go`
- `cmd/touchlog/commands/new.go` (stub for Phase 4)
- `cmd/touchlog/commands/config.go` (stub for Phase 2)
- `cmd/touchlog/commands/list.go` (future)
- `cmd/touchlog/commands/search.go` (future)

**Dependencies to Add**:

- `github.com/spf13/cobra` (or alternative)

**See Also**: [shared-concerns.md](./shared-concerns.md) for dependency management

---

## Implementation Checklist

- [ ] Platform detection module
- [ ] Version management with build integration
- [ ] CLI framework setup (cobra)
- [ ] Root command with `--help`, `--version`
- [ ] Default wizard command (stub - will be implemented in Phase 6)
- [ ] Tests for platform detection
- [ ] Tests for version output
- [ ] Tests for CLI routing
- [ ] Update CI/CD workflows for version injection
- [ ] Update build scripts/Makefile

## Testing Requirements

### Unit Tests

- `internal/platform/platform_test.go`
  - Test platform detection on different OS
  - Test WSL detection
  - Test Windows rejection
  - Test error messages

- `internal/version/version_test.go`
  - Test version string formatting
  - Test version output

- `cmd/touchlog/commands/root_test.go`
  - Test CLI routing
  - Test flag parsing
  - Test help output

### Integration Tests

- Test platform check in main()
- Test version flag output
- Test help flag output

## Success Criteria

- ✅ Platform detection works correctly (macOS/Linux/WSL)
- ✅ Windows native is rejected with clear error
- ✅ Version flag outputs correct version string
- ✅ CLI framework is set up with subcommand structure
- ✅ Root command works with `--help` and `--version`
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 1, proceed to:
- **[Phase 2: Configuration System Enhancement](./02-configuration.md)** - Requires Phase 1 (CLI framework)

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns
- Cobra documentation: https://github.com/spf13/cobra


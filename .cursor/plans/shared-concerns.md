# Shared Concerns: Cross-Cutting Requirements

This document outlines concerns that apply across all phases of the refactoring. Each phase plan should reference this document for consistency.

## Dependencies to Add

### Required

- `github.com/spf13/cobra` - CLI framework
  - Version: Latest stable (v1.x)
  - Purpose: Subcommand routing, flag parsing, help generation
  - Migration: Replace `flag` package usage
- `github.com/charmbracelet/bubbles` - TUI components (already have)
  - Verify version compatibility with new wizard requirements
  - May need: `list`, `textarea`, `textinput`, `spinner` components
- `github.com/charmbracelet/bubbletea` - TUI framework (already have)
  - Verify version supports all required features
- `github.com/charmbracelet/lipgloss` - Styling (already have)
  - Verify version supports styling requirements

### Optional (Later)

- `github.com/pelletier/go-toml/v2` - TOML support
  - Version: Latest v2.x
  - Purpose: TOML config file parsing
  - When: Phase 2.2 (after YAML enhancements)
- `gopkg.in/yaml.v3` - YAML support (already have)
  - Verify version supports all YAML features needed

### Dependency Management

- Update `go.mod` with new dependencies
- Run `go mod tidy` after adding dependencies
- Pin versions in `go.mod` (don't use `@latest` in production)
- Document any breaking changes in dependencies
- Test compatibility with existing dependencies
- Update `.github/dependabot.yml` if needed

## Migration Strategy

### Backward Compatibility

- **Config Files**: Maintain backward compatibility with existing YAML configs
  - Old format: `templates: [{name: "...", file: "..."}]` still works
  - New format: `templates: {daily: "..."}` (inline) also supported
  - Both formats can coexist (inline takes precedence)
- **Templates**: Support both file-based and inline templates
  - Existing file-based templates continue to work
  - New inline templates are optional enhancement
- **API**: Keep `internal/api/api.go` working (may need updates)
  - Update `api.Run()` to support new wizard mode
  - Maintain `Options` struct compatibility
  - Add new options for wizard vs non-interactive mode
  - Example:
    ```go
    type Options struct {
        OutputDirectory string
        ConfigPath      string
        Mode            string  // "wizard" or "new" or "legacy"
        // ... new fields
    }
    ```

### Breaking Changes

- **CLI**: New subcommand structure (old flags may not work)
- **Default Behavior**: `touchlog` with no args now launches wizard (was TUI editor)
- **Config Format**: New config structure (but old format still supported)

### Migration Path

1. Document breaking changes in CHANGELOG
2. Provide migration guide for config files
3. Update README with new usage
4. Version bump (major if breaking, minor if compatible)

## CI/CD Integration

### Build System Updates

**Version Injection in CI/CD**:

- Update `.github/workflows/ci.yml` to inject version during test builds
- Update `.github/workflows/release.yml` to inject version during release builds
- Use GoReleaser hooks for version injection in `.goreleaser.yml`

**Example CI Integration**:

```yaml
# .github/workflows/ci.yml
- name: Build with version
  run: |
    VERSION=$(git describe --tags --always || echo "dev")
    COMMIT=$(git rev-parse --short HEAD)
    DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    go build -ldflags "-X github.com/sv4u/touchlog/internal/version.Version=$VERSION -X github.com/sv4u/touchlog/internal/version.GitCommit=$COMMIT -X github.com/sv4u/touchlog/internal/version.BuildDate=$DATE" ./cmd/touchlog
```

### Test Coverage Requirements

- Maintain minimum 20% coverage (existing threshold)
- Target 80%+ coverage for new code
- Add coverage reports for new packages
- Update codecov configuration if needed

### Linting and Code Quality

- Ensure golangci-lint passes for all new code
- Run `go vet` and `staticcheck` on new packages
- Maintain code formatting standards (`gofmt`)

## Documentation Requirements

### Code Documentation

- Add package-level documentation for all new packages
- Add function-level documentation for exported functions
- Use Go doc conventions
- Include examples in documentation where helpful

### User Documentation

- Update `README.md` with new CLI structure
- Document new configuration options
- Add migration guide for existing users
- Update examples to reflect new architecture
- Document breaking changes in `CHANGELOG.md`

### Developer Documentation

- Document architecture decisions
- Add inline comments for complex logic
- Document test helper usage
- Add troubleshooting guide

## Performance Considerations

### Startup Performance

- **Target**: <100ms for simple commands (`--version`, `--help`)
- **Target**: <200ms for `touchlog new` with minimal flags
- Platform detection should be <10ms
- Config loading should be <50ms (with caching)

### Runtime Performance

- File operations should be efficient (use buffered I/O)
- Template processing should handle large templates (>10KB)
- TUI should remain responsive during file operations
- Use async operations for slow file system operations

### Memory Considerations

- Avoid loading entire log files into memory unnecessarily
- Use streaming for large stdin inputs
- Limit template size (configurable limit)

## Security Considerations

### Input Validation

- Validate all user inputs (paths, filenames, config values)
- Sanitize filenames to prevent path traversal
- Validate UTF-8 encoding for all text inputs
- Limit file size for stdin input (configurable)

### File System Security

- Validate output directory permissions before writing
- Prevent overwriting files outside output directory
- Use secure file permissions (0644 for files, 0755 for directories)
- Handle symlinks safely

### Editor Launch Security

- Validate editor paths before launching
- Prevent command injection in editor resolution
- Use `exec.Command` with proper argument handling
- Don't execute arbitrary shell commands
- Sanitize file paths passed to editors
- Validate editor executables exist and are executable
- Don't allow relative paths that escape intended directories

### Concurrent Access Handling

- Handle concurrent file creation (multiple `touchlog new` commands)
- Use file locking or atomic operations for filename collision detection
- Handle race conditions in filename generation
- Ensure atomic file writes (write to temp file, then rename)
- Handle concurrent wizard instances gracefully (if supported)

## Edge Cases and Error Scenarios

### File System Edge Cases

- Disk full during write → Return clear error, don't leave partial files
- Permission denied after file creation → Clean up created file, return error
- Concurrent file creation (race conditions) → Use atomic operations, handle gracefully
- Network filesystem delays → Add timeout, show spinner
- Invalid characters in filenames → Sanitize, use safe defaults
- Symlinks in output directory → Resolve safely, prevent traversal
- Output directory is a file (not directory) → Validate, return clear error

### Configuration Edge Cases

- Missing config file (should use defaults) → No error, use defaults
- Invalid YAML/TOML syntax → Return parse error with line number
- Circular template references (if supported) → Detect and prevent
- Very large config files → Add size limit, warn if exceeded
- Invalid timezone names → Fallback to system timezone, warn user
- Config file in invalid location → Use XDG fallback, warn
- Config file permissions (read-only) → Handle gracefully, use defaults

### Template Edge Cases

- Missing template variables → Leave as-is (don't replace), or use empty string
- Very large templates → Handle efficiently, add size limit
- Nested template variables (if supported) → Support or document limitation
- Invalid template syntax → Validate, return clear error
- Template file permissions → Return permission error
- Template contains `{{` in content → Escape properly or document limitation
- Empty template → Allow, use as-is

### Editor Edge Cases

- Editor executable missing after resolution → Fallback to internal editor
- Editor fails to launch → Return error, keep file
- Editor exits immediately → Consider success (user may have saved quickly)
- Multiple editor processes → Allow, document behavior
- Editor modifies file during wizard review → Re-read file before final save
- Editor with arguments in EDITOR env var → Parse correctly (e.g., `EDITOR="vim -f"`)

### Wizard Edge Cases

- User interrupts (Ctrl+C) during wizard → Clean up created file if exists

## Logging and Debugging Strategy

### Logging Approach

- **No logging by default** (CLI tool, not a service)
- **Debug mode**: Add `--debug` flag for verbose output
- **Error messages**: Always clear and actionable
- **Structured errors**: Use error types for consistent messaging

### Debug Mode Implementation

```go
// internal/debug/debug.go
package debug

var Enabled bool

func Log(format string, args ...interface{})
func LogError(err error, context string)
func DumpConfig(cfg *config.Config)
func DumpState(state interface{})
```

**Usage**:

- `touchlog --debug new --message "test"` → Verbose output
- `touchlog --debug` → Debug wizard mode
- Debug output goes to stderr
- Include: config loading, file operations, editor resolution, state transitions

### Error Reporting

- Clear, user-friendly error messages
- Include context (what operation failed, why)
- Suggest solutions when possible
- Use structured errors for programmatic access

## Rollback and Recovery Strategies

### File Creation Rollback

- If file creation fails partway through, clean up partial files
- Use atomic writes: write to temp file, then rename
- If rename fails, delete temp file
- Never leave orphaned files

### Configuration Rollback

- If config parsing fails, use defaults
- If config validation fails, return error before any operations
- Don't modify user's config file automatically

### Wizard State Recovery

- If wizard crashes, don't leave created files (unless at review screen)
- At review screen, file is considered "saved" (user can recover)
- Document recovery procedures for users


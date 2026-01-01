# Phase 2: Configuration System Enhancement

**Goal**: Support YAML/TOML, inline templates, strict validation, precedence

**Duration**: 1 week  
**Complexity**: Medium  
**Status**: Not Started

## Prerequisites

- Phase 1 (Foundation) - CLI framework must be in place

## Dependencies on Other Phases

- **Requires**: Phase 1 (CLI framework for `--config` flag)
- **Enables**: Phase 3 (Template system needs config), Phase 4 (Entry creation needs config)

## Overview

This phase enhances the configuration system to support:
- Config file discovery (XDG, current directory, explicit)
- YAML format enhancements (inline templates)
- Configuration precedence (CLI > Config > Defaults)
- Strict mode validation
- Timezone configuration support

## Tasks

### 2.1 Configuration Format Support

**Location**: `internal/config/config.go` (extend)

**Requirements**:

- **Priority**: YAML support (existing, enhance)
- **Later**: TOML support (add)
- Auto-detect format by file extension
- Support both `.yaml`/`.yml` and `.toml`
- Config file discovery: XDG paths, current directory, explicit path
- Support config file in current directory (`touchlog.yaml`, `touchlog.toml`)

**Config File Discovery Order**:

1. Explicit `--config` flag path
2. Current directory: `./touchlog.yaml` or `./touchlog.toml`
3. XDG config directory: `$XDG_CONFIG_HOME/touchlog/config.yaml` (or `~/.config/touchlog/config.yaml`)
4. If none found, use defaults (no error)

**Config File Loading**:

```go
func FindConfigFile(explicitPath string) (string, error)
func LoadConfigFromPath(path string) (*Config, error)
func DetectConfigFormat(path string) (ConfigFormat, error)
```

**YAML Enhancements**:

- Support inline templates in config:
  ```yaml
  templates:
    daily: |
      # {{date}}
      ## Title
      {{title}}
      ## Notes
      {{message}}
  ```

- Support template name configuration:
  ```yaml
  template:
    name: "daily"
  ```

**TOML Support** (Phase 2.2 - Future):

- Add `github.com/pelletier/go-toml/v2` dependency
- Create TOML parser
- Map TOML structure to same Config struct

**Files to Modify**:

- `internal/config/config.go`
- `internal/config/config_test.go`

**Files to Create**:

- `internal/config/loader.go` (format detection)
- `internal/config/toml.go` (TOML parser - later)

**Dependencies to Add**:

- `github.com/pelletier/go-toml/v2` (for TOML support - later)

**See Also**: [shared-concerns.md](./shared-concerns.md) for dependency management

---

### 2.2 Configuration Precedence

**Location**: `internal/config/config.go` (extend)

**Requirements**:

- Precedence order: CLI flags > Config file > Defaults
- CLI flags override config values
- Config file overrides defaults
- No merging (first non-empty value wins)

**Implementation**:

```go
type Config struct {
    OutputDir string
    Template  string
    Editor    string
    // ... other fields
}

func LoadWithPrecedence(configPath string, cliFlags *CLIFlags) (*Config, error)
```

**Files to Modify**:

- `internal/config/config.go`

---

### 2.3 Strict Mode Validation

**Location**: `internal/config/config.go` (extend)

**Requirements**:

- Reject unknown config keys in strict mode
- `--strict` flag for validation
- Clear error messages with key names

**Implementation**:

```go
func ValidateStrict(cfg *Config, knownKeys []string) error
```

**Files to Modify**:

- `internal/config/config.go`
- `internal/config/config_test.go`

---

### 2.4 Timezone Configuration Support

**Location**: `internal/config/config.go` (extend)

**Requirements**:

- Support timezone configuration in config file (optional)
- Default to system timezone if not specified
- Use IANA timezone database (e.g., "America/Denver", "UTC")
- Handle timezone parsing errors gracefully

**Timezone Configuration**:

```yaml
timezone: "America/Denver"  # Optional, defaults to system timezone
```

**Files to Modify**:

- `internal/config/config.go`

**See Also**: Phase 3 (Template system will use timezone for date formatting)

---

## Implementation Checklist

- [ ] Config file discovery (XDG, current dir, explicit)
- [ ] YAML format detection and loading
- [ ] Inline template support in YAML
- [ ] Template name configuration
- [ ] Config precedence (CLI > Config > Defaults)
- [ ] Strict mode validation
- [ ] Timezone configuration support
- [ ] Tests for config loading
- [ ] Tests for precedence
- [ ] Tests for strict mode
- [ ] Tests for timezone handling

## Testing Requirements

### Unit Tests

- `internal/config/config_test.go`
  - Test config file discovery order
  - Test YAML parsing with inline templates
  - Test precedence resolution
  - Test strict mode validation
  - Test timezone configuration

- `internal/config/loader_test.go`
  - Test format detection
  - Test file discovery

### Integration Tests

- Test config loading from different locations
- Test CLI flag precedence over config
- Test strict mode with invalid config

## Success Criteria

- ✅ Config file discovery works (XDG, current dir, explicit)
- ✅ YAML parsing works with inline templates
- ✅ Configuration precedence works correctly
- ✅ Strict mode validation works
- ✅ Timezone configuration works
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 2, proceed to:
- **[Phase 3: Template System Refactoring](./03-templates.md)** - Requires Phase 2 (config system)

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns


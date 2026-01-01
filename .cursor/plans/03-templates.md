# Phase 3: Template System Refactoring

**Goal**: Support both inline and file-based templates

**Duration**: 1 week  
**Complexity**: Medium  
**Status**: Not Started

## Prerequisites

- Phase 2 (Configuration System) - Config system must support inline templates

## Dependencies on Other Phases

- **Requires**: Phase 2 (Config system for inline templates)
- **Enables**: Phase 4 (Entry creation needs templates), Phase 7 (Metadata integration)

## Overview

This phase refactors the template system to support:
- Both inline templates (from config) and file-based templates
- Template resolution (inline first, then file)
- Timezone-aware date/time formatting
- Variable escaping for security
- Metadata variable support

## Tasks

### 3.1 Template Resolution

**Location**: `internal/template/template.go` (refactor)

**Requirements**:

- Support inline templates (from config)
- Support file-based templates (existing)
- Template name resolution:
  1. Check inline templates first
  2. Fall back to file-based templates
  3. Error if not found

**Implementation**:

```go
type TemplateSource interface {
    GetTemplate(name string) (string, error)
}

type InlineTemplateSource struct {
    templates map[string]string
}

type FileTemplateSource struct {
    templatesDir string
}

func ResolveTemplate(name string, config *config.Config) (string, error)
```

**Files to Modify**:

- `internal/template/template.go`
- `internal/template/template_test.go`

---

### 3.2 Template Variable System

**Location**: `internal/template/template.go` (extend)

**Requirements**:

- Support variables: `{{date}}`, `{{title}}`, `{{message}}`, `{{tags}}`
- Date formatting based on config and timezone
- Custom variables from config
- Metadata variables (user, host, git) - will be integrated in Phase 7
- Timezone-aware date/time formatting
- Variable escaping (prevent injection of template syntax in user input)

**Timezone Handling**:

- Support timezone configuration in config file (optional)
- Default to system timezone if not specified
- Use IANA timezone database (e.g., "America/Denver", "UTC")
- Format dates/times in specified timezone
- Handle timezone parsing errors gracefully

**Variable Escaping**:

- Escape user-provided variables (title, message, tags) to prevent template injection
- Don't escape system variables (date, time, etc.) - they're trusted
- Escape strategy: Replace `{{` with `{{ "{{" }}` in user input (or similar)

**Implementation**:

```go
type TemplateVars struct {
    Date     string
    Time     string
    DateTime string
    Title    string
    Message  string
    Tags     []string
    User     string
    Host     string
    Git      *GitContext
}

type TemplateProcessor struct {
    timezone *time.Location
    config   *config.Config
}

func NewTemplateProcessor(cfg *config.Config, tz *time.Location) *TemplateProcessor
func (p *TemplateProcessor) ProcessTemplate(content string, vars *TemplateVars) (string, error)
func (p *TemplateProcessor) EscapeUserInput(input string) string
func (p *TemplateProcessor) FormatDate(t time.Time) string
func (p *TemplateProcessor) FormatTime(t time.Time) string
func (p *TemplateProcessor) FormatDateTime(t time.Time) string
```

**Timezone Configuration**:

```yaml
timezone: "America/Denver"  # Optional, defaults to system timezone
```

**Files to Modify**:

- `internal/template/template.go`
- `internal/template/template_test.go`

**Files to Create**:

- `internal/template/timezone.go` (timezone handling)
- `internal/template/escape.go` (variable escaping)

**See Also**: [shared-concerns.md](./shared-concerns.md) for security guidelines

---

## Implementation Checklist

- [ ] Template source interface
- [ ] Inline template source
- [ ] File-based template source
- [ ] Template resolution (inline first, then file)
- [ ] Template variable system
- [ ] Timezone-aware date formatting
- [ ] Variable escaping
- [ ] Metadata variable support (stub for Phase 7)
- [ ] Tests for template resolution
- [ ] Tests for variable substitution
- [ ] Tests for timezone handling
- [ ] Tests for variable escaping

## Testing Requirements

### Unit Tests

- `internal/template/template_test.go`
  - Test template resolution (inline vs file-based)
  - Test variable substitution
  - Test timezone handling
  - Test variable escaping

- `internal/template/timezone_test.go`
  - Test timezone parsing
  - Test date/time formatting in different timezones

- `internal/template/escape_test.go`
  - Test variable escaping
  - Test template injection prevention

## Success Criteria

- ✅ Template resolution works (inline first, then file)
- ✅ Variable substitution works correctly
- ✅ Timezone-aware date formatting works
- ✅ Variable escaping prevents template injection
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 3, proceed to:
- **[Phase 4: Non-Interactive Mode](./04-non-interactive.md)** - Requires Phase 2 (config), Phase 3 (templates)

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns (security)


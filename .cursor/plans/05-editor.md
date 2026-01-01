# Phase 5: Editor Integration

**Goal**: Support both internal (Bubble Tea) and external editor handoff

**Duration**: 1 week  
**Complexity**: Medium  
**Status**: Not Started

## Prerequisites

- Phase 1 (Foundation) - CLI framework must be in place
- Phase 2 (Configuration System) - Config system for editor settings

## Dependencies on Other Phases

- **Requires**: Phase 1 (CLI), Phase 2 (Config)
- **Enables**: Phase 4 (Non-interactive mode can use editors), Phase 6 (Wizard can use editors)

## Overview

This phase implements editor integration supporting both internal (Bubble Tea) and external editors:
- Editor resolution with precedence chain
- External editor launch (non-blocking)
- Internal editor refactoring (Bubble Tea)
- Fallback logic (Option C: both modes can use either editor type)

## Tasks

### 5.1 Editor Resolution

**Location**: `internal/editor/resolver.go`

**Requirements**:

- Precedence: CLI `--editor` > Config `editor` > `EDITOR` env var > `vi` > `nano` > fallback to internal editor
- Detect if editor executable exists on PATH
- Return editor type (external/internal) and command/args
- Support both external and internal editor modes (Option C decision)

**Implementation**:

```go
type EditorType int

const (
    EditorTypeExternal EditorType = iota
    EditorTypeInternal
)

type EditorInfo struct {
    Type     EditorType
    Command  string
    Args     []string
    UseInternal bool
}

type EditorResolver struct {
    cliEditor    string
    configEditor string
    envEditor    string
    fallbackToInternal bool
}

func (r *EditorResolver) Resolve() (*EditorInfo, error)
func (r *EditorResolver) ResolveExternal() (string, []string, error)
func FindEditorOnPath(name string) (string, error)
func ShouldUseInternalEditor(externalErr error) bool
```

**Editor Resolution Logic**:

1. Try CLI `--editor` flag
2. Try Config `editor` setting
3. Try `EDITOR` environment variable
4. Try `vi` on PATH
5. Try `nano` on PATH
6. If all fail and `fallbackToInternal` is true, return internal editor
7. If all fail and `fallbackToInternal` is false, return error

**Files to Create**:

- `internal/editor/resolver.go`
- `internal/editor/resolver_test.go`

**See Also**: [shared-concerns.md](./shared-concerns.md) for editor launch security

---

### 5.2 External Editor Launch

**Location**: `internal/editor/launcher.go`

**Requirements**:

- Launch external editor with file path
- Non-blocking (touchlog exits after launch)
- Handle launch failures gracefully
- File remains even if editor fails

**Implementation**:

```go
func LaunchEditor(editor string, args []string, filePath string) error
```

**Files to Create**:

- `internal/editor/launcher.go`
- `internal/editor/launcher_test.go`

**See Also**: [shared-concerns.md](./shared-concerns.md) for editor launch security

---

### 5.3 Internal Editor (Bubble Tea)

**Location**: `internal/editor/tui.go` (refactor existing)

**Requirements**:

- Keep existing Bubble Tea editor as fallback (Option C)
- Use when no external editor available OR when explicitly requested
- Support both wizard and non-interactive modes
- Maintain current functionality (template selection, editing, saving)
- Extract reusable TUI components

**Editor Usage Strategy (Option C)**:

- **Wizard Mode**: Prefer external editor, fallback to internal if unavailable
- **Non-Interactive Mode (`touchlog new --edit`)**: Prefer external editor, fallback to internal if unavailable
- **Both modes**: Can explicitly request internal editor via config or flag

**Implementation**:

```go
type InternalEditor struct {
    filePath string
    content  string
    config   *config.Config
}

func NewInternalEditor(filePath string, initialContent string, cfg *config.Config) (*InternalEditor, error)
func (e *InternalEditor) Run() error
func (e *InternalEditor) GetContent() string
```

**Files to Modify**:

- `internal/editor/editor.go` (refactor to separate TUI logic, extract reusable components)

**Files to Create**:

- `internal/editor/tui.go` (extract TUI logic for internal editor)
- `internal/editor/components.go` (reusable TUI components)

---

## Implementation Checklist

- [ ] Editor resolver with precedence
- [ ] External editor launch
- [ ] Internal editor (Bubble Tea) refactoring
- [ ] Editor type selection (Option C)
- [ ] Fallback logic
- [ ] Tests for editor resolution
- [ ] Tests for editor launch
- [ ] Tests for fallback

## Testing Requirements

### Unit Tests

- `internal/editor/resolver_test.go`
  - Test editor resolution precedence
  - Test PATH detection
  - Test fallback logic

- `internal/editor/launcher_test.go`
  - Test external editor launch
  - Test error handling

- `internal/editor/tui_test.go`
  - Test internal editor functionality
  - Test TUI components

### Integration Tests

- Test editor integration in `touchlog new --edit`
- Test editor integration in wizard (stub for Phase 6)

## Success Criteria

- ✅ Editor resolution works with correct precedence
- ✅ External editor launch works (non-blocking)
- ✅ Internal editor works as fallback
- ✅ Both modes can use either editor type
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 5, proceed to:
- **[Phase 6: REPL Wizard](./06-wizard.md)** - Can be done in parallel with Phase 5

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns (security)


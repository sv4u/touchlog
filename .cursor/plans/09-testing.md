# Phase 9: Testing Strategy

**Goal**: Comprehensive test coverage

**Duration**: Ongoing + 1-2 weeks comprehensive  
**Complexity**: Medium  
**Status**: Not Started

## Prerequisites

- All previous phases (testing should be added throughout)

## Dependencies on Other Phases

- **Requires**: All phases (testing is ongoing)
- **Enables**: Confidence in refactored code

## Overview

This phase implements comprehensive testing:
- Unit tests for all packages
- Integration tests for commands
- Gherkin-style test helpers (future)
- Test coverage >80%

## Tasks

### 9.1 Unit Tests

**Requirements**:

- Test all internal packages
- Mock external dependencies
- Test error cases
- Test edge cases

**Test Files**:

- `internal/platform/platform_test.go`
- `internal/config/config_test.go`
- `internal/template/template_test.go`
- `internal/entry/entry_test.go`
- `internal/editor/resolver_test.go`
- `internal/wizard/wizard_test.go`
- `internal/metadata/metadata_test.go`
- `internal/validation/validation_test.go`

---

### 9.2 Integration Tests

**Requirements**:

- Test CLI commands end-to-end
- Test file operations
- Test editor integration
- Test configuration loading

**Test Files**:

- `cmd/touchlog/commands/new_test.go`
- `cmd/touchlog/commands/config_test.go`
- `integration/integration_test.go`

---

### 9.3 Gherkin/Cucumber Tests (Future)

**Requirements**:

- Map scenarios from `SCENARIOS.md` to test cases
- Use Go testing framework (not full Cucumber)
- Test scenario coverage
- Create reusable test helpers matching Gherkin steps

**Test Helper Functions**:

```go
// features/helpers/helpers.go
package helpers

// Given steps
func GivenCleanTempDir(t *testing.T) string
func GivenConfigFile(t *testing.T, path string, content string)
func GivenSystemTime(t *testing.T, timeStr string)
func GivenTimezone(t *testing.T, tz string)
func GivenEnvVar(t *testing.T, key string, value string)
func GivenGitRepo(t *testing.T, branch string, commit string)
func GivenExecutable(t *testing.T, name string, exists bool)

// When steps
func WhenRunCommand(t *testing.T, cmd string) *CommandResult
func WhenInput(t *testing.T, input string)
func WhenStart(t *testing.T, cmd string) *InteractiveSession

// Then steps
func ThenExitCode(t *testing.T, result *CommandResult, expected int)
func ThenStdoutContains(t *testing.T, result *CommandResult, text string)
func ThenStderrContains(t *testing.T, result *CommandResult, text string)
func ThenFileExists(t *testing.T, path string)
func ThenFileContentContains(t *testing.T, path string, text string)
func ThenProcessLaunched(t *testing.T, cmd string, args []string)
```

**Test Structure**:

```go
// features/new_test.go
func TestNewCommand_CreateBasicEntry(t *testing.T) {
    dir := helpers.GivenCleanTempDir(t)
    defer os.RemoveAll(dir)
    
    helpers.GivenSystemTime(t, "2025-12-31T12:00:00-07:00")
    helpers.GivenTimezone(t, "America/Denver")
    
    result := helpers.WhenRunCommand(t, "touchlog new --message 'Did code review' --output "+dir)
    
    helpers.ThenExitCode(t, result, 0)
    helpers.ThenStdoutContains(t, result, "Wrote log to")
    helpers.ThenFileExists(t, filepath.Join(dir, "2025-12-31_*.md"))
}
```

**Files to Create** (Future):

- `features/helpers/helpers.go` (test helper functions)
- `features/helpers/command.go` (command execution helpers)
- `features/helpers/session.go` (interactive session helpers)
- `features/platform_test.go`
- `features/cli_test.go`
- `features/config_test.go`
- `features/templates_test.go`
- `features/new_test.go`
- `features/editor_test.go`
- `features/repl_wizard_test.go`
- `features/repl_ui_test.go`
- `features/metadata_test.go`
- `features/errors_test.go`

---

## Implementation Checklist

- [ ] Unit tests for all packages
- [ ] Integration tests for commands
- [ ] Gherkin-style test helpers (future)
- [ ] Test coverage >80%
- [ ] All tests passing

## Testing Requirements

### Coverage Goals

- Minimum: 20% (existing threshold)
- Target: 80%+ for new code
- All critical paths covered
- All error paths covered

### Test Quality

- Tests are readable and maintainable
- Tests use descriptive names
- Tests are independent and isolated
- Tests use proper mocking
- Tests cover edge cases

## Success Criteria

- ✅ Unit tests for all packages
- ✅ Integration tests for all commands
- ✅ Test coverage >80%
- ✅ All tests pass
- ✅ Tests are maintainable and readable

## Next Steps

After completing Phase 9:
- Review test coverage reports
- Add missing test cases
- Refactor tests for maintainability
- Document test helper usage

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns (test coverage requirements)
- `SCENARIOS.md` - Scenario definitions for Gherkin-style tests


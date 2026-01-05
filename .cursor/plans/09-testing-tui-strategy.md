# Phase 9: TUI Testing Strategy (Option C - Comprehensive)

**Status**: Future Work  
**Priority**: Low (after reaching 80% coverage on business logic)  
**Goal**: Comprehensive test coverage for TUI components using mocking and test helpers

## Overview

This document outlines a comprehensive strategy for testing TUI (Terminal User Interface) components built with Bubble Tea. The goal is to achieve high test coverage for TUI code without requiring full interactive terminal sessions.

## Challenges

1. **Bubble Tea Programs Require Terminal**: Full TUI execution requires a real terminal
2. **Interactive Nature**: User input, key presses, and terminal events are hard to simulate
3. **State Machine Complexity**: TUI state machines have many transitions and edge cases
4. **Terminal Dependencies**: Terminal size, colors, and capabilities affect behavior

## Strategy: Layered Testing Approach

### Layer 1: Model Initialization & Configuration Tests

**Goal**: Test that models are created correctly with various configurations

**Approach**:
- Test `NewWizardModel()` with different configs
- Test `NewModel()` with different options
- Verify initial state is correct
- Test configuration validation

**Example**:
```go
func TestNewWizardModel_WithConfig(t *testing.T) {
    cfg := config.CreateDefaultConfig()
    model, err := NewWizardModel(cfg, true)
    if err != nil {
        t.Fatalf("NewWizardModel() error = %v", err)
    }
    if model.wizard == nil {
        t.Error("model.wizard is nil")
    }
    // Verify initial state
    if model.wizard.State() != StateMainMenu {
        t.Errorf("initial state = %v, want %v", model.wizard.State(), StateMainMenu)
    }
}
```

### Layer 2: Update Function Tests (Message Handling)

**Goal**: Test that Update() functions correctly handle all message types

**Approach**:
- Create model instances
- Send various messages (tea.KeyMsg, tea.WindowSizeMsg, custom messages)
- Verify state transitions
- Test error handling

**Example**:
```go
func TestWizardModel_Update_KeyPresses(t *testing.T) {
    model, _ := NewWizardModel(nil, false)
    
    tests := []struct {
        name    string
        key     string
        wantState State
    }{
        {"Enter on main menu", "enter", StateTemplateSelection},
        {"Escape cancels", "esc", StateCancelled},
        {"Q quits", "q", StateQuit},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
            newModel, cmd := model.Update(msg)
            // Verify state transition
            // Verify command returned
        })
    }
}
```

### Layer 3: View Function Tests (Rendering)

**Goal**: Test that View() functions render correctly

**Approach**:
- Create models in various states
- Call View() and verify output
- Test edge cases (empty lists, long text, etc.)
- Verify styling is applied correctly

**Example**:
```go
func TestWizardModel_View_MainMenu(t *testing.T) {
    model, _ := NewWizardModel(nil, false)
    model.width = 80
    model.height = 24
    
    view := model.View()
    if view == "" {
        t.Error("View() returned empty string")
    }
    // Verify contains expected elements
    if !strings.Contains(view, "Main Menu") {
        t.Error("View() missing 'Main Menu' text")
    }
}
```

### Layer 4: Command Tests (Side Effects)

**Goal**: Test that commands (tea.Cmd) are created correctly

**Approach**:
- Test command creation for various actions
- Verify commands are properly typed
- Test command chaining

**Example**:
```go
func TestWizardModel_Commands(t *testing.T) {
    model, _ := NewWizardModel(nil, false)
    
    // Trigger action that creates a command
    msg := tea.KeyMsg{Type: tea.KeyEnter}
    _, cmd := model.Update(msg)
    
    if cmd == nil {
        t.Error("Update() should return a command")
    }
    // Verify command type
}
```

### Layer 5: Integration Tests with Mock Terminal

**Goal**: Test full TUI flows with a mocked terminal

**Approach**:
- Create a mock terminal implementation
- Simulate user interactions (key presses, window resizes)
- Verify complete flows (e.g., create note wizard flow)
- Test error scenarios

**Mock Terminal Helper**:
```go
type MockTerminal struct {
    width  int
    height int
    events []tea.Msg
}

func (m *MockTerminal) SendKey(key string) {
    m.events = append(m.events, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func (m *MockTerminal) SendWindowSize(width, height int) {
    m.events = append(m.events, tea.WindowSizeMsg{Width: width, Height: height})
}

func RunModelWithMockTerminal(t *testing.T, model tea.Model, terminal *MockTerminal) tea.Model {
    for _, event := range terminal.events {
        var cmd tea.Cmd
        model, cmd = model.Update(event)
        if cmd != nil {
            // Handle command (could be async, spinner, etc.)
        }
    }
    return model
}
```

## Test Helper Functions

### Terminal Simulation Helpers

```go
// helpers/terminal.go
package helpers

// SendKeyPress simulates a key press
func SendKeyPress(model tea.Model, key string) (tea.Model, tea.Cmd) {
    return model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

// SendWindowResize simulates terminal resize
func SendWindowResize(model tea.Model, width, height int) (tea.Model, tea.Cmd) {
    return model.Update(tea.WindowSizeMsg{Width: width, Height: height})
}

// RunModelUntilIdle runs a model until no more commands are returned
func RunModelUntilIdle(t *testing.T, model tea.Model, maxIterations int) tea.Model {
    for i := 0; i < maxIterations; i++ {
        var cmd tea.Cmd
        model, cmd = model.Update(tea.TickMsg{})
        if cmd == nil {
            break
        }
    }
    return model
}
```

### State Verification Helpers

```go
// helpers/state.go
package helpers

// AssertState verifies a wizard is in the expected state
func AssertWizardState(t *testing.T, wizard *Wizard, want State) {
    if wizard.State() != want {
        t.Errorf("wizard.State() = %v, want %v", wizard.State(), want)
    }
}

// AssertViewContains verifies view contains expected text
func AssertViewContains(t *testing.T, view string, want string) {
    if !strings.Contains(view, want) {
        t.Errorf("View() does not contain %q", want)
    }
}
```

## Test Coverage Targets

### High Priority (Core Functionality)
- [ ] Model initialization with all configuration options
- [ ] All state transitions (StateMainMenu → StateTemplateSelection → etc.)
- [ ] Key press handling (Enter, Escape, Arrow keys, etc.)
- [ ] Error state handling
- [ ] Window resize handling

### Medium Priority (User Flows)
- [ ] Complete wizard flow (start to finish)
- [ ] Template selection flow
- [ ] Output directory input flow
- [ ] Review screen flow
- [ ] Cancellation flow

### Low Priority (Edge Cases)
- [ ] Very long input text
- [ ] Terminal too small
- [ ] Rapid key presses
- [ ] Concurrent events

## Implementation Plan

### Phase 1: Foundation (Week 1)
1. Create test helper package (`internal/testhelpers/tui/`)
2. Implement basic terminal simulation helpers
3. Add model initialization tests
4. Add basic Update() tests for key presses

### Phase 2: State Machine Testing (Week 2)
1. Test all state transitions
2. Test state history/navigation
3. Test error state handling
4. Add state verification helpers

### Phase 3: View Testing (Week 3)
1. Test View() for all states
2. Test rendering edge cases
3. Test styling application
4. Add view assertion helpers

### Phase 4: Integration Testing (Week 4)
1. Create mock terminal implementation
2. Test complete user flows
3. Test error scenarios
4. Test concurrent events

## Tools and Libraries

### Recommended Testing Libraries
- **Standard `testing` package**: Core testing framework
- **Custom helpers**: Terminal simulation, state verification
- **Mock implementations**: For terminal, file system, etc.

### Coverage Tools
- `go test -cover`: Standard coverage tool
- `go tool cover -html`: HTML coverage reports
- Coverage exclusions for TUI code (if needed)

## Success Criteria

1. **Coverage**: >80% coverage for TUI code
2. **State Transitions**: All state transitions tested
3. **User Flows**: All documented user flows have integration tests
4. **Error Handling**: All error paths tested
5. **Maintainability**: Tests are readable and maintainable

## Notes

- This strategy requires significant effort but provides comprehensive coverage
- Consider this after reaching 80% coverage on business logic
- Some TUI code may be difficult to test (e.g., complex animations)
- Focus on testing behavior, not implementation details
- Use table-driven tests for similar scenarios

## Future Enhancements

- Visual regression testing for TUI rendering
- Performance testing for large datasets
- Accessibility testing for keyboard navigation
- Cross-platform testing (different terminal capabilities)


# Phase 6: REPL Wizard (Interactive Mode)

**Goal**: Replace current TUI with scenario-based REPL wizard

**Duration**: 2-3 weeks  
**Complexity**: High  
**Status**: Not Started

## Prerequisites

- Phase 4 (Non-Interactive Mode) - Entry creation logic must be in place
- Phase 5 (Editor Integration) - Editor integration must be working

## Dependencies on Other Phases

- **Requires**: Phase 4 (Entry creation), Phase 5 (Editor)
- **Enables**: Full interactive mode functionality

## Overview

This phase implements the REPL wizard to replace the current TUI:
- Wizard state machine with transitions
- Back navigation system
- Wizard TUI (Bubble Tea components)
- Review screen with vim-style commands

## Tasks

### 6.1 Wizard State Machine

**Location**: `internal/wizard/wizard.go`

**Requirements**:

- State machine: Main Menu → Output Dir → Title → Tags → Message → File Created → Editor Prompt → Review Screen
- Back navigation (except after file creation)
- Cancel option (deletes file)
- Confirm option (saves file)
- Vim-style commands: `:wq`, `:q`, `:q!`
- State transition validation
- History tracking for back navigation

**State Machine with Transitions**:

```go
type State int

const (
    StateMainMenu State = iota
    StateOutputDir
    StateTitle
    StateTags
    StateMessage
    StateFileCreated
    StateEditorPrompt
    StateReviewScreen
)

type Wizard struct {
    state      State
    prevState  State  // For back navigation
    history    []State // Navigation history
    outputDir  string
    title      string
    tags       []string
    message    string
    filePath   string
    config     *config.Config
    // ...
}

// State transition methods
func (w *Wizard) TransitionTo(newState State) error
func (w *Wizard) CanGoBack() bool
func (w *Wizard) GoBack() error
func (w *Wizard) ValidateState() error
```

**State Transition Rules**:

- Main Menu → Output Dir (when "Create new entry" selected)
- Output Dir → Title (when path entered/validated)
- Title → Tags (when title entered or skipped)
- Tags → Message (when tags entered or skipped)
- Message → File Created (when message entered or skipped)
- File Created → Editor Prompt (always, after file creation)
- Editor Prompt → Review Screen (after editor decision)
- Review Screen → Editor (when "Open editor again" selected)
- Review Screen → Exit (when confirm/cancel/quit selected)

**Back Navigation Rules**:

- Available on: Output Dir, Title, Tags, Message
- NOT available on: File Created, Editor Prompt, Review Screen
- Back from Output Dir → Main Menu
- Back from Title → Output Dir
- Back from Tags → Title
- Back from Message → Tags

**Files to Create**:

- `internal/wizard/wizard.go`
- `internal/wizard/wizard_test.go`
- `internal/wizard/state.go` (state definitions and transitions)
- `internal/wizard/navigation.go` (back navigation logic)

---

### 6.2 Wizard TUI Implementation

**Location**: `internal/wizard/tui.go`

**Requirements**:

- Use Bubble Tea for TUI
- Use Bubbles components (list, textinput, textarea)
- Use Lip Gloss for styling
- Show help footer with keybindings
- Inline error messages
- Spinner for file creation (>200ms)

**Implementation**:

```go
type WizardModel struct {
    wizard *Wizard
    // Bubble Tea components
    list     list.Model
    textarea textarea.Model
    spinner  spinner.Model
    // ...
}

func (m WizardModel) Init() tea.Cmd
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m WizardModel) View() string
```

**Files to Create**:

- `internal/wizard/tui.go`
- `internal/wizard/components.go` (reusable components)

---

### 6.3 Wizard Flow Logic

**Location**: `internal/wizard/flow.go`

**Requirements**:

- Handle user input at each step
- Validate inputs (paths, etc.)
- Create file before editor prompt
- Support multiple editor launches
- Handle review screen actions

**Implementation**:

```go
func (w *Wizard) HandleInput(input string) error
func (w *Wizard) CreateFile() error
func (w *Wizard) LaunchEditor() error
func (w *Wizard) Confirm() error
func (w *Wizard) Cancel() error
```

**Files to Create**:

- `internal/wizard/flow.go`
- `internal/wizard/flow_test.go`

**See Also**: [shared-concerns.md](./shared-concerns.md) for wizard state recovery

---

## Implementation Checklist

- [ ] Wizard state machine
- [ ] State transitions
- [ ] Back navigation
- [ ] Wizard TUI (Bubble Tea)
- [ ] Wizard flow logic
- [ ] File creation in wizard
- [ ] Editor integration in wizard
- [ ] Review screen
- [ ] Vim-style commands
- [ ] Tests for wizard state machine
- [ ] Tests for wizard flow
- [ ] Tests for TUI rendering

## Testing Requirements

### Unit Tests

- `internal/wizard/wizard_test.go`
  - Test state machine
  - Test state transitions
  - Test validation

- `internal/wizard/state_test.go`
  - Test state definitions
  - Test transition rules

- `internal/wizard/navigation_test.go`
  - Test back navigation
  - Test navigation history

- `internal/wizard/flow_test.go`
  - Test flow logic
  - Test file creation
  - Test editor integration

### Integration Tests

- Test wizard end-to-end
- Test all state transitions
- Test back navigation
- Test vim-style commands
- Test review screen actions

## Success Criteria

- ✅ Wizard state machine works correctly
- ✅ All state transitions work as expected
- ✅ Back navigation works correctly
- ✅ Wizard TUI is responsive and user-friendly
- ✅ File creation works in wizard
- ✅ Editor integration works in wizard
- ✅ Review screen works with all actions
- ✅ Vim-style commands work
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 6, proceed to:
- **[Phase 7: Metadata Capture](./07-metadata.md)** - Can be added anytime after Phase 4
- **[Phase 8: Error Handling](./08-errors.md)** - Ongoing throughout all phases

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns


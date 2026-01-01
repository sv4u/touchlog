# touchlog Refactoring Plan: Scenario-Based Architecture - Overview

**Date**: 2025-12-31  
**Branch**: `refactor/scenario-based-architecture`  
**Status**: Planning Phase (Refinement Pass 3/3 - Complete)  
**Plan Status**: Ready for Review and Implementation**

## Executive Summary

This plan outlines the comprehensive refactoring of `touchlog` to align with the scenario-based architecture defined in `SCENARIOS.md` and `SCENARIOS_DIAGRAMS.md`. The refactoring transforms touchlog from a simple template-based TUI editor into a full-featured CLI tool with both interactive (REPL wizard) and non-interactive modes, platform detection, flexible configuration, and editor integration.

The refactoring is organized into **9 phases**, each with its own detailed plan document. This overview provides the high-level context, architecture, and links to phase-specific plans.

## Document Structure

- **[00-overview.md](./00-overview.md)** (this document) - Overview, architecture, and phase summaries
- **[01-foundation.md](./01-foundation.md)** - Phase 1: Foundation & Platform Detection
- **[02-configuration.md](./02-configuration.md)** - Phase 2: Configuration System Enhancement
- **[03-templates.md](./03-templates.md)** - Phase 3: Template System Refactoring
- **[04-non-interactive.md](./04-non-interactive.md)** - Phase 4: Non-Interactive Mode (`touchlog new`)
- **[05-editor.md](./05-editor.md)** - Phase 5: Editor Integration
- **[06-wizard.md](./06-wizard.md)** - Phase 6: REPL Wizard (Interactive Mode)
- **[07-metadata.md](./07-metadata.md)** - Phase 7: Metadata Capture
- **[08-errors.md](./08-errors.md)** - Phase 8: Error Handling & Validation
- **[09-testing.md](./09-testing.md)** - Phase 9: Testing Strategy
- **[shared-concerns.md](./shared-concerns.md)** - Cross-cutting concerns (Dependencies, Security, Performance, etc.)

## Current State Analysis

### Existing Architecture

- **Entry Point**: `cmd/touchlog/main.go` - Simple flag parsing, launches TUI
- **TUI**: `internal/editor/editor.go` - Bubble Tea-based template selection and editing
- **Config**: `internal/config/config.go` - YAML-only, file-based templates
- **Templates**: `internal/template/template.go` - File-based template loading
- **API**: `internal/api/api.go` - Programmatic interface (minimal)

### Key Limitations

1. No platform detection (macOS/Linux/WSL)
2. No subcommand structure (new, config, list, search)
3. No non-interactive mode (`touchlog new`)
4. No version flag support
5. YAML-only configuration (no TOML)
6. File-based templates only (no inline templates)
7. No external editor integration
8. No metadata capture (user, host, git)
9. No REPL wizard (current TUI is different)
10. Limited error handling and validation

## Target Architecture

### High-Level Structure

```text
touchlog
├── Platform Detection (macOS/Linux/WSL only)
├── CLI Router (subcommands + flags)
│   ├── new (non-interactive entry creation)
│   ├── config (config validation)
│   ├── list (future - log listing)
│   ├── search (future - log search)
│   └── [default] (REPL wizard - interactive)
├── Configuration System
│   ├── YAML support (priority)
│   ├── TOML support (later)
│   ├── Precedence: CLI flags > Config file > Defaults
│   └── Strict mode validation
├── Template System
│   ├── Inline templates (in config)
│   ├── File-based templates (existing)
│   └── Variable substitution
├── Entry Creation
│   ├── Non-interactive mode (touchlog new)
│   ├── Interactive mode (REPL wizard)
│   └── Editor integration (internal + external)
├── Metadata Capture
│   ├── User/host information
│   └── Git context
└── Error Handling
    └── Robust validation and safe behavior
```

## Phase Summaries

### [Phase 1: Foundation & Platform Detection](./01-foundation.md)

**Goal**: Establish platform support and basic CLI structure

**Key Deliverables**:

- Platform detection module (macOS/Linux/WSL)
- Version management with build-time injection
- CLI framework setup (Cobra)
- Root command with `--help` and `--version`

**Duration**: 1-2 weeks  
**Complexity**: Medium  
**Dependencies**: None (foundational)

---

### [Phase 2: Configuration System Enhancement](./02-configuration.md)

**Goal**: Support YAML/TOML, inline templates, strict validation, precedence

**Key Deliverables**:

- Config file discovery (XDG, current dir, explicit)
- YAML format enhancements (inline templates)
- Configuration precedence system (CLI > Config > Defaults)
- Strict mode validation

**Duration**: 1 week  
**Complexity**: Medium  
**Dependencies**: Phase 1 (CLI framework)

---

### [Phase 3: Template System Refactoring](./03-templates.md)

**Goal**: Support both inline and file-based templates

**Key Deliverables**:

- Template source interface (inline + file-based)
- Template resolution (inline first, then file)
- Timezone-aware date formatting
- Variable escaping for security

**Duration**: 1 week  
**Complexity**: Medium  
**Dependencies**: Phase 2 (config system)

---

### [Phase 4: Non-Interactive Mode (`touchlog new`)](./04-non-interactive.md)

**Goal**: Implement non-interactive log entry creation

**Key Deliverables**:

- `touchlog new` command with full flag support
- Filename generation with slug algorithm
- Entry creation logic
- Stdin support and UTF-8 validation

**Duration**: 1-2 weeks  
**Complexity**: Medium-High  
**Dependencies**: Phase 2 (config), Phase 3 (templates)

---

### [Phase 5: Editor Integration](./05-editor.md)

**Goal**: Support both internal (Bubble Tea) and external editor handoff

**Key Deliverables**:

- Editor resolution with precedence chain
- External editor launch (non-blocking)
- Internal editor refactoring (Bubble Tea)
- Fallback logic (Option C)

**Duration**: 1 week  
**Complexity**: Medium  
**Dependencies**: Phase 1 (CLI), Phase 2 (config)

---

### [Phase 6: REPL Wizard (Interactive Mode)](./06-wizard.md)

**Goal**: Replace current TUI with scenario-based REPL wizard

**Key Deliverables**:

- Wizard state machine with transitions
- Back navigation system
- Wizard TUI (Bubble Tea components)
- Review screen with vim-style commands

**Duration**: 2-3 weeks  
**Complexity**: High  
**Dependencies**: Phase 4 (entry creation), Phase 5 (editor)

---

### [Phase 7: Metadata Capture](./07-metadata.md)

**Goal**: Add metadata capture (user, host, git context)

**Key Deliverables**:

- Metadata collection (user, host)
- Git context detection
- Metadata integration in templates
- Config and CLI flags for metadata

**Duration**: 3-5 days  
**Complexity**: Low-Medium  
**Dependencies**: Phase 3 (templates), Phase 4 (entry creation)

---

### [Phase 8: Error Handling & Validation](./08-errors.md)

**Goal**: Robust error handling and safe behavior

**Key Deliverables**:

- Error types definition
- Validation functions
- User-friendly error messages
- Comprehensive error wrapping

**Duration**: Ongoing + 1 week polish  
**Complexity**: Low-Medium  
**Dependencies**: All phases (ongoing)

---

### [Phase 9: Testing Strategy](./09-testing.md)

**Goal**: Comprehensive test coverage

**Key Deliverables**:

- Unit tests for all packages
- Integration tests for commands
- Gherkin-style test helpers (future)
- Test coverage >80%

**Duration**: Ongoing + 1-2 weeks comprehensive  
**Complexity**: Medium  
**Dependencies**: All phases (ongoing)

---

## Implementation Order

### Recommended Sequence

1. **Phase 1**: Foundation (Platform, Version, CLI Framework)
   - Establishes base structure
   - Enables all other work

2. **Phase 2**: Configuration System
   - Needed by all other features
   - Can be done in parallel with Phase 3

3. **Phase 3**: Template System
   - Needed by entry creation
   - Can be done in parallel with Phase 2

4. **Phase 4**: Non-Interactive Mode (`touchlog new`)
   - Core functionality
   - Can use existing template system

5. **Phase 5**: Editor Integration
   - Needed by both new command and wizard
   - Can be done in parallel with Phase 6

6. **Phase 6**: REPL Wizard
   - Interactive mode
   - Depends on entry creation

7. **Phase 7**: Metadata Capture
   - Enhancement feature
   - Can be added anytime after Phase 4

8. **Phase 8**: Error Handling
   - Ongoing throughout all phases
   - Final polish

9. **Phase 9**: Testing
   - Ongoing throughout all phases
   - Final comprehensive coverage

## File Structure After Refactoring

```text
touchlog/
├── cmd/
│   └── touchlog/
│       ├── main.go
│       └── commands/
│           ├── root.go
│           ├── new.go
│           ├── config.go
│           ├── list.go (future)
│           └── search.go (future)
├── internal/
│   ├── platform/          # Phase 1
│   ├── version/            # Phase 1
│   ├── config/             # Phase 2
│   ├── template/           # Phase 3
│   ├── entry/              # Phase 4
│   ├── editor/             # Phase 5
│   ├── wizard/             # Phase 6
│   ├── metadata/           # Phase 7
│   ├── validation/         # Phase 8
│   ├── errors/             # Phase 8
│   └── xdg/                # (existing)
├── features/               # Phase 9 (future - Gherkin-style tests)
├── go.mod
├── go.sum
├── SCENARIOS.md
├── SCENARIOS_DIAGRAMS.md
└── README.md
```

## Success Criteria

### Functional

- ✅ All scenarios in `SCENARIOS.md` are implemented
- ✅ Platform detection works (macOS/Linux/WSL)
- ✅ Non-interactive mode works (`touchlog new`)
- ✅ Interactive wizard works (`touchlog` with no args)
- ✅ Editor integration works (internal + external)
- ✅ Configuration system works (YAML, precedence)
- ✅ Template system works (inline + file-based)
- ✅ Metadata capture works (user, host, git)

### Quality

- ✅ Comprehensive test coverage (>80%)
- ✅ All tests pass
- ✅ No linter errors
- ✅ Documentation updated
- ✅ Examples work

### Performance

- ✅ Fast startup (<100ms for simple commands)
- ✅ Responsive TUI
- ✅ Efficient file operations

## Risk Assessment

### High Risk

- **REPL Wizard Complexity**: Complex state machine, may need iteration
- **Editor Integration**: Cross-platform editor launching can be tricky
- **Template System**: Supporting both inline and file-based may cause conflicts

### Medium Risk

- **Configuration Precedence**: Need to ensure correct order
- **Filename Generation**: Collision handling needs to be robust
- **Platform Detection**: WSL detection may be platform-specific

### Low Risk

- **Version Management**: Standard Go build-time injection
- **Metadata Capture**: Standard OS APIs
- **Error Handling**: Standard Go error patterns

## Timeline Estimate

### Phase-by-Phase Estimates

- **Phase 1**: 1-2 weeks (Medium complexity)
- **Phase 2**: 1 week (Medium complexity)
- **Phase 3**: 1 week (Medium complexity)
- **Phase 4**: 1-2 weeks (Medium-High complexity)
- **Phase 5**: 1 week (Medium complexity)
- **Phase 6**: 2-3 weeks (High complexity)
- **Phase 7**: 3-5 days (Low-Medium complexity)
- **Phase 8**: Ongoing + 1 week polish (Low-Medium complexity)
- **Phase 9**: Ongoing + 1-2 weeks comprehensive (Medium complexity)

### Total Estimate

- **Minimum**: 8-10 weeks
- **Realistic**: 10-12 weeks
- **With Buffer**: 12-15 weeks

## Questions & Decisions Log

### Answered Questions

1. ✅ Version: Build-time injection
2. ✅ Config: YAML first, TOML later
3. ✅ Templates: Both inline and file-based
4. ✅ TUI: Replace with new wizard
5. ✅ Editor: Both internal and external (Option C)
6. ✅ Metadata: New feature
7. ✅ Platform: First step
8. ✅ Testing: Go tests required, Gherkin later

### Decisions Made

- Use `cobra` for CLI framework
- Maintain backward compatibility where possible
- Prioritize YAML over TOML
- Support both template types
- Replace current TUI with wizard
- **Editor Integration (Option C)**: Both wizard and non-interactive modes can use either editor type, with preference for external and fallback to internal

## Cross-Cutting Concerns

For details on shared concerns that apply across all phases, see:

- **[shared-concerns.md](./shared-concerns.md)** - Dependencies, Migration Strategy, CI/CD Integration, Documentation, Performance, Security, Edge Cases, Logging, Rollback Strategies

## Next Steps

1. **Review this overview** and phase plans
2. **Prioritize phases** based on requirements
3. **Set up development environment** (branch, dependencies)
4. **Begin Phase 1** (Foundation) - see [01-foundation.md](./01-foundation.md)
5. **Iterate** based on feedback

## References

- `SCENARIOS.md` - Complete scenario definitions
- `SCENARIOS_DIAGRAMS.md` - Visual flow diagrams
- Current codebase structure
- Go best practices
- Bubble Tea documentation
- Cobra CLI framework documentation

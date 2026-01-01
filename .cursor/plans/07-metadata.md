# Phase 7: Metadata Capture

**Goal**: Add metadata capture (user, host, git context)

**Duration**: 3-5 days  
**Complexity**: Low-Medium  
**Status**: Not Started

## Prerequisites

- Phase 3 (Template System) - Template system must support metadata variables
- Phase 4 (Non-Interactive Mode) - Entry creation must be in place

## Dependencies on Other Phases

- **Requires**: Phase 3 (Templates), Phase 4 (Entry creation)
- **Enables**: Enhanced template variables with metadata

## Overview

This phase adds metadata capture functionality:
- Metadata collection (user, host)
- Git context detection
- Metadata integration in templates
- Config and CLI flags for metadata

## Tasks

### 7.1 Metadata Collection

**Location**: `internal/metadata/metadata.go`

**Requirements**:

- Capture username (configurable)
- Capture hostname (configurable)
- Capture git context (branch, commit) when in git repo
- Config flags: `include_user`, `include_host`
- CLI flag: `--include-git`

**Implementation**:

```go
type Metadata struct {
    User string
    Host string
    Git  *GitContext
}

type GitContext struct {
    Branch string
    Commit string
}

func CollectMetadata(config *config.Config, includeGit bool) (*Metadata, error)
func GetUsername() (string, error)
func GetHostname() (string, error)
func GetGitContext() (*GitContext, error)
```

**Files to Create**:

- `internal/metadata/metadata.go`
- `internal/metadata/metadata_test.go`
- `internal/metadata/git.go`

---

### 7.2 Metadata Integration

**Location**: `internal/template/template.go` (extend)

**Requirements**:

- Add metadata variables to template vars
- Support `{{user}}`, `{{host}}`, `{{branch}}`, `{{commit}}`
- Include in template processing

**Files to Modify**:

- `internal/template/template.go`

**See Also**: Phase 3 (Template system) for variable system

---

## Implementation Checklist

- [ ] Metadata collection (user, host)
- [ ] Git context detection
- [ ] Metadata integration in templates
- [ ] Config flags for metadata
- [ ] CLI flag for git context
- [ ] Tests for metadata collection
- [ ] Tests for git context

## Testing Requirements

### Unit Tests

- `internal/metadata/metadata_test.go`
  - Test username capture
  - Test hostname capture
  - Test config flags

- `internal/metadata/git_test.go`
  - Test git context detection
  - Test git repo detection
  - Test branch/commit extraction

### Integration Tests

- Test metadata in templates
- Test metadata in entry creation
- Test metadata in wizard

## Success Criteria

- ✅ Metadata collection works (user, host)
- ✅ Git context detection works
- ✅ Metadata variables work in templates
- ✅ Config flags work
- ✅ CLI flag works
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 7, proceed to:
- **[Phase 8: Error Handling](./08-errors.md)** - Ongoing throughout all phases
- **[Phase 9: Testing](./09-testing.md)** - Ongoing throughout all phases

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns


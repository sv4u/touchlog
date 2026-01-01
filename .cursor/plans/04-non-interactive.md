# Phase 4: Non-Interactive Mode (`touchlog new`)

**Goal**: Implement non-interactive log entry creation

**Duration**: 1-2 weeks  
**Complexity**: Medium-High  
**Status**: Not Started

## Prerequisites

- Phase 2 (Configuration System) - Config system must be in place
- Phase 3 (Template System) - Template system must be working

## Dependencies on Other Phases

- **Requires**: Phase 2 (Config), Phase 3 (Templates)
- **Enables**: Phase 6 (Wizard can reuse entry creation logic)

## Overview

This phase implements the `touchlog new` command for non-interactive log entry creation:
- Full flag support (`--message`, `--title`, `--tag`, `--output`, `--template`, `--edit`, `--editor`, `--overwrite`, `--stdin`)
- Filename generation with slug algorithm
- Entry creation logic
- Stdin support and UTF-8 validation

## Tasks

### 4.1 New Command Implementation

**Location**: `cmd/touchlog/commands/new.go`

**Requirements**:

- Flags: `--message`, `--title`, `--tag`, `--output`, `--template`, `--edit`, `--editor`, `--overwrite`, `--stdin`
- Create log file with proper naming: `YYYY-MM-DD_slug.md`
- Handle collisions with numeric suffix
- Support `--overwrite` flag
- Support reading from stdin with `--stdin`
- Validate UTF-8 input

**Implementation**:

```go
var newCmd = &cobra.Command{
    Use:   "new",
    Short: "Create a new log entry",
    RunE:  runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
    // Load config
    // Parse flags
    // Validate inputs
    // Generate filename
    // Apply template
    // Write file
    // Optionally launch editor
    return nil
}
```

**Files to Create**:

- `cmd/touchlog/commands/new.go`
- `internal/entry/entry.go` (entry creation logic)
- `internal/entry/entry_test.go`

**Files to Modify**:

- `cmd/touchlog/commands/root.go` (add new command)

---

### 4.2 Filename Generation

**Location**: `internal/entry/filename.go`

**Requirements**:

- Format: `YYYY-MM-DD_slug.md` (e.g., `2025-12-31_standup-notes.md`)
- Slug generation algorithm:
  - Use title if provided and non-empty
  - Fall back to first line of message if title is empty
  - Convert to lowercase
  - Replace spaces and special chars with hyphens
  - Remove consecutive hyphens
  - Trim hyphens from start/end
  - Limit to 50 characters (truncate if needed)
  - If resulting slug is empty, use `untitled`
- Handle collisions: `_1.md`, `_2.md`, etc. (numeric suffix)
- Stable and deterministic (same input = same output)
- Timezone-aware date formatting (use local timezone from config or system)

**Slug Generation Details**:

```go
func GenerateSlug(title string, message string) string {
    // 1. Prefer title, fallback to first line of message
    source := title
    if source == "" {
        lines := strings.Split(message, "\n")
        if len(lines) > 0 {
            source = strings.TrimSpace(lines[0])
        }
    }
    
    // 2. Convert to lowercase
    source = strings.ToLower(source)
    
    // 3. Replace spaces and special chars with hyphens
    // 4. Remove consecutive hyphens
    // 5. Trim hyphens
    // 6. Limit to 50 chars
    // 7. Handle empty case
}
```

**Implementation**:

```go
func GenerateFilename(outputDir string, title string, message string, timestamp time.Time, timezone *time.Location) (string, error)
func GenerateSlug(title string, message string) string
func FindAvailableFilename(basePath string) (string, error)
func FormatDate(t time.Time, tz *time.Location) string
```

**Edge Cases**:

- Empty title and message → use `untitled`
- Very long title/message → truncate to 50 chars
- Special characters → sanitize appropriately
- Multiple files with same timestamp → use numeric suffix
- Invalid characters in slug → replace with hyphens

**Files to Create**:

- `internal/entry/filename.go`
- `internal/entry/filename_test.go`
- `internal/entry/slug.go` (slug generation logic)

**See Also**: [shared-concerns.md](./shared-concerns.md) for edge cases and security

---

### 4.3 Entry Creation Logic

**Location**: `internal/entry/entry.go`

**Requirements**:

- Collect all entry data (title, message, tags, metadata)
- Apply template
- Write to file
- Handle errors gracefully
- Support overwrite mode

**Implementation**:

```go
type Entry struct {
    Title    string
    Message  string
    Tags     []string
    Metadata *Metadata
    Date     time.Time
}

func CreateEntry(entry *Entry, config *config.Config, outputDir string, overwrite bool) (string, error)
```

**Files to Create**:

- `internal/entry/entry.go`
- `internal/entry/entry_test.go`

**See Also**: [shared-concerns.md](./shared-concerns.md) for file system security and rollback strategies

---

## Implementation Checklist

- [ ] `touchlog new` command
- [ ] Flag parsing (--message, --title, --tag, --output, --template, --edit, --editor, --overwrite, --stdin)
- [ ] Filename generation with slug
- [ ] Collision handling
- [ ] Entry creation logic
- [ ] UTF-8 validation
- [ ] Stdin support
- [ ] Tests for new command
- [ ] Tests for filename generation
- [ ] Tests for entry creation

## Testing Requirements

### Unit Tests

- `internal/entry/entry_test.go`
  - Test entry creation
  - Test template application
  - Test overwrite mode

- `internal/entry/filename_test.go`
  - Test filename generation
  - Test slug generation
  - Test collision handling
  - Test timezone handling

- `internal/entry/slug_test.go`
  - Test slug generation algorithm
  - Test edge cases (empty, long, special chars)

### Integration Tests

- `cmd/touchlog/commands/new_test.go`
  - Test `touchlog new` command end-to-end
  - Test all flags
  - Test stdin support
  - Test editor integration (stub for Phase 5)

## Success Criteria

- ✅ `touchlog new` command works with all flags
- ✅ Filename generation works correctly
- ✅ Slug generation handles all edge cases
- ✅ Entry creation works with templates
- ✅ Stdin support works
- ✅ UTF-8 validation works
- ✅ All tests pass
- ✅ No linter errors

## Next Phase

After completing Phase 4, proceed to:
- **[Phase 5: Editor Integration](./05-editor.md)** - Can be done in parallel with Phase 6
- **[Phase 6: REPL Wizard](./06-wizard.md)** - Requires Phase 4 (entry creation)

## References

- [Overview](./00-overview.md) - Overall plan and architecture
- [shared-concerns.md](./shared-concerns.md) - Cross-cutting concerns


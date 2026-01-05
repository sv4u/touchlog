# Bug Report: PR 24 - Scenario-Based Architecture Refactoring

**Date**: 2026-01-04  
**Branch**: `refactor/scenario-based-architecture`  
**Status**: All tests passing, but bugs identified in collision handling logic

## Executive Summary

This PR implements a comprehensive refactoring of `touchlog` into a scenario-based architecture with Cobra CLI and modular components. While all tests are passing, there are **critical bugs** in the filename collision handling logic that prevent the system from working as designed. The bugs affect the core functionality of creating multiple entries with the same title on the same day.

## Test Status

✅ **All tests passing** (including integration tests)  
⚠️ **Test warnings indicate expected behavior not working**: `TestIntegration_NewCommand_MultipleEntriesSameDay` logs a warning that collision handling isn't working as expected

---

## Bug #1: Filename Collision Handling Not Working

**Severity**: 🔴 **CRITICAL**  
**Location**: `internal/entry/entry.go`, lines 79-91  
**Impact**: Users cannot create multiple entries with the same title on the same day without using `--overwrite`

### Description

The `CreateEntry` function checks if the base filename exists **before** calling `GenerateFilename`, which prevents automatic collision handling from working. This means:

- If `2026-01-04_first.md` exists, the function returns an error immediately
- It never calls `GenerateFilename()` which would create `2026-01-04_first_1.md` automatically
- Users must manually use `--overwrite` or choose different titles

### Current Code (Buggy)

```go
// Check if base file exists and handle overwrite
if !overwrite {
    // Check if base file exists
    if _, err := os.Stat(basePath); err == nil {
        return "", fmt.Errorf("file already exists: %s (use --overwrite to overwrite)", basePath)
    }
    // If base doesn't exist, generate filename with collision handling
    filename, err := GenerateFilename(expandedDir, entry.Title, entry.Message, entry.Date, tz)
    if err != nil {
        return "", fmt.Errorf("failed to generate filename: %w", err)
    }
    basePath = filename
}
```

### Expected Behavior

According to the design (see `SCENARIOS.md` and `SCENARIOS_DIAGRAMS.md`), when a file already exists:

- **Without `--overwrite`**: Automatically append numeric suffix (`_1`, `_2`, etc.)
- **With `--overwrite`**: Overwrite the base filename

### Root Cause

The logic flow is inverted:

1. ❌ **Current**: Check if base exists → if yes, error; if no, call `GenerateFilename`
2. ✅ **Should be**: Always call `GenerateFilename` first (handles collisions) → if `overwrite=true`, use base filename instead

### Evidence

1. **Integration Test Warning**: `TestIntegration_NewCommand_MultipleEntriesSameDay` logs:

   ```
   Second command reported collision (may be expected), output: Error: failed to create entry: file already exists
   Expected at least 2 markdown files, got 1. Collision handling may prevent overwriting without --overwrite flag.
   ```

2. **Test Comment**: The test explicitly notes this is "acceptable behavior" but it contradicts the design docs

3. **Design Documents**: `SCENARIOS_DIAGRAMS.md` line 175 shows collision handling should add numeric suffix automatically

### Fix Required

**Always call `GenerateFilename` first** (it handles collisions internally), then override with base filename only if `overwrite=true`:

```go
// Always generate filename with collision handling first
filename, err := GenerateFilename(expandedDir, entry.Title, entry.Message, entry.Date, tz)
if err != nil {
    return "", fmt.Errorf("failed to generate filename: %w", err)
}

// If overwrite is true, use the base filename instead (overwrites existing file)
if overwrite {
    baseFilename := FormatDate(entry.Date, tz) + "_" + GenerateSlug(entry.Title, entry.Message) + ".md"
    filename = filepath.Join(expandedDir, baseFilename)
}
```

---

## Bug #2: Overwrite Logic Edge Case

**Severity**: 🟡 **MEDIUM**  
**Location**: `internal/entry/entry.go`, lines 79-94  
**Impact**: When `overwrite=true` but base file doesn't exist, behavior is inconsistent

### Description

When `overwrite=true`:

- If base file exists: ✅ Works correctly (overwrites)
- If base file doesn't exist but numbered files do (e.g., `2026-01-04_first_1.md` exists): ❌ Creates new numbered file instead of using base filename

### Current Behavior

```go
if !overwrite {
    // ... collision handling ...
} else {
    // Use basePath directly
    filename = basePath
}
```

### Expected Behavior

When `overwrite=true`, the user explicitly wants to overwrite the **base filename**:

- If `2026-01-04_first.md` exists → overwrite it
- If `2026-01-04_first.md` doesn't exist → create it (even if `_1.md`, `_2.md` exist)

### Edge Case Scenario

1. User creates entry → `2026-01-04_first.md` created
2. User creates another entry (no overwrite) → `2026-01-04_first_1.md` created
3. User runs with `--overwrite` → Should overwrite `2026-01-04_first.md`, not create `_2.md`

### Fix Required

The fix for Bug #1 will also address this, but we need to ensure:

- When `overwrite=true`, always use base filename (regardless of numbered files)
- When `overwrite=false`, always use `GenerateFilename` (handles collisions)

---

## Bug #3: Race Condition in File Creation

**Severity**: 🟡 **MEDIUM**  
**Location**: `internal/entry/entry.go`, `internal/entry/filename.go`  
**Impact**: Concurrent `touchlog new` commands might create files with same name

### Description

There's a potential race condition between:

1. `FindAvailableFilename` checking if a file exists
2. Another process creating that file
3. The current process writing to the same filename

### Current Implementation

```go
func FindAvailableFilename(basePath string) (string, error) {
    if _, err := os.Stat(basePath); os.IsNotExist(err) {
        return basePath, nil
    }
    // Try suffixes...
}
```

### Problem

Between `os.Stat` check and file creation, another process could:

1. Check the same filename
2. Both determine it's available
3. Both try to create it
4. One fails or overwrites the other

### Mitigation

The code uses atomic writes (write to `.tmp` then rename), which helps, but the filename generation itself isn't atomic.

### Recommended Fix

1. **Option A**: Use file locking (complex, may not be necessary)
2. **Option B**: Retry logic if file creation fails due to race condition
3. **Option C**: Document as known limitation (acceptable for CLI tool)

**Recommendation**: Option C (document limitation) - This is a CLI tool, not a high-concurrency service. The atomic write pattern already mitigates most issues.

---

## Bug #4: Template Resolution Edge Case

**Severity**: 🟢 **LOW**  
**Location**: `cmd/touchlog/commands/new.go`, lines 165-169  
**Impact**: Template override logic might not work as expected in all cases

### Description

When `--template` flag is provided, the code sets both `cfg.DefaultTemplate` and `cfg.Template.Name`:

```go
if templateName != "" {
    cfg.DefaultTemplate = templateName
    cfg.Template.Name = templateName
}
```

But later, `CreateEntry` only checks `cfg.GetDefaultTemplate()`:

```go
templateName := ""
if cfg != nil {
    templateName = cfg.GetDefaultTemplate()
}
```

### Potential Issue

If `cfg.Template.Name` is set but `cfg.DefaultTemplate` is empty, the template might not be resolved correctly. However, this is likely fine since `GetDefaultTemplate()` should handle this.

### Verification Needed

Check if there are any edge cases where `Template.Name` and `DefaultTemplate` can be out of sync.

---

## Bug #5: Inconsistent Error Messages

**Severity**: 🟢 **LOW**  
**Location**: Multiple files  
**Impact**: User experience - error messages could be more consistent

### Description

Error messages use different formats:

- `"file already exists: %s (use --overwrite to overwrite)"`
- `"failed to create entry: %w"`
- `"invalid output directory: %w"`

### Recommendation

Standardize error message format for consistency. Consider using structured errors from `internal/errors/errors.go`.

---

## Additional Findings

### ✅ Good Practices Found

1. **Atomic file writes**: Uses temp file + rename pattern
2. **Comprehensive validation**: `ValidateOutputDir`, `ValidateUTF8`, etc.
3. **Error wrapping**: Good use of `fmt.Errorf` with `%w`
4. **Test coverage**: Extensive unit and integration tests
5. **Read-only config search**: `ConfigFilePathReadOnly()` prevents side effects

### ⚠️ Areas for Improvement

1. **Documentation**: Some edge cases not fully documented
2. **Error types**: Could use structured error types more consistently
3. **Concurrency**: Race condition documentation needed

---

## Recommended Fix Priority

1. **🔴 CRITICAL**: Bug #1 (Collision Handling) - **Fix immediately**
2. **🟡 MEDIUM**: Bug #2 (Overwrite Edge Case) - Fix with Bug #1
3. **🟡 MEDIUM**: Bug #3 (Race Condition) - Document or add retry logic
4. **🟢 LOW**: Bug #4 (Template Resolution) - Verify and fix if needed
5. **🟢 LOW**: Bug #5 (Error Messages) - Polish for consistency

---

## Testing Recommendations

After fixing Bug #1 and #2, add/update tests:

1. **Test**: Multiple entries same day, same title (no overwrite) → should create `_1.md`, `_2.md`
2. **Test**: Overwrite when base file exists → should overwrite base file
3. **Test**: Overwrite when base file doesn't exist but numbered files do → should create base file
4. **Test**: Concurrent file creation (if implementing retry logic)

---

## Implementation Plan

### Phase 1: Fix Critical Bug (#1)

1. Modify `CreateEntry` to always call `GenerateFilename` first
2. Add logic to override with base filename when `overwrite=true`
3. Update tests to verify correct behavior
4. Run integration tests to ensure no regressions

### Phase 2: Fix Overwrite Edge Case (#2)

1. Ensure overwrite logic uses base filename regardless of numbered files
2. Add test cases for edge scenarios
3. Verify behavior matches design docs

### Phase 3: Address Race Condition (#3)

1. Document known limitation OR
2. Implement retry logic for file creation failures
3. Add tests for concurrent scenarios (if implementing retry)

### Phase 4: Polish (#4, #5)

1. Verify template resolution edge cases
2. Standardize error messages
3. Update documentation

---

## Files to Modify

1. **`internal/entry/entry.go`** - Fix collision handling logic (lines 75-94)
2. **`internal/entry/entry_test.go`** - Add/update tests for collision scenarios
3. **`integration/integration_test.go`** - Update `TestIntegration_NewCommand_MultipleEntriesSameDay` to expect correct behavior
4. **Documentation** - Update if race condition is documented limitation

---

## Questions for Review

1. **Overwrite semantics**: When `--overwrite` is used, should it:
   - Always overwrite base filename (even if numbered files exist)? ✅ **Recommended**
   - Overwrite the most recent numbered file?
   - Overwrite all files with same base name?

2. **Race condition handling**: Should we:
   - Document as known limitation? ✅ **Recommended for CLI tool**
   - Implement file locking?
   - Add retry logic?

3. **Backward compatibility**: Are there users relying on current behavior (error on collision)? If so, we may need a flag to control behavior.

---

## Conclusion

The PR is well-structured and comprehensive, but **Bug #1 is critical** and prevents the system from working as designed. The fix is straightforward and should be implemented before merging. Bugs #2-5 are lower priority but should be addressed for a polished release.

**Recommendation**: Fix Bug #1 and #2 before merging. Address Bugs #3-5 in follow-up PRs or document as known limitations.

# Bug Fix Implementation Plan: PR 24

**Date**: 2026-01-04  
**Related**: [BUG_REPORT_PR24.md](./BUG_REPORT_PR24.md)

## Overview

This document provides a detailed, step-by-step implementation plan to fix the bugs identified in PR 24. The fixes are prioritized by severity and impact.

---

## Fix #1: Filename Collision Handling (CRITICAL)

### Problem Summary

The `CreateEntry` function checks if base filename exists before calling `GenerateFilename`, preventing automatic collision handling.

### Current Code

```go
// internal/entry/entry.go, lines 75-94
// Generate base filename (without collision handling)
baseFilename := FormatDate(entry.Date, tz) + "_" + GenerateSlug(entry.Title, entry.Message) + ".md"
basePath := filepath.Join(expandedDir, baseFilename)

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

// Use base path for overwrite, or generated path for new file
filename := basePath
```

### Fixed Code

```go
// internal/entry/entry.go, lines 75-94
// Always generate filename with collision handling first
// This ensures we get an available filename (handles numbered suffixes automatically)
filename, err := GenerateFilename(expandedDir, entry.Title, entry.Message, entry.Date, tz)
if err != nil {
    return "", fmt.Errorf("failed to generate filename: %w", err)
}

// If overwrite is true, use the base filename instead (overwrites existing base file)
if overwrite {
    baseFilename := FormatDate(entry.Date, tz) + "_" + GenerateSlug(entry.Title, entry.Message) + ".md"
    filename = filepath.Join(expandedDir, baseFilename)
    // Note: We don't check if base file exists here - overwrite means overwrite
}
```

### Step-by-Step Implementation

1. **Backup current code** (already in git, but good practice)
2. **Modify `CreateEntry` function**:
   - Remove the `if !overwrite` block that checks for base file existence
   - Move `GenerateFilename` call to the top (before overwrite check)
   - Add `if overwrite` block that overrides filename with base filename
3. **Test the fix**:

   ```bash
   go test ./internal/entry -v
   go test ./integration -v -run TestIntegration_NewCommand_MultipleEntriesSameDay
   ```

4. **Update integration test**:
   - Remove the "may be expected" comment in `TestIntegration_NewCommand_MultipleEntriesSameDay`
   - Update test to expect 2 files to be created successfully
   - Remove the fallback logic that tries to create a third file

### Expected Test Results After Fix

**Before Fix**:

```
TestIntegration_NewCommand_MultipleEntriesSameDay
    integration_test.go:1231: Second command reported collision (may be expected)
    integration_test.go:1276: Expected at least 2 markdown files, got 1
--- PASS: TestIntegration_NewCommand_MultipleEntriesSameDay (0.44s)
```

**After Fix**:

```
TestIntegration_NewCommand_MultipleEntriesSameDay
    integration_test.go:1297: Created 2 unique files: [2026-01-04_first.md, 2026-01-04_first_1.md]
--- PASS: TestIntegration_NewCommand_MultipleEntriesSameDay (0.44s)
```

### Edge Cases to Test

1. ✅ Create entry → Create another with same title → Should create `_1.md`
2. ✅ Create entry → Create another with same title → Create another → Should create `_2.md`
3. ✅ Create entry → Use `--overwrite` → Should overwrite base file
4. ✅ Create entry → Create `_1.md` → Use `--overwrite` → Should overwrite base file (not create `_2.md`)

---

## Fix #2: Overwrite Logic Edge Case (MEDIUM)

### Problem Summary

When `overwrite=true` but base file doesn't exist (only numbered files exist), the current logic would create a new numbered file instead of using the base filename.

### Solution

The fix for Bug #1 already addresses this! When `overwrite=true`, we explicitly set the filename to the base filename, regardless of what numbered files exist.

### Additional Test Case

Add this test to `internal/entry/entry_test.go`:

```go
func TestCreateEntryOverwrite_WhenBaseFileDoesNotExist(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "touchlog-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)

    cfg := config.CreateDefaultConfig()
    cfg.NotesDirectory = tmpDir
    cfg.DefaultTemplate = "daily"

    // Setup template...
    // (same as TestCreateEntryOverwrite)

    entry := &Entry{
        Title:    "Test",
        Message:  "First entry",
        Tags:     []string{},
        Metadata: nil,
        Date:     time.Now(),
    }

    // Create first entry (will be base file)
    filePath1, err := CreateEntry(entry, cfg, tmpDir, false)
    if err != nil {
        t.Fatalf("Failed to create first entry: %v", err)
    }

    // Delete base file, but create a numbered file
    os.Remove(filePath1)
    
    // Create numbered file manually
    entry.Message = "Numbered entry"
    filePath2, err := CreateEntry(entry, cfg, tmpDir, false)
    if err != nil {
        t.Fatalf("Failed to create numbered entry: %v", err)
    }
    
    // Verify numbered file was created
    if !strings.Contains(filepath.Base(filePath2), "_1") {
        t.Errorf("Expected numbered file, got: %s", filePath2)
    }

    // Now use overwrite - should create base file (not _2)
    entry.Message = "Overwrite entry"
    filePath3, err := CreateEntry(entry, cfg, tmpDir, true)
    if err != nil {
        t.Fatalf("Failed to create with overwrite: %v", err)
    }

    // Should be base filename (not numbered)
    baseName := filepath.Base(filePath3)
    if strings.Contains(baseName, "_1") || strings.Contains(baseName, "_2") {
        t.Errorf("Overwrite should create base filename, got: %s", baseName)
    }

    // Verify content
    content, err := os.ReadFile(filePath3)
    if err != nil {
        t.Fatalf("Failed to read file: %v", err)
    }
    if !strings.Contains(string(content), "Overwrite entry") {
        t.Error("File content was not updated")
    }
}
```

---

## Fix #3: Race Condition Documentation (MEDIUM)

### Problem Summary

Potential race condition between filename availability check and file creation.

### Solution

**Option A: Document as Known Limitation** (Recommended)

Add documentation to `internal/entry/entry.go`:

```go
// CreateEntry creates a new log entry file with the given entry data
// It applies the template, generates the filename, and writes the file
// Returns the path to the created file
//
// Note: This function uses atomic file writes (write to .tmp then rename)
// to prevent partial file writes. However, there is a small window for
// race conditions if multiple processes create entries with the same title
// simultaneously. In practice, this is rare for CLI usage and the atomic
// write pattern mitigates most issues.
func CreateEntry(entry *Entry, cfg *config.Config, outputDir string, overwrite bool) (string, error) {
    // ... implementation
}
```

**Option B: Add Retry Logic** (If needed)

If we want to handle race conditions, add retry logic:

```go
// Generate filename with retry on collision
maxRetries := 3
var filename string
var err error

for i := 0; i < maxRetries; i++ {
    filename, err = GenerateFilename(expandedDir, entry.Title, entry.Message, entry.Date, tz)
    if err != nil {
        return "", fmt.Errorf("failed to generate filename: %w", err)
    }

    // Try to create file atomically
    tempFile := filename + ".tmp"
    if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
        if i < maxRetries-1 {
            // Retry with new filename
            time.Sleep(10 * time.Millisecond) // Small delay
            continue
        }
        return "", fmt.Errorf("failed to write file: %w", err)
    }

    // Try to rename (atomic)
    if err := os.Rename(tempFile, filename); err != nil {
        _ = os.Remove(tempFile) // Clean up
        if i < maxRetries-1 {
            // File might have been created by another process, retry
            time.Sleep(10 * time.Millisecond)
            continue
        }
        return "", fmt.Errorf("failed to rename file: %w", err)
    }

    // Success!
    return filename, nil
}

return "", fmt.Errorf("failed after %d retries", maxRetries)
```

**Recommendation**: Start with Option A (documentation). Only implement Option B if users report issues.

---

## Fix #4: Template Resolution Verification (LOW)

### Problem Summary

Need to verify template resolution works correctly in all edge cases.

### Verification Steps

1. **Test template precedence**:

   ```go
   // Test that CLI flag overrides config
   cfg.DefaultTemplate = "config-template"
   // Set via CLI flag
   cfg.DefaultTemplate = "cli-template"
   // Verify it's used
   ```

2. **Test template fallback**:

   ```go
   // Test that empty template name falls back to "daily"
   cfg.DefaultTemplate = ""
   // Should use "daily" as fallback
   ```

3. **Test inline vs file-based**:

   ```go
   // Test that inline templates take precedence over file-based
   cfg.InlineTemplates = map[string]string{"daily": "inline content"}
   // Should use inline, not file
   ```

### Current Code Review

Looking at `cmd/touchlog/commands/new.go` lines 165-169:

```go
// Override template if specified
if templateName != "" {
    cfg.DefaultTemplate = templateName
    cfg.Template.Name = templateName
}
```

And `internal/entry/entry.go` lines 96-104:

```go
// Get template name (use default if not specified)
templateName := ""
if cfg != nil {
    templateName = cfg.GetDefaultTemplate()
}
// If still empty, use "daily" as fallback
if templateName == "" {
    templateName = "daily"
}
```

**Analysis**: This looks correct. `GetDefaultTemplate()` should return the CLI-overridden value. The fallback to "daily" is also correct.

### Action

✅ **No fix needed** - Current implementation is correct. Add a test to verify edge cases if not already covered.

---

## Fix #5: Error Message Standardization (LOW)

### Problem Summary

Error messages use inconsistent formats.

### Current Error Messages

1. `"file already exists: %s (use --overwrite to overwrite)"`
2. `"failed to create entry: %w"`
3. `"invalid output directory: %w"`
4. `"failed to generate filename: %w"`

### Standardization Plan

Use structured error types from `internal/errors/errors.go` where possible:

```go
// Instead of:
return "", fmt.Errorf("file already exists: %s (use --overwrite to overwrite)", basePath)

// Use:
return "", fmt.Errorf("%w: %s (use --overwrite to overwrite)", errors.ErrFileExists, basePath)
```

### Implementation

1. **Check if `ErrFileExists` exists** in `internal/errors/errors.go`
   - ✅ It exists (line 55)
2. **Update `CreateEntry` to use structured errors**:

   ```go
   // Before fix, this error won't occur (Bug #1 fix)
   // But if we want to keep explicit check for overwrite case:
   if overwrite {
       // Check if base file exists for better error message
       if _, err := os.Stat(basePath); err == nil {
           // File exists, overwrite will work - continue
       }
   }
   ```

3. **Standardize other error messages**:
   - Use `%w` for error wrapping consistently
   - Use structured errors where available
   - Keep user-friendly messages

### Priority

🟢 **Low priority** - Can be done as polish after critical fixes.

---

## Testing Strategy

### Unit Tests

1. **`internal/entry/entry_test.go`**:
   - ✅ `TestCreateEntry` - Verify basic creation
   - ✅ `TestCreateEntryOverwrite` - Verify overwrite works
   - ➕ Add `TestCreateEntry_MultipleEntriesSameTitle` - Verify collision handling
   - ➕ Add `TestCreateEntryOverwrite_WhenBaseFileDoesNotExist` - Verify edge case

2. **`internal/entry/filename_test.go`**:
   - ✅ `TestGenerateFilename` - Verify filename generation
   - ✅ `TestFindAvailableFilename` - Verify collision handling
   - ✅ Tests look comprehensive

### Integration Tests

1. **`integration/integration_test.go`**:
   - ✅ `TestIntegration_NewCommand_MultipleEntriesSameDay` - **Update this test**
   - ✅ `TestIntegration_NewCommand_WithOverwrite` - Verify still works
   - ➕ Add test for overwrite when numbered files exist

### Test Execution Plan

```bash
# 1. Run unit tests
go test ./internal/entry -v

# 2. Run integration tests
go test ./integration -v

# 3. Run all tests
go test ./... -v

# 4. Run specific test
go test ./integration -v -run TestIntegration_NewCommand_MultipleEntriesSameDay
```

---

## Implementation Checklist

### Phase 1: Critical Fix (Bug #1)

- [ ] Modify `internal/entry/entry.go` - Fix collision handling logic
- [ ] Run unit tests - Verify `TestCreateEntry` passes
- [ ] Run integration tests - Verify `TestIntegration_NewCommand_MultipleEntriesSameDay` passes
- [ ] Update integration test - Remove "may be expected" comments
- [ ] Add test case for multiple entries with same title
- [ ] Verify all existing tests still pass

### Phase 2: Overwrite Edge Case (Bug #2)

- [ ] Add test case `TestCreateEntryOverwrite_WhenBaseFileDoesNotExist`
- [ ] Verify test passes with Fix #1
- [ ] Document behavior in code comments

### Phase 3: Race Condition (Bug #3)

- [ ] Add documentation comment to `CreateEntry`
- [ ] (Optional) Implement retry logic if needed
- [ ] Update shared-concerns.md with race condition notes

### Phase 4: Polish (Bugs #4, #5)

- [ ] Verify template resolution edge cases (no fix needed)
- [ ] Standardize error messages (low priority)
- [ ] Update documentation

---

## Rollback Plan

If fixes cause issues:

1. **Git revert**: `git revert <commit-hash>`
2. **Test**: Run full test suite to ensure no regressions
3. **Investigate**: Identify what broke and why
4. **Fix**: Address issues before re-applying

---

## Success Criteria

✅ **All tests pass** (unit + integration)  
✅ **Collision handling works** - Multiple entries with same title create numbered files  
✅ **Overwrite works** - `--overwrite` flag overwrites base filename  
✅ **No regressions** - All existing functionality still works  
✅ **Documentation updated** - Code comments reflect behavior  

---

## Timeline Estimate

- **Fix #1 (Critical)**: 1-2 hours
  - Code change: 15 minutes
  - Testing: 30 minutes
  - Test updates: 30 minutes
  - Verification: 15 minutes

- **Fix #2 (Medium)**: 30 minutes
  - Test case: 15 minutes
  - Verification: 15 minutes

- **Fix #3 (Medium)**: 15 minutes
  - Documentation: 15 minutes

- **Fix #4 (Low)**: 15 minutes
  - Verification: 15 minutes

- **Fix #5 (Low)**: 1 hour (polish)
  - Error standardization: 1 hour

**Total**: ~3-4 hours for critical fixes, +1-2 hours for polish

---

## Next Steps

1. ✅ Review this plan
2. ⏭️ Implement Fix #1 (Critical)
3. ⏭️ Test Fix #1
4. ⏭️ Implement Fix #2 (if needed)
5. ⏭️ Document Fix #3
6. ⏭️ Verify Fix #4
7. ⏭️ Polish Fix #5 (optional)

---

## Questions?

If you have questions about this implementation plan, please ask before starting implementation.

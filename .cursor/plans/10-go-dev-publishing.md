# Phase 10: Go Module Publishing to go.dev

**Date**: 2025-01-15  
**Status**: Planning  
**Dependencies**: Phase 1 (Foundation), existing release workflow

## Goal

Enhance the GitHub Actions release workflow to ensure the `touchlog` Go module is properly published and indexed on go.dev (pkg.go.dev) after each release. This includes proactive triggering of module indexing, robust verification, and comprehensive error handling.

## Duration

**Estimated**: 1-2 days  
**Complexity**: Low-Medium

## Prerequisites

- Existing release workflow (`.github/workflows/release.yml`)
- Valid `go.mod` file with proper module path
- `go.sum` file present and up-to-date
- Semantic version tags (vX.Y.Z format)

## Overview

The Go module proxy (proxy.golang.org) automatically indexes modules when they are requested via `go get`. However, this indexing can take time (minutes to hours), and there's no guarantee it happens immediately after a release. This phase enhances the release workflow to:

1. **Ensure prerequisites** - Validate `go.mod` and `go.sum` are correct
2. **Trigger indexing proactively** - Use `go get` to force the proxy to fetch the module
3. **Verify via multiple methods** - Check both proxy API and pkg.go.dev
4. **Robust error handling** - Retry with exponential backoff, warn but don't fail

## Implementation Tasks

### 10.1 Pre-Release Validation

**Goal**: Ensure module is ready for publishing before release

**Tasks**:
- [ ] Add step to validate `go.mod` format and module path
- [ ] Verify `go.sum` is present and valid
- [ ] Check that module path matches repository (github.com/sv4u/touchlog)
- [ ] Validate semantic version format matches tag format

**Implementation**:
- Add validation step before tag creation
- Use `go mod verify` to check module integrity
- Use `go mod tidy` check (ensure dependencies are clean)

### 10.2 Proactive Indexing Trigger

**Goal**: Force Go proxy to fetch and index the module immediately after tag push

**Tasks**:
- [ ] Add step to trigger module fetch via `go get`
- [ ] Use `GOPROXY=direct` to bypass cache and force fetch from source
- [ ] Fetch the specific tagged version
- [ ] Handle errors gracefully (may fail if tag not yet propagated)

**Implementation**:
```bash
# After tag push, trigger indexing
GOPROXY=direct go get github.com/sv4u/touchlog@${NEXT_TAG}
```

**Location**: After "Create and push tag" step, before GoReleaser

### 10.3 Enhanced Verification

**Goal**: Verify module is indexed using multiple methods with retries

**Tasks**:
- [ ] Replace existing simple verification with comprehensive multi-method check
- [ ] Check proxy API endpoint: `https://proxy.golang.org/github.com/sv4u/touchlog/@v/${NEXT_TAG}.info`
- [ ] Check pkg.go.dev: `https://pkg.go.dev/github.com/sv4u/touchlog@${NEXT_TAG}`
- [ ] Implement retry logic with exponential backoff (3-5 attempts)
- [ ] Add detailed logging for each verification attempt

**Implementation**:
- Create verification script with retry logic
- Check both endpoints in parallel or sequentially
- Wait intervals: 10s, 30s, 60s, 120s (exponential backoff)
- Log status of each attempt

**Location**: Replace existing "Verify pkg.go.dev indexing" step

### 10.4 Error Handling Strategy

**Goal**: Ensure verification failures don't block successful releases

**Tasks**:
- [ ] Implement retry logic with exponential backoff (3-5 attempts)
- [ ] After retries, log warning but don't fail workflow
- [ ] Provide clear messaging about indexing delays
- [ ] Include links to check manually later

**Implementation**:
- Use `continue-on-error: true` for verification step
- Log warnings with actionable information
- Include manual check instructions in summary

### 10.5 Documentation Updates

**Goal**: Document the publishing process and expected behavior

**Tasks**:
- [ ] Update release workflow comments
- [ ] Add notes about indexing delays
- [ ] Document manual verification steps
- [ ] Update release summary output

## Implementation Checklist

### Pre-Release Validation
- [ ] Add `go mod verify` check
- [ ] Add `go mod tidy` check (verify no changes needed)
- [ ] Validate module path matches repository
- [ ] Validate version format

### Proactive Triggering
- [ ] Add `go get` step after tag push
- [ ] Configure `GOPROXY=direct` to force fetch
- [ ] Handle tag propagation delays gracefully
- [ ] Add appropriate error handling

### Enhanced Verification
- [ ] Create verification script/step
- [ ] Implement proxy API check
- [ ] Implement pkg.go.dev check
- [ ] Add retry logic with exponential backoff
- [ ] Add detailed logging

### Error Handling
- [ ] Configure `continue-on-error: true`
- [ ] Add warning messages
- [ ] Include manual check instructions
- [ ] Update release summary

### Testing
- [ ] Test with dry-run mode
- [ ] Verify retry logic works
- [ ] Test error scenarios
- [ ] Verify warning messages are clear

## Testing Requirements

### Unit Testing
- N/A (workflow changes only)

### Integration Testing
- Test workflow in dry-run mode
- Verify validation steps catch issues
- Test retry logic with simulated delays
- Verify error handling doesn't block release

### Manual Testing
- Run actual release (or dry-run) to verify end-to-end
- Check that module appears on go.dev after release
- Verify warning messages are helpful

## Success Criteria

1. ✅ Module validation runs before release
2. ✅ Proactive indexing trigger executes after tag push
3. ✅ Verification checks both proxy API and pkg.go.dev
4. ✅ Retry logic works with exponential backoff
5. ✅ Verification failures warn but don't block release
6. ✅ Release summary includes go.dev links
7. ✅ Documentation is clear and helpful

## Risk Assessment

### Low Risk
- Adding validation steps (non-breaking)
- Improving verification (enhancement only)
- Warning instead of failing (safe)

### Medium Risk
- `go get` may fail if tag not propagated (handled with retries)
- Proxy API may be slow (handled with retries and warnings)

### Mitigation
- All verification steps use `continue-on-error: true`
- Retry logic handles transient failures
- Clear warnings guide manual verification if needed

## Dependencies

- Existing release workflow (`.github/workflows/release.yml`)
- Go toolchain (already in workflow)
- Internet access (for proxy and pkg.go.dev checks)

## Related Files

- `.github/workflows/release.yml` - Main release workflow
- `go.mod` - Module definition
- `go.sum` - Module checksums

## Notes

- Go module indexing is asynchronous and can take time
- The proxy may cache responses, so immediate verification may fail
- Manual verification at https://pkg.go.dev/github.com/sv4u/touchlog is always available
- This phase enhances existing functionality without breaking changes


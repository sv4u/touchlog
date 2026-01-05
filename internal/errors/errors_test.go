package errors

import (
	"errors"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrPlatformUnsupported",
			err:  ErrPlatformUnsupported,
			want: "unsupported platform: touchlog only supports macOS, Linux, and WSL",
		},
		{
			name: "ErrConfigNotFound",
			err:  ErrConfigNotFound,
			want: "config file not found",
		},
		{
			name: "ErrConfigInvalid",
			err:  ErrConfigInvalid,
			want: "failed to parse config",
		},
		{
			name: "ErrTemplateNotFound",
			err:  ErrTemplateNotFound,
			want: "template not found",
		},
		{
			name: "ErrPermissionDenied",
			err:  ErrPermissionDenied,
			want: "permission denied",
		},
		{
			name: "ErrInvalidUTF8",
			err:  ErrInvalidUTF8,
			want: "invalid UTF-8",
		},
		{
			name: "ErrFileExists",
			err:  ErrFileExists,
			want: "file already exists",
		},
		{
			name: "ErrEditorNotFound",
			err:  ErrEditorNotFound,
			want: "editor not found",
		},
		{
			name: "ErrOutputDirRequired",
			err:  ErrOutputDirRequired,
			want: "output directory is required",
		},
		{
			name: "ErrTimezoneInvalid",
			err:  ErrTimezoneInvalid,
			want: "invalid timezone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error message = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	baseErr := ErrConfigNotFound
	wrapped := WrapError(baseErr, "failed to load config")

	if wrapped == nil {
		t.Fatal("WrapError() returned nil")
	}

	if !errors.Is(wrapped, baseErr) {
		t.Error("WrapError() should wrap the base error")
	}

	expectedMsg := "failed to load config: config file not found"
	if wrapped.Error() != expectedMsg {
		t.Errorf("WrapError() message = %q, want %q", wrapped.Error(), expectedMsg)
	}
}

func TestWrapError_Nil(t *testing.T) {
	wrapped := WrapError(nil, "context")
	if wrapped != nil {
		t.Error("WrapError(nil) should return nil")
	}
}

func TestWrapErrorf(t *testing.T) {
	baseErr := ErrTemplateNotFound
	wrapped := WrapErrorf(baseErr, "template %q not found", "daily")

	if wrapped == nil {
		t.Fatal("WrapErrorf() returned nil")
	}

	if !errors.Is(wrapped, baseErr) {
		t.Error("WrapErrorf() should wrap the base error")
	}

	expectedMsg := "template \"daily\" not found: template not found"
	if wrapped.Error() != expectedMsg {
		t.Errorf("WrapErrorf() message = %q, want %q", wrapped.Error(), expectedMsg)
	}
}

func TestWrapErrorf_Nil(t *testing.T) {
	wrapped := WrapErrorf(nil, "template %q not found", "daily")
	if wrapped != nil {
		t.Error("WrapErrorf(nil) should return nil")
	}
}

func TestErrorIs(t *testing.T) {
	// Test that errors.Is works correctly with wrapped errors
	baseErr := ErrConfigInvalid
	wrapped := WrapError(baseErr, "failed to parse config file")

	if !errors.Is(wrapped, baseErr) {
		t.Error("errors.Is() should return true for wrapped error")
	}

	if !errors.Is(wrapped, ErrConfigInvalid) {
		t.Error("errors.Is() should return true for error type")
	}
}

package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	// Get the actual home directory for testing
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v (needed for test setup)", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "path starting with ~",
			path:    "~/notes",
			want:    filepath.Join(homeDir, "notes"),
			wantErr: false,
		},
		{
			name:    "path that is just ~",
			path:    "~",
			want:    homeDir,
			wantErr: false,
		},
		{
			name:    "path that doesn't start with ~",
			path:    "/absolute/path",
			want:    "/absolute/path",
			wantErr: false,
		},
		{
			name:    "path with ~ in middle - should not expand",
			path:    "/path/with~tilde",
			want:    "/path/with~tilde",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "path with ~/ at start",
			path:    "~/Documents/notes",
			want:    filepath.Join(homeDir, "Documents", "notes"),
			wantErr: false,
		},
		{
			name:    "relative path without ~",
			path:    "notes/test",
			want:    "notes/test",
			wantErr: false,
		},
		{
			name: "path with multiple ~ at start",
			// Note: expandPath only checks if path starts with "~", so "~~/notes"
			// becomes path[2:] which is "/notes", resulting in homeDir + "/notes"
			path:    "~~/notes",
			want:    filepath.Join(homeDir, "notes"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expandPath(%q) expected error, got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("expandPath(%q) unexpected error = %v", tt.path, err)
				}
				if got != tt.want {
					t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.want)
				}
			}
		})
	}
}

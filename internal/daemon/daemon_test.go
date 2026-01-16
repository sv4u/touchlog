package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDaemon_GetPID_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	pid := d.GetPID()
	if pid != 0 {
		t.Errorf("expected PID 0 when no PID file exists, got %d", pid)
	}
}

func TestDaemon_GetPID_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	// Create invalid PID file
	pidDir := filepath.Dir(d.pidPath)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		t.Fatalf("creating PID directory: %v", err)
	}

	if err := os.WriteFile(d.pidPath, []byte("invalid\n"), 0644); err != nil {
		t.Fatalf("writing invalid PID file: %v", err)
	}

	pid := d.GetPID()
	if pid != 0 {
		t.Errorf("expected PID 0 for invalid PID file, got %d", pid)
	}
}

func TestDaemon_GetPID_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	// Write valid PID
	expectedPID := 12345
	if err := d.WritePID(expectedPID); err != nil {
		t.Fatalf("writing PID: %v", err)
	}

	pid := d.GetPID()
	if pid != expectedPID {
		t.Errorf("expected PID %d, got %d", expectedPID, pid)
	}
}

func TestDaemon_IsRunning_NoPIDFile(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	if d.IsRunning() {
		t.Error("expected daemon not to be running when PID file doesn't exist")
	}
}

func TestDaemon_WritePID(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDaemon(tmpDir)

	pid := 12345
	if err := d.WritePID(pid); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(d.pidPath); err != nil {
		t.Fatalf("PID file was not created: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(d.pidPath)
	if err != nil {
		t.Fatalf("reading PID file: %v", err)
	}

	expected := "12345\n"
	if string(data) != expected {
		t.Errorf("expected PID file content %q, got %q", expected, string(data))
	}
}

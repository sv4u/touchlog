package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDaemon_Start_Behavior_CreatesPIDFile tests Start creates PID file
func TestDaemon_Start_Behavior_CreatesPIDFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if strings.Contains(err.Error(), "bind: invalid argument") || strings.Contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = daemon.Stop()
	}()

	// Verify PID file was created
	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Error("expected PID file to be created")
	}
}

// TestDaemon_Stop_Behavior_RemovesPIDFile tests Stop removes PID file
func TestDaemon_Stop_Behavior_RemovesPIDFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if strings.Contains(err.Error(), "bind: invalid argument") || strings.Contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}

	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Fatal("PID file should exist after Start")
	}

	if err := daemon.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify PID file was removed
	if _, err := os.Stat(pidPath); err == nil {
		t.Error("expected PID file to be removed after Stop")
	}
}

// TestDaemon_Status_Behavior_WhenRunning tests Status when daemon is running
func TestDaemon_Status_Behavior_WhenRunning(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if strings.Contains(err.Error(), "bind: invalid argument") || strings.Contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = daemon.Stop()
	}()

	running, pid, err := daemon.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if !running {
		t.Error("expected daemon to be running")
	}

	if pid <= 0 {
		t.Errorf("expected valid PID, got %d", pid)
	}
}

// TestDaemon_Status_Behavior_WhenStopped tests Status when daemon is stopped
func TestDaemon_Status_Behavior_WhenStopped(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)

	running, pid, err := daemon.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if running {
		t.Error("expected daemon not to be running")
	}

	if pid != 0 {
		t.Errorf("expected PID 0 when not running, got %d", pid)
	}
}

// TestDaemon_GetPID_Behavior_ReturnsPID tests GetPID returns PID from file
func TestDaemon_GetPID_Behavior_ReturnsPID(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if strings.Contains(err.Error(), "bind: invalid argument") || strings.Contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = daemon.Stop()
	}()

	pid := daemon.GetPID()
	if pid <= 0 {
		t.Errorf("expected valid PID, got %d", pid)
	}
}

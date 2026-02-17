package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isUnixSocketError checks if an error is related to Unix socket limitations
// that should cause the test to be skipped rather than failed
func isUnixSocketError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "bind: invalid argument") ||
		strings.Contains(errMsg, "AF_UNIX path too long") ||
		strings.Contains(errMsg, "address too long") ||
		strings.Contains(errMsg, "listen unix") ||
		strings.Contains(errMsg, "Unix socket")
}

// shortTempDir creates a temp directory with a short path to avoid
// exceeding the ~104 character limit for Unix domain socket paths on macOS.
func shortTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "tl-")
	if err != nil {
		t.Fatalf("creating short temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

// testSocketPath derives the socket path for a test vault and registers cleanup
func testSocketPath(t *testing.T, vaultRoot string) string {
	t.Helper()
	sockPath, err := SocketPathForVault(vaultRoot)
	if err != nil {
		t.Fatalf("computing socket path: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(sockPath)
	})
	return sockPath
}

// TestDaemon_Start_Behavior_CreatesPIDFile tests startServer creates PID file
func TestDaemon_Start_Behavior_CreatesPIDFile(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	d := NewDaemon(tmpDir)
	if err := d.startServer(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("startServer failed: %v", err)
	}
	defer d.cleanup()

	// Verify PID file was created
	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	if _, err := os.Stat(pidPath); err != nil {
		t.Error("expected PID file to be created")
	}

	// Verify PID is the current process
	pid := d.GetPID()
	if pid != os.Getpid() {
		t.Errorf("expected PID %d, got %d", os.Getpid(), pid)
	}
}

// TestDaemon_Stop_Behavior_RemovesPIDFile tests cleanup removes PID file
func TestDaemon_Stop_Behavior_RemovesPIDFile(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	d := NewDaemon(tmpDir)
	if err := d.startServer(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("startServer failed: %v", err)
	}

	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	if _, err := os.Stat(pidPath); err != nil {
		t.Fatal("PID file should exist after startServer")
	}

	// Stop the daemon (in-process server, so Stop detects server reference)
	if err := d.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify PID file was removed
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("expected PID file to be removed after Stop")
	}
}

// TestDaemon_Status_Behavior_WhenRunning tests Status when daemon is running
func TestDaemon_Status_Behavior_WhenRunning(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	d := NewDaemon(tmpDir)
	if err := d.startServer(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("startServer failed: %v", err)
	}
	defer d.cleanup()

	running, pid, err := d.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if !running {
		t.Error("expected daemon to be running")
	}

	if pid != os.Getpid() {
		t.Errorf("expected PID %d, got %d", os.Getpid(), pid)
	}
}

// TestDaemon_Status_Behavior_WhenStopped tests Status when daemon is stopped
func TestDaemon_Status_Behavior_WhenStopped(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	d := NewDaemon(tmpDir)

	running, pid, err := d.Status()
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
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	d := NewDaemon(tmpDir)
	if err := d.startServer(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("startServer failed: %v", err)
	}
	defer d.cleanup()

	pid := d.GetPID()
	if pid <= 0 {
		t.Errorf("expected valid PID, got %d", pid)
	}

	if pid != os.Getpid() {
		t.Errorf("expected current process PID %d, got %d", os.Getpid(), pid)
	}
}

// TestDaemon_Cleanup_Behavior_RemovesAllFiles tests cleanup removes both PID and socket files
func TestDaemon_Cleanup_Behavior_RemovesAllFiles(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	d := NewDaemon(tmpDir)
	if err := d.startServer(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("startServer failed: %v", err)
	}

	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	sockPath := d.SocketPath()

	// Verify files exist
	if _, err := os.Stat(pidPath); err != nil {
		t.Fatal("PID file should exist after startServer")
	}
	if _, err := os.Stat(sockPath); err != nil {
		t.Fatalf("socket file should exist after startServer at %s", sockPath)
	}

	// Cleanup
	d.cleanup()

	// Verify both files are removed
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("expected PID file to be removed after cleanup")
	}
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("expected socket file to be removed after cleanup")
	}
}

// TestDaemon_IsDaemonChild_Default tests IsDaemonChild is false by default
func TestDaemon_IsDaemonChild_Default(t *testing.T) {
	if IsDaemonChild() {
		t.Error("expected IsDaemonChild to be false by default")
	}
}

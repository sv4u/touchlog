package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestDaemon_Start_Behavior_CreatesPIDFile tests Start creates PID file
func TestDaemon_Start_Behavior_CreatesPIDFile(t *testing.T) {
	// Set a timeout for this test to prevent hanging
	deadline, ok := t.Deadline()
	if ok {
		timeout := time.Until(deadline) - 100*time.Millisecond // Leave some buffer
		if timeout <= 0 {
			timeout = 30 * time.Second // Default timeout if deadline is too soon
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		go func() {
			<-timer.C
			t.Error("test timeout: test took too long, possible hang detected")
		}()
	}

	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		// Ensure cleanup happens even if test fails
		if err := daemon.Stop(); err != nil {
			t.Logf("Warning: Stop() returned error during cleanup: %v", err)
		}
	}()

	// Verify PID file was created with retry logic
	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	maxRetries := 10
	retryDelay := 50 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(pidPath); err == nil {
			return // PID file exists, test passes
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	t.Error("expected PID file to be created")
}

// TestDaemon_Stop_Behavior_RemovesPIDFile tests Stop removes PID file
func TestDaemon_Stop_Behavior_RemovesPIDFile(t *testing.T) {
	// Set a timeout for this test to prevent hanging
	deadline, ok := t.Deadline()
	if ok {
		timeout := time.Until(deadline) - 100*time.Millisecond
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		go func() {
			<-timer.C
			t.Error("test timeout: test took too long, possible hang detected")
		}()
	}

	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		if err := daemon.Stop(); err != nil {
			t.Logf("Warning: Stop() returned error during cleanup: %v", err)
		}
	}()

	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	// Wait for PID file with retry
	maxRetries := 10
	retryDelay := 50 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(pidPath); err == nil {
			break
		}
		if i == maxRetries-1 {
			t.Fatal("PID file should exist after Start")
		}
		time.Sleep(retryDelay)
	}

	if err := daemon.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify PID file was removed with retry
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(pidPath); os.IsNotExist(err) {
			return // PID file removed, test passes
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	t.Error("expected PID file to be removed after Stop")
}

// TestDaemon_Status_Behavior_WhenRunning tests Status when daemon is running
func TestDaemon_Status_Behavior_WhenRunning(t *testing.T) {
	// Set a timeout for this test to prevent hanging
	deadline, ok := t.Deadline()
	if ok {
		timeout := time.Until(deadline) - 100*time.Millisecond
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		go func() {
			<-timer.C
			t.Error("test timeout: test took too long, possible hang detected")
		}()
	}

	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		if err := daemon.Stop(); err != nil {
			t.Logf("Warning: Stop() returned error during cleanup: %v", err)
		}
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
	// Set a timeout for this test to prevent hanging
	deadline, ok := t.Deadline()
	if ok {
		timeout := time.Until(deadline) - 100*time.Millisecond
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		go func() {
			<-timer.C
			t.Error("test timeout: test took too long, possible hang detected")
		}()
	}

	tmpDir := t.TempDir()
	setupTestVault(t, tmpDir)

	daemon := NewDaemon(tmpDir)
	if err := daemon.Start(); err != nil {
		if isUnixSocketError(err) {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		if err := daemon.Stop(); err != nil {
			t.Logf("Warning: Stop() returned error during cleanup: %v", err)
		}
	}()

	// Wait for PID file to be written with retry
	pidPath := filepath.Join(tmpDir, ".touchlog", "daemon.pid")
	maxRetries := 10
	retryDelay := 50 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(pidPath); err == nil {
			break
		}
		if i == maxRetries-1 {
			t.Fatalf("PID file not created after %d retries", maxRetries)
		}
		time.Sleep(retryDelay)
	}

	pid := daemon.GetPID()
	if pid <= 0 {
		t.Errorf("expected valid PID, got %d", pid)
	}
}

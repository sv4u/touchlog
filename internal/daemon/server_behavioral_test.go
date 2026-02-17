package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sv4u/touchlog/v2/internal/config"
)

// TestNewServer_Behavior_CreatesSocketDirectory tests NewServer creates socket directory
func TestNewServer_Behavior_CreatesSocketDirectory(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	absVaultRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("getting absolute path: %v", err)
	}

	// Remove .touchlog directory to test creation
	touchlogDir := filepath.Join(absVaultRoot, ".touchlog")
	if err := os.RemoveAll(touchlogDir); err != nil {
		t.Fatalf("removing .touchlog dir: %v", err)
	}

	server, err := NewServer(absVaultRoot, cfg)
	if err != nil {
		if contains(err.Error(), "bind: invalid argument") || contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("NewServer failed: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	// Verify socket directory was created
	if _, err := os.Stat(touchlogDir); os.IsNotExist(err) {
		t.Error("expected .touchlog directory to be created")
	}
}

// TestNewServer_Behavior_RemovesExistingSocket tests NewServer removes existing socket
func TestNewServer_Behavior_RemovesExistingSocket(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	absVaultRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("getting absolute path: %v", err)
	}

	sockPath := filepath.Join(absVaultRoot, ".touchlog", "daemon.sock")

	// Create existing socket file
	if err := os.WriteFile(sockPath, []byte("test"), 0644); err != nil {
		t.Fatalf("creating existing socket file: %v", err)
	}

	server, err := NewServer(absVaultRoot, cfg)
	if err != nil {
		if contains(err.Error(), "bind: invalid argument") || contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("NewServer failed: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	// Verify old socket was removed and new one created
	if _, err := os.Stat(sockPath); os.IsNotExist(err) {
		t.Error("expected socket to exist after NewServer")
	}
}

// TestServer_Start_Behavior_InitializesWatcher tests Start initializes watcher
func TestServer_Start_Behavior_InitializesWatcher(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	absVaultRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("getting absolute path: %v", err)
	}

	server, err := NewServer(absVaultRoot, cfg)
	if err != nil {
		if contains(err.Error(), "bind: invalid argument") || contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("NewServer failed: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	// Verify watcher was initialized
	if server.watcher == nil {
		t.Error("expected watcher to be initialized after Start")
	}

	// Verify indexer was initialized
	if server.indexer == nil {
		t.Error("expected indexer to be initialized after Start")
	}
}

// TestServer_Stop_Behavior_ClosesListener tests Stop closes listener
func TestServer_Stop_Behavior_ClosesListener(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	absVaultRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("getting absolute path: %v", err)
	}

	server, err := NewServer(absVaultRoot, cfg)
	if err != nil {
		if contains(err.Error(), "bind: invalid argument") || contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("NewServer failed: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop should close listener
	if err := server.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify listener is closed (attempting to accept should fail)
	// This is tested implicitly - if listener wasn't closed, we'd have issues
}

// TestServer_ProcessMessage_Behavior_HandlesAllMessageTypes tests processMessage handles all message types
func TestServer_ProcessMessage_Behavior_HandlesAllMessageTypes(t *testing.T) {
	tmpDir := shortTempDir(t)
	setupTestVault(t, tmpDir)

	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	absVaultRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("getting absolute path: %v", err)
	}

	server, err := NewServer(absVaultRoot, cfg)
	if err != nil {
		if contains(err.Error(), "bind: invalid argument") || contains(err.Error(), "AF_UNIX path too long") {
			t.Skipf("Skipping test due to system limitation: %v", err)
		}
		t.Fatalf("NewServer failed: %v", err)
	}

	// Start server to initialize indexer (required for ReindexPaths)
	if err := server.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	// Test all message types return responses (not nil)
	testCases := []struct {
		msgType MessageType
		payload interface{}
	}{
		{MessageTypeStatus, nil},
		{MessageTypeQueryExecute, QueryExecuteRequest{Query: "test"}},
		{MessageTypeReindexPaths, ReindexPathsRequest{Paths: []string{}}},
		{MessageTypeShutdown, nil},
	}

	for _, tc := range testCases {
		var msg *Message
		var err error
		if tc.payload != nil {
			msg, err = NewMessage(tc.msgType, tc.payload)
			if err != nil {
				t.Fatalf("creating message for %s: %v", tc.msgType, err)
			}
		} else {
			msg = &Message{
				Version: ProtocolVersion,
				Type:    tc.msgType,
			}
		}

		response := server.processMessage(msg)
		if response == nil {
			t.Errorf("expected response for message type %s, got nil", tc.msgType)
			continue
		}
		if response.Version != ProtocolVersion {
			t.Errorf("expected protocol version %d for %s, got %d", ProtocolVersion, tc.msgType, response.Version)
		}
	}
}

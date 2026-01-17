package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sv4u/touchlog/internal/config"
)

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestServer_UnixSocket(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup vault
	setupTestVault(t, tmpDir)

	// Load config
	cfg, err := config.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	// Get absolute path (Unix sockets may require absolute paths on some systems)
	absVaultRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("getting absolute path: %v", err)
	}

	// Create server
	// Note: On some systems (especially macOS with certain temp directory configurations),
	// Unix socket creation in temp directories may fail due to filesystem limitations.
	// This is a system limitation, not a code bug. The server code is correct.
	server, err := NewServer(absVaultRoot, cfg)
	if err != nil {
		// Check if it's a socket creation error (system limitation)
		if err.Error() != "" && (contains(err.Error(), "bind: invalid argument") || contains(err.Error(), "AF_UNIX path too long")) {
			t.Skipf("Skipping Unix socket test due to system limitation: %v (this is expected on some macOS configurations)", err)
		}
		t.Fatalf("NewServer failed: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Give server time to create socket
	time.Sleep(200 * time.Millisecond)

	// Connect to server
	sockPath := filepath.Join(absVaultRoot, ".touchlog", "daemon.sock")

	// Check if socket exists
	if _, err := os.Stat(sockPath); os.IsNotExist(err) {
		t.Fatalf("socket was not created at %s", sockPath)
	}

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dialing server: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Send status request
	msg, err := NewMessage(MessageTypeStatus, nil)
	if err != nil {
		t.Fatalf("creating message: %v", err)
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("sending message: %v", err)
	}

	// Receive response
	decoder := json.NewDecoder(conn)
	var response Response
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("receiving response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected successful response, got error: %s", response.Error)
	}

	if response.Version != ProtocolVersion {
		t.Errorf("expected protocol version %d, got %d", ProtocolVersion, response.Version)
	}
}

func TestProtocol_MessageTypes(t *testing.T) {
	// Test creating different message types
	msg, err := NewMessage(MessageTypeStatus, nil)
	if err != nil {
		t.Fatalf("creating status message: %v", err)
	}
	if msg.Type != MessageTypeStatus {
		t.Errorf("expected type %s, got %s", MessageTypeStatus, msg.Type)
	}

	req := QueryExecuteRequest{Query: "test"}
	msg, err = NewMessage(MessageTypeQueryExecute, req)
	if err != nil {
		t.Fatalf("creating query message: %v", err)
	}
	if msg.Type != MessageTypeQueryExecute {
		t.Errorf("expected type %s, got %s", MessageTypeQueryExecute, msg.Type)
	}
}

func TestProtocol_Response(t *testing.T) {
	// Test successful response
	resp := NewResponse(true, "test data", nil)
	if !resp.Success {
		t.Error("expected successful response")
	}
	if resp.Error != "" {
		t.Errorf("expected no error, got %s", resp.Error)
	}

	// Test error response
	resp = NewResponse(false, nil, fmt.Errorf("test error"))
	if resp.Success {
		t.Error("expected failed response")
	}
	if resp.Error != "test error" {
		t.Errorf("expected error 'test error', got %s", resp.Error)
	}
}

// setupTestVault creates a test vault with config
func setupTestVault(t *testing.T, tmpDir string) {
	t.Helper()

	// Create .touchlog directory
	touchlogDir := filepath.Join(tmpDir, ".touchlog")
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		t.Fatalf("creating .touchlog dir: %v", err)
	}

	// Create config file
	configPath := filepath.Join(touchlogDir, "config.yaml")
	configContent := `version: 1
types:
  note:
    description: A note
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
tags:
  preferred: []
edges:
  related-to:
    description: General relationship
templates:
  root: templates
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
}

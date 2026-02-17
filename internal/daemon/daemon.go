package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/index"
)

// daemonChildEnv is the environment variable used to identify the daemon child process.
// When set to "1", the current process is the forked daemon child.
const daemonChildEnv = "_TOUCHLOG_DAEMON_CHILD"

// Daemon manages the touchlog daemon lifecycle
type Daemon struct {
	vaultRoot string
	pidPath   string
	sockPath  string
	server    *Server
}

// SocketPathForVault derives a deterministic Unix domain socket path for the given vault.
// Unix domain socket paths are limited to 104 bytes on macOS (108 on Linux).
// To avoid exceeding this limit with long vault paths, we derive a short
// hash-based path in /tmp that is always well under the limit.
func SocketPathForVault(vaultRoot string) (string, error) {
	absPath, err := filepath.Abs(vaultRoot)
	if err != nil {
		return "", fmt.Errorf("resolving vault path: %w", err)
	}
	// Resolve symlinks for consistency (e.g., /var -> /private/var on macOS)
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = resolved
	}
	hash := sha256.Sum256([]byte(absPath))
	hexHash := hex.EncodeToString(hash[:6]) // 12 hex chars = sufficient uniqueness
	return filepath.Join("/tmp", fmt.Sprintf("touchlog-%s.sock", hexHash)), nil
}

// NewDaemon creates a new daemon instance for a vault
func NewDaemon(vaultRoot string) *Daemon {
	touchlogDir := filepath.Join(vaultRoot, ".touchlog")
	sockPath, err := SocketPathForVault(vaultRoot)
	if err != nil {
		// Fallback to vault-local path if hash computation fails
		sockPath = filepath.Join(touchlogDir, "daemon.sock")
	}
	return &Daemon{
		vaultRoot: vaultRoot,
		pidPath:   filepath.Join(touchlogDir, "daemon.pid"),
		sockPath:  sockPath,
	}
}

// SocketPath returns the path to the daemon's Unix domain socket
func (d *Daemon) SocketPath() string {
	return d.sockPath
}

// IsDaemonChild returns true if the current process is the daemon child process
func IsDaemonChild() bool {
	return os.Getenv(daemonChildEnv) == "1"
}

// Start launches the daemon as a background process.
// It forks a child process that runs the daemon loop via Run().
// The parent process returns after confirming the child started successfully.
func (d *Daemon) Start() error {
	// Validate vault exists
	configPath := filepath.Join(d.vaultRoot, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not initialized (run 'touchlog init' first)")
	}

	// Check if daemon is already running
	if d.IsRunning() {
		return fmt.Errorf("daemon is already running (PID: %d)", d.GetPID())
	}

	// Clean up any stale files from previous crashed runs
	d.cleanupFiles()

	// Get the path to the current executable for re-exec
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable path: %w", err)
	}

	// Resolve absolute vault path so the child process doesn't depend on CWD
	absVaultRoot, err := filepath.Abs(d.vaultRoot)
	if err != nil {
		return fmt.Errorf("resolving vault path: %w", err)
	}

	// Launch child process as a daemon with a new session (detached from terminal)
	cmd := exec.Command(exe, "--vault", absVaultRoot, "daemon", "start") // #nosec G204 -- exe is from os.Executable() (self re-exec), not user input
	cmd.Env = append(os.Environ(), daemonChildEnv+"=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting daemon process: %w", err)
	}

	childPID := cmd.Process.Pid

	// Release the child process so we don't wait for it
	_ = cmd.Process.Release()

	// Wait for the child to write its PID file, confirming it started successfully
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		pid := d.GetPID()
		if pid == childPID {
			if process, err := os.FindProcess(childPID); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return nil
				}
			}
		}
	}

	// Timeout: daemon child failed to start or write PID file
	if process, err := os.FindProcess(childPID); err == nil {
		_ = process.Kill()
	}
	d.cleanupFiles()
	return fmt.Errorf("daemon failed to start (timed out waiting for PID file)")
}

// Run runs the daemon in the foreground (blocking).
// This is called by the daemon child process after forking.
// It blocks until the daemon receives a shutdown signal (SIGINT, SIGTERM)
// or the IPC server is shut down via a Shutdown command.
func (d *Daemon) Run() error {
	if err := d.startServer(); err != nil {
		return err
	}

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal or server done
	select {
	case <-sigCh:
		// Received shutdown signal from OS or 'daemon stop'
	case <-d.server.Done():
		// Server shut down via IPC Shutdown command
	}

	// Graceful cleanup: stop server, remove PID/socket files
	d.cleanup()
	return nil
}

// startServer starts the IPC server in the current process.
// It validates the vault, loads config, rebuilds the index if needed,
// creates and starts the IPC server, and writes the PID file.
// This method returns immediately after setup; goroutines handle connections.
func (d *Daemon) startServer() error {
	// Validate vault
	configPath := filepath.Join(d.vaultRoot, ".touchlog", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not initialized (run 'touchlog init' first)")
	}

	// Load config
	cfg, err := config.LoadConfig(d.vaultRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Check if daemon is already running
	if d.IsRunning() {
		return fmt.Errorf("daemon is already running (PID: %d)", d.GetPID())
	}

	// Auto-rebuild index if missing
	indexPath := filepath.Join(d.vaultRoot, ".touchlog", "index.db")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		builder := index.NewBuilder(d.vaultRoot, cfg)
		if err := builder.Rebuild(); err != nil {
			return fmt.Errorf("rebuilding index: %w", err)
		}
	}

	// Start IPC server
	server, err := NewServer(d.vaultRoot, d.sockPath, cfg)
	if err != nil {
		return fmt.Errorf("creating IPC server: %w", err)
	}

	if err := server.Start(); err != nil {
		_ = server.Stop()
		return fmt.Errorf("starting IPC server: %w", err)
	}

	d.server = server

	// Write PID file
	pid := os.Getpid()
	if err := d.WritePID(pid); err != nil {
		d.cleanup()
		return fmt.Errorf("writing PID file: %w", err)
	}

	return nil
}

// Stop stops the daemon
func (d *Daemon) Stop() error {
	// If we have an in-process server reference (e.g. test scenario), stop it directly
	if d.server != nil {
		d.cleanup()
		return nil
	}

	if !d.IsRunning() {
		d.cleanupFiles()
		return fmt.Errorf("daemon is not running")
	}

	pid := d.GetPID()
	if pid == 0 {
		d.cleanupFiles()
		return fmt.Errorf("could not determine daemon PID")
	}

	// Self-PID check: if the PID file points to this process, just clean up
	// (can happen in test scenarios where daemon was started in-process)
	if pid == os.Getpid() {
		d.cleanupFiles()
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		d.cleanupFiles()
		return fmt.Errorf("finding daemon process: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		d.cleanupFiles()
		return fmt.Errorf("sending signal to daemon: %w", err)
	}

	// Wait for the daemon to exit gracefully (up to 5 seconds)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		if err := process.Signal(syscall.Signal(0)); err != nil {
			// Process has exited - the daemon cleans up its own files on signal,
			// but remove any leftovers just in case
			d.cleanupFiles()
			return nil
		}
	}

	// Timeout: force kill and clean up
	_ = process.Kill()
	d.cleanupFiles()
	return nil
}

// Status returns the daemon status
func (d *Daemon) Status() (bool, int, error) {
	if !d.IsRunning() {
		return false, 0, nil
	}

	pid := d.GetPID()
	if pid == 0 {
		return false, 0, nil
	}

	// Verify process is actually running
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0, nil
	}

	// Check if process is alive by sending signal 0
	if err := process.Signal(syscall.Signal(0)); err != nil {
		// Process doesn't exist, clean up stale files
		d.cleanupFiles()
		return false, 0, nil
	}

	return true, pid, nil
}

// IsRunning checks if the daemon is running
func (d *Daemon) IsRunning() bool {
	if _, err := os.Stat(d.pidPath); os.IsNotExist(err) {
		return false
	}

	pid := d.GetPID()
	if pid == 0 {
		return false
	}

	// Verify process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Check if process is alive by sending signal 0
	if err := process.Signal(syscall.Signal(0)); err != nil {
		// Process doesn't exist, clean up stale files
		d.cleanupFiles()
		return false
	}

	return true
}

// GetPID returns the daemon PID from the PID file
func (d *Daemon) GetPID() int {
	data, err := os.ReadFile(d.pidPath)
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}

	return pid
}

// WritePID writes the PID to the PID file
func (d *Daemon) WritePID(pid int) error {
	pidDir := filepath.Dir(d.pidPath)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("creating PID directory: %w", err)
	}

	return os.WriteFile(d.pidPath, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

// cleanup stops the in-process server and removes PID/socket files
func (d *Daemon) cleanup() {
	if d.server != nil {
		_ = d.server.Stop()
		d.server = nil
	}
	d.cleanupFiles()
}

// cleanupFiles removes stale PID and socket files
func (d *Daemon) cleanupFiles() {
	_ = os.Remove(d.pidPath)
	_ = os.Remove(d.sockPath)
}

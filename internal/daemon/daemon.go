package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/index"
)

// Daemon manages the touchlog daemon lifecycle
type Daemon struct {
	vaultRoot string
	pidPath   string
	sockPath  string
}

// NewDaemon creates a new daemon instance for a vault
func NewDaemon(vaultRoot string) *Daemon {
	touchlogDir := filepath.Join(vaultRoot, ".touchlog")
	return &Daemon{
		vaultRoot: vaultRoot,
		pidPath:   filepath.Join(touchlogDir, "daemon.pid"),
		sockPath:  filepath.Join(touchlogDir, "daemon.sock"),
	}
}

// Start starts the daemon
func (d *Daemon) Start() error {
	// Validate vault (check for config.yaml)
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

	// Auto-rebuild index if missing or schema mismatch
	// Check if index exists
	indexPath := filepath.Join(d.vaultRoot, ".touchlog", "index.db")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Index doesn't exist, rebuild it
		builder := index.NewBuilder(d.vaultRoot, cfg)
		if err := builder.Rebuild(); err != nil {
			return fmt.Errorf("rebuilding index: %w", err)
		}
	} else {
		// Index exists, check schema version
		// For Phase 3, we'll assume schema is current
		// In a full implementation, we'd check schema version and migrate if needed
	}

	// Start IPC server
	server, err := NewServer(d.vaultRoot, cfg)
	if err != nil {
		return fmt.Errorf("creating IPC server: %w", err)
	}

	if err := server.Start(); err != nil {
		return fmt.Errorf("starting IPC server: %w", err)
	}

	// Write PID file
	pid := os.Getpid()
	if err := d.WritePID(pid); err != nil {
		server.Stop()
		return fmt.Errorf("writing PID file: %w", err)
	}

	// For Phase 3, the server runs in the foreground
	// In a full implementation, we'd fork and run in background
	// For now, we'll keep the server reference and let it run
	// The actual daemon process management would be handled by the OS or a process manager

	return nil
}

// Stop stops the daemon
func (d *Daemon) Stop() error {
	if !d.IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	pid := d.GetPID()
	if pid == 0 {
		return fmt.Errorf("could not determine daemon PID")
	}

	// Send SIGTERM to the daemon process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding daemon process: %w", err)
	}

	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("sending signal to daemon: %w", err)
	}

	// Wait a bit for graceful shutdown
	// In a full implementation, we'd wait for the process to exit
	// For Phase 3, we'll just remove the PID file
	os.Remove(d.pidPath)
	os.Remove(d.sockPath)

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
	if err := process.Signal(os.Signal(nil)); err != nil {
		// Process doesn't exist, clean up PID file
		os.Remove(d.pidPath)
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

	// Check if process is alive
	if err := process.Signal(os.Signal(nil)); err != nil {
		// Process doesn't exist, clean up
		os.Remove(d.pidPath)
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

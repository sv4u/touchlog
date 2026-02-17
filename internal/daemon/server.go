package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/query"
	"github.com/sv4u/touchlog/v2/internal/watch"
)

// Server handles IPC communication via Unix domain socket
type Server struct {
	vaultRoot string
	cfg       *config.Config
	listener  net.Listener
	watcher   *watch.Watcher
	indexer   *watch.IncrementalIndexer
	done      chan struct{}
	once      sync.Once
}

// NewServer creates a new IPC server
func NewServer(vaultRoot string, cfg *config.Config) (*Server, error) {
	// Ensure vault root is absolute (required for Unix sockets on some systems)
	absVaultRoot, err := filepath.Abs(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute vault path: %w", err)
	}

	// Resolve symlinks (important on macOS where /var -> /private/var)
	resolvedVaultRoot, err := filepath.EvalSymlinks(absVaultRoot)
	if err == nil {
		absVaultRoot = resolvedVaultRoot
	}

	sockPath := filepath.Join(absVaultRoot, ".touchlog", "daemon.sock")

	// Ensure directory exists
	sockDir := filepath.Dir(sockPath)
	if err := os.MkdirAll(sockDir, 0755); err != nil {
		return nil, fmt.Errorf("creating socket directory: %w", err)
	}

	// Remove existing socket if it exists (check both as file and socket)
	if info, err := os.Stat(sockPath); err == nil {
		if info.Mode()&os.ModeSocket != 0 || info.Mode().IsRegular() {
			if err := os.Remove(sockPath); err != nil {
				return nil, fmt.Errorf("removing existing socket: %w", err)
			}
			// Small delay to ensure filesystem has processed the removal
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Create Unix domain socket listener
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("creating Unix socket listener: %w", err)
	}

	return &Server{
		vaultRoot: absVaultRoot,
		cfg:       cfg,
		listener:  listener,
		done:      make(chan struct{}),
	}, nil
}

// Start starts the IPC server
func (s *Server) Start() error {
	// Start filesystem watcher
	watcher, err := watch.NewWatcher(s.vaultRoot)
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	s.watcher = watcher

	if err := watcher.Start(); err != nil {
		return fmt.Errorf("starting watcher: %w", err)
	}

	// Create incremental indexer (uses short-lived DB connections per event)
	s.indexer = watch.NewIncrementalIndexer(s.vaultRoot, s.cfg)

	// Start event processing goroutine
	go s.processWatchEvents()

	// Start accepting connections
	go s.acceptConnections()

	return nil
}

// Done returns a channel that's closed when the server is shut down.
// This allows callers to block until the server exits (e.g., via IPC Shutdown).
func (s *Server) Done() <-chan struct{} {
	return s.done
}

// Stop stops the IPC server
func (s *Server) Stop() error {
	s.once.Do(func() {
		close(s.done)
	})
	if s.watcher != nil {
		_ = s.watcher.Stop()
	}
	return s.listener.Close()
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections() {
	for {
		// Check if we should stop before attempting to accept
		select {
		case <-s.done:
			return
		default:
		}

		// Set a deadline so Accept() doesn't block indefinitely
		// This allows us to periodically check s.done
		// UnixListener supports SetDeadline
		if unixListener, ok := s.listener.(*net.UnixListener); ok {
			_ = unixListener.SetDeadline(time.Now().Add(100 * time.Millisecond))
		}

		conn, err := s.listener.Accept()
		if err != nil {
			// Check if error is due to timeout (expected) or shutdown
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout is expected, check if we should stop
				select {
				case <-s.done:
					return
				default:
					continue
				}
			}
			// Other errors (like closed listener) - check if we should stop
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}

		// Handle connection in a goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			// Connection closed or invalid message
			return
		}

		// Process message
		response := s.processMessage(&msg)

		// Send response
		if err := encoder.Encode(response); err != nil {
			return
		}
	}
}

// processMessage processes an IPC message and returns a response
func (s *Server) processMessage(msg *Message) *Response {
	switch msg.Type {
	case MessageTypeStatus:
		return s.handleStatus()

	case MessageTypeQueryExecute:
		return s.handleQueryExecute(msg)

	case MessageTypeReindexPaths:
		return s.handleReindexPaths(msg)

	case MessageTypeShutdown:
		return s.handleShutdown()

	default:
		return NewResponse(false, nil, fmt.Errorf("unknown message type: %s", msg.Type))
	}
}

// handleStatus handles a status request
func (s *Server) handleStatus() *Response {
	daemon := NewDaemon(s.vaultRoot)
	running, pid, err := daemon.Status()
	if err != nil {
		return NewResponse(false, nil, err)
	}

	return NewResponse(true, StatusResponse{
		Running: running,
		PID:     pid,
	}, nil)
}

// handleQueryExecute handles a query execution request
func (s *Server) handleQueryExecute(msg *Message) *Response {
	var req QueryExecuteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return NewResponse(false, nil, fmt.Errorf("parsing query request: %w", err))
	}

	// Parse query string into SearchQuery AST.
	// Supports key:value pairs like "type:note state:published tag:important limit:10".
	searchQuery, err := query.ParseSearchQuery(req.Query)
	if err != nil {
		return NewResponse(false, nil, fmt.Errorf("parsing query string: %w", err))
	}

	// Execute query
	results, err := query.ExecuteSearch(s.vaultRoot, searchQuery)
	if err != nil {
		return NewResponse(false, nil, err)
	}

	return NewResponse(true, QueryExecuteResponse{
		Results: results,
	}, nil)
}

// handleReindexPaths handles a reindex paths request
func (s *Server) handleReindexPaths(msg *Message) *Response {
	var req ReindexPathsRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return NewResponse(false, nil, fmt.Errorf("parsing reindex request: %w", err))
	}

	// Check if indexer is initialized (server must be started)
	if s.indexer == nil {
		return NewResponse(false, nil, fmt.Errorf("server not started: indexer not initialized"))
	}

	// Process each path
	processed := 0
	for _, path := range req.Paths {
		// Create a synthetic event for the path
		event := watch.Event{
			Path:      path,
			Op:        fsnotify.Write, // Assume write for reindex
			Timestamp: time.Now(),
		}

		if err := s.indexer.ProcessEvent(event); err != nil {
			// Log error but continue processing other paths
			continue
		}

		processed++
	}

	return NewResponse(true, ReindexPathsResponse{
		Processed: processed,
	}, nil)
}

// handleShutdown handles a shutdown request
func (s *Server) handleShutdown() *Response {
	// Signal shutdown
	go func() {
		time.Sleep(100 * time.Millisecond) // Give time for response
		s.once.Do(func() {
			close(s.done)
		})
	}()

	return NewResponse(true, nil, nil)
}

// processWatchEvents processes filesystem watch events
func (s *Server) processWatchEvents() {
	// Use a ticker to periodically check s.done
	// This ensures we can exit even if watcher channels never receive values
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case event := <-s.watcher.Events():
			// Process event with incremental indexer
			if s.indexer != nil {
				if err := s.indexer.ProcessEvent(event); err != nil {
					// Log error but continue processing
					// In a full implementation, we'd have proper logging
					_ = err
				}
			}

		case err := <-s.watcher.Errors():
			// Log error but continue
			_ = err

		case <-s.done:
			return

		case <-ticker.C:
			// Periodic check - if s.done is closed, we'll catch it in the next iteration
			// This ensures we can exit even if no events occur
			continue
		}
	}
}

package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/query"
	"github.com/sv4u/touchlog/internal/store"
	"github.com/sv4u/touchlog/internal/watch"
)

// Server handles IPC communication via Unix domain socket
type Server struct {
	vaultRoot string
	cfg       *config.Config
	listener  net.Listener
	watcher   *watch.Watcher
	indexer   *watch.IncrementalIndexer
	done      chan struct{}
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
		vaultRoot: vaultRoot,
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

	// Create incremental indexer (will use short-lived connections)
	db, err := store.OpenOrCreateDB(s.vaultRoot)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	s.indexer = watch.NewIncrementalIndexer(s.vaultRoot, s.cfg, db)
	_ = db.Close()

	// Start event processing goroutine
	go s.processWatchEvents()

	// Start accepting connections
	go s.acceptConnections()

	return nil
}

// Stop stops the IPC server
func (s *Server) Stop() error {
	close(s.done)
	if s.watcher != nil {
		_ = s.watcher.Stop()
	}
	return s.listener.Close()
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.done:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
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

	// Parse query from request
	// For Phase 3, we'll support basic search queries
	// In a full implementation, we'd parse the query string into a SearchQuery AST
	searchQuery := query.NewSearchQuery()
	// TODO: Parse req.Query into searchQuery

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
		close(s.done)
	}()

	return NewResponse(true, nil, nil)
}

// processWatchEvents processes filesystem watch events
func (s *Server) processWatchEvents() {
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
		}
	}
}

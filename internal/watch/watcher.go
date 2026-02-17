package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// pendingEvent tracks the most recent event for a debounced path
type pendingEvent struct {
	Op        fsnotify.Op
	Timestamp time.Time
}

// Watcher manages filesystem watching for a vault
type Watcher struct {
	vaultRoot string
	watcher   *fsnotify.Watcher
	events    chan Event
	errors    chan error
	done      chan struct{}
	mu        sync.Mutex
	debounce  time.Duration
	pending   map[string]pendingEvent
}

// Event represents a filesystem event
type Event struct {
	Path      string
	Op        fsnotify.Op
	Timestamp time.Time
}

// NewWatcher creates a new filesystem watcher
func NewWatcher(vaultRoot string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating fsnotify watcher: %w", err)
	}

	w := &Watcher{
		vaultRoot: vaultRoot,
		watcher:   fsWatcher,
		events:    make(chan Event, 100),
		errors:    make(chan error, 10),
		done:      make(chan struct{}),
		debounce:  200 * time.Millisecond,
		pending:   make(map[string]pendingEvent),
	}

	return w, nil
}

// Start starts watching the filesystem
func (w *Watcher) Start() error {
	// Add vault root recursively
	if err := w.addRecursive(w.vaultRoot); err != nil {
		return fmt.Errorf("adding watch paths: %w", err)
	}

	// Start event processing goroutine
	go w.processEvents()

	// Start watching loop
	go w.watchLoop()

	return nil
}

// Stop stops watching
func (w *Watcher) Stop() error {
	close(w.done)
	return w.watcher.Close()
}

// Events returns the events channel
func (w *Watcher) Events() <-chan Event {
	return w.events
}

// Errors returns the errors channel
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// addRecursive adds a directory and all subdirectories to the watcher
func (w *Watcher) addRecursive(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only watch directories (fsnotify watches directories, not files)
		// We'll get events for files in watched directories
		if info.IsDir() {
			// Skip .touchlog directory (contains index.db, etc.)
			if filepath.Base(p) == ".touchlog" {
				return filepath.SkipDir
			}

			if err := w.watcher.Add(p); err != nil {
				return fmt.Errorf("adding watch for %s: %w", p, err)
			}
		}

		return nil
	})
}

// watchLoop is the main watching loop
func (w *Watcher) watchLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Filter for .Rmd files only, but also handle newly created
			// directories so that .Rmd files inside them trigger events.
			if filepath.Ext(event.Name) != ".Rmd" {
				// For Create events, check if the path is a directory
				// regardless of its name (directories may contain dots,
				// e.g. "notes.2024", "v1.0").
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if filepath.Base(event.Name) != ".touchlog" {
							_ = w.watcher.Add(event.Name)
						}
					}
				}
				continue
			}

			// Debounce events
			w.debounceEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			select {
			case w.errors <- err:
			default:
				// Error channel full, skip
			}

		case <-w.done:
			return
		}
	}
}

// debounceEvent debounces filesystem events
func (w *Watcher) debounceEvent(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Record the most recent event for this path
	w.pending[event.Name] = pendingEvent{
		Op:        event.Op,
		Timestamp: time.Now(),
	}

	// Schedule debounced emission
	go func(path string) {
		time.Sleep(w.debounce)

		w.mu.Lock()
		defer w.mu.Unlock()

		// Check if this event is still pending for this path
		if pe, ok := w.pending[path]; ok {
			// Emit with the most recent Op (not the one that spawned this goroutine)
			select {
			case w.events <- Event{
				Path:      path,
				Op:        pe.Op,
				Timestamp: pe.Timestamp,
			}:
			default:
				// Event channel full, skip
			}

			// Remove from pending
			delete(w.pending, path)
		}
	}(event.Name)
}

// processEvents processes and coalesces events
func (w *Watcher) processEvents() {
	// This goroutine can be used for additional event processing
	// For Phase 3, we just pass events through
	<-w.done
}

package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sv4u/touchlog/internal/model"
)

// OpenOrCreateDB opens or creates the SQLite database at the vault root
func OpenOrCreateDB(vaultRoot string) (*sql.DB, error) {
	dbPath := filepath.Join(vaultRoot, ".touchlog", "index.db")

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

// ApplyMigrations applies all necessary migrations to bring the database to the current schema version
func ApplyMigrations(db *sql.DB) error {
	// Check current schema version
	currentVersion, err := getCurrentSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("getting current schema version: %w", err)
	}

	// Apply migrations if needed
	if currentVersion == 0 {
		// Fresh database, create schema v1
		if err := createSchemaV1(db); err != nil {
			return fmt.Errorf("creating schema v1: %w", err)
		}
	} else if currentVersion < model.IndexSchemaVersion {
		// Future: apply incremental migrations
		return fmt.Errorf("schema version %d requires migration (not implemented)", currentVersion)
	} else if currentVersion > model.IndexSchemaVersion {
		return fmt.Errorf("database schema version %d is newer than supported version %d", currentVersion, model.IndexSchemaVersion)
	}

	return nil
}

// getCurrentSchemaVersion returns the current schema version, or 0 if the database is empty
func getCurrentSchemaVersion(db *sql.DB) (int, error) {
	// Check if meta table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT name FROM sqlite_master 
			WHERE type='table' AND name='meta'
		)
	`).Scan(&exists)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	var version int
	err = db.QueryRow("SELECT schema_version FROM meta WHERE schema_version = ?", model.IndexSchemaVersion).Scan(&version)
	if err == sql.ErrNoRows {
		// Table exists but no row for current version, check for any version
		err = db.QueryRow("SELECT MAX(schema_version) FROM meta").Scan(&version)
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return version, err
	}
	return version, err
}

// createSchemaV1 creates the initial database schema
func createSchemaV1(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Create meta table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS meta (
			schema_version INTEGER PRIMARY KEY,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("creating meta table: %w", err)
	}

	// Create nodes table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			key TEXT NOT NULL,
			title TEXT NOT NULL,
			state TEXT NOT NULL,
			created TEXT NOT NULL,
			updated TEXT NOT NULL,
			path TEXT NOT NULL,
			mtime_ns INTEGER NOT NULL,
			size_bytes INTEGER NOT NULL,
			hash TEXT,
			UNIQUE(type, key)
		)
	`)
	if err != nil {
		return fmt.Errorf("creating nodes table: %w", err)
	}

	// Create edges table
	// Note: to_id may be NULL for unresolved links (Phase 0 doesn't resolve links)
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS edges (
			from_id TEXT NOT NULL,
			to_id TEXT,
			edge_type TEXT NOT NULL,
			raw_target TEXT NOT NULL,
			span TEXT NOT NULL,
			PRIMARY KEY(from_id, to_id, edge_type, raw_target, span),
			FOREIGN KEY(from_id) REFERENCES nodes(id) ON DELETE CASCADE,
			FOREIGN KEY(to_id) REFERENCES nodes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("creating edges table: %w", err)
	}

	// Create tags table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			node_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			PRIMARY KEY(node_id, tag),
			FOREIGN KEY(node_id) REFERENCES nodes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("creating tags table: %w", err)
	}

	// Create diagnostics table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS diagnostics (
			node_id TEXT NOT NULL,
			level TEXT NOT NULL,
			code TEXT NOT NULL,
			message TEXT NOT NULL,
			span TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(node_id) REFERENCES nodes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("creating diagnostics table: %w", err)
	}

	// Create indexes for performance optimization
	// Edge indexes (for graph traversal)
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_from_id ON edges(from_id)`)
	if err != nil {
		return fmt.Errorf("creating edges from_id index: %w", err)
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_to_id ON edges(to_id)`)
	if err != nil {
		return fmt.Errorf("creating edges to_id index: %w", err)
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_edge_type ON edges(edge_type)`)
	if err != nil {
		return fmt.Errorf("creating edges edge_type index: %w", err)
	}

	// Node indexes (for search queries)
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type)`)
	if err != nil {
		return fmt.Errorf("creating nodes type index: %w", err)
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_nodes_state ON nodes(state)`)
	if err != nil {
		return fmt.Errorf("creating nodes state index: %w", err)
	}

	// Note: nodes(type, key) already has UNIQUE constraint which acts as an index

	// Tag indexes (for tag filtering)
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_tags_tag ON tags(tag)`)
	if err != nil {
		return fmt.Errorf("creating tags tag index: %w", err)
	}

	// Diagnostic indexes (for diagnostics queries)
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_diagnostics_node_id ON diagnostics(node_id)`)
	if err != nil {
		return fmt.Errorf("creating diagnostics node_id index: %w", err)
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_diagnostics_level ON diagnostics(level)`)
	if err != nil {
		return fmt.Errorf("creating diagnostics level index: %w", err)
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_diagnostics_code ON diagnostics(code)`)
	if err != nil {
		return fmt.Errorf("creating diagnostics code index: %w", err)
	}

	// Insert schema version
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec(`
		INSERT INTO meta (schema_version, created_at, updated_at)
		VALUES (?, ?, ?)
	`, model.IndexSchemaVersion, now, now)
	if err != nil {
		return fmt.Errorf("inserting schema version: %w", err)
	}

	return tx.Commit()
}

// UpsertNode inserts or updates a node in the database
func UpsertNode(db *sql.DB, nodeID model.NoteID, nodeType model.TypeName, key model.Key, title, state string, created, updated time.Time, path string, mtimeNs int64, sizeBytes int64, hash string) error {
	_, err := db.Exec(`
		INSERT INTO nodes (id, type, key, title, state, created, updated, path, mtime_ns, size_bytes, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			key = excluded.key,
			title = excluded.title,
			state = excluded.state,
			created = excluded.created,
			updated = excluded.updated,
			path = excluded.path,
			mtime_ns = excluded.mtime_ns,
			size_bytes = excluded.size_bytes,
			hash = excluded.hash
	`, nodeID, nodeType, key, title, state, created.Format(time.RFC3339), updated.Format(time.RFC3339), path, mtimeNs, sizeBytes, hash)
	return err
}

// ReplaceEdgesForNode replaces all edges for a given node
func ReplaceEdgesForNode(db *sql.DB, fromID model.NoteID, edges []model.RawLink) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete existing edges
	_, err = tx.Exec("DELETE FROM edges WHERE from_id = ?", fromID)
	if err != nil {
		return fmt.Errorf("deleting existing edges: %w", err)
	}

	// Insert new edges
	for _, edge := range edges {
		// Serialize raw target and span to JSON for storage
		targetJSON, err := json.Marshal(edge.Target)
		if err != nil {
			return fmt.Errorf("marshaling target: %w", err)
		}

		spanJSON, err := json.Marshal(edge.Span)
		if err != nil {
			return fmt.Errorf("marshaling span: %w", err)
		}

		// Set to_id if link is resolved (Phase 2+)
		var toID *string
		if edge.ResolvedToID != nil {
			toIDStr := string(*edge.ResolvedToID)
			toID = &toIDStr
		}
		// If ResolvedToID is nil, to_id will be NULL (unresolved link)

		_, err = tx.Exec(`
			INSERT INTO edges (from_id, to_id, edge_type, raw_target, span)
			VALUES (?, ?, ?, ?, ?)
		`, fromID, toID, edge.EdgeType, string(targetJSON), string(spanJSON))
		if err != nil {
			return fmt.Errorf("inserting edge: %w", err)
		}
	}

	return tx.Commit()
}

// ReplaceTagsForNode replaces all tags for a given node
func ReplaceTagsForNode(db *sql.DB, nodeID model.NoteID, tags []string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete existing tags
	_, err = tx.Exec("DELETE FROM tags WHERE node_id = ?", nodeID)
	if err != nil {
		return fmt.Errorf("deleting existing tags: %w", err)
	}

	// Insert new tags
	for _, tag := range tags {
		_, err = tx.Exec("INSERT INTO tags (node_id, tag) VALUES (?, ?)", nodeID, tag)
		if err != nil {
			return fmt.Errorf("inserting tag: %w", err)
		}
	}

	return tx.Commit()
}

// InsertDiagnostics inserts diagnostics for a node
func InsertDiagnostics(db *sql.DB, nodeID model.NoteID, diags []model.Diagnostic) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete existing diagnostics for this node
	_, err = tx.Exec("DELETE FROM diagnostics WHERE node_id = ?", nodeID)
	if err != nil {
		return fmt.Errorf("deleting existing diagnostics: %w", err)
	}

	// Insert new diagnostics
	now := time.Now().UTC().Format(time.RFC3339)
	for _, diag := range diags {
		spanJSON, err := json.Marshal(diag.Span)
		if err != nil {
			return fmt.Errorf("marshaling span: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO diagnostics (node_id, level, code, message, span, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, nodeID, diag.Level, diag.Code, diag.Message, string(spanJSON), now)
		if err != nil {
			return fmt.Errorf("inserting diagnostic: %w", err)
		}
	}

	return tx.Commit()
}

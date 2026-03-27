// Package store provides a SQLite-backed result store that captures pipeline
// and agent run results with configurable retention policies.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // register the sqlite driver
)

// Run represents a recorded pipeline or agent run.
type Run struct {
	ID         int64
	Kind       string // "pipeline" | "agent"
	Name       string
	StartedAt  int64  // unix millis
	FinishedAt *int64 // nil if in-flight
	ExitStatus *int   // nil if in-flight
	Stdout     string
	Stderr     string
	Metadata   string // JSON blob
}

// Store manages the SQLite result database.
type Store struct {
	db     *sql.DB
	writer *writer
	cfg    RetentionConfig
}

// Open opens or creates the store at ~/.local/share/orcai/orcai.db.
// It enables WAL mode and applies the schema migration.
func Open() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("store: resolve home dir: %w", err)
	}
	path := filepath.Join(home, ".local", "share", "orcai", "orcai.db")
	return OpenAt(path)
}

// OpenAt opens or creates the store at the given path.
// It enables WAL mode and applies the schema migration.
// This function exists primarily for testing with t.TempDir().
func OpenAt(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("store: mkdir: %w", err)
	}

	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}

	if err := applySchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: apply schema: %w", err)
	}

	cfg := loadRetentionConfig()

	s := &Store{
		db:     db,
		writer: newWriter(db),
		cfg:    cfg,
	}
	return s, nil
}

// DB returns the underlying *sql.DB for direct read queries.
// Write operations should go through RecordRunStart/RecordRunComplete
// to benefit from the serialized write queue.
func (s *Store) DB() *sql.DB { return s.db }

// Close shuts down the writer goroutine and closes the database.
func (s *Store) Close() error {
	s.writer.close()
	return s.db.Close()
}

package store

import "database/sql"

// createSchema is the DDL for the runs table.
const createSchema = `CREATE TABLE IF NOT EXISTS runs (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  kind        TEXT NOT NULL,
  name        TEXT NOT NULL,
  started_at  INTEGER NOT NULL,
  finished_at INTEGER,
  exit_status INTEGER,
  stdout      TEXT,
  stderr      TEXT,
  metadata    TEXT
);`

// applySchema runs the schema migration against db.
func applySchema(db *sql.DB) error {
	_, err := db.Exec(createSchema)
	return err
}

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

// addStepsColumn is the migration that adds the steps column to an existing
// runs table that was created before this column existed.
const addStepsColumn = `ALTER TABLE runs ADD COLUMN steps TEXT DEFAULT '[]'`

// applySchema runs the schema migration against db.
func applySchema(db *sql.DB) error {
	if _, err := db.Exec(createSchema); err != nil {
		return err
	}
	return applyStepsColumnMigration(db)
}

// applyStepsColumnMigration adds the steps column if it does not already exist.
// modernc.org/sqlite does not support ALTER TABLE ... ADD COLUMN IF NOT EXISTS,
// so we probe pragma_table_info first.
func applyStepsColumnMigration(db *sql.DB) error {
	var count int
	row := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('runs') WHERE name='steps'`)
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := db.Exec(addStepsColumn); err != nil {
			return err
		}
	}
	return nil
}

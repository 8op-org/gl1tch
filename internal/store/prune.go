package store

import "database/sql"

// autoPruneDB performs pruning directly on db without going through the writer.
// It must only be called from within a writer.send callback.
func autoPruneDB(db *sql.DB, maxAgeDays, maxRows int) error {
	// Delete rows older than maxAgeDays.
	cutoffMillis := int64(maxAgeDays) * 86400000
	_, err := db.Exec(
		`DELETE FROM runs WHERE started_at < (CAST(strftime('%s', 'now') AS INTEGER) * 1000 - ?)`,
		cutoffMillis,
	)
	if err != nil {
		return err
	}

	// Delete oldest rows when total count exceeds maxRows.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&count); err != nil {
		return err
	}
	if count > maxRows {
		excess := count - maxRows
		_, err = db.Exec(
			`DELETE FROM runs WHERE id IN (
				SELECT id FROM runs ORDER BY started_at ASC LIMIT ?
			)`,
			excess,
		)
		return err
	}
	return nil
}

// AutoPrune deletes rows older than maxAgeDays OR when total rows exceed maxRows.
// Called automatically after RecordRunComplete, and available for direct use.
func (s *Store) AutoPrune(maxAgeDays, maxRows int) error {
	return s.writer.send(func(db *sql.DB) error {
		return autoPruneDB(db, maxAgeDays, maxRows)
	})
}

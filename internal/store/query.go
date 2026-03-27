package store

import "database/sql"

// QueryRuns returns up to limit runs ordered by started_at descending.
// Reads use s.db directly since WAL mode allows concurrent readers.
func (s *Store) QueryRuns(limit int) ([]Run, error) {
	rows, err := s.db.Query(
		`SELECT id, kind, name, started_at, finished_at, exit_status, stdout, stderr, metadata
		   FROM runs
		  ORDER BY started_at DESC
		  LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		var finishedAt sql.NullInt64
		var exitStatus sql.NullInt64
		var stdout, stderr, metadata sql.NullString

		if err := rows.Scan(
			&r.ID, &r.Kind, &r.Name, &r.StartedAt,
			&finishedAt, &exitStatus,
			&stdout, &stderr, &metadata,
		); err != nil {
			return nil, err
		}

		if finishedAt.Valid {
			r.FinishedAt = &finishedAt.Int64
		}
		if exitStatus.Valid {
			v := int(exitStatus.Int64)
			r.ExitStatus = &v
		}
		r.Stdout = stdout.String
		r.Stderr = stderr.String
		r.Metadata = metadata.String

		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// DeleteRun removes the run with the given id.
func (s *Store) DeleteRun(id int64) error {
	return s.writer.send(func(db *sql.DB) error {
		_, err := db.Exec(`DELETE FROM runs WHERE id = ?`, id)
		return err
	})
}

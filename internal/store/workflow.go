package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// WorkflowRun represents a recorded workflow orchestration run.
type WorkflowRun struct {
	ID          int64
	Name        string
	Status      string
	Input       string
	Output      string
	Error       string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// WorkflowCheckpoint is a saved step state within a workflow run.
type WorkflowCheckpoint struct {
	ID          int64
	RunID       int64
	StepID      string
	Status      string
	ContextJSON string
	CreatedAt   time.Time
}

// CreateWorkflowRun inserts a new workflow run record and returns its ID.
func (s *Store) CreateWorkflowRun(ctx context.Context, name, input string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO workflow_runs (name, status, input) VALUES (?, 'running', ?)`,
		name, input,
	)
	if err != nil {
		return 0, fmt.Errorf("store: create workflow run: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("store: create workflow run last id: %w", err)
	}
	return id, nil
}

// CompleteWorkflowRun marks a workflow run as completed with the given output.
func (s *Store) CompleteWorkflowRun(ctx context.Context, id int64, output string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE workflow_runs SET status = 'completed', output = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`,
		output, id,
	)
	if err != nil {
		return fmt.Errorf("store: complete workflow run: %w", err)
	}
	return nil
}

// FailWorkflowRun marks a workflow run as failed with the given error message.
func (s *Store) FailWorkflowRun(ctx context.Context, id int64, errMsg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE workflow_runs SET status = 'failed', error = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`,
		errMsg, id,
	)
	if err != nil {
		return fmt.Errorf("store: fail workflow run: %w", err)
	}
	return nil
}

// GetWorkflowRun retrieves a workflow run by ID.
func (s *Store) GetWorkflowRun(ctx context.Context, id int64) (*WorkflowRun, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, status, COALESCE(input,''), COALESCE(output,''), COALESCE(error,''), created_at, completed_at
		   FROM workflow_runs WHERE id = ?`,
		id,
	)
	var wr WorkflowRun
	var completedAt sql.NullTime
	var createdAtStr string
	if err := row.Scan(&wr.ID, &wr.Name, &wr.Status, &wr.Input, &wr.Output, &wr.Error, &createdAtStr, &completedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("store: workflow run %d not found", id)
		}
		return nil, fmt.Errorf("store: get workflow run: %w", err)
	}
	// Parse created_at from SQLite DATETIME string.
	if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
		wr.CreatedAt = t
	}
	if completedAt.Valid {
		t := completedAt.Time
		wr.CompletedAt = &t
	}
	return &wr, nil
}

// SaveWorkflowCheckpoint inserts a checkpoint for a step in a workflow run.
func (s *Store) SaveWorkflowCheckpoint(ctx context.Context, runID int64, stepID, status, contextJSON string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO workflow_checkpoints (run_id, step_id, status, context_json) VALUES (?, ?, ?, ?)`,
		runID, stepID, status, contextJSON,
	)
	if err != nil {
		return fmt.Errorf("store: save workflow checkpoint: %w", err)
	}
	return nil
}

// LoadWorkflowCheckpoints returns all checkpoints for a workflow run ordered by creation time.
func (s *Store) LoadWorkflowCheckpoints(ctx context.Context, runID int64) ([]WorkflowCheckpoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, run_id, step_id, status, context_json, created_at
		   FROM workflow_checkpoints
		  WHERE run_id = ?
		  ORDER BY id ASC`,
		runID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: load workflow checkpoints: %w", err)
	}
	defer rows.Close()

	var checkpoints []WorkflowCheckpoint
	for rows.Next() {
		var cp WorkflowCheckpoint
		var createdAtStr string
		if err := rows.Scan(&cp.ID, &cp.RunID, &cp.StepID, &cp.Status, &cp.ContextJSON, &createdAtStr); err != nil {
			return nil, fmt.Errorf("store: load workflow checkpoints scan: %w", err)
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			cp.CreatedAt = t
		}
		checkpoints = append(checkpoints, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: load workflow checkpoints rows: %w", err)
	}
	return checkpoints, nil
}

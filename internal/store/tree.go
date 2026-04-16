package store

import "fmt"

// RunNode is a minimal run record used by tree queries.
type RunNode struct {
	ID           int64
	Name         string
	WorkflowName string
	Kind         string
	ExitStatus   *int
	StartedAt    int64
	FinishedAt   *int64
	ParentRunID  int64
	Children     []RunNode
}

// ListChildren returns direct children of a run.
func (s *Store) ListChildren(parentID int64) ([]RunNode, error) {
	rows, err := s.db.Query(
		`SELECT id, name, COALESCE(workflow_name,''), kind, exit_status, started_at, finished_at, COALESCE(parent_run_id,0)
		 FROM runs WHERE parent_run_id = ? ORDER BY id ASC`, parentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RunNode
	for rows.Next() {
		var n RunNode
		var exitInt *int
		var fin *int64
		if err := rows.Scan(&n.ID, &n.Name, &n.WorkflowName, &n.Kind, &exitInt, &n.StartedAt, &fin, &n.ParentRunID); err != nil {
			return nil, err
		}
		n.ExitStatus = exitInt
		n.FinishedAt = fin
		out = append(out, n)
	}
	return out, rows.Err()
}

// GetRunTree returns the run rooted at id with all descendants populated.
// Depth is capped to guard against pathological cycles in corrupted data;
// call-workflow cycle detection is the real guard upstream.
func (s *Store) GetRunTree(id int64) (RunNode, error) {
	row := s.db.QueryRow(
		`SELECT id, name, COALESCE(workflow_name,''), kind, exit_status, started_at, finished_at, COALESCE(parent_run_id,0)
		 FROM runs WHERE id = ?`, id,
	)
	var n RunNode
	var exit *int
	var fin *int64
	if err := row.Scan(&n.ID, &n.Name, &n.WorkflowName, &n.Kind, &exit, &n.StartedAt, &fin, &n.ParentRunID); err != nil {
		return RunNode{}, err
	}
	n.ExitStatus = exit
	n.FinishedAt = fin
	return s.populateChildren(n, 0)
}

const maxTreeDepth = 64

func (s *Store) populateChildren(n RunNode, depth int) (RunNode, error) {
	if depth > maxTreeDepth {
		return RunNode{}, fmt.Errorf("run tree depth exceeded %d at id=%d", maxTreeDepth, n.ID)
	}
	kids, err := s.ListChildren(n.ID)
	if err != nil {
		return RunNode{}, err
	}
	for i := range kids {
		sub, err := s.populateChildren(kids[i], depth+1)
		if err != nil {
			return RunNode{}, err
		}
		kids[i] = sub
	}
	n.Children = kids
	return n, nil
}

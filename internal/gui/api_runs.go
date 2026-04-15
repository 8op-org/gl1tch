package gui

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type runEntry struct {
	ID         int64  `json:"id"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Input      string `json:"input"`
	Output     string `json:"output,omitempty"`
	ExitStatus int    `json:"exit_status"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at,omitempty"`
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]runEntry{})
		return
	}
	rows, err := s.store.DB().Query(
		`SELECT id, kind, name, COALESCE(input,''), COALESCE(output,''),
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0)
		 FROM runs ORDER BY id DESC LIMIT 100`,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var runs []runEntry
	for rows.Next() {
		var re runEntry
		if err := rows.Scan(&re.ID, &re.Kind, &re.Name, &re.Input, &re.Output,
			&re.ExitStatus, &re.StartedAt, &re.FinishedAt); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		runs = append(runs, re)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if runs == nil {
		runs = []runEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "store not available", 500)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}

	var run runEntry
	err = s.store.DB().QueryRow(
		`SELECT id, kind, name, COALESCE(input,''), COALESCE(output,''),
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0)
		 FROM runs WHERE id = ?`, id,
	).Scan(&run.ID, &run.Kind, &run.Name, &run.Input, &run.Output,
		&run.ExitStatus, &run.StartedAt, &run.FinishedAt)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	type stepEntry struct {
		StepID     string `json:"step_id"`
		Model      string `json:"model"`
		DurationMs int64  `json:"duration_ms"`
	}

	stepRows, _ := s.store.DB().Query(
		`SELECT step_id, COALESCE(model,''), COALESCE(duration_ms,0)
		 FROM steps WHERE run_id = ?`, id)
	var steps []stepEntry
	if stepRows != nil {
		defer stepRows.Close()
		for stepRows.Next() {
			var se stepEntry
			if err := stepRows.Scan(&se.StepID, &se.Model, &se.DurationMs); err != nil {
				continue
			}
			steps = append(steps, se)
		}
	}
	if steps == nil {
		steps = []stepEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"run":   run,
		"steps": steps,
	})
}

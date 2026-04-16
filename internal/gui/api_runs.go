package gui

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/8op-org/gl1tch/internal/store"
)

type runEntry struct {
	ID           int64   `json:"id"`
	Kind         string  `json:"kind"`
	Name         string  `json:"name"`
	Input        string  `json:"input"`
	Output       string  `json:"output,omitempty"`
	ExitStatus   int     `json:"exit_status"`
	StartedAt    int64   `json:"started_at"`
	FinishedAt   int64   `json:"finished_at,omitempty"`
	WorkflowFile string  `json:"workflow_file,omitempty"`
	Repo         string  `json:"repo,omitempty"`
	Model        string  `json:"model,omitempty"`
	TokensIn     int64   `json:"tokens_in"`
	TokensOut    int64   `json:"tokens_out"`
	CostUSD      float64 `json:"cost_usd"`
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]runEntry{})
		return
	}
	if p := r.URL.Query().Get("parent_id"); p != "" {
		parentID, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			http.Error(w, "invalid parent_id", http.StatusBadRequest)
			return
		}
		kids, err := s.store.ListChildren(parentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if kids == nil {
			kids = []store.RunNode{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(kids)
		return
	}
	rows, err := s.store.DB().Query(
		`SELECT id, kind, name, COALESCE(input,''), COALESCE(output,''),
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0),
		        COALESCE(workflow_file,''), COALESCE(repo,''), COALESCE(model,''),
		        COALESCE(tokens_in,0), COALESCE(tokens_out,0), COALESCE(cost_usd,0)
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
			&re.ExitStatus, &re.StartedAt, &re.FinishedAt,
			&re.WorkflowFile, &re.Repo, &re.Model,
			&re.TokensIn, &re.TokensOut, &re.CostUSD); err != nil {
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
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0),
		        COALESCE(workflow_file,''), COALESCE(repo,''), COALESCE(model,''),
		        COALESCE(tokens_in,0), COALESCE(tokens_out,0), COALESCE(cost_usd,0)
		 FROM runs WHERE id = ?`, id,
	).Scan(&run.ID, &run.Kind, &run.Name, &run.Input, &run.Output,
		&run.ExitStatus, &run.StartedAt, &run.FinishedAt,
		&run.WorkflowFile, &run.Repo, &run.Model,
		&run.TokensIn, &run.TokensOut, &run.CostUSD)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	type stepEntry struct {
		StepID     string `json:"step_id"`
		Model      string `json:"model"`
		DurationMs int64  `json:"duration_ms"`
		Kind       string `json:"kind,omitempty"`
		ExitStatus *int   `json:"exit_status,omitempty"`
		TokensIn   int64  `json:"tokens_in"`
		TokensOut  int64  `json:"tokens_out"`
		GatePassed *bool  `json:"gate_passed,omitempty"`
	}

	stepRows, err := s.store.DB().Query(
		`SELECT step_id, COALESCE(model,''), COALESCE(duration_ms,0),
		        COALESCE(kind,''), exit_status, COALESCE(tokens_in,0),
		        COALESCE(tokens_out,0), gate_passed
		 FROM steps WHERE run_id = ?`, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer stepRows.Close()

	var steps []stepEntry
	for stepRows.Next() {
		var se stepEntry
		var exitStatus, gatePassed sql.NullInt64
		if err := stepRows.Scan(&se.StepID, &se.Model, &se.DurationMs,
			&se.Kind, &exitStatus, &se.TokensIn, &se.TokensOut, &gatePassed); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if exitStatus.Valid {
			v := int(exitStatus.Int64)
			se.ExitStatus = &v
		}
		if gatePassed.Valid {
			v := gatePassed.Int64 == 1
			se.GatePassed = &v
		}
		steps = append(steps, se)
	}
	if err := stepRows.Err(); err != nil {
		http.Error(w, err.Error(), 500)
		return
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

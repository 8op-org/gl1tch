package store

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestOpenForWorkspace(t *testing.T) {
	wsDir := t.TempDir()
	s, err := OpenForWorkspace(wsDir)
	if err != nil {
		t.Fatalf("OpenForWorkspace: %v", err)
	}
	defer s.Close()

	dbPath := filepath.Join(wsDir, ".glitch", "glitch.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected DB at %s: %v", dbPath, err)
	}

	id, err := s.RecordRun(RunRecord{Kind: "test", Name: "ws-test", Input: ""})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestOpenAt_BusyTimeout(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	var timeout int
	err = s.db.QueryRow("PRAGMA busy_timeout").Scan(&timeout)
	if err != nil {
		t.Fatalf("PRAGMA busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Fatalf("busy_timeout: got %d, want 5000", timeout)
	}
}

func TestRecordRun(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	id, err := s.RecordRun(RunRecord{Kind: "pipeline", Name: "test-run", Input: "some input"})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	if err := s.FinishRun(id, "some output", 0); err != nil {
		t.Fatalf("FinishRun: %v", err)
	}
}

func TestRecordStep(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	runID, err := s.RecordRun(RunRecord{Kind: "pipeline", Name: "step-test", Input: ""})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	if err := s.RecordStep(StepRecord{
		RunID: runID, StepID: "step-1", Prompt: "my prompt",
		Output: "model output", Model: "qwen2.5:7b", DurationMs: 123,
	}); err != nil {
		t.Fatalf("RecordStep: %v", err)
	}

	// Insert OR REPLACE — should not error on duplicate step_id
	if err := s.RecordStep(StepRecord{
		RunID: runID, StepID: "step-1", Prompt: "updated prompt",
		Output: "new output", Model: "qwen2.5:7b", DurationMs: 456,
	}); err != nil {
		t.Fatalf("RecordStep (replace): %v", err)
	}
}

func TestRecordResearchEvent(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	evt := ResearchEvent{
		QueryID:        "q1",
		Question:       "What is the deployment frequency?",
		Researchers:    "git,es",
		CompositeScore: 0.85,
		Reason:         "high confidence from git data",
	}
	if err := s.RecordResearchEvent(evt); err != nil {
		t.Fatalf("RecordResearchEvent: %v", err)
	}
}

func TestRecordRunEnriched(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	id, err := s.RecordRun(RunRecord{
		Kind:         "workflow",
		Name:         "pr-review",
		Input:        "review this",
		WorkflowFile: "pr-review.glitch",
		Repo:         "elastic/ensemble",
		Model:        "qwen2.5:7b",
	})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	if err := s.FinishRun(id, "done", 0, RunTotals{
		TokensIn:  1500,
		TokensOut: 300,
		CostUSD:   0.005,
	}); err != nil {
		t.Fatalf("FinishRun: %v", err)
	}
}

func TestRecordStepEnriched(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	runID, _ := s.RecordRun(RunRecord{Kind: "workflow", Name: "test", Input: ""})

	err = s.RecordStep(StepRecord{
		RunID:      runID,
		StepID:     "fetch",
		Prompt:     "echo hello",
		Output:     "hello",
		Model:      "qwen2.5:7b",
		DurationMs: 150,
		Kind:       "run",
		ExitStatus: intPtr(0),
	})
	if err != nil {
		t.Fatalf("RecordStep: %v", err)
	}

	err = s.RecordStep(StepRecord{
		RunID:      runID,
		StepID:     "gate-check",
		Prompt:     "verify",
		Output:     "PASS",
		Model:      "",
		DurationMs: 50,
		Kind:       "gate",
		ExitStatus: intPtr(0),
		GatePassed: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("RecordStep gate: %v", err)
	}
}

func TestRecordRun_Workspace(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	id, err := s.RecordRun(RunRecord{
		Kind:      "pipeline",
		Name:      "test-run",
		Input:     "some input",
		Workspace: "stokagent",
	})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	var ws string
	err = s.db.QueryRow("SELECT workspace FROM runs WHERE id = ?", id).Scan(&ws)
	if err != nil {
		t.Fatalf("query workspace: %v", err)
	}
	if ws != "stokagent" {
		t.Fatalf("workspace: got %q, want stokagent", ws)
	}
}

func TestRecordStepArtifacts(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	runID, err := s.RecordRun(RunRecord{Kind: "pipeline", Name: "artifact-test", Input: ""})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	artifacts := []string{"/tmp/results/output.md", "/tmp/results/summary.json"}
	err = s.RecordStep(StepRecord{
		RunID:      runID,
		StepID:     "save-step",
		Output:     "saved",
		Kind:       "save",
		ExitStatus: intPtr(0),
		Artifacts:  artifacts,
	})
	if err != nil {
		t.Fatalf("RecordStep with artifacts: %v", err)
	}

	var raw string
	err = s.db.QueryRow("SELECT artifacts FROM steps WHERE run_id = ? AND step_id = ?", runID, "save-step").Scan(&raw)
	if err != nil {
		t.Fatalf("query artifacts: %v", err)
	}

	var got []string
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal artifacts: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(got))
	}
	if got[0] != artifacts[0] || got[1] != artifacts[1] {
		t.Fatalf("artifacts mismatch: got %v, want %v", got, artifacts)
	}
}

func intPtr(n int) *int    { return &n }
func boolPtr(b bool) *bool { return &b }

func TestOpenAt_RecreatesStaleSchema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Create a DB with a runs table missing the "input" column
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		kind TEXT NOT NULL,
		name TEXT NOT NULL,
		exit_status INTEGER,
		started_at INTEGER NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create stale table: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE steps (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("create stale steps: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE research_events (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("create stale research_events: %v", err)
	}
	db.Close()

	// OpenAt should detect the missing "input" column, drop, and recreate
	s, err := OpenAt(dbPath)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	// Verify the input column exists by inserting a run with input
	_, err = s.RecordRun(RunRecord{Kind: "test", Name: "drift-test", Input: "hello"})
	if err != nil {
		t.Fatalf("RecordRun after drift fix should succeed: %v", err)
	}
}

func TestSimilarResearchEvents(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	evtA := ResearchEvent{
		QueryID:        "qa",
		Question:       "What is the deployment frequency for this service?",
		Researchers:    "git",
		CompositeScore: 0.9,
		Reason:         "based on git tags",
	}
	evtB := ResearchEvent{
		QueryID:        "qb",
		Question:       "How many open pull requests exist in the repository?",
		Researchers:    "es",
		CompositeScore: 0.7,
		Reason:         "es index data",
	}
	if err := s.RecordResearchEvent(evtA); err != nil {
		t.Fatalf("seed evtA: %v", err)
	}
	if err := s.RecordResearchEvent(evtB); err != nil {
		t.Fatalf("seed evtB: %v", err)
	}

	results, err := s.SimilarResearchEvents("What is the release deployment frequency?", 2)
	if err != nil {
		t.Fatalf("SimilarResearchEvents: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	// The deployment-frequency event (evtA) should be ranked first.
	if results[0].QueryID != "qa" {
		t.Errorf("expected first result to be 'qa' (deployment frequency), got %q", results[0].QueryID)
	}
}

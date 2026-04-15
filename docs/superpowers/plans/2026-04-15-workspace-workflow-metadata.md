# Workspace & Workflow Metadata Enrichment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add workspace manifest parsing, workflow metadata keywords, and enriched SQLite columns so the GUI can present richer context about workspaces, workflows, and runs.

**Architecture:** Three additive layers — (1) new `workspace.glitch` parser producing a `Workspace` struct, (2) four new optional keywords on the existing `(workflow ...)` form, (3) enriched SQLite schema with new columns on `runs`/`steps` tables. A new `/api/workspace` GUI endpoint and enriched payloads on existing endpoints surface the data.

**Tech Stack:** Go, s-expression parser (`internal/sexpr`), SQLite (`internal/store`), net/http GUI (`internal/gui`)

---

### Task 1: Add Workspace Types and Parser

**Files:**
- Create: `internal/workspace/workspace.go`
- Create: `internal/workspace/workspace_test.go`

- [ ] **Step 1: Write the failing test for workspace parsing**

```go
package workspace

import (
	"testing"
)

func TestParseWorkspace_Full(t *testing.T) {
	src := []byte(`
(workspace "stokagent"
  :description "Cross-repo research and automation"
  :owner "adam"

  (repos
    "elastic/observability-robots"
    "elastic/ensemble")

  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"))
`)
	ws, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "stokagent" {
		t.Errorf("expected name stokagent, got %q", ws.Name)
	}
	if ws.Description != "Cross-repo research and automation" {
		t.Errorf("description mismatch: %q", ws.Description)
	}
	if ws.Owner != "adam" {
		t.Errorf("owner mismatch: %q", ws.Owner)
	}
	if len(ws.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(ws.Repos))
	}
	if ws.Repos[0] != "elastic/observability-robots" {
		t.Errorf("repo 0: %q", ws.Repos[0])
	}
	if ws.Defaults.Model != "qwen2.5:7b" {
		t.Errorf("default model: %q", ws.Defaults.Model)
	}
	if ws.Defaults.Provider != "ollama" {
		t.Errorf("default provider: %q", ws.Defaults.Provider)
	}
}

func TestParseWorkspace_Minimal(t *testing.T) {
	src := []byte(`(workspace "minimal")`)
	ws, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "minimal" {
		t.Errorf("expected name minimal, got %q", ws.Name)
	}
	if len(ws.Repos) != 0 {
		t.Errorf("expected no repos, got %d", len(ws.Repos))
	}
}

func TestParseWorkspace_NoWorkspaceForm(t *testing.T) {
	src := []byte(`(workflow "not-a-workspace")`)
	_, err := ParseFile(src)
	if err == nil {
		t.Fatal("expected error for missing (workspace ...) form")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -v -run TestParseWorkspace`
Expected: FAIL — package does not exist

- [ ] **Step 3: Write the workspace parser**

```go
package workspace

import (
	"fmt"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Workspace represents a parsed workspace.glitch manifest.
type Workspace struct {
	Name        string
	Description string
	Owner       string
	Repos       []string
	Defaults    Defaults
}

// Defaults holds workspace-level config overrides.
type Defaults struct {
	Model    string
	Provider string
}

// ParseFile parses workspace.glitch source bytes into a Workspace.
func ParseFile(src []byte) (*Workspace, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.IsList() && len(n.Children) > 0 && n.Children[0].StringVal() == "workspace" {
			return convertWorkspace(n)
		}
	}
	return nil, fmt.Errorf("no (workspace ...) form found")
}

func convertWorkspace(n *sexpr.Node) (*Workspace, error) {
	children := n.Children[1:]
	if len(children) == 0 {
		return nil, fmt.Errorf("line %d: workspace missing name", n.Line)
	}

	ws := &Workspace{}
	ws.Name = children[0].StringVal()
	if ws.Name == "" {
		return nil, fmt.Errorf("line %d: workspace name must be a string", children[0].Line)
	}
	children = children[1:]

	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "description":
				ws.Description = val.StringVal()
			case "owner":
				ws.Owner = val.StringVal()
			default:
				return nil, fmt.Errorf("line %d: unknown workspace keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].StringVal()
			switch head {
			case "repos":
				for _, r := range child.Children[1:] {
					ws.Repos = append(ws.Repos, r.StringVal())
				}
			case "defaults":
				d, err := parseDefaults(child)
				if err != nil {
					return nil, err
				}
				ws.Defaults = d
			default:
				return nil, fmt.Errorf("line %d: unknown workspace form %q", child.Line, head)
			}
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in workspace", child.Line)
	}
	if ws.Repos == nil {
		ws.Repos = []string{}
	}
	return ws, nil
}

func parseDefaults(n *sexpr.Node) (Defaults, error) {
	d := Defaults{}
	children := n.Children[1:]
	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return d, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "model":
				d.Model = val.StringVal()
			case "provider":
				d.Provider = val.StringVal()
			default:
				return d, fmt.Errorf("line %d: unknown defaults keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		return d, fmt.Errorf("line %d: unexpected form in defaults", child.Line)
	}
	return d, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -v -run TestParseWorkspace`
Expected: PASS (all 3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/workspace/workspace.go internal/workspace/workspace_test.go
git commit -m "feat(workspace): add workspace.glitch parser and types"
```

---

### Task 2: Add Workflow Metadata Fields

**Files:**
- Modify: `internal/pipeline/types.go:13-18`
- Modify: `internal/pipeline/sexpr.go:50-81` (convertWorkflow)
- Modify: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write the failing test for workflow metadata parsing**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Metadata(t *testing.T) {
	src := []byte(`
(workflow "pr-review"
  :description "Review PRs"
  :tags ("review" "ci" "code-quality")
  :author "adam"
  :version "1.0"
  :created "2026-04-01"
  (step "s1" (run "echo hi")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "pr-review" {
		t.Errorf("name: %q", w.Name)
	}
	if len(w.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(w.Tags))
	}
	if w.Tags[0] != "review" || w.Tags[1] != "ci" || w.Tags[2] != "code-quality" {
		t.Errorf("tags: %v", w.Tags)
	}
	if w.Author != "adam" {
		t.Errorf("author: %q", w.Author)
	}
	if w.Version != "1.0" {
		t.Errorf("version: %q", w.Version)
	}
	if w.Created != "2026-04-01" {
		t.Errorf("created: %q", w.Created)
	}
}

func TestSexprWorkflow_NoMetadata(t *testing.T) {
	src := []byte(`(workflow "bare" (step "s1" (run "echo hi")))`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Tags) != 0 {
		t.Errorf("expected no tags, got %v", w.Tags)
	}
	if w.Author != "" {
		t.Errorf("expected empty author, got %q", w.Author)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestSexprWorkflow_Metadata`
Expected: FAIL — `w.Tags` field does not exist

- [ ] **Step 3: Add metadata fields to Workflow struct**

In `internal/pipeline/types.go`, change the Workflow struct:

```go
// Workflow is a named sequence of steps loaded from YAML.
type Workflow struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Tags        []string       `yaml:"tags,omitempty"`
	Author      string         `yaml:"author,omitempty"`
	Version     string         `yaml:"version,omitempty"`
	Created     string         `yaml:"created,omitempty"`
	Steps       []Step         `yaml:"steps"`
	Items       []WorkflowItem `yaml:"-"`
}
```

- [ ] **Step 4: Handle new keywords in convertWorkflow**

In `internal/pipeline/sexpr.go`, update the keyword switch in `convertWorkflow()`. Replace the `default` case that errors on unknown keywords:

```go
			switch key {
			case "description":
				w.Description = resolveVal(val, defs)
			case "author":
				w.Author = resolveVal(val, defs)
			case "version":
				w.Version = resolveVal(val, defs)
			case "created":
				w.Created = resolveVal(val, defs)
			case "tags":
				if val.IsList() {
					for _, t := range val.Children {
						w.Tags = append(w.Tags, resolveVal(t, defs))
					}
				} else {
					w.Tags = append(w.Tags, resolveVal(val, defs))
				}
			default:
				return nil, fmt.Errorf("line %d: unknown workflow keyword :%s", child.Line, key)
			}
```

- [ ] **Step 5: Initialize Tags to empty slice**

After the existing `w := &Workflow{}` line in `convertWorkflow()`, add:

```go
	w.Tags = []string{}
```

This ensures `Tags` is never nil (matches the `NoMetadata` test expectation and JSON serialization).

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestSexprWorkflow`
Expected: PASS (all workflow tests including new metadata tests)

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): add tags, author, version, created workflow metadata keywords"
```

---

### Task 3: Enrich SQLite Schema

**Files:**
- Modify: `internal/store/schema.go`
- Modify: `internal/store/store.go:65-93`
- Modify: `internal/store/store_test.go`

- [ ] **Step 1: Write the failing test for enriched RecordRun**

Add to `internal/store/store_test.go`:

```go
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

func intPtr(n int) *int       { return &n }
func boolPtr(b bool) *bool    { return &b }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ -v -run TestRecordRunEnriched`
Expected: FAIL — `RunRecord` type does not exist

- [ ] **Step 3: Update schema with new columns**

In `internal/store/schema.go`:

```go
const createSchema = `
CREATE TABLE IF NOT EXISTS runs (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  kind          TEXT NOT NULL,
  name          TEXT NOT NULL,
  input         TEXT,
  output        TEXT,
  exit_status   INTEGER,
  started_at    INTEGER NOT NULL,
  finished_at   INTEGER,
  metadata      TEXT,
  workflow_file TEXT,
  repo          TEXT,
  model         TEXT,
  tokens_in     INTEGER,
  tokens_out    INTEGER,
  cost_usd      REAL,
  variant       TEXT
);

CREATE TABLE IF NOT EXISTS steps (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id      INTEGER NOT NULL,
  step_id     TEXT NOT NULL,
  prompt      TEXT,
  output      TEXT,
  model       TEXT,
  duration_ms INTEGER,
  kind        TEXT,
  exit_status INTEGER,
  tokens_in   INTEGER,
  tokens_out  INTEGER,
  gate_passed INTEGER,
  UNIQUE(run_id, step_id)
);

CREATE TABLE IF NOT EXISTS research_events (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  query_id        TEXT NOT NULL,
  question        TEXT NOT NULL,
  researchers     TEXT NOT NULL,
  composite_score REAL,
  reason          TEXT,
  created_at      INTEGER NOT NULL
);
`
```

- [ ] **Step 4: Add record types and update store methods**

In `internal/store/store.go`, add the new types and update the methods. Replace `RecordRun`, `FinishRun`, and `RecordStep`:

```go
// RunRecord holds the fields for inserting a new run.
type RunRecord struct {
	Kind         string
	Name         string
	Input        string
	WorkflowFile string
	Repo         string
	Model        string
	Variant      string
}

// RunTotals holds accumulated totals for finishing a run.
type RunTotals struct {
	TokensIn  int64
	TokensOut int64
	CostUSD   float64
}

// StepRecord holds the fields for inserting a step.
type StepRecord struct {
	RunID      int64
	StepID     string
	Prompt     string
	Output     string
	Model      string
	DurationMs int64
	Kind       string
	ExitStatus *int
	TokensIn   int64
	TokensOut  int64
	GatePassed *bool
}

// RecordRun inserts a new run record and returns the new row ID.
func (s *Store) RecordRun(rec RunRecord) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO runs (kind, name, input, started_at, workflow_file, repo, model, variant)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.Kind, rec.Name, rec.Input, time.Now().UnixMilli(),
		rec.WorkflowFile, rec.Repo, rec.Model, rec.Variant,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FinishRun updates an existing run with its output, exit status, and totals.
func (s *Store) FinishRun(id int64, output string, exitStatus int, totals ...RunTotals) error {
	if len(totals) > 0 {
		t := totals[0]
		_, err := s.db.Exec(
			`UPDATE runs SET output = ?, exit_status = ?, finished_at = ?,
			 tokens_in = ?, tokens_out = ?, cost_usd = ? WHERE id = ?`,
			output, exitStatus, time.Now().UnixMilli(),
			t.TokensIn, t.TokensOut, t.CostUSD, id,
		)
		return err
	}
	_, err := s.db.Exec(
		`UPDATE runs SET output = ?, exit_status = ?, finished_at = ? WHERE id = ?`,
		output, exitStatus, time.Now().UnixMilli(), id,
	)
	return err
}

// RecordStep inserts or replaces a step record for a given run.
func (s *Store) RecordStep(rec StepRecord) error {
	var gatePassed *int
	if rec.GatePassed != nil {
		v := 0
		if *rec.GatePassed {
			v = 1
		}
		gatePassed = &v
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO steps (run_id, step_id, prompt, output, model, duration_ms,
		 kind, exit_status, tokens_in, tokens_out, gate_passed)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.RunID, rec.StepID, rec.Prompt, rec.Output, rec.Model, rec.DurationMs,
		rec.Kind, rec.ExitStatus, rec.TokensIn, rec.TokensOut, gatePassed,
	)
	return err
}
```

- [ ] **Step 5: Run store tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ -v`
Expected: FAIL — existing tests use old `RecordRun(kind, name, input)` signature

- [ ] **Step 6: Update existing store tests to use new API**

In `internal/store/store_test.go`, update `TestRecordRun`:

```go
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
```

Update `TestRecordStep`:

```go
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
```

- [ ] **Step 7: Run all store tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/store/schema.go internal/store/store.go internal/store/store_test.go
git commit -m "feat(store): enrich runs/steps schema with workflow_file, repo, model, tokens, cost, gate_passed"
```

---

### Task 4: Update Callers of Store API

**Files:**
- Modify: `internal/gui/api_runs_test.go`
- Modify: `internal/gui/smoke_test.go`
- Grep for all `RecordRun(` and `RecordStep(` calls and update them

- [ ] **Step 1: Find all callers**

Run: `cd /Users/stokes/Projects/gl1tch && grep -rn 'RecordRun\|RecordStep\|FinishRun' --include='*.go' | grep -v store.go | grep -v store_test.go`

- [ ] **Step 2: Update each caller to use the new struct-based API**

For each caller found, update from positional args to struct literals. Common patterns:

Old: `s.RecordRun("workflow", "hello", "")`
New: `s.RecordRun(store.RunRecord{Kind: "workflow", Name: "hello", Input: ""})`

Old: `s.RecordStep(id, "step1", "prompt", "output", "gpt-4", 1500)`
New: `s.RecordStep(store.StepRecord{RunID: id, StepID: "step1", Prompt: "prompt", Output: "output", Model: "gpt-4", DurationMs: 1500})`

Old: `s.FinishRun(id, "done", 0)`
New: `s.FinishRun(id, "done", 0)` (unchanged — variadic totals)

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -30`
Expected: PASS across all packages

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor(store): update all callers to use RunRecord/StepRecord struct API"
```

---

### Task 5: GUI Workspace Endpoint

**Files:**
- Modify: `internal/gui/server.go`
- Create: `internal/gui/api_workspace.go`
- Create: `internal/gui/api_workspace_test.go`

- [ ] **Step 1: Write the failing test**

```go
package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetWorkspace(t *testing.T) {
	srv := testServer(t)
	src := `(workspace "test-ws" :description "a test" :owner "adam" (repos "elastic/ensemble") (defaults :model "qwen2.5:7b" :provider "ollama"))`
	os.WriteFile(filepath.Join(srv.workspace, "workspace.glitch"), []byte(src), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workspace", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != "test-ws" {
		t.Errorf("expected name test-ws, got %v", resp["name"])
	}
	if resp["owner"] != "adam" {
		t.Errorf("expected owner adam, got %v", resp["owner"])
	}
}

func TestGetWorkspace_Missing(t *testing.T) {
	srv := testServer(t)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workspace", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200 even when missing, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != nil && resp["name"] != "" {
		t.Errorf("expected empty name for missing workspace, got %v", resp["name"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/gui/ -v -run TestGetWorkspace`
Expected: FAIL — no route for /api/workspace

- [ ] **Step 3: Write the workspace handler**

Create `internal/gui/api_workspace.go`:

```go
package gui

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/workspace"
)

type workspaceResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Owner       string            `json:"owner"`
	Repos       []string          `json:"repos"`
	Defaults    workspaceDefaults `json:"defaults"`
}

type workspaceDefaults struct {
	Model    string `json:"model,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.workspace, "workspace.glitch")
	data, err := os.ReadFile(path)
	if err != nil {
		// No workspace.glitch — return empty response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaceResponse{Repos: []string{}})
		return
	}

	ws, err := workspace.ParseFile(data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaceResponse{
		Name:        ws.Name,
		Description: ws.Description,
		Owner:       ws.Owner,
		Repos:       ws.Repos,
		Defaults: workspaceDefaults{
			Model:    ws.Defaults.Model,
			Provider: ws.Defaults.Provider,
		},
	})
}
```

- [ ] **Step 4: Register the route in server.go**

In `internal/gui/server.go`, add to the `routes()` method after the existing API routes:

```go
	s.mux.HandleFunc("GET /api/workspace", s.handleGetWorkspace)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/gui/ -v -run TestGetWorkspace`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/gui/api_workspace.go internal/gui/api_workspace_test.go internal/gui/server.go
git commit -m "feat(gui): add GET /api/workspace endpoint for workspace manifest"
```

---

### Task 6: Enrich GUI Workflow and Run API Responses

**Files:**
- Modify: `internal/gui/api_workflows.go:18-22`
- Modify: `internal/gui/api_runs.go:9-18,76-80`
- Modify: `internal/gui/api_workflows_test.go`
- Modify: `internal/gui/api_runs_test.go`

- [ ] **Step 1: Write the failing test for enriched workflow list**

Add to `internal/gui/api_workflows_test.go`:

```go
func TestListWorkflows_WithMetadata(t *testing.T) {
	srv := testServer(t)
	wfDir := filepath.Join(srv.workspace, "workflows")
	src := `(workflow "pr-review" :description "Review PRs" :tags ("review" "ci") :author "adam" :version "1.0" :created "2026-04-01" (step "s1" (run "echo hi")))`
	os.WriteFile(filepath.Join(wfDir, "pr-review.glitch"), []byte(src), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var wfs []workflowEntry
	json.Unmarshal(w.Body.Bytes(), &wfs)
	if len(wfs) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(wfs))
	}
	if len(wfs[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(wfs[0].Tags))
	}
	if wfs[0].Author != "adam" {
		t.Errorf("expected author adam, got %q", wfs[0].Author)
	}
	if wfs[0].Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", wfs[0].Version)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/gui/ -v -run TestListWorkflows_WithMetadata`
Expected: FAIL — `workflowEntry` has no `Tags` field

- [ ] **Step 3: Enrich workflowEntry struct**

In `internal/gui/api_workflows.go`:

```go
type workflowEntry struct {
	Name        string   `json:"name"`
	File        string   `json:"file"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
	Created     string   `json:"created,omitempty"`
}
```

Update the workflow list builder in `handleListWorkflows` to populate the new fields:

```go
		workflows = append(workflows, workflowEntry{
			Name:        wf.Name,
			File:        e.Name(),
			Description: wf.Description,
			Tags:        wf.Tags,
			Author:      wf.Author,
			Version:     wf.Version,
			Created:     wf.Created,
		})
```

- [ ] **Step 4: Enrich runEntry and stepEntry structs**

In `internal/gui/api_runs.go`, update `runEntry`:

```go
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
```

Update the `handleListRuns` query:

```go
	rows, err := s.store.DB().Query(
		`SELECT id, kind, name, COALESCE(input,''), COALESCE(output,''),
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0),
		        COALESCE(workflow_file,''), COALESCE(repo,''), COALESCE(model,''),
		        COALESCE(tokens_in,0), COALESCE(tokens_out,0), COALESCE(cost_usd,0)
		 FROM runs ORDER BY id DESC LIMIT 100`,
	)
```

And the scan:

```go
		rows.Scan(&re.ID, &re.Kind, &re.Name, &re.Input, &re.Output,
			&re.ExitStatus, &re.StartedAt, &re.FinishedAt,
			&re.WorkflowFile, &re.Repo, &re.Model,
			&re.TokensIn, &re.TokensOut, &re.CostUSD)
```

Update the `handleGetRun` query similarly, and enrich `stepEntry`:

```go
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
```

Update the `handleGetRun` step query:

```go
	stepRows, _ := s.store.DB().Query(
		`SELECT step_id, COALESCE(model,''), COALESCE(duration_ms,0),
		        COALESCE(kind,''), exit_status, COALESCE(tokens_in,0),
		        COALESCE(tokens_out,0), gate_passed
		 FROM steps WHERE run_id = ?`, id)
```

And the step scan (using nullable intermediaries for `exit_status` and `gate_passed`):

```go
		for stepRows.Next() {
			var se stepEntry
			var exitStatus, gatePassed sql.NullInt64
			stepRows.Scan(&se.StepID, &se.Model, &se.DurationMs,
				&se.Kind, &exitStatus, &se.TokensIn, &se.TokensOut, &gatePassed)
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
```

Add the `database/sql` import to `api_runs.go`.

Update the `handleGetRun` run query to also select the enriched columns:

```go
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
```

- [ ] **Step 5: Run all GUI tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/gui/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/gui/api_workflows.go internal/gui/api_runs.go internal/gui/api_workflows_test.go internal/gui/api_runs_test.go
git commit -m "feat(gui): enrich workflow list with tags/author/version, runs with model/tokens/cost"
```

---

### Task 7: CLI `--tag` Filter on `workflow list`

**Files:**
- Modify: `cmd/workflow.go:17-58`

- [ ] **Step 1: Add `--tag` flag to `workflowListCmd`**

In `cmd/workflow.go`, add a flag variable and register it:

```go
var workflowTagFilter string
```

In `init()`, add:

```go
	workflowListCmd.Flags().StringVar(&workflowTagFilter, "tag", "", "filter workflows by tag")
```

- [ ] **Step 2: Update the list command to filter by tag**

In the `workflowListCmd` RunE, after building the sorted `names` slice, add filtering before the tabwriter loop:

```go
		if workflowTagFilter != "" {
			var filtered []string
			for _, name := range names {
				w := workflows[name]
				for _, tag := range w.Tags {
					if tag == workflowTagFilter {
						filtered = append(filtered, name)
						break
					}
				}
			}
			names = filtered
		}
```

- [ ] **Step 3: Verify manually**

Run: `cd /Users/stokes/Projects/gl1tch && go build -o /tmp/glitch . && /tmp/glitch workflow list --tag review`
Expected: builds and runs without error (may show no results if no workflows have that tag)

- [ ] **Step 4: Commit**

```bash
git add cmd/workflow.go
git commit -m "feat(cli): add --tag filter to workflow list command"
```

---

### Task 8: Full Integration Test

**Files:**
- No new files — validation only

- [ ] **Step 1: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -40`
Expected: PASS across all packages

- [ ] **Step 2: Run vet and build**

Run: `cd /Users/stokes/Projects/gl1tch && go vet ./... && go build ./...`
Expected: No errors

- [ ] **Step 3: Delete old SQLite database (pre-1.0 wipe)**

Run: `rm -f ~/.local/share/glitch/glitch.db`

This is required because the schema changed and there are no migrations pre-1.0.

- [ ] **Step 4: Commit (if any fixes were needed)**

Only if previous steps revealed issues that required fixes.

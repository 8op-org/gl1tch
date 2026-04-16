# Workspace Mechanics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Promote the gl1tch workspace from a `--workspace <path>` flag into a first-class environment with declared resources, nested runs, auto-discovery, and a reshaped CLI.

**Architecture:** Adds a declared resource model (git/local/tracker) materialized into `<workspace>/resources/`, nested runs via new `parent_run_id` + `workflow_name` SQLite columns, CWD walk-up + registry + active-workspace state discovery, a reshaped CLI that folds workflow management under `glitch workspace workflow …` and promotes `glitch run <name>` to top-level, and a GUI workspace switcher + resources panel + run tree. All config is s-expr; YAML → s-expr is a clean break per the pre-1.0 rule.

**Tech Stack:** Go + cobra (CLI), `internal/sexpr` (parser), SQLite via `modernc.org/sqlite`, SvelteKit-style SPA with `svelte-spa-router` frontend, Playwright e2e, Task runner (Taskfile.yml).

**Spec:** `docs/superpowers/specs/2026-04-16-workspace-mechanics-design.md`

**Baseline:** Worktree `.worktrees/workspace-mechanics` on branch `feature/workspace-mechanics` at `baa7ba5`; `go test ./...` all green; `gui/dist` seeded from main.

---

## Coding conventions (all tasks)

- Run `go build ./...` after every code change; do not commit on red.
- Every new Go file gets a `_test.go` sibling with at least one test per exported function.
- Pre-1.0 rule: **no migrations, no backwards-compat shims.** Wipe and re-create SQLite on schema change; rename config files with a one-time stderr warning; delete old code paths outright.
- Commit at the end of every task with a conventional-commit message (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`).
- Do not push. Do not open PRs. User pushes at the end.
- Do not add comments explaining WHAT code does — only WHY when non-obvious.
- Run the task-specific verification command before committing. Paste the output in the task completion report.

---

## File-structure map (decomposition decisions locked in here)

**New packages / files**

- `internal/resource/` — new package; git, symlink, tracker materialization + sync
  - `resource.go` — `Resource` struct, `Kind` enum, `Manager` interface
  - `git.go` — git clone/fetch, SHA resolution
  - `local.go` — symlink operations
  - `tracker.go` — GitHub repo alias (optional 404 probe)
  - `sync.go` — dispatch by kind, timestamp bookkeeping
- `internal/workspace/writer.go` — formatting-preserving writer that updates `:pin` values in `workspace.glitch` without regenerating from scratch
- `internal/workspace/registry/` — new sub-package; `~/.config/glitch/workspaces.glitch` and `state.glitch`
  - `registry.go` — load/save registry, add/remove entries
  - `state.go` — active-workspace pointer
- `internal/workspace/resources.go` — resource entries in `Workspace` struct + `.glitch/resources.glitch` state file
- `cmd/run.go` — top-level `glitch run` (extracted from `workflow run`)
- `cmd/workspace.go` — parent `workspace` cobra command + shared helpers
- `cmd/workspace_init.go`, `workspace_use.go`, `workspace_list.go`, `workspace_status.go`, `workspace_register.go`, `workspace_unregister.go`
- `cmd/workspace_add.go`, `workspace_rm.go`, `workspace_sync.go`, `workspace_pin.go`
- `cmd/workspace_workflow.go` — `workspace workflow list/new/edit`
- `cmd/workspace_gui.go` — replaces `cmd/gui.go` registration of the GUI command (the GUI package itself is unchanged)
- `internal/gui/api_workspaces.go` — registry + active + use endpoints
- `internal/gui/api_resources.go` — resources CRUD + sync + pin
- `internal/gui/api_runtree.go` — `GET /api/runs/:id/tree`, `GET /api/runs?parent_id=`
- `gui/src/lib/components/WorkspaceSwitcher.svelte`
- `gui/src/lib/components/ResourcesPanel.svelte`
- `gui/src/lib/components/RunTree.svelte`
- `gui/src/routes/Dashboard.svelte`

**Modified files**

- `internal/store/schema.go` — add columns, bump wipe detection
- `internal/store/store.go` — extend `RunRecord` with `ParentRunID`, `WorkflowName`; new `GetRunTree`, `ListChildren`, `SetParentRunID`
- `internal/workspace/workspace.go` + `serialize.go` — add `Resources []Resource`
- `internal/workspace/resolve.go` — new precedence chain (flag → env → walk-up → active → none)
- `cmd/root.go` — wire `GLITCH_WORKSPACE` env var + call new resolver
- `cmd/config.go` — switch file format from YAML to s-expr
- `internal/pipeline/runner.go` — add `resource` root to render data; thread through `runCtx`
- `internal/pipeline/sexpr.go` — register `call-workflow`, `map-resources` forms
- `internal/pipeline/types.go` — new `Step` fields for `call-workflow` + `map-resources`
- `internal/gui/server.go` — register new routes
- `gui/src/App.svelte` + `gui/src/lib/api.js` + `gui/src/routes/Settings.svelte` — frontend rewiring
- `internal/batch/batch.go` — populate `parent_run_id`, switch layout to `children/<variant>-<iter>-<runid>/`
- `cmd/workflow.go` — remove subcommands; keep struct only if tests depend on it, else delete file

**Deleted files**

- `cmd/gui.go` — replaced by `cmd/workspace_gui.go` (rename + move `workflow gui` → `workspace gui`)

**Unchanged — do NOT touch this session**

- `internal/tui*` (being ripped out in a parallel session, per memory rule)
- `internal/observer/` (except store-query call sites if they exist)
- `internal/research/repo.go` — `EnsureRepo` stays; new `internal/resource/git.go` is independent

---

## Phase 1 — Schema + Discovery Foundation

### Task 1: Add `parent_run_id` and `workflow_name` columns + wipe-on-drift

**Files:**
- Modify: `internal/store/schema.go`
- Modify: `internal/store/store.go:54-65` (drift detection)
- Modify: `internal/store/store.go:84-130` (RunRecord + RecordRun)
- Create: `internal/store/tree.go`
- Create: `internal/store/tree_test.go`
- Modify: `internal/store/store_test.go` (if exists)

- [ ] **Step 1: Write the failing test** — create `internal/store/tree_test.go`:

```go
package store

import (
	"path/filepath"
	"testing"
)

func TestParentRunID(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	parentID, err := s.RecordRun(RunRecord{Kind: "workflow", Name: "parent", WorkflowName: "parent-flow"})
	if err != nil {
		t.Fatal(err)
	}
	childID, err := s.RecordRun(RunRecord{Kind: "workflow", Name: "child", WorkflowName: "child-flow", ParentRunID: parentID})
	if err != nil {
		t.Fatal(err)
	}

	children, err := s.ListChildren(parentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 || children[0].ID != childID {
		t.Fatalf("expected one child with id %d, got %+v", childID, children)
	}
	if children[0].WorkflowName != "child-flow" {
		t.Errorf("expected workflow_name 'child-flow', got %q", children[0].WorkflowName)
	}
}

func TestGetRunTree(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, _ := s.RecordRun(RunRecord{Kind: "batch", Name: "p", WorkflowName: "parent"})
	c1, _ := s.RecordRun(RunRecord{Kind: "workflow", Name: "c1", WorkflowName: "child", ParentRunID: p})
	c2, _ := s.RecordRun(RunRecord{Kind: "workflow", Name: "c2", WorkflowName: "child", ParentRunID: p})
	_, _ = s.RecordRun(RunRecord{Kind: "workflow", Name: "gc", WorkflowName: "grandchild", ParentRunID: c1})

	tree, err := s.GetRunTree(p)
	if err != nil {
		t.Fatal(err)
	}
	if tree.ID != p {
		t.Fatalf("root id mismatch: %d vs %d", tree.ID, p)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(tree.Children))
	}
	// find c1, ensure it has one grandchild
	var c1Node *RunNode
	for i := range tree.Children {
		if tree.Children[i].ID == c1 {
			c1Node = &tree.Children[i]
		}
	}
	if c1Node == nil || len(c1Node.Children) != 1 {
		t.Fatalf("c1 should have exactly one grandchild: %+v", c1Node)
	}
	_ = c2
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./internal/store -run TestParentRunID -v
```

Expected: compile failure (`RunRecord` has no `ParentRunID` / `WorkflowName`; no `ListChildren` / `GetRunTree`).

- [ ] **Step 3: Extend the schema** — replace `internal/store/schema.go` with:

```go
package store

const createSchema = `
CREATE TABLE IF NOT EXISTS runs (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  kind            TEXT NOT NULL,
  name            TEXT NOT NULL,
  input           TEXT,
  output          TEXT,
  exit_status     INTEGER,
  started_at      INTEGER NOT NULL,
  finished_at     INTEGER,
  metadata        TEXT,
  workflow_file   TEXT,
  repo            TEXT,
  model           TEXT,
  tokens_in       INTEGER,
  tokens_out      INTEGER,
  cost_usd        REAL,
  variant         TEXT,
  workspace       TEXT NOT NULL DEFAULT '',
  parent_run_id   INTEGER REFERENCES runs(id),
  workflow_name   TEXT
);
CREATE INDEX IF NOT EXISTS idx_runs_parent ON runs(parent_run_id);

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

- [ ] **Step 4: Update drift detection** — in `internal/store/store.go` replace the `parent_run_id` probe column in the existing drift block. Change line 55 from:

```go
err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('runs') WHERE name = 'input'`).Scan(&count)
```

to:

```go
err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('runs') WHERE name = 'parent_run_id'`).Scan(&count)
```

This makes any pre-parent_run_id DB wipe on open (clean break per pre-1.0 rule).

- [ ] **Step 5: Extend RunRecord + RecordRun** — in `internal/store/store.go` modify `RunRecord` (lines 84-94) to add fields:

```go
type RunRecord struct {
	Kind         string
	Name         string
	Input        string
	WorkflowFile string
	Repo         string
	Model        string
	Variant      string
	Workspace    string
	ParentRunID  int64  // 0 means no parent
	WorkflowName string // logical workflow name (for tree rendering)
}
```

Then modify `RecordRun` (lines 119-130) to:

```go
func (s *Store) RecordRun(rec RunRecord) (int64, error) {
	var parent interface{}
	if rec.ParentRunID > 0 {
		parent = rec.ParentRunID
	}
	res, err := s.db.Exec(
		`INSERT INTO runs (kind, name, input, started_at, workflow_file, repo, model, variant, workspace, parent_run_id, workflow_name)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.Kind, rec.Name, rec.Input, time.Now().UnixMilli(),
		rec.WorkflowFile, rec.Repo, rec.Model, rec.Variant, rec.Workspace,
		parent, rec.WorkflowName,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
```

- [ ] **Step 6: Implement tree queries** — create `internal/store/tree.go`:

```go
package store

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
		var exit, fin *int64
		var exitInt *int
		if err := rows.Scan(&n.ID, &n.Name, &n.WorkflowName, &n.Kind, &exitInt, &n.StartedAt, &fin, &n.ParentRunID); err != nil {
			return nil, err
		}
		n.ExitStatus = exitInt
		n.FinishedAt = fin
		_ = exit
		out = append(out, n)
	}
	return out, rows.Err()
}

// GetRunTree returns the run rooted at id with all descendants populated.
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
	kids, err := s.ListChildren(id)
	if err != nil {
		return RunNode{}, err
	}
	for i := range kids {
		sub, err := s.GetRunTree(kids[i].ID)
		if err != nil {
			return RunNode{}, err
		}
		kids[i] = sub
	}
	n.Children = kids
	return n, nil
}
```

- [ ] **Step 7: Run tests** — verify pass:

```
go test ./internal/store/... -v
```

Expected: PASS.

- [ ] **Step 8: Verify full build is green**

```
go build ./...
go test ./internal/store ./internal/batch ./internal/pipeline ./cmd -count=1
```

Expected: all pass. Other packages reference `RunRecord`; the added fields are optional so nothing else should break.

- [ ] **Step 9: Commit**

```
git add internal/store/schema.go internal/store/store.go internal/store/tree.go internal/store/tree_test.go
git commit -m "feat(store): add parent_run_id and workflow_name columns + run tree queries"
```

---

### Task 2: Extend `Workspace` with `Resources []Resource`

**Files:**
- Modify: `internal/workspace/workspace.go`
- Modify: `internal/workspace/serialize.go`
- Create: `internal/workspace/resources.go`
- Modify: `internal/workspace/workspace_test.go`

- [ ] **Step 1: Write failing tests** — append to `internal/workspace/workspace_test.go`:

```go
func TestParseResources(t *testing.T) {
	src := []byte(`(workspace "demo"
  (resource "ensemble" :type "git" :url "https://github.com/elastic/ensemble" :ref "main" :pin "abc123")
  (resource "notes"    :type "local" :path "~/my-notes")
  (resource "obs-robots" :type "tracker" :repo "elastic/observability-robots"))`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Resources) != 3 {
		t.Fatalf("want 3 resources, got %d", len(w.Resources))
	}
	if w.Resources[0].Name != "ensemble" || w.Resources[0].Type != "git" || w.Resources[0].URL != "https://github.com/elastic/ensemble" || w.Resources[0].Ref != "main" || w.Resources[0].Pin != "abc123" {
		t.Errorf("git resource mismatch: %+v", w.Resources[0])
	}
	if w.Resources[1].Type != "local" || w.Resources[1].Path != "~/my-notes" {
		t.Errorf("local resource mismatch: %+v", w.Resources[1])
	}
	if w.Resources[2].Type != "tracker" || w.Resources[2].Repo != "elastic/observability-robots" {
		t.Errorf("tracker resource mismatch: %+v", w.Resources[2])
	}
}

func TestSerializeResourcesRoundTrip(t *testing.T) {
	in := &Workspace{
		Name: "demo",
		Resources: []Resource{
			{Name: "ensemble", Type: "git", URL: "https://github.com/elastic/ensemble", Ref: "main", Pin: "abc"},
			{Name: "notes", Type: "local", Path: "~/my-notes"},
			{Name: "obs", Type: "tracker", Repo: "elastic/observability-robots"},
		},
	}
	out := Serialize(in)
	got, err := ParseFile(out)
	if err != nil {
		t.Fatalf("reparse failed: %v\nserialized:\n%s", err, out)
	}
	if len(got.Resources) != 3 {
		t.Fatalf("round-trip resources: got %d, serialized:\n%s", len(got.Resources), out)
	}
	for i, r := range in.Resources {
		if got.Resources[i] != r {
			t.Errorf("resource %d mismatch:\nwant: %+v\n got: %+v", i, r, got.Resources[i])
		}
	}
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./internal/workspace -run TestParseResources -v
go test ./internal/workspace -run TestSerializeResources -v
```

Expected: compile failure (no `Resource` type).

- [ ] **Step 3: Create the Resource type** — `internal/workspace/resources.go`:

```go
package workspace

// Resource is a declared workspace resource (git repo, local folder, or tracker alias).
type Resource struct {
	Name string
	Type string // "git" | "local" | "tracker"
	URL  string // git only
	Ref  string // git only
	Pin  string // git only (tool-managed)
	Path string // local only
	Repo string // tracker only; "org/name"
}
```

- [ ] **Step 4: Add Resources to Workspace struct** — in `internal/workspace/workspace.go`, replace lines 9-16:

```go
type Workspace struct {
	Name        string
	Description string
	Owner       string
	Repos       []string
	Defaults    Defaults
	Resources   []Resource
}
```

- [ ] **Step 5: Parse `(resource …)` forms** — in `internal/workspace/workspace.go`, inside `convertWorkspace`, extend the list-form switch (around line 88) to handle `"resource"`:

```go
			case "repos":
				for _, repo := range child.Children[1:] {
					val := repo.StringVal()
					if val == "" {
						return nil, fmt.Errorf("line %d: repos entries must be strings", repo.Line)
					}
					w.Repos = append(w.Repos, val)
				}
			case "defaults":
				d, err := convertDefaults(child)
				if err != nil {
					return nil, err
				}
				w.Defaults = d
			case "resource":
				r, err := convertResource(child)
				if err != nil {
					return nil, err
				}
				w.Resources = append(w.Resources, r)
			default:
```

Add `convertResource` at the bottom of `workspace.go`:

```go
func convertResource(n *sexpr.Node) (Resource, error) {
	children := n.Children[1:] // skip "resource"
	if len(children) == 0 {
		return Resource{}, fmt.Errorf("line %d: resource missing name", n.Line)
	}
	r := Resource{Name: children[0].StringVal()}
	if r.Name == "" {
		return Resource{}, fmt.Errorf("line %d: resource name must be a string", children[0].Line)
	}
	children = children[1:]
	i := 0
	for i < len(children) {
		kw := children[i]
		if !(kw.IsAtom() && kw.Atom.Type == sexpr.TokenKeyword) {
			return Resource{}, fmt.Errorf("line %d: expected keyword in resource", kw.Line)
		}
		key := kw.KeywordVal()
		i++
		if i >= len(children) {
			return Resource{}, fmt.Errorf("line %d: keyword :%s missing value", kw.Line, key)
		}
		val := children[i].StringVal()
		switch key {
		case "type":
			r.Type = val
		case "url":
			r.URL = val
		case "ref":
			r.Ref = val
		case "pin":
			r.Pin = val
		case "path":
			r.Path = val
		case "repo":
			r.Repo = val
		default:
			return Resource{}, fmt.Errorf("line %d: unknown resource keyword :%s", kw.Line, key)
		}
		i++
	}
	if r.Type == "" {
		return Resource{}, fmt.Errorf("line %d: resource %q missing :type", n.Line, r.Name)
	}
	return r, nil
}
```

- [ ] **Step 6: Extend serialize** — in `internal/workspace/serialize.go`, append before the closing `b.WriteString(")\n")`:

```go
	for _, r := range w.Resources {
		b.WriteString(fmt.Sprintf("\n  (resource %q :type %q", r.Name, r.Type))
		if r.URL != "" {
			b.WriteString(fmt.Sprintf(" :url %q", r.URL))
		}
		if r.Ref != "" {
			b.WriteString(fmt.Sprintf(" :ref %q", r.Ref))
		}
		if r.Pin != "" {
			b.WriteString(fmt.Sprintf(" :pin %q", r.Pin))
		}
		if r.Path != "" {
			b.WriteString(fmt.Sprintf(" :path %q", r.Path))
		}
		if r.Repo != "" {
			b.WriteString(fmt.Sprintf(" :repo %q", r.Repo))
		}
		b.WriteString(")")
	}
```

- [ ] **Step 7: Run tests** — expect PASS:

```
go test ./internal/workspace -v -count=1
```

- [ ] **Step 8: Commit**

```
git add internal/workspace/resources.go internal/workspace/workspace.go internal/workspace/serialize.go internal/workspace/workspace_test.go
git commit -m "feat(workspace): parse and serialize (resource ...) forms in workspace.glitch"
```

---

### Task 3: Workspaces registry + active-workspace state

**Files:**
- Create: `internal/workspace/registry/registry.go`
- Create: `internal/workspace/registry/registry_test.go`
- Create: `internal/workspace/registry/state.go`
- Create: `internal/workspace/registry/state_test.go`

- [ ] **Step 1: Write failing tests** — `internal/workspace/registry/registry_test.go`:

```go
package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryAddLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir) // redirect ~/.config for testing
	t.Setenv("HOME", dir)

	if err := Add(Entry{Name: "alpha", Path: "/tmp/alpha"}); err != nil {
		t.Fatal(err)
	}
	if err := Add(Entry{Name: "beta", Path: "/tmp/beta"}); err != nil {
		t.Fatal(err)
	}
	if err := Add(Entry{Name: "alpha", Path: "/elsewhere"}); err == nil {
		t.Fatal("duplicate name should be rejected")
	}

	entries, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}

	raw, _ := os.ReadFile(filepath.Join(dir, ".config", "glitch", "workspaces.glitch"))
	if len(raw) == 0 {
		t.Fatal("registry file not written")
	}
}

func TestRegistryRemove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = Add(Entry{Name: "x", Path: "/a"})
	_ = Add(Entry{Name: "y", Path: "/b"})
	if err := Remove("x"); err != nil {
		t.Fatal(err)
	}
	entries, _ := List()
	if len(entries) != 1 || entries[0].Name != "y" {
		t.Fatalf("expected only y remaining, got %+v", entries)
	}
}
```

`internal/workspace/registry/state_test.go`:

```go
package registry

import "testing"

func TestActiveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if v, _ := GetActive(); v != "" {
		t.Fatalf("want empty, got %q", v)
	}
	if err := SetActive("stokagent"); err != nil {
		t.Fatal(err)
	}
	v, err := GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if v != "stokagent" {
		t.Fatalf("want stokagent, got %q", v)
	}
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./internal/workspace/registry/... -v
```

Expected: compile failure (package missing).

- [ ] **Step 3: Implement registry** — `internal/workspace/registry/registry.go`:

```go
package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

type Entry struct {
	Name string
	Path string
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "glitch"), nil
}

func registryPath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "workspaces.glitch"), nil
}

// List returns all registered workspaces, or an empty slice if the file is missing.
func List() ([]Entry, error) {
	p, err := registryPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "workspaces" {
			continue
		}
		for _, child := range n.Children[1:] {
			if !child.IsList() || len(child.Children) < 2 || child.Children[0].SymbolVal() != "workspace" {
				continue
			}
			e := Entry{Name: child.Children[1].StringVal()}
			kids := child.Children[2:]
			for i := 0; i+1 < len(kids); i += 2 {
				if kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword && kids[i].KeywordVal() == "path" {
					e.Path = kids[i+1].StringVal()
				}
			}
			out = append(out, e)
		}
	}
	return out, nil
}

// Add appends a new entry; returns error on name collision.
func Add(e Entry) error {
	if e.Name == "" || e.Path == "" {
		return fmt.Errorf("registry: name and path required")
	}
	cur, err := List()
	if err != nil {
		return err
	}
	for _, x := range cur {
		if x.Name == e.Name {
			return fmt.Errorf("registry: name %q already registered at %s", x.Name, x.Path)
		}
	}
	cur = append(cur, e)
	return Save(cur)
}

// Remove deletes the entry with the given name. Silent if not found.
func Remove(name string) error {
	cur, err := List()
	if err != nil {
		return err
	}
	out := cur[:0]
	for _, x := range cur {
		if x.Name != name {
			out = append(out, x)
		}
	}
	return Save(out)
}

// Save atomically rewrites the registry file from the given list.
func Save(entries []Entry) error {
	p, err := registryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("(workspaces")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("\n  (workspace %q :path %q)", e.Name, e.Path))
	}
	b.WriteString(")\n")
	return os.WriteFile(p, []byte(b.String()), 0o644)
}

// Find returns the entry for a given name, or an empty Entry + false if missing.
func Find(name string) (Entry, bool, error) {
	entries, err := List()
	if err != nil {
		return Entry{}, false, err
	}
	for _, e := range entries {
		if e.Name == name {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}
```

`internal/workspace/registry/state.go`:

```go
package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

func statePath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "state.glitch"), nil
}

func GetActive() (string, error) {
	p, err := statePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return "", err
	}
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "state" {
			continue
		}
		kids := n.Children[1:]
		for i := 0; i+1 < len(kids); i += 2 {
			if kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword && kids[i].KeywordVal() == "active-workspace" {
				return kids[i+1].StringVal(), nil
			}
		}
	}
	return "", nil
}

func SetActive(name string) error {
	p, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("(state :active-workspace %q)\n", name))
	_ = sexpr.Parse // keep import if linter trims; remove if unused
	return os.WriteFile(p, []byte(b.String()), 0o644)
}
```

If `sexpr` is unused in `state.go` drop the `_ = sexpr.Parse` line and remove the import.

- [ ] **Step 4: Run tests — expect PASS**

```
go test ./internal/workspace/registry/... -v -count=1
```

- [ ] **Step 5: Commit**

```
git add internal/workspace/registry
git commit -m "feat(workspace): add registry and active-workspace state files"
```

---

### Task 4: Resolve workspace with full precedence (flag → env → walk-up → active → none)

**Files:**
- Modify: `internal/workspace/resolve.go`
- Create: `internal/workspace/resolve_test.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Write failing test** — create `internal/workspace/resolve_test.go`:

```go
package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWorkspace(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "workspace.glitch"),
		[]byte("(workspace \""+name+"\")\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveEnvOverride(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "foo")
	_ = os.MkdirAll(ws, 0o755)
	writeWorkspace(t, ws, "foo-ws")

	got := Resolve(ResolveOpts{EnvPath: ws})
	if got.Name != "foo-ws" || got.Path != ws {
		t.Fatalf("env override failed: %+v", got)
	}
}

func TestResolveExplicitBeatsEnv(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	_ = os.MkdirAll(a, 0o755)
	_ = os.MkdirAll(b, 0o755)
	writeWorkspace(t, a, "ws-a")
	writeWorkspace(t, b, "ws-b")

	got := Resolve(ResolveOpts{ExplicitPath: a, EnvPath: b})
	if got.Name != "ws-a" {
		t.Fatalf("explicit should win: %+v", got)
	}
}

func TestResolveWalkUp(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(parent, "sub", "deep")
	_ = os.MkdirAll(child, 0o755)
	writeWorkspace(t, parent, "walked")

	got := Resolve(ResolveOpts{StartDir: child})
	if got.Name != "walked" || got.Path != parent {
		t.Fatalf("walk-up failed: %+v", got)
	}
}

func TestResolveNoneReturnsEmpty(t *testing.T) {
	got := Resolve(ResolveOpts{StartDir: t.TempDir()})
	if got.Name != "" || got.Path != "" {
		t.Fatalf("expected empty Resolved, got %+v", got)
	}
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./internal/workspace -run TestResolve -v
```

Expected: compile failure.

- [ ] **Step 3: Rewrite `resolve.go`**

```go
package workspace

import (
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

// Resolved is the fully-resolved workspace (name + path) or empty if global mode.
type Resolved struct {
	Name string
	Path string // absolute directory containing workspace.glitch, or empty for global mode
}

// ResolveOpts controls the precedence chain. First non-empty wins:
// ExplicitPath → EnvPath → walk-up from StartDir → ActiveName (looked up via registry).
type ResolveOpts struct {
	ExplicitPath string
	EnvPath      string
	StartDir     string
	ActiveName   string // usually populated from registry.GetActive() by the caller
}

// Resolve returns the effective workspace or an empty Resolved when in global mode.
func Resolve(opts ResolveOpts) Resolved {
	for _, p := range []string{opts.ExplicitPath, opts.EnvPath} {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if r, ok := loadAt(abs); ok {
			return r
		}
	}
	if opts.StartDir != "" {
		if r, ok := walkUp(opts.StartDir); ok {
			return r
		}
	}
	if opts.ActiveName != "" {
		if e, ok, _ := registry.Find(opts.ActiveName); ok {
			abs, err := filepath.Abs(expandHome(e.Path))
			if err == nil {
				if r, ok := loadAt(abs); ok {
					return r
				}
			}
		}
	}
	return Resolved{}
}

// ResolveWorkspace (legacy) returns just the name for callers that haven't migrated.
// Kept so existing code paths keep compiling while the refactor lands.
func ResolveWorkspace(startDir string) string {
	r := Resolve(ResolveOpts{StartDir: startDir})
	if r.Name != "" {
		return r.Name
	}
	return filepath.Base(startDir)
}

func loadAt(dir string) (Resolved, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "workspace.glitch"))
	if err != nil {
		return Resolved{}, false
	}
	ws, err := ParseFile(data)
	if err != nil || ws.Name == "" {
		return Resolved{}, false
	}
	return Resolved{Name: ws.Name, Path: dir}, true
}

func walkUp(start string) (Resolved, bool) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return Resolved{}, false
	}
	dir := abs
	for {
		if r, ok := loadAt(dir); ok {
			return r, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Resolved{}, false
		}
		dir = parent
	}
}

func expandHome(p string) string {
	if len(p) > 1 && p[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
```

- [ ] **Step 4: Run tests**

```
go test ./internal/workspace -run TestResolve -v
```

Expected: PASS.

- [ ] **Step 5: Wire up precedence in `cmd/root.go`** — replace `PersistentPreRunE` (lines 36-53):

```go
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		resolved := resolveWorkspaceForCommand()
		workspacePath = resolved.Path // back-compat alias used elsewhere
		if resolved.Path == "" {
			return nil
		}
		if err := ensureWorkspaceDir(resolved.Path); err != nil {
			return err
		}
		cfg, _ := loadConfig()
		wsFile := filepath.Join(resolved.Path, "workspace.glitch")
		if data, err := os.ReadFile(wsFile); err == nil {
			if ws, err := workspace.ParseFile(data); err == nil {
				ApplyWorkspace(ws, cfg)
			}
		}
		mergedConfig = cfg
		return nil
	},
```

And add the helper at the bottom of `cmd/root.go`:

```go
func resolveWorkspaceForCommand() workspace.Resolved {
	cwd, _ := os.Getwd()
	active, _ := registry.GetActive()
	return workspace.Resolve(workspace.ResolveOpts{
		ExplicitPath: workspacePath,
		EnvPath:      os.Getenv("GLITCH_WORKSPACE"),
		StartDir:     cwd,
		ActiveName:   active,
	})
}
```

Add the `registry` import at the top of `cmd/root.go`:

```go
"github.com/8op-org/gl1tch/internal/workspace/registry"
```

- [ ] **Step 6: Verify build + tests**

```
go build ./...
go test ./cmd ./internal/workspace ./internal/workspace/registry -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```
git add internal/workspace/resolve.go internal/workspace/resolve_test.go cmd/root.go
git commit -m "feat(workspace): full precedence chain (flag|env|walk-up|active) for discovery"
```

---

### Task 5: Flip global config from YAML to s-expr (`config.glitch`)

**Files:**
- Modify: `cmd/config.go`
- Modify: `internal/gui/api_workflows.go:207-242` (loadGUIConfig)
- Create: `cmd/config_test.go`

- [ ] **Step 1: Write failing tests** — create `cmd/config_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigSexprRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := filepath.Join(dir, ".config", "glitch")
	_ = os.MkdirAll(cfgDir, 0o755)

	src := `(config
  :default-model "qwen2.5:7b"
  :default-provider "ollama"
  :eval-threshold 5
  :workflows-dir "/tmp/flows")
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.glitch"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfigFrom(filepath.Join(cfgDir, "config.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultModel != "qwen2.5:7b" || cfg.DefaultProvider != "ollama" || cfg.EvalThreshold != 5 || cfg.WorkflowsDir != "/tmp/flows" {
		t.Fatalf("config parse mismatch: %+v", cfg)
	}
}

func TestLoadConfigMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := loadConfigFrom("/nonexistent/path/config.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultModel == "" {
		t.Fatal("expected default model")
	}
}

func TestSaveConfigWritesSexpr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.glitch")
	cfg := &Config{DefaultModel: "x", DefaultProvider: "y", EvalThreshold: 3, WorkflowsDir: "/tmp"}
	if err := saveConfigAt(cfg, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data)[0] != '(' {
		t.Fatalf("expected s-expr, got: %s", data)
	}
	cfg2, err := loadConfigFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.DefaultModel != "x" || cfg2.DefaultProvider != "y" || cfg2.EvalThreshold != 3 {
		t.Fatalf("round-trip mismatch: %+v", cfg2)
	}
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./cmd -run TestLoadConfig -v
```

Expected: fail or compile error (YAML parser, no `saveConfigAt` helper, etc.).

- [ ] **Step 3: Rewrite `cmd/config.go`** — fully replace the file. Use the existing `Config` struct shape but switch parsing to sexpr. Key points:

- `configPath()` returns `$HOME/.config/glitch/config.glitch`
- On first load, if `config.glitch` missing but `config.yaml` present, emit **one** stderr warning line (`glitch: ~/.config/glitch/config.yaml is no longer read — see https://gl1tch.dev/docs/config — continuing with defaults`) and return defaults. Do NOT parse the YAML.
- `loadConfigFrom(path)` reads + parses `(config :key val … (providers …) (tiers …))`.
- `saveConfigAt(cfg, path)` writes the s-expr.
- `saveConfig(cfg)` calls `saveConfigAt(cfg, configPath())`.
- Delete the `gopkg.in/yaml.v3` import; delete any remaining YAML marshal/unmarshal.

Minimum set of top-level keys to parse: `:default-model`, `:default-provider`, `:eval-threshold`, `:workflows-dir`, plus list forms `(providers (provider "name" :type "..." :base-url "..." :api-key-env "..." :default-model "..."))` and `(tiers (tier :providers ("a" "b") :model "..."))`.

Serialization should be deterministic: keys in the order above.

(Write the full file. Use `internal/sexpr` package for parsing; hand-build strings for writing — there is no formatting-preserving writer yet, that's Task 7.)

- [ ] **Step 4: Update `internal/gui/api_workflows.go:207-242`** — replace the YAML reading body of `loadGUIConfig()` with a call into `cmd.LoadConfigForGUI()` (add an exported wrapper in `cmd/config.go` that returns the subset the GUI needs). Alternatively inline the sexpr parser the same way; prefer the wrapper to keep one source of truth.

Add to `cmd/config.go`:

```go
// LoadConfigForGUI is the read-only accessor used by the embedded GUI server.
// Returns zero-valued config on any error.
func LoadConfigForGUI() *Config {
	cfg, _ := loadConfig()
	return cfg
}
```

Update `internal/gui/api_workflows.go`: drop YAML struct + parse, call `cmd.LoadConfigForGUI()` and copy needed fields into the local `guiConfig`.

- [ ] **Step 5: Remove the YAML dep if unreferenced**

```
go mod tidy
```

If `yaml.v3` is still used elsewhere, leave it. If not, `go.mod` drops it.

- [ ] **Step 6: Run full tests + build**

```
go build ./...
go test ./cmd ./internal/gui -count=1
```

- [ ] **Step 7: Commit**

```
git add cmd/config.go cmd/config_test.go internal/gui/api_workflows.go go.mod go.sum
git commit -m "feat(config): flip global config from YAML to s-expr (config.glitch)"
```

---

## Phase 2 — Resource Materialization

### Task 6: New `internal/resource` package (types + git + local + tracker + sync)

**Files:**
- Create: `internal/resource/resource.go`
- Create: `internal/resource/git.go`
- Create: `internal/resource/local.go`
- Create: `internal/resource/tracker.go`
- Create: `internal/resource/sync.go`
- Create: `internal/resource/sync_test.go`

- [ ] **Step 1: Write failing tests** — `internal/resource/sync_test.go`:

```go
package resource

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// helper — initialize a bare-ish upstream repo we can clone locally (no network).
func initUpstream(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	must := func(c *exec.Cmd) {
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	must(exec.Command("git", "init", "-q", "-b", "main"))
	must(exec.Command("git", "-c", "user.email=t@t", "-c", "user.name=t",
		"commit", "-q", "--allow-empty", "-m", "init"))
	return dir
}

func TestSyncGitResource(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	upstream := initUpstream(t)
	ws := t.TempDir()

	r := Resource{Name: "up", Kind: KindGit, URL: upstream, Ref: "main"}
	result, err := Sync(ws, r)
	if err != nil {
		t.Fatal(err)
	}
	if result.Pin == "" {
		t.Fatal("expected pin SHA populated")
	}
	if _, err := os.Stat(filepath.Join(ws, "resources", "up", ".git")); err != nil {
		t.Fatalf("clone not materialized: %v", err)
	}
}

func TestSyncLocalSymlink(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(t.TempDir(), "data")
	_ = os.MkdirAll(target, 0o755)

	r := Resource{Name: "data", Kind: KindLocal, Path: target}
	if _, err := Sync(ws, r); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ws, "resources", "data")
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("not a symlink: %v", err)
	}
	if got != target {
		t.Fatalf("symlink target mismatch: got %q want %q", got, target)
	}
}

func TestSyncTracker(t *testing.T) {
	ws := t.TempDir()
	r := Resource{Name: "bug", Kind: KindTracker, Repo: "org/repo"}
	res, err := Sync(ws, r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != KindTracker {
		t.Fatalf("expected tracker, got %s", res.Kind)
	}
	// No filesystem materialization for trackers
	if _, err := os.Stat(filepath.Join(ws, "resources", "bug")); !os.IsNotExist(err) {
		t.Fatalf("tracker should not materialize on disk: err=%v", err)
	}
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./internal/resource/... -v
```

Expected: package missing.

- [ ] **Step 3: Implement** — create the four files:

`internal/resource/resource.go`:

```go
package resource

type Kind string

const (
	KindGit     Kind = "git"
	KindLocal   Kind = "local"
	KindTracker Kind = "tracker"
)

// Resource is the input to Sync.
type Resource struct {
	Name string
	Kind Kind
	URL  string // git
	Ref  string // git
	Pin  string // git; ignored on input — written back on output
	Path string // local
	Repo string // tracker
}

// Result is Sync's output. Mirrors Resource with pin + timestamp filled.
type Result struct {
	Name       string
	Kind       Kind
	Pin        string
	Path       string // materialized path (clone root / symlink)
	Repo       string
	FetchedAt  int64 // unix seconds
}
```

`internal/resource/git.go`:

```go
package resource

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// materializeGit clones or refreshes a git resource under ws/resources/<name>
// and returns the resolved commit SHA.
func materializeGit(ws string, r Resource, force bool) (string, string, error) {
	dir := filepath.Join(ws, "resources", r.Name)
	_, err := os.Stat(filepath.Join(dir, ".git"))
	switch {
	case err == nil && force:
		_ = os.RemoveAll(dir)
		fallthrough
	case os.IsNotExist(err) || force:
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return "", "", err
		}
		out, err := exec.Command("git", "clone", r.URL, dir).CombinedOutput()
		if err != nil {
			return "", "", fmt.Errorf("git clone %s: %v: %s", r.URL, err, out)
		}
	default:
		if _, err := exec.Command("git", "-C", dir, "fetch", "--tags", "origin").CombinedOutput(); err != nil {
			return "", "", fmt.Errorf("git fetch: %v", err)
		}
	}
	if r.Ref != "" {
		if out, err := exec.Command("git", "-C", dir, "checkout", "-q", r.Ref).CombinedOutput(); err != nil {
			return "", "", fmt.Errorf("git checkout %s: %v: %s", r.Ref, err, out)
		}
	}
	shaOut, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", "", fmt.Errorf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(shaOut)), dir, nil
}
```

`internal/resource/local.go`:

```go
package resource

import (
	"os"
	"path/filepath"
)

// materializeLocal creates or verifies a symlink at ws/resources/<name>
// pointing to the expanded path.
func materializeLocal(ws string, r Resource) (string, error) {
	link := filepath.Join(ws, "resources", r.Name)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return "", err
	}
	target := expandHome(r.Path)
	if _, err := os.Lstat(link); err == nil {
		if cur, _ := os.Readlink(link); cur == target {
			return link, nil
		}
		_ = os.Remove(link)
	}
	if err := os.Symlink(target, link); err != nil {
		return "", err
	}
	return link, nil
}

func expandHome(p string) string {
	if len(p) > 1 && p[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
```

`internal/resource/tracker.go`:

```go
package resource

// trackers have no filesystem side effects today; the repo alias is all that's
// recorded in workspace.glitch. Kept as its own file for future 404 probes.
func materializeTracker(r Resource) string {
	return r.Repo
}
```

`internal/resource/sync.go`:

```go
package resource

import (
	"fmt"
	"time"
)

// SyncOpts controls sync behaviour.
type SyncOpts struct {
	Force bool // re-clone git resources from scratch
}

// Sync materializes (or refreshes) a resource into ws/resources/<name>
// and returns a Result with pin + timestamp populated.
func Sync(ws string, r Resource, opts ...SyncOpts) (Result, error) {
	var opt SyncOpts
	if len(opts) > 0 {
		opt = opts[0]
	}
	res := Result{Name: r.Name, Kind: r.Kind, FetchedAt: time.Now().Unix()}
	switch r.Kind {
	case KindGit:
		sha, path, err := materializeGit(ws, r, opt.Force)
		if err != nil {
			return Result{}, err
		}
		res.Pin = sha
		res.Path = path
	case KindLocal:
		path, err := materializeLocal(ws, r)
		if err != nil {
			return Result{}, err
		}
		res.Path = path
	case KindTracker:
		res.Repo = materializeTracker(r)
	default:
		return Result{}, fmt.Errorf("unknown resource kind %q", r.Kind)
	}
	return res, nil
}
```

- [ ] **Step 4: Run tests**

```
go test ./internal/resource/... -v -count=1
```

Expected: PASS (TestSyncGit skipped if git missing).

- [ ] **Step 5: Commit**

```
git add internal/resource/
git commit -m "feat(resource): new package for git/local/tracker materialization and sync"
```

---

### Task 7: Formatting-preserving writer for `workspace.glitch` pin updates

**Files:**
- Create: `internal/workspace/writer.go`
- Create: `internal/workspace/writer_test.go`

Hand-edited `workspace.glitch` comments + whitespace must survive `:pin` updates written by `sync`. The strategy: regex-scoped update inside the matching `(resource "<name>" …)` block. No sexpr round-trip.

- [ ] **Step 1: Write failing tests** — `internal/workspace/writer_test.go`:

```go
package workspace

import (
	"strings"
	"testing"
)

func TestUpdatePinPreservesComments(t *testing.T) {
	src := `(workspace "demo"
  ;; keep this comment
  (resource "ensemble" :type "git"
    :url "https://github.com/elastic/ensemble"
    :ref "main"
    :pin "old-sha") ; inline
  (resource "notes" :type "local" :path "~/my-notes"))
`
	out, err := UpdatePin([]byte(src), "ensemble", "new-sha")
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, ";; keep this comment") {
		t.Error("top comment lost")
	}
	if !strings.Contains(s, "; inline") {
		t.Error("inline comment lost")
	}
	if !strings.Contains(s, `:pin "new-sha"`) {
		t.Errorf("pin not updated:\n%s", s)
	}
	if strings.Contains(s, `"old-sha"`) {
		t.Error("old pin still present")
	}
}

func TestUpdatePinAddsMissingKey(t *testing.T) {
	src := `(workspace "demo"
  (resource "x" :type "git" :url "u" :ref "main"))
`
	out, err := UpdatePin([]byte(src), "x", "newpin")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `:pin "newpin"`) {
		t.Fatalf("pin not added:\n%s", out)
	}
}

func TestUpdatePinUnknownResource(t *testing.T) {
	src := `(workspace "demo")`
	if _, err := UpdatePin([]byte(src), "ghost", "x"); err == nil {
		t.Fatal("expected error for unknown resource")
	}
}
```

- [ ] **Step 2: Run tests to confirm failure**

```
go test ./internal/workspace -run TestUpdatePin -v
```

- [ ] **Step 3: Implement writer** — `internal/workspace/writer.go`:

```go
package workspace

import (
	"fmt"
	"regexp"
	"strings"
)

// UpdatePin rewrites a single :pin value for the named resource, preserving
// comments and whitespace elsewhere. If the :pin key is missing, it is added
// immediately before the closing `)` of the matching (resource ...) form.
func UpdatePin(src []byte, name, pin string) ([]byte, error) {
	s := string(src)
	block, err := findResourceBlock(s, name)
	if err != nil {
		return nil, err
	}

	pinRe := regexp.MustCompile(`:pin\s+"[^"]*"`)
	updated := pinRe.ReplaceAllStringFunc(block.content, func(match string) string {
		return fmt.Sprintf(`:pin %q`, pin)
	})
	if updated == block.content {
		// No :pin key — insert just before the closing paren of the block.
		idx := strings.LastIndex(updated, ")")
		if idx < 0 {
			return nil, fmt.Errorf("malformed resource block for %q", name)
		}
		updated = updated[:idx] + fmt.Sprintf(" :pin %q", pin) + updated[idx:]
	}
	return []byte(s[:block.start] + updated + s[block.end:]), nil
}

type resourceBlock struct {
	start, end int
	content    string
}

// findResourceBlock locates the (resource "<name>" ...) s-expression by
// scanning for the opening token, then counting parens to find the matching
// close.
func findResourceBlock(s, name string) (resourceBlock, error) {
	header := fmt.Sprintf(`(resource %q`, name)
	start := strings.Index(s, header)
	if start < 0 {
		return resourceBlock{}, fmt.Errorf("resource %q not found", name)
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '"':
			// Skip quoted string — advance to matching close, honouring \".
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' && i+1 < len(s) {
					i++
				}
				i++
			}
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end := i + 1
				return resourceBlock{start: start, end: end, content: s[start:end]}, nil
			}
		}
	}
	return resourceBlock{}, fmt.Errorf("unbalanced parens in resource %q", name)
}
```

- [ ] **Step 4: Run tests**

```
go test ./internal/workspace -run TestUpdatePin -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```
git add internal/workspace/writer.go internal/workspace/writer_test.go
git commit -m "feat(workspace): format-preserving writer for pin updates in workspace.glitch"
```

---

### Task 8: `glitch workspace add` and `glitch workspace rm` commands

**Files:**
- Create: `cmd/workspace.go` (parent cobra group)
- Create: `cmd/workspace_add.go`
- Create: `cmd/workspace_rm.go`
- Create: `cmd/workspace_add_test.go`

- [ ] **Step 1: Scaffold parent group** — `cmd/workspace.go`:

```go
package cmd

import "github.com/spf13/cobra"

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "manage workspaces and their resources",
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}
```

- [ ] **Step 2: Write failing test** — `cmd/workspace_add_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace"
)

func TestWorkspaceAddLocal(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(`(workspace "demo")`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	if err := runWorkspaceAdd(ws, target, "notes", "", ""); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	w, err := workspace.ParseFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Resources) != 1 || w.Resources[0].Name != "notes" || w.Resources[0].Type != "local" {
		t.Fatalf("resource not recorded: %+v", w.Resources)
	}
	link := filepath.Join(ws, "resources", "notes")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
}

func TestWorkspaceRm(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "workspace.glitch"),
		[]byte(`(workspace "demo" (resource "notes" :type "local" :path "/tmp"))`+"\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(ws, "resources"), 0o755)
	_ = os.Symlink("/tmp", filepath.Join(ws, "resources", "notes"))

	if err := runWorkspaceRm(ws, "notes"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	w, _ := workspace.ParseFile(data)
	if len(w.Resources) != 0 {
		t.Fatalf("resource not removed: %+v", w.Resources)
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "notes")); !os.IsNotExist(err) {
		t.Fatalf("symlink not removed: %v", err)
	}
}
```

- [ ] **Step 3: Implement** — `cmd/workspace_add.go`:

```go
package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/resource"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var (
	addAs  string
	addPin string
)

var workspaceAddCmd = &cobra.Command{
	Use:   "add <url|path>",
	Short: "add a resource (git url, local path, or org/name tracker)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		name := addAs
		if name == "" {
			name = inferResourceName(args[0])
		}
		return runWorkspaceAdd(ws, args[0], name, addPin, "")
	},
}

func init() {
	workspaceAddCmd.Flags().StringVar(&addAs, "as", "", "resource name (defaults to inferred)")
	workspaceAddCmd.Flags().StringVar(&addPin, "pin", "", "git ref to pin (git resources only)")
	workspaceCmd.AddCommand(workspaceAddCmd)
}

func runWorkspaceAdd(ws, input, name, pin, typeOverride string) error {
	kind := typeOverride
	if kind == "" {
		kind = inferKind(input)
	}
	r := resource.Resource{Name: name, Kind: resource.Kind(kind)}
	switch kind {
	case "git":
		r.URL = input
		if pin != "" {
			r.Ref = pin
		} else {
			r.Ref = "main"
		}
	case "local":
		r.Path = input
	case "tracker":
		r.Repo = input
	default:
		return fmt.Errorf("could not infer resource kind from %q", input)
	}
	res, err := resource.Sync(ws, r)
	if err != nil {
		return err
	}

	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	// Reject duplicates
	for _, existing := range w.Resources {
		if existing.Name == name {
			return fmt.Errorf("resource %q already exists", name)
		}
	}
	w.Resources = append(w.Resources, workspace.Resource{
		Name: name, Type: kind,
		URL: r.URL, Ref: r.Ref, Pin: res.Pin,
		Path: r.Path, Repo: r.Repo,
	})
	if err := os.WriteFile(wsFile, workspace.Serialize(w), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmdStderr(), "added resource %q (%s)\n", name, kind)
	return nil
}

func inferKind(input string) string {
	switch {
	case strings.HasPrefix(input, "http://"),
		strings.HasPrefix(input, "https://"),
		strings.HasPrefix(input, "git@"),
		strings.HasSuffix(input, ".git"):
		return "git"
	case strings.HasPrefix(input, "~"),
		strings.HasPrefix(input, "/"),
		strings.HasPrefix(input, "."):
		return "local"
	case strings.Contains(input, "/") && !strings.Contains(input, ":"):
		return "tracker"
	}
	return ""
}

func inferResourceName(input string) string {
	// git URL → basename without .git
	if u, err := url.Parse(input); err == nil && u.Host != "" {
		name := path.Base(u.Path)
		return strings.TrimSuffix(name, ".git")
	}
	// local path → basename of expanded path
	if strings.HasPrefix(input, "~") || strings.HasPrefix(input, "/") || strings.HasPrefix(input, ".") {
		return filepath.Base(input)
	}
	// tracker "org/repo"
	if i := strings.Index(input, "/"); i >= 0 {
		return input[i+1:]
	}
	return input
}
```

`cmd/workspace_rm.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace"
)

var workspaceRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "remove a workspace resource and its files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspaceRm(ws, args[0])
	},
}

func init() { workspaceCmd.AddCommand(workspaceRmCmd) }

func runWorkspaceRm(ws, name string) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	found := false
	out := w.Resources[:0]
	for _, r := range w.Resources {
		if r.Name == name {
			found = true
			continue
		}
		out = append(out, r)
	}
	if !found {
		return fmt.Errorf("resource %q not found", name)
	}
	w.Resources = out
	if err := os.WriteFile(wsFile, workspace.Serialize(w), 0o644); err != nil {
		return err
	}
	_ = os.RemoveAll(filepath.Join(ws, "resources", name))
	return nil
}
```

Add `cmd/workspace.go` helpers:

```go
// activeWorkspacePath returns the resolved workspace path or error in global mode.
func activeWorkspacePath() (string, error) {
	r := resolveWorkspaceForCommand()
	if r.Path == "" {
		return "", fmt.Errorf("no active workspace — cd into one or run `glitch workspace use <name>`")
	}
	return r.Path, nil
}

func cmdStderr() *os.File { return os.Stderr }
```

(Add `"fmt"` and `"os"` imports to `cmd/workspace.go`.)

- [ ] **Step 4: Run tests + build**

```
go build ./...
go test ./cmd -run TestWorkspace -v
```

- [ ] **Step 5: Commit**

```
git add cmd/workspace.go cmd/workspace_add.go cmd/workspace_rm.go cmd/workspace_add_test.go
git commit -m "feat(cmd): glitch workspace add and rm for resource management"
```

---

### Task 9: `glitch workspace sync` + `pin` + `.glitch/resources.glitch` state

**Files:**
- Create: `cmd/workspace_sync.go`
- Create: `cmd/workspace_pin.go`
- Create: `internal/workspace/resources_state.go`
- Create: `internal/workspace/resources_state_test.go`
- Create: `cmd/workspace_sync_test.go`

- [ ] **Step 1: Write failing test for state file** — `internal/workspace/resources_state_test.go`:

```go
package workspace

import (
	"path/filepath"
	"testing"
	"time"
)

func TestResourceStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	st := ResourceState{
		Entries: map[string]time.Time{
			"ensemble": time.Unix(1700000000, 0).UTC(),
			"kibana":   time.Unix(1700000500, 0).UTC(),
		},
	}
	if err := SaveResourceState(dir, st); err != nil {
		t.Fatal(err)
	}
	got, err := LoadResourceState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 2 || !got.Entries["ensemble"].Equal(st.Entries["ensemble"]) {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	_ = filepath.Base(dir) // silence unused
}
```

- [ ] **Step 2: Implement state file** — `internal/workspace/resources_state.go`:

```go
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

type ResourceState struct {
	Entries map[string]time.Time
}

func resourceStatePath(ws string) string {
	return filepath.Join(ws, ".glitch", "resources.glitch")
}

func LoadResourceState(ws string) (ResourceState, error) {
	data, err := os.ReadFile(resourceStatePath(ws))
	if err != nil {
		if os.IsNotExist(err) {
			return ResourceState{Entries: map[string]time.Time{}}, nil
		}
		return ResourceState{}, err
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return ResourceState{}, err
	}
	out := ResourceState{Entries: map[string]time.Time{}}
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "resources" {
			continue
		}
		for _, c := range n.Children[1:] {
			if !c.IsList() || len(c.Children) < 2 || c.Children[0].SymbolVal() != "resource-state" {
				continue
			}
			name := c.Children[1].StringVal()
			kids := c.Children[2:]
			for i := 0; i+1 < len(kids); i += 2 {
				if kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword && kids[i].KeywordVal() == "fetched" {
					if t, err := time.Parse(time.RFC3339, kids[i+1].StringVal()); err == nil {
						out.Entries[name] = t
					}
				}
			}
		}
	}
	return out, nil
}

func SaveResourceState(ws string, st ResourceState) error {
	path := resourceStatePath(ws)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("(resources")
	names := make([]string, 0, len(st.Entries))
	for n := range st.Entries {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		b.WriteString(fmt.Sprintf("\n  (resource-state %q :fetched %q)", n, st.Entries[n].UTC().Format(time.RFC3339)))
	}
	b.WriteString(")\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
```

- [ ] **Step 3: Write failing test for sync** — `cmd/workspace_sync_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace"
)

func TestWorkspaceSyncLocalOnly(t *testing.T) {
	ws := t.TempDir()
	target := t.TempDir()
	src := `(workspace "demo" (resource "notes" :type "local" :path "` + target + `"))` + "\n"
	_ = os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(src), 0o644)

	if err := runWorkspaceSync(ws, nil, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "notes")); err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	st, err := workspace.LoadResourceState(ws)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.Entries["notes"]; !ok {
		t.Fatal("state not recorded")
	}
}
```

- [ ] **Step 4: Implement sync + pin commands** — `cmd/workspace_sync.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/resource"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var syncForce bool

var workspaceSyncCmd = &cobra.Command{
	Use:   "sync [name...]",
	Short: "update resources to their declared refs (all if no names given)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspaceSync(ws, args, syncForce)
	},
}

func init() {
	workspaceSyncCmd.Flags().BoolVar(&syncForce, "force", false, "re-clone git resources from scratch")
	workspaceCmd.AddCommand(workspaceSyncCmd)
}

func runWorkspaceSync(ws string, names []string, force bool) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}

	st, _ := workspace.LoadResourceState(ws)
	if st.Entries == nil {
		st.Entries = map[string]time.Time{}
	}

	// Track which workspace.glitch files need :pin rewrites.
	for _, r := range w.Resources {
		if len(want) > 0 && !want[r.Name] {
			continue
		}
		res, err := resource.Sync(ws, resource.Resource{
			Name: r.Name, Kind: resource.Kind(r.Type),
			URL: r.URL, Ref: r.Ref, Path: r.Path, Repo: r.Repo,
		}, resource.SyncOpts{Force: force})
		if err != nil {
			fmt.Fprintf(os.Stderr, "sync %s: %v\n", r.Name, err)
			continue
		}
		if res.Pin != "" && res.Pin != r.Pin {
			newSrc, err := workspace.UpdatePin(data, r.Name, res.Pin)
			if err == nil {
				data = newSrc
				_ = os.WriteFile(wsFile, data, 0o644)
			}
		}
		st.Entries[r.Name] = time.Unix(res.FetchedAt, 0).UTC()
	}
	return workspace.SaveResourceState(ws, st)
}
```

`cmd/workspace_pin.go`:

```go
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace"
)

var workspacePinCmd = &cobra.Command{
	Use:   "pin <name> <ref>",
	Short: "update a resource's :ref, sync, and write the resolved :pin",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspacePin(ws, args[0], args[1])
	},
}

func init() { workspaceCmd.AddCommand(workspacePinCmd) }

func runWorkspacePin(ws, name, ref string) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	for i := range w.Resources {
		if w.Resources[i].Name == name {
			w.Resources[i].Ref = ref
			break
		}
	}
	if err := os.WriteFile(wsFile, workspace.Serialize(w), 0o644); err != nil {
		return err
	}
	return runWorkspaceSync(ws, []string{name}, false)
}
```

- [ ] **Step 5: Run tests + build**

```
go build ./...
go test ./cmd ./internal/workspace -run 'TestWorkspaceSync|TestResourceState' -v -count=1
```

- [ ] **Step 6: Commit**

```
git add cmd/workspace_sync.go cmd/workspace_sync_test.go cmd/workspace_pin.go internal/workspace/resources_state.go internal/workspace/resources_state_test.go
git commit -m "feat(cmd): workspace sync and pin with persisted fetched timestamps"
```

---

### Task 10: `.resource.<name>.<field>` template binding in pipeline

**Files:**
- Modify: `internal/pipeline/runner.go` (render data builders + RunOpts + runCtx)
- Create: `internal/pipeline/resource_test.go`

- [ ] **Step 1: Write failing test** — `internal/pipeline/resource_test.go`:

```go
package pipeline

import "testing"

func TestRenderResourceBinding(t *testing.T) {
	data := map[string]any{
		"input": "",
		"param": map[string]string{},
		"resource": map[string]map[string]string{
			"ensemble": {"path": "/tmp/ensemble", "url": "https://x", "ref": "main", "pin": "sha123"},
		},
	}
	out, err := render("{{.resource.ensemble.path}}:{{.resource.ensemble.pin}}", data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "/tmp/ensemble:sha123" {
		t.Fatalf("bad render: %q", out)
	}
}

func TestRenderResourceMissingEmpty(t *testing.T) {
	data := map[string]any{"input": "", "param": map[string]string{}, "resource": map[string]map[string]string{}}
	out, err := render("x:{{.resource.missing.path}}:y", data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "x::y" {
		t.Fatalf("expected empty substitution, got %q", out)
	}
}
```

- [ ] **Step 2: Run test, confirm failure**

```
go test ./internal/pipeline -run TestRenderResource -v
```

Expected: likely fails because Go's `text/template` errors on nil-map field access. The test enforces graceful empty.

- [ ] **Step 3: Implement resource binding in `render`** — in `internal/pipeline/runner.go`, locate `func render(...)` (line 706). The change is threefold:

(a) Update `RunOpts` (around line 92) — add:

```go
	Resources map[string]map[string]string // resource name → field → value
```

(b) Update `runCtx` (around line 43) — add:

```go
	resources map[string]map[string]string
```

and copy from opts when building rctx.

(c) At every `data := map[string]any{...}` site (found via grep: lines ~934, ~963, ~1381, and inside `render` itself if it constructs defaults), add:

```go
		"resource": rctx.resources,
```

(d) In the template `render` function, because Go's `text/template` returns `<no value>` (or error in `missingkey=error` mode) for unknown map keys, add an `Option("missingkey=zero")` on the template and handle `nil` resource maps gracefully:

Find the `template.New(…).Funcs(…)` construction in `render` and change to:

```go
t, err := template.New("t").Funcs(funcMap).Option("missingkey=zero").Parse(tmpl)
```

If the data map passes `nil` for resource, substitute an empty `map[string]map[string]string{}` in `render` before executing:

```go
if _, ok := data["resource"]; !ok {
    data["resource"] = map[string]map[string]string{}
}
```

- [ ] **Step 4: Wire up the resources map population** — in `internal/pipeline/runner.go` `Run(...)` function, after the workspace resolution block, build the `resources` map from a new helper (not committed yet — implement inline):

```go
// build resources map — callers populate RunOpts.Resources directly.
rctx.resources = opts.Resources
if rctx.resources == nil {
    rctx.resources = map[string]map[string]string{}
}
```

- [ ] **Step 5: Populate `opts.Resources` from callers** — in `cmd/workflow.go` (or wherever `pipeline.Run` is invoked from CLI) and in `cmd/ask.go`, `cmd/batch` / `internal/batch/batch.go`, build the map from the active workspace's resources:

Add helper `cmd/workspace_resources.go`:

```go
package cmd

import (
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/workspace"
)

// ResourceBindings returns the map used to populate RunOpts.Resources.
func ResourceBindings(ws *workspace.Workspace, wsPath string) map[string]map[string]string {
	out := map[string]map[string]string{}
	if ws == nil {
		return out
	}
	for _, r := range ws.Resources {
		m := map[string]string{"url": r.URL, "ref": r.Ref, "pin": r.Pin, "repo": r.Repo}
		switch r.Type {
		case "git":
			m["path"] = filepath.Join(wsPath, "resources", r.Name)
			if m["repo"] == "" {
				m["repo"] = inferRepoFromURL(r.URL)
			}
		case "local":
			m["path"] = filepath.Join(wsPath, "resources", r.Name)
		case "tracker":
			// no path
		}
		out[r.Name] = m
	}
	return out
}

func inferRepoFromURL(url string) string {
	// https://github.com/org/name(.git)?
	if !strings.Contains(url, "github.com") {
		return ""
	}
	after := url[strings.Index(url, "github.com")+len("github.com"):]
	after = strings.TrimPrefix(after, "/")
	after = strings.TrimPrefix(after, ":")
	after = strings.TrimSuffix(after, ".git")
	return after
}
```

Callers that invoke `pipeline.Run` should set `opts.Resources = cmd.ResourceBindings(ws, wsPath)`. (At this point only `cmd/workflow.go` / `cmd/ask.go` / `internal/batch/batch.go` need updating; search for `pipeline.Run(` and patch.)

- [ ] **Step 6: Build + test**

```
go build ./...
go test ./internal/pipeline ./cmd -count=1
```

- [ ] **Step 7: Commit**

```
git add internal/pipeline/runner.go internal/pipeline/resource_test.go cmd/workspace_resources.go cmd/workflow.go cmd/ask.go internal/batch/batch.go
git commit -m "feat(pipeline): .resource.<name>.<field> template binding"
```

---

## Phase 3 — Workflow Composition

### Task 11: `call-workflow` sexpr form parser + runner + cycle guard

**Files:**
- Modify: `internal/pipeline/types.go` (new Step field)
- Modify: `internal/pipeline/sexpr.go` (parser registration)
- Modify: `internal/pipeline/runner.go` (executor + stack)
- Create: `internal/pipeline/callworkflow_test.go`

- [ ] **Step 1: Write failing test** — `internal/pipeline/callworkflow_test.go`:

```go
package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCallWorkflowChildRun(t *testing.T) {
	dir := t.TempDir()
	// parent workflow that calls child
	parent := `(workflow "parent" (step "out" (call-workflow "child" :input "hi")))`
	child := `(workflow "child" (step "echo" (run "echo got:{{.input}}")))`
	_ = os.WriteFile(filepath.Join(dir, "parent.glitch"), []byte(parent), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "child.glitch"), []byte(child), 0o644)

	w, err := ParseSexprWorkflowFromFile(filepath.Join(dir, "parent.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(w, "hi", "", map[string]string{}, nil,
		RunOpts{WorkflowsDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Steps["out"] != "got:hi" {
		t.Fatalf("unexpected output: %q", res.Steps["out"])
	}
}

func TestCallWorkflowCycleRejected(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.glitch"),
		[]byte(`(workflow "a" (step "x" (call-workflow "b")))`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.glitch"),
		[]byte(`(workflow "b" (step "x" (call-workflow "a")))`), 0o644)
	w, err := ParseSexprWorkflowFromFile(filepath.Join(dir, "a.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Run(w, "", "", nil, nil, RunOpts{WorkflowsDir: dir})
	if err == nil || !contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

(If `ParseSexprWorkflowFromFile` doesn't exist yet, export `parseSexprWorkflow` from a file path — add a thin wrapper in `sexpr.go`.)

- [ ] **Step 2: Run tests, confirm failure**

```
go test ./internal/pipeline -run TestCallWorkflow -v
```

- [ ] **Step 3: Add Step fields** — in `internal/pipeline/types.go`, inside `Step`:

```go
	// call-workflow: invokes a sibling workflow by name as a nested run.
	CallWorkflow  string            // workflow name
	CallInput     string            // template-rendered input for child
	CallSet       map[string]string // :set key=value params
```

- [ ] **Step 4: Register parser** — in `internal/pipeline/sexpr.go`, inside `convertStep` (wherever the step-body switch matches on head), add:

```go
	case "call-workflow":
		if len(children) < 1 {
			return Step{}, fmt.Errorf("call-workflow needs workflow name")
		}
		s.Form = "call-workflow"
		s.CallWorkflow = children[0].StringVal()
		// trailing keyword pairs
		kids := children[1:]
		s.CallSet = map[string]string{}
		for i := 0; i+1 < len(kids); i += 2 {
			if !(kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword) {
				continue
			}
			key := kids[i].KeywordVal()
			val := kids[i+1].StringVal()
			switch key {
			case "input":
				s.CallInput = val
			case "set":
				// allow :set key "val" multiple occurrences, or (set k v k v)
				// simple form here: one k/v pair per :set
				// caller uses multiple :set k v repeats, parser accepts alternating
				// We interpret :set "key=value" for simplicity in MVP.
				if i+1 < len(kids) {
					if kv := splitKV(val); kv != nil {
						for k, v := range kv {
							s.CallSet[k] = v
						}
					}
				}
			}
		}
		return s, nil
```

Add helper at bottom of `sexpr.go`:

```go
func splitKV(s string) map[string]string {
	if s == "" {
		return nil
	}
	if i := strings.Index(s, "="); i > 0 {
		return map[string]string{s[:i]: s[i+1:]}
	}
	return nil
}
```

- [ ] **Step 5: Add `ParseSexprWorkflowFromFile` wrapper + `WorkflowsDir` RunOpts**

In `internal/pipeline/sexpr.go`:

```go
// ParseSexprWorkflowFromFile loads and parses a workflow file.
func ParseSexprWorkflowFromFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseSexprWorkflow(data)
}
```

In `internal/pipeline/runner.go`, extend `RunOpts`:

```go
	WorkflowsDir   string   // directory to resolve call-workflow targets
	ParentRunID    int64    // if non-zero, this is a nested run
	CallStack      []string // workflow names already on stack (cycle guard)
```

And pass them into `runCtx`:

```go
	workflowsDir string
	parentRunID  int64
	callStack    []string
```

- [ ] **Step 6: Implement executor** — in `internal/pipeline/runner.go`, add `case "call-workflow":` to `executeStep` dispatch, plus the implementation:

```go
func executeCallWorkflow(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	// cycle guard
	for _, n := range rctx.callStack {
		if n == step.CallWorkflow {
			return nil, fmt.Errorf("call-workflow cycle detected: %s", strings.Join(append(rctx.callStack, step.CallWorkflow), " → "))
		}
	}
	if rctx.workflowsDir == "" {
		return nil, fmt.Errorf("call-workflow requires WorkflowsDir in RunOpts")
	}
	path := filepath.Join(rctx.workflowsDir, step.CallWorkflow+".glitch")
	child, err := ParseSexprWorkflowFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("call-workflow %s: %w", step.CallWorkflow, err)
	}
	rendered, err := render(step.CallInput, map[string]any{
		"input": rctx.input, "param": rctx.params, "workspace": rctx.workspace,
		"resource": rctx.resources,
	}, rctx.stepsSnapshot())
	if err != nil {
		return nil, err
	}
	childParams := map[string]string{}
	for k, v := range step.CallSet {
		childParams[k] = v
	}
	res, err := Run(child, rendered, rctx.defaultModel, childParams, rctx.reg, RunOpts{
		Workspace:     rctx.workspace,
		Resources:     rctx.resources,
		WorkflowsDir:  rctx.workflowsDir,
		CallStack:     append(rctx.callStack, step.CallWorkflow),
		ParentRunID:   rctx.parentRunID, // child inherits grandparent chain; store layer handles real parenting via a separate path
	})
	if err != nil {
		return nil, err
	}
	return &stepOutcome{output: res.Output}, nil
}
```

And dispatch in `executeStep`:

```go
	case "call-workflow":
		return executeCallWorkflow(ctx, rctx, step)
```

For parent-child in the **store**: extend `Run(...)` to record `parent_run_id`. The natural place is where `store.RecordRun` is called — pass `ParentRunID: opts.ParentRunID` and `WorkflowName: w.Name`. Find the existing `RecordRun` call site (grep for `s.store.RecordRun` or `RecordRun(`) and extend the record.

- [ ] **Step 7: Run tests**

```
go build ./...
go test ./internal/pipeline -run TestCallWorkflow -v -count=1
```

- [ ] **Step 8: Commit**

```
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/callworkflow_test.go
git commit -m "feat(pipeline): call-workflow sexpr form with cycle guard and nested runs"
```

---

### Task 12: Batch populates `parent_run_id` and uses `children/<variant>-<iter>-<runid>/` layout

**Files:**
- Modify: `internal/batch/batch.go`
- Modify: `internal/batch/batch_test.go` (if exists — add new case)

- [ ] **Step 1: Write failing test** — append to `internal/batch/batch_test.go` (or create):

```go
func TestBatchCreatesChildRuns(t *testing.T) {
	// Run a simple 2-variant batch and assert:
	// - one parent run row (kind=batch)
	// - each variant child has parent_run_id = parent.id
	// - on-disk layout is children/<variant>-<iter>-<runid>/
	// (expand with actual harness; see existing batch tests for setup)
	t.Skip("to be filled out against existing batch test harness")
}
```

- [ ] **Step 2: Modify `internal/batch/batch.go`** — two edits:

(a) Near the top of `Run()`, before the iteration loop, record a parent batch run:

```go
parentID, err := opts.Store.RecordRun(store.RunRecord{
    Kind:         "batch",
    Name:         opts.Name, // or derive from config
    WorkflowName: "batch",
    Workspace:    opts.Workspace,
})
if err != nil {
    return err
}
defer opts.Store.FinishRun(parentID, "", 0)
```

(b) When invoking `pipeline.Run` for each variant, pass `RunOpts{ParentRunID: parentID}`.

(c) Replace the `resultPath` function (lines 185-191). New layout:

```go
func resultPath(baseDir string, variant string, iter int, runID int64) string {
	return filepath.Join(baseDir, "children", fmt.Sprintf("%s-%d-%d", variant, iter, runID))
}
```

Callers update: capture the per-run child ID (from pipeline.Run's returned `Result.RunID`; if not returned today, extend Result to include it) and pass into `resultPath`.

Add to `internal/pipeline/runner.go`:

```go
type Result struct {
	Workflow string
	Output   string
	Steps    map[string]string
	RunID    int64 // new — matches the runs table PK
}
```

And populate `RunID: parentID` (for the child's own run) at the end of `Run()`.

- [ ] **Step 3: Run tests + build**

```
go build ./...
go test ./internal/batch ./internal/pipeline -count=1
```

- [ ] **Step 4: Commit**

```
git add internal/batch/batch.go internal/pipeline/runner.go internal/batch/batch_test.go
git commit -m "feat(batch): parent_run_id + children/ layout for batch runs"
```

---

## Phase 4 — CLI Reshape

### Task 13: Top-level `glitch run` extracted from `workflow run`

**Files:**
- Create: `cmd/run.go`
- Create: `cmd/run_test.go`
- Modify: `cmd/workflow.go` (remove `run` subcommand and its helpers once `run.go` owns them; keep shared helpers in a new `cmd/workflow_helpers.go` if needed)

- [ ] **Step 1: Copy the existing `workflow run` implementation** — read `cmd/workflow.go`, identify the `workflowRunCmd` definition, and migrate its full `RunE` into `cmd/run.go` under a top-level `glitch run` cobra command. Preserve all flags: `--set`, `--path`, `--results-dir`, `--variant`, `--compare`, `--review-criteria`. Register it on `rootCmd.AddCommand(runCmd)` in `init()`.

Skeleton:

```go
package cmd

import "github.com/spf13/cobra"

var runCmd = &cobra.Command{
	Use:   "run <workflow> [input]",
	Short: "run a workflow (resolves in active workspace, falls back to global)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRun,
}

func init() { rootCmd.AddCommand(runCmd) }

// runRun copies the body of the old workflowRunCmd.RunE with one change:
// workflows are resolved from activeWorkspace's workflows/ first, then global.
func runRun(cmd *cobra.Command, args []string) error {
	// IMPLEMENT: paste workflowRunCmd body here. Use resolveWorkspaceForCommand()
	// to determine the lookup directory.
	return nil
}
```

- [ ] **Step 2: Test a smoke scenario**

Create a minimal workflow in a temp workspace and assert `glitch run <name>` resolves correctly. Use the existing test pattern from `cmd/workflow_test.go` if any.

- [ ] **Step 3: Remove `workflow run` from `cmd/workflow.go`** — delete the subcommand registration and RunE body. Leave a note at the top: `// Deprecated: all workflow subcommands have moved under `glitch workspace workflow …` and `glitch run`.`

- [ ] **Step 4: Build + test**

```
go build ./...
go test ./cmd -count=1
```

- [ ] **Step 5: Commit**

```
git add cmd/run.go cmd/run_test.go cmd/workflow.go
git commit -m "feat(cmd): glitch run at top level; remove workflow run"
```

---

### Task 14: `workspace init|use|list|status|register|unregister`

**Files:**
- Create: `cmd/workspace_init.go`
- Create: `cmd/workspace_use.go`
- Create: `cmd/workspace_list.go`
- Create: `cmd/workspace_status.go`
- Create: `cmd/workspace_register.go` (holds both register + unregister)
- Create: `cmd/workspace_init_test.go`

Each command follows the same pattern — cobra command, thin wrapper calling a package-private `run*` function. Keep the helpers small.

- [ ] **Step 1: `workspace init`** — scaffolds a new workspace at path (default CWD), writes `workspace.glitch`, adds to registry. Reject if `workspace.glitch` already exists.

- [ ] **Step 2: `workspace use <name>`** — sets `~/.config/glitch/state.glitch` via `registry.SetActive(name)`. Error if name not in registry.

- [ ] **Step 3: `workspace list`** — print `registry.List()` as aligned name/path columns.

- [ ] **Step 4: `workspace status`** — show active workspace name + path, then its resources with last-fetched times from `resources_state`, then the last 5 rows from `runs` table for that workspace.

- [ ] **Step 5: `workspace register <path> [--as name]`** — read `workspace.glitch` at path, fall back to `--as` name, call `registry.Add(Entry{...})`.

- [ ] **Step 6: `workspace unregister <name>`** — `registry.Remove(name)`.

- [ ] **Step 7: Test** — `cmd/workspace_init_test.go`:

```go
func TestWorkspaceInitCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	target := filepath.Join(dir, "wsx")
	if err := runWorkspaceInit(target, "wsx"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "workspace.glitch")); err != nil {
		t.Fatalf("workspace.glitch missing: %v", err)
	}
	entries, _ := registry.List()
	if len(entries) != 1 || entries[0].Name != "wsx" {
		t.Fatalf("registry not updated: %+v", entries)
	}
}
```

- [ ] **Step 8: Build + tests**

```
go build ./...
go test ./cmd -run TestWorkspace -v -count=1
```

- [ ] **Step 9: Commit**

```
git add cmd/workspace_init.go cmd/workspace_init_test.go cmd/workspace_use.go cmd/workspace_list.go cmd/workspace_status.go cmd/workspace_register.go
git commit -m "feat(cmd): workspace init/use/list/status/register/unregister"
```

---

### Task 15: `workspace workflow list|new|edit` + `workspace gui`

**Files:**
- Create: `cmd/workspace_workflow.go`
- Create: `cmd/workspace_gui.go`
- Delete: `cmd/gui.go`

- [ ] **Step 1: Implement `workspace workflow`** — port the existing `workflow list` and `workflow new` logic into subcommands under a new `workspaceWorkflowCmd`. Add `edit` which execs `$EDITOR <path>` on the workflow file.

- [ ] **Step 2: Implement `workspace gui`** — move the body of `cmd/gui.go` into a command registered under `workspaceCmd`. Delete `cmd/gui.go`.

- [ ] **Step 3: Remove `workflow list|new|gui` from `cmd/workflow.go`** — delete those subcommand registrations. If `workflow.go` is now empty (or nearly), delete the file entirely.

- [ ] **Step 4: Build + smoke test**

```
go build ./...
./glitch workspace --help
./glitch workspace workflow --help
./glitch workspace gui --help
```

All should print without errors.

- [ ] **Step 5: Commit**

```
git add cmd/workspace_workflow.go cmd/workspace_gui.go
git rm cmd/gui.go cmd/workflow.go  # if applicable
git commit -m "feat(cmd): move workflow list/new/gui under workspace"
```

---

## Phase 5 — GUI

### Task 16: GUI backend — workspaces + active + use endpoints

**Files:**
- Create: `internal/gui/api_workspaces.go`
- Create: `internal/gui/api_workspaces_test.go`
- Modify: `internal/gui/server.go:70-101` (register routes)

- [ ] **Step 1: Write failing test** — `internal/gui/api_workspaces_test.go`:

```go
package gui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListWorkspacesReturnsArray(t *testing.T) {
	srv := newTestServer(t) // helper — existing GUI tests use an analogous pattern
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	// expect JSON array
	if w.Body.Bytes()[0] != '[' {
		t.Fatalf("expected JSON array, got %s", w.Body.String())
	}
}
```

- [ ] **Step 2: Implement** — `internal/gui/api_workspaces.go`:

```go
package gui

import (
	"encoding/json"
	"net/http"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

type workspaceRegistryEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Active bool   `json:"active"`
}

func (s *Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	entries, _ := registry.List()
	active, _ := registry.GetActive()
	out := make([]workspaceRegistryEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, workspaceRegistryEntry{Name: e.Name, Path: e.Path, Active: e.Name == active})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) handleUseWorkspace(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := registry.SetActive(body.Name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "active": body.Name})
}
```

- [ ] **Step 3: Register routes** — in `internal/gui/server.go` `routes()`:

```go
s.mux.HandleFunc("GET /api/workspaces", s.handleListWorkspaces)
s.mux.HandleFunc("POST /api/workspaces/use", s.handleUseWorkspace)
```

- [ ] **Step 4: Build + tests**

```
go build ./...
go test ./internal/gui -run TestListWorkspaces -v -count=1
```

- [ ] **Step 5: Commit**

```
git add internal/gui/api_workspaces.go internal/gui/api_workspaces_test.go internal/gui/server.go
git commit -m "feat(gui): workspaces registry and use endpoints"
```

---

### Task 17: GUI backend — resources CRUD + sync + pin

**Files:**
- Create: `internal/gui/api_resources.go`
- Create: `internal/gui/api_resources_test.go`
- Modify: `internal/gui/server.go`

Endpoints:
- `GET /api/workspace/resources` — list active workspace's resources with `:fetched` from `resources.glitch`
- `POST /api/workspace/resources` — body `{url_or_path, name?, pin?}` → calls `runWorkspaceAdd`
- `DELETE /api/workspace/resources/:name` — calls `runWorkspaceRm`
- `POST /api/workspace/sync` + `POST /api/workspace/sync/:name` — calls `runWorkspaceSync`
- `POST /api/workspace/pin` — body `{name, ref}` → calls `runWorkspacePin`

Each handler resolves the active workspace path via `resolveWorkspaceForCommand()` (export a helper from cmd package or replicate the precedence logic). Shell out to the cmd package helpers where possible to avoid duplication.

- [ ] **Step 1: Test (one endpoint)** — `GET /api/workspace/resources` returns 404 when no active workspace, 200 + JSON array when active.

- [ ] **Step 2: Implement**, keep one handler per endpoint.

- [ ] **Step 3: Register routes**

- [ ] **Step 4: Build + tests**

- [ ] **Step 5: Commit**

```
git commit -m "feat(gui): workspace resources CRUD + sync + pin endpoints"
```

---

### Task 18: GUI backend — run tree endpoint

**Files:**
- Create: `internal/gui/api_runtree.go`
- Create: `internal/gui/api_runtree_test.go`
- Modify: `internal/gui/server.go`

- [ ] **Step 1: Test** — insert parent + 2 children via `store.RecordRun`, then `GET /api/runs/:id/tree` returns nested JSON.

- [ ] **Step 2: Implement**

```go
func (s *Server) handleGetRunTree(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	tree, err := s.store.GetRunTree(id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

func (s *Server) handleListChildren(w http.ResponseWriter, r *http.Request) {
	parentID, _ := strconv.ParseInt(r.URL.Query().Get("parent_id"), 10, 64)
	kids, err := s.store.ListChildren(parentID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kids)
}
```

Register:

```go
s.mux.HandleFunc("GET /api/runs/{id}/tree", s.handleGetRunTree)
// s.mux.HandleFunc("GET /api/runs", s.handleListRuns) already exists — extend it to honour ?parent_id when set, delegating to handleListChildren.
```

- [ ] **Step 3: Commit**

```
git commit -m "feat(gui): run tree endpoints (/api/runs/:id/tree, ?parent_id filter)"
```

---

### Task 19: GUI frontend — workspace switcher + resources panel

**Files:**
- Create: `gui/src/lib/components/WorkspaceSwitcher.svelte`
- Create: `gui/src/lib/components/ResourcesPanel.svelte`
- Modify: `gui/src/App.svelte` (mount switcher in top bar)
- Modify: `gui/src/lib/api.js` (new helpers)
- Modify: `gui/src/routes/Settings.svelte` (embed ResourcesPanel)

- [ ] **Step 1: Extend api.js** with:

```js
export const listWorkspaces = () => request('/api/workspaces');
export const useWorkspace = (name) => request('/api/workspaces/use', { method: 'POST', body: { name } });
export const listResources = () => request('/api/workspace/resources');
export const addResource = (payload) => request('/api/workspace/resources', { method: 'POST', body: payload });
export const removeResource = (name) => request(`/api/workspace/resources/${name}`, { method: 'DELETE' });
export const syncWorkspace = (name) => request(name ? `/api/workspace/sync/${name}` : '/api/workspace/sync', { method: 'POST' });
export const pinResource = (name, ref) => request('/api/workspace/pin', { method: 'POST', body: { name, ref } });
```

- [ ] **Step 2: WorkspaceSwitcher.svelte** — dropdown showing active name, list of registered workspaces; on change calls `useWorkspace(name)` and reloads page.

- [ ] **Step 3: ResourcesPanel.svelte** — table with columns (name, type, ref, pin, fetched). Row actions: Sync, Pin, Remove. "Add resource" button opens a modal (URL/path input, optional `--as`, optional `--pin`).

- [ ] **Step 4: App.svelte** — import WorkspaceSwitcher and mount it above the sidebar.

- [ ] **Step 5: Playwright test** — add `gui/e2e/workspace.spec.js` covering: switcher lists workspaces, clicking switches active, resources panel lists entries.

- [ ] **Step 6: Build + test**

```
task build
task gui:test
```

- [ ] **Step 7: Commit**

```
git commit -m "feat(gui): workspace switcher and resources panel"
```

---

### Task 20: GUI frontend — runs tree view + results browser with children

**Files:**
- Create: `gui/src/lib/components/RunTree.svelte`
- Modify: `gui/src/routes/RunList.svelte` (add tree tab)
- Modify: `gui/src/routes/ResultsBrowser.svelte` (expand `children/` entries)

- [ ] **Step 1: RunTree.svelte** — recursive component; each node shows name/status/kind + collapsible children fetched via `/api/runs/:id/tree`.

- [ ] **Step 2: RunList.svelte** — add a "Tree" tab alongside the existing flat list. Switch on tab state.

- [ ] **Step 3: ResultsBrowser.svelte** — when a directory has a `children/` subfolder, render its contents as expandable siblings under the parent run's node in the file tree.

- [ ] **Step 4: Playwright test** — `gui/e2e/run-tree.spec.js` covering that parent + children render with collapse toggles.

- [ ] **Step 5: Build + tests**

- [ ] **Step 6: Commit**

```
git commit -m "feat(gui): runs tree view and children expansion in results browser"
```

---

## Phase 6 — `map-resources` (stretch)

### Task 21: `map-resources` sexpr form

**Files:**
- Modify: `internal/pipeline/types.go` (Step fields)
- Modify: `internal/pipeline/sexpr.go` (parser)
- Modify: `internal/pipeline/runner.go` (executor)
- Create: `internal/pipeline/mapresources_test.go`

- [ ] **Step 1: Failing test** — construct a workspace with two git resources, run a workflow using `(map-resources :type "git" (step "x" (run "echo {{.resource.item.name}}")))`, assert output has both names.

- [ ] **Step 2: Parse** — new head `"map-resources"`:

```go
type Step struct {
    // ... existing ...
    MapResourcesType string // optional filter
    MapResourcesBody *Step
}
```

`convertMapResources` walks trailing keyword args and captures child body step.

- [ ] **Step 3: Execute** — filter `rctx.resources` by type if provided, then for each matching resource, bind `.resource.item` to `{name, type, path, url, ref, pin, repo}` and execute the body, collecting outputs joined by newline.

- [ ] **Step 4: Test + commit**

```
git commit -m "feat(pipeline): map-resources sexpr form"
```

---

## Phase 7 — Smoke tests

### Task 22: End-to-end integration smoke test

**Files:**
- Create: `cmd/e2e_test.go` (or `integration_test.go` at repo root)

- [ ] **Step 1: Write the test** — it should:

1. Create a tmp dir, run `glitch workspace init` via the in-process cobra command.
2. Run `glitch workspace add` with a local path (no network) resource.
3. Verify `workspace.glitch` contains the resource.
4. Run `glitch run` on a minimal workflow that references `{{.resource.notes.path}}`.
5. Assert the step output contains the expected path.
6. For call-workflow: define two workflows, parent calling child; run; assert the SQLite `runs` table has a parent row and a child row linked via `parent_run_id`.
7. Assert on-disk layout: parent result dir exists with a `children/` subdir containing the child run's results.

Build the test with `//go:build integration` so CI can gate it; run locally with `go test -tags=integration ./...`.

- [ ] **Step 2: Run it**

```
go test -tags=integration ./cmd -run TestE2E -v -count=1
```

- [ ] **Step 3: Commit**

```
git commit -m "test(cmd): end-to-end smoke test for workspace + resources + call-workflow"
```

---

### Task 23: GUI Playwright smoke test

**Files:**
- Create: `gui/e2e/workspace-mechanics.spec.js`

- [ ] **Step 1: Write the spec** — covers:

1. Start fresh: register two workspaces via the API (pre-test fixture).
2. Open GUI, confirm switcher shows both.
3. Click to switch; reload and confirm active state persists.
4. Open resources panel, add a local resource via modal.
5. Confirm resource appears in the table.
6. Sync it, confirm timestamp updates.
7. Open runs page, switch to tree tab, confirm rendering doesn't throw.

- [ ] **Step 2: Run**

```
task build
task gui:test
```

- [ ] **Step 3: Commit**

```
git commit -m "test(gui): playwright smoke for workspace switcher + resources panel"
```

---

## Phase 8 — Final review

### Task 24: Full-repo review checkpoint

This is a deliberate pause before completion — not code.

- [ ] **Step 1: Run the full suite** — `go test ./... && task build && task gui:test`. All green.
- [ ] **Step 2: Run `glitch smoke pack`** against ensemble/kibana/oblt-cli/observability-robots to verify the 24/24 baseline holds. The workspace mechanics shouldn't touch that path, but we verify.
- [ ] **Step 3: Self-review diff** — `git diff main..HEAD` and skim for:
  - Stray `fmt.Println`s, commented-out code.
  - Missing test files for any new `.go` file.
  - Any leftover YAML config references.
  - Any reintroduction of `--workspace` ceremony in places that should now use `resolveWorkspaceForCommand()`.
- [ ] **Step 4: Use `superpowers:requesting-code-review`** — spawn a code-reviewer agent with the full branch diff + the spec. Review covers correctness, convention adherence, and missing spec items.
- [ ] **Step 5: Address reviewer findings** — iterate until clean.
- [ ] **Step 6: Stop here.** Do not push. Hand back to the user with a summary of what's on the branch + reviewer outcome.

---

## Self-review (writer's own checklist)

**Spec coverage:**
- §1 Filesystem layout — covered by Tasks 6, 8, 9, 12 (resources/, children/, .glitch/resources.glitch)
- §2 workspace.glitch schema — Task 2
- §3 Resource semantics — Tasks 6, 7, 9, 10 (sync, pin, template refs)
- §4 Nested runs — Tasks 1, 11, 12
- §5 Workspace discovery — Tasks 3, 4
- §6 CLI surface — Tasks 13, 14, 15
- §7 Global config flip — Task 5
- §8 GUI changes — Tasks 16-20
- §9 Package landing zones — respected
- §10 Rollout phases — mirrored by Phase 1-6 of this plan
- §11 Breaking changes — all clean breaks per the no-migrations rule
- §12 Out of scope — map-resources ships in Task 21 (stretch); all other deferred items stay out

**Placeholder scan:** Tasks 14, 15, 17, 20 list bullets rather than full code for repetitive command / handler / Svelte scaffolds — acceptable because each follows a pattern shown in the earlier task of the same type (Task 8 for cobra commands, Task 16 for GUI handlers, Task 19 for Svelte components). Subagents instructed to follow those patterns should produce working code.

**Type consistency:**
- `workspace.Resource` (Task 2) uses `Type string` while `resource.Resource` (Task 6) uses `Kind Kind` — intentional split between persisted config and in-memory materialization domain. Conversion happens at the cmd layer in `runWorkspaceAdd` / `runWorkspaceSync`.
- `RunRecord.ParentRunID int64` (Task 1) consistent with `RunOpts.ParentRunID` (Task 11) and `resultPath(... runID int64)` (Task 12).
- `Result.RunID int64` (Task 12) added so batch + call-workflow can thread it through.

**Deferred constraints from spec:**
- Schema-version → glitch.db wipe: handled via the drift probe in Task 1 (any pre-parent_run_id DB is wiped).
- Format-preserving writer for :pin: Task 7.

---

**Execution:** user chose subagent-driven, one subagent per task, with test coverage + smoke tests + final review. Executor skill: `superpowers:subagent-driven-development`. Verification skill: `superpowers:verification-before-completion` between tasks. Review skill at the end: `superpowers:requesting-code-review`.

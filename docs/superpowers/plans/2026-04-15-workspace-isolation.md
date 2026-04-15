# Workspace Isolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make workspaces fully isolated — per-workspace SQLite, run-scoped result directories, explicit config precedence, and workspace identity in telemetry.

**Architecture:** Four independent changes: (1) `store.OpenForWorkspace` routes the DB into `<workspace>/.glitch/glitch.db`, (2) `SaveLoopResult` writes into timestamped run subdirectories with a `latest` symlink, (3) config load applies workspace defaults between global config and CLI flags, (4) `run.json` and the `runs` table carry a `workspace` field.

**Tech Stack:** Go, SQLite (modernc.org/sqlite), cobra, os/filepath, symlinks

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `internal/store/store.go` | Add `OpenForWorkspace` constructor, add `busy_timeout` pragma |
| Modify | `internal/store/store_test.go` | Tests for workspace-scoped DB and busy timeout |
| Modify | `internal/store/schema.go` | Add `workspace` column to `runs` table |
| Modify | `internal/store/store.go` | `RecordRun` accepts and inserts `workspace` field |
| Modify | `internal/research/results.go` | `runScopedDir`, `latest` symlink, `workspace` fields in `runJSON` |
| Modify | `internal/research/results_test.go` | Tests for run-scoped dirs, symlink, workspace in run.json |
| Modify | `cmd/config.go` | `ApplyWorkspace` function |
| Modify | `cmd/config_test.go` | Tests for config precedence with workspace defaults |
| Modify | `internal/gui/server.go` | Use `OpenForWorkspace` when workspace is set |
| Modify | `cmd/workspace_test.go` | Update integration test for run-scoped dirs |

---

### Task 1: Per-Workspace SQLite — Store Constructor

**Files:**
- Modify: `internal/store/store.go:30-52`
- Modify: `internal/store/store_test.go`

- [ ] **Step 1: Write the failing test for `OpenForWorkspace`**

Add to `internal/store/store_test.go`:

```go
func TestOpenForWorkspace(t *testing.T) {
	wsDir := t.TempDir()
	s, err := OpenForWorkspace(wsDir)
	if err != nil {
		t.Fatalf("OpenForWorkspace: %v", err)
	}
	defer s.Close()

	// DB should live at <workspace>/.glitch/glitch.db
	dbPath := filepath.Join(wsDir, ".glitch", "glitch.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected DB at %s: %v", dbPath, err)
	}

	// Should be functional
	id, err := s.RecordRun(RunRecord{Kind: "test", Name: "ws-test", Input: ""})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/store/ -run TestOpenForWorkspace -v`
Expected: FAIL — `OpenForWorkspace` not defined

- [ ] **Step 3: Implement `OpenForWorkspace` and add `busy_timeout` pragma**

In `internal/store/store.go`, add after the `Open` function (after line 36):

```go
// OpenForWorkspace opens a workspace-scoped store at <workspace>/.glitch/glitch.db.
func OpenForWorkspace(workspacePath string) (*Store, error) {
	return OpenAt(filepath.Join(workspacePath, ".glitch", "glitch.db"))
}
```

Also update `OpenAt` to include `busy_timeout` — change the `sql.Open` line:

```go
db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/store/ -run TestOpenForWorkspace -v`
Expected: PASS

- [ ] **Step 5: Write test for busy_timeout pragma**

Add to `internal/store/store_test.go`:

```go
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
```

- [ ] **Step 6: Run test to verify busy_timeout**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/store/ -run TestOpenAt_BusyTimeout -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add internal/store/store.go internal/store/store_test.go
git commit -m "feat(store): add OpenForWorkspace constructor and busy_timeout pragma"
```

---

### Task 2: Workspace Column in Schema and RecordRun

**Files:**
- Modify: `internal/store/schema.go:4-21` (runs table)
- Modify: `internal/store/store.go:65-73` (RunRecord struct)
- Modify: `internal/store/store.go:98-109` (RecordRun method)
- Modify: `internal/store/store_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/store/store_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/store/ -run TestRecordRun_Workspace -v`
Expected: FAIL — `RunRecord` has no `Workspace` field

- [ ] **Step 3: Add workspace column to schema**

In `internal/store/schema.go`, add `workspace` column to the `runs` table. Replace the full `CREATE TABLE IF NOT EXISTS runs` block with:

```sql
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
  variant       TEXT,
  workspace     TEXT NOT NULL DEFAULT ''
);
```

- [ ] **Step 4: Add `Workspace` field to `RunRecord` and update `RecordRun`**

In `internal/store/store.go`, add to `RunRecord`:

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
}
```

Update `RecordRun` to insert the workspace:

```go
func (s *Store) RecordRun(rec RunRecord) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO runs (kind, name, input, started_at, workflow_file, repo, model, variant, workspace)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.Kind, rec.Name, rec.Input, time.Now().UnixMilli(),
		rec.WorkflowFile, rec.Repo, rec.Model, rec.Variant, rec.Workspace,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/store/ -v`
Expected: ALL PASS (including existing tests — `Workspace` defaults to empty string)

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add internal/store/schema.go internal/store/store.go internal/store/store_test.go
git commit -m "feat(store): add workspace column to runs table and RunRecord"
```

---

### Task 3: Run-Scoped Result Directories

**Files:**
- Modify: `internal/research/results.go:63-89` (resultDir → runScopedDir)
- Modify: `internal/research/results.go:101-182` (SaveLoopResult)
- Modify: `internal/research/results.go:44-61` (runJSON struct)
- Modify: `internal/research/results_test.go`

- [ ] **Step 1: Write the failing test for run-scoped directories**

Add to `internal/research/results_test.go`:

```go
func TestSaveLoopResult_RunScopedDir(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-scoped-001",
		Document: ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/55",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "55"},
		},
		Goal:     GoalSummarize,
		Output:   "# Summary\n\nTest scoped result." + strings.Repeat(" pad", 200),
		LLMCalls: 1,
		Duration: 1 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	issueDir := filepath.Join(base, "elastic", "ensemble", "issue-55")

	// A "latest" symlink should exist
	latestLink := filepath.Join(issueDir, "latest")
	target, err := os.Readlink(latestLink)
	if err != nil {
		t.Fatalf("expected 'latest' symlink: %v", err)
	}

	// The target should be a directory name (not a full path)
	if filepath.IsAbs(target) {
		t.Fatalf("latest symlink should be relative, got %q", target)
	}

	// run.json should exist inside the run-scoped dir
	runDir := filepath.Join(issueDir, target)
	if _, err := os.Stat(filepath.Join(runDir, "run.json")); err != nil {
		t.Fatal("run.json not in run-scoped dir")
	}
	if _, err := os.Stat(filepath.Join(runDir, "README.md")); err != nil {
		t.Fatal("README.md not in run-scoped dir")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/research/ -run TestSaveLoopResult_RunScopedDir -v`
Expected: FAIL — no `latest` symlink exists (current code writes directly into the issue dir)

- [ ] **Step 3: Add `runScopedDir` and update `SaveLoopResult`**

In `internal/research/results.go`, add the import `"math/rand"` to the import block. Then add after the existing `resultDir` function:

```go
// runScopedDir returns a unique run directory under the issue/PR path.
// Format: <baseDir>/<org>/<repo>/<type>-<number>/<YYYYMMDDTHHMMSS>-<4hex>/
func runScopedDir(baseDir string, result LoopResult) string {
	parent := resultDir(baseDir, result)
	stamp := time.Now().UTC().Format("20060102T150405")
	suffix := fmt.Sprintf("%04x", rand.Intn(0xFFFF))
	return filepath.Join(parent, stamp+"-"+suffix)
}

// updateLatestSymlink creates or replaces a "latest" symlink in parentDir
// pointing to the given run directory name.
func updateLatestSymlink(parentDir, runDirName string) error {
	link := filepath.Join(parentDir, "latest")
	// Remove existing symlink (ignore error if it doesn't exist)
	os.Remove(link)
	return os.Symlink(runDirName, link)
}
```

Update `SaveLoopResult` — replace the line `dir := resultDir(baseDir, result)` with:

```go
dir := runScopedDir(baseDir, result)
```

And at the end of `SaveLoopResult`, before the final `return nil`, add:

```go
// Create/update the "latest" symlink in the parent directory
parentDir := filepath.Dir(dir)
runDirName := filepath.Base(dir)
if err := updateLatestSymlink(parentDir, runDirName); err != nil {
    return fmt.Errorf("results: update latest symlink: %w", err)
}
```

- [ ] **Step 4: Run new test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/research/ -run TestSaveLoopResult_RunScopedDir -v`
Expected: PASS

- [ ] **Step 5: Update existing tests that assert exact directory paths**

The existing tests (`TestSaveLoopResult`, `TestRunJSON_StandardFields`, `TestSaveLoopResult_WritesReadme`, `TestSaveLoopResultImplement`) check for files at `<base>/elastic/ensemble/issue-NNN/run.json` etc. They need updating to follow the `latest` symlink instead.

In `TestSaveLoopResult`, replace the line:
```go
dir := filepath.Join(base, "elastic", "ensemble", "issue-872")
```
with:
```go
issueDir := filepath.Join(base, "elastic", "ensemble", "issue-872")
dir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
if err != nil {
    t.Fatalf("latest symlink: %v", err)
}
```

In `TestRunJSON_StandardFields`, replace:
```go
dir := filepath.Join(base, "elastic", "ensemble", "issue-50")
data, err := os.ReadFile(filepath.Join(dir, "run.json"))
```
with:
```go
issueDir := filepath.Join(base, "elastic", "ensemble", "issue-50")
latestDir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
if err != nil {
    t.Fatalf("latest symlink: %v", err)
}
data, err := os.ReadFile(filepath.Join(latestDir, "run.json"))
```

In `TestSaveLoopResult_WritesReadme`, replace:
```go
dir := filepath.Join(base, "elastic", "ensemble", "issue-42")
readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
```
with:
```go
issueDir := filepath.Join(base, "elastic", "ensemble", "issue-42")
latestDir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
if err != nil {
    t.Fatalf("latest symlink: %v", err)
}
readme, err := os.ReadFile(filepath.Join(latestDir, "README.md"))
```

In `TestSaveLoopResultImplement`, replace:
```go
dir := filepath.Join(base, "elastic", "ensemble", "issue-100")
```
with:
```go
issueDir := filepath.Join(base, "elastic", "ensemble", "issue-100")
dir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
if err != nil {
    t.Fatalf("latest symlink: %v", err)
}
```

- [ ] **Step 6: Run all research tests**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/research/ -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add internal/research/results.go internal/research/results_test.go
git commit -m "feat(results): run-scoped directories with latest symlink"
```

---

### Task 4: Workspace Identity in run.json

**Files:**
- Modify: `internal/research/results.go:44-61` (runJSON struct)
- Modify: `internal/research/results.go:101-140` (SaveLoopResult — runJSON population)
- Modify: `internal/research/results_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/research/results_test.go`:

```go
func TestRunJSON_WorkspaceFields(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-ws-001",
		Document: ResearchDocument{
			Source:   "github_issue",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "77"},
		},
		Goal:          GoalSummarize,
		Output:        "summary" + strings.Repeat(" pad", 200),
		Workspace:     "stokagent",
		WorkspacePath: "/home/user/stokagent",
		LLMCalls:      1,
		Duration:      1 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	issueDir := filepath.Join(base, "elastic", "ensemble", "issue-77")
	latestDir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
	if err != nil {
		t.Fatalf("latest symlink: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(latestDir, "run.json"))
	if err != nil {
		t.Fatalf("read run.json: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if raw["workspace"] != "stokagent" {
		t.Fatalf("workspace: got %v, want stokagent", raw["workspace"])
	}
	if raw["workspace_path"] != "/home/user/stokagent" {
		t.Fatalf("workspace_path: got %v", raw["workspace_path"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/research/ -run TestRunJSON_WorkspaceFields -v`
Expected: FAIL — `LoopResult` has no `Workspace` field

- [ ] **Step 3: Add workspace fields to LoopResult, runJSON, and SaveLoopResult**

In `internal/research/results.go`, add to the `runJSON` struct (after `Escalations`):

```go
Workspace     string `json:"workspace"`
WorkspacePath string `json:"workspace_path"`
```

Find the `LoopResult` struct (in `internal/research/toolloop.go` or wherever it's defined) and add:

```go
Workspace     string
WorkspacePath string
```

In `SaveLoopResult`, update the `meta := runJSON{...}` block to include:

```go
Workspace:     result.Workspace,
WorkspacePath: result.WorkspacePath,
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/research/ -run TestRunJSON_WorkspaceFields -v`
Expected: PASS

- [ ] **Step 5: Run all research tests**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/research/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add internal/research/results.go internal/research/results_test.go internal/research/toolloop.go
git commit -m "feat(results): add workspace and workspace_path to run.json"
```

---

### Task 5: Config Precedence — ApplyWorkspace

**Files:**
- Modify: `cmd/config.go:19-26` (Config struct, add ApplyWorkspace)
- Modify: `cmd/config_test.go`

- [ ] **Step 1: Write the failing test**

Add to `cmd/config_test.go`:

```go
func TestApplyWorkspace(t *testing.T) {
	cfg := &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
	}

	ws := &workspace.Workspace{
		Defaults: workspace.Defaults{
			Model:    "llama3.2:3b",
			Provider: "lm-studio",
		},
	}

	ApplyWorkspace(ws, cfg)

	if cfg.DefaultModel != "llama3.2:3b" {
		t.Fatalf("DefaultModel: got %q, want llama3.2:3b", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "lm-studio" {
		t.Fatalf("DefaultProvider: got %q, want lm-studio", cfg.DefaultProvider)
	}
}

func TestApplyWorkspace_NilWorkspace(t *testing.T) {
	cfg := &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
	}

	ApplyWorkspace(nil, cfg)

	if cfg.DefaultModel != "qwen3:8b" {
		t.Fatalf("DefaultModel changed to %q", cfg.DefaultModel)
	}
}

func TestApplyWorkspace_PartialDefaults(t *testing.T) {
	cfg := &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
	}

	ws := &workspace.Workspace{
		Defaults: workspace.Defaults{
			Model: "llama3.2:3b",
			// Provider left empty — should not override
		},
	}

	ApplyWorkspace(ws, cfg)

	if cfg.DefaultModel != "llama3.2:3b" {
		t.Fatalf("DefaultModel: got %q", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "ollama" {
		t.Fatalf("DefaultProvider should stay ollama, got %q", cfg.DefaultProvider)
	}
}
```

Add the import `"github.com/8op-org/gl1tch/internal/workspace"` to the test file's import block.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./cmd/ -run TestApplyWorkspace -v`
Expected: FAIL — `ApplyWorkspace` not defined

- [ ] **Step 3: Implement `ApplyWorkspace`**

Add to `cmd/config.go`, after the imports (add `"github.com/8op-org/gl1tch/internal/workspace"` to imports):

```go
// ApplyWorkspace overlays workspace defaults onto the config.
// Call after loadConfig, before CLI flag overrides.
func ApplyWorkspace(ws *workspace.Workspace, cfg *Config) {
	if ws == nil {
		return
	}
	if ws.Defaults.Model != "" {
		cfg.DefaultModel = ws.Defaults.Model
	}
	if ws.Defaults.Provider != "" {
		cfg.DefaultProvider = ws.Defaults.Provider
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./cmd/ -run TestApplyWorkspace -v`
Expected: ALL 3 PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add cmd/config.go cmd/config_test.go
git commit -m "feat(config): add ApplyWorkspace for workspace config precedence"
```

---

### Task 6: Wire Workspace Store into GUI Server

**Files:**
- Modify: `internal/gui/server.go:32-51`

- [ ] **Step 1: Write the failing test**

Add a new file `internal/gui/server_workspace_test.go`:

```go
package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_WorkspaceStore(t *testing.T) {
	wsDir := t.TempDir()
	os.MkdirAll(filepath.Join(wsDir, "workflows"), 0o755)

	s, err := New(":0", wsDir, false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// DB should be in the workspace .glitch dir
	dbPath := filepath.Join(wsDir, ".glitch", "glitch.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected workspace DB at %s: %v", dbPath, err)
	}

	_ = s // server created successfully with workspace store
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/gui/ -run TestNew_WorkspaceStore -v`
Expected: FAIL — DB is created at global path, not workspace path

- [ ] **Step 3: Update `gui.New` to use workspace-scoped store**

In `internal/gui/server.go`, replace:

```go
st, err := store.Open()
if err != nil {
    return nil, fmt.Errorf("open store: %w", err)
}
```

with:

```go
var st *store.Store
var err error
if workspace != "" {
    st, err = store.OpenForWorkspace(workspace)
} else {
    st, err = store.Open()
}
if err != nil {
    return nil, fmt.Errorf("open store: %w", err)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/gui/ -run TestNew_WorkspaceStore -v`
Expected: PASS

- [ ] **Step 5: Run full GUI test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./internal/gui/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add internal/gui/server.go internal/gui/server_workspace_test.go
git commit -m "feat(gui): use workspace-scoped store when --workspace is set"
```

---

### Task 7: .glitch Directory Auto-Creation

**Files:**
- Modify: `cmd/ask.go` (or wherever workspace setup runs early)
- Modify: `cmd/workspace_test.go`

- [ ] **Step 1: Write the failing test**

Add to `cmd/workspace_test.go`:

```go
func TestEnsureWorkspaceDir(t *testing.T) {
	wsDir := t.TempDir()

	ensureWorkspaceDir(wsDir)

	dotGlitch := filepath.Join(wsDir, ".glitch")
	info, err := os.Stat(dotGlitch)
	if err != nil {
		t.Fatalf(".glitch dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".glitch is not a directory")
	}

	// Inner .gitignore should contain "*"
	gi, err := os.ReadFile(filepath.Join(dotGlitch, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	if string(gi) != "*\n" {
		t.Fatalf(".gitignore: got %q, want %q", string(gi), "*\n")
	}
}

func TestEnsureWorkspaceDir_Idempotent(t *testing.T) {
	wsDir := t.TempDir()

	ensureWorkspaceDir(wsDir)
	ensureWorkspaceDir(wsDir) // second call should not error

	gi, _ := os.ReadFile(filepath.Join(wsDir, ".glitch", ".gitignore"))
	if string(gi) != "*\n" {
		t.Fatalf(".gitignore content wrong after second call: %q", string(gi))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./cmd/ -run TestEnsureWorkspaceDir -v`
Expected: FAIL — `ensureWorkspaceDir` not defined

- [ ] **Step 3: Implement `ensureWorkspaceDir`**

Add to `cmd/ask.go` (near the other resolve functions):

```go
// ensureWorkspaceDir creates the .glitch state directory inside a workspace
// with an inner .gitignore. Idempotent — safe to call on every run.
func ensureWorkspaceDir(wsPath string) {
	dotGlitch := filepath.Join(wsPath, ".glitch")
	os.MkdirAll(dotGlitch, 0o755)
	giPath := filepath.Join(dotGlitch, ".gitignore")
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		os.WriteFile(giPath, []byte("*\n"), 0o644)
	}
}
```

Then call it early — in `resolveResultsDir()`, add at the top:

```go
func resolveResultsDir() string {
	if workspacePath != "" {
		ensureWorkspaceDir(workspacePath)
	}
	// ... rest unchanged
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./cmd/ -run TestEnsureWorkspaceDir -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add cmd/ask.go cmd/workspace_test.go
git commit -m "feat(workspace): auto-create .glitch state dir with .gitignore"
```

---

### Task 8: Update Workspace Integration Test

**Files:**
- Modify: `cmd/workspace_test.go`

- [ ] **Step 1: Update `TestWorkspaceIntegration` for run-scoped dirs**

Replace the existing result assertions in `TestWorkspaceIntegration` with symlink-aware checks:

```go
// Check result landed in workspace (run-scoped dir with latest symlink)
issueDir := filepath.Join(wsDir, "results", "elastic", "ensemble", "issue-99")
latestDir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
if err != nil {
	t.Fatalf("latest symlink: %v", err)
}
if _, err := os.Stat(filepath.Join(latestDir, "README.md")); err != nil {
	t.Fatal("README.md not in workspace results")
}
if _, err := os.Stat(filepath.Join(latestDir, "run.json")); err != nil {
	t.Fatal("run.json not in workspace results")
}
```

This replaces the block from line 58-65 in the current test.

- [ ] **Step 2: Run the updated integration test**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./cmd/ -run TestWorkspaceIntegration -v`
Expected: PASS

- [ ] **Step 3: Run all cmd tests**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./cmd/ -v`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec
git add cmd/workspace_test.go
git commit -m "test(workspace): update integration test for run-scoped dirs"
```

---

### Task 9: Full Build and Test Verification

- [ ] **Step 1: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go test ./...`
Expected: ALL PASS

- [ ] **Step 2: Build binary**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go build -o /dev/null ./cmd/glitch`
Expected: clean build, exit 0

- [ ] **Step 3: Verify no compilation warnings**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/workspace-isolation-spec && go vet ./...`
Expected: no output (clean)

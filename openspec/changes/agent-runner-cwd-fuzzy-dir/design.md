## Context

The agent runner overlay (`viewAgentModalBox` in `internal/switchboard/switchboard.go`) currently presents PROVIDER, MODEL, and PROMPT but has no concept of a working directory. Sessions are always launched from wherever orcai was started.

The new session flow uses a separate full-screen picker (`internal/picker/picker.go`) that duplicates provider and model selection. It also has its own working directory discovery (git repo scanning, worktree creation) that is not accessible from the agent runner overlay.

Pipelines are launched without any CWD selection — they always execute in orcai's launch directory.

A reusable fuzzy directory picker overlay is needed in both the agent runner (for ad-hoc agents) and the pipeline launcher.

## Goals / Non-Goals

**Goals:**
- Add a WORKING DIRECTORY field to the agent runner overlay, defaulting to `os.Getwd()` at orcai startup
- Implement a reusable fuzzy dir picker overlay that searches directory names under `~/` in real time
- Replace the new session full-screen picker with the agent runner overlay (same provider/model/prompt/cwd fields)
- Add a pipeline directory picker overlay that appears when launching a pipeline, using the same reusable component
- Remove or tombstone the legacy `picker.go` full-screen UI

**Non-Goals:**
- Changing how worktrees are created or named (existing `GetOrCreateWorktreeFrom` logic stays)
- Supporting arbitrary remote paths or non-local directories
- Full file browser (just directory names, not file contents)
- Sorting/ranking by recency or frequency

## Decisions

### D1: Reusable `dirpicker` BubbleTea component, not inline state

**Decision:** Extract directory picking into a standalone `internal/switchboard/dirpicker.go` component with its own `Model`, `Init()`, `Update()`, `View()`.

**Rationale:** Both agent runner and pipeline launcher need the same fuzzy dir behaviour. Inline state in switchboard would duplicate logic. A self-contained component can be embedded in any overlay.

**Alternative considered:** Add dir picking inline to `agentSection` struct. Rejected — would need to duplicate for pipelines.

---

### D2: Walk `~/` asynchronously on overlay open, cache results

**Decision:** When the dir picker opens, fire a `tea.Cmd` that walks `~/` (one level at a time, breadth-first, depth-limited to ~3) and streams directory names back as messages. Results accumulate in the component's model; the fuzzy filter re-runs on each new batch.

**Rationale:** `~/` can contain thousands of directories. Blocking the UI is not acceptable. Streaming results let the user start typing immediately while the walk continues.

**Alternative considered:** Pre-cache on orcai startup. Rejected — startup cost not justified; results would be stale if directories change.

**Depth limit:** 3 levels (e.g. `~/Projects/foo/bar`) balances coverage vs. walk time. Configurable via constant.

---

### D3: Fuzzy filter using `go-fuzzyfinder` algorithm (substring rank), not a dependency

**Decision:** Implement a simple inline fuzzy scorer (score by contiguous match length + position bonus) rather than importing a fuzzy-find library.

**Rationale:** The existing codebase uses `charmbracelet/bubbles` but has no fuzzy-find dependency. Adding one for a single component is heavyweight. A 30-line scorer is sufficient for directory name matching.

**Alternative considered:** `sahilm/fuzzy` package. Rejected — unnecessary dependency for this use case.

---

### D4: New session keybinding opens agent runner overlay, not picker

**Decision:** The keybinding that currently calls `picker.Run()` (the full-screen new session picker) is rewired to open the agent runner overlay (`agentModalOpen = true`) in switchboard. The agent runner overlay already handles provider/model/prompt; CWD is added in this change.

**Rationale:** The full-screen picker duplicates all the fields in the agent runner overlay. Removing it reduces surface area and unifies the UX.

**Migration:** `picker.go`'s git worktree utilities (`GetOrCreateWorktreeFrom`, `scanGitRepos`, `copyDotEnv`) are retained as a library package. Only the full-screen UI (`pickerModel`, `Run()`, `View()`) is removed.

---

### D5: Pipeline dir picker fires before pipeline execution, not as a pipeline step

**Decision:** When the user triggers pipeline run, a new `pipelineDirPickerOpen bool` state in switchboard shows the dir picker overlay first. On confirmation, the selected CWD is passed to the pipeline runner as an initial context variable (`cwd`).

**Rationale:** CWD selection is a launch-time concern, not a pipeline step. Keeping it in the UI layer avoids modifying the pipeline YAML schema.

## Risks / Trade-offs

- **Walk performance on large home dirs** → Mitigated by depth limit (3) and async streaming; user can type before walk completes
- **Legacy picker removal breaks existing callers** → The picker's `Run()` is only called from switchboard's new session handler; one callsite to migrate
- **Dir picker shows too many results** → Mitigated by only showing directories (not files) and fuzzy filtering; top 50 results shown
- **Pipeline CWD not persisted between runs** → Acceptable for now; future work could store per-pipeline default CWD in config

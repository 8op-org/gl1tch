# Agent Modal: use_brain Toggle + Top Border Fix — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix the agent modal top border appearing shifted left (overlaying the ORCAI title bar), and add a `use_brain` checkbox to the agent modal that injects brain context into both immediate and scheduled runs.

**Architecture:** Two independent changes — (1) a 2-line fix in `panelrender.OverlayCenter` to clamp `startRow` to min 1 and clip rows; (2) a new `agentUseBrain bool` field on the switchboard model, a 6th focus state in the agent modal, and brain preamble injection in `submitAgentJob` plus `writeSingleStepPipeline`.

**Tech Stack:** Go, BubbleTea TUI, `internal/panelrender`, `internal/switchboard`, `internal/pipeline` (`StoreBrainInjector`/`NewStoreBrainInjector`)

---

## Task 1: Fix OverlayCenter — write failing test

**Files:**
- Create: `internal/panelrender/overlay_test.go`

**Step 1: Write the failing test**

Create `internal/panelrender/overlay_test.go`:

```go
package panelrender_test

import (
	"strings"
	"testing"

	"github.com/adam-stokes/orcai/internal/panelrender"
)

// TestOverlayCenter_TitleBarPreserved verifies that when an overlay is exactly
// as tall as the base (startRow would be 0), the top row of the base is not
// overwritten by the overlay.
func TestOverlayCenter_TitleBarPreserved(t *testing.T) {
	base := strings.Join([]string{
		"TITLE BAR",
		"row1",
		"row2",
		"row3",
	}, "\n")
	// Overlay is same height as base (4 rows) — without the fix, startRow=0
	// and the overlay's first row replaces "TITLE BAR".
	overlay := strings.Join([]string{
		"[top border]",
		"[content1]",
		"[content2]",
		"[bottom]",
	}, "\n")

	result := panelrender.OverlayCenter(base, overlay, 40, 4)
	lines := strings.Split(result, "\n")

	if len(lines) == 0 {
		t.Fatal("empty result")
	}
	if strings.Contains(lines[0], "[top border]") {
		t.Errorf("row 0 (title bar) was overwritten by overlay: %q", lines[0])
	}
	if !strings.Contains(lines[0], "TITLE BAR") {
		t.Errorf("row 0 should still contain title bar, got: %q", lines[0])
	}
}

// TestOverlayCenter_SmallOverlay verifies that small overlays are still centered
// correctly (startRow stays > 1, so the min-1 clamp has no effect).
func TestOverlayCenter_SmallOverlay(t *testing.T) {
	base := strings.Repeat("base line\n", 20)
	overlay := "line1\nline2\nline3"

	result := panelrender.OverlayCenter(base, overlay, 20, 20)
	lines := strings.Split(result, "\n")

	// Overlay height = 3, base height = 20, expected startRow = (20-3)/2 = 8
	// The overlay content should appear around the middle, not at the top.
	if strings.Contains(lines[0], "line1") {
		t.Errorf("small overlay should be vertically centered, not at top")
	}
	if strings.Contains(lines[1], "line1") {
		t.Errorf("small overlay should not appear at row 1")
	}
}

// TestOverlayCenter_TallOverlayClipped verifies rows beyond h are not appended.
func TestOverlayCenter_TallOverlayClipped(t *testing.T) {
	base := strings.Repeat("base\n", 5)
	base = strings.TrimSuffix(base, "\n") // 5 lines
	// Overlay is 8 rows tall — would extend beyond h=5 without clipping.
	overlay := strings.Join([]string{"a", "b", "c", "d", "e", "f", "g", "h"}, "\n")

	result := panelrender.OverlayCenter(base, overlay, 10, 5)
	lines := strings.Split(result, "\n")

	if len(lines) > 5 {
		t.Errorf("result should be capped at h=5 lines, got %d", len(lines))
	}
}
```

**Step 2: Run it to confirm it fails**

```bash
cd /Users/stokes/Projects/orcai
go test ./internal/panelrender/... -run TestOverlayCenter -v
```

Expected: FAIL — `TestOverlayCenter_TitleBarPreserved` reports row 0 was overwritten.

---

## Task 2: Fix OverlayCenter implementation

**Files:**
- Modify: `internal/panelrender/panelrender.go:294-318`

**Step 1: Apply the fix**

In `OverlayCenter` at line ~305, change:
```go
startRow := pmax((h-popH)/2, 0)
```
to:
```go
startRow := pmax((h-popH)/2, 1)
```

Then in the loop at line ~308, add a clipping guard as the first line inside the `for`:
```go
for i, oLine := range overlayLines {
    row := startRow + i
    if row >= h {
        break
    }
    // ... rest unchanged
```

**Step 2: Run tests to confirm they pass**

```bash
go test ./internal/panelrender/... -run TestOverlayCenter -v
```

Expected: all three tests PASS.

**Step 3: Run full test suite to check for regressions**

```bash
go test ./internal/panelrender/... -v
go test ./internal/switchboard/... -v
```

Expected: all PASS.

**Step 4: Commit**

```bash
git add internal/panelrender/panelrender.go internal/panelrender/overlay_test.go
git commit -m "fix(panelrender): clamp overlay startRow to min 1, clip rows at h"
```

---

## Task 3: Export WriteSingleStepPipeline and add useBrain parameter

**Files:**
- Modify: `internal/switchboard/switchboard.go:715-746` (the `writeSingleStepPipeline` function)

**Step 1: Write the failing test**

Add to `internal/switchboard/switchboard_test.go`:

```go
// TestWriteSingleStepPipeline_UseBrain verifies use_brain: true appears in YAML
// when useBrain=true, and is absent when useBrain=false.
func TestWriteSingleStepPipeline_UseBrain(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ORCAI_PIPELINES_DIR", dir)

	path, err := switchboard.WriteSingleStepPipeline("test-brain", "opencode", "", "do a thing", true)
	if err != nil {
		t.Fatalf("WriteSingleStepPipeline: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pipeline: %v", err)
	}
	yaml := string(data)
	if !strings.Contains(yaml, "use_brain: true") {
		t.Errorf("expected use_brain: true in YAML, got:\n%s", yaml)
	}
}

func TestWriteSingleStepPipeline_NoBrain(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ORCAI_PIPELINES_DIR", dir)

	path, err := switchboard.WriteSingleStepPipeline("test-no-brain", "opencode", "", "do a thing", false)
	if err != nil {
		t.Fatalf("WriteSingleStepPipeline: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pipeline: %v", err)
	}
	if strings.Contains(string(data), "use_brain") {
		t.Errorf("expected no use_brain key in YAML when useBrain=false")
	}
}
```

**Step 2: Run to confirm it fails**

```bash
go test ./internal/switchboard/... -run TestWriteSingleStepPipeline -v
```

Expected: FAIL — `WriteSingleStepPipeline` is undefined (unexported).

**Step 3: Export and update writeSingleStepPipeline**

In `switchboard.go` at line ~719, rename and add the `useBrain bool` parameter:

```go
// WriteSingleStepPipeline generates a minimal single-step pipeline YAML for a
// scheduled agent run and writes it to the .agents/ subdirectory of the pipelines
// directory so it does not appear in the PIPELINES launcher panel. Returns the
// absolute path of the written file so the caller can reference it in a cron entry.
// Exported so tests can call it directly.
func WriteSingleStepPipeline(name, providerID, modelID, prompt string, useBrain bool) (string, error) {
    dir := filepath.Join(pipelinesDir(), ".agents")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return "", err
    }
    path := filepath.Join(dir, name+".pipeline.yaml")

    // Indent every line of the prompt for the YAML block scalar.
    var promptLines strings.Builder
    for _, line := range strings.Split(prompt, "\n") {
        promptLines.WriteString("      ")
        promptLines.WriteString(line)
        promptLines.WriteString("\n")
    }

    model := ""
    if modelID != "" {
        model = "\n    model: " + modelID
    }

    useBrainYAML := ""
    if useBrain {
        useBrainYAML = "\n    use_brain: true"
    }

    content := fmt.Sprintf("name: %s\nversion: \"1\"\nsteps:\n  - id: run\n    executor: %s%s%s\n    prompt: |\n%s",
        name, providerID, model, useBrainYAML, promptLines.String())

    if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
        return "", err
    }
    return path, nil
}
```

Then update the internal call site (line ~2518):
```go
pipelineFile, pipelineErr := WriteSingleStepPipeline(entryName, prov.ID, modelID, input, m.agentUseBrain)
```

(Note: `m.agentUseBrain` is added in Task 4 — use `false` as a placeholder compile stub for now, update after Task 4.)

**Step 4: Run tests**

```bash
go test ./internal/switchboard/... -run TestWriteSingleStepPipeline -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/switchboard/switchboard.go internal/switchboard/switchboard_test.go
git commit -m "feat(switchboard): export WriteSingleStepPipeline, add useBrain param"
```

---

## Task 4: Add agentUseBrain field and accessor to Model

**Files:**
- Modify: `internal/switchboard/switchboard.go` — `Model` struct, accessor

**Step 1: Add field and accessor**

In the `Model` struct (around line 278 near `agentCWD`), add:
```go
agentUseBrain         bool
```

Add an exported accessor (near the other `Agent*` accessors around line 445):
```go
func (m Model) AgentUseBrain() bool { return m.agentUseBrain }
```

**Step 2: Build to confirm it compiles**

```bash
go build ./internal/switchboard/...
```

Expected: success.

**Step 3: Write a test for the new accessor**

Add to `switchboard_test.go`:

```go
func TestAgentUseBrain_DefaultFalse(t *testing.T) {
    m := switchboard.New()
    if m.AgentUseBrain() {
        t.Error("agentUseBrain should default to false")
    }
}
```

**Step 4: Run the test**

```bash
go test ./internal/switchboard/... -run TestAgentUseBrain -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/switchboard/switchboard.go internal/switchboard/switchboard_test.go
git commit -m "feat(switchboard): add agentUseBrain field and AgentUseBrain accessor"
```

---

## Task 5: Update focus cycling and key handling (renumber 3→4 cwd, 4→5 schedule)

**Files:**
- Modify: `internal/switchboard/switchboard.go` — `handleAgentModal` and related

**Context:** Current focus map: 0=provider, 1=model, 2=prompt, 3=cwd, 4=schedule.
New map: 0=provider, 1=model, 2=prompt, **3=use_brain**, 4=cwd, 5=schedule.

**Step 1: Update tab cycling modulus**

Find line `m.agentModalFocus = (m.agentModalFocus + 1) % 5` (line ~2077) and change to `% 6`.

Find line `m.agentModalFocus = (m.agentModalFocus + 4) % 5` (line ~2092) and change `+ 4` to `+ 5` AND `% 5` to `% 6`.

**Step 2: Add use_brain toggle at focus 3, shift cwd to 4, schedule to 5**

Find the CWD handler block (line ~2148):
```go
if m.agentModalFocus == 3 {
```
Change to `== 4`.

Find the schedule textarea forward block (line ~2164):
```go
if m.agentModalFocus == 4 {
```
Change to `== 5`.

Find the up/down pass-through conditions (lines ~2107, ~2124):
```go
if m.agentModalFocus == 2 || m.agentModalFocus == 4 {
```
Change both to `== 2 || m.agentModalFocus == 5`.

**Step 3: Add focus 3 handler (use_brain toggle)**

Insert before the CWD handler (the `if m.agentModalFocus == 4 {` block after your rename):

```go
// USE BRAIN toggle: space or enter toggles the flag.
if m.agentModalFocus == 3 {
    if key == " " || key == "enter" {
        m.agentUseBrain = !m.agentUseBrain
    }
    return m, nil
}
```

**Step 4: Update focus on tab — prompt focus/blur calls**

In the switch inside the tab handler (around line ~2078), check that the case for focusing the prompt textarea (focus 2) still says `m.agent.prompt.Focus()` and the case for schedule (now focus 5) says `m.agentSchedule.Focus()`. Check that focus 3 and 4 blur both. The current `default:` case blurs both — that's fine for focus 3 (use_brain) and 4 (cwd).

**Step 5: Build and run tests**

```bash
go build ./internal/switchboard/...
go test ./internal/switchboard/... -v
```

Expected: all pass.

**Step 6: Commit**

```bash
git add internal/switchboard/switchboard.go
git commit -m "feat(switchboard): renumber agent modal focus 3→use_brain, 4→cwd, 5→schedule"
```

---

## Task 6: Update viewAgentModalBox rendering

**Files:**
- Modify: `internal/switchboard/switchboard.go:3084-3245`

**Step 1: Update section focus numbers**

In `viewAgentModalBox`, change:
- `m.agentModalFocus == 3` (WORKING DIRECTORY header) → `== 4` (line ~3187)
- `m.agentModalFocus == 3` ("press enter to browse" conditional) → `== 4` (line ~3197, 3201)
- `m.agentModalFocus == 4` (SCHEDULE header) → `== 5` (line ~3207)

**Step 2: Add use_brain checkbox row**

After the prompt textarea loop (line ~3183) and before the WORKING DIRECTORY spacer, insert:

```go
// ── USE BRAIN toggle ──────────────────────────────────────────────────────
rows = append(rows, boxRow("", modalW, modalBorderColor))
useBrainCheck := "[ ]"
if m.agentUseBrain {
    useBrainCheck = "[x]"
}
useBrainColor := pal.Dim
if m.agentModalFocus == 3 {
    useBrainColor = pal.Accent + aBld
}
useBrainRow := "  " + useBrainColor + useBrainCheck + " use brain context" + aRst
rows = append(rows, boxRow(useBrainRow, modalW, modalBorderColor))
```

**Step 3: Build and verify**

```bash
go build ./internal/switchboard/...
go test ./internal/switchboard/... -v
```

Expected: all pass.

**Step 4: Commit**

```bash
git add internal/switchboard/switchboard.go
git commit -m "feat(switchboard): render use_brain checkbox in agent modal"
```

---

## Task 7: Wire brain injection into submitAgentJob

**Files:**
- Modify: `internal/switchboard/switchboard.go` — imports section and `submitAgentJob`

**Step 1: Add pipeline import**

In the import block (line ~6), add:
```go
"github.com/adam-stokes/orcai/internal/pipeline"
```

**Step 2: Inject brain context for immediate runs**

In `submitAgentJob`, find the section where `jh.storeRunID` is set (line ~2632):

```go
if m.store != nil {
    if runID, err := m.store.RecordRunStart("agent", title, runMetadataJSON("", cwd, modelID)); err == nil {
        jh.storeRunID = runID
    }
}
```

Immediately after this block, add:

```go
if m.agentUseBrain && m.store != nil {
    inj := pipeline.NewStoreBrainInjector(m.store)
    if preamble, err := inj.ReadContext(context.Background(), jh.storeRunID); err == nil && preamble != "" {
        input = preamble + "\n\n" + input
    }
}
```

**Step 3: Update the WriteSingleStepPipeline call**

Find line ~2518:
```go
pipelineFile, pipelineErr := WriteSingleStepPipeline(entryName, prov.ID, modelID, input, false)
```
Change `false` to `m.agentUseBrain`.

**Step 4: Build**

```bash
go build ./internal/switchboard/...
```

Expected: success.

**Step 5: Run full test suite**

```bash
go test ./internal/switchboard/... -v
go test ./internal/panelrender/... -v
go test ./internal/pipeline/... -v
```

Expected: all pass.

**Step 6: Commit**

```bash
git add internal/switchboard/switchboard.go
git commit -m "feat(switchboard): inject brain context in agent modal immediate and scheduled runs"
```

---

## Task 8: Final smoke test and cleanup

**Step 1: Full repo build**

```bash
go build ./...
```

Expected: zero errors.

**Step 2: Full test suite**

```bash
go test ./... 2>&1 | tail -20
```

Expected: all PASS, zero FAIL.

**Step 3: Manual smoke check**

Run the switchboard and open the agent modal:
```bash
make run  # or: go run ./cmd/orcai/main.go
```

Verify:
- [ ] The agent modal top border `┌─── AGENT ───┐` is properly aligned with content borders
- [ ] The title bar row is untouched when the modal opens
- [ ] Tab cycles through: provider → model → prompt → **use brain** → cwd → schedule → provider
- [ ] Space/enter on "use brain" focus toggles `[ ]` ↔ `[x]`
- [ ] Submitting a run with `[x]` prepends brain context to the prompt (visible in tmux window)
- [ ] Scheduling with `[x]` produces a pipeline YAML with `use_brain: true`

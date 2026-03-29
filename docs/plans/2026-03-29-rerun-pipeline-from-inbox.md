# Re-run Pipeline/Agent from Inbox Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an `r` keybinding to the inbox that opens a `RerunModal` overlay for adding optional context and changing provider/model before re-running any inbox item.

**Architecture:** New `RerunModal` BubbleTea component in `internal/modal/rerun.go` embeds an existing `AgentPickerModel` and a `textarea`. The switchboard stores `showRerun bool` + `rerunModal modal.RerunModal`, opens the modal on `r` in the inbox, dispatches the job on `RerunConfirmedMsg`, and renders it as a full-screen overlay via `overlayCenter`. The inbox hint bar loses the `f filter:label` hint and gains `r re-run`.

**Tech Stack:** Go, BubbleTea (`github.com/charmbracelet/bubbles/textarea`), `internal/modal`, `internal/panelrender`, `internal/store`, `internal/switchboard`

---

### Task 1: Add `RerunModal` component

**Files:**
- Create: `internal/modal/rerun.go`

**Step 1: Write the file**

```go
package modal

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/adam-stokes/orcai/internal/panelrender"
	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/store"
	"github.com/adam-stokes/orcai/internal/styles"
)

type rerunFocus int

const (
	rerunFocusContext rerunFocus = iota
	rerunFocusPicker
)

// RerunConfirmedMsg is dispatched when the user confirms a re-run.
type RerunConfirmedMsg struct {
	Run               store.Run
	AdditionalContext string // empty if not provided
	ProviderID        string
	ModelID           string
}

// RerunCancelledMsg is dispatched when the user cancels the re-run.
type RerunCancelledMsg struct{}

// RerunModal is a reusable overlay that prompts for optional additional context
// and lets the user pick a different provider/model before re-running a run.
type RerunModal struct {
	run      store.Run
	textarea textarea.Model
	picker   AgentPickerModel
	focus    rerunFocus
}

// NewRerunModal constructs a RerunModal for the given run.
// The picker is pre-seeded from run.Metadata if it contains a "model" slug.
// For agent runs the textarea is pre-populated with the first step's prompt.
func NewRerunModal(run store.Run, providers []picker.ProviderDef) RerunModal {
	ta := textarea.New()
	ta.Placeholder = "Additional context… (optional)"
	ta.CharLimit = 4096
	ta.ShowLineNumbers = false
	ta.SetWidth(60)
	ta.SetHeight(4)
	ta.Focus()

	// For agent runs, pre-fill with the original prompt from the first step.
	if run.Kind == "agent" && len(run.Steps) > 0 && run.Steps[0].Prompt != "" {
		ta.SetValue(run.Steps[0].Prompt)
	}

	pk := NewAgentPickerModel(providers)

	// Seed picker from run metadata "model" field (format: "providerID/modelID").
	var meta struct {
		Model string `json:"model"`
	}
	if json.Unmarshal([]byte(run.Metadata), &meta) == nil && meta.Model != "" {
		pk = pk.SelectBySlug(meta.Model)
	}

	return RerunModal{
		run:      run,
		textarea: ta,
		picker:   pk,
		focus:    rerunFocusContext,
	}
}

// Run returns the run this modal was constructed for.
func (m RerunModal) Run() store.Run { return m.run }

// Update handles key input. Returns the updated model and an optional tea.Cmd.
// Emits RerunConfirmedMsg or RerunCancelledMsg as appropriate.
func (m RerunModal) Update(msg tea.KeyMsg) (RerunModal, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return RerunCancelledMsg{} }

	case "ctrl+r":
		// Shortcut: run immediately from any focus zone.
		return m, m.confirmedCmd()

	case "tab", "shift+tab":
		if m.focus == rerunFocusContext {
			m.focus = rerunFocusPicker
			m.textarea.Blur()
		} else {
			m.focus = rerunFocusContext
			m.textarea.Focus()
		}
		return m, nil
	}

	if m.focus == rerunFocusContext {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	// Picker focus.
	var evt AgentPickerEvent
	m.picker, evt = m.picker.Update(msg)
	switch evt {
	case AgentPickerConfirmed:
		return m, m.confirmedCmd()
	case AgentPickerCancelled:
		return m, func() tea.Msg { return RerunCancelledMsg{} }
	}
	return m, nil
}

func (m RerunModal) confirmedCmd() tea.Cmd {
	return func() tea.Msg {
		return RerunConfirmedMsg{
			Run:               m.run,
			AdditionalContext: strings.TrimSpace(m.textarea.Value()),
			ProviderID:        m.picker.SelectedProviderID(),
			ModelID:           m.picker.SelectedModelID(),
		}
	}
}

// ViewBox renders the modal as a bordered box for use with panelrender.OverlayCenter.
func (m RerunModal) ViewBox(w int, pal styles.ANSIPalette) string {
	innerW := w - 4
	rst := panelrender.RST
	bld := panelrender.BLD

	// Truncate run name if needed.
	title := "RE-RUN: " + m.run.Name
	if len(title) > w-6 {
		title = title[:w-9] + "..."
	}

	rows := []string{
		panelrender.BoxTop(w, title, pal.Border, pal.Accent),
		panelrender.BoxRow("", w, pal.Border),
	}

	// Context section label.
	ctxLabel := pal.Dim + "  additional context" + rst
	if m.focus == rerunFocusContext {
		ctxLabel = pal.Accent + bld + "  ADDITIONAL CONTEXT" + rst
	}
	rows = append(rows, panelrender.BoxRow(ctxLabel, w, pal.Border))

	// Textarea lines.
	for _, line := range strings.Split(m.textarea.View(), "\n") {
		rows = append(rows, panelrender.BoxRow("  "+line, w, pal.Border))
	}

	rows = append(rows, panelrender.BoxRow("", w, pal.Border))

	// Picker section.
	for _, r := range m.picker.ViewRows(innerW, pal) {
		rows = append(rows, panelrender.BoxRow("  "+r, w, pal.Border))
	}

	hint := panelrender.HintBar([]panelrender.Hint{
		{Key: "tab", Desc: "focus"},
		{Key: "ctrl+r", Desc: "run"},
		{Key: "enter", Desc: "run (picker)"},
		{Key: "esc", Desc: "cancel"},
	}, innerW, pal)
	rows = append(rows, panelrender.BoxRow("  "+hint, w, pal.Border))
	rows = append(rows, panelrender.BoxRow("", w, pal.Border))
	rows = append(rows, panelrender.BoxBot(w, pal.Border))

	return strings.Join(rows, "\n")
}
```

**Step 2: Build to verify it compiles**

```bash
cd /Users/stokes/Projects/orcai && go build ./internal/modal/...
```

Expected: no errors.

**Step 3: Commit**

```bash
git add internal/modal/rerun.go
git commit -m "feat(modal): add RerunModal component with textarea + agent picker"
```

---

### Task 2: Wire `RerunModal` into the switchboard model

**Files:**
- Modify: `internal/switchboard/switchboard.go`

The changes in this task are split across four locations in the file.

**Step 1: Add state fields to `Model` struct** (near line 335, after `inboxEditorOrigContent`)

```go
// Re-run modal
rerunModal    modal.RerunModal
showRerun     bool
```

**Step 2: Guard keys when rerun modal is open** (line 1187, in the `if m.confirmQuit || m.helpOpen || ...` condition)

Add `|| m.showRerun` to the condition:

```go
if m.confirmQuit || m.helpOpen || m.jumpOpen || m.agentModalOpen || m.themePicker.Open || m.dirPickerOpen || m.confirmDelete || m.pipelineLaunchMode != plModeNone || m.showRerun {
```

**Step 3: Route keys to the rerun modal in `handleKey`** (at the top of `handleKey`, immediately after the `dirPickerOpen` block, around line 1228)

```go
// Re-run modal — capture all keys when open.
if m.showRerun {
    var cmd tea.Cmd
    m.rerunModal, cmd = m.rerunModal.Update(msg)
    return m, cmd
}
```

**Step 4: Handle `RerunConfirmedMsg` and `RerunCancelledMsg`** (in `Update`, after the `inbox.RunCompletedMsg` case around line 1209)

```go
case modal.RerunConfirmedMsg:
    m.showRerun = false
    return m.submitRerun(msg)

case modal.RerunCancelledMsg:
    m.showRerun = false
    return m, nil
```

**Step 5: Add `case "r":` in the inbox panel key handler** (in the `switch key {` block around line 1585, after the `"x":` case)

```go
case "r":
    runs := m.filteredInboxRuns()
    if len(runs) > 0 && m.inboxPanel.selectedIdx >= 0 && m.inboxPanel.selectedIdx < len(runs) {
        run := runs[m.inboxPanel.selectedIdx]
        m.rerunModal = modal.NewRerunModal(run, m.agent.providers)
        m.showRerun = true
    }
    return m, nil
```

**Step 6: Add `submitRerun` helper method** (after `submitAgentJob`, around line 2797)

```go
// submitRerun handles a confirmed re-run from the RerunModal.
// Agent runs are reconstructed as a new single-step pipeline with the (optionally
// appended) original prompt. Pipeline runs re-launch the original YAML file.
func (m Model) submitRerun(msg modal.RerunConfirmedMsg) (Model, tea.Cmd) {
    run := msg.Run
    additionalContext := msg.AdditionalContext

    var meta struct {
        PipelineFile string `json:"pipeline_file"`
        CWD          string `json:"cwd"`
    }
    _ = json.Unmarshal([]byte(run.Metadata), &meta)
    cwd := meta.CWD
    if cwd == "" {
        cwd = m.launchCWD
    }

    switch run.Kind {
    case "agent":
        // Reconstruct original prompt from the first step.
        originalPrompt := ""
        if len(run.Steps) > 0 {
            originalPrompt = run.Steps[0].Prompt
        }
        fullPrompt := originalPrompt
        if additionalContext != "" {
            fullPrompt += "\n\n---\nAdditional context:\n" + additionalContext
        }
        if fullPrompt == "" {
            return m, nil
        }

        entryName := fmt.Sprintf("agent-%s-%d", msg.ProviderID, time.Now().UnixNano())
        pipelineFile, err := WriteSingleStepPipeline(entryName, msg.ProviderID, msg.ModelID, fullPrompt, false)
        if err != nil {
            return m, nil
        }
        m.pendingPipelineName = entryName
        m.pendingPipelineYAML = pipelineFile
        return m.launchPendingPipeline(cwd)

    default: // "pipeline" and any future kinds
        yamlPath := meta.PipelineFile
        if yamlPath == "" {
            return m, nil
        }
        feedID := fmt.Sprintf("pipe-%d", time.Now().UnixNano())
        orcaiBin := orcaiBinaryPath()
        shellCmd := orcaiBin + " pipeline run " + yamlPath
        if additionalContext != "" {
            // Pass additional context as an env var the pipeline can reference.
            escaped := strings.ReplaceAll(additionalContext, "'", "'\\''")
            shellCmd = "ORCAI_CONTEXT='" + escaped + "' " + shellCmd
        }
        windowName, logFile, doneFile := createJobWindow(feedID, shellCmd, run.Name, cwd)
        ch := make(chan tea.Msg, 256)
        _, cancel := context.WithCancel(context.Background())
        jh := &jobHandle{id: feedID, cancel: cancel, ch: ch, tmuxWindow: windowName, logFile: logFile, pipelineName: run.Name}
        m.activeJobs[feedID] = jh
        startLogWatcher(feedID, logFile, doneFile, ch)
        return m, drainChan(ch)
    }
}
```

**Step 7: Build to verify it compiles**

```bash
cd /Users/stokes/Projects/orcai && go build ./...
```

Expected: no errors.

**Step 8: Commit**

```bash
git add internal/switchboard/switchboard.go
git commit -m "feat(switchboard): wire RerunModal — r key, submitRerun handler"
```

---

### Task 3: Render the rerun modal overlay in `View()`

**Files:**
- Modify: `internal/switchboard/switchboard.go`

**Step 1: Add the overlay render block** (in `View()`, after the `agentModalOpen` block, around line 3001)

```go
// Re-run modal — floating overlay on top of the switchboard.
if m.showRerun {
    base := topBar + "\n" + body
    return overlayCenter(base, m.rerunModal.ViewBox(w, m.ansiPalette()), w, h)
}
```

**Step 2: Build and run to verify visually**

```bash
cd /Users/stokes/Projects/orcai && go build ./... && echo "build ok"
```

Expected: no errors.

**Step 3: Commit**

```bash
git add internal/switchboard/switchboard.go
git commit -m "feat(switchboard): render RerunModal overlay in View"
```

---

### Task 4: Update inbox hint bar

**Files:**
- Modify: `internal/switchboard/switchboard.go`

**Step 1: Find the hint bar definition** (around line 3684)

Current:
```go
inboxHints = []panelrender.Hint{
    {Key: "enter", Desc: "open"},
    {Key: "x", Desc: "mark read"},
    {Key: "/", Desc: "search"},
    {Key: "f", Desc: "filter:" + filterLabel},
}
```

Replace with:
```go
inboxHints = []panelrender.Hint{
    {Key: "enter", Desc: "open"},
    {Key: "x", Desc: "mark read"},
    {Key: "/", Desc: "search"},
    {Key: "r", Desc: "re-run"},
}
```

**Step 2: Build**

```bash
cd /Users/stokes/Projects/orcai && go build ./...
```

**Step 3: Commit**

```bash
git add internal/switchboard/switchboard.go
git commit -m "feat(inbox): replace filter:unread hint with r re-run"
```

---

### Task 5: Write tests for RerunModal

**Files:**
- Create: `internal/modal/rerun_test.go`

**Step 1: Write failing tests**

```go
package modal_test

import (
	"encoding/json"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/adam-stokes/orcai/internal/modal"
	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/store"
	"github.com/adam-stokes/orcai/internal/styles"
)

func testRun(kind string) store.Run {
	meta, _ := json.Marshal(map[string]string{"model": "test-provider/model-a", "cwd": "/tmp"})
	r := store.Run{
		ID:       1,
		Kind:     kind,
		Name:     "test-run",
		Metadata: string(meta),
	}
	if kind == "agent" {
		r.Steps = []store.StepRecord{{ID: "step1", Prompt: "hello world"}}
	}
	return r
}

func testProviders() []picker.ProviderDef {
	return []picker.ProviderDef{{
		ID:    "test-provider",
		Label: "Test Provider",
		Models: []picker.ModelOption{
			{ID: "model-a", Label: "Model A"},
			{ID: "model-b", Label: "Model B"},
		},
	}}
}

func TestNewRerunModal_SeedsPickerFromMetadata(t *testing.T) {
	m := modal.NewRerunModal(testRun("agent"), testProviders())
	// Provider and model should be pre-seeded from metadata slug.
	if got := m.Run().Name; got != "test-run" {
		t.Fatalf("unexpected run name: %q", got)
	}
}

func TestNewRerunModal_AgentPreFillsTextarea(t *testing.T) {
	m := modal.NewRerunModal(testRun("agent"), testProviders())
	// ViewBox must render the original prompt somewhere in the box.
	pal := styles.ANSIPalette{Border: "\x1b[36m", Accent: "\x1b[35m", FG: "\x1b[97m", Dim: "\x1b[2m"}
	view := m.ViewBox(60, pal)
	if view == "" {
		t.Fatal("ViewBox returned empty string")
	}
}

func TestRerunModal_EscEmitsCancelledMsg(t *testing.T) {
	m := modal.NewRerunModal(testRun("agent"), testProviders())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a command on esc")
	}
	msg := cmd()
	if _, ok := msg.(modal.RerunCancelledMsg); !ok {
		t.Fatalf("expected RerunCancelledMsg, got %T", msg)
	}
}

func TestRerunModal_CtrlREmitsConfirmedMsg(t *testing.T) {
	m := modal.NewRerunModal(testRun("agent"), testProviders())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("expected a command on ctrl+r")
	}
	msg := cmd()
	confirmed, ok := msg.(modal.RerunConfirmedMsg)
	if !ok {
		t.Fatalf("expected RerunConfirmedMsg, got %T", msg)
	}
	if confirmed.Run.Name != "test-run" {
		t.Fatalf("unexpected run name in confirmed msg: %q", confirmed.Run.Name)
	}
}

func TestRerunModal_TabCyclesFocus(t *testing.T) {
	m := modal.NewRerunModal(testRun("agent"), testProviders())
	// Default focus is context (textarea).
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	// After tab, picker should have focus — a subsequent enter should emit confirmed.
	_, cmd := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command on enter in picker focus")
	}
	msg := cmd()
	if _, ok := msg.(modal.RerunConfirmedMsg); !ok {
		t.Fatalf("expected RerunConfirmedMsg after tab+enter, got %T", msg)
	}
}
```

**Step 2: Run the tests to verify they fail**

```bash
cd /Users/stokes/Projects/orcai && go test ./internal/modal/... -run TestRerunModal -v 2>&1 | head -40
```

Expected: FAIL (modal types not exported yet or test compilation issues — verify fails for the right reason).

**Step 3: Verify tests pass with the implementation in place**

```bash
cd /Users/stokes/Projects/orcai && go test ./internal/modal/... -run TestRerunModal -v
```

Expected: all PASS.

**Step 4: Commit**

```bash
git add internal/modal/rerun_test.go
git commit -m "test(modal): add RerunModal unit tests"
```

---

### Task 6: Final build + smoke test

**Step 1: Full build and test run**

```bash
cd /Users/stokes/Projects/orcai && go build ./... && go test ./... 2>&1 | tail -20
```

Expected: no failures.

**Step 2: Commit if any fixups were needed, otherwise done.**

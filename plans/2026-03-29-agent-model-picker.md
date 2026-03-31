# Agent/Model Picker Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a shared `AgentPickerModel` in `internal/modal/agentpicker.go` that replaces the promptmgr's broken model selector and the switchboard's inline provider/model picker code.

**Architecture:** `AgentPickerModel` owns all provider/model selection state and key handling. It exposes `ViewRows()` for inline embedding (switchboard) and `ViewBox()` for overlay use (promptmgr). A simple `AgentPickerEvent` enum signals confirm/cancel without tea.Cmd plumbing.

**Tech Stack:** Go, BubbleTea, `internal/panelrender`, `internal/styles`, `internal/picker`

---

### Task 1: AgentPickerModel — struct, constructor, accessors

**Files:**
- Create: `internal/modal/agentpicker.go`
- Create: `internal/modal/agentpicker_test.go`

**Step 1: Write failing tests**

```go
// internal/modal/agentpicker_test.go
package modal_test

import (
	"testing"

	"github.com/adam-stokes/orcai/internal/modal"
	"github.com/adam-stokes/orcai/internal/picker"
)

func testProviders() []picker.ProviderDef {
	return []picker.ProviderDef{
		{
			ID: "claude", Label: "Claude",
			Models: []picker.ModelOption{
				{ID: "sonnet", Label: "Sonnet"},
				{ID: "opus", Label: "Opus"},
			},
		},
		{
			ID: "ollama", Label: "Ollama",
			Models: []picker.ModelOption{
				{ID: "llama3", Label: "Llama 3"},
			},
		},
		{ID: "shell", Label: "Shell"},
	}
}

func TestAgentPickerModel_NewSelectsFirst(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	if m.SelectedProviderID() != "claude" {
		t.Errorf("want claude, got %q", m.SelectedProviderID())
	}
	if m.SelectedModelID() != "sonnet" {
		t.Errorf("want sonnet, got %q", m.SelectedModelID())
	}
	if m.SelectedModelLabel() != "Sonnet" {
		t.Errorf("want Sonnet, got %q", m.SelectedModelLabel())
	}
}

func TestAgentPickerModel_EmptyProviders(t *testing.T) {
	m := modal.NewAgentPickerModel(nil)
	if m.SelectedProviderID() != "" {
		t.Errorf("want empty, got %q", m.SelectedProviderID())
	}
	if m.SelectedModelID() != "" {
		t.Errorf("want empty, got %q", m.SelectedModelID())
	}
}

func TestAgentPickerModel_ProviderWithNoModels(t *testing.T) {
	m := modal.NewAgentPickerModel([]picker.ProviderDef{{ID: "shell", Label: "Shell"}})
	if m.SelectedProviderID() != "shell" {
		t.Errorf("want shell, got %q", m.SelectedProviderID())
	}
	if m.SelectedModelID() != "" {
		t.Errorf("want empty model, got %q", m.SelectedModelID())
	}
}
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/modal/... -run TestAgentPickerModel -v 2>&1 | head -20
```

Expected: `FAIL` — `modal.NewAgentPickerModel` undefined.

**Step 3: Write the struct, constructor, and accessors**

```go
// internal/modal/agentpicker.go
package modal

import (
	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/panelrender"
	"github.com/adam-stokes/orcai/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

const agentPickerWindow = 4

// AgentPickerEvent signals the result of a key event in the picker.
type AgentPickerEvent int

const (
	AgentPickerNone      AgentPickerEvent = iota
	AgentPickerConfirmed                  // enter was pressed
	AgentPickerCancelled                  // esc was pressed
)

// AgentPickerModel owns all provider/model selection state.
// Use NewAgentPickerModel to construct; pass providers from picker.BuildProviders().
type AgentPickerModel struct {
	providers         []picker.ProviderDef
	selectedProvider  int
	selectedModel     int
	provScrollOffset  int
	modelScrollOffset int
	focus             int // 0 = provider list, 1 = model list
}

// NewAgentPickerModel creates a picker pre-selecting the first provider and model.
func NewAgentPickerModel(providers []picker.ProviderDef) AgentPickerModel {
	return AgentPickerModel{providers: providers}
}

// SelectedProviderID returns the ID of the currently highlighted provider, or "".
func (m AgentPickerModel) SelectedProviderID() string {
	if m.selectedProvider >= len(m.providers) {
		return ""
	}
	return m.providers[m.selectedProvider].ID
}

// SelectedModelID returns the ID of the currently highlighted model, or "".
func (m AgentPickerModel) SelectedModelID() string {
	models := m.currentModels()
	if m.selectedModel >= len(models) {
		return ""
	}
	return models[m.selectedModel].ID
}

// SelectedModelLabel returns the display label of the highlighted model, or "".
func (m AgentPickerModel) SelectedModelLabel() string {
	models := m.currentModels()
	if m.selectedModel >= len(models) {
		return ""
	}
	lbl := models[m.selectedModel].Label
	if lbl == "" {
		lbl = models[m.selectedModel].ID
	}
	return lbl
}

// currentModels returns non-separator models for the selected provider.
func (m AgentPickerModel) currentModels() []picker.ModelOption {
	if m.selectedProvider >= len(m.providers) {
		return nil
	}
	return agentPickerFilterModels(m.providers[m.selectedProvider].Models)
}

// agentPickerFilterModels strips separator entries from a model list.
func agentPickerFilterModels(models []picker.ModelOption) []picker.ModelOption {
	out := make([]picker.ModelOption, 0, len(models))
	for _, mo := range models {
		if !mo.Separator {
			out = append(out, mo)
		}
	}
	return out
}
```

**Step 4: Run tests**

```bash
go test ./internal/modal/... -run TestAgentPickerModel -v
```

Expected: all three tests pass.

**Step 5: Commit**

```bash
git add internal/modal/agentpicker.go internal/modal/agentpicker_test.go
git commit -m "feat(modal): AgentPickerModel struct, constructor, accessors"
```

---

### Task 2: AgentPickerModel — Update (key handling)

**Files:**
- Modify: `internal/modal/agentpicker.go`
- Modify: `internal/modal/agentpicker_test.go`

**Step 1: Write failing tests**

Add to `agentpicker_test.go`:

```go
func key(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func keySpecial(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func TestAgentPickerUpdate_MoveDown(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	m2, ev := m.Update(key("j"))
	if ev != modal.AgentPickerNone {
		t.Errorf("want None, got %v", ev)
	}
	if m2.SelectedProviderID() != "ollama" {
		t.Errorf("want ollama after j, got %q", m2.SelectedProviderID())
	}
	// model resets when provider changes
	if m2.SelectedModelID() != "llama3" {
		t.Errorf("want llama3, got %q", m2.SelectedModelID())
	}
}

func TestAgentPickerUpdate_MoveUpAtTop(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	m2, _ := m.Update(key("k"))
	if m2.SelectedProviderID() != "claude" {
		t.Errorf("want claude (clamped), got %q", m2.SelectedProviderID())
	}
}

func TestAgentPickerUpdate_TabSwitchesFocus(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	m2, _ := m.Update(keySpecial(tea.KeyTab))
	// now in model focus; j moves model selection
	m3, _ := m2.Update(key("j"))
	if m3.SelectedModelID() != "opus" {
		t.Errorf("want opus, got %q", m3.SelectedModelID())
	}
	if m3.SelectedProviderID() != "claude" {
		t.Errorf("provider should not change in model focus, got %q", m3.SelectedProviderID())
	}
}

func TestAgentPickerUpdate_EnterConfirms(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	_, ev := m.Update(keySpecial(tea.KeyEnter))
	if ev != modal.AgentPickerConfirmed {
		t.Errorf("want Confirmed, got %v", ev)
	}
}

func TestAgentPickerUpdate_EscCancels(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	_, ev := m.Update(keySpecial(tea.KeyEsc))
	if ev != modal.AgentPickerCancelled {
		t.Errorf("want Cancelled, got %v", ev)
	}
}

func TestAgentPickerUpdate_ScrollClamps(t *testing.T) {
	// Build 6 providers to test scroll window
	providers := make([]picker.ProviderDef, 6)
	for i := range providers {
		providers[i] = picker.ProviderDef{ID: fmt.Sprintf("p%d", i), Label: fmt.Sprintf("P%d", i)}
	}
	m := modal.NewAgentPickerModel(providers)
	// Move down 5 times (past window of 4)
	for range 5 {
		m, _ = m.Update(key("j"))
	}
	if m.SelectedProviderID() != "p5" {
		t.Errorf("want p5, got %q", m.SelectedProviderID())
	}
}
```

You will need `"fmt"` imported in the test file.

**Step 2: Run to confirm fail**

```bash
go test ./internal/modal/... -run TestAgentPickerUpdate -v 2>&1 | head -30
```

Expected: FAIL — `Update` undefined.

**Step 3: Implement Update**

Add to `agentpicker.go`:

```go
// Update handles a key event and returns the updated model and an event signal.
// It does not emit tea.Cmd — all state changes are synchronous.
func (m AgentPickerModel) Update(msg tea.KeyMsg) (AgentPickerModel, AgentPickerEvent) {
	switch msg.String() {
	case "enter":
		return m, AgentPickerConfirmed
	case "esc":
		return m, AgentPickerCancelled
	case "tab":
		m.focus = 1 - m.focus
	case "j", "down":
		if m.focus == 0 {
			if m.selectedProvider < len(m.providers)-1 {
				m.selectedProvider++
				m.selectedModel = 0
				m.modelScrollOffset = 0
			}
			m = m.clampProvScroll()
		} else {
			models := m.currentModels()
			if m.selectedModel < len(models)-1 {
				m.selectedModel++
			}
			m = m.clampModelScroll()
		}
	case "k", "up":
		if m.focus == 0 {
			if m.selectedProvider > 0 {
				m.selectedProvider--
				m.selectedModel = 0
				m.modelScrollOffset = 0
			}
			m = m.clampProvScroll()
		} else {
			if m.selectedModel > 0 {
				m.selectedModel--
			}
			m = m.clampModelScroll()
		}
	}
	return m, AgentPickerNone
}

func (m AgentPickerModel) clampProvScroll() AgentPickerModel {
	if m.selectedProvider < m.provScrollOffset {
		m.provScrollOffset = m.selectedProvider
	}
	if m.selectedProvider >= m.provScrollOffset+agentPickerWindow {
		m.provScrollOffset = m.selectedProvider - agentPickerWindow + 1
	}
	return m
}

func (m AgentPickerModel) clampModelScroll() AgentPickerModel {
	if m.selectedModel < m.modelScrollOffset {
		m.modelScrollOffset = m.selectedModel
	}
	if m.selectedModel >= m.modelScrollOffset+agentPickerWindow {
		m.modelScrollOffset = m.selectedModel - agentPickerWindow + 1
	}
	return m
}
```

**Step 4: Run tests**

```bash
go test ./internal/modal/... -run TestAgentPickerUpdate -v
```

Expected: all pass.

**Step 5: Commit**

```bash
git add internal/modal/agentpicker.go internal/modal/agentpicker_test.go
git commit -m "feat(modal): AgentPickerModel.Update key handling + scroll clamping"
```

---

### Task 3: AgentPickerModel — ViewRows and ViewBox

**Files:**
- Modify: `internal/modal/agentpicker.go`
- Modify: `internal/modal/agentpicker_test.go`

**Step 1: Write failing tests**

```go
func TestAgentPickerViewRows_ContainsProviderAndModel(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	pal := styles.ANSIPalette{Accent: "", Dim: "", FG: "", SelBG: "", Border: ""}
	rows := m.ViewRows(60, pal)
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "PROVIDER") {
		t.Error("ViewRows missing PROVIDER label")
	}
	if !strings.Contains(joined, "MODEL") {
		t.Error("ViewRows missing MODEL label")
	}
	if !strings.Contains(joined, "Claude") {
		t.Error("ViewRows missing provider label Claude")
	}
	if !strings.Contains(joined, "Sonnet") {
		t.Error("ViewRows missing model label Sonnet")
	}
}

func TestAgentPickerViewBox_HasBorders(t *testing.T) {
	m := modal.NewAgentPickerModel(testProviders())
	pal := styles.ANSIPalette{}
	box := m.ViewBox(60, pal)
	if !strings.Contains(box, "AGENT / MODEL") {
		t.Error("ViewBox missing title")
	}
	// BoxTop starts with ┌ and BoxBot ends with ┘
	if !strings.Contains(box, "┌") || !strings.Contains(box, "└") {
		t.Error("ViewBox missing box-drawing borders")
	}
}
```

Also add `"strings"` to the test imports and `"github.com/adam-stokes/orcai/internal/styles"`.

**Step 2: Run to confirm fail**

```bash
go test ./internal/modal/... -run TestAgentPickerView -v 2>&1 | head -20
```

**Step 3: Implement ViewRows and ViewBox**

Add to `agentpicker.go`:

```go
// ViewRows returns ANSI-rendered rows for embedding into any panel.
// innerW is the available content width (no borders). Each row is a plain string
// that callers wrap in panelrender.BoxRow when embedding inline.
func (m AgentPickerModel) ViewRows(innerW int, pal styles.ANSIPalette) []string {
	rst := panelrender.RST
	bld := panelrender.BLD
	var rows []string

	// ── PROVIDER section ─────────────────────────────────────────────────────
	if m.focus == 0 {
		rows = append(rows, pal.Accent+bld+"  PROVIDER"+rst)
	} else {
		rows = append(rows, pal.Dim+"  PROVIDER"+rst)
	}

	if len(m.providers) == 0 {
		rows = append(rows, pal.Dim+"  no providers"+rst)
	} else {
		start := m.provScrollOffset
		end := start + agentPickerWindow
		if end > len(m.providers) {
			end = len(m.providers)
		}
		for i := start; i < end; i++ {
			p := m.providers[i]
			lbl := p.Label
			if lbl == "" {
				lbl = p.ID
			}
			if i == m.selectedProvider {
				if m.focus == 0 {
					rows = append(rows, pal.SelBG+"\x1b[97m"+"  > "+lbl+rst)
				} else {
					rows = append(rows, pal.Accent+"  > "+lbl+rst)
				}
			} else {
				rows = append(rows, pal.Dim+"    "+pal.FG+lbl+rst)
			}
		}
	}

	rows = append(rows, "")

	// ── MODEL section ─────────────────────────────────────────────────────────
	if m.focus == 1 {
		rows = append(rows, pal.Accent+bld+"  MODEL"+rst)
	} else {
		rows = append(rows, pal.Dim+"  MODEL"+rst)
	}

	models := m.currentModels()
	if len(models) == 0 {
		rows = append(rows, pal.Dim+"  no models"+rst)
	} else {
		start := m.modelScrollOffset
		end := start + agentPickerWindow
		if end > len(models) {
			end = len(models)
		}
		for i := start; i < end; i++ {
			mo := models[i]
			lbl := mo.Label
			if lbl == "" {
				lbl = mo.ID
			}
			if i == m.selectedModel {
				if m.focus == 1 {
					rows = append(rows, pal.SelBG+"\x1b[97m"+"  > "+lbl+rst)
				} else {
					rows = append(rows, pal.Accent+"  > "+lbl+rst)
				}
			} else {
				rows = append(rows, pal.Dim+"    "+pal.FG+lbl+rst)
			}
		}
	}

	rows = append(rows, "")
	rows = append(rows, panelrender.HintBar([]panelrender.Hint{
		{Key: "j/k", Desc: "nav"},
		{Key: "tab", Desc: "switch"},
		{Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "cancel"},
	}, innerW, pal))

	return rows
}

// ViewBox renders the picker as a standalone bordered overlay box suitable for
// use with panelrender.OverlayCenter.
func (m AgentPickerModel) ViewBox(boxW int, pal styles.ANSIPalette) string {
	rows := []string{
		panelrender.BoxTop(boxW, "AGENT / MODEL", pal.Border, pal.Accent),
		panelrender.BoxRow("", boxW, pal.Border),
	}
	for _, r := range m.ViewRows(boxW-4, pal) {
		rows = append(rows, panelrender.BoxRow("  "+r, boxW, pal.Border))
	}
	rows = append(rows, panelrender.BoxRow("", boxW, pal.Border))
	rows = append(rows, panelrender.BoxBot(boxW, pal.Border))
	return strings.Join(rows, "\n")
}
```

**Step 4: Run all modal tests**

```bash
go test ./internal/modal/... -v
```

Expected: all pass.

**Step 5: Build check**

```bash
go build ./...
```

**Step 6: Commit**

```bash
git add internal/modal/agentpicker.go internal/modal/agentpicker_test.go
git commit -m "feat(modal): AgentPickerModel.ViewRows and ViewBox rendering"
```

---

### Task 4: Promptmgr — wire AgentPickerModel, fix sub-focus

**Files:**
- Modify: `internal/promptmgr/model.go`
- Modify: `internal/promptmgr/update.go`

**Step 1: Update model.go**

In `model.go`, replace the `modelSlugs []string` and `modelIdx int` fields with the picker, and load providers on init.

Remove these fields from `Model`:
```go
modelSlugs []string
modelIdx   int
```

Add instead:
```go
agentPicker      modal.AgentPickerModel
agentPickerActive bool
```

Remove the `loadModelSlugsCmd` call from `Init()`. Instead, populate the picker in `New()`:

```go
import "github.com/adam-stokes/orcai/internal/picker"

func New(st *store.Store, pluginMgr *plugin.Manager, bundle *themes.Bundle) *Model {
    // ... existing textinput setup ...
    providers := picker.BuildProviders()
    return &Model{
        store:       st,
        pluginMgr:   pluginMgr,
        themeState:  tuikit.NewThemeState(bundle),
        filterInput: fi,
        titleInput:  ti,
        bodyInput:   bi,
        dirPicker:   modal.NewDirPickerModel(),
        agentPicker: modal.NewAgentPickerModel(providers),
    }
}
```

Also update `editorSubFocus` comment — it now has 4 slots: `0=title, 1=body, 2=model, 3=cwd`.

**Step 2: Update update.go**

Remove `modelSlugsLoadedMsg`, `loadModelSlugsCmd`, and all references to `m.modelSlugs` / `m.modelIdx`.

In `Update()`, add a new routing block **before** the `tea.KeyMsg` switch — when `agentPickerActive`, route all key events to the picker:

```go
case tea.KeyMsg:
    // When agent picker overlay is active, route all keys through it.
    if m.agentPickerActive {
        newPicker, ev := m.agentPicker.Update(msg)
        m.agentPicker = newPicker
        switch ev {
        case modal.AgentPickerConfirmed, modal.AgentPickerCancelled:
            m.agentPickerActive = false
        }
        return m, nil
    }
    // ... existing key handling ...
```

In `updateEditorPanel`, change `editorSubFocus` to cycle `0 → 1 → 2 → 3 → 0` (4 slots):

```go
case "tab":
    m.editorSubFocus = (m.editorSubFocus + 1) % 4
    m.syncEditorFocus()
    switch m.editorSubFocus {
    case 2:
        // model sub-focus: open picker on enter (not immediately)
    case 3:
        m.dirPickerActive = true
        return m, modal.DirPickerInit()
    }
```

Add `enter` handling in `updateEditorPanel` default/switch:

```go
case "enter":
    switch m.editorSubFocus {
    case 2:
        m.agentPickerActive = true
        return m, nil
    case 3:
        m.dirPickerActive = true
        return m, modal.DirPickerInit()
    }
```

Remove the old `[ ]` model cycling from `updateEditorPanel`.

Update `syncEditorFocus` — slot 2 is model (no text input), slot 3 is CWD (no text input):

```go
func (m *Model) syncEditorFocus() {
    m.titleInput.Blur()
    m.bodyInput.Blur()
    switch m.editorSubFocus {
    case 0:
        m.titleInput.Focus()
    case 1:
        m.bodyInput.Focus()
    // case 2: model picker — no text input
    // case 3: CWD dir picker — no text input
    }
}
```

Also update `updateEditorPanel`'s `shift+tab` to cycle `(m.editorSubFocus + 3) % 4` (reverse of 4-slot).

**Step 3: Build check**

```bash
go build ./...
```

Fix any compile errors (unused imports like `plugin` if `loadModelSlugsCmd` used it, etc.).

**Step 4: Commit**

```bash
git add internal/promptmgr/model.go internal/promptmgr/update.go
git commit -m "feat(promptmgr): replace model slug selector with AgentPickerModel"
```

---

### Task 5: Promptmgr — update view for picker overlay

**Files:**
- Modify: `internal/promptmgr/view.go`

**Step 1: Update buildEditorRows**

Replace the model section (currently `modelRow` with `◀ ▶` arrows) with a collapsed one-liner that shows the current selection and a hint:

```go
// MODEL row (collapsed — enter opens overlay).
modelActive := m.focusPanel == 1 && m.editorSubFocus == 2
provID := m.agentPicker.SelectedProviderID()
modelLabel := m.agentPicker.SelectedModelLabel()
modelSummary := "(none)"
if provID != "" {
    modelSummary = provID
    if modelLabel != "" {
        modelSummary += " / " + modelLabel
    }
}
pickHint := pal.Dim + "  [enter to pick]" + panelrender.RST
modelRow := "  " + sectionLabel("MODEL", modelActive) +
    "  " + pal.FG + modelSummary + panelrender.RST + pickHint
rows = append(rows, panelrender.BoxRow(modelRow, w, borderColor))
```

Update the CWD row to use sub-focus slot 3 (not 2):
```go
cwdActive := m.focusPanel == 1 && m.editorSubFocus == 3
```

**Step 2: Update View() to overlay agent picker**

In `View()`, after the dir picker overlay block, add:

```go
if m.agentPickerActive {
    overlay := m.agentPicker.ViewBox(min(m.width-4, 60), pal)
    return panelrender.OverlayCenter(base, overlay, m.width, m.height)
}
```

Place this **before** the `return base` at the end of `View()`.

**Step 3: Build check + run promptmgr tests**

```bash
go build ./... && go test ./internal/promptmgr/... -v
```

Expected: all tests pass. Fix any failures — likely test helpers that check `editorSubFocus` cycling.

**Step 4: Commit**

```bash
git add internal/promptmgr/view.go
git commit -m "feat(promptmgr): model picker overlay, collapsed MODEL row in editor"
```

---

### Task 6: Switchboard — replace inline provider/model state with AgentPickerModel

**Files:**
- Modify: `internal/switchboard/switchboard.go`

**Step 1: Update agentSection struct**

Find the `agentSection` struct (around line 125) and replace the four selection fields:

```go
// Remove:
selectedProvider       int
selectedModel          int
agentScrollOffset      int
agentModelScrollOffset int

// Add:
agentPicker modal.AgentPickerModel
```

Add import for `"github.com/adam-stokes/orcai/internal/modal"` at the top.

**Step 2: Initialize picker when modal opens**

Search for the two places where `agentModalOpen = true` is set with `agentModalFocus = 0` (around lines 2480-2481). At those points, also initialize the picker:

```go
m.agentModalOpen = true
m.agentModalFocus = 0
m.agent.agentPicker = modal.NewAgentPickerModel(m.agent.providers)
```

Also reset it when the modal closes (search for `m.agentModalOpen = false` and add `m.agent.agentPicker = modal.NewAgentPickerModel(m.agent.providers)` — or just leave the picker state; it re-initializes on next open).

**Step 3: Update handleAgentModal — replace focus 0/1 key handling**

Find the `handleAgentModal` function. Replace the `"up", "k"` and `"down", "j"` cases that check `agentModalFocus == 0` or `1` with picker delegation.

Replace this entire pattern:
```go
case "up", "k":
    if m.agentModalFocus == 2 || m.agentModalFocus == 5 {
        break
    }
    switch m.agentModalFocus {
    case 0:
        if m.agent.selectedProvider > 0 {
            m.agent.selectedProvider--
        }
    case 1:
        if m.agent.selectedModel > 0 {
            m.agent.selectedModel--
        }
    }
    return m, nil

case "down", "j":
    // ... similar
```

With:
```go
case "up", "k", "down", "j":
    if m.agentModalFocus == 2 || m.agentModalFocus == 5 {
        break // let textarea handle
    }
    if m.agentModalFocus == 0 || m.agentModalFocus == 1 {
        newPicker, ev := m.agent.agentPicker.Update(msg)
        m.agent.agentPicker = newPicker
        if ev == modal.AgentPickerConfirmed {
            m.agentModalFocus = 2
            m.agent.prompt.Focus()
        }
        return m, nil
    }
```

Also handle `tab` within focus 0/1: currently tab cycles all 6 slots. The picker handles its own internal tab (provider↔model). We need to differentiate: when `agentModalFocus == 0`, tab should go to the picker's model list first, then on next tab advance to focus slot 2.

Simplest approach: keep `agentModalFocus` at 0 while the picker is active (it handles provider/model focus internally via its own `focus` field). Advance to slot 2 only on `enter`. So tab in `handleAgentModal` when `agentModalFocus == 0` routes to the picker:

```go
case "tab":
    if m.agentModalFocus == 0 {
        // Let picker handle tab (switches provider↔model internally).
        // When picker focus wraps back to provider (second tab), advance to slot 2.
        if m.agent.agentPicker.Focus() == 1 {
            // currently on model; tab would wrap to provider, so advance outer focus
            m.agentModalFocus = 2
            m.agent.prompt.Focus()
            return m, nil
        }
        newPicker, _ := m.agent.agentPicker.Update(msg)
        m.agent.agentPicker = newPicker
        return m, nil
    }
    m.agentModalFocus = (m.agentModalFocus + 1) % 6
    // ... existing focus/blur logic
```

To support this, expose `Focus() int` on `AgentPickerModel` — add to `agentpicker.go`:

```go
// Focus returns the current internal focus: 0 = provider list, 1 = model list.
func (m AgentPickerModel) Focus() int { return m.focus }
```

**Step 4: Update submitAgentJob and currentProvider**

Replace `m.currentProvider()` usages in `submitAgentJob` with direct picker reads:

```go
provID := m.agent.agentPicker.SelectedProviderID()
if provID == "" {
    return m, nil
}
modelID := m.agent.agentPicker.SelectedModelID()

// Find the ProviderDef for pipeline args etc.
var prov *picker.ProviderDef
for i := range m.agent.providers {
    if m.agent.providers[i].ID == provID {
        prov = &m.agent.providers[i]
        break
    }
}
if prov == nil {
    return m, nil
}
```

Delete the `currentProvider()` method and `nonSepModels()` helper if they are no longer used anywhere else. (`grep -n "currentProvider\|nonSepModels"` to confirm.)

**Step 5: Update viewAgentModalBox — replace inline sections**

Find the PROVIDER and MODEL rendering sections in `viewAgentModalBox` (roughly lines 3097–3180). Replace them with:

```go
// ── PROVIDER + MODEL picker ───────────────────────────────────────────────
for _, r := range m.agent.agentPicker.ViewRows(modalW-4, pal) {
    rows = append(rows, boxRow("  "+r, modalW, modalBorderColor))
}
```

Also update the `fixedOverhead` constant at the top of `viewAgentModalBox` — reduce it by the difference in rows (the old inline sections had ~12 rows; the new ViewRows produces ~12 rows too, so the overhead stays roughly the same; double-check by counting).

**Step 6: Build and test**

```bash
go build ./... && go test ./internal/switchboard/... -v 2>&1 | tail -30
```

Fix any test failures. The switchboard tests that call `AgentModalFocus()` and navigate via `Update` will need to be aware that focus slot 0 now delegates to the picker.

**Step 7: Commit**

```bash
git add internal/switchboard/switchboard.go internal/modal/agentpicker.go
git commit -m "feat(switchboard): replace inline provider/model picker with AgentPickerModel"
```

---

### Task 7: Final wiring, cleanup, and full test run

**Files:**
- Modify: `internal/modal/agentpicker_test.go` (add Focus() test)
- Any remaining cleanup

**Step 1: Add Focus accessor test**

```go
func TestAgentPickerModel_FocusAccessor(t *testing.T) {
    m := modal.NewAgentPickerModel(testProviders())
    if m.Focus() != 0 {
        t.Errorf("want focus 0 initially, got %d", m.Focus())
    }
    m2, _ := m.Update(keySpecial(tea.KeyTab))
    if m2.Focus() != 1 {
        t.Errorf("want focus 1 after tab, got %d", m2.Focus())
    }
}
```

**Step 2: Full test suite**

```bash
go test ./... 2>&1 | tail -30
```

Expected: all pass.

**Step 3: Remove dead code**

- Delete `loadModelSlugsCmd` and `modelSlugsLoadedMsg` from `promptmgr/update.go` if still present (they should be gone from Task 4 but verify)
- Delete `runTokenMsg` and `statusClearMsg` types from `promptmgr/update.go` if unused
- Delete unused `keys` var from `promptmgr/keys.go` if unreferenced (or keep — it's harmless)

**Step 4: Final build**

```bash
go build ./...
```

**Step 5: Commit**

```bash
git add -u
git commit -m "chore(promptmgr): remove dead model-slug code and unused types"
```

---

**Plan complete and saved to `docs/plans/2026-03-29-agent-model-picker.md`.**

Two execution options:

**1. Subagent-Driven (this session)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Parallel Session (separate)** — Open a new session in this worktree, use `superpowers:executing-plans` to work through the tasks with checkpoints.

Which approach?

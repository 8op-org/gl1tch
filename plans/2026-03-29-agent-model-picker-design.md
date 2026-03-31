# Agent/Model Picker — Design

**Date:** 2026-03-29
**Status:** Approved

## Problem

The promptmgr model selector (`[`/`]` key cycling) is broken and has no visual list. The agent runner modal has a working scrollable provider+model picker, but it is 100+ lines of inline switchboard code that cannot be reused. The cwd picker also opens on tab instead of enter, which is unintuitive.

## Goal

A single shared `AgentPickerModel` component in `internal/modal/agentpicker.go` that both the promptmgr and switchboard agent modal use. The promptmgr uses it as an overlay; the switchboard slots its rows inline into the existing agent modal layout.

## Architecture

### New: `internal/modal/agentpicker.go`

Follows the same pattern as `DirPickerModel`.

**Types:**

```go
// AgentPickerModel owns all provider/model selection state.
type AgentPickerModel struct {
    providers         []picker.ProviderDef
    selectedProvider  int
    selectedModel     int
    provScrollOffset  int
    modelScrollOffset int
    focus             int // 0 = provider list, 1 = model list
}

// AgentPickerEvent is returned from Update to signal caller intent.
type AgentPickerEvent int
const (
    AgentPickerNone      AgentPickerEvent = iota
    AgentPickerConfirmed                  // enter pressed
    AgentPickerCancelled                  // esc pressed
)
```

**API:**

```go
func NewAgentPickerModel(providers []picker.ProviderDef) AgentPickerModel
func (m AgentPickerModel) Update(msg tea.KeyMsg) (AgentPickerModel, AgentPickerEvent)
func (m AgentPickerModel) ViewRows(innerW int, pal styles.ANSIPalette) []string
func (m AgentPickerModel) ViewBox(boxW int, pal styles.ANSIPalette) string
func (m AgentPickerModel) SelectedProviderID() string
func (m AgentPickerModel) SelectedModelID() string
func (m AgentPickerModel) SelectedModelLabel() string
```

**Key map (inside Update):**

| Key | Action |
|-----|--------|
| `j` / `↓` | move down in focused list |
| `k` / `↑` | move up in focused list |
| `tab` | switch focus: provider ↔ model |
| `enter` | return `AgentPickerConfirmed` |
| `esc` | return `AgentPickerCancelled` |

**Visual layout (ViewRows):**

```
  PROVIDER                     ← accent label when focused, dim otherwise
  > claude-code                ← SelBG + white when focused; accent ">" when not
    ollama
    shell
                               ← blank divider row
  MODEL
  > claude-sonnet-4-6
    claude-opus-4-6
                               ← blank divider row
  tab switch · enter confirm · esc cancel
```

Visible window: 4 rows per list (scrolls when providers/models exceed 4).

`ViewBox` wraps `ViewRows` in `BoxTop("AGENT / MODEL") + BoxBot`.

---

### Promptmgr changes (`internal/promptmgr/`)

**Model struct:**
- Remove `modelSlugs []string`, `modelIdx int`
- Add `agentPicker modal.AgentPickerModel`, `agentPickerActive bool`
- Providers loaded at init via `picker.BuildProviders()` (same as promptbuilder)

**Editor sub-focus** gains a 4th slot:

| Slot | Field | Activation |
|------|-------|------------|
| 0 | Title | tab |
| 1 | Body | tab |
| 2 | Model | tab → **enter** opens picker overlay |
| 3 | CWD | tab → **enter** opens dir picker (replaces current tab-to-open) |

Tab cycles `0 → 1 → 2 → 3 → 0`.

**Editor panel** shows collapsed one-liner when picker is closed:
```
  MODEL  claude-code / claude-sonnet-4-6  [enter to pick]
```

**Overlay:** when `agentPickerActive`, `View()` overlays `agentPicker.ViewBox()` centered on screen via `panelrender.OverlayCenter`.

**Update:** when `agentPickerActive`, route all key events to `agentPicker.Update()`. On `AgentPickerConfirmed`, close overlay and store selection. On `AgentPickerCancelled`, close overlay and discard.

---

### Switchboard changes (`internal/switchboard/switchboard.go`)

**Agent inner struct:**
- Remove `selectedProvider int`, `agentScrollOffset int`, `selectedModel int`, `agentModelScrollOffset int`
- Add `agentPicker modal.AgentPickerModel`
- Initialize picker when agent modal opens (same point providers were already loaded)

**`handleAgentModal`:**
- For focus slots 0/1, delegate to `agentPicker.Update(msg)` and check `AgentPickerEvent`
- On `AgentPickerConfirmed`, advance `agentModalFocus` to 2 (prompt textarea)
- On `AgentPickerCancelled`, close agent modal

**`viewAgentModalBox`:**
- Replace inline PROVIDER + MODEL section rows with `agentPicker.ViewRows(modalW-2, pal)` wrapped in `boxRow` calls
- Remove `nonSepModels()` helper (now internal to the picker)
- The fixed overhead constant adjusts to match new row count

**`submitAgentJob`:**
- Replace `m.agent.providers[m.agent.selectedProvider]` / `m.agent.selectedModel` reads with `m.agent.agentPicker.SelectedProviderID()` / `SelectedModelID()`

---

## Testing

- Unit tests for `AgentPickerModel`: navigation, scroll clamping, tab focus switch, enter/esc events, empty providers, providers with no models
- Existing promptmgr tests updated for new sub-focus numbering
- Existing switchboard agent modal tests updated for picker delegation
- Build + `go test ./...` must pass

## Files Touched

| File | Change |
|------|--------|
| `internal/modal/agentpicker.go` | new |
| `internal/modal/agentpicker_test.go` | new |
| `internal/promptmgr/model.go` | replace model/slug fields with picker |
| `internal/promptmgr/update.go` | route picker events, fix CWD to enter-to-open |
| `internal/promptmgr/view.go` | collapsed model line + picker overlay |
| `internal/switchboard/switchboard.go` | replace inline sections with shared picker |

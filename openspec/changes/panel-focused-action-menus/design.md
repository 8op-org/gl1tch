## Context

The switchboard TUI has a single `viewBottomBar(width int)` function (switchboard.go ~3299) that inspects which panel is currently focused and emits the corresponding hint strip as a full-terminal-width bar below all panels. The cron TUI has the equivalent `viewHintBar(width int)` (crontui/view.go ~267). Neither panel owns its own footer — they are passive boxes that rely on an external observer to produce help text at the screen edge.

This means:
- A user looking at a panel cannot tell, from within the panel, what keys it accepts.
- Adding a new panel or changing hints requires editing a central switch in a different part of the file.
- Hint-formatting code (`pal.Accent` key + `pal.Dim` description + `" · "` separator + padding) is inline and duplicated in every branch.

`panelrender` already provides shared header and box primitives used by all panels. It is the right home for a shared hint-bar helper.

## Goals / Non-Goals

**Goals:**
- Each panel/pane renders its own single-row hint footer **inside** its border box, visible only when that panel is focused.
- A single `panelrender.HintBar(hints []Hint, width int, pal Palette) string` helper encapsulates hint formatting and is used by all panels.
- The global `viewBottomBar` and `viewHintBar` functions are removed.
- Panel height arithmetic is updated so the footer row is accounted for without shrinking content.

**Non-Goals:**
- Multi-row action menus or pop-up keybinding overlays.
- Changing which actions are exposed per panel.
- Hint-bar animations or transitions.
- Panels not currently in the switchboard or cron TUI.

## Decisions

### 1. HintBar lives in `panelrender`, not a new package

**Decision:** Add `HintBar` to `internal/panelrender/panelrender.go`.

**Rationale:** `panelrender` already owns all shared panel-chrome primitives (header, box top/bottom, box row). Adding a footer primitive there keeps the package cohesive and avoids a new dependency. Both switchboard and crontui already import `panelrender`.

**Alternative considered:** A `hintbar` sub-package. Rejected — too small to warrant its own package, and it would fragment the panel-chrome responsibility.

---

### 2. Hint struct defined in `panelrender`

**Decision:** Define `type Hint struct { Key, Desc string }` in `panelrender`.

**Rationale:** Both switchboard panels and crontui panes need to pass hints to `HintBar`. A shared type avoids parallel definitions. The struct is minimal and stable.

**Alternative considered:** `[]string` pairs. Rejected — error-prone ordering, no named fields.

---

### 3. Empty hint list → no footer row rendered (zero height)

**Decision:** When `hints` is nil or empty, `HintBar` returns `""`. Each panel omits the footer row from its layout when unfocused.

**Rationale:** Panels should not reserve space for a footer they are not showing. Returning an empty string lets callers use `lipgloss.JoinVertical` without leaving a blank gap.

**Alternative considered:** Always render a blank footer row. Rejected — wastes a line on every unfocused panel and makes the focused panel's footer harder to see.

---

### 4. Height accounting: subtract 1 from content height when footer is shown

**Decision:** Each panel view function calculates `contentH = panelH - headerH - 1` when focused (footer present) and `contentH = panelH - headerH` when unfocused.

**Rationale:** This is the minimal change to keep existing scroll/layout math correct. It mirrors how the existing global bar was already "consuming" one terminal row — we are just moving that accounting into the panel.

**Alternative considered:** Always reserve the footer row (render blank when unfocused). Rejected — per decision 3 above.

---

### 5. Remove global bars entirely (not keep as fallback)

**Decision:** Delete `viewBottomBar` in switchboard and `viewHintBar` in crontui once all panels have their own footers.

**Rationale:** Leaving the global bar alongside per-panel footers would double-render hints and waste a terminal row. A clean removal is preferable to a guarded no-op.

## Risks / Trade-offs

- **Height arithmetic errors** → Each panel view function must carefully subtract 1 from content height when the footer is present. Off-by-one errors will cause scroll glitches or overflow. Mitigation: test each panel at both focused and unfocused states; add a comment marking the footer accounting line.
- **Panel too narrow for hints** → If a panel is very narrow, hints may overflow or wrap. Mitigation: `HintBar` truncates the rendered string to `width` with `lipgloss` max-width constraint, same pattern used by `viewBottomBar` today.
- **Cron TUI embedded inside switchboard** → The cron panel inside switchboard (`buildCronSection`) is a simplified embedded view, not the full `crontui`. Its hints must be added inline, not via `crontui` internals. Mitigation: document the distinction clearly; the full crontui and the embedded cron panel are separate rendering paths.

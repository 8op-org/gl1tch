## Context

ORCAI's UI is built on BubbleTea + Lipgloss. Themes are loaded from YAML bundles (`internal/themes`) and a hardcoded Dracula fallback lives in `internal/styles`. The switchboard TUI holds all modal rendering logic (`resolveModalColors`, `viewHelpModal`, `viewQuitConfirm`) as private methods, making those components impossible to reuse. Crontui, plugins, and future subcommands must each reinvent modal UI or go without it.

Panel headers currently generate a fixed-width ANSI bar using `▄`/`▀`/`█` block characters with left-aligned text. There is no support for CP437 decorative patterns, TDF block-letter title fonts, gradient borders, or user-editable label translations.

The theme registry lives only inside the switchboard model; no other package can observe theme changes or load the active bundle without being coupled to switchboard.

## Goals / Non-Goals

**Goals:**
- Single `internal/modal` package that any BubbleTea program can import to render confirm, alert, and scrollable-help overlays.
- Process-level `themes.GlobalRegistry` singleton with an observer pattern for theme-change events.
- CP437 pattern header engine with at least 8 named patterns (waves, texture, checkerboard, double-line, shadow, noise, brick, scanline).
- TDF font renderer that reads `.tdf` files from `internal/assets/fonts/tdf/` and renders a string as ANSI block letters at a target width.
- Gradient border renderer that interpolates between 2–4 hex color stops per side.
- `~/.config/orcai/translations.yaml` loader with raw ANSI escape support for all labeled strings in the UI.
- `welcome-dashboard` header updated to full-width centered layout using the new header engine.

**Non-Goals:**
- Full i18n / locale switching (translations are hacker overrides, not multi-language support).
- Dynamic ANSI art import from external URLs at runtime.
- Changing the BubbleTea program structure or message loop architecture.
- Plugin sandbox isolation for the modal API (plugins run in-process).

## Decisions

### 1. `internal/modal` as a pure rendering package (no BubbleTea model)

**Decision:** `internal/modal` exports only rendering functions (`RenderConfirm`, `RenderAlert`, `RenderScroll`) that take a `Config` and return a `string`. Each caller overlays the returned string in its own `View()` method.

**Why:** BubbleTea programs already own their event loop. Embedding a sub-model adds `Init`/`Update`/`View` plumbing that every importer must wire. A pure render function is trivially composable and testable without a BubbleTea runtime.

**Alternative considered:** A standalone `tea.Model` modal component. Rejected because BubbleTea doesn't have a standard component composition story — each program would still need to proxy messages manually.

---

### 2. Global registry singleton via `sync.Once`

**Decision:** `themes.GlobalRegistry()` returns a package-level `*Registry` initialized via `sync.Once`. Callers that want to observe theme changes call `registry.Subscribe(ch)`.

**Why:** The registry is effectively a process singleton (one active theme per process). A singleton avoids dependency injection boilerplate throughout every subcommand that just needs "the current theme." Explicit injection remains possible by passing `*Registry` directly where needed.

**Alternative considered:** Context-propagated registry. Rejected — BubbleTea models don't carry `context.Context` through `Update`, so injecting via context would require wrapping every message type.

---

### 3. CP437 pattern header as a character-tiling engine

**Decision:** Define a `PatternDef` as a string of CP437 characters that tile horizontally, plus a repeat strategy (tile/mirror). The header engine renders: top-pattern row (accent color) → title row (centered, styled with TDF or plain text) → bottom-pattern row (accent color mirrored). Character sequences stored as UTF-8 equivalents of the CP437 codepoints (rendered correctly by modern terminal emulators that speak CP437→Unicode mapping).

**Why:** Pure string rendering, no special terminal modes required. Works in tmux, iTerm2, and Kitty. Width adapts to exact panel width at render time.

**Alternative considered:** Actual ANSI/SAUCE .ans file sprites. Kept as optional override (existing `Headers` map) but patterns are the primary path because they auto-size.

---

### 4. TDF font renderer as a standalone engine

**Decision:** Implement a minimal TDF parser (`internal/tdf`) that reads the binary `.tdf` format (TheDraw Font format) and renders a string into ANSI escape sequences. Ship 3–5 embedded TDF fonts in `internal/assets/fonts/tdf/`. Theme YAML declares `header_font: "standard"` (or empty to disable).

**Why:** TDF files are compact (typically 2–8 KB each), the format is well-documented, and dozens of high-quality fonts are available under permissive terms from the ACiD/iCE packs. Shipping a native renderer keeps the binary self-contained with no C dependencies (no libcaca/figlet).

**Alternative considered:** FIGlet `.flf` fonts. More common, but FIGlet's format is ASCII-art only — no color. TDF supports color and is native to the BBS aesthetic. Both could coexist but TDF is the priority.

---

### 5. Gradient borders via `lipgloss.Border` with per-cell coloring

**Decision:** A gradient border is rendered by constructing the border string segment by segment and applying per-cell foreground escape codes that interpolate between 2–4 color stops. The result is a `string` (not a `lipgloss.Style`) that callers place around their panel content.

**Why:** Lipgloss does not natively support per-cell border coloring. Constructing the border string directly gives full control. The gradient interpolator can be shared with the pattern header engine.

**Alternative considered:** Rendering a colored shadow using layered `lipgloss.NewStyle().Border()` calls. Not viable — Lipgloss applies one border color per style, not per side or per character.

---

### 6. Translations as a YAML key→value map with ANSI passthrough

**Decision:** `translations.yaml` is a flat YAML map: `key: value` where value may contain raw ANSI escape sequences (written as `\e[...m` or `\033[...m`, expanded at load time). A `translations.Provider` interface exposes `T(key, fallback string) string`. The fallback is always the current hardcoded label.

**Why:** Flat key/value is the simplest override mechanism. Operators can target any label without understanding the Go internals. Raw ANSI support gives gradient text and 256-color labels without a DSL.

**Alternative considered:** Structured YAML with explicit `text:` and `color:` fields. More validated but more verbose and still requires knowledge of color codes.

## Risks / Trade-offs

- **CP437 rendering in non-UTF-8 terminals** → Most modern terminals (iTerm2, Kitty, tmux ≥3.2) map CP437 Unicode equivalents correctly. We document that CP437 patterns require a UTF-8 terminal. The existing plain-block fallback activates when pattern rendering is disabled.

- **TDF parser edge cases** → The TDF format has several font variants (color vs. mono, different block sizes). We implement the most common variant first; edge cases produce a plain-text fallback. A fuzzer test covers the parser.

- **Global registry race conditions** → `sync.Once` initialization is safe, but subscriber channels must be non-blocking (buffered) to avoid deadlocks if a subscriber is slow. Document this contract; provide a `SafeSubscribe` helper that wraps with a 1-element buffer.

- **Translations with malformed ANSI** → User-supplied escape sequences can break terminal state. We apply `\x1b[0m` (reset) after every translated string at render time.

- **Gradient borders + tmux** → tmux multiplexes terminal output and may coalesce adjacent escape sequences. We test gradient borders under `$TERM=screen-256color` and `tmux` in CI.

## Migration Plan

1. Add `internal/modal` with `RenderConfirm`, `RenderAlert`, `RenderScroll`.
2. Add `themes.GlobalRegistry()` singleton; initialize it in the switchboard bootstrap path (unchanged existing wiring).
3. Update switchboard modal methods to delegate to `internal/modal` — no behavior change, pure refactor.
4. Update crontui quit confirm to use `internal/modal`.
5. Add `internal/tdf` parser + embedded fonts.
6. Add CP437 pattern engine to `internal/styles`; update header rendering in switchboard and crontui.
7. Add gradient border renderer; update `theme.yaml` schema with optional `gradient_border` field.
8. Add `translations.Provider` and YAML loader; wire into switchboard and crontui at startup.
9. Update `welcome-dashboard` header to centered full-width layout.

Each step is independently deployable. Steps 1–4 are pure refactors with no UX change. Steps 5–9 add opt-in features controlled by theme YAML and translations YAML.

**Rollback:** All new features are opt-in via theme YAML flags (`header_pattern`, `header_font`, `gradient_border`). Themes that omit these fields use existing rendering paths unchanged. Translations file absence is a no-op.

## Open Questions

- Which TDF fonts to ship as defaults? Candidates: `fire` (iCE), `standard` (ACiD), `gothic`, `blocky`, `future`. Need license review.
- Should `translations.yaml` support per-theme overrides (e.g., different labels when Dracula is active vs. Nord)?
- Should the gradient border support alpha blending for tmux compatibility, or always use nearest 256-color approximation?

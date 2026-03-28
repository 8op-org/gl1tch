## Why

The current theme system is fragmented: styles are defined in `internal/styles`, palette structs in `internal/themes`, and modal rendering is private to the switchboard package — making it impossible to reuse modals, themed headers, or consistent UI components in crontui, go plugins, or future subcommands. Panel headers use simple block characters and the header title is left-aligned at a fixed offset, breaking the visual symmetry hackers expect from a true BBS aesthetic.

## What Changes

- **Shared modal component** — extract modal rendering from switchboard into a standalone `internal/modal` package; any orcai command (cron, plugin host, future subcommands) can render a quit-confirm, alert, or help modal without duplicating logic.
- **Theme registry global accessor** — expose `themes.Registry` through a process-level singleton so cron TUI, go plugins, and CLI adapters can load and observe the active theme without being wired through switchboard.
- **Centered full-width header titles** — panel headers recalculate title position at render time to center text across the exact terminal width; the header bar spans 100% of width so panel tops form a continuous horizontal line.
- **CP437/extended ASCII pattern headers** — header decorations use characters from the 128–255 (CP437) range: waves (`≈`), half-blocks, box-drawing heavy variants, and texture fills; each theme can declare a `header_pattern` choosing from a named pattern library.
- **TDF (TheDraw Font) renderer** — integrate a TDF font engine so panel titles can be rendered as full ANSI block-letter art using fonts from the TheDraw/ACiD font packs; fonts ship as embedded assets.
- **Gradient border rendering** — panel borders accept a `gradient` color list in `theme.yaml`; the renderer steps through the palette left-to-right (top/bottom) and top-to-bottom (sides), producing a smooth color sweep.
- **UI translations layer** — a `~/.config/orcai/translations.yaml` file lets operators remap any labeled string in the application (panel names, menu items, header titles, status bar tokens) with plain text or raw ANSI escape sequences including color codes; the translations layer sits above the theme, so it works with any theme.

## Capabilities

### New Capabilities

- `modal-system`: Shared reusable modal component (`internal/modal`) providing alert, confirm, and scrollable-content modals usable from any orcai TUI or subcommand.
- `ansi-pattern-headers`: CP437 pattern library + TDF font renderer for decorative panel headers; themes declare a pattern and optional TDF font; the engine renders at the exact panel width.
- `ui-translations`: Operator-configurable label/text overrides with raw ANSI support, loaded from `~/.config/orcai/translations.yaml` and surfaced via a `translations.Provider` interface.
- `gradient-borders`: Lipgloss-compatible gradient border renderer that interpolates between 2–4 hex stops per border side, declared in `theme.yaml`.
- `theme-global-registry`: Process-level theme registry singleton + observer pattern so any package can subscribe to theme changes without switchboard coupling.

### Modified Capabilities

- `welcome-dashboard`: Header title rendering updated to use centered full-width layout via new header engine.

## Impact

- `internal/themes` — adds `GradientBorder`, `HeaderPattern`, `TDFFont` fields to `Bundle`; adds `Registry` singleton accessor.
- `internal/modal` — new package; switchboard, crontui, plugin host import it.
- `internal/styles` — gradient border helpers added; pattern header renderer added.
- `internal/switchboard` — `resolveModalColors`, `viewHelpModal`, `viewQuitConfirm`, `viewThemePicker` delegated to `internal/modal`.
- `internal/crontui` — quit confirm modal uses `internal/modal`; theme bundle wired through global registry.
- Plugin host interface — plugins receive a `modal.API` and `translations.Provider` via the existing plugin context.
- New config file: `~/.config/orcai/translations.yaml`.
- New embedded assets: TDF fonts in `internal/assets/fonts/tdf/`.

## 1. Theme Global Registry

- [x] 1.1 Add `GlobalRegistry()` singleton function to `internal/themes/registry.go` using `sync.Once`
- [x] 1.2 Add `Subscribe(ch chan<- string)` and `Unsubscribe(ch chan<- string)` methods to `Registry` with non-blocking send
- [x] 1.3 Add `SafeSubscribe()` helper that wraps caller channel with a 1-element buffer
- [x] 1.4 Update main entrypoint to initialize `GlobalRegistry` with the loaded registry before any TUI starts
- [x] 1.5 Update switchboard to populate `GlobalRegistry` when it loads the theme registry
- [x] 1.6 Update crontui `New()` to use `GlobalRegistry().Active()` instead of requiring an explicit bundle argument
- [x] 1.7 Add crontui subscription to theme changes and re-render on new bundle
- [x] 1.8 Write unit tests for Subscribe/Unsubscribe/non-blocking send behavior

## 2. Shared Modal System

- [x] 2.1 Create `internal/modal` package with `Config` struct (`Bundle *themes.Bundle`, `Title`, `ConfirmLabel`, `DismissLabel string`, `TDFFont string`)
- [x] 2.2 Implement `RenderConfirm(cfg Config, w, h int) string` with Dracula fallback when Bundle is nil
- [x] 2.3 Implement `RenderAlert(cfg Config, message string, w, h int) string`
- [x] 2.4 Implement `RenderScroll(cfg Config, lines []string, offset, w, h int) string` with scroll indicators
- [x] 2.5 Refactor switchboard `viewQuitConfirm` to delegate to `modal.RenderConfirm`
- [x] 2.6 Refactor switchboard `viewHelpModal` to delegate to `modal.RenderScroll`
- [x] 2.7 Refactor switchboard `viewThemePicker` modal wrapper to use `internal/modal`
- [x] 2.8 Refactor switchboard `inbox_detail` modal wrapper to use `internal/modal`
- [x] 2.9 Remove `resolveModalColors()` from switchboard (logic moves to `internal/modal`)
- [x] 2.10 Update crontui quit confirm to use `modal.RenderConfirm`
- [x] 2.11 Define `modal.API` interface and expose it in the plugin host context
- [x] 2.12 Write render tests for all three modal variants (with and without bundle)

## 3. CP437 Pattern Header Engine

- [x] 3.1 Add `internal/styles/patterns.go` defining `PatternDef` struct and the 8 named pattern sequences (waves, texture, checkerboard, double-line, shadow, noise, brick, scanline)
- [x] 3.2 Implement `TilePattern(pattern PatternDef, width int) string` that tiles or truncates the character sequence to exactly `width` visible chars
- [x] 3.3 Implement `MirrorPattern(pattern PatternDef, width int) string` for bottom-row mirroring
- [x] 3.4 Update `internal/styles/styles.go` `DynamicHeader` (or equivalent) to use pattern tiling when `bundle.HeaderPattern` is set
- [x] 3.5 Add `HeaderPattern string` field to `themes.Bundle` and `themes.HeaderStyle`
- [x] 3.6 Update `theme.yaml` loader to parse `header_pattern` and `header_font` fields
- [x] 3.7 Write table tests covering all 8 patterns at widths 40, 80, 120, 200
- [x] 3.8 Update ABS and Dracula `theme.yaml` with example `header_pattern` values

## 4. TDF Font Renderer

- [x] 4.1 Create `internal/tdf` package with TDF binary format parser (header block + character blocks)
- [x] 4.2 Implement `Font.Render(text string, maxWidth int) (string, error)` returning ANSI block-letter art
- [x] 4.3 Add width-estimation function `Font.MeasureWidth(text string) int` for clamping logic
- [x] 4.4 Add fallback to plain text when rendered width exceeds panel width
- [x] 4.5 Add `internal/assets/fonts/tdf/` directory with 3–5 embedded TDF fonts (license-reviewed)
- [x] 4.6 Wire `embed.FS` for `internal/assets/fonts/tdf/` in `internal/assets/assets.go`
- [x] 4.7 Add `FontRegistry` in `internal/tdf` that maps font names to parsed `Font` instances using the embedded FS
- [x] 4.8 Add `HeaderFont string` field to `themes.Bundle`
- [x] 4.9 Integrate `tdf.FontRegistry.Render()` into the pattern header engine (step 3.4) when `HeaderFont` is set
- [x] 4.10 Write fuzz test for `internal/tdf` parser covering malformed/truncated input
- [x] 4.11 Verify license compatibility of selected TDF fonts; add attributions to `internal/assets/fonts/tdf/LICENSES.md`

## 5. Centered Full-Width Header Title

- [x] 5.1 Update switchboard main header render function to set width to terminal width (from `tea.WindowSizeMsg`)
- [x] 5.2 Center title text using `lipgloss.Style.Width(termWidth).Align(lipgloss.Center)` or equivalent manual padding
- [x] 5.3 Ensure all panel column left edges align with header left edge (remove any existing left indent offset)
- [x] 5.4 Update panel header generator to center title within each panel's exact column width
- [x] 5.5 Add resize test asserting header width equals terminal width at 80, 120, 200, 220 column widths

## 6. Gradient Border Renderer

- [x] 6.1 Add `GradientBorder []string` field to `themes.Bundle` (hex color stops, 0–4 values)
- [x] 6.2 Add `GradientBorder []string` to `themes.PanelHeaderStyle` for per-panel overrides
- [x] 6.3 Implement `internal/styles/gradient.go` with `InterpolateRGB(from, to lipgloss.Color, steps int) []lipgloss.Color`
- [x] 6.4 Implement `RenderGradientBorder(content string, stops []string, termType string) string` that constructs top/bottom/side segments with per-character color escapes
- [x] 6.5 Add 256-color fallback path when `$TERM` is `screen*` or `xterm-256color`
- [x] 6.6 Apply `\x1b[0m` reset after each gradient border side
- [x] 6.7 Wire gradient border renderer into switchboard panel rendering when `GradientBorder` is non-empty
- [x] 6.8 Wire panel-level gradient override from `PanelHeaderStyle.GradientBorder`
- [x] 6.9 Update ABS and Dracula `theme.yaml` with example gradient stop values
- [x] 6.10 Write gradient interpolation unit tests (black→white at 10 steps, 4-stop blending)

## 7. UI Translations Layer

- [x] 7.1 Create `internal/translations` package with `Provider` interface (`T(key, fallback string) string`)
- [x] 7.2 Implement `YAMLProvider` that loads `~/.config/orcai/translations.yaml` and expands `\e[...]m` / `\033[...]m` escape notation to raw bytes
- [x] 7.3 Implement `NopProvider` (all calls return fallback) used when translations file is absent
- [x] 7.4 Apply `\x1b[0m` reset after every translated string at render time
- [x] 7.5 Add `translations.GlobalProvider()` singleton (same pattern as `themes.GlobalRegistry()`)
- [x] 7.6 Define the full canonical key list as constants in `internal/translations/keys.go` (panel titles, modal titles, status bar tokens, header title)
- [x] 7.7 Replace hardcoded label strings in switchboard panel headers with `provider.T(key, fallback)` calls
- [x] 7.8 Replace hardcoded label strings in crontui panel headers with `provider.T(key, fallback)` calls
- [x] 7.9 Replace modal title strings in `internal/modal` `Config` with provider lookups at call sites
- [x] 7.10 Inject `translations.Provider` into plugin host context
- [x] 7.11 Write unit tests: known key returns value, unknown key returns fallback, missing file returns NopProvider, ANSI expansion correctness, malformed escape safety

## 8. Integration & Documentation

- [x] 8.1 Add example `translations.yaml` to `internal/assets/examples/translations.yaml` showing all available keys with ANSI color examples
- [x] 8.2 Update `internal/assets/themes/abs/theme.yaml` with `header_pattern`, `header_font`, and `gradient_border` showcase values
- [x] 8.3 Update `internal/assets/themes/dracula/theme.yaml` with example new fields
- [x] 8.4 Run `go vet ./...` and `go test ./...` and confirm all existing tests pass
- [ ] 8.5 Manually verify: switchboard renders gradient border + pattern header + TDF title in iTerm2 and under tmux
- [ ] 8.6 Manually verify: crontui quit confirm uses themed modal colors
- [ ] 8.7 Manually verify: editing `translations.yaml` and restarting orcai shows custom panel labels

## Context

Orcai currently hardcodes AI provider binaries in two places: `internal/discovery/discovery.go` (`knownCLITools`) and `internal/bridge/manager.go` (`adapterDefs`). Per-provider logic lives in `internal/adapters/{claude,gemini,copilot}/` as compiled Go code. The welcome splash is a baked-in BubbleTea component in `internal/welcome/`. There is no mechanism for contributors to add providers, widgets, or themes without modifying core and issuing a release.

The existing `internal/bus` package provides a pub/sub backbone. The existing `internal/plugin` package (CliAdapter, Manager) handles sidecar YAML wrappers and remains unchanged.

## Goals / Non-Goals

**Goals:**
- Replace all hardcoded provider lists with data-driven profile loading
- Define a contributor path for providers (YAML profile), widgets (manifest + binary), and themes (directory bundle) that requires zero changes to this repo
- Add a local Unix socket event bus so widgets can react to orcai state changes
- Ship first-party widgets (welcome, sysop panel) that use the same contract as contributor widgets
- Preserve all existing `internal/plugin` and `internal/pipeline` behavior

**Non-Goals:**
- A plugin marketplace or remote registry
- Hot-reload of plugins without restarting orcai
- Supporting non-Unix platforms (Windows) for the bus socket
- Changing the proto/gRPC definitions
- Replacing `internal/plugin` CliAdapter or Manager

## Decisions

### Decision 1: Provider profiles as embedded YAML, not Go structs

**Choice:** `//go:embed internal/assets/providers/*.yaml` — profile data lives outside Go code.

**Rationale:** Adding a new provider requires writing YAML, not Go. Bundled profiles stay auditable and diffable. Users can override any bundled profile by dropping a same-named file in `~/.config/orcai/providers/`. No recompile needed for users customizing an existing profile.

**Alternative considered:** Typed Go structs per provider (current adapter pattern). Rejected — every new provider is a repo PR.

### Decision 2: Separate manifest formats per plugin kind, not a single `type:` field

**Choice:** Providers, widgets, and themes each have their own schema and discovery directory.

**Rationale:** A contributor writing a weather widget should not need to understand provider profile fields. Separate schemas mean separate docs, separate validation errors, and a simpler mental model per contributor type. The daemon connecting them is invisible infrastructure.

**Alternative considered:** Single `plugin.yaml` with a `type:` discriminator. Rejected — one format means one set of docs; contributors hit irrelevant fields.

### Decision 3: Widget protocol is newline-delimited JSON over Unix socket, not gRPC

**Choice:** `internal/busd` serves a Unix socket; framing is `\n`-delimited JSON.

**Rationale:** Low barrier for contributors — a widget can be a bash script or a Python one-liner. No proto compilation step. The existing `internal/bus` package already handles pub/sub logic; `busd` is just a socket transport wrapping it. gRPC would require contributors to generate stubs and understand proto definitions.

**Alternative considered:** gRPC (extend existing bridge pattern). Rejected — existing bridge already proves complex for 3 providers; impractical for many lightweight widgets.

### Decision 4: Widgets run in tmux panes, orcai does not own layout

**Choice:** Orcai launches each widget binary via `tmux new-window` or `tmux split-window`; the user or the widget manifest controls positioning. Orcai does not compose a layout.

**Rationale:** tmux is already the layout manager. Fighting it adds complexity with no gain. Widgets that want specific layouts can use tmux commands themselves or document their preferred setup.

**Alternative considered:** Orcai managing a tiling layout (like a sidebar + main pane grid). Rejected — too opinionated, breaks user tmux configs.

### Decision 5: First-party widgets eat their own dogfood

**Choice:** `internal/welcome/` and the sysop panel are rewritten as widget binaries following the same manifest/protocol as contributor widgets.

**Rationale:** Validates the protocol design. Ensures first-party widgets are not privileged — if the contract works for core components, it works for contributors.

### Decision 6: Theme switching broadcasts full palette via bus event

**Choice:** `theme.changed` event payload includes the full resolved palette (all named colors as hex strings).

**Rationale:** Widgets should not need to read config files or know the active theme name. They receive a self-contained payload and re-render. This keeps widgets stateless with respect to theme.

## Risks / Trade-offs

- **Bus socket unavailable at widget startup** → Widgets MUST retry connection with backoff; orcai MUST start the socket before launching any widget binary.
- **Widget binary crashes or exits** → Orcai prunes the disconnected client from subscriber lists on next publish; no crash propagation.
- **Embedded profiles become stale** (model lists, pricing) → Profile updates ship with orcai releases; users can override via `~/.config/orcai/providers/` without waiting for a release.
- **Breaking the widget protocol** → Any change to event/command field names is a breaking change for contributor widgets. Protocol versioning (a `version` field in the handshake) is out of scope for v2 but should be added before the protocol is declared stable.
- **Theme bundle missing assets** → Orcai falls back to the bundled ABS theme; missing `splash.ans` is not fatal.

## Migration Plan

1. Add `internal/assets/` with embedded provider profiles and ABS theme bundle
2. Add `internal/providers/` package (profile loading, discovery, binary detection)
3. Add `internal/busd/` package (Unix socket daemon wrapping `internal/bus`)
4. Add `internal/widgets/` package (manifest loading, widget lifecycle)
5. Add `internal/themes/` package (bundle loading, palette resolution, theme switching)
6. Rewrite `internal/discovery/` to use provider profile registry (remove `knownCLITools`)
7. Rewrite `internal/bridge/manager.go` to use provider profile registry (remove `adapterDefs`)
8. Migrate `internal/welcome/` to first-party widget binary
9. Remove `internal/adapters/{claude,gemini,copilot}/` after bridge rewrite
10. Wire busd startup into orcai's main init sequence

Rollback: Steps 1–5 are additive. Cutover at steps 6–7 is the risk point; keep old discovery/bridge behind a build tag until profiles are validated.

## Open Questions

- Should widget manifests declare a minimum orcai version for protocol compatibility?
- Should provider profiles support multiple binary candidates (e.g., `claude` or `claude-code` depending on install method)?
- Should `~/.config/orcai/widgets/` support inline single-file manifests (no subdirectory) for simple one-binary widgets?

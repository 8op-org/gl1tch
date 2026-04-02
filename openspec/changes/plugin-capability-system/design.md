## Context

gl1tch discovers executor plugins via sidecar YAML files in `~/.config/glitch/wrappers/`. The `SidecarSchema` struct and `CliAdapter` handle execution. All UI integration (BUSD subscriptions, logo swaps, mode routing) is currently hardcoded in `switchboard.go` and `glitch_panel.go`. Any plugin that wants deep UI integration requires a gl1tch source change.

The goal is to let a sidecar YAML fully describe a plugin's integration contract — executor behaviour, UI mode, and event subscriptions — so gl1tch wires everything at startup from config alone.

## Goals / Non-Goals

**Goals:**
- `SidecarSchema` gains optional `mode` and `signals` blocks; zero-value safe for existing plugins
- A `WidgetRegistry` scans wrappers at startup and builds a dynamic dispatch table
- Slash command triggers, logo text, speaker labels, exit commands, and on-activate commands are all config-driven
- A named `SignalHandlerRegistry` maps handler names (`companion`, `score`, `log`) to gl1tch behaviour; plugins reference by name
- BUSD subscriptions are derived from loaded widget sidecars, not hardcoded
- gl1tch-mud sidecar ships as the canonical three-capability example

**Non-Goals:**
- Plugins rendering custom TUI panels directly (widget mode routes IO through the existing chat panel only)
- Hot-reload of sidecar config while gl1tch is running
- Plugin sandboxing or permission gating
- New external Go dependencies

## Decisions

### 1. Optional blocks, not a new `kind`

`kind` stays as the executor classification (`agent`, `tool`). `mode` and `signals` are independent optional blocks. This avoids combinatorial `kind` values (`agent+widget`, `tool+widget+signal`) and keeps backward compatibility trivial — a missing block is a zero-value struct.

**Alternative considered**: A `capabilities: [executor, widget, signal]` list. Rejected — more verbose, harder to validate, and the block presence already implies the capability.

### 2. Named signal handlers, not inline code

Plugins declare `handler: companion` not a shell command or script. gl1tch owns the handler implementations. This keeps the attack surface small (no arbitrary code execution from sidecar config) and makes handler behaviour consistent across all plugins that use the same name.

**Alternative considered**: `handler: run: "ollama run llama3.2"` inline shell. Rejected — security risk, hard to test, inconsistent UX.

### 3. Widget mode routes through the existing chat panel

A widget in `mode:` takes over the glitch chat panel's input routing — it does not get a new TUI panel. The `on_activate` command is piped to the plugin binary on stdin; subsequent user inputs are piped one-shot per command. Output renders as a `glitchSpeakerGame`-style entry labelled with `speaker`.

**Alternative considered**: Plugin owns a separate tmux pane. Rejected — defeats the "gl1tch transforms" UX and splits the companion commentary from the game output.

### 4. WidgetRegistry built at startup, stored on Model

`switchboard.New()` calls `LoadWidgetRegistry(wrappersDir)` which returns a `WidgetRegistry` containing all loaded widget configs. The Model stores it. BUSD subscription topics are derived from it at connect time (merged with the existing hardcoded topic list). Slash command dispatch checks it before the hardcoded switch.

**Alternative considered**: Reload registry on each `/mud`-style command. Rejected — adds latency and complexity for no benefit; wrappers dir is stable at runtime.

### 5. gl1tch-mud sidecar ships in gl1tch-mud repo, not gl1tch core

`~/.config/glitch/wrappers/gl1tch-mud.yaml` is installed by the gl1tch-mud package (e.g. `make install` copies it). gl1tch has no compile-time knowledge of gl1tch-mud. This is the correct plugin boundary.

## Risks / Trade-offs

- **Sidecar misconfiguration silently no-ops** → Mitigation: `LoadWidgetRegistry` logs warnings for invalid `mode`/`signals` blocks at DEBUG; unknown handler names log at WARN on first dispatch.
- **Topic wildcards (`mud.*`) already work in BUSD** → No risk; existing wildcard matching handles this.
- **Two widgets declare the same trigger** → First loaded wins; subsequent logs a WARN. Acceptable for v1.
- **On-activate binary not found** → `mudExecCmd` equivalent already handles this with a narration message; generalised version does the same.

## Migration Plan

1. Extend `SidecarSchema` and `CliAdapter` (non-breaking, additive)
2. Implement `WidgetRegistry` and `SignalHandlerRegistry` as new packages/files
3. Wire registry into `switchboard.New()` and BUSD subscription
4. Replace hardcoded `/mud` handler and `mud.*` BUSD subscription with registry-driven dispatch
5. Ship `gl1tch-mud.yaml` sidecar in gl1tch-mud repo
6. Remove in-progress hardcoded gl1tch-mud changes from `glitch_panel.go` and `switchboard.go`

No migration needed for existing users — wrappers without `mode`/`signals` are untouched.

## Open Questions

- Should `score` handler be wired now or deferred? (token score engine already exists; connecting it via signal config is low-effort)
- Should the `log` handler write to a debug panel or a file? (suggest file at `~/.local/share/glitch/plugin-signals.log` for v1)

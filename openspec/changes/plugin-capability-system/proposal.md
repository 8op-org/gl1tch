## Why

Plugins (sidecar YAML wrappers) currently declare only executor behaviour — they can be called as pipeline steps but cannot integrate with gl1tch's UI or event bus. Adding a new plugin (like gl1tch-mud) that wants to swap the logo, own a chat mode, and react to BUSD events requires hardcoding those behaviours in `switchboard.go`. This makes gl1tch non-extensible and forces every deep integration to become a core change.

## What Changes

- **Extend `SidecarSchema`** with two optional top-level blocks: `mode` (widget/UI takeover) and `signals` (BUSD topic subscriptions with named handlers).
- **Widget mode**: a plugin with a `mode:` block can declare a slash command trigger, a TDF logo string, a chat speaker label, an exit command, and an on-activate command. gl1tch wires this entirely from config — no source changes needed.
- **Signal subscriptions**: a plugin with a `signals:` block declares which BUSD topics it emits and which named handler gl1tch should invoke (e.g. `companion`, `score`). gl1tch subscribes at startup and dispatches to the registered handler.
- **Named signal handlers**: a small registry maps handler names to gl1tch behaviour (`companion` → Ollama narration, `score` → XP engine, `log` → debug log). New handlers can be added to gl1tch without touching individual plugin configs.
- **Backward compatibility**: existing sidecars with no `mode:` or `signals:` blocks are completely unaffected.
- **gl1tch-mud sidecar** becomes the canonical example of a plugin using all three capabilities (executor + widget + signal emitter).

## Capabilities

### New Capabilities

- `sidecar-mode-block`: Schema, parsing, and validation for the `mode:` block in a sidecar YAML — trigger command, logo text, speaker label, exit command, on-activate command.
- `sidecar-signals-block`: Schema, parsing, and validation for the `signals:` block — topic pattern, named handler, optional filter.
- `widget-registry`: Runtime registry in gl1tch that scans wrappers at startup, loads widget-capable sidecars, and registers their slash commands, logo configs, and BUSD subscriptions dynamically.
- `signal-handler-registry`: Named handler dispatch table (`companion`, `score`, `log`) that signal subscriptions reference by name.

### Modified Capabilities

- `cli-adapter-sidecar`: `SidecarSchema` struct and `NewCliAdapterFromSidecar` gain `Mode` and `Signals` fields. Existing behaviour unchanged; new fields are optional and zero-value safe.

## Impact

- `internal/executor/cli_adapter.go` — struct extension
- `internal/console/switchboard.go` — widget registry wired at startup; dynamic BUSD subscription; `glitchMudModeMsg` handling generalised to widget mode events
- `internal/console/glitch_panel.go` — `/mud` handler replaced by generic widget mode activation; `mudExecCmd` generalised
- `~/.config/glitch/wrappers/gl1tch-mud.yaml` — new sidecar file (ships with gl1tch-mud repo)
- No new external dependencies

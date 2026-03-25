## Why

Orcai's core hardcodes AI provider binaries (claude, gemini, copilot) in discovery and bridge layers, making every new provider a repo PR and recompile. There is no contributor path for adding providers, widgets, or themes without touching core code.

## What Changes

- **BREAKING**: Remove `knownCLITools` hardcoded slice from `internal/discovery/discovery.go`
- **BREAKING**: Remove `adapterDefs` hardcoded slice from `internal/bridge/manager.go`
- Remove per-provider adapter packages `internal/adapters/{claude,gemini,copilot}/`
- Add embedded capability profiles for popular providers (claude, gemini, opencode, aider, goose, copilot) as YAML data files via `//go:embed`
- Add user-installable provider profile discovery from `~/.config/orcai/providers/`
- Add widget plugin system: manifest-driven binaries in `~/.config/orcai/widgets/`, launched in tmux panes
- Add theme plugin system: directory bundles in `~/.config/orcai/themes/` covering palette, ANSI art, borders, status bar
- Add local event bus daemon (`internal/busd`) serving a Unix socket for widget ↔ orcai communication
- Migrate first-party components (welcome splash, sysop panel) to first-party widget binaries
- Move ANSI art assets into theme bundles

## Capabilities

### New Capabilities

- `provider-profiles`: Embedded + user-installable YAML capability profiles for AI provider binaries; replaces hardcoded discovery and bridge lists
- `widget-plugins`: Manifest-driven external widget binaries launched in tmux panes with framed JSON event protocol
- `theme-plugins`: Directory bundle themes covering palette, ANSI art, borders, and status bar; active theme broadcasts to widgets via bus
- `bus-daemon`: Local Unix socket event bus enabling bidirectional orcai ↔ widget communication

### Modified Capabilities

- `cli-adapter-discovery`: Provider binary detection moves from hardcoded list to profile-driven discovery; existing sidecar YAML path (`~/.config/orcai/plugins/`) is unchanged
- `welcome-dashboard`: Welcome splash migrates from baked-in BubbleTea component to first-party widget binary using the widget plugin contract

## Impact

- `internal/discovery/` — rewritten; `knownCLITools` removed
- `internal/bridge/` — rewritten or removed; `adapterDefs` removed
- `internal/adapters/` — removed (claude, gemini, copilot packages)
- `internal/welcome/` — migrated to widget binary
- `internal/ansiart/` — assets move into bundled theme
- New: `internal/busd/`, `internal/providers/`, `internal/widgets/`
- New embedded assets: `internal/assets/providers/*.yaml`, `internal/assets/themes/abs/`
- No changes to `internal/pipeline/`, `internal/plugin/` (CliAdapter, Manager), or proto definitions

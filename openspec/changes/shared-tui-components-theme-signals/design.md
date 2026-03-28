## Context

`busd` is orcai's Unix socket event bus. It already defines `theme.changed` as the canonical cross-process topic (`internal/themes.TopicThemeChanged`). However, no TUI component currently subscribes to busd for theme events.

The current cross-process theme propagation path in `crontui` is a 5-second file poll (`pollThemeFile` / `themeFilePollMsg`). This means switching themes in switchboard takes up to 5 seconds to appear in `cron` (or any other standalone sub-TUI). Every new sub-TUI would need to copy the same polling boilerplate.

The in-process path (via `GlobalRegistry.SafeSubscribe()`) works fine for components running inside the switchboard process but is unavailable to separate OS processes.

The `internal/panelrender` package already exposes shared header/sprite rendering. However the lower-level box drawing helpers (`boxTop`, `boxBot`, `boxRow`) are private wrappers inside `switchboard`, duplicated or absent in sub-TUIs.

## Goals / Non-Goals

**Goals:**
- Replace file polling in `crontui` (and establish the pattern for all future sub-TUIs) with a busd subscription that delivers `theme.changed` events in real time.
- Provide a reusable `tea.Cmd` + message type in a shared package so any BubbleTea model can subscribe to theme changes with a single call.
- Have switchboard publish `theme.changed` to busd when the user picks a new theme.
- Export `boxTop` / `boxBot` / `boxRow` helpers from `internal/panelrender` so sub-TUIs get box drawing without re-implementing it.

**Non-Goals:**
- Rewriting the in-process `GlobalRegistry.SafeSubscribe()` path — it still works within-process and remains the fast path for embedded TUIs.
- Moving all UI rendering code out of switchboard in this change — only shared primitives move.
- Replacing busd with a different IPC mechanism.
- Adding busd subscriptions for events other than `theme.changed` in this change.

## Decisions

### D1: busd as the cross-process theme channel

**Decision**: Use busd Unix socket for cross-process `theme.changed` delivery; remove the file-poll fallback once busd is wired.

**Rationale**: busd is already the declared mechanism for this event. File polling is inherently racy and slow. busd delivers events within the `sendDeadline` (5ms) of publication.

**Alternative considered**: fsnotify on the active-theme file — already used for cron config; familiar but still has ~1-5s debounce latency and every process must own its own watcher.

### D2: New `busd.ThemeSubscribeCmd` helper

**Decision**: Add a `ThemeSubscribeCmd(ctx context.Context) tea.Cmd` (or similar name) to a new `internal/tuikit` package (or directly to `internal/busd`) that:
1. Dials the busd socket.
2. Sends a registration frame subscribing to `theme.changed`.
3. Returns a `tea.Cmd` that blocks reading the next event and returns a typed `ThemeChangedMsg`.

Each TUI re-issues the cmd after handling a message (the standard BubbleTea listener pattern).

**Rationale**: Keeps the subscription wiring out of each TUI's Init/Update. A helper function is the minimal shared abstraction — no new interface types required.

**Alternative considered**: Exposing a channel-based API like `GlobalRegistry.SafeSubscribe()` but backed by busd — adds complexity and the channel pattern requires goroutine lifecycle management that `tea.Cmd` already handles.

### D3: Export box helpers from `panelrender`

**Decision**: Promote `boxTop`, `boxBot`, `boxRow` (currently private in `switchboard`) to exported functions `panelrender.BoxTop`, `panelrender.BoxBot`, `panelrender.BoxRow`. The existing private wrappers in switchboard become thin delegating calls (or are inlined away).

**Rationale**: `panelrender` already exists for this purpose; `crontui` uses `panelrender.BoxTop` in view code. The private switchboard functions are one-liners that just call the panelrender versions — they only exist because the exported versions didn't exist first.

**Alternative considered**: A new `internal/tuikit` package for all shared TUI primitives. Deferred — panelrender already has the right scope; adding another package for these 3 functions is premature.

### D4: switchboard publishes to busd on theme change

**Decision**: In the switchboard theme-picker apply path, after calling `registry.SetActive(name)`, call `busd.Publish(themes.TopicThemeChanged, themes.ThemeChangedPayload{Name: name})` on the daemon reference held by switchboard.

**Rationale**: Switchboard already owns the busd daemon lifecycle (started at boot, stopped on quit). Extending the existing publish call is a 2-line change.

**Alternative considered**: Having the registry itself publish to busd — couples themes package to busd, a bad dependency direction.

## Risks / Trade-offs

- **busd daemon not running** → `ThemeSubscribeCmd` must handle dial failure gracefully (return nil cmd, log a warning) so sub-TUIs degrade silently rather than crash.
- **Socket path changes** → `busd.SocketPath()` is the single source of truth; all callers use it.
- **File-poll removal breaks offline testing** → tests that mock theme switching via file writes will break. Mitigation: update tests to use the registry channel path or inject a mock busd.
- **boxTop/boxBot promotion is a small API surface change** → No external consumers; internal-only breakage is trivially fixed.

## Migration Plan

1. Export `BoxTop` / `BoxBot` / `BoxRow` in `panelrender`; update switchboard private wrappers to delegate.
2. Add `ThemeSubscribeCmd` to `internal/busd` (or new `internal/tuikit`).
3. Wire switchboard theme-picker to publish `theme.changed` via busd.
4. Replace `pollThemeFile` / `themeFilePollMsg` in `crontui` with the new busd subscription cmd.
5. Apply the same pattern to any other sub-TUI that currently polls (`jumpwindow`, etc.).
6. Delete the file-poll helpers once all consumers are migrated.

Rollback: The file-poll path can be reinstated by reverting crontui's Init/Update changes — no data migration involved.

## Open Questions

- Should `ThemeSubscribeCmd` live in `internal/busd` (coupling the event bus package to BubbleTea) or in a new thin `internal/tuikit` package? The busd package currently has no BubbleTea dependency; adding one is a minor but real coupling. Preference: `internal/tuikit` keeps concerns separated.
- Does `jumpwindow` also use file-poll? If so, it should be migrated in this same change.

## ADDED Requirements

### Requirement: Process-level theme registry singleton
The `themes` package SHALL expose a `GlobalRegistry()` function that returns a `*Registry` initialized exactly once via `sync.Once` for the process lifetime.

#### Scenario: Multiple callers receive the same registry instance
- **WHEN** `themes.GlobalRegistry()` is called from two different goroutines
- **THEN** both calls SHALL return the same `*Registry` pointer

#### Scenario: GlobalRegistry is safe to call before switchboard init
- **WHEN** a subcommand calls `themes.GlobalRegistry()` before the switchboard TUI starts
- **THEN** it SHALL receive a valid (possibly empty) registry without panicking

### Requirement: Theme change observer pattern
The `Registry` SHALL support subscribing to theme change events via a channel-based observer: `registry.Subscribe(ch chan<- string)` and `registry.Unsubscribe(ch chan<- string)`.

#### Scenario: Subscriber notified on theme change
- **WHEN** a subscriber channel is registered and the active theme changes
- **THEN** the new theme name SHALL be sent to the channel without blocking the caller

#### Scenario: Slow subscriber does not block theme change
- **WHEN** a subscriber's channel is full (unbuffered or backpressured)
- **THEN** the registry SHALL skip that subscriber's notification and continue (non-blocking send)

#### Scenario: Unsubscribed channel receives no further events
- **WHEN** a channel is unsubscribed and the theme changes again
- **THEN** no value SHALL be sent to the unsubscribed channel

### Requirement: Crontui uses GlobalRegistry for theme bundle
The crontui initialization SHALL obtain the active theme bundle from `themes.GlobalRegistry()` rather than requiring the bundle to be passed as a constructor argument.

#### Scenario: Crontui starts with no explicit bundle
- **WHEN** `crontui.New()` is called without a bundle argument
- **THEN** it SHALL query `themes.GlobalRegistry().Active()` to obtain the current bundle

#### Scenario: Crontui responds to theme change events
- **WHEN** the user switches themes in switchboard while crontui is running in a separate tmux pane
- **THEN** crontui SHALL subscribe to the registry and re-render using the new bundle colors on the next tick

### Requirement: Plugin host provides theme bundle via GlobalRegistry
The plugin host SHALL not pass the theme bundle as a constructor parameter; plugins SHALL call `themes.GlobalRegistry().Active()` to get the current bundle.

#### Scenario: Plugin reads active bundle
- **WHEN** a plugin calls `themes.GlobalRegistry().Active()`
- **THEN** it SHALL receive the same bundle as the switchboard is currently using

### Requirement: GlobalRegistry initialized before any TUI starts
The main binary entrypoint SHALL initialize `themes.GlobalRegistry()` with the loaded registry before starting any TUI or subcommand, ensuring the singleton is populated.

#### Scenario: Registry populated at entrypoint
- **WHEN** `orcai switchboard` or `orcai cron` is invoked
- **THEN** by the time any TUI model's `Init()` runs, `themes.GlobalRegistry().Active()` SHALL return a non-nil bundle

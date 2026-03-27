## MODIFIED Requirements

### Requirement: ExecutionContext is a named type with path-based accessors
The pipeline package SHALL expose an `ExecutionContext` struct that wraps a `map[string]any` with a `sync.RWMutex` for concurrent access. It SHALL provide `Get(path string) (any, bool)` and `Set(path string, value any)` methods. `Get` SHALL support dot-separated path expressions (e.g. `"step.fetch.data.url"`) by walking nested maps. The struct SHALL additionally hold an optional reference to the result store (`*store.Store`) accessible via `ec.DB()`, returning `nil` if no store is configured.

#### Scenario: Set and Get round-trip
- **WHEN** `ec.Set("param.env", "staging")` is called
- **THEN** `ec.Get("param.env")` returns `("staging", true)`

#### Scenario: Dot-path traversal
- **WHEN** the context contains `{"step": {"fetch": {"data": {"url": "https://x.com"}}}}`
- **THEN** `ec.Get("step.fetch.data.url")` returns `("https://x.com", true)`

#### Scenario: Missing path returns false
- **WHEN** `ec.Get("step.missing.data.key")` is called on a context without that path
- **THEN** it returns `(nil, false)`

#### Scenario: Snapshot returns a deep copy
- **WHEN** `ec.Snapshot()` is called
- **THEN** it returns a `map[string]any` that reflects the current state; subsequent `Set` calls do not modify the snapshot

#### Scenario: DB returns nil when store not configured
- **WHEN** `ec.DB()` is called on a context created without a store reference
- **THEN** it returns `nil`

#### Scenario: DB returns store when configured
- **WHEN** `ec.DB()` is called on a context created with `WithStore(s)`
- **THEN** it returns the `*store.Store` instance `s`

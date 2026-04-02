## 1. Extend SidecarSchema

- [x] 1.1 Add `ModeBlock` struct with `Trigger`, `Logo`, `Speaker`, `ExitCommand`, `OnActivate` fields and `IsZero()` method to `internal/executor/cli_adapter.go`
- [x] 1.2 Add `SignalDeclaration` struct with `Topic` and `Handler` fields to `internal/executor/cli_adapter.go`
- [x] 1.3 Add `Mode ModeBlock` and `Signals []SignalDeclaration` fields to `SidecarSchema`
- [x] 1.4 Write unit tests confirming zero-value safety for sidecars without `mode:`/`signals:` blocks

## 2. Implement WidgetRegistry

- [x] 2.1 Create `internal/console/widget_registry.go` with `WidgetConfig` and `WidgetRegistry` types
- [x] 2.2 Implement `LoadWidgetRegistry(wrappersDir string) *WidgetRegistry` — scans YAML files, skips zero-mode sidecars, warns on missing required fields and duplicate triggers
- [x] 2.3 Implement `WidgetRegistry.FindByTrigger(trigger string) *WidgetConfig`
- [x] 2.4 Implement `WidgetRegistry.AllSignalTopics() []string` — deduplicated topics from all loaded `signals:` blocks
- [x] 2.5 Write unit tests for `LoadWidgetRegistry`, `FindByTrigger`, and `AllSignalTopics`

## 3. Implement SignalHandlerRegistry

- [x] 3.1 Create `internal/console/signal_handlers.go` with `SignalHandlerRegistry` type (`map[string]func(topic, payload string)`)
- [x] 3.2 Implement `companion` handler — fires Ollama narration goroutine via `game.GameEngine.Respond()` with plugin system prompt, sends result to narration channel
- [x] 3.3 Implement `score` handler — parses token usage payload, forwards to scoring package
- [x] 3.4 Implement `log` handler — appends RFC3339 timestamp + topic + payload to `~/.local/share/glitch/plugin-signals.log`
- [x] 3.5 Implement `BuildSignalHandlerRegistry(narrationCh chan<- string) SignalHandlerRegistry` constructor that registers all built-in handlers
- [x] 3.6 Write unit tests for `log` handler and unknown handler drop behaviour

## 4. Wire Registry into Switchboard

- [x] 4.1 Call `LoadWidgetRegistry(wrappersDir)` in `switchboard.New()` and store on `Model` as `widgetRegistry`
- [x] 4.2 Call `BuildSignalHandlerRegistry(m.narrationCh)` in `switchboard.New()` and store as `signalHandlers`
- [x] 4.3 Replace hardcoded `"mud.*"` in `pipeline_bus.go` subscription list with `m.widgetRegistry.AllSignalTopics()` merged with existing hardcoded topics
- [x] 4.4 In BUSD event handler (`switchboard.go`), replace hardcoded `mud.*` companion dispatch with registry lookup: find all signal declarations matching the topic, invoke their handlers

## 5. Generalise Widget Mode in glitch_panel.go

- [x] 5.1 Replace hardcoded `mudMode bool` field with `activeWidget *WidgetConfig` on `glitchChatPanel`
- [x] 5.2 Replace hardcoded `/mud` slash command handler with generic `WidgetRegistry.FindByTrigger(cmd)` lookup
- [x] 5.3 Generalise `mudExecCmd` into `widgetExecCmd(cfg *WidgetConfig, input string) tea.Cmd` — uses `cfg.Mode.Speaker` for output label, reads binary path from sidecar
- [x] 5.4 Update `refreshTDFHeader()` (or equivalent) to render `activeWidget.Mode.Logo` when in widget mode, `"GL1TCH"` otherwise
- [x] 5.5 Generalise exit command check to use `activeWidget.Mode.ExitCommand` instead of hardcoded `"quit"`
- [x] 5.6 Generalise `on_activate` — pipe it to binary on widget activation if set
- [x] 5.7 Remove all remaining hardcoded `gl1tch-mud` / `mud` / `GIBSON` strings from `glitch_panel.go` and `switchboard.go`

## 6. gl1tch-mud Sidecar YAML

- [x] 6.1 Create `~/.config/glitch/wrappers/gl1tch-mud.yaml` (to be shipped with gl1tch-mud repo) with `kind: tool`, `mode:` block (`trigger: /mud`, `logo: THE GIBSON`, `speaker: GIBSON`, `exit_command: quit`, `on_activate: init`), and `signals:` block (`mud.* → companion`)
- [x] 6.2 Verify gl1tch loads the sidecar, registers `/mud` trigger, subscribes to `mud.*`, and wires companion handler at startup

## 7. Cleanup and Integration Tests

- [x] 7.1 Write integration test: load a test sidecar YAML with mode + signals, confirm `WidgetRegistry` and `SignalHandlerRegistry` are populated correctly
- [x] 7.2 Write routing test: `FindByTrigger` returns correct config; unknown trigger returns nil
- [x] 7.3 Remove any leftover hardcoded `glitchMudModeMsg`, `mudMode`, `mudExecCmd` symbols if replaced by generic equivalents
- [x] 7.4 Run `go build ./...` and `go test ./...` — all green

## 1. Extract shared components into internal/buildershared

- [x] 1.1 Create `internal/buildershared` package with `Sidebar` sub-model (searchable scrollable list, emits `SidebarSelectMsg`)
- [x] 1.2 Create `EditorPanel` sub-model in `internal/buildershared` (name input + multi-line content textarea, tab-cycles focus)
- [x] 1.3 Create `RunnerPanel` sub-model in `internal/buildershared` (streaming output display, `Clear()` and `InjectPrompt()` methods)
- [x] 1.4 Write table-driven unit tests for Sidebar filtering and selection messages
- [x] 1.5 Write unit tests for EditorPanel name/content accessors and tab focus cycle
- [x] 1.6 Write unit tests for RunnerPanel `Clear()` and line-append behavior

## 2. Refactor promptbuilder to use shared components

- [x] 2.1 Replace promptbuilder's inline sidebar rendering with `buildershared.Sidebar`
- [x] 2.2 Replace promptbuilder's editor area with `buildershared.EditorPanel`
- [x] 2.3 Replace promptbuilder's test runner output area with `buildershared.RunnerPanel`
- [x] 2.4 Add right-column feedback loop panel state to `BubbleModel` (draft content + response history)
- [x] 2.5 Wire agent runner chat input as inline widget at bottom of right column
- [x] 2.6 Implement feedback loop: first send clears draft, shows loading; subsequent sends append
- [x] 2.7 Bind `ctrl+s` to save current name + content from any focus position
- [x] 2.8 Bind `ctrl+r` to clear RunnerPanel and start test run with first-sent prompt
- [x] 2.9 Verify existing promptbuilder tests still pass

## 3. Refactor pipelineeditor to use shared components

- [x] 3.1 Replace pipelineeditor's left list panel with `buildershared.Sidebar` (field added to Model)
- [x] 3.2 Replace pipelineeditor's name input + prompt textarea with `buildershared.EditorPanel` (field added to Model)
- [x] 3.3 Replace pipelineeditor's runner output area with `buildershared.RunnerPanel` (field added to Model)
- [x] 3.4 Update right-column layout: EditorPanel (YAML content) above, agent runner chat input below
- [x] 3.5 Implement feedback loop in pipelineeditor (same behavior as promptbuilder)
- [x] 3.6 Bind `ctrl+s` to save pipeline YAML to disk and refresh sidebar
- [x] 3.7 Bind `ctrl+r` to inject first step's prompt into RunnerPanel and start test run
- [x] 3.8 Verify existing pipelineeditor tests still pass (no test files, build succeeds)

## 4. Add orcai prompt-builder subcommand

- [x] 4.1 Add `promptBuilderCmd` to `cmd/cmd_prompts.go` under root as `orcai prompt-builder`
- [x] 4.2 Wire provider/plugin setup (same as current `orcai prompts tui`) and launch refactored BubbleModel with alt-screen
- [x] 4.3 Add `promptBuilderStartCmd` — `orcai prompt-builder start` that opens in a new tmux window named `orcai-prompt-builder`

## 5. Add orcai pipeline-builder subcommand

- [x] 5.1 Add `pipelineBuilderCmd` to `cmd/pipeline.go` as `orcai pipeline-builder` subcommand
- [x] 5.2 Wire store, provider, plugin setup and launch refactored pipelineeditor Model with alt-screen
- [x] 5.3 Add `orcai pipeline-builder start` that opens in a new tmux window named `orcai-pipeline-builder`

## 6. Extract jump window to orcai widget jump-window

- [x] 6.1 Create `cmd/widget.go` with `widgetCmd` cobra group (`orcai widget`)
- [x] 6.2 Add `jumpWindowCmd` subcommand (`orcai widget jump-window`) that runs `jumpwindow.Run()` in non-embedded mode
- [x] 6.3 Register `widgetCmd` on `rootCmd` in `cmd/widget.go` `init()`

## 7. Update jump window sysop entries

- [x] 7.1 Update the prompts sysop entry in `internal/jumpwindow/jumpwindow.go` to open a new tmux window running `orcai prompt-builder`
- [x] 7.2 Update the pipelines sysop entry to open a new tmux window running `orcai pipeline-builder`
- [x] 7.3 Update or add `PipelinesMsg` / prompts message handlers in the parent switchboard if needed (switchboard's embedded PipelinesMsg handler is unchanged and correct; standalone mode now launches `orcai pipeline-builder` via tmux)

## 8. Integration and validation

- [x] 8.1 Build project (`go build ./...`) with zero errors (verified)
- [x] 8.2 Run unit tests for changed packages pass (buildershared, promptbuilder, cmd)
- [ ] 8.3 Manually verify `orcai prompt-builder` opens correct two-column TUI — requires manual terminal testing
- [ ] 8.4 Manually verify `orcai pipeline-builder` opens correct two-column TUI — requires manual terminal testing
- [ ] 8.5 Manually verify `orcai widget jump-window` opens standalone jump window — requires manual terminal testing
- [ ] 8.6 Manually verify jump window sysop entries open builder subcommands in new tmux windows — requires manual terminal testing

## 1. Dir Picker Overlay Component

- [x] 1.1 Create `internal/switchboard/dirpicker.go` with `DirPickerModel` struct, `Init()`, `Update()`, `View()`
- [x] 1.2 Implement async home directory walk (`tea.Cmd`) with depth limit of 3, streaming results as `dirWalkResultMsg`
- [x] 1.3 Implement inline fuzzy scorer (contiguous match length + position bonus) for directory name filtering
- [x] 1.4 Wire up arrow-key navigation, Enter to emit `DirSelectedMsg`, Esc to emit `DirCancelledMsg`
- [x] 1.5 Cap displayed results at 50; render overlay box with query input at top and filtered list below
- [x] 1.6 Write unit tests for fuzzy scorer and dir walk result filtering

## 2. Agent Runner Overlay — CWD Field

- [x] 2.1 Capture orcai launch directory via `os.Getwd()` at startup and store in switchboard `Model`
- [x] 2.2 Add `agentCWD string` field to switchboard `Model`; default to launch directory
- [x] 2.3 Add WORKING DIRECTORY section to `viewAgentModalBox()` showing current `agentCWD`
- [x] 2.4 Extend `agentModalFocus` cycle to include CWD field (PROVIDER → MODEL → PROMPT → CWD → PROVIDER)
- [x] 2.5 On Enter/trigger key when CWD focused, open dir picker overlay (`dirPickerOpen = true`, `dirPickerCtx = "agent"`)
- [x] 2.6 Handle `DirSelectedMsg` in switchboard: update `agentCWD` and close dir picker
- [x] 2.7 Pass `agentCWD` to `submitAgentJob()` so sessions launch in the selected directory
- [x] 2.8 Update help bar hint to include CWD field keybinding

## 3. New Session — Replace Full-Screen Picker

- [x] 3.1 Identify the keybinding/handler in switchboard that calls `picker.Run()` for new session
- [x] 3.2 Replace the `picker.Run()` call with `agentModalOpen = true` (opens agent runner overlay)
- [x] 3.3 Remove the full-screen `pickerModel` struct, `Run()`, `View()`, `Update()` from `internal/picker/picker.go`
- [x] 3.4 Retain `GetOrCreateWorktreeFrom`, `scanGitRepos`, `copyDotEnv`, `findGitRoot` as unexported or exported utility functions
- [x] 3.5 Ensure `submitAgentJob()` calls `GetOrCreateWorktreeFrom` with the selected CWD when applicable
- [x] 3.6 Clean up any remaining references to the deleted picker UI types

## 4. Pipeline Dir Picker Integration

- [x] 4.1 Add `pipelineDirPickerOpen bool` and `pendingPipelineID string` fields to switchboard `Model`
- [x] 4.2 Intercept the pipeline run trigger: instead of executing immediately, set `pipelineDirPickerOpen = true` and store the pending pipeline
- [x] 4.3 Render the dir picker overlay when `pipelineDirPickerOpen` is true (reuse `DirPickerModel`, `dirPickerCtx = "pipeline"`)
- [x] 4.4 Handle `DirSelectedMsg` for pipeline context: pass selected path as `cwd` to pipeline execution context, then start pipeline
- [x] 4.5 Handle `DirCancelledMsg` for pipeline context: clear pending pipeline, return to normal state
- [x] 4.6 Ensure pipeline step template interpolation resolves `{{cwd}}` from the execution context `cwd` key

## 5. Cleanup & Verification

- [x] 5.1 Remove or update any references to the old `picker.Run()` full-screen UI throughout the codebase
- [x] 5.2 Verify `go build ./...` passes with no errors
- [ ] 5.3 Manually test: open agent runner overlay, cycle focus to CWD, open dir picker, select a dir, submit agent
- [ ] 5.4 Manually test: trigger new session — confirm agent runner overlay opens (not old picker)
- [ ] 5.5 Manually test: trigger a pipeline run — confirm dir picker appears, selection passed to pipeline
- [ ] 5.6 Manually test: Esc from dir picker cancels without side effects in both agent and pipeline contexts

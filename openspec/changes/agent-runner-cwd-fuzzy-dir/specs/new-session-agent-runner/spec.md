## ADDED Requirements

### Requirement: New session keybinding opens the agent runner overlay
The new session keybinding (previously opening the full-screen picker) SHALL open the agent runner overlay directly. The agent runner overlay's PROVIDER, MODEL, PROMPT, and WORKING DIRECTORY fields serve as the complete new session configuration.

#### Scenario: New session keybinding triggers agent runner overlay
- **WHEN** the user triggers the new session keybinding
- **THEN** the agent runner overlay opens (not the legacy full-screen picker)

#### Scenario: Session launched from agent runner overlay
- **WHEN** the user fills in provider, model, prompt, and CWD then submits
- **THEN** a new tmux session is created with the specified configuration

### Requirement: Legacy full-screen picker UI is removed
The `picker.Run()` full-screen UI SHALL be removed. Its git/worktree utility functions (`GetOrCreateWorktreeFrom`, `scanGitRepos`, `copyDotEnv`) SHALL remain available as library functions.

#### Scenario: No full-screen picker displayed for new session
- **WHEN** a new session is initiated
- **THEN** the full-screen picker overlay is not shown; the agent runner overlay is shown instead

#### Scenario: Worktree utilities still usable
- **WHEN** a session is launched with a git repo as the CWD
- **THEN** `GetOrCreateWorktreeFrom` is still called to create an isolated worktree

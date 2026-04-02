---
title: "Console — The Switchboard"
description: "Navigate the GL1TCH Switchboard TUI, launch pipelines, run agents, and manage your workspace."
order: 2
---

GL1TCH runs inside a tmux session. The **Switchboard** (window 0) is your primary control panel — a full-screen BubbleTea TUI where you launch pipelines, run agents, check activity, and navigate your workspace. Everything streams live to the Activity Feed with status badges and timestamps.


## The Switchboard Layout

The Switchboard divides into three regions:

```
┌─────────────────────────────────────────────────────────────┐
│ GL1TCH DECK — CONTROL PANEL                         hh:mm:ss │
├──────────────────────┬────────────────────────────────────────┤
│                      │                                        │
│  PIPELINE LAUNCHER   │                                        │
│  ─────────────────   │      ACTIVITY FEED                     │
│  • backup.pipeline   │      (live output, results,            │
│  • deploy.pipeline   │       status badges)                   │
│  • sync.pipeline     │                                        │
│                      │                                        │
│  AGENT RUNNER        │                                        │
│  ─────────────────   │                                        │
│  Provider: ollama    │                                        │
│  Model: neural-chat  │                                        │
│                      │                                        │
│  Prompt: ___________ │                                        │
│  [Enter to send]     │                                        │
│                      │                                        │
├──────────────────────┴────────────────────────────────────────┤
│ [TAB] focus  [j/k] navigate  [↵] launch  [ESC] back  [T] theme│
└───────────────────────────────────────────────────────────────┘
```

**Left Sidebar**: Pipeline Launcher (saved `.pipeline.yaml` files) and Agent Runner (provider/model picker + inline prompt input).

**Center/Right**: Activity Feed — timestamped output from pipeline runs and agent invocations, with status badges (`[running]`, `[done]`, `[failed]`).

**Bottom Bar**: Compact keybinding reference showing the most important shortcuts.


## Navigation

### Focus Movement

| Key | Action |
|---|---|
| `tab` | Cycle focus between left sidebar sections and the Activity Feed |
| `shift+tab` | Cycle focus backwards |

### Selection

| Key | Action |
|---|---|
| `j` | Move selection down (in lists, left sidebar) |
| `k` | Move selection up (in lists, left sidebar) |
| `↵` (Enter) | Launch or open the selected item |
| `esc` | Close overlay / return to Switchboard |

### Visual Controls

| Key | Action |
|---|---|
| `T` | Open theme picker modal |


## Pipeline Launcher

The **Pipeline Launcher** scans `~/.config/glitch/pipelines/` for saved `.pipeline.yaml` files and displays them as a scrollable list in the left sidebar.

### Launch a Pipeline

1. Use `j`/`k` to navigate to the pipeline you want
2. Press `↵` to launch it
3. Output streams live to the Activity Feed
4. A status badge (`[running]` → `[done]` or `[failed]`) tracks progress
5. Each run is timestamped and added to the Inbox

### Active Job

While a pipeline is running, the launcher is disabled and shows a `[running]` badge. You can still navigate other parts of the Switchboard and watch the output stream in real time. When the job completes, the launcher re-enables and the badge updates.

> [!TIP]
> Multi-job queuing is planned for a future release. For now, wait for the current job to finish before launching another.


## Agent Runner

The **Agent Runner** is an inline agent invocation tool in the left sidebar. Pick a provider and model, type a prompt, and send it directly to your local Ollama or cloud provider.

### Anatomy

```
Provider: [ollama                    ↓]
Model:    [neural-chat              ↓]

Prompt:   ____________________________
          [Your prompt here]
          [↵ to send]
```

**Provider**: Dropdown showing installed sidecars and configured cloud providers. Managed via `apm.yml`.

**Model**: Dropdown populated from the selected provider's available models (refreshed at startup and on `R` key).

**Prompt Input**: Textarea for typing your message. Press `↵` to submit; output streams to the Activity Feed.

### Running an Agent

1. Navigate to the Agent Runner section (use `tab`)
2. Use `↓` to open the Provider dropdown and select a sidecar or cloud provider
3. Use `↓` again to open the Model dropdown and pick a model
4. Click in the Prompt field and type your message
5. Press `↵` to send
6. Output appears in the Activity Feed with a `[running]` badge
7. When the agent responds, the badge updates to `[done]` or `[failed]`

### Providers and Models

GL1TCH auto-detects installed Ollama sidescar via `~/.local/share/ollama/models` and registered cloud providers from `apm.yml`. Available models are pulled at startup.

To refresh the model list after installing a new sidecar, press `R` in the Switchboard. This re-calls the provider discovery without restarting the TUI.


## Activity Feed

The **Activity Feed** (center/right pane) displays real-time output from pipeline runs and agent invocations. Each entry is timestamped and tagged with its source.

### Entry Structure

```
[14:23:05] [running] pipeline:backup.yaml
  → Backing up database...
  → Compressing files...

[14:23:18] [done] agent:ask
  → Response from neural-chat:
  Here's what I think about that...
```

**Timestamp**: Shows when the activity started (not when output was written).

**Status Badge**: 
- `[running]` — job is active
- `[done]` — job completed successfully
- `[failed]` — job exited with non-zero status or error

**Source**: Either `pipeline:<name>` or `agent:<provider>:<model>`.

**Output**: Streamed line-by-line. Long pipelines show all stdout/stderr captured during execution.

### Selecting and Inspecting Results

The Activity Feed is scrollable. Navigate with `j`/`k` to move between entries. Press `↵` to open the full result in the **Inbox Detail Modal**.

### Live Streaming

Output is streamed in real time as it arrives from the pipeline or agent. The feed is throttled (50ms debounce) to avoid excessive BubbleTea re-renders on high-frequency output. Very long or complex pipelines may batch lines together.


## Chord Shortcuts

Chord shortcuts start with `^spc` (ctrl+space) followed by a key. This is the primary way to navigate your GL1TCH workspace beyond the Switchboard.

| Chord | Action |
|---|---|
| `^spc h` | Show help screen (full keybinding reference) |
| `^spc j` | Jump to any window (tmux window switcher) |
| `^spc c` | Create a new window |
| `^spc d` | Detach from the session (safely exit tmux) |
| `^spc r` | Reload GL1TCH (pick up a new binary without restarting) |
| `^spc q` | Quit GL1TCH entirely (close the session) |
| `^spc [` | Previous window in the session |
| `^spc ]` | Next window in the session |
| `^spc x` | Kill the current pane |
| `^spc X` | Kill the current window |
| `^spc a` | Jump to the GL1TCH assistant (shell/REPL pane) |
| `^spc t` | Open the theme picker |
| `^spc n` | New workspace session (alternative to `^spc c` for named sessions) |
| `^spc p` | Open the pipeline/prompt builder |

### Jump Window (`^spc j`)

`^spc j` opens a modal showing all active windows in your GL1TCH session. Use `j`/`k` to navigate, press `↵` to jump. Windows include:

- **switchboard** — This control panel
- **assistant** — Shell prompt where you can type commands directly
- **inbox** — Full-screen inbox modal (from `^spc j`, select "inbox")
- **themes** — Theme picker modal
- **prompts** — Prompt manager (view/edit saved prompts)
- Any custom windows you've created with `^spc c`


## Modal Workflows

Several features open as full-screen modals layered on top of the Switchboard. You can always press `esc` to close and return to the main view.

### Theme Picker (`T` or `^spc t`)

Opens a modal showing available themes. Use `j`/`k` to navigate, press `↵` to select. The theme applies immediately to all GL1TCH panels (Switchboard, Inbox, prompts, etc.) and persists across sessions.

Currently available themes:
- **Dracula** (default) — Dark mode with purples, cyans, and high contrast
- **Nord** — Cool, arctic colors
- **Gruvbox** — Warm, earthy tones

### Inbox Detail (`↵` from Activity Feed entry)

Selecting an entry in the Activity Feed and pressing `↵` opens the **Inbox Detail Modal**:

```
┌─────────────────────────────────────────────────┐
│ Pipeline: backup.yaml                      [12 of 47]│
│ Status: done | Exited: 0 | Runtime: 12.3s      │
├─────────────────────────────────────────────────┤
│                                                 │
│ STDOUT:                                         │
│ ─────────────────────────────────────────────── │
│ $ Starting backup...                            │
│ $ Database backed up to /data/backup.tar.gz    │
│ $ Compressing files...                          │
│ $ Done.                                         │
│                                                 │
│ [p]revious  [n]ext  [d]elete  [r]erun  [esc]back │
└─────────────────────────────────────────────────┘
```

**Navigation inside the modal**:
- `p` — Jump to the previous result in the Inbox
- `n` — Jump to the next result in the Inbox
- `d` — Delete the current result
- `r` — Re-run the pipeline or agent with the same parameters
- `esc` — Close and return to the Switchboard

### Prompt Builder (`^spc p`)

Opens a full-screen accordion-style editor for authoring and testing new pipelines or agent prompts. This is a separate view from the Switchboard and is useful for iterative development.

```
┌──────────────────────────────────────────────────┐
│ PROMPT BUILDER                                   │
├──────────────────────────────────────────────────┤
│ Title: [New Pipeline              ]              │
│                                                  │
│ ▼ Step 1: Input                                  │
│   Type: [user-input      ▼]                      │
│   Prompt: [Ask user...   ]                       │
│                                                  │
│ ▼ Step 2: Brain                                  │
│   Model: [gpt-4          ▼]                      │
│   System: [You are...    ]                       │
│                                                  │
│ ▼ Step 3: Output                                 │
│   Format: [markdown      ▼]                      │
│                                                  │
│ [save] [test] [esc] back                         │
└──────────────────────────────────────────────────┘
```

**Key actions**:
- `tab` — Move between sections
- `↵` — Expand/collapse a step
- `j`/`k` — Navigate between steps
- `space` — Expand/collapse the current step
- `[save]` button — Save your pipeline to `~/.config/glitch/pipelines/`
- `[test]` button — Run a test execution and show output inline
- `esc` — Return to the Switchboard

### Theme Persistence

The active theme is stored in `~/.config/glitch/state.json` and loaded automatically when you reattach to your GL1TCH session. All panels (Switchboard, modals, pane borders, text colors) respect the theme.


## Status Indicators

The Switchboard shows several status indicators to keep you aware of what's happening:

### Chat Panel Subtitle

The Switchboard window title and subtitle show:
- Current time (top right)
- Active job badge (`[1 running]` if pipelines are active)
- Theme indicator (brief name of the active theme)

### Exit Status

When a pipeline or agent finishes, the Activity Feed entry shows:
- `[done]` — Exited cleanly (status 0)
- `[failed]` — Non-zero exit code or error
- Exit code displayed in the Inbox Detail modal

### Provider/Model Availability

If a provider is unavailable (Ollama not running, cloud API key invalid), the Agent Runner shows a warning and disables the Model dropdown until the provider is reachable again.


## Tips and Patterns

### Quick Pipeline Testing

1. Navigate to your pipeline in the launcher
2. Press `↵` to run it
3. Watch output stream in the Activity Feed
4. If it fails, press `^spc p` to open the builder, make edits, save, and test again
5. When satisfied, the pipeline is already saved and ready for production

### Agent Iteration

1. Open the Agent Runner section (use `tab`)
2. Write your prompt in the textarea
3. Press `↵` to send
4. Read the response in the Activity Feed
5. Use `↵` on the activity entry to open the full result
6. Refine your prompt and press `↵` again
7. Responses are saved to the Inbox for reference

### Multi-Window Workflow

While a pipeline is running in the Switchboard:
1. Press `^spc c` to open a new window
2. Type commands directly in the shell (e.g., `tail -f /var/log/app.log`)
3. Press `^spc j` to jump back to the Switchboard
4. Check the Activity Feed for your pipeline's progress
5. When done, use `^spc ]` to cycle back through windows

### Saving and Re-running Results

Pipeline outputs are automatically saved to the **Inbox**. To re-run an old pipeline:
1. Open the Activity Feed in the Switchboard
2. Navigate to the result you want to repeat
3. Press `↵` to open the Inbox Detail modal
4. Press `r` to re-run with the same parameters
5. Output from the new run streams to the Activity Feed

## See Also

- [Pipelines](./pipelines/overview.md) — Author and structure pipeline YAML
- [Agents](./agents/overview.md) — Configure providers and models
- [Shortcuts](./keyboard.md) — Full keybinding reference
- [Themes](./themes.md) — Customize colors and appearance
- [Inbox](./inbox.md) — Manage pipeline results

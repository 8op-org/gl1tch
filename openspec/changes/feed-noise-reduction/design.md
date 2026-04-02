## Context

Narration in gl1tch flows through two paths:

1. **`glitchNarrationMsg`** — game scoring narration arrives on `narrationCh chan string`, drained by `waitForNarrationCmd`; delivered to `glitchChat.update()` from `deck.go` Update() at line 1311. Also fired directly for "→ pipeline started" at line 1286.
2. **`glitchRunEventMsg`** — fired when `topics.RunCompleted` or `topics.RunFailed` arrives in `pipeline_bus.go`, injected into `glitchChat.update()` at line 1300.

The `glitchChatPanel` already tracks `streaming bool` (LLM token stream active) and `routing bool` (async intent router running). These are the canonical "conversation active" signals. Both are unexported fields.

## Goals / Non-Goals

**Goals:**
- Drop unsolicited narration while a conversation is in-flight or within 30s of user input
- Throttle unsolicited narration to 1 per run completion; suppress in busy mode
- Direct responses to user messages are never gated

**Non-Goals:**
- Queuing suppressed narration for later delivery — just drop it
- Configurable thresholds via settings file (hardcoded constants are fine for now)
- Gating step-level feed badges (`StepStatusMsg`) — those are always shown

## Decisions

### D1: `IsActive() bool` on `glitchChatPanel`

Rather than exporting `streaming`/`routing` or reaching into them from `Model.Update()`, add a single method:

```go
func (p glitchChatPanel) IsActive() bool {
    return p.streaming || p.routing
}
```

This keeps the panel's internals encapsulated and gives `Model` a stable predicate.

### D2: `lastUserMsgAt` set in `Model.Update()`, not in `glitchChatPanel`

`Model.Update()` already intercepts all `tea.KeyMsg` and routes to `glitchChat`. The submit action (Enter key when panel is focused) is where we set `lastUserMsgAt`. Avoids threading time state through the panel.

*Alternative:* a new `glitchUserSubmitMsg` — adds indirection for no benefit here.

### D3: Two separate gates, composed

```go
func (m Model) conversationActive() bool {
    return m.glitchChat.IsActive() || time.Since(m.lastUserMsgAt) < 30*time.Second
}

func (m Model) narrationAllowed() bool {
    if m.conversationActive() {
        return false
    }
    // Busy mode: 2+ completions in last 60s → suppress
    if time.Since(m.runWindowStart) < 60*time.Second && m.recentRunCount >= 2 {
        return false
    }
    return true
}
```

`conversationActive()` is pure suppression — binary, no counting.
`narrationAllowed()` adds the rate-limit layer on top.

### D4: Run completion counting in the existing `pipeline_bus.go` event handler

`topics.RunCompleted` and `topics.RunFailed` already land in `deck.go Update()` at the bus event case. Increment `recentRunCount` there; reset window when `time.Since(runWindowStart) >= 60s`.

### D5: `lastUserMsgAt` detection point

The user submit path in `deck.go` runs when `glitchChat.focused` and Enter is pressed. Search for where `glitchChat.update(msg)` is called with a key message and the panel is focused — that's the set point. No need to change `glitch_panel.go` for this.

## Risks / Trade-offs

- **30s silence window feels wrong in practice** → easy to tune; extracting as a constant makes it a one-line change
- **Dropping narration vs queuing** → dropping is correct here; queued narration arriving after a long delay is more confusing than silence
- **Busy mode threshold (2 runs/60s)** → conservative; a single fast pipeline loop won't trigger it

## Migration Plan

1. Add `IsActive()` to `glitchChatPanel` in `glitch_panel.go`
2. Add five fields to `Model` in `deck.go`
3. Add `conversationActive()` and `narrationAllowed()` helpers to `deck.go`
4. Set `lastUserMsgAt` at the user submit point in `deck.go Update()`
5. Increment `recentRunCount` / reset window at `RunCompleted`/`RunFailed` events
6. Gate `glitchNarrationMsg` and `glitchRunEventMsg` cases in `Update()` with `narrationAllowed()`
7. `go build ./...` + `go vet ./...`

No rollback complexity — all changes are additive state fields and conditional guards.

## Open Questions

- Should `/quiet` toggle be a follow-up or in scope? (Recommendation: follow-up — the automatic suppression covers the core case)

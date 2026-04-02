## 1. Expose active state from glitchChatPanel

- [x] 1.1 Add `IsActive() bool` method to `glitchChatPanel` in `internal/console/glitch_panel.go`: returns `p.streaming || p.routing`

## 2. Add suppression state to Model

- [x] 2.1 Add `lastUserMsgAt time.Time` field to `Model` in `internal/console/deck.go`
- [x] 2.2 Add `recentRunCount int`, `runWindowStart time.Time` fields to `Model` for busy-mode tracking
- [x] 2.3 Add `conversationActive() bool` method to `Model`: returns `m.glitchChat.IsActive() || time.Since(m.lastUserMsgAt) < 30*time.Second`
- [x] 2.4 Add `narrationAllowed() bool` method to `Model`: returns false if `conversationActive()`; returns false if `recentRunCount >= 2 && time.Since(m.runWindowStart) < 60*time.Second`; otherwise true

## 3. Set lastUserMsgAt on user submit

- [x] 3.1 In `deck.go Update()`, find where the user submits input to the glitch panel (Enter key while `glitchChat.focused`) and set `m.lastUserMsgAt = time.Now()`

## 4. Track run completions for busy mode

- [x] 4.1 In `deck.go Update()`, in the bus event handler for `topics.RunCompleted` and `topics.RunFailed`: if `time.Since(m.runWindowStart) >= 60*time.Second`, reset `m.recentRunCount = 0` and `m.runWindowStart = time.Now()`; then increment `m.recentRunCount`

## 5. Gate unsolicited narration delivery

- [x] 5.1 In `deck.go Update()`, wrap the `glitchNarrationMsg` case: if `!m.narrationAllowed()`, return early (drop the message)
- [x] 5.2 In `deck.go Update()`, wrap the `glitchRunEventMsg` case: if `!m.narrationAllowed()`, return early (drop the message)
- [x] 5.3 In `deck.go Update()`, wrap the `glitchNarrationMsg` case at the pipeline-started narration point (line ~1286): same `narrationAllowed()` gate

## 6. Build gate

- [x] 6.1 Run `go build ./...` and `go vet ./...` — zero errors

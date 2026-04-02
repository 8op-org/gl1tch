## Why

gl1tch's feed fills up with ambient narration and flavor commentary during normal operation — pipeline lifecycle events, game scoring quips, "another user entered the network" announcements — even while you're actively typing or mid-conversation. The result is a noisy interface that trains the operator to ignore the feed, which defeats the purpose of having one. Two surgical fixes address the core cases: shutting up when a conversation is live, and throttling unsolicited quips when pipelines are busy.

## What Changes

- Add `conversationActive()` gate to `Model`: returns true when `glitchChat.streaming || glitchChat.routing || time.Since(lastUserMsgAt) < 30s`
- Track `lastUserMsgAt time.Time` on `Model` — set whenever the user submits input to the glitch panel
- Gate unsolicited `glitchNarrationMsg` and `glitchRunEventMsg` delivery in `deck.go Update()` — drop when conversation is active
- Add narration rate limiter to `Model`: max 1 unsolicited narration per completed run; suppress entirely in busy mode (2+ runs completed in last 60s)
- Expose `IsActive() bool` on `glitchChatPanel` so `Model` can read `streaming || routing` without accessing unexported fields directly

## Capabilities

### New Capabilities

- `conversation-suppression`: When a conversation thread is live (streaming, routing, or user sent a message <30s ago), unsolicited narration and run-event commentary are dropped
- `narration-rate-limit`: Unsolicited narration is capped at 1 per completed run; busy mode (2+ completions in 60s) suppresses all unsolicited narration

### Modified Capabilities

- `feed-step-output`: No spec-level behavior change — step output still appears; only unsolicited narration commentary is gated

## Impact

- `internal/console/glitch_panel.go` — add `IsActive() bool` method to `glitchChatPanel`
- `internal/console/deck.go` — add `lastUserMsgAt`, `narrationCount`, `narrationWindowStart`, `recentRunCount`, `runWindowStart` fields to `Model`; add `conversationActive()` and `narrationAllowed()` helpers; gate `glitchNarrationMsg` and `glitchRunEventMsg` in `Update()`
- No new dependencies; no schema or API changes

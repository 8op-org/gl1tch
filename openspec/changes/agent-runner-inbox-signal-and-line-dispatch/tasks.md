## 1. Bus Topics

- [x] 1.1 Add `AgentRunStarted`, `AgentRunCompleted`, `AgentRunFailed` constants to `internal/busd/topics/topics.go`

## 2. Agent Run Bus Publishing

- [x] 2.1 Verify `busd.Client` on the switchboard has a `Publish(topic string, payload any) error` method (add if missing)
- [x] 2.2 In switchboard `Update()`, on `jobStartedMsg` (or equivalent), publish `topics.AgentRunStarted` with `run_id` and `agent` fields
- [x] 2.3 On `jobDoneMsg`, publish `topics.AgentRunCompleted` with `run_id` and `exit_status: 0`
- [x] 2.4 On `jobFailedMsg`, publish `topics.AgentRunFailed` with `run_id` and `exit_status` (non-zero)
- [x] 2.5 Ensure publish errors are silently swallowed (log only) so job lifecycle is unaffected

## 3. Inbox Subscription

- [x] 3.1 In `internal/inbox/model.go`, add `"agent.run.*"` to the bus subscription wildcard list alongside `"pipeline.run.*"`
- [x] 3.2 On receipt of any `agent.run.*` bus event, trigger the same `refreshRuns()` path used for pipeline completion events
- [ ] 3.3 Verify inbox list updates in real time when an agent run completes (manual smoke test)

## 4. Activity Feed Line Dispatch

- [x] 4.1 Confirm `feedMarked` / `feedMarkedContent` are populated correctly when `m` is pressed on feed lines
- [x] 4.2 Wire `r` key in feed-focused state: collect marked line content in order, join with newline
- [x] 4.3 Open agent runner modal with pre-filled prompt set to the collected content
- [x] 4.4 Clear all marks from `feedMarked` / `feedMarkedContent` after the modal opens
- [x] 4.5 Verify `r` with no marks opens an empty agent modal (existing behavior preserved)

## 5. Visual Mark Highlight

- [x] 5.1 In `viewActivityFeed()`, apply a distinct style (e.g., marked background color) to lines whose absolute index is in `feedMarked`
- [x] 5.2 Ensure cursor-line highlight and mark highlight are visually distinct and composable (cursor on a marked line shows both)

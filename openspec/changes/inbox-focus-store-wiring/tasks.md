## 1. Switchboard store field

- [x] 1.1 Add `store *store.Store` field to `Model` struct in `internal/switchboard/switchboard.go`
- [x] 1.2 Set `m.store = s` in `NewWithStore(s *store.Store)`

## 2. Inbox bottom-bar hints

- [x] 2.1 Add `case m.inboxFocused:` branch in `viewBottomBar` before the default case
- [x] 2.2 Render hint row: `enter` open · `d` delete · `r` re-run · `tab` focus · `i` inbox (using existing hint helper functions)

## 3. Pipeline CLI store wiring

- [x] 3.1 In `cmd/pipeline.go` `pipelineRunCmd.RunE`, open the store with `store.Open()`; log and continue if it fails (do not abort)
- [x] 3.2 Defer `s.Close()` after successful open
- [x] 3.3 Pass `pipeline.WithRunStore(s)` to `pipeline.Run()`

## 4. Verification

- [x] 4.1 `go build ./...` — no errors
- [x] 4.2 Run `orcai pipeline run <any-yaml>` then `sqlite3 ~/.local/share/orcai/orcai.db 'select name,exit_status from runs order by id desc limit 5'` — confirm rows appear
- [ ] 4.3 Open switchboard, focus inbox (`i`), confirm bottom bar shows inbox hints not launcher hints

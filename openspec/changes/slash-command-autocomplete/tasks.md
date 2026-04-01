## 1. Data Model

- [x] 1.1 Define `slashSuggestion` struct `{ cmd, hint string }` and package-level `var glitchSlashCommands []slashSuggestion` adjacent to the slash-command switch in `glitch_panel.go`
- [x] 1.2 Add `acSuggestions []slashSuggestion`, `acCursor int`, and `acActive bool` fields to the `glitchChatPanel` struct

## 2. Filtering Logic

- [x] 2.1 Write `filterSuggestions(input string) []slashSuggestion` that strips the leading `/`, calls `fuzzyScore` from `internal/modal/fuzzypicker.go` against each command name, and returns ranked matches (score > 0); return all commands when input is exactly `/`
- [x] 2.2 Call `filterSuggestions` from the `Update` path whenever the textinput value changes and the value starts with `/`; set `acActive = len(results) > 0` and reset `acCursor = 0`

## 3. Key Handling

- [x] 3.1 In the `Update` method, when `acActive == true`, intercept `tab`, `up`, `down` to move `acCursor` (with wraparound) before forwarding to textinput
- [x] 3.2 When `acActive == true` and the user presses `Enter`, insert `acSuggestions[acCursor].cmd + " "` via `p.input.SetValue` + `p.input.CursorEnd()`, set `acActive = false`, and return without sending to chat backend
- [x] 3.3 When `acActive == true` and the user presses `Esc`, set `acActive = false` and mark an `acSuppressed bool` flag so the overlay doesn't reopen until the next input change
- [x] 3.4 Clear `acSuppressed` whenever the input value changes (on any non-navigation keystroke)

## 4. Rendering

- [x] 4.1 Write `viewSuggestions() string` that renders the suggestion list as a lipgloss box using the existing palette and overlay style (consistent with `modelPickerBox`)
- [x] 4.2 Each suggestion row renders as `  /cmd   hint text  ` with the selected row highlighted using `styles.Selected` or the theme accent colour
- [x] 4.3 Truncate hint text with `…` when the row would exceed `p.width`
- [x] 4.4 Call `viewSuggestions()` in the panel's `View()` method and position it directly above the input row (insert it into the vertical stack before the input line)

## 5. Tests

- [x] 5.1 Unit-test `filterSuggestions`: exact prefix, partial match, no match, bare `/` returns all
- [x] 5.2 BubbleTea model test: send `/` keypress → assert `acActive == true` and `acSuggestions` is full list
- [x] 5.3 BubbleTea model test: send `tab` when overlay active → assert `acCursor` advances and wraps
- [x] 5.4 BubbleTea model test: send `enter` when suggestion selected → assert input value equals `cmd + " "` and `acActive == false`
- [x] 5.5 BubbleTea model test: send `esc` when overlay active → assert `acActive == false` and input value unchanged

## TDF Parser Fix Tasks

- [x] 1. Fix magic constant: `"\x13TheDraw FONTS file\x1a"` (lowercase "file", 20 bytes)
- [x] 2. Skip `55 AA 00 FF` identification marker (4 bytes after magic)
- [x] 3. Parse name as 1-byte length prefix + 12-byte buffer + 4-byte reserved
- [x] 4. Read 94 character offset pointers for ASCII 33–126 (not 95 for 32–126)
- [x] 5. Replace fixed-block character storage with `map[byte][][]cell` (variable-length rows)
- [x] 6. `parseCharRows`: reads CR-terminated rows, NUL-ended character, 2-byte cells for Color fonts
- [x] 7. `Font.Height()`: returns max row count across all defined characters
- [x] 8. `Font.Render()`: returns multi-line `\n`-joined string; falls back to plain text if too wide or no data
- [x] 9. Replace `placeholder.tdf` with `fire.tdf`, `tiny.tdf`, `elite.tdf` from the roysac TDF collection
- [x] 10. Update `DynamicHeader` in `ansi_render.go` to expand to N+2 rows for multi-line TDF art
- [x] 11. Update `abs/theme.yaml` → `header_font: fire`, `dracula/theme.yaml` → `header_font: tiny`
- [ ] 12. Verify build: `go build ./...`
- [ ] 13. Manual visual verification: panel headers show TDF block-letter titles in iTerm2
- [ ] 14. Manual verification: header height expands correctly, no layout overflow

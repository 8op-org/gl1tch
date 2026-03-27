## Why

The TDF (TheDraw Font) parser introduced in `bbs-theme-system` produces no output because it checks for `"\x13TheDraw FONTS FILE"` (uppercase) while every real TDF file contains `"\x13TheDraw FONTS file"` (lowercase "file"). Additionally the parser assumed fixed-size character blocks, whereas the actual format uses variable-length rows with CR/NUL terminators and a 94-entry offset table for ASCII 33–126 — not 32–126. The `placeholder.tdf` embedded asset also used the wrong magic, preventing any font from loading.

## What Changes

- Fix magic constant to lowercase: `"\x13TheDraw FONTS file\x1a"`
- Skip the 4-byte `55 AA 00 FF` identification marker after magic
- Parse name as 1-byte length prefix + 12-byte buffer + 4-byte reserved gap
- Read 94 character offset pointers (ASCII 33–126), not 95
- Parse variable-length character rows (CR-terminated, NUL-ended) instead of fixed-size blocks
- Support all three font types: Color (2 bytes/cell), Block (1 byte/cell), Outline (1 byte/cell)
- Replace `placeholder.tdf` with real embedded fonts: `fire.tdf` (860 B), `tiny.tdf` (855 B), `elite.tdf` (1093 B)
- Update `DynamicHeader` to expand panel headers to N+2 rows when TDF art is multi-line
- Update ABS theme to use `header_font: fire`, Dracula to use `header_font: tiny`

## Capabilities

### New Capabilities

- `tdf-rendering`: Panel headers render multi-row TDF block-letter art from embedded fonts

### Modified Capabilities

- `bbs-theme-system`: Parser correctness — existing spec behavior unchanged but now actually works

## Impact

- `internal/tdf/tdf.go`: complete rewrite of `Parse()`, `Render()`, cell representation
- `internal/assets/fonts/tdf/`: remove placeholder, add fire/tiny/elite
- `internal/switchboard/ansi_render.go`: `DynamicHeader` supports multi-row TDF output
- `internal/assets/themes/*/theme.yaml`: font names updated

# TDF Rendering

## Overview

Panel headers render multi-row ANSI block-letter art from embedded TDF (TheDraw Font) files when `header_font` is set in the active theme.

## Requirements

### Parser

- **REQ-TDF-1**: `Parse()` MUST accept files whose magic is exactly `"\x13TheDraw FONTS file\x1a"` (20 bytes, lowercase "file").
- **REQ-TDF-2**: `Parse()` MUST skip the 4-byte `55 AA 00 FF` marker after the magic.
- **REQ-TDF-3**: Name is read as 1-byte length prefix + 12-byte null-padded buffer; length > 12 is clamped.
- **REQ-TDF-4**: 4 reserved bytes are skipped between name buffer and type byte.
- **REQ-TDF-5**: Offset table contains exactly 94 LE 2-byte entries for ASCII 33–126; `0xFFFF` = undefined.
- **REQ-TDF-6**: Character data is variable-length: Color fonts use 2-byte (char, attr) cell pairs; Block/Outline use 1 byte; `0x0D` ends a row; `0x00` ends the character.

### Rendering

- **REQ-TDF-7**: `Font.Render(text, maxWidth)` returns a multi-line string (rows joined with `\n`) when the font has character data and the rendered width ≤ maxWidth.
- **REQ-TDF-8**: Falls back to plain text when: no char data, or rendered width > maxWidth.
- **REQ-TDF-9**: Color attributes use the 16-color DOS palette mapped to 24-bit ANSI escapes.
- **REQ-TDF-10**: CP437 glyph bytes are converted to UTF-8 equivalents (block chars, line-drawing, etc.).
- **REQ-TDF-11**: All characters in a rendered string share the same row height (`Font.Height()`); shorter characters are padded with spaces.

### Panel Header Integration

- **REQ-TDF-12**: When `DynamicHeader` receives multi-line TDF output, the returned `[]string` is `[topLine, ...tdfLines..., botLine]` (height expands dynamically).
- **REQ-TDF-13**: Each TDF line is centered within the panel width using ANSI-aware visible-width calculation.

### Embedded Assets

- **REQ-TDF-14**: The embedded FS at `internal/assets/fonts/tdf/` MUST contain at least one valid TDF file with the correct magic.
- **REQ-TDF-15**: Font registry loads files by lowercase stem name; theme references match that name.

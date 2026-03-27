## TDF Parser — Binary Format

A TDF file has a fixed 233-byte header followed by variable-length character data:

| Offset | Length | Field |
|--------|--------|-------|
| 0 | 20 | Magic: `\x13TheDraw FONTS file\x1a` |
| 20 | 4 | ID marker: `55 AA 00 FF` |
| 24 | 1 | Name length (0–12) |
| 25 | 12 | Name buffer (null-padded) |
| 37 | 4 | Reserved (zeros) |
| 41 | 1 | Font type (0=Outline, 1=Block, 2=Color) |
| 42 | 1 | Letter spacing |
| 43 | 2 | Block size LE (total char data bytes) |
| 45 | 188 | Offset table: 94 × 2-byte LE entries for ASCII 33–126; `0xFFFF` = undefined |
| 233 | N | Character data |

Character data uses variable-length encoding per character:
- **Color fonts**: pairs of (CP437 glyph byte, color attribute byte)
- **Block/Outline fonts**: single glyph bytes
- Row ends with `0x0D` (CR)
- Character ends with `0x00` (NUL, also terminates the last row)

## Cell struct

Replace the old `map[byte][]byte` (flat bytes) with `map[byte][][]cell` where:

```go
type cell struct { char, attr byte }
```

This makes row iteration direct and removes the height/charWidth arithmetic that was
fragile under the old fixed-block assumption.

## Height and width

- `Font.Height()`: iterates all defined chars, returns max row count.
- `Font.charWidth(ascii)`: max cell count across that char's rows.
- `Font.MeasureWidth(text)`: sum of charWidths + spacing.

## DynamicHeader multi-row expansion

When `bundle.HeaderFont` is set and the render result contains `\n`:
1. Split on `\n` → `tdfLines`
2. Return `[]string{topLine} + centeredTDFLines + []string{botLine}`
3. Centering uses `visibleWidth()` (ANSI-aware) to compute padding.

## Embedded fonts

Three real TDF fonts replace `placeholder.tdf`:
- `fire.tdf` (860 B) — Color font, 8-row block letters
- `tiny.tdf` (855 B) — Color font, compact 3-row block letters
- `elite.tdf` (1093 B) — Color font, classic BBS style

Registry loads by filename stem → theme references by same name (e.g. `header_font: fire`).

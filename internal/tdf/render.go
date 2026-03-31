package tdf

import (
	"fmt"
	"strings"
)

var (
	fgANSI = [16]uint8{30, 34, 32, 36, 31, 35, 33, 37, 90, 94, 92, 96, 91, 95, 93, 97}
	bgANSI = [8]uint8{40, 44, 42, 46, 41, 45, 43, 47}
)

func colorEscape(color byte) string {
	fg := color & 0x0f
	bg := (color & 0xf0) >> 4
	if bg >= 8 {
		bg = 0
	}
	return fmt.Sprintf("\x1b[%d;%dm", fgANSI[fg], bgANSI[bg])
}

// RenderString renders text using font f, returning one ANSI-colored string per row.
func RenderString(text string, f *Font) []string {
	// Collect glyphs for each rune in text (skip missing).
	type entry struct {
		g   *Glyph
		sep int // spacing columns after this glyph
	}
	var glyphs []entry
	runes := []rune(text)
	for i, r := range runes {
		g, ok := f.Glyph(r)
		if !ok {
			continue
		}
		sep := 0
		if i < len(runes)-1 {
			sep = int(f.Spacing)
		}
		glyphs = append(glyphs, entry{g, sep})
	}

	lines := make([]string, f.Height)
	for row := range int(f.Height) {
		var sb strings.Builder
		for _, e := range glyphs {
			g := e.g
			var lastColor byte = 255 // invalid sentinel
			for col := range int(g.Width) {
				idx := row*int(g.Width) + col
				if idx >= len(g.Cells) {
					sb.WriteRune(' ')
					continue
				}
				cell := g.Cells[idx]
				if cell.Color != lastColor {
					sb.WriteString(colorEscape(cell.Color))
					lastColor = cell.Color
				}
				sb.WriteRune(cell.Ch)
			}
			sb.WriteString("\x1b[0m")
			// spacing
			for range e.sep {
				sb.WriteRune(' ')
			}
		}
		lines[row] = sb.String()
	}
	return lines
}

// StripANSI removes ANSI escape sequences from s. Exported for tests.
func StripANSI(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// skip until 'm' or end
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

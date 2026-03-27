package switchboard

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// readmeContent returns the README.md content, falling back to inline text.
func readmeContent() string {
	// Try reading from the binary's directory, then repo root.
	self, _ := os.Executable()
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}
	for _, candidate := range []string{
		filepath.Join(filepath.Dir(self), "README.md"),
		"README.md",
	} {
		if data, err := os.ReadFile(candidate); err == nil {
			return string(data)
		}
	}
	return fallbackReadme
}

// viewHelpModal renders the getting-started guide as a centered overlay.
func (m Model) viewHelpModal(w, h int) string {
	mc := m.resolveModalColors()

	innerW := min(w-8, 80)
	if innerW < 40 {
		innerW = 40
	}
	outerW := innerW + 2

	headerStyle := lipgloss.NewStyle().
		Background(mc.titleBG).
		Foreground(mc.titleFG).
		Bold(true).
		Width(innerW).
		Padding(0, 1)

	// Render markdown with glamour.
	rendered := renderMarkdown(readmeContent(), innerW)

	// Split into lines and apply scroll.
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	visibleH := h - 6 // header + border + some padding
	if visibleH < 4 {
		visibleH = 4
	}
	offset := m.helpScrollOffset
	if offset > len(lines)-visibleH {
		offset = max(len(lines)-visibleH, 0)
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + visibleH
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[offset:end]

	// Scroll indicator.
	total := len(lines)
	scrollInfo := ""
	if total > visibleH {
		scrollInfo = lipgloss.NewStyle().Foreground(mc.dim).
			Render(strings.Repeat(" ", innerW-12) +
				lipgloss.NewStyle().Foreground(mc.accent).Render("j/k") +
				lipgloss.NewStyle().Foreground(mc.dim).Render(" scroll  esc close"))
	} else {
		scrollInfo = lipgloss.NewStyle().Foreground(mc.dim).
			Width(innerW).Align(lipgloss.Right).Padding(0, 1).
			Render("esc close")
	}

	body := lipgloss.NewStyle().
		Width(innerW).
		Padding(0, 1).
		Render(strings.Join(visible, "\n"))

	content := strings.Join([]string{
		headerStyle.Render("ORCAI  getting started"),
		body,
		scrollInfo,
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mc.border).
		Width(outerW).
		Render(content)
}

// renderMarkdown renders md using glamour, falling back to plain text on error.
func renderMarkdown(md string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return out
}

const fallbackReadme = `# ORCAI — Getting Started

Press **^spc** (ctrl+space) to access chord shortcuts.

**Navigation:** tab · j/k · enter

**Panels:** Pipelines · Agent Runner · Signal Board · Activity Feed

**Chord shortcuts:**
- ^spc h  this help
- ^spc q  quit
- ^spc d  detach
- ^spc r  reload
- ^spc m  themes
- ^spc j  jump to window
`

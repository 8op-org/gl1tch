package crontui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/adam-stokes/orcai/internal/cron"
)

// Lipgloss styles (built lazily since terminal detection is best-effort).
var (
	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(draculaComment))

	styleHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaPurple)).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color(draculaCurrent)).
			Foreground(lipgloss.Color(draculaPurple)).
			Bold(true)

	styleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaFg))

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaComment))

	styleGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaGreen))

	styleRed = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaRed))

	styleCyan = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaCyan))

	styleOrange = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaOrange))

	styleOverlay = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(draculaPink)).
			Background(lipgloss.Color(draculaBg)).
			Padding(1, 2)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color(draculaRed))
)

// View renders the full TUI screen.
func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	// Split height: 60% jobs, 40% logs, minus 1 row for the hint bar.
	topH, botH := splitHeight(m.height-1, 0.6, 6)

	top := m.viewJobList(m.width, topH)
	bot := m.viewLogPane(m.width, botH)
	bar := m.viewHintBar(m.width)

	content := lipgloss.JoinVertical(lipgloss.Left, top, bot, bar)

	// Render overlays on top if open.
	if m.editOverlay != nil {
		return renderOverlay(content, m.viewEditOverlay(), m.width, m.height)
	}
	if m.deleteConfirm != nil {
		return renderOverlay(content, m.viewDeleteConfirm(), m.width, m.height)
	}
	return content
}

// viewJobList renders the top pane showing the list of cron entries.
func (m Model) viewJobList(width, height int) string {
	// Inner width accounts for border (2) and padding.
	inner := width - 4
	if inner < 10 {
		inner = 10
	}

	// Build header line.
	var headerRight string
	if m.filtering {
		headerRight = " " + m.filterInput.View()
	}
	title := styleHeader.Render("CRON JOBS")
	header := title + headerRight
	header = padRight(header, inner)

	// Build rows.
	var rows []string
	if len(m.filtered) == 0 {
		if m.filterInput.Value() != "" {
			rows = append(rows, styleDim.Render("  no matches"))
		} else {
			rows = append(rows, styleDim.Render("  no scheduled jobs"))
		}
	} else {
		// Determine visible window.
		visibleRows := height - 4 // header + border top + border bot + header row
		if visibleRows < 1 {
			visibleRows = 1
		}
		m.clampScrollForList(len(m.filtered), visibleRows)

		start := m.scrollOffset
		end := start + visibleRows
		if end > len(m.filtered) {
			end = len(m.filtered)
		}

		for i := start; i < end; i++ {
			e := m.filtered[i]
			row := m.formatEntryRow(e, inner)
			if i == m.selectedIdx {
				row = styleSelected.Width(inner).Render(row)
			} else {
				row = styleNormal.Render(row)
			}
			rows = append(rows, row)
		}
	}

	// Assemble pane content.
	lines := []string{header}
	lines = append(lines, rows...)

	// Pad to fill height (accounting for borders).
	contentH := height - 2 // 2 for borders
	for len(lines) < contentH {
		lines = append(lines, "")
	}

	body := strings.Join(lines, "\n")
	paneStyle := styleBorder.Width(width - 2).Height(height - 2)
	if m.activePane == 0 {
		paneStyle = paneStyle.BorderForeground(lipgloss.Color(draculaPurple))
	}
	return paneStyle.Render(body)
}

// formatEntryRow formats a single entry as a fixed-width row string.
func (m Model) formatEntryRow(e cron.Entry, width int) string {
	nextStr := ""
	if t, err := cron.NextRun(e); err == nil {
		nextStr = cron.FormatRelative(t)
	} else {
		nextStr = styleRed.Render("invalid")
	}

	kindStyle := styleCyan
	if e.Kind == "agent" {
		kindStyle = styleOrange
	}

	// Columns: name (30%), schedule (25%), kind (10%), next (rest)
	nameW := width * 30 / 100
	schedW := width * 25 / 100
	kindW := 10

	name := truncate(e.Name, nameW)
	sched := truncate(e.Schedule, schedW)
	kind := kindStyle.Render(truncate(e.Kind, kindW))
	next := styleDim.Render(nextStr)

	return fmt.Sprintf("%-*s %-*s %-*s %s",
		nameW, name,
		schedW, sched,
		kindW, kind,
		next,
	)
}

// viewLogPane renders the bottom pane showing recent log output.
func (m Model) viewLogPane(width, height int) string {
	inner := width - 4
	if inner < 10 {
		inner = 10
	}

	title := styleHeader.Render("LOG OUTPUT")
	header := padRight(title, inner)

	// Available lines for log content.
	contentLines := height - 4 // header + 2 borders + header row
	if contentLines < 1 {
		contentLines = 1
	}

	// Calculate scroll: auto-scroll to bottom unless user scrolled up.
	totalLogs := len(m.logBuf)
	maxScroll := totalLogs - contentLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	offset := maxScroll - m.logScrollOffset
	if offset < 0 {
		offset = 0
	}

	end := offset + contentLines
	if end > totalLogs {
		end = totalLogs
	}

	var logLines []string
	if totalLogs == 0 {
		logLines = append(logLines, styleDim.Render("  waiting for log output..."))
	} else {
		slice := m.logBuf[offset:end]
		for _, l := range slice {
			l = strings.TrimRight(l, "\n\r")
			l = truncate(l, inner)
			logLines = append(logLines, styleDim.Render(l))
		}
	}

	lines := []string{header}
	lines = append(lines, logLines...)

	// Pad to fill content height.
	paneH := height - 2
	for len(lines) < paneH {
		lines = append(lines, "")
	}

	body := strings.Join(lines, "\n")
	paneStyle := styleBorder.Width(width - 2).Height(height - 2)
	if m.activePane == 1 {
		paneStyle = paneStyle.BorderForeground(lipgloss.Color(draculaPurple))
	}
	return paneStyle.Render(body)
}

// viewHintBar renders the single-line hint strip at the bottom.
func (m Model) viewHintBar(_ int) string {
	var hints string
	if m.activePane == 0 {
		if m.filtering {
			hints = styleDim.Render("[esc] clear filter  [enter] confirm  [tab] logs pane")
		} else {
			hints = styleDim.Render("[j/k] navigate  [e] edit  [d] delete  [enter/r] run now  [/] filter  [tab] logs  [q] quit")
		}
	} else {
		hints = styleDim.Render("[j/k] scroll  [tab] jobs pane  [q] quit")
	}
	return hints
}

// viewEditOverlay renders the edit form overlay.
func (m Model) viewEditOverlay() string {
	ov := m.editOverlay
	labels := [5]string{"Name", "Schedule", "Kind", "Target", "Timeout"}

	var sb strings.Builder
	sb.WriteString(styleHeader.Render("EDIT CRON JOB") + "\n\n")

	for i, f := range ov.fields {
		label := labels[i]
		if i == ov.focusIdx {
			label = styleGreen.Render("> " + label)
		} else {
			label = styleDim.Render("  " + label)
		}
		sb.WriteString(fmt.Sprintf("%-20s %s\n", label, f.View()))
	}

	if ov.errMsg != "" {
		sb.WriteString("\n" + styleError.Render("Error: "+ov.errMsg))
	}

	sb.WriteString("\n" + styleDim.Render("[tab] next field  [enter] save  [esc] cancel"))

	return styleOverlay.Render(sb.String())
}

// viewDeleteConfirm renders the delete confirmation overlay.
func (m Model) viewDeleteConfirm() string {
	name := m.deleteConfirm.entry.Name
	prompt := fmt.Sprintf("Delete %q?\n\n%s  %s",
		name,
		styleRed.Render("[y] yes"),
		styleDim.Render("[n/esc] cancel"),
	)
	return styleOverlay.Render(styleHeader.Render("DELETE CRON JOB") + "\n\n" + prompt)
}

// splitHeight divides total rows into top and bottom, applying a ratio and
// enforcing a minimum row count for each pane.
func splitHeight(total int, ratio float64, minRows int) (top, bot int) {
	top = int(float64(total) * ratio)
	bot = total - top
	if top < minRows {
		top = minRows
		bot = total - top
	}
	if bot < minRows {
		bot = minRows
		top = total - bot
	}
	return
}

// renderOverlay places overlayContent centered over background using lipgloss.Place.
func renderOverlay(background, overlayContent string, width, height int) string {
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlayContent,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(draculaBg)),
	)
}

// padRight pads s to at least n runes wide using spaces.
func padRight(s string, n int) string {
	vis := lipgloss.Width(s)
	if vis >= n {
		return s
	}
	return s + strings.Repeat(" ", n-vis)
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}

package pipelineeditor

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/adam-stokes/orcai/internal/panelrender"
)

// ANSI helpers (package-local).
const (
	aRst = "\x1b[0m"
	aBld = "\x1b[1m"
	aDim = "\x1b[2m"
)

// View renders the full-screen two-column pipeline editor.
func (m Model) View(w, h int) string {
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 40
	}

	pal := m.pal
	leftW := w / 4
	if leftW < 20 {
		leftW = 20
	}
	rightW := w - leftW

	// Top bar.
	topBar := pal.Accent + aBld + " PIPELINE BUILDER" + aRst

	// Body height = total - topbar(1).
	bodyH := h - 1

	leftLines := m.buildLeft(leftW, bodyH)
	rightLines := m.buildRight(rightW, bodyH)

	// Merge left and right columns.
	var rows []string
	rows = append(rows, topBar)
	maxRows := bodyH
	for i := range maxRows {
		var l, r string
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		// Pad left to leftW visible chars.
		lv := lipgloss.Width(l)
		if lv < leftW {
			l = l + strings.Repeat(" ", leftW-lv)
		}
		rows = append(rows, l+r)
	}

	return strings.Join(rows, "\n")
}

// buildLeft delegates to the shared Sidebar sub-model.
func (m Model) buildLeft(w, h int) []string {
	sb := m.sidebar.SetFocused(m.focus == FocusList)
	return sb.View(w, h, m.pal)
}

// buildRight renders: editor (top ~55%) + runner (middle) + chat input (bottom ~15%).
func (m Model) buildRight(w, h int) []string {
	chatH := 4
	if h < 20 {
		chatH = 3
	}
	remaining := h - chatH
	editorH := remaining * 55 / 100
	if editorH < 10 {
		editorH = 10
	}
	runnerH := remaining - editorH
	if runnerH < 5 {
		runnerH = 5
	}

	var rows []string
	rows = append(rows, m.buildEditorBox(w, editorH)...)
	rows = append(rows, m.buildRunnerBox(w, runnerH)...)
	rows = append(rows, m.buildChatBox(w, chatH)...)
	return rows
}

// buildEditorBox delegates to the shared EditorPanel sub-model.
func (m Model) buildEditorBox(w, h int) []string {
	ed := m.editor.SetFocused(m.focus == FocusEditor)
	return ed.View(w, h, m.pal)
}

// buildRunnerBox delegates to the shared RunnerPanel sub-model.
func (m Model) buildRunnerBox(w, h int) []string {
	rn := m.runner.SetFocused(m.focus == FocusRunner)
	return rn.View(w, h, m.pal)
}

// buildChatBox renders the agent runner chat input at the bottom of the right column.
func (m Model) buildChatBox(w, h int) []string {
	pal := m.pal
	borderColor := pal.Border
	if m.focus == FocusChat {
		borderColor = pal.Accent
	}

	var rows []string
	rows = append(rows, panelrender.BoxTop(w, "SEND", borderColor, pal.Accent))

	// Chat input row.
	m.chatInput.Width = w - 6
	if m.chatInput.Width < 10 {
		m.chatInput.Width = 10
	}
	inputLine := "  " + m.chatInput.View()
	rows = append(rows, panelrender.BoxRow(inputLine, w, borderColor))

	// Fill remaining rows.
	for len(rows) < h-2 {
		rows = append(rows, panelrender.BoxRow("", w, borderColor))
	}

	// Hint row.
	hints := []panelrender.Hint{
		{Key: "enter", Desc: "send"},
		{Key: "ctrl+r", Desc: "re-run"},
		{Key: "ctrl+s", Desc: "save"},
		{Key: "shift+tab", Desc: "editor"},
	}
	rows = append(rows, panelrender.BoxRow(panelrender.HintBar(hints, w-2, pal), w, borderColor))
	rows = append(rows, panelrender.BoxBot(w, borderColor))
	return rows
}


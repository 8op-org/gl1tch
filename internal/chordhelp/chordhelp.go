// Package chordhelp runs the global ^spc C-Space chord-key popup.
// Shows a read-only shortcut reference; all actions are handled by the switchboard.
package chordhelp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/adam-stokes/orcai/internal/themes"
)

// helpPalette holds resolved lipgloss colors for the chord-help popup.
type helpPalette struct {
	titleBG lipgloss.Color
	titleFG lipgloss.Color
	fg      lipgloss.Color
	accent  lipgloss.Color
	dim     lipgloss.Color
}

// loadHelpPalette reads the persisted active theme and derives popup colors.
// Falls back to Nord values when no theme is configured.
func loadHelpPalette() helpPalette {
	p := helpPalette{
		titleBG: lipgloss.Color("#88c0d0"),
		titleFG: lipgloss.Color("#2e3440"),
		fg:      lipgloss.Color("#eceff4"),
		accent:  lipgloss.Color("#88c0d0"),
		dim:     lipgloss.Color("#4c566a"),
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	userThemesDir := filepath.Join(home, ".config", "orcai", "themes")
	reg, err := themes.NewRegistry(userThemesDir)
	if err != nil {
		return p
	}
	b := reg.Active()
	if b == nil {
		return p
	}
	if v := b.ResolveRef(b.Modal.TitleBG); v != "" {
		p.titleBG = lipgloss.Color(v)
	}
	if v := b.ResolveRef(b.Modal.TitleFG); v != "" {
		p.titleFG = lipgloss.Color(v)
	}
	if v := b.Palette.FG; v != "" {
		p.fg = lipgloss.Color(v)
	}
	if v := b.Palette.Accent; v != "" {
		p.accent = lipgloss.Color(v)
	}
	if v := b.Palette.Dim; v != "" {
		p.dim = lipgloss.Color(v)
	}
	return p
}

type model struct {
	width  int
	height int
	self   string
	pal    helpPalette
}

func newModel() model {
	self, _ := os.Executable()
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}
	return model{self: self, pal: loadHelpPalette()}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			if m.self != "" {
				exec.Command("tmux", "display-popup", "-E",
					"-w", "42", "-h", "20", m.self, "_picker").Run() //nolint:errcheck
			}
			return m, tea.Quit
		case "s":
			if m.self != "" {
				exec.Command("tmux", "display-popup", "-E",
					"-w", "44", "-h", "6", m.self, "_opsx").Run() //nolint:errcheck
			}
			return m, tea.Quit
		case "esc", "`", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	w := m.width
	if w <= 0 {
		w = 50
	}

	headerStyle := lipgloss.NewStyle().
		Background(m.pal.titleBG).
		Foreground(m.pal.titleFG).
		Bold(true).
		Width(w).
		Padding(0, 1)

	keyStyle := lipgloss.NewStyle().
		Foreground(m.pal.accent).
		Bold(true).
		Width(10)

	descStyle := lipgloss.NewStyle().
		Foreground(m.pal.fg)

	sectionStyle := lipgloss.NewStyle().
		Foreground(m.pal.dim).
		Width(w).
		Padding(0, 1).
		PaddingTop(1)

	rowStyle := lipgloss.NewStyle().
		Width(w).
		Padding(0, 1)

	row := func(key, desc string) string {
		return rowStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				keyStyle.Render(key),
				descStyle.Render(desc),
			),
		)
	}

	dimStyle := lipgloss.NewStyle().Foreground(m.pal.dim)

	rows := []string{
		headerStyle.Render("ORCAI  shortcuts"),
		sectionStyle.Render("session"),
		row("^spc q", "quit workspace"),
		row("^spc d", "detach  (reconnect with: orcai)"),
		row("^spc r", "reload  (updated binary, sessions preserved)"),
		row("n", "new session"),
		row("s", "OpenSpec — propose a feature"),
		sectionStyle.Render("navigation"),
		row("^spc t", "switchboard"),
		row("^spc m", "themes"),
		row("^spc j", "jump to window"),
		row("^spc c", "new window"),
		row("^spc x", "kill pane"),
		rowStyle.Render(""),
		rowStyle.Render(dimStyle.Render("esc  dismiss")),
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// Run starts the chord-help popup as a bubbletea program.
func Run() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("chordhelp error: %v\n", err)
	}
}

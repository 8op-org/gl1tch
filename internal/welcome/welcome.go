// Package welcome implements the ANSI splash screen shown in window 0 on
// fresh ORCAI launch. Enter opens the provider picker; any other keypress
// exits to $SHELL via syscall.Exec.
package welcome

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// ── ANSI palette ───────────────────────────────────────────────────────────────

const (
	aPurple = "\x1b[38;5;141m"
	aPink   = "\x1b[38;5;212m"
	aBold   = "\x1b[1;38;5;212m"
	aBlue   = "\x1b[38;5;61m"
	aDimT   = "\x1b[38;5;66m"
	aReset  = "\x1b[0m"
)

// ── Banner ─────────────────────────────────────────────────────────────────────

func buildWelcomeArt(width int) string {
	if width < 10 {
		width = 52
	}
	inner := width - 2

	pad := func(n int) string {
		if n <= 0 {
			return ""
		}
		return strings.Repeat(" ", n)
	}

	top := aPurple + "╔" + strings.Repeat("═", inner) + "╗" + aReset

	const logoPrefixLen = 37
	logoLine := aPurple + "║" + aPink + " ░▒▓ " + aBold + "O R C A I" + aReset +
		aPink + " ▓▒░" + aBlue + "  Your AI Workspace" + pad(inner-logoPrefixLen) +
		aPurple + "║" + aReset

	const subtitlePrefixLen = 38
	subtitleLine := aPurple + "║" + aBlue + "      tmux · AI agents · open sessions" +
		pad(inner-subtitlePrefixLen) + aPurple + "║" + aReset

	mid := aPurple + "╠" + strings.Repeat("═", inner) + "╣" + aReset

	scanContent := strings.Repeat("▄▀", inner/2)
	if inner%2 == 1 {
		scanContent += "▄"
	}
	scanLine := aPurple + "║" + aPink + scanContent + aPurple + "║" + aReset

	bot := aPurple + "╚" + strings.Repeat("═", inner) + "╝" + aReset

	return strings.Join([]string{top, logoLine, subtitleLine, mid, scanLine, bot}, "\n")
}

// ── Help text ──────────────────────────────────────────────────────────────────

func buildHelp(width int) string {
	col := aDimT + strings.Repeat("─", width) + aReset

	lines := []string{
		col,
		"",
		aBlue + "  Press  " + aPink + "ctrl+space" + aBlue + "  to open the chord menu from anywhere." + aReset,
		"",
		aBlue + "    " + aPink + "n" + aDimT + "  new session   " + aBlue + "(pick AI provider + model)" + aReset,
		aBlue + "    " + aPink + "t" + aDimT + "  sysop panel   " + aBlue + "(agent monitor in current window)" + aReset,
		aBlue + "    " + aPink + "p" + aDimT + "  prompt builder" + aBlue + aReset,
		aBlue + "    " + aPink + "q" + aDimT + "  quit ORCAI" + aReset,
		aBlue + "    " + aPink + "d" + aDimT + "  detach        " + aBlue + "(reconnect later: orcai)" + aReset,
		"",
		col,
		"",
		aDimT + "  ── enter new session · any key continue ──" + aReset,
	}
	return strings.Join(lines, "\n")
}

// ── BubbleTea model ────────────────────────────────────────────────────────────

type model struct {
	width  int
	height int
	self   string
}

func newModel() model {
	self, _ := os.Executable()
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}
	return model{self: self}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "enter" && m.self != "" {
			exec.Command("tmux", "display-popup", "-E",
				"-w", "120", "-h", "40", m.self, "_picker").Run() //nolint:errcheck
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	return buildWelcomeArt(w) + "\n" + buildHelp(w)
}

// ── Entry point ────────────────────────────────────────────────────────────────

// Run launches the splash TUI. After the user presses any key the process
// is replaced by $SHELL.
func Run() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "welcome: %v\n", err)
	}
	execShell()
}

func execShell() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	if err := syscall.Exec(shell, []string{shell}, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "welcome: exec shell: %v\n", err)
		os.Exit(0)
	}
}

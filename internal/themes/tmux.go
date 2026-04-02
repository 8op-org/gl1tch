package themes

import (
	"fmt"
	"os/exec"
)

// TmuxStatusCenterFormat builds the window-status-current-format for the centered hint bar.
// Keys are rendered in accent color, descriptions in dim color.
func TmuxStatusCenterFormat(accent, dim string) string {
	key := func(k string) string { return fmt.Sprintf("#[fg=%s]%s", accent, k) }
	desc := func(d string) string { return fmt.Sprintf("#[fg=%s]%s", dim, d) }
	sp := fmt.Sprintf("#[fg=%s]  ", dim)
	return " " +
		key("^spc d") + desc(" detach") + sp +
		key("^spc r") + desc(" reload") + sp +
		key("/quit") + desc(" quit") + " "
}

// ApplyTmux pushes theme colors to the running tmux session via set-option.
// Called after the user selects a new theme so the status bar updates immediately.
func ApplyTmux(b *Bundle) {
	if b == nil {
		return
	}
	accent := b.Palette.Accent
	dim := b.Palette.Dim
	border := b.Palette.Border

	if accent == "" {
		accent = "#88c0d0"
	}
	if dim == "" {
		dim = "#4c566a"
	}
	if border == "" {
		border = "#3b4252"
	}

	opts := [][]string{
		{"set-option", "-g", "status-style", fmt.Sprintf("fg=%s,bg=default", accent)},
		{"set-option", "-g", "status-left", ""},
		{"set-option", "-g", "status-left-length", "0"},
		{"set-option", "-g", "status-right", ""},
		{"set-option", "-g", "status-right-length", "0"},
		{"set-option", "-g", "status-justify", "centre"},
		{"set-option", "-g", "window-status-format", ""},
		{"set-option", "-g", "window-status-current-format", TmuxStatusCenterFormat(accent, dim)},
		{"set-option", "-g", "pane-border-style", fmt.Sprintf("fg=%s", border)},
		{"set-option", "-g", "pane-active-border-style", fmt.Sprintf("fg=%s", accent)},
	}

	for _, args := range opts {
		_ = exec.Command("tmux", args...).Run()
	}
}

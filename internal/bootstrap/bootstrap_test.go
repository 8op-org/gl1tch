package bootstrap_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/adam-stokes/orcai/internal/bootstrap"
)

func TestWriteTmuxConf(t *testing.T) {
	dir := t.TempDir()
	confPath, err := bootstrap.WriteTmuxConf(dir, "/fake/orcai")
	if err != nil {
		t.Fatalf("WriteTmuxConf: %v", err)
	}
	data, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("tmux.conf not written: %v", err)
	}
	if len(data) == 0 {
		t.Error("tmux.conf is empty")
	}
	expected := filepath.Join(dir, "tmux.conf")
	if confPath != expected {
		t.Errorf("confPath = %q, want %q", confPath, expected)
	}
	if !strings.Contains(string(data), "status-position bottom") {
		t.Error("tmux.conf missing status-position bottom")
	}
}

func TestBuildTmuxConf_Keybindings(t *testing.T) {
	dir := t.TempDir()
	_, err := bootstrap.WriteTmuxConf(dir, "/fake/orcai")
	if err != nil {
		t.Fatalf("WriteTmuxConf: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "tmux.conf"))
	if err != nil {
		t.Fatalf("read tmux.conf: %v", err)
	}
	conf := string(data)
	// Strip tmux color directives (#[fg=...]) from the conf for plain-text assertions.
	colorRe := regexp.MustCompile(`#\[[^\]]*\]`)
	plain := colorRe.ReplaceAllString(conf, "")

	// ctrl+space leader must be present.
	if !strings.Contains(conf, "C-Space") {
		t.Error("tmux.conf missing C-Space leader binding")
	}
	// Backtick leader must be absent.
	if strings.Contains(conf, "bind-key -n `") {
		t.Error("tmux.conf still contains backtick leader binding")
	}
	// Global ESC binding must be absent.
	if strings.Contains(conf, "bind-key -n Escape select-pane") {
		t.Error("tmux.conf still contains global ESC intercept")
	}
	// Status bar must contain new hints; old ^spc n new hint must be gone.
	if strings.Contains(plain, "^spc n new") {
		t.Error("tmux.conf status-right still contains removed '^spc n new' hint")
	}
	if !strings.Contains(plain, "^spc c win") {
		t.Error("tmux.conf status-right missing '^spc c win' hint")
	}
	if !strings.Contains(plain, "^spc j jump") {
		t.Error("tmux.conf status-right missing '^spc j jump' hint")
	}
	// ^spc t switchboard hint must be gone.
	if strings.Contains(plain, "^spc t switchboard") {
		t.Error("tmux.conf status-right still contains removed '^spc t switchboard' hint")
	}
	if !strings.Contains(plain, "^spc t themes") {
		t.Error("tmux.conf status-right missing '^spc t themes' hint")
	}
	// ^spc t must send T (theme picker), not navigate to switchboard window.
	if strings.Contains(conf, "bind-key -T orcai-chord t     { switch-client -T root ; switch-client -t orcai") {
		t.Error("tmux.conf ^spc t binding still navigates to switchboard")
	}
}

func TestBuildTmuxConf_WindowPaneChords(t *testing.T) {
	dir := t.TempDir()
	_, err := bootstrap.WriteTmuxConf(dir, "/fake/orcai")
	if err != nil {
		t.Fatalf("WriteTmuxConf: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "tmux.conf"))
	if err != nil {
		t.Fatalf("read tmux.conf: %v", err)
	}
	conf := string(data)

	chords := []struct {
		key  string
		desc string
	}{
		{"orcai-chord c", "new window (c)"},
		{"orcai-chord [", "previous window ([)"},
		{"orcai-chord ]", "next window (])"},
		{"orcai-chord |", "split pane right (|)"},
		{"orcai-chord -", "split pane down (-)"},
		{"orcai-chord Left", "select pane left"},
		{"orcai-chord Right", "select pane right"},
		{"orcai-chord Up", "select pane up"},
		{"orcai-chord Down", "select pane down"},
		{"orcai-chord x", "kill pane (x)"},
		{"orcai-chord j", "session/window jump (j)"},
	}
	for _, c := range chords {
		if !strings.Contains(conf, c.key) {
			t.Errorf("tmux.conf missing chord binding for %s", c.desc)
		}
	}
}

func TestBuildTmuxConf_WindowStatusFormats(t *testing.T) {
	dir := t.TempDir()
	_, err := bootstrap.WriteTmuxConf(dir, "/fake/orcai")
	if err != nil {
		t.Fatalf("WriteTmuxConf: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "tmux.conf"))
	if err != nil {
		t.Fatalf("read tmux.conf: %v", err)
	}
	conf := string(data)

	// Window list must be suppressed (blank format strings hide all windows).
	if !strings.Contains(conf, `window-status-format ""`) {
		t.Error("window-status-format must be blank to suppress window list")
	}
	if !strings.Contains(conf, `window-status-current-format ""`) {
		t.Error("window-status-current-format must be blank to suppress window list")
	}
}

func TestSessionExists_NoSuchSession(t *testing.T) {
	if !bootstrap.HasTmux() {
		t.Skip("tmux not in PATH")
	}
	got := bootstrap.SessionExists("orcai-test-nonexistent-xyz")
	if got {
		t.Error("SessionExists returned true for a session that should not exist")
	}
}

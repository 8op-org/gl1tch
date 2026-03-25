package welcome

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMsg constructs a tea.KeyMsg for a given key string.
func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func keyMsgSpecial(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func TestUpdate_UnhandledKeyNoOp(t *testing.T) {
	m := newModel()
	next, cmd := m.Update(keyMsg("z"))
	if cmd != nil {
		t.Error("pressing unhandled 'z' should return nil cmd")
	}
	nm := next.(model)
	if nm.launchPicker {
		t.Error("pressing unhandled 'z' should not set launchPicker")
	}
}

func TestUpdate_QQuits(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("pressing 'q' should return a non-nil cmd")
	}
	// tea.Quit is a function; verify it produces a QuitMsg.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("pressing 'q' should return tea.Quit cmd, got %T", msg)
	}
}

func TestUpdate_EscQuits(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(keyMsgSpecial(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("pressing esc should return a non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("pressing esc should return tea.Quit cmd, got %T", msg)
	}
}

func TestUpdate_EnterSetsPickerAndQuits(t *testing.T) {
	m := newModel()
	next, cmd := m.Update(keyMsgSpecial(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("pressing enter should return a non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("pressing enter should return tea.Quit cmd, got %T", msg)
	}
	nm := next.(model)
	if !nm.launchPicker {
		t.Error("pressing enter should set launchPicker = true")
	}
}

func TestUpdate_NSetsPickerAndQuits(t *testing.T) {
	m := newModel()
	next, cmd := m.Update(keyMsg("n"))
	if cmd == nil {
		t.Fatal("pressing 'n' should return a non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("pressing 'n' should return tea.Quit cmd, got %T", msg)
	}
	nm := next.(model)
	if !nm.launchPicker {
		t.Error("pressing 'n' should set launchPicker = true")
	}
}

// isQuitCmd returns true if the cmd produces a tea.QuitMsg.
func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestUpdate_SessionKeysFireAndStay(t *testing.T) {
	keys := []string{"t", "p", "d", "c", "|", "-"}
	for _, key := range keys {
		m := newModel()
		next, cmd := m.Update(keyMsg(key))
		if cmd == nil {
			t.Errorf("pressing %q should return a non-nil cmd (tmux side effect)", key)
			continue
		}
		if isQuitCmd(cmd) {
			t.Errorf("pressing %q should NOT return tea.Quit", key)
		}
		nm := next.(model)
		if nm.launchPicker {
			t.Errorf("pressing %q should not set launchPicker", key)
		}
	}
}

func TestUpdate_ArrowKeysFireAndStay(t *testing.T) {
	arrows := []tea.KeyType{tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight}
	for _, kt := range arrows {
		m := newModel()
		_, cmd := m.Update(keyMsgSpecial(kt))
		if cmd == nil {
			t.Errorf("arrow key %v should return a non-nil cmd", kt)
			continue
		}
		if isQuitCmd(cmd) {
			t.Errorf("arrow key %v should NOT return tea.Quit", kt)
		}
	}
}

func TestBuildHelp_ContainsQEscHint(t *testing.T) {
	p := absDefaults()
	help := buildHelp(80, p)
	if !strings.Contains(help, "q / esc") {
		t.Error("buildHelp should contain 'q / esc' close hint")
	}
}

func TestBuildHelp_NoAnyKeyContinue(t *testing.T) {
	p := absDefaults()
	help := buildHelp(80, p)
	if strings.Contains(help, "any key continue") {
		t.Error("buildHelp should NOT contain 'any key continue'")
	}
}

func TestBuildHelp_ContainsNewShellWindow(t *testing.T) {
	p := absDefaults()
	help := buildHelp(80, p)
	if !strings.Contains(help, "new shell window") {
		t.Error("buildHelp should contain 'new shell window' chord entry")
	}
}

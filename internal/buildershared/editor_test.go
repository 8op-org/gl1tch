package buildershared

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditorPanelNameContent(t *testing.T) {
	e := NewEditorPanel(nil)
	e = e.SetName("my-prompt")
	e = e.SetContent("do something cool")

	if got := e.Name(); got != "my-prompt" {
		t.Errorf("Name: want my-prompt, got %q", got)
	}
	if got := e.Content(); got != "do something cool" {
		t.Errorf("Content: want 'do something cool', got %q", got)
	}
}

func TestEditorPanelFocusCycle(t *testing.T) {
	e := NewEditorPanel(nil)
	e = e.SetFocused(true)
	// Start at picker (0).
	if e.FocusField() != EditorFocusPicker {
		t.Fatalf("expected initial focus = EditorFocusPicker, got %d", e.FocusField())
	}

	// Tab from picker with internal focus=0 should advance picker internal focus,
	// not move to name. Tab again (now internal=1) should advance to name.
	e2, _ := e.Update(tea.KeyMsg{Type: tea.KeyTab})
	// picker internal focus moved to 1
	e3, _ := e2.Update(tea.KeyMsg{Type: tea.KeyTab})
	if e3.FocusField() != EditorFocusName {
		t.Errorf("after two tabs from picker: expected EditorFocusName, got %d", e3.FocusField())
	}

	// Tab from name -> content.
	e4, _ := e3.Update(tea.KeyMsg{Type: tea.KeyTab})
	if e4.FocusField() != EditorFocusContent {
		t.Errorf("tab from name: expected EditorFocusContent, got %d", e4.FocusField())
	}

	// Tab from content -> EditorTabOutMsg.
	_, cmd := e4.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("tab from content: expected EditorTabOutMsg cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(EditorTabOutMsg); !ok {
		t.Errorf("tab from content: expected EditorTabOutMsg, got %T", msg)
	}
}

func TestEditorPanelShiftTabOut(t *testing.T) {
	e := NewEditorPanel(nil)
	e = e.SetFocused(true)
	// Focus is at picker — shift+tab should emit EditorShiftTabOutMsg.
	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if cmd == nil {
		t.Fatal("shift+tab from picker: expected EditorShiftTabOutMsg cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(EditorShiftTabOutMsg); !ok {
		t.Errorf("shift+tab from picker: expected EditorShiftTabOutMsg, got %T", msg)
	}
}

func TestEditorPanelShiftTabNavigation(t *testing.T) {
	e := NewEditorPanel(nil)
	e = e.SetFocused(true)
	e.focus = EditorFocusContent

	// shift+tab from content -> name.
	e2, _ := e.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if e2.FocusField() != EditorFocusName {
		t.Errorf("shift+tab from content: expected EditorFocusName, got %d", e2.FocusField())
	}

	// shift+tab from name -> picker.
	e3, _ := e2.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if e3.FocusField() != EditorFocusPicker {
		t.Errorf("shift+tab from name: expected EditorFocusPicker, got %d", e3.FocusField())
	}
}

func TestEditorPanelSetFocused(t *testing.T) {
	e := NewEditorPanel(nil)
	if e.focused {
		t.Error("new EditorPanel should not be focused")
	}
	e = e.SetFocused(true)
	if !e.focused {
		t.Error("SetFocused(true) should set focused")
	}
	e = e.SetFocused(false)
	if e.focused {
		t.Error("SetFocused(false) should clear focused")
	}
}

package buildershared

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSidebarFiltered(t *testing.T) {
	items := []string{"alpha", "beta", "gamma", "alphabet"}
	s := NewSidebar("TEST", items)

	// No query — all items returned.
	if got := s.Filtered(); len(got) != 4 {
		t.Fatalf("no query: want 4 items, got %d", len(got))
	}

	// Set query via Update.
	s.query = "alpha"
	filtered := s.Filtered()
	if len(filtered) != 2 {
		t.Fatalf("query=alpha: want 2 items (alpha, alphabet), got %d: %v", len(filtered), filtered)
	}
	for _, item := range filtered {
		if item != "alpha" && item != "alphabet" {
			t.Errorf("unexpected item %q", item)
		}
	}
}

func TestSidebarSelectedName(t *testing.T) {
	s := NewSidebar("TEST", []string{"foo", "bar", "baz"})
	// Default sel=0.
	if got := s.SelectedName(); got != "foo" {
		t.Errorf("want foo, got %q", got)
	}
	s.sel = 2
	if got := s.SelectedName(); got != "baz" {
		t.Errorf("want baz, got %q", got)
	}
}

func TestSidebarSetItems(t *testing.T) {
	s := NewSidebar("TEST", []string{"a", "b", "c"})
	s.sel = 2
	s = s.SetItems([]string{"x", "y"})
	if s.sel != 1 { // clamped to len-1=1
		t.Errorf("expected sel clamped to 1, got %d", s.sel)
	}
}

func TestSidebarNavigationKeys(t *testing.T) {
	s := NewSidebar("TEST", []string{"a", "b", "c"})
	s = s.SetFocused(true)

	press := func(key string) Sidebar {
		s2, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		return s2
	}
	pressSpecial := func(key tea.KeyType) Sidebar {
		s2, _ := s.Update(tea.KeyMsg{Type: key})
		return s2
	}

	s = press("j")
	if s.sel != 1 {
		t.Errorf("j: expected sel=1, got %d", s.sel)
	}
	s = press("j")
	if s.sel != 2 {
		t.Errorf("j: expected sel=2, got %d", s.sel)
	}
	s = press("j") // at end, should not exceed
	if s.sel != 2 {
		t.Errorf("j past end: expected sel=2, got %d", s.sel)
	}
	s = press("k")
	if s.sel != 1 {
		t.Errorf("k: expected sel=1, got %d", s.sel)
	}
	_ = pressSpecial(tea.KeyUp)
	_ = pressSpecial(tea.KeyDown)
}

func TestSidebarSelectMsg(t *testing.T) {
	s := NewSidebar("TEST", []string{"foo", "bar"})
	s.sel = 0

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter: expected a command, got nil")
	}
	msg := cmd()
	sel, ok := msg.(SidebarSelectMsg)
	if !ok {
		t.Fatalf("expected SidebarSelectMsg, got %T", msg)
	}
	if sel.Name != "foo" {
		t.Errorf("expected Name=foo, got %q", sel.Name)
	}
}

func TestSidebarDeleteMsg(t *testing.T) {
	s := NewSidebar("TEST", []string{"one", "two"})
	s.sel = 1

	// Press 'd' to enter confirm mode.
	s2, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if !s2.confirmDelete {
		t.Fatal("expected confirmDelete=true after d")
	}

	// Press 'y' to confirm.
	s3, cmd := s2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if s3.confirmDelete {
		t.Error("expected confirmDelete=false after y")
	}
	if cmd == nil {
		t.Fatal("expected SidebarDeleteMsg command")
	}
	msg := cmd()
	del, ok := msg.(SidebarDeleteMsg)
	if !ok {
		t.Fatalf("expected SidebarDeleteMsg, got %T", msg)
	}
	if del.Name != "two" {
		t.Errorf("expected Name=two, got %q", del.Name)
	}
}

func TestSidebarSearch(t *testing.T) {
	s := NewSidebar("TEST", []string{"apple", "apricot", "banana"})

	// Press '/' to start search.
	s2, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if !s2.searching {
		t.Fatal("expected searching=true after /")
	}

	// Type 'ap'.
	s2, _ = s2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	s2, _ = s2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if s2.query != "ap" {
		t.Errorf("expected query=ap, got %q", s2.query)
	}
	if got := s2.Filtered(); len(got) != 2 {
		t.Errorf("want 2 filtered items for 'ap', got %d: %v", len(got), got)
	}

	// Esc cancels search.
	s3, _ := s2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s3.searching {
		t.Error("expected searching=false after esc")
	}
	if s3.query != "" {
		t.Errorf("expected query cleared, got %q", s3.query)
	}
}

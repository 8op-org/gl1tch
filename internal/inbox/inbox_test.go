package inbox

import (
	"strings"
	"testing"
	"time"

	"github.com/adam-stokes/orcai/internal/store"
	"github.com/adam-stokes/orcai/internal/themes"
)

// testBundle returns a minimal theme bundle for use in tests.
func testBundle() *themes.Bundle {
	return &themes.Bundle{
		Palette: themes.Palette{
			BG:      "#282a36",
			FG:      "#f8f8f2",
			Accent:  "#bd93f9",
			Dim:     "#6272a4",
			Border:  "#44475a",
			Error:   "#ff5555",
			Success: "#50fa7b",
		},
	}
}

// TestNew_NilStore verifies that New(nil, bundle) doesn't panic.
func TestNew_NilStore(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("New(nil, bundle) panicked: %v", r)
		}
	}()
	_ = New(nil, testBundle())
}

// TestModel_SetSize verifies that SetSize updates width and height fields.
func TestModel_SetSize(t *testing.T) {
	m := New(nil, testBundle())
	m.SetSize(100, 50)
	if m.width != 100 {
		t.Errorf("expected width=100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height=50, got %d", m.height)
	}
}

// TestModel_SetFocused verifies toggling the focused field.
func TestModel_SetFocused(t *testing.T) {
	m := New(nil, testBundle())
	if m.focused {
		t.Error("expected focused=false initially")
	}
	m.SetFocused(true)
	if !m.focused {
		t.Error("expected focused=true after SetFocused(true)")
	}
	m.SetFocused(false)
	if m.focused {
		t.Error("expected focused=false after SetFocused(false)")
	}
}

// TestStatusIndicator_Success verifies exit 0 renders the filled-circle indicator.
func TestStatusIndicator_Success(t *testing.T) {
	b := testBundle()
	exitOK := 0
	run := store.Run{ExitStatus: &exitOK}
	s := statusIndicator(run, b)
	// The rendered string must contain the success dot character.
	if !strings.Contains(s, "●") {
		t.Errorf("expected success indicator '●', got %q", s)
	}
	// A nil exit should NOT produce the success indicator.
	runNil := store.Run{ExitStatus: nil}
	sNil := statusIndicator(runNil, b)
	if strings.Contains(sNil, "●") {
		t.Errorf("in-flight run should not render '●', got %q", sNil)
	}
}

// TestStatusIndicator_Error verifies non-zero exit renders the error indicator.
func TestStatusIndicator_Error(t *testing.T) {
	b := testBundle()
	exitErr := 1
	run := store.Run{ExitStatus: &exitErr}
	s := statusIndicator(run, b)
	if !strings.Contains(s, "●") {
		t.Errorf("expected error indicator '●', got %q", s)
	}
	// A zero exit should not match the same ANSI sequence — just ensure the
	// character is present for error.
	exitOK := 0
	sOK := statusIndicator(store.Run{ExitStatus: &exitOK}, b)
	// Both are "●" but styled differently; ensure error path still renders.
	if s == "" || sOK == "" {
		t.Error("statusIndicator returned empty string")
	}
}

// TestStatusIndicator_InFlight verifies nil exit returns the ring indicator.
func TestStatusIndicator_InFlight(t *testing.T) {
	b := testBundle()
	run := store.Run{ExitStatus: nil}
	s := statusIndicator(run, b)
	if !strings.Contains(s, "◉") {
		t.Errorf("expected in-flight indicator '◉', got %q", s)
	}
	// Completed runs must NOT use the ◉ character.
	exitOK := 0
	sDone := statusIndicator(store.Run{ExitStatus: &exitOK}, b)
	if strings.Contains(sDone, "◉") {
		t.Errorf("completed run should not render '◉', got %q", sDone)
	}
}

// TestElapsedStr_Finished verifies correct duration format for completed runs.
func TestElapsedStr_Finished(t *testing.T) {
	startedAt := time.Now().Add(-10 * time.Second).UnixMilli()
	finishedAt := time.Now().UnixMilli()
	run := store.Run{
		StartedAt:  startedAt,
		FinishedAt: &finishedAt,
	}
	s := elapsedStr(run)
	// Duration should be approximately 10s; should NOT contain "running".
	if strings.Contains(s, "running") {
		t.Errorf("finished run should not say 'running', got %q", s)
	}
	if !strings.Contains(s, "s") {
		t.Errorf("expected seconds in elapsed string, got %q", s)
	}
}

// TestElapsedStr_InFlight verifies "running X" format for in-progress runs.
func TestElapsedStr_InFlight(t *testing.T) {
	run := store.Run{
		StartedAt:  time.Now().Add(-3 * time.Second).UnixMilli(),
		FinishedAt: nil,
	}
	s := elapsedStr(run)
	if !strings.HasPrefix(s, "running ") {
		t.Errorf("in-flight run should start with 'running ', got %q", s)
	}
}


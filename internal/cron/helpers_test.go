package cron

import (
	"strings"
	"testing"
	"time"
)

func TestNextRun_Valid(t *testing.T) {
	e := Entry{Name: "test", Schedule: "* * * * *", Kind: "pipeline", Target: "x"}
	next, err := NextRun(e)
	if err != nil {
		t.Fatalf("NextRun: unexpected error: %v", err)
	}
	if !next.After(time.Now()) {
		t.Errorf("expected next run to be in the future, got %v", next)
	}
}

func TestNextRun_Invalid(t *testing.T) {
	e := Entry{Name: "bad", Schedule: "not-a-cron", Kind: "pipeline", Target: "x"}
	_, err := NextRun(e)
	if err == nil {
		t.Fatal("expected error for invalid schedule, got nil")
	}
}

func TestFormatRelative_Minutes(t *testing.T) {
	t.Parallel()
	future := time.Now().Add(4*time.Minute + 30*time.Second)
	got := FormatRelative(future)
	if !strings.HasPrefix(got, "in ") {
		t.Errorf("expected 'in ...' prefix, got %q", got)
	}
	if !strings.Contains(got, "m") {
		t.Errorf("expected minutes in output, got %q", got)
	}
}

func TestFormatRelative_Hours(t *testing.T) {
	t.Parallel()
	future := time.Now().Add(2*time.Hour + 30*time.Minute)
	got := FormatRelative(future)
	if !strings.Contains(got, "h") {
		t.Errorf("expected hours in output for 2h30m, got %q", got)
	}
}

func TestFormatRelative_Days(t *testing.T) {
	t.Parallel()
	future := time.Now().Add(25 * time.Hour)
	got := FormatRelative(future)
	if !strings.Contains(got, "d") {
		t.Errorf("expected days in output for 25h, got %q", got)
	}
}

func TestFormatRelative_Past(t *testing.T) {
	t.Parallel()
	past := time.Now().Add(-1 * time.Minute)
	got := FormatRelative(past)
	if got != "now" {
		t.Errorf("expected 'now' for past time, got %q", got)
	}
}

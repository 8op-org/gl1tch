package pipeline

import (
	"strings"
	"testing"
)

func TestStdlib_StringsLoads(t *testing.T) {
	src := `(include "std/strings") (kebab-case "Hello World")`
	got := evalHelper(t, src)
	if got != "hello-world" {
		t.Fatalf("want %q, got %q", "hello-world", got)
	}
}

func TestStdlib_SnakeCase(t *testing.T) {
	src := `(include "std/strings") (snake-case "Hello World")`
	got := evalHelper(t, src)
	if got != "hello_world" {
		t.Fatalf("want %q, got %q", "hello_world", got)
	}
}

func TestStdlib_CollectionsLoads(t *testing.T) {
	// compact removes empty strings from a list
	src := `(include "std/collections") (compact (list "alice" "" "bob"))`
	got := evalHelper(t, src)
	if !strings.Contains(got, "alice") || !strings.Contains(got, "bob") {
		t.Fatalf("expected alice and bob, got %q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("expected 2 items (1 newline separator), got %q", got)
	}
}

func TestStdlib_IOLoads(t *testing.T) {
	// Just verify it parses and loads without error
	src := `(include "std/io")`
	got := evalHelper(t, src)
	_ = got
}

func TestStdlib_BlankAndPresent(t *testing.T) {
	src := `(include "std/strings") (list (blank? "") (blank? "hello") (present? "") (present? "hello"))`
	got := evalHelper(t, src)
	// blank? "" = true, blank? "hello" = false, present? "" = false, present? "hello" = true
	if !strings.Contains(got, "true") {
		t.Fatalf("expected some true values, got %q", got)
	}
}

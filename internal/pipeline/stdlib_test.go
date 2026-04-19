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

func TestStdlib_BlankAndPresent(t *testing.T) {
	src := `(include "std/strings") (list (blank? "") (blank? "hello") (present? "") (present? "hello"))`
	got := evalHelper(t, src)
	// blank? "" = true, blank? "hello" = false, present? "" = false, present? "hello" = true
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 values, got %d: %q", len(lines), got)
	}
	if lines[0] != "true" || lines[1] != "false" || lines[2] != "false" || lines[3] != "true" {
		t.Fatalf("unexpected blank?/present? results: %q", got)
	}
}

func TestStdlib_Words(t *testing.T) {
	src := `(include "std/strings") (unwords (words "  hello world  "))`
	got := evalHelper(t, src)
	if got != "hello world" {
		t.Fatalf("want %q, got %q", "hello world", got)
	}
}

func TestBuiltin_SliceLiteral(t *testing.T) {
	// slice works with literal numeric args
	got := evalHelper(t, `(slice "Hello World" 0 5)`)
	if got != "Hello" {
		t.Fatalf("want %q, got %q", "Hello", got)
	}
}

func TestStdlib_CollectionsCompact(t *testing.T) {
	src := `(include "std/collections") (compact (list "alice" "" "bob"))`
	got := evalHelper(t, src)
	if !strings.Contains(got, "alice") || !strings.Contains(got, "bob") {
		t.Fatalf("expected alice and bob, got %q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("expected 2 items (1 newline separator), got %q", got)
	}
}

func TestStdlib_CollectionsUnique(t *testing.T) {
	src := `(include "std/collections") (unique (list "a" "b" "a" "c" "b"))`
	got := evalHelper(t, src)
	if got != "a\nb\nc" {
		t.Fatalf("want %q, got %q", "a\nb\nc", got)
	}
}

func TestStdlib_IOLoads(t *testing.T) {
	src := `(include "std/io")`
	got := evalHelper(t, src)
	_ = got
}

func TestStdlib_IOReadLines(t *testing.T) {
	src := `(include "std/io") (count (read-lines "testdata/par-demo.glitch"))`
	got := evalHelper(t, src)
	if got == "0" || got == "" {
		t.Fatal("expected non-zero line count")
	}
}

// Test the native builtins directly (not via stdlib)

func TestBuiltin_Upper(t *testing.T) {
	got := evalHelper(t, `(upper "hello")`)
	if got != "HELLO" {
		t.Fatalf("want %q, got %q", "HELLO", got)
	}
}

func TestBuiltin_Lower(t *testing.T) {
	got := evalHelper(t, `(lower "HELLO")`)
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestBuiltin_Trim(t *testing.T) {
	got := evalHelper(t, `(trim "  hello  ")`)
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestBuiltin_Replace(t *testing.T) {
	got := evalHelper(t, `(replace "a/b/c" "/" "-")`)
	if got != "a-b-c" {
		t.Fatalf("want %q, got %q", "a-b-c", got)
	}
}

func TestBuiltin_Contains(t *testing.T) {
	got := evalHelper(t, `(contains "foobar" "foo")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestBuiltin_SplitJoin(t *testing.T) {
	got := evalHelper(t, `(join (split "a/b/c" "/") "-")`)
	if got != "a-b-c" {
		t.Fatalf("want %q, got %q", "a-b-c", got)
	}
}

func TestBuiltin_Lines(t *testing.T) {
	got := evalHelper(t, `(count (lines "a\nb\nc"))`)
	if got != "3" {
		t.Fatalf("want %q, got %q", "3", got)
	}
}

func TestBuiltin_RegexMatch(t *testing.T) {
	// regex-match returns all matches as a list
	got := evalHelper(t, `(regex-match "[a-z]+" "hello world")`)
	if got != "hello\nworld" {
		t.Fatalf("want %q, got %q", "hello\nworld", got)
	}
	// No match returns empty list
	got2 := evalHelper(t, `(count (regex-match "[0-9]+" "no digits"))`)
	if got2 != "0" {
		t.Fatalf("want %q, got %q", "0", got2)
	}
}

func TestBuiltin_RegexFind(t *testing.T) {
	// regex-find returns first match
	got := evalHelper(t, `(regex-find "[0-9]+" "abc123def456")`)
	if got != "123" {
		t.Fatalf("want %q, got %q", "123", got)
	}
}

func TestBuiltin_Set(t *testing.T) {
	got := evalHelper(t, `(set (list "a" "b" "a" "c"))`)
	if got != "a\nb\nc" {
		t.Fatalf("want %q, got %q", "a\nb\nc", got)
	}
}

func TestBuiltin_Difference(t *testing.T) {
	got := evalHelper(t, `(difference (list "a" "b" "c") (list "b"))`)
	if got != "a\nc" {
		t.Fatalf("want %q, got %q", "a\nc", got)
	}
}

func TestBuiltin_Assert(t *testing.T) {
	got := evalHelper(t, `(assert (= "1" "1") "should be equal")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestBuiltin_AssertFails(t *testing.T) {
	ev := NewEvaluator()
	_, err := ev.RunSource([]byte(`(assert (= "1" "2") "not equal")`))
	if err == nil {
		t.Fatal("expected assert to fail")
	}
	if !strings.Contains(err.Error(), "not equal") {
		t.Fatalf("expected error to contain message, got: %s", err)
	}
}

func TestBuiltin_Lt(t *testing.T) {
	got := evalHelper(t, `(< "1" "2")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestBuiltin_StartsEndsWith(t *testing.T) {
	got := evalHelper(t, `(starts-with "hello world" "hello")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
	got2 := evalHelper(t, `(ends-with "hello world" "world")`)
	if got2 != "true" {
		t.Fatalf("want %q, got %q", "true", got2)
	}
}

func TestBuiltin_Flatten(t *testing.T) {
	got := evalHelper(t, `(flatten (list (list "a" "b") (list "c" "d")))`)
	if got != "a\nb\nc\nd" {
		t.Fatalf("want %q, got %q", "a\nb\nc\nd", got)
	}
}

func TestInclude_CircularGuard(t *testing.T) {
	// Including the same std module twice should not error or double-define
	src := `(include "std/strings") (include "std/strings") (kebab-case "Hello World")`
	got := evalHelper(t, src)
	if got != "hello-world" {
		t.Fatalf("want %q, got %q", "hello-world", got)
	}
}

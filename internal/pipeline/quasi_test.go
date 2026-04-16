// internal/pipeline/quasi_test.go
package pipeline

import "testing"

func TestLexQuasiLiteralOnly(t *testing.T) {
	parts, err := lexQuasi("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("want 1 part, got %d", len(parts))
	}
	if parts[0].Kind != partLiteral || parts[0].Literal != "hello world" {
		t.Errorf("got %+v", parts[0])
	}
}

func TestLexQuasiBareRef(t *testing.T) {
	parts, err := lexQuasi("hi ~name there")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []quasiPart{
		{Kind: partLiteral, Literal: "hi "},
		{Kind: partRef, RefBase: "name"},
		{Kind: partLiteral, Literal: " there"},
	}
	assertParts(t, parts, want)
}

func TestLexQuasiDottedRef(t *testing.T) {
	parts, _ := lexQuasi("repo=~param.repo x")
	want := []quasiPart{
		{Kind: partLiteral, Literal: "repo="},
		{Kind: partRef, RefBase: "param", RefPath: []string{"repo"}},
		{Kind: partLiteral, Literal: " x"},
	}
	assertParts(t, parts, want)
}

func TestLexQuasiEscapedTilde(t *testing.T) {
	parts, _ := lexQuasi(`cp a \~/dest`)
	if len(parts) != 1 {
		t.Fatalf("want 1 part, got %d", len(parts))
	}
	if parts[0].Literal != "cp a ~/dest" {
		t.Errorf("got literal %q", parts[0].Literal)
	}
}

func TestLexQuasiNoInterpolationWithoutTilde(t *testing.T) {
	parts, _ := lexQuasi("plain text no refs")
	if len(parts) != 1 || parts[0].Kind != partLiteral {
		t.Fatalf("got %+v", parts)
	}
}

func assertParts(t *testing.T, got, want []quasiPart) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len: got %d, want %d; got=%+v", len(got), len(want), got)
	}
	for i := range got {
		if got[i].Kind != want[i].Kind ||
			got[i].Literal != want[i].Literal ||
			got[i].RefBase != want[i].RefBase ||
			!stringSlicesEqual(got[i].RefPath, want[i].RefPath) {
			t.Errorf("part %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

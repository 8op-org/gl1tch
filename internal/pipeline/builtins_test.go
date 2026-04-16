// internal/pipeline/builtins_test.go
package pipeline

import "testing"

func TestBuiltinStep(t *testing.T) {
	s := NewScope()
	s.SetSteps(map[string]string{"fetch": "hello"})
	v, err := callBuiltin("step", []string{"fetch"}, s)
	if err != nil || v != "hello" {
		t.Errorf("got %q err=%v", v, err)
	}
}

func TestBuiltinStepUnknown(t *testing.T) {
	s := NewScope()
	s.SetSteps(map[string]string{})
	_, err := callBuiltin("step", []string{"missing"}, s)
	if err == nil {
		t.Fatal("expected error on unknown step")
	}
}

func TestBuiltinOr(t *testing.T) {
	s := NewScope()
	v, _ := callBuiltin("or", []string{"", "fallback"}, s)
	if v != "fallback" {
		t.Errorf("got %q", v)
	}
	v2, _ := callBuiltin("or", []string{"first", "second"}, s)
	if v2 != "first" {
		t.Errorf("got %q", v2)
	}
}

func TestBuiltinUnknown(t *testing.T) {
	_, err := callBuiltin("doesnotexist", nil, NewScope())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStringBuiltins(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"upper", []string{"hello"}, "HELLO"},
		{"lower", []string{"HI"}, "hi"},
		{"trim", []string{"  x  "}, "x"},
		{"trimPrefix", []string{"pre-", "pre-thing"}, "thing"},
		{"trimSuffix", []string{"-end", "thing-end"}, "thing"},
		{"replace", []string{"/", "-", "a/b/c"}, "a-b-c"},
		{"truncate", []string{"5", "abcdefg"}, "abcde"},
		{"truncate noop", []string{"100", "abc"}, "abc"},
		{"contains", []string{"foobar", "foo"}, "true"},
		{"hasPrefix", []string{"elastic/x", "elastic"}, "true"},
		{"hasSuffix", []string{"elastic/x", "/x"}, "true"},
		{"split", []string{"/", "a/b/c"}, "a\nb\nc"},
		{"join", []string{"-", "a\nb\nc"}, "a-b-c"},
		{"first", []string{"a\nb\nc"}, "a"},
		{"last", []string{"a\nb\nc"}, "c"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := callBuiltin(strings_functionName(tc.name), tc.args, NewScope())
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// strings_functionName strips any " noop"/" short" suffix used only for test naming.
func strings_functionName(n string) string {
	if i := indexOf(n, ' '); i >= 0 {
		return n[:i]
	}
	return n
}

func indexOf(s string, ch byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ch {
			return i
		}
	}
	return -1
}

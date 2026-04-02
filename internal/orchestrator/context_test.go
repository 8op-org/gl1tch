package orchestrator

import (
	"strings"
	"testing"
)

func TestWorkflowContextSetGet(t *testing.T) {
	c := NewWorkflowContext()
	c.Set("foo", "bar")
	if got := c.Get("foo"); got != "bar" {
		t.Errorf("Get(%q) = %q, want %q", "foo", got, "bar")
	}
	if got := c.Get("missing"); got != "" {
		t.Errorf("Get(missing) = %q, want empty string", got)
	}
}

func TestWorkflowContextTruncation(t *testing.T) {
	c := NewWorkflowContext()
	// Build a value larger than 16 KB.
	big := strings.Repeat("x", maxContextValueBytes+100)
	c.Set("bigkey", big)
	got := c.Get("bigkey")
	if len(got) != maxContextValueBytes {
		t.Errorf("truncated value len = %d, want %d", len(got), maxContextValueBytes)
	}
}

func TestWorkflowContextMarshalUnmarshal(t *testing.T) {
	c := NewWorkflowContext()
	c.Set("a", "1")
	c.Set("b", "2")

	b, err := c.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	c2 := NewWorkflowContext()
	if err := c2.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := c2.Get("a"); got != "1" {
		t.Errorf("after unmarshal, Get(%q) = %q, want %q", "a", got, "1")
	}
	if got := c2.Get("b"); got != "2" {
		t.Errorf("after unmarshal, Get(%q) = %q, want %q", "b", got, "2")
	}
}

func TestExpandTemplate(t *testing.T) {
	c := NewWorkflowContext()
	c.Set("foo", "world")
	c.Set("step1.output", "result")

	tests := []struct {
		input string
		want  string
	}{
		{"hello {{ctx.foo}}", "hello world"},
		{"{{ctx.step1.output}} done", "result done"},
		{"no placeholders", "no placeholders"},
		{"unknown: {{ctx.missing}}", "unknown: "},
		{"{{ctx.foo}} and {{ctx.step1.output}}", "world and result"},
	}

	for _, tt := range tests {
		got := ExpandTemplate(tt.input, c)
		if got != tt.want {
			t.Errorf("ExpandTemplate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWorkflowContextUnmarshalInvalidJSON(t *testing.T) {
	c := NewWorkflowContext()
	if err := c.Unmarshal([]byte("not-json")); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

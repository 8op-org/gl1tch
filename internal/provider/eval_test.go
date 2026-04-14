package provider

import (
	"strings"
	"testing"
)

func TestBuildEvalPrompt(t *testing.T) {
	prompt := BuildEvalPrompt("classify this bug", "It's a UI rendering issue")
	if prompt == "" {
		t.Fatal("expected non-empty eval prompt")
	}
	if !strings.Contains(prompt, "classify this bug") {
		t.Error("eval prompt missing original task")
	}
	if !strings.Contains(prompt, "UI rendering issue") {
		t.Error("eval prompt missing response")
	}
}

func TestParseEvalScore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"bare number", "4", 4},
		{"with newline", "5\n", 5},
		{"with text", "I'd rate this a 3 because...", 3},
		{"markdown", "**4**", 4},
		{"no number", "this is great", 0},
		{"out of range high", "7", 0},
		{"out of range low", "0", 0},
		{"with whitespace", "  3  ", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseEvalScore(tt.input); got != tt.expected {
				t.Errorf("ParseEvalScore(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

package provider

import (
	"testing"
)

func TestIsAgent(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"claude", true},
		{"copilot", true},
		{"gemini", true},
		{"ollama", false},
		{"openrouter", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAgent(tt.name); got != tt.want {
				t.Errorf("IsAgent(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		agent string
		model string
		want  []string
	}{
		{"claude", "", []string{"claude", "-p", "--output-format", "json"}},
		{"claude", "sonnet", []string{"claude", "-p", "--output-format", "json", "--model", "sonnet"}},
		{"copilot", "", []string{"gh", "copilot", "-p"}},
		{"gemini", "", []string{"gemini", "-p"}},
	}
	for _, tt := range tests {
		t.Run(tt.agent+"/"+tt.model, func(t *testing.T) {
			a := KnownAgents[tt.agent]
			got := a.buildArgs(tt.model)
			if len(got) != len(tt.want) {
				t.Fatalf("buildArgs() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("buildArgs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseAgentJSON(t *testing.T) {
	t.Run("valid claude json", func(t *testing.T) {
		raw := `{"result":"hello world","usage":{"input_tokens":100,"output_tokens":20},"cost_usd":0.005}`
		parsed, ok := parseAgentJSON(raw)
		if !ok {
			t.Fatal("expected successful parse")
		}
		if parsed.response != "hello world" {
			t.Errorf("response = %q, want %q", parsed.response, "hello world")
		}
		if parsed.tokensIn != 100 {
			t.Errorf("tokensIn = %d, want 100", parsed.tokensIn)
		}
		if parsed.tokensOut != 20 {
			t.Errorf("tokensOut = %d, want 20", parsed.tokensOut)
		}
	})

	t.Run("plain text fallback", func(t *testing.T) {
		_, ok := parseAgentJSON("just plain text")
		if ok {
			t.Error("expected parse failure for plain text")
		}
	})

	t.Run("json without result field", func(t *testing.T) {
		_, ok := parseAgentJSON(`{"foo":"bar"}`)
		if ok {
			t.Error("expected parse failure for json without result")
		}
	})
}

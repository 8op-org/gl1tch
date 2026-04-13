package provider

import "testing"

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 2},
		{"hello world this is a test", 7},
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got != tt.want {
			t.Errorf("EstimateTokens(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestEstimateCost_Claude(t *testing.T) {
	cost := EstimateCost("claude", 1000, 500)
	if cost <= 0 {
		t.Errorf("expected non-zero cost for claude, got %f", cost)
	}
	// 1000 * 3.00/1M + 500 * 15.00/1M = 0.003 + 0.0075 = 0.0105
	expected := 0.0105
	if cost < expected-0.0001 || cost > expected+0.0001 {
		t.Errorf("EstimateCost(claude, 1000, 500) = %f, want ~%f", cost, expected)
	}
}

func TestEstimateCost_Ollama(t *testing.T) {
	cost := EstimateCost("ollama", 1000, 500)
	if cost != 0 {
		t.Errorf("expected zero cost for ollama, got %f", cost)
	}
}

func TestEstimateCost_Unknown(t *testing.T) {
	cost := EstimateCost("unknown-provider", 1000, 500)
	if cost != 0 {
		t.Errorf("expected zero cost for unknown provider, got %f", cost)
	}
}

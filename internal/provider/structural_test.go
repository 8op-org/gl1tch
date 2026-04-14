package provider

import "testing"

func TestCheckStructure_JSON(t *testing.T) {
	tests := []struct {
		name   string
		format string
		input  string
		ok     bool
	}{
		{"valid json", "json", `{"key": "value"}`, true},
		{"invalid json", "json", `not json at all`, false},
		{"empty json", "json", "", false},
		{"json array", "json", `[1, 2, 3]`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStructure(tt.format, tt.input); got != tt.ok {
				t.Errorf("CheckStructure(%q, %q) = %v, want %v", tt.format, tt.input, got, tt.ok)
			}
		})
	}
}

func TestCheckStructure_YAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ok    bool
	}{
		{"valid yaml", "key: value\nlist:\n  - a\n  - b", true},
		{"empty", "", false},
		{"bare string", "just a string", true}, // valid YAML scalar
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStructure("yaml", tt.input); got != tt.ok {
				t.Errorf("CheckStructure(yaml, %q) = %v, want %v", tt.input, got, tt.ok)
			}
		})
	}
}

func TestCheckStructure_NoFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ok    bool
	}{
		{"normal text", "Here is my analysis of the issue.", true},
		{"empty", "", false},
		{"whitespace only", "   \n  ", false},
		{"refusal I cannot", "I cannot help with that request.", false},
		{"refusal I'm sorry", "I'm sorry, I can't assist with that.", false},
		{"refusal as AI", "As an AI language model, I cannot", false},
		{"contains cannot but not refusal", "The system cannot connect to the database.", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStructure("", tt.input); got != tt.ok {
				t.Errorf("CheckStructure('', %q) = %v, want %v", tt.input, got, tt.ok)
			}
		})
	}
}

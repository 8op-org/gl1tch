package provider

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckStructure validates an LLM response based on the expected format.
// Returns true if the response passes structural checks.
// Format values: "json", "yaml", or "" (plain text).
func CheckStructure(format, response string) bool {
	trimmed := strings.TrimSpace(response)
	if trimmed == "" {
		return false
	}

	switch strings.ToLower(format) {
	case "json":
		var v any
		return json.Unmarshal([]byte(trimmed), &v) == nil
	case "yaml":
		var v any
		return yaml.Unmarshal([]byte(trimmed), &v) == nil
	default:
		return !isRefusal(trimmed)
	}
}

// isRefusal checks if a response looks like an LLM refusal.
func isRefusal(s string) bool {
	lower := strings.ToLower(s)
	prefixes := []string{
		"i cannot",
		"i can't",
		"i'm sorry",
		"i am sorry",
		"as an ai",
		"i'm not able",
		"i am not able",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

package research

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	doc := ResearchDocument{
		Source: "github-issue",
		Title:  "Fix flaky test in CI",
		Repo:   "elastic/kibana",
		Body:   "The test suite fails intermittently on Linux.",
	}

	tools := []Tool{
		{Name: "grep_code", Description: "Search code for a regex pattern", Params: "pattern (required)"},
		{Name: "read_file", Description: "Read the first 200 lines of a file", Params: "path (required)"},
	}

	prompt := buildSystemPrompt(doc, GoalSummarize, tools)

	checks := []struct {
		label    string
		contains string
	}{
		{"source", "github-issue"},
		{"title", "Fix flaky test in CI"},
		{"repo", "elastic/kibana"},
		{"tool grep_code", "grep_code"},
		{"tool read_file", "read_file"},
		{"goal summary", "summary"},
		{"json format", `"tool"`},
		{"budget", "15"},
	}

	for _, c := range checks {
		if !strings.Contains(prompt, c.contains) {
			t.Errorf("prompt missing %s (%q)", c.label, c.contains)
		}
	}
}

func TestBuildSystemPromptImplementGoal(t *testing.T) {
	doc := ResearchDocument{
		Source: "github-issue",
		Title:  "Add retry logic",
		Body:   "Need exponential backoff.",
	}

	prompt := buildSystemPrompt(doc, GoalImplement, []Tool{
		{Name: "grep_code", Description: "Search code", Params: "pattern"},
	})

	if !strings.Contains(prompt, "implementation") {
		t.Error("implement goal prompt should contain 'implementation'")
	}
}

func TestToolCallParsing(t *testing.T) {
	raw := `{"tool":"grep_code","params":{"pattern":"error"}}`
	var tc ToolCall
	if err := json.Unmarshal([]byte(raw), &tc); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if tc.Tool != "grep_code" {
		t.Errorf("expected tool grep_code, got %s", tc.Tool)
	}
	if tc.Params["pattern"] != "error" {
		t.Errorf("expected param pattern=error, got %s", tc.Params["pattern"])
	}
}

func TestToolCallParsingRejectsNonJSON(t *testing.T) {
	raw := `Here is my analysis of the issue...`
	var tc ToolCall
	err := json.Unmarshal([]byte(raw), &tc)
	if err == nil && tc.Tool != "" {
		t.Error("expected unmarshal to fail or tool to be empty for non-JSON input")
	}
}

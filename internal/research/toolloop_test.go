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
		{"goal summary", "root cause"},
		{"json format", `"tool"`},
		{"budget", "15"},
		{"relative paths", "relative"},
		{"min tool calls", "at least 3"},
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
	if tc.Params()["pattern"] != "error" {
		t.Errorf("expected param pattern=error, got %s", tc.Params()["pattern"])
	}
}

func TestToolCallParsingWithIntParams(t *testing.T) {
	raw := `{"tool": "fetch_pr", "params": {"repo": "elastic/ensemble", "number": 747}}`
	var tc ToolCall
	if err := json.Unmarshal([]byte(raw), &tc); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if tc.Tool != "fetch_pr" {
		t.Errorf("expected tool fetch_pr, got %s", tc.Tool)
	}
	p := tc.Params()
	if p["repo"] != "elastic/ensemble" {
		t.Errorf("expected repo=elastic/ensemble, got %s", p["repo"])
	}
	if p["number"] != "747" {
		t.Errorf("expected number=747, got %s", p["number"])
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

func TestCleanLLMResponseStripsThinkTags(t *testing.T) {
	input := "<think>I should call grep_code...</think>\n{\"tool\": \"grep_code\", \"params\": {\"pattern\": \"error\"}}"
	cleaned := cleanLLMResponse(input)
	if !strings.HasPrefix(cleaned, "{") {
		t.Errorf("expected JSON after stripping think tags, got: %s", cleaned)
	}
}

func TestCleanLLMResponseStripsCodeFences(t *testing.T) {
	input := "```json\n{\"tool\": \"grep_code\", \"params\": {}}\n```"
	cleaned := cleanLLMResponse(input)
	if !strings.HasPrefix(cleaned, "{") {
		t.Errorf("expected JSON after stripping fences, got: %s", cleaned)
	}
}

func TestStripCLINoise(t *testing.T) {
	input := "OpenAI Codex v0.117.0\n--------\nworkdir: /tmp\n--------\nHere is the analysis.\ntokens used\n45,282\nHere is the analysis."
	stripped := stripCLINoise(input)
	if strings.Contains(stripped, "Codex") {
		t.Errorf("expected preamble stripped, got: %s", stripped)
	}
	if strings.Contains(stripped, "tokens used") {
		t.Errorf("expected footer stripped, got: %s", stripped)
	}
	if !strings.Contains(stripped, "analysis") {
		t.Errorf("expected content preserved, got: %s", stripped)
	}
}

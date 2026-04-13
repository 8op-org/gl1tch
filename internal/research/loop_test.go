package research

import (
	"context"
	"fmt"
	"testing"
)

func scriptedLLM(responses ...string) LLMFn {
	i := 0
	return func(_ context.Context, _ string) (string, error) {
		if i >= len(responses) {
			return "", fmt.Errorf("exhausted after %d calls", len(responses))
		}
		r := responses[i]
		i++
		return r, nil
	}
}

func TestLoopHappyPath(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(&stubResearcher{
		name:     "github-prs",
		describe: "search GitHub PRs",
		evidence: Evidence{Source: "github", Title: "PR #42", Body: "fixed the bug"},
	})

	llm := scriptedLLM(
		// Plan response
		`["github-prs"]`,
		// Draft response
		"The bug was fixed in PR #42.",
		// Critique response
		`[{"text": "The bug was fixed in PR #42", "label": "grounded"}]`,
		// Judge response
		"0.9",
	)

	loop := NewLoop(reg, llm)
	result, err := loop.Run(context.Background(), ResearchQuery{Question: "What happened?"}, DefaultBudget())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Draft == "" {
		t.Error("expected non-empty draft")
	}
	if result.Reason != ReasonAccepted {
		t.Errorf("expected ReasonAccepted, got %q", result.Reason)
	}
}

func TestLoopEmptyPlanShortCircuits(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(&stubResearcher{name: "unused", describe: "unused"})

	llm := scriptedLLM(`[]`)

	loop := NewLoop(reg, llm)
	result, err := loop.Run(context.Background(), ResearchQuery{Question: "Nothing?"}, DefaultBudget())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Reason != ReasonUnscored {
		t.Errorf("expected ReasonUnscored, got %q", result.Reason)
	}
}

func TestLoopRejectsNilLLM(t *testing.T) {
	reg := NewRegistry()
	loop := NewLoop(reg, nil)
	_, err := loop.Run(context.Background(), ResearchQuery{Question: "test"}, DefaultBudget())
	if err == nil {
		t.Fatal("expected error for nil LLM")
	}
}

func TestParsePlanTolerant(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"json array", `["alpha", "beta"]`, []string{"alpha", "beta"}},
		{"preamble", `Here are my picks: ["alpha"]`, []string{"alpha"}},
		{"backslash escaped", `[\"alpha\", \"beta\"]`, []string{"alpha", "beta"}},
		{"bare identifiers", `[alpha, beta-test]`, []string{"alpha", "beta-test"}},
		{"empty", `[]`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlan(tt.input)
			if tt.want == nil {
				if err != nil && len(got) != 0 {
					// empty is fine either way
				}
				if len(got) != 0 {
					t.Errorf("ParsePlan(%q) = %v, want empty", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePlan(%q): %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParsePlan(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParsePlan(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLoopWithEventSink(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(&stubResearcher{
		name:     "test-src",
		describe: "test source",
		evidence: Evidence{Source: "test", Body: "data"},
	})

	llm := scriptedLLM(
		`["test-src"]`,
		"Answer based on evidence.",
		`[{"text": "Answer based on evidence", "label": "grounded"}]`,
		"0.9",
	)

	sink := &MemoryEventSink{}
	loop := NewLoop(reg, llm).WithEventSink(sink)
	_, err := loop.Run(context.Background(), ResearchQuery{Question: "test?"}, DefaultBudget())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	events := sink.Events()
	if len(events) == 0 {
		t.Error("expected events to be emitted")
	}
}

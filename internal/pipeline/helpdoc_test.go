package pipeline

import (
	"sort"
	"testing"
)

func TestExtractImplicitRefs_FlatSteps(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", Run: "echo ~param.topic"},
			{ID: "b", Run: "echo ~param.audience ~param.topic"},
		},
	}
	params, usesInput := ExtractImplicitRefs(w)
	sort.Strings(params)
	want := []string{"audience", "topic"}
	if !equalStrings(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
	if usesInput {
		t.Errorf("usesInput = true, want false")
	}
}

func TestExtractImplicitRefs_DetectsInput(t *testing.T) {
	w := &Workflow{
		Steps: []Step{{ID: "a", Run: "echo ~input"}},
	}
	_, usesInput := ExtractImplicitRefs(w)
	if !usesInput {
		t.Errorf("usesInput = false, want true")
	}
}

func TestExtractImplicitRefs_InsideForms(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", Run: `echo ~(or param.topic "default") ~(upper param.audience)`},
		},
	}
	params, _ := ExtractImplicitRefs(w)
	sort.Strings(params)
	want := []string{"audience", "topic"}
	if !equalStrings(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
}

func TestExtractImplicitRefs_SkipsQuoteFirstArg(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", Run: `~(step diff) ~(stepfile foo)`},
		},
	}
	params, _ := ExtractImplicitRefs(w)
	if len(params) != 0 {
		t.Errorf("params should be empty (step/stepfile are quote-first-arg), got %v", params)
	}
}

func TestExtractImplicitRefs_LLMPrompt(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", LLM: &LLMStep{Prompt: "Summarise ~param.topic for ~param.audience"}},
		},
	}
	params, _ := ExtractImplicitRefs(w)
	sort.Strings(params)
	if !equalStrings(params, []string{"audience", "topic"}) {
		t.Errorf("params = %v", params)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

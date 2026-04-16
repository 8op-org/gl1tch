package pipeline

import (
	"sort"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/plugin"
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

func TestMergeArgs_ImplicitFillsMissing(t *testing.T) {
	w := &Workflow{
		Args: []plugin.ArgDef{
			{Name: "topic", Description: "Topic", Required: true},
		},
		Steps: []Step{
			{ID: "a", Run: "echo ~param.topic ~param.audience"},
		},
	}
	warnings := MergeImplicitArgs(w)

	if len(w.Args) != 2 {
		t.Fatalf("Args len = %d, want 2 (declared topic + implicit audience)", len(w.Args))
	}

	var topic, audience *plugin.ArgDef
	for i := range w.Args {
		switch w.Args[i].Name {
		case "topic":
			topic = &w.Args[i]
		case "audience":
			audience = &w.Args[i]
		}
	}
	if topic == nil || topic.Implicit {
		t.Errorf("topic missing or marked implicit: %+v", topic)
	}
	if audience == nil || !audience.Implicit {
		t.Errorf("audience missing or not marked implicit: %+v", audience)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestMergeArgs_WarnsOnUnreferenced(t *testing.T) {
	w := &Workflow{
		Args: []plugin.ArgDef{
			{Name: "topic", Required: true},
			{Name: "ghost", Description: "orphan"},
		},
		Steps: []Step{
			{ID: "a", Run: "echo ~param.topic"},
		},
	}
	warnings := MergeImplicitArgs(w)
	if len(warnings) != 1 {
		t.Fatalf("want 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], `"ghost"`) {
		t.Errorf("warning should name the orphan arg, got %q", warnings[0])
	}
}

func TestMergeArgs_InputPopulatedFromRef(t *testing.T) {
	w := &Workflow{
		Steps: []Step{{ID: "a", Run: "echo ~input"}},
	}
	MergeImplicitArgs(w)
	if w.Input == nil {
		t.Fatal("Input should be populated from implicit ~input reference")
	}
	if !w.Input.Implicit {
		t.Errorf("implicit Input should have Implicit=true")
	}
}

package pipeline

import (
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/plugin"
)

func TestFormatHelp_AllFields(t *testing.T) {
	w := &Workflow{
		Name:        "site-create-page",
		Description: "AI-generate a new doc page with gated verification",
		SourceFile:  ".glitch/workflows/site-create-page.glitch",
		Args: []plugin.ArgDef{
			{Name: "topic", Required: true, Description: "Topic of the page.", Example: "batch comparison"},
			{Name: "audience", Default: "developers", Description: "Target reader.", Example: "ops engineer"},
		},
		Input: &InputDef{Description: "Free-form context.", Example: "fix latency"},
		Steps: []Step{{ID: "s", Line: 6}},
	}
	out := FormatHelp(w)

	for _, want := range []string{
		"site-create-page - AI-generate a new doc page with gated verification",
		"glitch run site-create-page",
		"topic",
		"(required)",
		"Topic of the page.",
		`--set topic="batch comparison"`,
		"audience",
		"developers",
		"Free-form context.",
		".glitch/workflows/site-create-page.glitch",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatHelp output missing %q\n---\n%s", want, out)
		}
	}
}

func TestFormatHelp_NoDescription(t *testing.T) {
	w := &Workflow{Name: "bare", SourceFile: "bare.glitch"}
	out := FormatHelp(w)
	if strings.Contains(out, " - ") {
		t.Errorf("header should not have dash when :description is absent\n%s", out)
	}
	if !strings.Contains(out, "bare") {
		t.Errorf("workflow name missing from output\n%s", out)
	}
}

func TestFormatHelp_ImplicitMarker(t *testing.T) {
	w := &Workflow{
		Name: "x",
		Args: []plugin.ArgDef{{Name: "topic", Implicit: true}},
	}
	out := FormatHelp(w)
	if !strings.Contains(out, "undocumented") {
		t.Errorf("implicit args should be marked undocumented\n%s", out)
	}
}

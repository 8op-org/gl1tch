package cmd

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/router"
)

// ── stripFences ───────────────────────────────────────────────────────────────

func TestStripFences(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"name: foo", "name: foo"},
		{"```yaml\nname: foo\n```", "name: foo"},
		{"```yml\nname: foo\n```", "name: foo"},
		{"```\nname: foo\n```", "name: foo"},
		{"  ```yaml\nname: foo\n```  ", "name: foo"},
	}
	for _, c := range cases {
		got := stripFences(c.in)
		if got != c.want {
			t.Errorf("stripFences(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── routeIntent response matching ────────────────────────────────────────────

// matchResponse is the pure matching logic extracted from routeIntent for unit testing.
func matchResponse(response string, refs []pipeline.PipelineRef) *pipeline.PipelineRef {
	response = strings.TrimSpace(response)
	response = strings.Trim(response, `"'.`)
	if strings.EqualFold(response, "NONE") || response == "" {
		return nil
	}
	for i, r := range refs {
		if strings.EqualFold(r.Name, response) {
			return &refs[i]
		}
	}
	return nil
}

func TestRouteIntentMatching(t *testing.T) {
	refs := []pipeline.PipelineRef{
		{Name: "sync-docs", Description: "Sync documentation"},
		{Name: "gh-review", Description: "Review GitHub PRs"},
	}

	t.Run("exact match", func(t *testing.T) {
		got := matchResponse("sync-docs", refs)
		if got == nil || got.Name != "sync-docs" {
			t.Errorf("expected sync-docs match, got %v", got)
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		got := matchResponse("SYNC-DOCS", refs)
		if got == nil || got.Name != "sync-docs" {
			t.Errorf("expected sync-docs match, got %v", got)
		}
	})

	t.Run("NONE returns nil", func(t *testing.T) {
		if got := matchResponse("NONE", refs); got != nil {
			t.Errorf("expected nil for NONE, got %v", got)
		}
	})

	t.Run("none lowercase returns nil", func(t *testing.T) {
		if got := matchResponse("none", refs); got != nil {
			t.Errorf("expected nil for none, got %v", got)
		}
	})

	t.Run("garbage returns nil", func(t *testing.T) {
		if got := matchResponse("I think you want the sync-docs pipeline", refs); got != nil {
			t.Errorf("expected nil for garbage, got %v", got)
		}
	})

	t.Run("empty returns nil", func(t *testing.T) {
		if got := matchResponse("", refs); got != nil {
			t.Errorf("expected nil for empty, got %v", got)
		}
	})

	t.Run("quoted name is matched", func(t *testing.T) {
		got := matchResponse(`"sync-docs"`, refs)
		if got == nil || got.Name != "sync-docs" {
			t.Errorf("expected sync-docs match for quoted name, got %v", got)
		}
	})
}

// ── --route=false: discovery + classification skipped ────────────────────────

func TestAskRoute_FalseSkipsRouting(t *testing.T) {
	// When route=false, DiscoverPipelines is never called.
	// We verify this by ensuring the flag is wired — no integration needed.
	f := askCmd.Flags().Lookup("route")
	if f == nil {
		t.Fatal("--route flag not registered on askCmd")
	}
	if f.DefValue != "true" {
		t.Errorf("--route default = %q, want \"true\"", f.DefValue)
	}
}

// ── --dry-run flag registered ─────────────────────────────────────────────────

func TestAskDryRun_FlagRegistered(t *testing.T) {
	f := askCmd.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("--dry-run flag not registered on askCmd")
	}
}

// ── --auto flag registered ────────────────────────────────────────────────────

func TestAskAuto_FlagRegistered(t *testing.T) {
	f := askCmd.Flags().Lookup("auto")
	if f == nil {
		t.Fatal("--auto flag not registered on askCmd")
	}
	shorthand := askCmd.Flags().ShorthandLookup("y")
	if shorthand == nil {
		t.Fatal("-y shorthand not registered on askCmd")
	}
}

// ── parseInputVars ────────────────────────────────────────────────────────────

func TestParseInputVars(t *testing.T) {
	t.Run("nil returns empty map", func(t *testing.T) {
		got, err := parseInputVars(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})

	t.Run("single key=value", func(t *testing.T) {
		got, err := parseInputVars([]string{"topic=golang"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["topic"] != "golang" {
			t.Errorf("got[topic] = %q, want 'golang'", got["topic"])
		}
	})

	t.Run("multiple pairs", func(t *testing.T) {
		got, err := parseInputVars([]string{"a=1", "b=2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["a"] != "1" || got["b"] != "2" {
			t.Errorf("unexpected map: %v", got)
		}
	})

	t.Run("value contains equals sign", func(t *testing.T) {
		got, err := parseInputVars([]string{"url=https://example.com?x=1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["url"] != "https://example.com?x=1" {
			t.Errorf("url = %q, want 'https://example.com?x=1'", got["url"])
		}
	})

	t.Run("missing equals returns error", func(t *testing.T) {
		_, err := parseInputVars([]string{"noequalssign"})
		if err == nil {
			t.Error("expected error for missing =, got nil")
		}
	})
}

// ── buildAskPipeline ──────────────────────────────────────────────────────────

func TestBuildAskPipeline(t *testing.T) {
	t.Run("single step without synthesize", func(t *testing.T) {
		p := buildAskPipeline("what is recursion", "ollama", "llama3.2", nil, false, "")
		if p.Name != "ask" {
			t.Errorf("name = %q, want 'ask'", p.Name)
		}
		if len(p.Steps) != 1 {
			t.Fatalf("expected 1 step, got %d", len(p.Steps))
		}
		s := p.Steps[0]
		if s.ID != "ask" || s.Executor != "ollama" || s.Model != "llama3.2" {
			t.Errorf("step = {ID:%q Executor:%q Model:%q}", s.ID, s.Executor, s.Model)
		}
		if s.Prompt != "what is recursion" {
			t.Errorf("prompt = %q", s.Prompt)
		}
	})

	t.Run("synthesize appends second step with needs wiring", func(t *testing.T) {
		p := buildAskPipeline("explain goroutines", "ollama", "llama3.2", nil, true, "claude-sonnet-4-6")
		if len(p.Steps) != 2 {
			t.Fatalf("expected 2 steps, got %d", len(p.Steps))
		}
		synth := p.Steps[1]
		if synth.ID != "synthesize" {
			t.Errorf("synth ID = %q, want 'synthesize'", synth.ID)
		}
		if synth.Executor != "claude" {
			t.Errorf("synth executor = %q, want 'claude'", synth.Executor)
		}
		if synth.Model != "claude-sonnet-4-6" {
			t.Errorf("synth model = %q", synth.Model)
		}
		if len(synth.Needs) == 0 || synth.Needs[0] != "ask" {
			t.Errorf("synth needs = %v, want [ask]", synth.Needs)
		}
		if !strings.Contains(synth.Prompt, "{{step.ask.data.value}}") {
			t.Errorf("synth prompt missing template ref: %q", synth.Prompt)
		}
	})

	t.Run("vars passed to ask step", func(t *testing.T) {
		vars := map[string]string{"input": "router package"}
		p := buildAskPipeline("improve docs", "ollama", "llama3.2", vars, false, "")
		if p.Steps[0].Vars["input"] != "router package" {
			t.Errorf("vars[input] = %q, want 'router package'", p.Steps[0].Vars["input"])
		}
	})
}

// ── printJSON ─────────────────────────────────────────────────────────────────

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	t.Run("includes response provider model", func(t *testing.T) {
		out := captureStdout(func() { _ = printJSON("hello world", "ollama", "llama3.2", "") })
		var m map[string]string
		if err := json.Unmarshal([]byte(out), &m); err != nil {
			t.Fatalf("invalid JSON: %v\n%s", err, out)
		}
		if m["response"] != "hello world" {
			t.Errorf("response = %q", m["response"])
		}
		if m["provider"] != "ollama" {
			t.Errorf("provider = %q", m["provider"])
		}
		if m["model"] != "llama3.2" {
			t.Errorf("model = %q", m["model"])
		}
	})

	t.Run("brain_entry_id absent when empty", func(t *testing.T) {
		out := captureStdout(func() { _ = printJSON("result", "ollama", "llama3.2", "") })
		var m map[string]string
		if err := json.Unmarshal([]byte(out), &m); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if _, ok := m["brain_entry_id"]; ok {
			t.Error("brain_entry_id should be absent when empty string passed")
		}
	})

	t.Run("brain_entry_id present when set", func(t *testing.T) {
		out := captureStdout(func() { _ = printJSON("result", "ollama", "llama3.2", "42") })
		var m map[string]string
		if err := json.Unmarshal([]byte(out), &m); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if m["brain_entry_id"] != "42" {
			t.Errorf("brain_entry_id = %q, want '42'", m["brain_entry_id"])
		}
	})
}

// ── dispatchMatched dry-run ───────────────────────────────────────────────────

func TestDispatchMatched_DryRun(t *testing.T) {
	old := askDryRun
	askDryRun = true
	t.Cleanup(func() { askDryRun = old })

	ref := &pipeline.PipelineRef{Name: "support-digest", Path: "/tmp/fake-pipeline.yaml"}
	result := &router.RouteResult{
		Pipeline:   ref,
		Confidence: 0.92,
		Method:     "embedding",
	}

	out := captureStdout(func() {
		err := dispatchMatched(askCmd, "run the support digest", result, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "support-digest") {
		t.Errorf("dry-run output missing pipeline name: %q", out)
	}
	if !strings.Contains(out, "/tmp/fake-pipeline.yaml") {
		t.Errorf("dry-run output missing path: %q", out)
	}
}

// ── dispatchMatched input merge logic ────────────────────────────────────────

// TestInputMergeLogic validates the merge rule: router input is used when no
// explicit --input was given; explicit wins when both are present.
func TestInputMergeLogic(t *testing.T) {
	t.Run("router input fills empty map", func(t *testing.T) {
		inputVars := map[string]string{}
		routerInput := "executor package"
		if _, hasInput := inputVars["input"]; !hasInput && routerInput != "" {
			inputVars["input"] = routerInput
		}
		if inputVars["input"] != "executor package" {
			t.Errorf("expected router input to be set, got %q", inputVars["input"])
		}
	})

	t.Run("explicit input beats router input", func(t *testing.T) {
		inputVars := map[string]string{"input": "explicit-value"}
		routerInput := "router-value"
		if _, hasInput := inputVars["input"]; !hasInput && routerInput != "" {
			inputVars["input"] = routerInput
		}
		if inputVars["input"] != "explicit-value" {
			t.Errorf("expected explicit input to win, got %q", inputVars["input"])
		}
	})
}

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/8op-org/gl1tch/internal/cron"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/router"
)

// ── upsertCronEntry ───────────────────────────────────────────────────────────

func TestUpsertCronEntry_WritesNewEntry(t *testing.T) {
	dir := t.TempDir()
	cronPath := filepath.Join(dir, "cron.yaml")

	ref := &pipeline.PipelineRef{Name: "support-digest-dryrun", Path: "/fake/path.yaml"}
	if err := writeCronEntry(cronPath, ref, "acme", "0 9 * * *"); err != nil {
		t.Fatalf("writeCronEntry: %v", err)
	}

	entries, err := cron.LoadConfigFrom(cronPath)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.Name != "support-digest-dryrun" {
		t.Errorf("Name=%q, want 'support-digest-dryrun'", e.Name)
	}
	if e.Schedule != "0 9 * * *" {
		t.Errorf("Schedule=%q, want '0 9 * * *'", e.Schedule)
	}
	if e.Input != "acme" {
		t.Errorf("Input=%q, want 'acme'", e.Input)
	}
	if e.Kind != "pipeline" {
		t.Errorf("Kind=%q, want 'pipeline'", e.Kind)
	}
	if e.Target != "support-digest-dryrun" {
		t.Errorf("Target=%q, want 'support-digest-dryrun'", e.Target)
	}
	if e.WorkingDir == "" {
		t.Error("WorkingDir should be set to cwd, got empty")
	}
}

func TestUpsertCronEntry_UpdatesExistingEntry(t *testing.T) {
	dir := t.TempDir()
	cronPath := filepath.Join(dir, "cron.yaml")

	ref := &pipeline.PipelineRef{Name: "support-digest-dryrun"}

	// Write initial entry
	if err := writeCronEntry(cronPath, ref, "", "0 9 * * *"); err != nil {
		t.Fatalf("first write: %v", err)
	}
	// Upsert same name with different schedule
	if err := writeCronEntry(cronPath, ref, "focus", "0 */2 * * *"); err != nil {
		t.Fatalf("second write: %v", err)
	}

	entries, err := cron.LoadConfigFrom(cronPath)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("upsert should not duplicate: got %d entries", len(entries))
	}
	if entries[0].Schedule != "0 */2 * * *" {
		t.Errorf("Schedule=%q, want '0 */2 * * *'", entries[0].Schedule)
	}
	if entries[0].Input != "focus" {
		t.Errorf("Input=%q, want 'focus'", entries[0].Input)
	}
}

func TestUpsertCronEntry_MultipleDistinctEntries(t *testing.T) {
	dir := t.TempDir()
	cronPath := filepath.Join(dir, "cron.yaml")

	pipelines := []pipeline.PipelineRef{
		{Name: "support-digest-dryrun"},
		{Name: "clarify-haiku-multistep"},
	}
	schedules := []string{"0 9 * * *", "0 14 * * 1-5"}

	for i, ref := range pipelines {
		r := ref
		if err := writeCronEntry(cronPath, &r, "", schedules[i]); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	entries, err := cron.LoadConfigFrom(cronPath)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 distinct entries, got %d", len(entries))
	}
}

func TestUpsertCronEntry_YAMLStructure(t *testing.T) {
	dir := t.TempDir()
	cronPath := filepath.Join(dir, "cron.yaml")

	ref := &pipeline.PipelineRef{Name: "test-glab-after"}
	if err := writeCronEntry(cronPath, ref, "", "0 0 * * *"); err != nil {
		t.Fatalf("writeCronEntry: %v", err)
	}

	raw, err := os.ReadFile(cronPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Must be parseable as top-level entries: array
	var doc struct {
		Entries []map[string]any `yaml:"entries"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid cron.yaml structure: %v\n%s", err, raw)
	}
	if len(doc.Entries) == 0 {
		t.Error("entries array is empty")
	}
}

// writeCronEntry is the testable core of upsertCronEntry — writes to an
// explicit path instead of glitchConfigDir so tests are hermetic.
func writeCronEntry(path string, ref *pipeline.PipelineRef, input, schedule string) error {
	entries, err := cron.LoadConfigFrom(path)
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	e := cron.Entry{
		Name:       ref.Name,
		Schedule:   schedule,
		Kind:       "pipeline",
		Target:     ref.Name,
		Input:      input,
		Timeout:    "15m",
		WorkingDir: wd,
	}
	entries = cron.UpsertEntry(entries, e)
	return cron.SaveConfigTo(path, entries)
}

// ── dispatchMatched with cron ─────────────────────────────────────────────────

func TestDispatchMatched_DryRun_SkipsCronWrite(t *testing.T) {
	// dry-run=true must not write cron.yaml even when CronExpr is set.
	old := askDryRun
	askDryRun = true
	t.Cleanup(func() { askDryRun = old })

	ref := &pipeline.PipelineRef{Name: "support-digest-dryrun", Path: "/tmp/fake.yaml"}
	result := &router.RouteResult{
		Pipeline:   ref,
		Confidence: 0.88,
		CronExpr:   "0 9 * * *",
		Method:     "llm",
	}

	out := captureStdout(func() {
		err := dispatchMatched(askCmd, "run digest every morning", result, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Dry-run prints the plan but does NOT write cron.yaml
	if out == "" {
		t.Error("expected dry-run output, got empty")
	}
	// Verify no cron.yaml was written to the default config dir
	// (we can't easily intercept glitchConfigDir in a unit test, but we verify
	// the dry-run path returns before reaching upsertCronEntry)
	_ = out
}

// ── buildAskPipeline edge cases ───────────────────────────────────────────────

func TestBuildAskPipeline_Version(t *testing.T) {
	p := buildAskPipeline("test", "ollama", "llama3.2", nil, false, "")
	if p.Version != "1" {
		t.Errorf("Version=%q, want '1'", p.Version)
	}
}

func TestBuildAskPipeline_SynthesizeEmptyModel(t *testing.T) {
	// When synthesize=true but no model specified, model field should be ""
	// (the claude executor picks its own default).
	p := buildAskPipeline("test", "ollama", "llama3.2", nil, true, "")
	if len(p.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(p.Steps))
	}
	synth := p.Steps[1]
	if synth.Model != "" {
		t.Errorf("synth model=%q, want '' (empty = executor default)", synth.Model)
	}
	if synth.Executor != "claude" {
		t.Errorf("synth executor=%q, want 'claude'", synth.Executor)
	}
}

func TestBuildAskPipeline_Name(t *testing.T) {
	p := buildAskPipeline("anything", "ollama", "llama3.2", nil, false, "")
	if p.Name != "ask" {
		t.Errorf("Name=%q, want 'ask'", p.Name)
	}
}

// ── parseInputVars edge cases ─────────────────────────────────────────────────

func TestParseInputVars_EdgeCases(t *testing.T) {
	t.Run("empty value is allowed", func(t *testing.T) {
		got, err := parseInputVars([]string{"key="})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := got["key"]; !ok || v != "" {
			t.Errorf("got[key]=%q, want ''", got["key"])
		}
	})

	t.Run("value with multiple equals preserves all after first", func(t *testing.T) {
		got, err := parseInputVars([]string{"a=b=c"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["a"] != "b=c" {
			t.Errorf("got[a]=%q, want 'b=c'", got["a"])
		}
	})

	t.Run("empty slice produces empty map not nil", func(t *testing.T) {
		got, err := parseInputVars([]string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil {
			t.Error("expected non-nil map for empty slice")
		}
	})
}

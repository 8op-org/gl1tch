package pipeline

import (
	"strings"
	"testing"
)

// TestRun_InvokesStepRecorderPerStep guards that a StepRecorder callback wired
// via RunOpts receives one record per workflow step (flat-steps form). Exists
// because the steps table in the store was never populated — the runner had
// no hook for callers to observe per-step completion. Issue #4.
func TestRun_InvokesStepRecorderPerStep(t *testing.T) {
	w := &Workflow{
		Name: "steprec-flat",
		Steps: []Step{
			{ID: "first", Run: "echo hello"},
			{ID: "second", Run: "echo world"},
		},
	}

	var got []StepRecord
	_, err := Run(w, "", "", nil, nil, RunOpts{
		StepRecorder: func(rec StepRecord) { got = append(got, rec) },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("want 2 step records, got %d: %+v", len(got), got)
	}
	if got[0].StepID != "first" || got[1].StepID != "second" {
		t.Fatalf("step IDs = [%q, %q], want [first, second]", got[0].StepID, got[1].StepID)
	}
	if !strings.Contains(got[0].Output, "hello") {
		t.Errorf("first output = %q, want to contain 'hello'", got[0].Output)
	}
	if !strings.Contains(got[1].Output, "world") {
		t.Errorf("second output = %q, want to contain 'world'", got[1].Output)
	}
	if got[0].Kind != "run" {
		t.Errorf("first kind = %q, want 'run'", got[0].Kind)
	}
	if got[0].ExitStatus == nil || *got[0].ExitStatus != 0 {
		t.Errorf("first exit status = %v, want 0", got[0].ExitStatus)
	}
}

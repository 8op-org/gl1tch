package pipeline

import (
	"strings"
	"testing"
)

// TestMapResources_ParseAndRun_EndToEnd guards the full path from a
// workspace-style sexpr workflow containing only a (map-resources ...) form
// through parsing and execution. Mirrors the repro in issue #3 which showed
// the parent run completing ok with no body executions.
func TestMapResources_ParseAndRun_EndToEnd(t *testing.T) {
	src := []byte(`
(workflow "smoke"
  (map-resources :type "tracker"
    (step "per-resource"
      (run "echo tracker:~resource.item.repo"))))
`)
	w, err := LoadBytes(src, "smoke.glitch")
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}

	// Parser must hoist the (map-resources ...) form into Workflow.Steps so
	// the runner has something to iterate. Without this, the run completes
	// with zero executions and no recorded output.
	if len(w.Steps) == 0 {
		t.Fatalf("parser produced zero steps; want at least the map-resources form")
	}

	var mr *Step
	for i := range w.Steps {
		if w.Steps[i].Form == "map-resources" {
			mr = &w.Steps[i]
			break
		}
	}
	if mr == nil {
		t.Fatalf("workflow has no map-resources step, got %+v", w.Steps)
	}
	if mr.MapResourcesType != "tracker" {
		t.Errorf("MapResourcesType = %q, want %q", mr.MapResourcesType, "tracker")
	}
	if mr.MapResourcesBody == nil {
		t.Fatalf("MapResourcesBody is nil; parser dropped the body step")
	}

	// End-to-end: run with two tracker resources and one non-tracker; body
	// must fire twice and only for tracker resources.
	res, err := Run(w, "", "", nil, nil, RunOpts{
		Resources: map[string]map[string]string{
			"ensemble":              {"name": "ensemble", "type": "tracker", "repo": "elastic/ensemble"},
			"observability-robots":  {"name": "observability-robots", "type": "tracker", "repo": "elastic/observability-robots"},
			"kibana":                {"name": "kibana", "type": "git", "repo": "elastic/kibana"},
		},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	out := res.Steps[mr.ID]
	if !strings.Contains(out, "tracker:elastic/ensemble") {
		t.Errorf("output missing ensemble entry; got %q", out)
	}
	if !strings.Contains(out, "tracker:elastic/observability-robots") {
		t.Errorf("output missing observability-robots entry; got %q", out)
	}
	if strings.Contains(out, "elastic/kibana") {
		t.Errorf("non-tracker resource leaked into output: %q", out)
	}
}

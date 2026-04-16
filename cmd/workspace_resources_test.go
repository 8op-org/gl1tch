package cmd

import (
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace"
)

// TestResourceBindings_IncludesType guards that the resource map passed to
// pipeline.RunOpts.Resources carries the resource "type" field. Without it,
// (map-resources :type "tracker" ...) filters out every resource because
// res["type"] is empty. Root cause for issue #3.
func TestResourceBindings_IncludesType(t *testing.T) {
	ws := &workspace.Workspace{
		Resources: []workspace.Resource{
			{Name: "ensemble", Type: "tracker", Repo: "elastic/ensemble"},
			{Name: "kibana", Type: "git", URL: "https://github.com/elastic/kibana"},
			{Name: "notes", Type: "local", Path: "/tmp/notes"},
		},
	}

	got := ResourceBindings(ws, "/ws")

	for _, name := range []string{"ensemble", "kibana", "notes"} {
		m, ok := got[name]
		if !ok {
			t.Fatalf("binding for %q missing", name)
		}
		if m["type"] == "" {
			t.Errorf("binding for %q missing 'type' field; got %+v", name, m)
		}
	}

	if got["ensemble"]["type"] != "tracker" {
		t.Errorf("ensemble type = %q, want tracker", got["ensemble"]["type"])
	}
	if got["ensemble"]["repo"] != "elastic/ensemble" {
		t.Errorf("ensemble repo = %q, want elastic/ensemble", got["ensemble"]["repo"])
	}
}

package pipeline

import (
	"sort"
	"strings"
	"testing"
)

func TestMapResourcesBasic(t *testing.T) {
	w := &Workflow{
		Name: "loop",
		Steps: []Step{
			{
				ID:               "each",
				Form:             "map-resources",
				MapResourcesType: "",
				MapResourcesBody: &Step{
					ID:  "inner",
					Run: "echo ~resource.item.name:~resource.item.path",
				},
			},
		},
	}
	res, err := Run(w, "", "", map[string]string{}, nil, RunOpts{
		Resources: map[string]map[string]string{
			"alpha": {"name": "alpha", "type": "git", "path": "/a"},
			"beta":  {"name": "beta", "type": "local", "path": "/b"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Each shell invocation emits one line; Run joins them with newlines.
	parts := strings.Split(strings.TrimSpace(res.Steps["each"]), "\n")
	sort.Strings(parts)
	if len(parts) != 2 || parts[0] != "alpha:/a" || parts[1] != "beta:/b" {
		t.Fatalf("got %+v", parts)
	}
}

func TestMapResourcesTypeFilter(t *testing.T) {
	w := &Workflow{
		Name: "loop",
		Steps: []Step{
			{
				ID:               "each",
				Form:             "map-resources",
				MapResourcesType: "git",
				MapResourcesBody: &Step{
					ID:  "inner",
					Run: "echo ~resource.item.name",
				},
			},
		},
	}
	res, err := Run(w, "", "", map[string]string{}, nil, RunOpts{
		Resources: map[string]map[string]string{
			"alpha": {"name": "alpha", "type": "git", "path": "/a"},
			"beta":  {"name": "beta", "type": "local", "path": "/b"},
			"gamma": {"name": "gamma", "type": "git", "path": "/g"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(res.Steps["each"]), "\n")
	sort.Strings(lines)
	if len(lines) != 2 || lines[0] != "alpha" || lines[1] != "gamma" {
		t.Fatalf("expected alpha+gamma, got %+v", lines)
	}
}

func TestMapResourcesEmpty(t *testing.T) {
	w := &Workflow{
		Name: "loop",
		Steps: []Step{{
			ID:               "each",
			Form:             "map-resources",
			MapResourcesBody: &Step{ID: "inner", Run: "echo x"},
		}},
	}
	res, err := Run(w, "", "", map[string]string{}, nil, RunOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(res.Steps["each"]) != "" {
		t.Fatalf("empty resources should yield empty output, got %q", res.Steps["each"])
	}
}

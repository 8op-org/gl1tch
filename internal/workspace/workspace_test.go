package workspace

import (
	"testing"
)

const fullExample = `
(workspace "stokagent"
  :description "Cross-repo research and automation"
  :owner "adam"
  (repos
    "elastic/observability-robots"
    "elastic/ensemble")
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"))
`

func TestParseWorkspace_Full(t *testing.T) {
	w, err := ParseFile([]byte(fullExample))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Name != "stokagent" {
		t.Errorf("Name: got %q, want %q", w.Name, "stokagent")
	}
	if w.Description != "Cross-repo research and automation" {
		t.Errorf("Description: got %q, want %q", w.Description, "Cross-repo research and automation")
	}
	if w.Owner != "adam" {
		t.Errorf("Owner: got %q, want %q", w.Owner, "adam")
	}

	wantRepos := []string{"elastic/observability-robots", "elastic/ensemble"}
	if len(w.Repos) != len(wantRepos) {
		t.Fatalf("Repos len: got %d, want %d", len(w.Repos), len(wantRepos))
	}
	for i, r := range wantRepos {
		if w.Repos[i] != r {
			t.Errorf("Repos[%d]: got %q, want %q", i, w.Repos[i], r)
		}
	}

	if w.Defaults.Model != "qwen2.5:7b" {
		t.Errorf("Defaults.Model: got %q, want %q", w.Defaults.Model, "qwen2.5:7b")
	}
	if w.Defaults.Provider != "ollama" {
		t.Errorf("Defaults.Provider: got %q, want %q", w.Defaults.Provider, "ollama")
	}
}

func TestParseWorkspace_Minimal(t *testing.T) {
	src := `(workspace "minimal")`
	w, err := ParseFile([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Name != "minimal" {
		t.Errorf("Name: got %q, want %q", w.Name, "minimal")
	}
	if w.Repos == nil {
		t.Error("Repos must not be nil")
	}
	if len(w.Repos) != 0 {
		t.Errorf("Repos: got %v, want empty slice", w.Repos)
	}
	if w.Description != "" {
		t.Errorf("Description: got %q, want empty", w.Description)
	}
	if w.Owner != "" {
		t.Errorf("Owner: got %q, want empty", w.Owner)
	}
}

func TestParseWorkspace_Elasticsearch(t *testing.T) {
	src := []byte(`
(workspace "test"
  :description "test workspace"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://es.internal:9200"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.Elasticsearch != "http://es.internal:9200" {
		t.Fatalf("elasticsearch = %q, want http://es.internal:9200", w.Defaults.Elasticsearch)
	}
}

func TestParseWorkspace_ElasticsearchDefault(t *testing.T) {
	src := []byte(`
(workspace "test"
  (defaults :model "qwen2.5:7b"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.Elasticsearch != "" {
		t.Fatalf("elasticsearch should be empty by default, got %q", w.Defaults.Elasticsearch)
	}
}

func TestParseWorkspace_NoWorkspaceForm(t *testing.T) {
	src := `(workflow "not-a-workspace")`
	_, err := ParseFile([]byte(src))
	if err == nil {
		t.Fatal("expected error for missing (workspace ...) form, got nil")
	}
}

func TestParseWorkspace_Params(t *testing.T) {
	src := []byte(`
(workspace "test"
  :description "test workspace"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    (params
      :repo "elastic/kibana"
      :results-dir "results/kibana")))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Defaults.Model != "qwen2.5:7b" {
		t.Errorf("Model: got %q, want %q", w.Defaults.Model, "qwen2.5:7b")
	}
	if len(w.Defaults.Params) != 2 {
		t.Fatalf("Params len: got %d, want 2", len(w.Defaults.Params))
	}
	if w.Defaults.Params["repo"] != "elastic/kibana" {
		t.Errorf("Params[repo]: got %q, want %q", w.Defaults.Params["repo"], "elastic/kibana")
	}
	if w.Defaults.Params["results-dir"] != "results/kibana" {
		t.Errorf("Params[results-dir]: got %q, want %q", w.Defaults.Params["results-dir"], "results/kibana")
	}
}

func TestParseWorkspace_NoParams(t *testing.T) {
	src := []byte(`
(workspace "test"
  (defaults :model "qwen2.5:7b"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.Params == nil {
		t.Fatal("Params should be initialized to empty map, not nil")
	}
	if len(w.Defaults.Params) != 0 {
		t.Errorf("Params should be empty, got %v", w.Defaults.Params)
	}
}

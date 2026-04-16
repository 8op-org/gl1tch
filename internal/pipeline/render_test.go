package pipeline

import "testing"

func TestRenderStringFunctions(t *testing.T) {
	scope := NewScope()
	scope.SetParam("repo", "elastic/elasticsearch")
	scope.SetParam("label", "  Bug Fix  ")

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{"split then join newline", `~(split "/" param.repo)`, "elastic\nelasticsearch"},
		{"first of split", `~(first (split "/" param.repo))`, "elastic"},
		{"last of split", `~(last (split "/" param.repo))`, "elasticsearch"},
		{"join", `~(join "-" (split "/" param.repo))`, "elastic-elasticsearch"},
		{"upper", `~(upper param.repo)`, "ELASTIC/ELASTICSEARCH"},
		{"lower literal", `~(lower "HELLO")`, "hello"},
		{"trim", `~(trim param.label)`, "Bug Fix"},
		{"trimPrefix", `~(trimPrefix "elastic/" param.repo)`, "elasticsearch"},
		{"trimSuffix", `~(trimSuffix "/elasticsearch" param.repo)`, "elastic"},
		{"replace", `~(replace "/" "-" param.repo)`, "elastic-elasticsearch"},
		{"truncate", `~(truncate 7 param.repo)`, "elastic"},
		{"truncate noop", `~(truncate 100 param.repo)`, "elastic/elasticsearch"},
		{"contains", `~(contains param.repo "elastic")`, "true"},
		{"hasPrefix", `~(hasPrefix param.repo "elastic")`, "true"},
		{"hasSuffix", `~(hasSuffix param.repo "search")`, "true"},
		{"nested chain", `~(upper (last (split "/" param.repo)))`, "ELASTICSEARCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := render(tt.tmpl, scope, nil)
			if err != nil {
				t.Fatalf("render(%q): %v", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

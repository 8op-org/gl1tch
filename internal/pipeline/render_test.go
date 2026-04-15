package pipeline

import "testing"

func TestRenderStringFunctions(t *testing.T) {
	data := map[string]any{
		"param": map[string]any{
			"repo":  "elastic/elasticsearch",
			"label": "  Bug Fix  ",
		},
	}

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{"split", `{{split "/" .param.repo}}`, `[elastic elasticsearch]`},
		{"first", `{{split "/" .param.repo | first}}`, `elastic`},
		{"last", `{{split "/" .param.repo | last}}`, `elasticsearch`},
		{"join", `{{split "/" .param.repo | join "-"}}`, `elastic-elasticsearch`},
		{"upper", `{{upper .param.repo}}`, `ELASTIC/ELASTICSEARCH`},
		{"lower", `{{lower "HELLO"}}`, `hello`},
		{"trim", `{{trim .param.label}}`, `Bug Fix`},
		{"trimPrefix", `{{trimPrefix "elastic/" .param.repo}}`, `elasticsearch`},
		{"trimSuffix", `{{trimSuffix "/elasticsearch" .param.repo}}`, `elastic`},
		{"replace", `{{replace "/" "-" .param.repo}}`, `elastic-elasticsearch`},
		{"truncate", `{{truncate 7 .param.repo}}`, `elastic`},
		{"truncate short string", `{{truncate 100 .param.repo}}`, `elastic/elasticsearch`},
		{"contains", `{{contains .param.repo "elastic"}}`, `true`},
		{"hasPrefix", `{{hasPrefix .param.repo "elastic"}}`, `true`},
		{"hasSuffix", `{{hasSuffix .param.repo "search"}}`, `true`},
		{"pipe chain", `{{split "/" .param.repo | last | upper}}`, `ELASTICSEARCH`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := render(tt.tmpl, data, nil)
			if err != nil {
				t.Fatalf("render(%q): %v", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

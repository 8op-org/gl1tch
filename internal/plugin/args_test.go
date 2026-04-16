// internal/plugin/args_test.go
package plugin

import (
	"testing"
)

// glitchSrc is a sample .glitch file with three arg forms.
var glitchSrc = []byte(`
(workflow "example" :description "test workflow")

(arg "query"
  :type :string
  :description "search query"
  :default "hello")

(arg "verbose"
  :type :flag
  :description "enable verbose output")

(arg "limit"
  :type :number
  :description "max results")
`)

func TestParseArgs(t *testing.T) {
	defs, err := ParseArgs(glitchSrc)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}
	if len(defs) != 3 {
		t.Fatalf("expected 3 ArgDefs, got %d", len(defs))
	}

	// --- query ---
	q := defs[0]
	if q.Name != "query" {
		t.Errorf("defs[0].Name = %q, want %q", q.Name, "query")
	}
	if q.Type != "string" {
		t.Errorf("defs[0].Type = %q, want %q", q.Type, "string")
	}
	if q.Default != "hello" {
		t.Errorf("defs[0].Default = %q, want %q", q.Default, "hello")
	}
	if q.Required {
		t.Errorf("defs[0].Required = true, want false (has default)")
	}
	if q.Description != "search query" {
		t.Errorf("defs[0].Description = %q, want %q", q.Description, "search query")
	}

	// --- verbose ---
	v := defs[1]
	if v.Name != "verbose" {
		t.Errorf("defs[1].Name = %q, want %q", v.Name, "verbose")
	}
	if v.Type != "flag" {
		t.Errorf("defs[1].Type = %q, want %q", v.Type, "flag")
	}
	if v.Required {
		t.Errorf("defs[1].Required = true, want false (flag type)")
	}

	// --- limit ---
	l := defs[2]
	if l.Name != "limit" {
		t.Errorf("defs[2].Name = %q, want %q", l.Name, "limit")
	}
	if l.Type != "number" {
		t.Errorf("defs[2].Type = %q, want %q", l.Type, "number")
	}
	if !l.Required {
		t.Errorf("defs[2].Required = false, want true (no default, not flag)")
	}
}

func TestParseArgs_NoArgs(t *testing.T) {
	src := []byte(`(workflow "no-args" :description "nothing here")`)
	defs, err := ParseArgs(src)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}
	if len(defs) != 0 {
		t.Errorf("expected 0 ArgDefs, got %d", len(defs))
	}
}

func TestBuildParams_FromFlags(t *testing.T) {
	defs, err := ParseArgs(glitchSrc)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	provided := map[string]string{
		"query":   "elasticsearch",
		"verbose": "true",
		"limit":   "50",
	}

	params, err := BuildParams(defs, provided)
	if err != nil {
		t.Fatalf("BuildParams error: %v", err)
	}

	for k, want := range provided {
		if got := params[k]; got != want {
			t.Errorf("params[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestBuildParams_Defaults(t *testing.T) {
	defs, err := ParseArgs(glitchSrc)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	// Only provide the required arg (limit has no default).
	provided := map[string]string{"limit": "10"}

	params, err := BuildParams(defs, provided)
	if err != nil {
		t.Fatalf("BuildParams error: %v", err)
	}

	if got := params["query"]; got != "hello" {
		t.Errorf("params[\"query\"] = %q, want %q", got, "hello")
	}
	if got := params["verbose"]; got != "" {
		t.Errorf("params[\"verbose\"] = %q, want empty string for flag", got)
	}
	if got := params["limit"]; got != "10" {
		t.Errorf("params[\"limit\"] = %q, want %q", got, "10")
	}
}

func TestBuildParams_RequiredMissing(t *testing.T) {
	defs, err := ParseArgs(glitchSrc)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	// Omit "limit" which is required (no default, type number).
	provided := map[string]string{"query": "test"}

	_, err = BuildParams(defs, provided)
	if err == nil {
		t.Fatal("expected error for missing required arg, got nil")
	}
}

func TestParseArgs_ExampleKeyword(t *testing.T) {
	src := []byte(`(arg "topic" :required true :description "The topic" :example "batch comparison")`)
	defs, err := ParseArgs(src)
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("want 1 def, got %d", len(defs))
	}
	if defs[0].Example != "batch comparison" {
		t.Errorf("Example = %q, want %q", defs[0].Example, "batch comparison")
	}
	if defs[0].Implicit {
		t.Errorf("Implicit should default to false")
	}
}

func TestParseArgs_UnknownKeywordRejected(t *testing.T) {
	src := []byte(`(arg "topic" :defalt "oops")`)
	_, err := ParseArgs(src)
	if err == nil {
		t.Fatal("expected parse error for unknown keyword :defalt")
	}
}

func TestParseArgs_RequiredWithDefaultRejected(t *testing.T) {
	src := []byte(`(arg "topic" :required true :default "x")`)
	_, err := ParseArgs(src)
	if err == nil {
		t.Fatal("expected parse error when :required and :default both set")
	}
}

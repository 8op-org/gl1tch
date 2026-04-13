// internal/sexpr/parser_test.go
package sexpr

import "testing"

func TestParse_EmptyList(t *testing.T) {
	nodes, err := Parse([]byte("()"))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if !nodes[0].IsList() {
		t.Fatal("expected list node")
	}
	if len(nodes[0].Children) != 0 {
		t.Fatalf("expected empty list, got %d children", len(nodes[0].Children))
	}
}

func TestParse_NestedList(t *testing.T) {
	nodes, err := Parse([]byte(`(workflow "test" (step "s1" (run "echo hi")))`))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 top-level node, got %d", len(nodes))
	}
	wf := nodes[0]
	if !wf.IsList() || len(wf.Children) != 3 {
		t.Fatalf("expected list with 3 children, got %d", len(wf.Children))
	}
	if wf.Children[0].StringVal() != "workflow" {
		t.Fatalf("expected 'workflow', got %q", wf.Children[0].StringVal())
	}
	step := wf.Children[2]
	if !step.IsList() || len(step.Children) != 3 {
		t.Fatalf("expected step list with 3 children, got %d", len(step.Children))
	}
}

func TestParse_Keywords(t *testing.T) {
	nodes, err := Parse([]byte(`(:name "test")`))
	if err != nil {
		t.Fatal(err)
	}
	list := nodes[0]
	if list.Children[0].KeywordVal() != "name" {
		t.Fatalf("expected keyword 'name', got %q", list.Children[0].KeywordVal())
	}
}

func TestParse_Discard(t *testing.T) {
	nodes, err := Parse([]byte(`("keep" #_"discard" "also-keep")`))
	if err != nil {
		t.Fatal(err)
	}
	list := nodes[0]
	if len(list.Children) != 2 {
		t.Fatalf("expected 2 children (discard removed one), got %d", len(list.Children))
	}
	if list.Children[0].StringVal() != "keep" {
		t.Fatalf("expected 'keep', got %q", list.Children[0].StringVal())
	}
	if list.Children[1].StringVal() != "also-keep" {
		t.Fatalf("expected 'also-keep', got %q", list.Children[1].StringVal())
	}
}

func TestParse_DiscardList(t *testing.T) {
	nodes, err := Parse([]byte(`("a" #_(skip "this" "entire" "thing") "b")`))
	if err != nil {
		t.Fatal(err)
	}
	list := nodes[0]
	if len(list.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(list.Children))
	}
}

func TestParse_UnmatchedRParen(t *testing.T) {
	_, err := Parse([]byte(")"))
	if err == nil {
		t.Fatal("expected error for unmatched )")
	}
}

func TestParse_UnmatchedLParen(t *testing.T) {
	_, err := Parse([]byte("("))
	if err == nil {
		t.Fatal("expected error for unmatched (")
	}
}

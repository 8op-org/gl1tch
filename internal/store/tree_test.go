package store

import (
	"path/filepath"
	"testing"
)

func TestParentRunID(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	parentID, err := s.RecordRun(RunRecord{Kind: "workflow", Name: "parent", WorkflowName: "parent-flow"})
	if err != nil {
		t.Fatal(err)
	}
	childID, err := s.RecordRun(RunRecord{Kind: "workflow", Name: "child", WorkflowName: "child-flow", ParentRunID: parentID})
	if err != nil {
		t.Fatal(err)
	}

	children, err := s.ListChildren(parentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 || children[0].ID != childID {
		t.Fatalf("expected one child with id %d, got %+v", childID, children)
	}
	if children[0].WorkflowName != "child-flow" {
		t.Errorf("expected workflow_name 'child-flow', got %q", children[0].WorkflowName)
	}
}

func TestGetRunTree(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, _ := s.RecordRun(RunRecord{Kind: "batch", Name: "p", WorkflowName: "parent"})
	c1, _ := s.RecordRun(RunRecord{Kind: "workflow", Name: "c1", WorkflowName: "child", ParentRunID: p})
	c2, _ := s.RecordRun(RunRecord{Kind: "workflow", Name: "c2", WorkflowName: "child", ParentRunID: p})
	_, _ = s.RecordRun(RunRecord{Kind: "workflow", Name: "gc", WorkflowName: "grandchild", ParentRunID: c1})

	tree, err := s.GetRunTree(p)
	if err != nil {
		t.Fatal(err)
	}
	if tree.ID != p {
		t.Fatalf("root id mismatch: %d vs %d", tree.ID, p)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(tree.Children))
	}
	var c1Node *RunNode
	for i := range tree.Children {
		if tree.Children[i].ID == c1 {
			c1Node = &tree.Children[i]
		}
	}
	if c1Node == nil || len(c1Node.Children) != 1 {
		t.Fatalf("c1 should have exactly one grandchild: %+v", c1Node)
	}
	_ = c2
}

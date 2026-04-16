package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/store"
)

func TestGetRunTreeEndpoint(t *testing.T) {
	s, err := store.OpenAt(filepath.Join(t.TempDir(), "tree.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, _ := s.RecordRun(store.RunRecord{Kind: "batch", Name: "p", WorkflowName: "parent"})
	c1, _ := s.RecordRun(store.RunRecord{Kind: "workflow", Name: "c1", WorkflowName: "child", ParentRunID: p})
	_, _ = s.RecordRun(store.RunRecord{Kind: "workflow", Name: "gc", WorkflowName: "grand", ParentRunID: c1})

	srv := &Server{store: s}
	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+itoa(p)+"/tree", nil)
	req.SetPathValue("id", itoa(p))
	rec := httptest.NewRecorder()
	srv.handleGetRunTree(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var tree store.RunNode
	if err := json.Unmarshal(rec.Body.Bytes(), &tree); err != nil {
		t.Fatalf("not JSON: %s", rec.Body.String())
	}
	if tree.ID != p {
		t.Fatalf("root id mismatch: %d vs %d", tree.ID, p)
	}
	if len(tree.Children) != 1 || tree.Children[0].ID != c1 {
		t.Fatalf("expected one child c1, got %+v", tree.Children)
	}
	if len(tree.Children[0].Children) != 1 {
		t.Fatalf("expected grandchild, got %+v", tree.Children[0].Children)
	}
}

func TestListRunsWithParentFilter(t *testing.T) {
	s, err := store.OpenAt(filepath.Join(t.TempDir(), "list.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, _ := s.RecordRun(store.RunRecord{Kind: "batch", Name: "p", WorkflowName: "parent"})
	c1, _ := s.RecordRun(store.RunRecord{Kind: "workflow", Name: "c1", WorkflowName: "child", ParentRunID: p})
	c2, _ := s.RecordRun(store.RunRecord{Kind: "workflow", Name: "c2", WorkflowName: "child", ParentRunID: p})

	srv := &Server{store: s}
	req := httptest.NewRequest(http.MethodGet, "/api/runs?parent_id="+itoa(p), nil)
	rec := httptest.NewRecorder()
	srv.handleListRuns(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var arr []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &arr)
	if len(arr) != 2 {
		t.Fatalf("expected 2 children, got %d: %s", len(arr), rec.Body.String())
	}
	ids := map[int64]bool{}
	for _, e := range arr {
		if v, ok := e["id"].(float64); ok {
			ids[int64(v)] = true
		} else if v, ok := e["ID"].(float64); ok {
			ids[int64(v)] = true
		}
	}
	if !ids[c1] || !ids[c2] {
		t.Fatalf("missing child rows: %+v", ids)
	}
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}

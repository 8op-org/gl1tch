package research

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolSetValidTool(t *testing.T) {
	ts := NewToolSet("/tmp", nil)

	if !ts.ValidTool("grep_code") {
		t.Error("grep_code should be valid")
	}
	if ts.ValidTool("hack_mainframe") {
		t.Error("hack_mainframe should be invalid")
	}
}

func TestToolSetDefinitions(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	defs := ts.Definitions()

	if len(defs) != 8 {
		t.Fatalf("expected 8 definitions, got %d", len(defs))
	}

	expected := []string{
		"grep_code", "read_file", "git_log", "git_diff",
		"search_es", "list_files", "fetch_issue", "fetch_pr",
	}
	for i, name := range expected {
		if defs[i].Name != name {
			t.Errorf("definition[%d]: expected %q, got %q", i, name, defs[i].Name)
		}
	}
}

func TestGrepCodeRequiresPattern(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	res := ts.Execute(context.Background(), "grep_code", map[string]string{})

	if res.Err == "" {
		t.Error("expected error for missing pattern")
	}
	if !strings.Contains(res.Err, "pattern") {
		t.Errorf("error should mention pattern, got: %s", res.Err)
	}
}

func TestReadFileWorks(t *testing.T) {
	dir := t.TempDir()
	content := "line one\nline two\nline three\n"
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ts := NewToolSet(dir, nil)
	res := ts.Execute(context.Background(), "read_file", map[string]string{"path": path})

	if res.Err != "" {
		t.Fatalf("unexpected error: %s", res.Err)
	}
	if !strings.Contains(res.Output, "line one") {
		t.Errorf("output should contain file content, got: %s", res.Output)
	}
}

func TestListFilesWorks(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.go", "b.txt"} {
		if err := os.WriteFile(filepath.Join(sub, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ts := NewToolSet(dir, nil)
	res := ts.Execute(context.Background(), "list_files", map[string]string{"path": dir, "depth": "3"})

	if res.Err != "" {
		t.Fatalf("unexpected error: %s", res.Err)
	}
	if !strings.Contains(res.Output, "a.go") {
		t.Errorf("output should list a.go, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "b.txt") {
		t.Errorf("output should list b.txt, got: %s", res.Output)
	}
}

func TestUnknownToolReturnsError(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	res := ts.Execute(context.Background(), "hack_mainframe", map[string]string{})

	if res.Err == "" {
		t.Error("expected error for unknown tool")
	}
	if !strings.Contains(res.Err, "unknown tool") {
		t.Errorf("error should mention unknown tool, got: %s", res.Err)
	}
}

func TestSearchESWithoutClient(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	res := ts.Execute(context.Background(), "search_es", map[string]string{"query": "test"})

	if res.Err != "" {
		t.Errorf("should not be an error, got: %s", res.Err)
	}
	if res.Output != "elasticsearch not available" {
		t.Errorf("expected 'elasticsearch not available', got: %s", res.Output)
	}
}

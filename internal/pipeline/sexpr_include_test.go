package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSexpr_Include_Basic(t *testing.T) {
	dir := t.TempDir()
	shared := `(def model "qwen2.5:7b")
(def greeting "hello")`
	if err := os.WriteFile(filepath.Join(dir, "shared.glitch"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}

	main := `(include "` + filepath.Join(dir, "shared.glitch") + `")

(workflow "test-include"
  :description "tests include"
  (step "s1"
    (run "echo hello")))`

	w, err := parseSexprWorkflow([]byte(main))
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "test-include" {
		t.Fatalf("expected name %q, got %q", "test-include", w.Name)
	}
}

func TestSexpr_Include_DefPropagates(t *testing.T) {
	dir := t.TempDir()
	shared := `(def model "qwen2.5:7b")`
	if err := os.WriteFile(filepath.Join(dir, "shared.glitch"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}

	main := `(include "` + filepath.Join(dir, "shared.glitch") + `")

(workflow "test-def-propagation"
  :description "included def resolves in workflow"
  (step "s1"
    (llm :model model :prompt "hello")))`

	w, err := parseSexprWorkflow([]byte(main))
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM == nil {
		t.Fatal("expected LLM step")
	}
	if w.Steps[0].LLM.Model != "qwen2.5:7b" {
		t.Fatalf("expected model %q, got %q", "qwen2.5:7b", w.Steps[0].LLM.Model)
	}
}

func TestSexpr_Include_CircularDetected(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "a.glitch")
	fileB := filepath.Join(dir, "b.glitch")

	os.WriteFile(fileA, []byte(`(include "`+fileB+`")
(def x "1")`), 0o644)
	os.WriteFile(fileB, []byte(`(include "`+fileA+`")
(def y "2")`), 0o644)

	main := `(include "` + fileA + `")
(workflow "circular" :description "test" (step "s" (run "echo")))`

	_, err := parseSexprWorkflow([]byte(main))
	if err == nil {
		t.Fatal("expected circular include error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Fatalf("expected circular error, got: %v", err)
	}
}

func TestSexpr_Include_FileNotFound(t *testing.T) {
	main := `(include "/nonexistent/file.glitch")
(workflow "test" :description "test" (step "s" (run "echo")))`

	_, err := parseSexprWorkflow([]byte(main))
	if err == nil {
		t.Fatal("expected file not found error")
	}
}

func TestSexpr_Include_OnlyImportsDefs(t *testing.T) {
	dir := t.TempDir()
	shared := `(def model "qwen2.5:7b")
(workflow "should-be-ignored" :description "x" (step "s" (run "echo")))`
	if err := os.WriteFile(filepath.Join(dir, "shared.glitch"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}

	main := `(include "` + filepath.Join(dir, "shared.glitch") + `")

(workflow "real-workflow"
  :description "only this workflow should exist"
  (step "s1"
    (llm :model model :prompt "hello")))`

	w, err := parseSexprWorkflow([]byte(main))
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "real-workflow" {
		t.Fatalf("expected name %q, got %q", "real-workflow", w.Name)
	}
}

func TestSexpr_DefReadFile(t *testing.T) {
	dir := t.TempDir()
	content := "these are my conventions"
	if err := os.WriteFile(filepath.Join(dir, "conventions.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	src := `(def conventions (read-file "` + filepath.Join(dir, "conventions.md") + `"))

(workflow "test-def-readfile"
  :description "def evaluates read-file"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefReadFileMultiple(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("file-a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("file-b"), 0o644)

	src := `(def combined (read-file "` + filepath.Join(dir, "a.md") + `" "` + filepath.Join(dir, "b.md") + `"))

(workflow "test-multi-readfile"
  :description "def reads multiple files"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefReadFileNotFound(t *testing.T) {
	src := `(def x (read-file "/nonexistent/path.md"))
(workflow "test" :description "test" (step "s" (run "echo")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err == nil {
		t.Fatal("expected error for missing file in def")
	}
}

func TestSexpr_DefGlob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.glitch"), []byte("workflow-a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.glitch"), []byte("workflow-b"), 0o644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("not-a-workflow"), 0o644)

	src := `(def files (glob "` + filepath.Join(dir, "*.glitch") + `"))

(workflow "test-glob"
  :description "def evaluates glob"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefThread_GlobMapReadFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.glitch"), []byte("content-a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.glitch"), []byte("content-b"), 0o644)

	src := `(def examples (-> (glob "` + filepath.Join(dir, "*.glitch") + `") (map read-file) (join "\n\n")))

(workflow "test-thread"
  :description "threading in def"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefThread_Lines(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "data.txt"), []byte("line1\nline2\nline3"), 0o644)

	src := `(def line-count (-> (read-file "` + filepath.Join(dir, "data.txt") + `") (lines) (join ", ")))

(workflow "test-thread-lines"
  :description "lines + join in thread"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefThread_Filter(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cmds.txt"), []byte("glitch run\nother thing\nglitch workflow list\nskip me"), 0o644)

	src := `(def cmds (-> (read-file "` + filepath.Join(dir, "cmds.txt") + `") (lines) (filter (contains "glitch")) (join "\n")))

(workflow "test-filter"
  :description "filter in thread"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

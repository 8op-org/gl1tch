package workspace

import (
	"strings"
	"testing"
)

func TestUpdatePinPreservesComments(t *testing.T) {
	src := `(workspace "demo"
  ;; keep this comment
  (resource "ensemble" :type "git"
    :url "https://github.com/elastic/ensemble"
    :ref "main"
    :pin "old-sha") ; inline
  (resource "notes" :type "local" :path "~/my-notes"))
`
	out, err := UpdatePin([]byte(src), "ensemble", "new-sha")
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, ";; keep this comment") {
		t.Error("top comment lost")
	}
	if !strings.Contains(s, "; inline") {
		t.Error("inline comment lost")
	}
	if !strings.Contains(s, `:pin "new-sha"`) {
		t.Errorf("pin not updated:\n%s", s)
	}
	if strings.Contains(s, `"old-sha"`) {
		t.Error("old pin still present")
	}
}

func TestUpdatePinAddsMissingKey(t *testing.T) {
	src := `(workspace "demo"
  (resource "x" :type "git" :url "u" :ref "main"))
`
	out, err := UpdatePin([]byte(src), "x", "newpin")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `:pin "newpin"`) {
		t.Fatalf("pin not added:\n%s", out)
	}
}

func TestUpdatePinUnknownResource(t *testing.T) {
	src := `(workspace "demo")`
	if _, err := UpdatePin([]byte(src), "ghost", "x"); err == nil {
		t.Fatal("expected error for unknown resource")
	}
}

func TestUpdateRefPreservesComments(t *testing.T) {
	src := `(workspace "demo"
  ;; keep this comment
  (resource "ensemble" :type "git"
    :url "https://github.com/elastic/ensemble"
    :ref "main"
    :pin "old-sha") ; inline
  (resource "notes" :type "local" :path "~/my-notes"))
`
	out, err := UpdateRef([]byte(src), "ensemble", "feature-branch")
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, ";; keep this comment") {
		t.Error("top comment lost")
	}
	if !strings.Contains(s, "; inline") {
		t.Error("inline comment lost")
	}
	if !strings.Contains(s, `:ref "feature-branch"`) {
		t.Errorf("ref not updated:\n%s", s)
	}
	if strings.Contains(s, `:ref "main"`) {
		t.Error("old ref still present")
	}
	// Pin should remain untouched.
	if !strings.Contains(s, `:pin "old-sha"`) {
		t.Errorf("pin was changed unexpectedly:\n%s", s)
	}
}

func TestUpdateRefAddsMissingKey(t *testing.T) {
	src := `(workspace "demo"
  (resource "x" :type "git" :url "u"))
`
	out, err := UpdateRef([]byte(src), "x", "main")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `:ref "main"`) {
		t.Fatalf("ref not added:\n%s", out)
	}
}

func TestUpdateRefUnknownResource(t *testing.T) {
	src := `(workspace "demo")`
	if _, err := UpdateRef([]byte(src), "ghost", "x"); err == nil {
		t.Fatal("expected error for unknown resource")
	}
}

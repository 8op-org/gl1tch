package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCheckCmd_ValidFile(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "good.glitch")
	os.WriteFile(wf, []byte(`(workflow "test" (step "a" (run "echo ok")))`), 0644)

	out := &bytes.Buffer{}
	checkCmd.SetOut(out)

	if err := runCheck(checkCmd, []string{wf}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("expected success message, got: %s", out.String())
	}
}

func TestCheckCmd_WithStdlib(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "uses-stdlib.glitch")
	os.WriteFile(wf, []byte(`(include "std/strings")
(workflow "test"
  :description "uses stdlib"
  (step "a" (kebab-case "Hello World")))
`), 0644)

	out := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "check", Args: cobra.ExactArgs(1), RunE: runCheck}
	cmd.SetOut(out)
	cmd.SetArgs([]string{wf})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("check should pass for valid stdlib usage: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("expected ok, got: %s", out.String())
	}
}

func TestCheckCmd_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "bad.glitch")
	os.WriteFile(wf, []byte(`(workflow "test" (frobnicate "x"))`), 0644)

	out := &bytes.Buffer{}
	checkCmd.SetOut(out)
	checkCmd.SetErr(out)

	err := runCheck(checkCmd, []string{wf})
	if err == nil {
		t.Fatal("expected error for unknown symbol 'frobnicate'")
	}
	if !strings.Contains(out.String(), "frobnicate") {
		t.Fatalf("expected diagnostic mentioning frobnicate, got: %s", out.String())
	}
}

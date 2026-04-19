package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

func TestEvalCmd_OneShot(t *testing.T) {
	out := &bytes.Buffer{}
	evalCmd.SetOut(out)

	// Set the flag directly and reset after test.
	prev := evalExpr
	evalExpr = `(str "hello" " " "world")`
	defer func() { evalExpr = prev }()

	if err := runEval(evalCmd, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hello world") {
		t.Fatalf("want %q in output, got %q", "hello world", out.String())
	}
}

func TestEvalCmd_WithStdlib(t *testing.T) {
	out := &bytes.Buffer{}

	ev := pipeline.NewEvaluator()
	env := pipeline.NewEnv(nil)
	ev.RegisterBuiltins(env)

	val, err := ev.RunSourceWithEnv(env, []byte(`(do (include "std/strings") (kebab-case "Hello World"))`))
	if err != nil {
		t.Fatal(err)
	}
	if val.String() != "hello-world" {
		t.Fatalf("want %q, got %q", "hello-world", val.String())
	}
	_ = out
}

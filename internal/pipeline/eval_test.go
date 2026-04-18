package pipeline

import (
	"strings"
	"testing"
)

func evalHelper(t *testing.T, src string) string {
	t.Helper()
	ev := NewEvaluator()
	ev.Input = "test-input"
	ev.Params = map[string]string{"repo": "gl1tch", "name": "test"}
	val, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return val.String()
}

func TestEval_DefAndSymbol(t *testing.T) {
	got := evalHelper(t, `(def x "hello") x`)
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestEval_Fn(t *testing.T) {
	src := `(def greet (fn (name) (str "hello " name))) (greet "world")`
	got := evalHelper(t, src)
	if got != "hello world" {
		t.Fatalf("want %q, got %q", "hello world", got)
	}
}

func TestEval_Let(t *testing.T) {
	got := evalHelper(t, `(let (x "a" y "b") (str x y))`)
	if got != "ab" {
		t.Fatalf("want %q, got %q", "ab", got)
	}
}

func TestEval_Closure(t *testing.T) {
	src := `(def make-adder (fn (prefix) (fn (s) (str prefix s)))) (def add-hello (make-adder "hello-")) (add-hello "world")`
	got := evalHelper(t, src)
	if got != "hello-world" {
		t.Fatalf("want %q, got %q", "hello-world", got)
	}
}

func TestEval_If(t *testing.T) {
	got := evalHelper(t, `(if true "yes" "no")`)
	if got != "yes" {
		t.Fatalf("true branch: want %q, got %q", "yes", got)
	}
	got = evalHelper(t, `(if false "yes" "no")`)
	if got != "no" {
		t.Fatalf("false branch: want %q, got %q", "no", got)
	}
}

func TestEval_Cond(t *testing.T) {
	got := evalHelper(t, `(cond false "a" true "b")`)
	if got != "b" {
		t.Fatalf("want %q, got %q", "b", got)
	}
}

func TestEval_WorkflowAndStep(t *testing.T) {
	src := `(workflow test-wf
  (step a (str "alpha"))
  (step b (str "beta")))`
	ev := NewEvaluator()
	ev.Input = "test-input"
	ev.Params = map[string]string{"repo": "gl1tch", "name": "test"}
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["a"] != "alpha" {
		t.Fatalf("step a: want %q, got %q", "alpha", steps["a"])
	}
	if steps["b"] != "beta" {
		t.Fatalf("step b: want %q, got %q", "beta", steps["b"])
	}
}

func TestEval_Par(t *testing.T) {
	src := `(workflow par-test
  (par
    (step x (str "X"))
    (step y (str "Y")))
  (step merge (str "got:" (step x) (step y))))`
	ev := NewEvaluator()
	ev.Input = "test-input"
	ev.Params = map[string]string{"repo": "gl1tch", "name": "test"}
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["merge"] != "got:XY" {
		t.Fatalf("merge step: want %q, got %q", "got:XY", steps["merge"])
	}
}

func TestEval_StepInterpolation(t *testing.T) {
	src := `(workflow interp-test
  (step a (str "hello"))
  (step b "result: ~(step a)"))`
	ev := NewEvaluator()
	ev.Input = "test-input"
	ev.Params = map[string]string{"repo": "gl1tch", "name": "test"}
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["b"] != "result: hello" {
		t.Fatalf("step b: want %q, got %q", "result: hello", steps["b"])
	}
}

func TestEval_ParamInterpolation(t *testing.T) {
	src := `(workflow param-test
  (step a "repo is ~param.repo"))`
	ev := NewEvaluator()
	ev.Input = "test-input"
	ev.Params = map[string]string{"repo": "gl1tch", "name": "test"}
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["a"] != "repo is gl1tch" {
		t.Fatalf("step a: want %q, got %q", "repo is gl1tch", steps["a"])
	}
}

func TestEval_Retry(t *testing.T) {
	got := evalHelper(t, `(retry "3" (str "ok"))`)
	if got != "ok" {
		t.Fatalf("want %q, got %q", "ok", got)
	}
}

func TestEval_Catch(t *testing.T) {
	// Use an undefined symbol reference which will error
	got := evalHelper(t, `(catch nonexistent-symbol (str "caught"))`)
	if got != "caught" {
		t.Fatalf("want %q, got %q", "caught", got)
	}
}

func TestEval_Or(t *testing.T) {
	got := evalHelper(t, `(or "" "" "found")`)
	if got != "found" {
		t.Fatalf("want %q, got %q", "found", got)
	}
}

func TestEval_Sh(t *testing.T) {
	got := evalHelper(t, `(sh "echo hello")`)
	if !strings.Contains(got, "hello") {
		t.Fatalf("want output containing %q, got %q", "hello", got)
	}
}

func TestEval_Map(t *testing.T) {
	src := `(workflow map-test
  (step source (str "a\nb\nc"))
  (map (step source) (str "item:" item)))`
	got := evalHelper(t, src)
	if !strings.Contains(got, "item:a") {
		t.Fatalf("want output containing %q, got %q", "item:a", got)
	}
	if !strings.Contains(got, "item:c") {
		t.Fatalf("want output containing %q, got %q", "item:c", got)
	}
}

func TestEval_Str(t *testing.T) {
	got := evalHelper(t, `(str "a" "b" "c")`)
	if got != "abc" {
		t.Fatalf("want %q, got %q", "abc", got)
	}
}

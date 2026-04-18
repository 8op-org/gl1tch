package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/sexpr"
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

func TestEval_Assoc(t *testing.T) {
	got := evalHelper(t, `(def m (assoc :name "gl1tch" :version "0.1")) (pick m :name)`)
	if got != "gl1tch" {
		t.Fatalf("want %q, got %q", "gl1tch", got)
	}
}

func TestEval_AssocNested(t *testing.T) {
	src := `(def m (assoc :meta (assoc :title "hello"))) (pick (pick m :meta) :title)`
	got := evalHelper(t, src)
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestEval_PickFromList(t *testing.T) {
	src := `(def items (list "a" "b" "c")) (pick items 1)`
	got := evalHelper(t, src)
	if got != "b" {
		t.Fatalf("want %q, got %q", "b", got)
	}
}

func TestEval_AssocInList(t *testing.T) {
	src := `(def pages (list (assoc :slug "intro") (assoc :slug "guide")))
	         (pick (pick pages 0) :slug)`
	got := evalHelper(t, src)
	if got != "intro" {
		t.Fatalf("want %q, got %q", "intro", got)
	}
}

func TestEval_Replace(t *testing.T) {
	got := evalHelper(t, `(replace "hello world" "world" "earth")`)
	if got != "hello earth" {
		t.Fatalf("want %q, got %q", "hello earth", got)
	}
}

func TestEval_ReplaceMultiple(t *testing.T) {
	got := evalHelper(t, `(replace "a/b/c.md" "a/b/" "" ".md" "")`)
	if got != "c" {
		t.Fatalf("want %q, got %q", "c", got)
	}
}

func TestEval_Contains(t *testing.T) {
	got := evalHelper(t, `(contains "hello world" "world")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_ContainsList(t *testing.T) {
	got := evalHelper(t, `(contains (list "a" "b" "c") "b")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_Join(t *testing.T) {
	got := evalHelper(t, `(join (list "a" "b" "c") ",")`)
	if got != "a,b,c" {
		t.Fatalf("want %q, got %q", "a,b,c", got)
	}
}

func TestEval_Split(t *testing.T) {
	src := `(def parts (split "a,b,c" ",")) (pick parts 1)`
	got := evalHelper(t, src)
	if got != "b" {
		t.Fatalf("want %q, got %q", "b", got)
	}
}

func TestEval_Trim(t *testing.T) {
	got := evalHelper(t, `(trim "  hello  ")`)
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestEval_Upper(t *testing.T) {
	got := evalHelper(t, `(upper "hello")`)
	if got != "HELLO" {
		t.Fatalf("want %q, got %q", "HELLO", got)
	}
}

func TestEval_Lower(t *testing.T) {
	got := evalHelper(t, `(lower "HELLO")`)
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestEval_Lines(t *testing.T) {
	src := `(def ls (lines "a\nb\nc")) (pick ls 1)`
	got := evalHelper(t, src)
	if got != "b" {
		t.Fatalf("want %q, got %q", "b", got)
	}
}

func TestEval_StartsWith(t *testing.T) {
	got := evalHelper(t, `(starts-with "glitch run" "glitch")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_EndsWith(t *testing.T) {
	got := evalHelper(t, `(ends-with "hello.md" ".md")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_Slice(t *testing.T) {
	got := evalHelper(t, `(slice "hello world" 6 11)`)
	if got != "world" {
		t.Fatalf("want %q, got %q", "world", got)
	}
}

func TestEval_RegexMatch(t *testing.T) {
	src := `(def matches (regex-match "\\d+" "a1b22c333")) (join matches ",")`
	got := evalHelper(t, src)
	if got != "1,22,333" {
		t.Fatalf("want %q, got %q", "1,22,333", got)
	}
}

func TestEval_RegexMatchCapture(t *testing.T) {
	src := `(def matches (regex-match "\\(([a-z]+)" "(foo) (bar)")) (join matches ",")`
	got := evalHelper(t, src)
	if got != "foo,bar" {
		t.Fatalf("want %q, got %q", "foo,bar", got)
	}
}

func TestEval_RegexFind(t *testing.T) {
	got := evalHelper(t, `(regex-find "\\d+" "abc123def")`)
	if got != "123" {
		t.Fatalf("want %q, got %q", "123", got)
	}
}

func TestEval_RegexFindNil(t *testing.T) {
	got := evalHelper(t, `(regex-find "\\d+" "abcdef")`)
	if got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestEval_Count(t *testing.T) {
	got := evalHelper(t, `(count (list "a" "b" "c"))`)
	if got != "3" {
		t.Fatalf("want %q, got %q", "3", got)
	}
}

func TestEval_CountString(t *testing.T) {
	got := evalHelper(t, `(count "hello")`)
	if got != "5" {
		t.Fatalf("want %q, got %q", "5", got)
	}
}

func TestEval_Some(t *testing.T) {
	got := evalHelper(t, `(some (list "a" "bb" "ccc") (fn (x) (starts-with x "b")))`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_SomeFalse(t *testing.T) {
	got := evalHelper(t, `(some (list "a" "b" "c") (fn (x) (starts-with x "z")))`)
	if got != "false" {
		t.Fatalf("want %q, got %q", "false", got)
	}
}

func TestEval_Every(t *testing.T) {
	got := evalHelper(t, `(every (list "ab" "ac" "ad") (fn (x) (starts-with x "a")))`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_Set(t *testing.T) {
	src := `(join (set (list "a" "b" "a" "c" "b")) ",")`
	got := evalHelper(t, src)
	if got != "a,b,c" {
		t.Fatalf("want %q, got %q", "a,b,c", got)
	}
}

func TestEval_Difference(t *testing.T) {
	src := `(join (difference (list "a" "b" "c") (list "b")) ",")`
	got := evalHelper(t, src)
	if got != "a,c" {
		t.Fatalf("want %q, got %q", "a,c", got)
	}
}

func TestEval_Flatten(t *testing.T) {
	src := `(join (flatten (list (list "a" "b") (list "c"))) ",")`
	got := evalHelper(t, src)
	if got != "a,b,c" {
		t.Fatalf("want %q, got %q", "a,b,c", got)
	}
}

func TestEval_Assert(t *testing.T) {
	got := evalHelper(t, `(assert true "should not fail")`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
}

func TestEval_AssertFail(t *testing.T) {
	ev := NewEvaluator()
	ev.Input = "test"
	ev.Params = map[string]string{}
	_, err := ev.RunSource([]byte(`(assert false "expected failure")`))
	if err == nil {
		t.Fatal("expected error from assert false")
	}
	if !strings.Contains(err.Error(), "expected failure") {
		t.Fatalf("want error containing %q, got %q", "expected failure", err.Error())
	}
}

func TestEval_LessThan(t *testing.T) {
	got := evalHelper(t, `(< 3 5)`)
	if got != "true" {
		t.Fatalf("want %q, got %q", "true", got)
	}
	got = evalHelper(t, `(< 5 3)`)
	if got != "false" {
		t.Fatalf("want %q, got %q", "false", got)
	}
}

func TestEval_WriteFilePositional(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/test.txt"
	src := fmt.Sprintf(`(write-file "%s" :content "hello")`, path)
	got := evalHelper(t, src)
	if got != path {
		t.Fatalf("want %q, got %q", path, got)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello" {
		t.Fatalf("file content: want %q, got %q", "hello", string(data))
	}
}

func TestEval_ManifestLoad(t *testing.T) {
	data, err := os.ReadFile("../../site-manifest.glitch")
	if err != nil {
		t.Skip("site-manifest.glitch not found")
	}
	ev := NewEvaluator()
	ev.Params = map[string]string{}
	result, err := ev.RunSource(data)
	if err != nil {
		t.Fatalf("manifest eval: %v", err)
	}
	m, ok := result.(*MapVal)
	if !ok {
		t.Fatalf("manifest should return MapVal, got %T", result)
	}
	site, found := m.Get("site")
	if !found || site.String() != "gl1tch" {
		t.Fatalf("want site=gl1tch, got %v", site)
	}
	sections, found := m.Get("sections")
	if !found {
		t.Fatal("missing :sections key")
	}
	if _, ok := sections.(ListVal); !ok {
		t.Fatalf("sections should be ListVal, got %T", sections)
	}
}

func TestEval_SavePositional(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/out.txt"
	src := fmt.Sprintf(`(workflow save-test
	  (step data "content here")
	  (save "%s" :from "data"))`, path)
	evalHelper(t, src)
	data, _ := os.ReadFile(path)
	if string(data) != "content here" {
		t.Fatalf("file content: want %q, got %q", "content here", string(data))
	}
}

func TestEval_GateWorkflowsParse(t *testing.T) {
	files, _ := filepath.Glob("../../workflows/gate-*.glitch")
	if len(files) == 0 {
		t.Skip("no gate workflows found")
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		_, err = sexpr.Parse(data)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
	}
}

func TestEval_SiteWorkflowsParse(t *testing.T) {
	patterns := []string{"../../workflows/site*.glitch", "../../site/shared.glitch"}
	for _, pat := range patterns {
		files, _ := filepath.Glob(pat)
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			_, err = sexpr.Parse(data)
			if err != nil {
				t.Fatalf("parse %s: %v", f, err)
			}
		}
	}
}

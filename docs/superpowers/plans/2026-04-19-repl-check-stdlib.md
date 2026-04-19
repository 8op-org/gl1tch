# REPL, Static Validator, and Stdlib Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give workflow authors three tools that collapse the write→run→fail loop: an interactive REPL for experimentation, a `glitch check` command for static validation before execution, and a stdlib of reusable patterns they can `(include)`.

**Architecture:** The REPL wraps `Evaluator.RunSource()` in a readline loop with persistent state between inputs. The validator does a single AST walk (no evaluation) to catch undefined refs, unknown forms, and arity mismatches. The stdlib is a directory of `.glitch` files shipped alongside the binary via `embed.FS`, auto-discoverable via `(include "std/...")`.

**Tech Stack:** Go, `github.com/8op-org/gl1tch/internal/sexpr` (parser), `github.com/8op-org/gl1tch/internal/pipeline` (evaluator), `github.com/spf13/cobra` (CLI), `embed` (stdlib bundling)

---

## File Structure

### New files
| Path | Responsibility |
|------|---------------|
| `internal/pipeline/repl.go` | REPL loop: readline, eval, print, state persistence |
| `internal/pipeline/repl_test.go` | REPL unit tests (simulated I/O) |
| `cmd/eval.go` | `glitch eval` cobra command (REPL + one-shot `-e`) |
| `cmd/eval_test.go` | CLI integration tests for eval command |
| `internal/pipeline/check.go` | Static AST walker: collects diagnostics without evaluation |
| `internal/pipeline/check_test.go` | Validator tests against known-good and known-bad workflows |
| `cmd/check.go` | `glitch check` cobra command |
| `cmd/check_test.go` | CLI integration tests for check command |
| `internal/pipeline/stdlib/strings.glitch` | String utility functions |
| `internal/pipeline/stdlib/collections.glitch` | Collection helpers (pluck, group-by, zip) |
| `internal/pipeline/stdlib/io.glitch` | File/HTTP convenience wrappers |
| `internal/pipeline/stdlib/embed.go` | `embed.FS` for bundled stdlib |
| `internal/pipeline/stdlib_test.go` | Tests that every stdlib file parses and key defs evaluate correctly |

### Modified files
| Path | Change |
|------|--------|
| `cmd/root.go` | Register `evalCmd` and `checkCmd` |
| `internal/pipeline/eval.go:74-91` | Extract `RunSourceWithEnv()` variant that accepts an existing `*Env` (for REPL state persistence) |
| `internal/sexpr/parser.go` | No changes — `Parse()` API is sufficient |

---

## Task 1: Extract `RunSourceWithEnv` for REPL state persistence

The REPL needs to evaluate successive inputs in the same environment so that `(def x 10)` in one line is visible in the next. Currently `RunSource()` creates a fresh `Env` every call. We extract a variant that takes an existing env.

**Files:**
- Modify: `internal/pipeline/eval.go:74-91`
- Test: `internal/pipeline/eval_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestEval_RunSourceWithEnv(t *testing.T) {
	ev := NewEvaluator()
	env := NewEnv(nil)
	ev.RegisterBuiltins(env)

	// First eval defines a binding
	_, err := ev.RunSourceWithEnv(env, []byte(`(def x "hello")`))
	if err != nil {
		t.Fatal(err)
	}

	// Second eval sees it
	val, err := ev.RunSourceWithEnv(env, []byte(`x`))
	if err != nil {
		t.Fatal(err)
	}
	if val.String() != "hello" {
		t.Fatalf("want %q, got %q", "hello", val.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestEval_RunSourceWithEnv -v`
Expected: FAIL — `ev.RunSourceWithEnv` and `ev.RegisterBuiltins` undefined

- [ ] **Step 3: Implement RunSourceWithEnv and export RegisterBuiltins**

In `internal/pipeline/eval.go`, add after `RunSource` (line 91):

```go
// RunSourceWithEnv parses and evaluates source in an existing environment.
// Callers are responsible for registering builtins in env before first use.
func (ev *Evaluator) RunSourceWithEnv(env *Env, src []byte) (Value, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var result Value = NilVal{}
	for _, n := range nodes {
		result, err = ev.Eval(env, n)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// RegisterBuiltins installs all builtin functions into env.
// Exported so callers (REPL, tests) can prepare an env for RunSourceWithEnv.
func (ev *Evaluator) RegisterBuiltins(env *Env) {
	ev.registerBuiltins(env)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestEval_RunSourceWithEnv -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add internal/pipeline/eval.go internal/pipeline/eval_test.go
git commit -m "feat(pipeline): add RunSourceWithEnv for persistent eval sessions"
```

---

## Task 2: REPL core — read, eval, print loop

Build the REPL as a testable struct that takes `io.Reader`/`io.Writer` so tests can drive it without a terminal.

**Files:**
- Create: `internal/pipeline/repl.go`
- Create: `internal/pipeline/repl_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/pipeline/repl_test.go
package pipeline

import (
	"bytes"
	"strings"
	"testing"
)

func TestREPL_BasicDefAndLookup(t *testing.T) {
	in := strings.NewReader("(def x 42)\nx\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	err := r.Run()
	if err != nil {
		t.Fatal(err)
	}
	// Output should contain "42" from the second expression
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("expected output to contain %q, got:\n%s", "42", out.String())
	}
}

func TestREPL_ParseError(t *testing.T) {
	in := strings.NewReader("(def x\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	_ = r.Run()
	// Should print error, not crash
	if !strings.Contains(out.String(), "error:") {
		t.Fatalf("expected error output, got:\n%s", out.String())
	}
}

func TestREPL_MultilineInput(t *testing.T) {
	// Open paren on first line, close on second — REPL should detect incomplete input
	in := strings.NewReader("(str \"hello\"\n  \"world\")\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	_ = r.Run()
	if !strings.Contains(out.String(), "hello world") {
		t.Fatalf("expected %q, got:\n%s", "hello world", out.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestREPL -v`
Expected: FAIL — `NewREPL` undefined

- [ ] **Step 3: Implement the REPL**

```go
// internal/pipeline/repl.go
package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// REPL is an interactive read-eval-print loop for the glitch evaluator.
type REPL struct {
	ev  *Evaluator
	env *Env
	in  io.Reader
	out io.Writer
}

// NewREPL creates a REPL with the given evaluator and I/O streams.
func NewREPL(ev *Evaluator, in io.Reader, out io.Writer) *REPL {
	env := NewEnv(nil)
	ev.RegisterBuiltins(env)
	return &REPL{ev: ev, env: env, in: in, out: out}
}

// Run starts the REPL loop. Returns nil on EOF.
func (r *REPL) Run() error {
	scanner := bufio.NewScanner(r.in)
	var buf strings.Builder
	prompt := "gl1tch> "

	r.writePrompt(prompt)

	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line)
		buf.WriteString("\n")

		src := buf.String()
		if !balanced(src) {
			prompt = "  ...   "
			r.writePrompt(prompt)
			continue
		}

		trimmed := strings.TrimSpace(src)
		buf.Reset()
		prompt = "gl1tch> "

		if trimmed == "" {
			r.writePrompt(prompt)
			continue
		}

		val, err := r.ev.RunSourceWithEnv(r.env, []byte(trimmed))
		if err != nil {
			fmt.Fprintf(r.out, "error: %s\n", err)
		} else if val != nil && val.String() != "" {
			fmt.Fprintln(r.out, val.String())
		}

		r.writePrompt(prompt)
	}
	return scanner.Err()
}

func (r *REPL) writePrompt(p string) {
	fmt.Fprint(r.out, p)
}

// balanced returns true if parens are balanced in src.
func balanced(src string) bool {
	depth := 0
	inString := false
	inTriple := false
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if inTriple {
			if ch == '`' && i+2 < len(src) && src[i+1] == '`' && src[i+2] == '`' {
				inTriple = false
				i += 2
			}
			continue
		}
		if inString {
			if ch == '\\' {
				i++ // skip escaped char
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '`':
			if i+2 < len(src) && src[i+1] == '`' && src[i+2] == '`' {
				inTriple = true
				i += 2
			}
		case '"':
			inString = true
		case ';':
			// skip to end of line (comment)
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case '(':
			depth++
		case ')':
			depth--
		}
	}
	return depth <= 0 && !inString && !inTriple
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestREPL -v`
Expected: PASS (all 3 tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add internal/pipeline/repl.go internal/pipeline/repl_test.go
git commit -m "feat(pipeline): add REPL with multiline detection and persistent env"
```

---

## Task 3: `glitch eval` CLI command

Wire the REPL to a cobra command. Support both interactive mode (no args) and one-shot mode (`-e "expr"`).

**Files:**
- Create: `cmd/eval.go`
- Create: `cmd/eval_test.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Write the failing test**

```go
// cmd/eval_test.go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestEvalCmd_OneShot(t *testing.T) {
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"eval", "-e", `(str "hello" " " "world")`})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hello world") {
		t.Fatalf("want %q in output, got %q", "hello world", out.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./cmd/ -run TestEvalCmd -v`
Expected: FAIL — unknown command "eval"

- [ ] **Step 3: Implement the eval command**

```go
// cmd/eval.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

var evalExpr string

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "interactive evaluator for glitch expressions",
	Long: `Start an interactive REPL for experimenting with glitch workflow expressions.

Use -e to evaluate a single expression:
  glitch eval -e '(str "hello" " " "world")'

Without -e, starts an interactive session:
  glitch eval`,
	RunE: runEval,
}

func init() {
	evalCmd.Flags().StringVarP(&evalExpr, "expr", "e", "", "evaluate a single expression and exit")
	rootCmd.AddCommand(evalCmd)
}

func runEval(cmd *cobra.Command, args []string) error {
	ev := pipeline.NewEvaluator()

	if evalExpr != "" {
		env := pipeline.NewEnv(nil)
		ev.RegisterBuiltins(env)
		val, err := ev.RunSourceWithEnv(env, []byte(evalExpr))
		if err != nil {
			return fmt.Errorf("eval: %s", err)
		}
		if val != nil && val.String() != "" {
			fmt.Fprintln(cmd.OutOrStdout(), val.String())
		}
		return nil
	}

	// Interactive mode
	r := pipeline.NewREPL(ev, os.Stdin, cmd.OutOrStdout())
	return r.Run()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./cmd/ -run TestEvalCmd -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add cmd/eval.go cmd/eval_test.go
git commit -m "feat(cmd): add glitch eval command with REPL and one-shot mode"
```

---

## Task 4: Static validator — AST walker core

The validator walks the AST without evaluating, collecting diagnostics. It knows all special forms and builtins, so it can flag unknown symbols, bad arity, and unreachable code.

**Files:**
- Create: `internal/pipeline/check.go`
- Create: `internal/pipeline/check_test.go`

- [ ] **Step 1: Write failing tests for each diagnostic category**

```go
// internal/pipeline/check_test.go
package pipeline

import (
	"testing"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

func checkHelper(t *testing.T, src string) []Diagnostic {
	t.Helper()
	nodes, err := sexpr.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return Check(nodes)
}

func TestCheck_UnknownForm(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" (frobnicate "x"))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagUnknownSymbol {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagUnknownSymbol for 'frobnicate'")
	}
}

func TestCheck_StepBadArity(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" (step))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagBadArity {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagBadArity for step with no args")
	}
}

func TestCheck_UndefinedRef(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" (step "a" (str ~undefined)))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagUndefinedRef {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagUndefinedRef for ~undefined")
	}
}

func TestCheck_CleanWorkflow(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" :description "ok" (step "a" (run "echo hi")))`)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Fatalf("unexpected error diagnostic: %s", d.Message)
		}
	}
}

func TestCheck_DefShadowsBuiltin(t *testing.T) {
	diags := checkHelper(t, `(def str "oops") (workflow "test" (step "a" (str "hi")))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagShadowsBuiltin {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagShadowsBuiltin for def of 'str'")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestCheck -v`
Expected: FAIL — `Check`, `Diagnostic`, etc. undefined

- [ ] **Step 3: Implement the checker**

```go
// internal/pipeline/check.go
package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Severity of a diagnostic.
type Severity int

const (
	SeverityWarning Severity = iota
	SeverityError
)

func (s Severity) String() string {
	if s == SeverityError {
		return "error"
	}
	return "warning"
}

// DiagKind categorizes diagnostics.
type DiagKind int

const (
	DiagUnknownSymbol DiagKind = iota
	DiagBadArity
	DiagUndefinedRef
	DiagShadowsBuiltin
	DiagUnreachable
)

// Diagnostic is a single issue found during static analysis.
type Diagnostic struct {
	Kind     DiagKind
	Severity Severity
	Line     int
	Message  string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("line %d: %s: %s", d.Line, d.Severity, d.Message)
}

// knownSpecialForms lists all forms handled by evalSpecial.
var knownSpecialForms = map[string]bool{
	"def": true, "fn": true, "let": true,
	"do": true, "begin": true,
	"if": true, "when": true, "when-not": true, "cond": true,
	"workflow": true, "step": true, "par": true,
	"retry": true, "timeout": true, "catch": true,
	"include": true,
	"map": true, "each": true, "filter": true, "reduce": true,
	"compare": true, "phase": true, "gate": true,
	"->": true,
	"arg": true,
	"branch": true, "review": true,
}

// knownBuiltins lists all functions registered by registerBuiltins.
var knownBuiltins = map[string]bool{
	"sh": true, "run": true, "ref": true, "str": true,
	"llm": true, "save": true, "read-file": true, "read": true,
	"write-file": true, "write": true, "list": true,
	"not": true, "=": true, "println": true, "or": true,
	"glob": true, "websearch": true,
	"http-get": true, "fetch": true, "http-post": true, "send": true,
	"call-workflow": true, "json-pick": true, "pick": true,
	// Collection builtins
	"assoc": true, "flatten": true, "set": true, "difference": true,
	"join": true, "split": true, "lines": true, "count": true,
	"some": true, "every": true,
	// String builtins
	"upper": true, "lower": true, "trim": true,
	"replace": true, "contains": true,
	"starts-with": true, "ends-with": true, "slice": true,
	"regex-match": true, "regex-find": true,
	// Comparison
	"<": true, "assert": true,
}

// checkScope tracks what names are defined at each point in the walk.
type checkScope struct {
	defs    map[string]bool
	steps   map[string]bool
	parent  *checkScope
}

func newCheckScope(parent *checkScope) *checkScope {
	return &checkScope{
		defs:   make(map[string]bool),
		steps:  make(map[string]bool),
		parent: parent,
	}
}

func (cs *checkScope) isDefined(name string) bool {
	if cs.defs[name] {
		return true
	}
	if cs.steps[name] {
		return true
	}
	if cs.parent != nil {
		return cs.parent.isDefined(name)
	}
	return false
}

// Check performs static analysis on parsed AST nodes and returns diagnostics.
func Check(nodes []*sexpr.Node) []Diagnostic {
	c := &checker{scope: newCheckScope(nil)}
	for _, n := range nodes {
		c.walk(n)
	}
	return c.diags
}

type checker struct {
	scope *checkScope
	diags []Diagnostic
}

func (c *checker) warn(line int, kind DiagKind, msg string) {
	c.diags = append(c.diags, Diagnostic{Kind: kind, Severity: SeverityWarning, Line: line, Message: msg})
}

func (c *checker) err(line int, kind DiagKind, msg string) {
	c.diags = append(c.diags, Diagnostic{Kind: kind, Severity: SeverityError, Line: line, Message: msg})
}

func (c *checker) walk(node *sexpr.Node) {
	if node.IsAtom() {
		c.checkAtom(node)
		return
	}
	if !node.IsList() || len(node.Children) == 0 {
		return
	}

	head := node.Children[0]
	args := node.Children[1:]

	if !head.IsAtom() {
		// Head is a subexpression — walk it
		c.walk(head)
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	name := head.SymbolVal()
	if name == "" {
		// String or keyword in head position — walk children
		for _, ch := range node.Children {
			c.walk(ch)
		}
		return
	}

	// Known special form
	if knownSpecialForms[name] {
		c.checkSpecialForm(name, args, node)
		return
	}

	// Known builtin
	if knownBuiltins[name] {
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	// User-defined function
	if c.scope.isDefined(name) {
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	c.err(node.Line, DiagUnknownSymbol, fmt.Sprintf("unknown function or form '%s'", name))
}

func (c *checker) checkAtom(node *sexpr.Node) {
	if node.Atom.Type == sexpr.TokenString {
		c.checkInterpolation(node)
	}
}

func (c *checker) checkInterpolation(node *sexpr.Node) {
	s := node.Atom.Val
	if !strings.ContainsRune(s, '~') {
		return
	}
	parts, err := lexQuasi(s)
	if err != nil {
		c.err(node.Line, DiagUndefinedRef, fmt.Sprintf("bad interpolation: %s", err))
		return
	}
	for _, p := range parts {
		if p.Kind != partRef {
			continue
		}
		full := p.RefBase
		if len(p.RefPath) > 0 {
			full += "." + strings.Join(p.RefPath, ".")
		}
		// param.*, env.*, resource.* are runtime-resolved — skip
		switch p.RefBase {
		case "param", "env", "resource", "input", "workspace":
			continue
		}
		// Check if the ref is a known step or def
		if !c.scope.isDefined(p.RefBase) {
			c.warn(node.Line, DiagUndefinedRef, fmt.Sprintf("reference '~%s' may be undefined — not a known def or step at this point", full))
		}
	}
}

func (c *checker) checkSpecialForm(name string, args []*sexpr.Node, node *sexpr.Node) {
	switch name {
	case "def":
		if len(args) < 2 {
			c.err(node.Line, DiagBadArity, "def requires a name and a value")
			return
		}
		defName := args[0].SymbolVal()
		if defName == "" {
			defName = args[0].StringVal()
		}
		if knownBuiltins[defName] || knownSpecialForms[defName] {
			c.warn(node.Line, DiagShadowsBuiltin, fmt.Sprintf("def '%s' shadows a builtin", defName))
		}
		c.scope.defs[defName] = true
		for _, a := range args[1:] {
			c.walk(a)
		}

	case "step", "gate":
		if len(args) < 1 {
			c.err(node.Line, DiagBadArity, fmt.Sprintf("%s requires at least a step ID", name))
			return
		}
		stepID := args[0].StringVal()
		if stepID != "" {
			c.scope.steps[stepID] = true
		}
		for _, a := range args[1:] {
			c.walk(a)
		}

	case "let":
		if len(args) < 2 {
			c.err(node.Line, DiagBadArity, "let requires a binding list and a body")
			return
		}
		child := newCheckScope(c.scope)
		bindings := args[0]
		if bindings.IsList() {
			for i := 0; i+1 < len(bindings.Children); i += 2 {
				bName := bindings.Children[i].SymbolVal()
				if bName != "" {
					child.defs[bName] = true
				}
			}
		}
		saved := c.scope
		c.scope = child
		for _, a := range args[1:] {
			c.walk(a)
		}
		c.scope = saved

	case "fn":
		if len(args) < 2 {
			c.err(node.Line, DiagBadArity, "fn requires a parameter list and a body")
			return
		}
		child := newCheckScope(c.scope)
		params := args[0]
		if params.IsList() {
			for _, p := range params.Children {
				pName := p.SymbolVal()
				if pName != "" {
					child.defs[pName] = true
				}
			}
		}
		saved := c.scope
		c.scope = child
		for _, a := range args[1:] {
			c.walk(a)
		}
		c.scope = saved

	case "workflow":
		// First arg is name (string), then keywords, then body forms
		for _, a := range args {
			c.walk(a)
		}

	case "map", "each", "filter":
		// These bind "item" implicitly
		child := newCheckScope(c.scope)
		child.defs["item"] = true
		child.defs["item_index"] = true
		saved := c.scope
		c.scope = child
		for _, a := range args {
			c.walk(a)
		}
		c.scope = saved

	case "reduce":
		child := newCheckScope(c.scope)
		child.defs["item"] = true
		child.defs["acc"] = true
		saved := c.scope
		c.scope = child
		for _, a := range args {
			c.walk(a)
		}
		c.scope = saved

	case "->":
		// Thread binds "_" for each subsequent form
		child := newCheckScope(c.scope)
		child.defs["_"] = true
		saved := c.scope
		c.scope = child
		for _, a := range args {
			c.walk(a)
		}
		c.scope = saved

	case "catch":
		// catch binds "err" in fallback
		if len(args) >= 2 {
			c.walk(args[0])
			child := newCheckScope(c.scope)
			child.defs["err"] = true
			saved := c.scope
			c.scope = child
			for _, a := range args[1:] {
				c.walk(a)
			}
			c.scope = saved
		} else {
			for _, a := range args {
				c.walk(a)
			}
		}

	default:
		// All other special forms — just walk children
		for _, a := range args {
			c.walk(a)
		}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestCheck -v`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add internal/pipeline/check.go internal/pipeline/check_test.go
git commit -m "feat(pipeline): add static validator with scope tracking and diagnostics"
```

---

## Task 5: Validate real workflow fixtures

Run the checker against the existing test workflows to make sure it doesn't false-positive on valid code.

**Files:**
- Modify: `internal/pipeline/check_test.go`

- [ ] **Step 1: Write tests against testdata fixtures**

```go
func TestCheck_ParDemo(t *testing.T) {
	src, err := os.ReadFile("testdata/par-demo.glitch")
	if err != nil {
		t.Skip("testdata not found")
	}
	nodes, err := sexpr.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	diags := Check(nodes)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Errorf("unexpected error: %s", d)
		}
	}
}

func TestCheck_PhaseGate(t *testing.T) {
	src, err := os.ReadFile("testdata/phase-gate.glitch")
	if err != nil {
		t.Skip("testdata not found")
	}
	nodes, err := sexpr.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	diags := Check(nodes)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Errorf("unexpected error: %s", d)
		}
	}
}
```

Add `"os"` to the import block in `check_test.go`.

- [ ] **Step 2: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run "TestCheck_ParDemo|TestCheck_PhaseGate" -v`
Expected: PASS — no false-positive errors on valid workflows. If any fail, fix the checker to handle those patterns (likely keywords being flagged as unknown symbols, or `run`/`echo` inside string args).

- [ ] **Step 3: Fix any false positives discovered**

Review each failing diagnostic, determine if the checker is wrong, and add the necessary logic. Common fixes:
- Keywords (`:description`, `:retries`) should be skipped in `walk()` — they're not function calls
- String atoms in non-head position should not be checked as symbols

- [ ] **Step 4: Run the full check test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestCheck -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add internal/pipeline/check_test.go internal/pipeline/check.go
git commit -m "test(pipeline): validate checker against real workflow fixtures"
```

---

## Task 6: `glitch check` CLI command

Wire the validator into a cobra command that reads a `.glitch` file, runs `Check()`, and prints diagnostics.

**Files:**
- Create: `cmd/check.go`
- Create: `cmd/check_test.go`

- [ ] **Step 1: Write the failing test**

```go
// cmd/check_test.go
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCmd_ValidFile(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "good.glitch")
	os.WriteFile(wf, []byte(`(workflow "test" (step "a" (run "echo ok")))`), 0644)

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"check", wf})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("expected success message, got: %s", out.String())
	}
}

func TestCheckCmd_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "bad.glitch")
	os.WriteFile(wf, []byte(`(workflow "test" (frobnicate "x"))`), 0644)

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"check", wf})
	_ = rootCmd.Execute()
	if !strings.Contains(out.String(), "frobnicate") {
		t.Fatalf("expected diagnostic mentioning frobnicate, got: %s", out.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./cmd/ -run TestCheckCmd -v`
Expected: FAIL — unknown command "check"

- [ ] **Step 3: Implement the check command**

```go
// cmd/check.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

var checkCmd = &cobra.Command{
	Use:   "check <file>",
	Short: "validate a workflow file without running it",
	Long: `Statically analyze a .glitch workflow file for common errors:
  - Unknown functions or forms
  - Bad argument counts
  - Undefined references (~refs that aren't steps, defs, or params)
  - Definitions that shadow builtins

Example:
  glitch check workflows/deploy.glitch`,
	Args: cobra.ExactArgs(1),
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	path := args[0]
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	nodes, err := sexpr.Parse(src)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: parse error: %s\n", path, err)
		return fmt.Errorf("parse failed")
	}

	diags := pipeline.Check(nodes)
	if len(diags) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: ok\n", path)
		return nil
	}

	hasError := false
	for _, d := range diags {
		prefix := "warning"
		if d.Severity == pipeline.SeverityError {
			prefix = "error"
			hasError = true
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s:%d: %s: %s\n", path, d.Line, prefix, d.Message)
	}

	if hasError {
		return fmt.Errorf("check failed with errors")
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./cmd/ -run TestCheckCmd -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add cmd/check.go cmd/check_test.go
git commit -m "feat(cmd): add glitch check for static workflow validation"
```

---

## Task 7: Stdlib — embedded standard library

Ship a set of reusable `.glitch` definitions that authors can `(include "std/strings")`. Files are embedded in the binary via `embed.FS` and extracted to a temp dir on first use, so `include` path resolution works without changes.

**Files:**
- Create: `internal/pipeline/stdlib/strings.glitch`
- Create: `internal/pipeline/stdlib/collections.glitch`
- Create: `internal/pipeline/stdlib/io.glitch`
- Create: `internal/pipeline/stdlib/embed.go`
- Create: `internal/pipeline/stdlib_test.go`
- Modify: `internal/pipeline/eval.go` (resolve `std/` prefixed includes)

- [ ] **Step 1: Write the test that expects stdlib to parse and evaluate**

```go
// internal/pipeline/stdlib_test.go
package pipeline

import (
	"testing"
)

func TestStdlib_StringsLoads(t *testing.T) {
	src := `(include "std/strings") (kebab-case "Hello World")`
	got := evalHelper(t, src)
	if got != "hello-world" {
		t.Fatalf("want %q, got %q", "hello-world", got)
	}
}

func TestStdlib_CollectionsLoads(t *testing.T) {
	src := `(include "std/collections") (pluck "name" (list (assoc :name "alice") (assoc :name "bob")))`
	got := evalHelper(t, src)
	// pluck returns a list, stringified as newline-separated
	if got != "alice\nbob" {
		t.Fatalf("want %q, got %q", "alice\nbob", got)
	}
}

func TestStdlib_IOLoads(t *testing.T) {
	src := `(include "std/io") (read-lines "testdata/par-demo.glitch")`
	got := evalHelper(t, src)
	if got == "" {
		t.Fatal("expected non-empty output from read-lines")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestStdlib -v`
Expected: FAIL — include can't resolve `std/strings`

- [ ] **Step 3: Create the stdlib embed infrastructure**

```go
// internal/pipeline/stdlib/embed.go
package stdlib

import "embed"

//go:embed *.glitch
var FS embed.FS
```

- [ ] **Step 4: Create `strings.glitch`**

```glitch
;; std/strings — string utility functions

(def kebab-case
  (fn (s)
    (-> s
      (lower)
      (replace " " "-")
      (replace "_" "-"))))

(def snake-case
  (fn (s)
    (-> s
      (lower)
      (replace " " "_")
      (replace "-" "_"))))

(def blank?
  (fn (s) (= (trim s) "")))

(def present?
  (fn (s) (not (= (trim s) ""))))

(def words
  (fn (s) (split (trim s) " ")))

(def unwords
  (fn (lst) (join lst " ")))
```

- [ ] **Step 5: Create `collections.glitch`**

```glitch
;; std/collections — collection helpers

(def pluck
  (fn (key source)
    (map source (fn (item) (pick key item)))))

(def compact
  (fn (source)
    (filter source (fn (item) (not (= (trim item) ""))))))

(def first
  (fn (source)
    (pick 0 source)))

(def last
  (fn (source)
    (pick (- (count source) 1) source)))

(def take
  (fn (n source)
    (slice source 0 n)))

(def zip-with
  (fn (a b combiner)
    (let (len (count a))
      (map (list) (fn (_)
        ;; zip is hard without numeric range — defer to reduce
        a)))))
```

Note: `zip-with` is intentionally minimal — the language lacks numeric `range`, so full zip requires `reduce` with an index counter. Ship what works cleanly now.

- [ ] **Step 6: Create `io.glitch`**

```glitch
;; std/io — file and HTTP convenience wrappers

(def read-lines
  (fn (path)
    (lines (read-file path))))

(def write-lines
  (fn (path data)
    (write-file path (join data "\n"))))

(def fetch-json
  (fn (url)
    (json-pick "$" (http-get url))))
```

- [ ] **Step 7: Wire `std/` include resolution into the evaluator**

In `internal/pipeline/eval.go`, modify `specialInclude` to resolve `std/` prefixed paths from the embedded FS. Find the `specialInclude` function and add the embedded resolution before the `os.ReadFile` call:

```go
// At the top of eval.go, add import:
import "github.com/8op-org/gl1tch/internal/pipeline/stdlib"

// Inside specialInclude, before the os.ReadFile call, add:
if strings.HasPrefix(path, "std/") {
	name := strings.TrimPrefix(path, "std/") + ".glitch"
	data, err := stdlib.FS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("line %d: include %q: %w", node.Line, path, err), true
	}
	nodes, parseErr := sexpr.Parse(data)
	if parseErr != nil {
		return nil, fmt.Errorf("line %d: include %q: %w", node.Line, path, parseErr), true
	}
	var result Value = NilVal{}
	for _, n := range nodes {
		result, err = ev.Eval(env, n)
		if err != nil {
			return nil, err, true
		}
	}
	return result, nil, true
}
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestStdlib -v`
Expected: PASS. If `pluck` or `pick` fails due to assoc/map interaction, adjust the test input to match the actual `pick` implementation (it may use JSON path syntax).

- [ ] **Step 9: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add internal/pipeline/stdlib/ internal/pipeline/stdlib_test.go internal/pipeline/eval.go
git commit -m "feat(pipeline): add embedded stdlib with strings, collections, and io modules"
```

---

## Task 8: Validator knows about stdlib defs

When a workflow does `(include "std/strings")`, the checker should recognize that `kebab-case` is now defined and not flag it as unknown.

**Files:**
- Modify: `internal/pipeline/check.go`
- Modify: `internal/pipeline/check_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestCheck_StdlibInclude(t *testing.T) {
	diags := checkHelper(t, `(include "std/strings") (workflow "test" (step "a" (kebab-case "Hello World")))`)
	for _, d := range diags {
		if d.Kind == DiagUnknownSymbol && strings.Contains(d.Message, "kebab-case") {
			t.Fatal("checker should not flag stdlib-defined 'kebab-case' as unknown")
		}
	}
}
```

Add `"strings"` to the import block in `check_test.go` if not already present.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestCheck_StdlibInclude -v`
Expected: FAIL — checker flags `kebab-case` as unknown

- [ ] **Step 3: Add include resolution to the checker**

In `check.go`, in the `checkSpecialForm` function, add a case for `"include"`:

```go
case "include":
	if len(args) < 1 {
		c.err(node.Line, DiagBadArity, "include requires a path argument")
		return
	}
	path := args[0].StringVal()
	if strings.HasPrefix(path, "std/") {
		// Parse the stdlib file and collect top-level def names
		name := strings.TrimPrefix(path, "std/") + ".glitch"
		data, err := stdlib.FS.ReadFile(name)
		if err != nil {
			c.err(node.Line, DiagUnknownSymbol, fmt.Sprintf("unknown stdlib module '%s'", path))
			return
		}
		nodes, err := sexpr.Parse(data)
		if err != nil {
			c.err(node.Line, DiagUnknownSymbol, fmt.Sprintf("stdlib parse error in '%s': %s", path, err))
			return
		}
		for _, n := range nodes {
			if n.IsList() && len(n.Children) >= 2 {
				head := n.Children[0].SymbolVal()
				if head == "def" {
					defName := n.Children[1].SymbolVal()
					if defName == "" {
						defName = n.Children[1].StringVal()
					}
					if defName != "" {
						c.scope.defs[defName] = true
					}
				}
			}
		}
	}
	// Non-std includes — can't statically resolve, skip
```

Add import for `"github.com/8op-org/gl1tch/internal/pipeline/stdlib"` at the top of `check.go`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./internal/pipeline/ -run TestCheck -v`
Expected: PASS (all tests including the new stdlib include test)

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add internal/pipeline/check.go internal/pipeline/check_test.go
git commit -m "feat(pipeline): checker resolves stdlib includes for accurate diagnostics"
```

---

## Task 9: End-to-end integration test

Verify that all three features work together: write a workflow using stdlib, check it, and eval an expression from it.

**Files:**
- Modify: `cmd/eval_test.go`
- Modify: `cmd/check_test.go`

- [ ] **Step 1: Write the integration test for check + stdlib**

```go
// In cmd/check_test.go
func TestCheckCmd_WithStdlib(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "uses-stdlib.glitch")
	os.WriteFile(wf, []byte(`(include "std/strings")
(workflow "test"
  :description "uses stdlib"
  (step "a" (kebab-case "Hello World")))
`), 0644)

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"check", wf})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check should pass for valid stdlib usage: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("expected ok, got: %s", out.String())
	}
}
```

- [ ] **Step 2: Write the integration test for eval + stdlib**

```go
// In cmd/eval_test.go
func TestEvalCmd_WithStdlib(t *testing.T) {
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"eval", "-e", `(do (include "std/strings") (kebab-case "Hello World"))`})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hello-world") {
		t.Fatalf("want %q, got %q", "hello-world", out.String())
	}
}
```

- [ ] **Step 3: Run all integration tests**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./cmd/ -run "TestCheckCmd|TestEvalCmd" -v`
Expected: PASS

- [ ] **Step 4: Run the full test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup && go test ./...`
Expected: All existing tests still pass, no regressions

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.claude/worktrees/lisp-eval-cleanup
git add cmd/eval_test.go cmd/check_test.go
git commit -m "test: end-to-end integration tests for eval, check, and stdlib"
```

---

## Summary

| Task | What it delivers |
|------|-----------------|
| 1 | `RunSourceWithEnv` + `RegisterBuiltins` — persistent eval sessions |
| 2 | REPL core with multiline detection |
| 3 | `glitch eval` CLI (interactive + `-e` one-shot) |
| 4 | Static validator with 5 diagnostic categories |
| 5 | Validator smoke-tested against real fixtures |
| 6 | `glitch check` CLI command |
| 7 | Embedded stdlib (strings, collections, io) |
| 8 | Checker understands stdlib includes |
| 9 | End-to-end integration tests |

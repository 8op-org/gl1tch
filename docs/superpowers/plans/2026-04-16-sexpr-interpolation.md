# Sexpr Interpolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Go `text/template` inside `.glitch` workflow strings with Clojure-style sexpr unquote (`~name`, `~param.x`, `~(form)`), failing loud on undefined refs, with a hard cut-over of all existing workflow files.

**Architecture:** Keep `Step` struct fields as `string`. Rewrite the single `render()` entry point in `internal/pipeline/runner.go` to parse `~` interpolation on demand and evaluate forms against a new `Scope` type. The eleven existing string helpers plus `pick`/`assoc` move from `text/template` funcs to a sexpr evaluator builtins table. All existing `.glitch` files are rewritten by a one-shot script checked in with the PR and deleted at merge.

**Tech Stack:** Go, existing `internal/sexpr` lexer/parser.

**Spec:** `docs/superpowers/specs/2026-04-16-sexpr-interpolation-design.md`

---

## File structure

New files:
- `internal/pipeline/scope.go` — `Scope` struct, resolution order, `UndefinedRefError`
- `internal/pipeline/scope_test.go`
- `internal/pipeline/quasi.go` — quasi-string tokenizer + renderer (`renderQuasi`)
- `internal/pipeline/quasi_test.go`
- `internal/pipeline/builtins.go` — sexpr builtins (`step`, `stepfile`, `pick`, string helpers, `or`, …)
- `internal/pipeline/builtins_test.go`
- `scripts/rewrite-quasi/main.go` — one-shot `{{…}}` → `~…` rewrite tool (deleted at merge)

Modified files:
- `internal/pipeline/runner.go` — rewrite `render()` body; delete `text/template` import, funcMap, and `Parse/Execute` plumbing (lines 752-903 today)
- `internal/pipeline/render_test.go` — rewrite tests to use new syntax
- `internal/pipeline/runner_test.go` — fix any test fixtures that use `{{…}}`
- `internal/pipeline/compare_review.go` / `_test.go` — audit for `{{…}}`
- `internal/pipeline/plugin_runner.go` / `_test.go` — audit for `{{…}}`
- All `.glitch` files under `examples/`, `test-workspace/workflows/`, `internal/pipeline/testdata/`, `.glitch/workflows/` — rewritten by the script

No changes to `internal/sexpr/` — the lexer already emits `TokenString` for both `"…"` and ```` ``` ``` ````. We don't need a new token type because interpolation is a render-time concern per the spec update (all strings interpolate uniformly).

---

## Task 1: Scope type + UndefinedRefError

**Files:**
- Create: `internal/pipeline/scope.go`
- Create: `internal/pipeline/scope_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/pipeline/scope_test.go
package pipeline

import (
	"strings"
	"testing"
)

func TestScopeResolvesBareSymbol(t *testing.T) {
	s := NewScope()
	s.SetLet("model", "qwen2.5:7b")
	v, err := s.Resolve("model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "qwen2.5:7b" {
		t.Errorf("got %q, want %q", v, "qwen2.5:7b")
	}
}

func TestScopeLetShadowsDef(t *testing.T) {
	s := NewScope()
	s.SetDef("model", "llama3")
	s.SetLet("model", "qwen2.5:7b")
	v, _ := s.Resolve("model")
	if v != "qwen2.5:7b" {
		t.Errorf("let should shadow def; got %q", v)
	}
}

func TestScopeUndefinedSymbolReturnsError(t *testing.T) {
	s := NewScope()
	s.SetDef("model", "llama3")
	_, err := s.Resolve("modle")
	if err == nil {
		t.Fatal("expected UndefinedRefError, got nil")
	}
	var ure *UndefinedRefError
	if !errorsAs(err, &ure) {
		t.Fatalf("want UndefinedRefError, got %T", err)
	}
	if !strings.Contains(err.Error(), "modle") {
		t.Errorf("error should mention symbol; got: %v", err)
	}
	if !strings.Contains(err.Error(), "model") {
		t.Errorf("error should include suggestion; got: %v", err)
	}
}

// errorsAs is a tiny wrapper so the test file doesn't need to import "errors".
func errorsAs(err error, target any) bool {
	return errorsAsStd(err, target)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/pipeline/ -run TestScope -v`
Expected: FAIL (`NewScope`, `UndefinedRefError` not defined).

- [ ] **Step 3: Implement minimal code**

```go
// internal/pipeline/scope.go
package pipeline

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

type Scope struct {
	lets     map[string]string
	defs     map[string]string
	params   map[string]string
	item     string
	itemIdx  int
	hasItem  bool
	steps    map[string]string
	input    string
	hasInput bool
}

func NewScope() *Scope {
	return &Scope{
		lets:   map[string]string{},
		defs:   map[string]string{},
		params: map[string]string{},
		steps:  map[string]string{},
	}
}

func (s *Scope) SetLet(name, val string)    { s.lets[name] = val }
func (s *Scope) SetDef(name, val string)    { s.defs[name] = val }
func (s *Scope) SetParam(name, val string)  { s.params[name] = val }
func (s *Scope) SetInput(v string)          { s.input = v; s.hasInput = true }
func (s *Scope) SetItem(v string, idx int)  { s.item = v; s.itemIdx = idx; s.hasItem = true }
func (s *Scope) SetSteps(st map[string]string) {
	s.steps = st
}

// Resolve looks up a bare symbol in precedence order: let, def, specials.
func (s *Scope) Resolve(name string) (string, error) {
	if v, ok := s.lets[name]; ok {
		return v, nil
	}
	if v, ok := s.defs[name]; ok {
		return v, nil
	}
	switch name {
	case "input":
		if s.hasInput {
			return s.input, nil
		}
	case "item":
		if s.hasItem {
			return s.item, nil
		}
	case "item_index":
		if s.hasItem {
			return fmt.Sprintf("%d", s.itemIdx), nil
		}
	}
	return "", &UndefinedRefError{Symbol: name, Suggestion: s.suggest(name)}
}

// ResolvePath resolves dotted paths like "param.repo" or "env.HOME".
func (s *Scope) ResolvePath(base string, path []string) (string, error) {
	switch base {
	case "param":
		if len(path) != 1 {
			return "", &UndefinedRefError{Symbol: "param." + joinDots(path), Suggestion: "param.x only supports one level"}
		}
		if v, ok := s.params[path[0]]; ok {
			return v, nil
		}
		return "", &UndefinedRefError{Symbol: "param." + path[0]}
	case "env":
		if len(path) != 1 {
			return "", &UndefinedRefError{Symbol: "env." + joinDots(path)}
		}
		return os.Getenv(path[0]), nil
	}
	return "", &UndefinedRefError{Symbol: base + "." + joinDots(path)}
}

func joinDots(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "."
		}
		out += p
	}
	return out
}

// suggest returns the closest known symbol by Levenshtein distance, or "".
func (s *Scope) suggest(name string) string {
	candidates := []string{}
	for k := range s.lets {
		candidates = append(candidates, k)
	}
	for k := range s.defs {
		candidates = append(candidates, k)
	}
	candidates = append(candidates, "input", "item", "item_index")

	sort.Strings(candidates)
	best := ""
	bestDist := len(name) + 1
	for _, c := range candidates {
		d := levenshtein(name, c)
		if d < bestDist && d <= 2 {
			bestDist = d
			best = c
		}
	}
	return best
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d1 := prev[j] + 1
			d2 := curr[j-1] + 1
			d3 := prev[j-1] + cost
			m := d1
			if d2 < m {
				m = d2
			}
			if d3 < m {
				m = d3
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

type UndefinedRefError struct {
	Symbol     string
	Suggestion string
	File       string
	Line       int
	Col        int
}

func (e *UndefinedRefError) Error() string {
	loc := ""
	if e.File != "" {
		loc = fmt.Sprintf(" at %s:%d:%d", e.File, e.Line, e.Col)
	}
	sug := ""
	if e.Suggestion != "" {
		sug = fmt.Sprintf(" (did you mean '%s'?)", e.Suggestion)
	}
	return fmt.Sprintf("undefined reference '%s'%s%s", e.Symbol, loc, sug)
}

// errorsAsStd wraps errors.As for test visibility.
func errorsAsStd(err error, target any) bool {
	return errors.As(err, target)
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go test ./internal/pipeline/ -run TestScope -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/scope.go internal/pipeline/scope_test.go
git commit -m "feat(pipeline): Scope type with let/def/specials and UndefinedRefError

Adds the resolution primitive for sexpr-level unquote. Precedence is
let > def > specials (input/item/item_index). Undefined refs return
UndefinedRefError with a Levenshtein-based suggestion."
```

---

## Task 2: Quasi-string lexer (literal + bare symbol)

**Files:**
- Create: `internal/pipeline/quasi.go`
- Create: `internal/pipeline/quasi_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/pipeline/quasi_test.go
package pipeline

import "testing"

func TestLexQuasiLiteralOnly(t *testing.T) {
	parts, err := lexQuasi("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("want 1 part, got %d", len(parts))
	}
	if parts[0].Kind != partLiteral || parts[0].Literal != "hello world" {
		t.Errorf("got %+v", parts[0])
	}
}

func TestLexQuasiBareRef(t *testing.T) {
	parts, err := lexQuasi("hi ~name there")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []quasiPart{
		{Kind: partLiteral, Literal: "hi "},
		{Kind: partRef, RefBase: "name"},
		{Kind: partLiteral, Literal: " there"},
	}
	assertParts(t, parts, want)
}

func TestLexQuasiDottedRef(t *testing.T) {
	parts, _ := lexQuasi("repo=~param.repo x")
	want := []quasiPart{
		{Kind: partLiteral, Literal: "repo="},
		{Kind: partRef, RefBase: "param", RefPath: []string{"repo"}},
		{Kind: partLiteral, Literal: " x"},
	}
	assertParts(t, parts, want)
}

func TestLexQuasiEscapedTilde(t *testing.T) {
	parts, _ := lexQuasi(`cp a \~/dest`)
	if len(parts) != 1 {
		t.Fatalf("want 1 part, got %d", len(parts))
	}
	if parts[0].Literal != "cp a ~/dest" {
		t.Errorf("got literal %q", parts[0].Literal)
	}
}

func TestLexQuasiNoInterpolationWithoutTilde(t *testing.T) {
	parts, _ := lexQuasi("plain text no refs")
	if len(parts) != 1 || parts[0].Kind != partLiteral {
		t.Fatalf("got %+v", parts)
	}
}

func assertParts(t *testing.T, got, want []quasiPart) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len: got %d, want %d; got=%+v", len(got), len(want), got)
	}
	for i := range got {
		if got[i].Kind != want[i].Kind ||
			got[i].Literal != want[i].Literal ||
			got[i].RefBase != want[i].RefBase ||
			!stringSlicesEqual(got[i].RefPath, want[i].RefPath) {
			t.Errorf("part %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Run tests, expect failure**

Run: `go test ./internal/pipeline/ -run TestLexQuasi -v`
Expected: FAIL (symbols undefined).

- [ ] **Step 3: Implement**

```go
// internal/pipeline/quasi.go
package pipeline

import (
	"fmt"
	"strings"
)

type quasiPartKind int

const (
	partLiteral quasiPartKind = iota
	partRef
	partForm
)

type quasiPart struct {
	Kind    quasiPartKind
	Literal string   // for partLiteral
	RefBase string   // for partRef: bare symbol or dotted base
	RefPath []string // for partRef: dotted trailing components
	Form    string   // for partForm: raw "( ... )" source including outer parens
	Line    int
	Col     int
}

// lexQuasi scans a raw string for ~... interpolations.
// Returns an ordered list of literal/ref/form parts. No-tilde strings return
// a single literal part.
func lexQuasi(src string) ([]quasiPart, error) {
	var parts []quasiPart
	var lit strings.Builder
	flushLit := func() {
		if lit.Len() > 0 {
			parts = append(parts, quasiPart{Kind: partLiteral, Literal: lit.String()})
			lit.Reset()
		}
	}

	i := 0
	for i < len(src) {
		ch := src[i]
		if ch == '\\' && i+1 < len(src) && src[i+1] == '~' {
			lit.WriteByte('~')
			i += 2
			continue
		}
		if ch != '~' {
			lit.WriteByte(ch)
			i++
			continue
		}
		// ch == '~'
		if i+1 >= len(src) {
			lit.WriteByte('~')
			i++
			continue
		}
		next := src[i+1]
		if next == '(' {
			end, err := findFormEnd(src, i+1)
			if err != nil {
				return nil, err
			}
			flushLit()
			parts = append(parts, quasiPart{Kind: partForm, Form: src[i+1 : end+1]})
			i = end + 1
			continue
		}
		if isRefStartByte(next) {
			j := i + 1
			for j < len(src) && isRefByte(src[j]) {
				j++
			}
			ref := src[i+1 : j]
			base, path := splitDotted(ref)
			flushLit()
			parts = append(parts, quasiPart{Kind: partRef, RefBase: base, RefPath: path})
			i = j
			continue
		}
		// Lone ~ not followed by a ref start — treat as literal.
		lit.WriteByte('~')
		i++
	}
	flushLit()
	return parts, nil
}

func isRefStartByte(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isRefByte(ch byte) bool {
	return isRefStartByte(ch) || (ch >= '0' && ch <= '9') || ch == '.' || ch == '-'
}

func splitDotted(ref string) (base string, path []string) {
	if !strings.Contains(ref, ".") {
		return ref, nil
	}
	parts := strings.Split(ref, ".")
	return parts[0], parts[1:]
}

// findFormEnd finds the matching ")" for the "(" at src[start].
// Tracks string literals to avoid miscounting parens inside "..." .
func findFormEnd(src string, start int) (int, error) {
	if src[start] != '(' {
		return 0, fmt.Errorf("findFormEnd: expected '(' at %d", start)
	}
	depth := 0
	i := start
	inString := false
	for i < len(src) {
		ch := src[i]
		if inString {
			if ch == '\\' && i+1 < len(src) {
				i += 2
				continue
			}
			if ch == '"' {
				inString = false
			}
			i++
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, nil
			}
		}
		i++
	}
	return 0, fmt.Errorf("unterminated ~( form starting at byte %d", start)
}
```

- [ ] **Step 4: Run tests, verify pass**

Run: `go test ./internal/pipeline/ -run TestLexQuasi -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/quasi.go internal/pipeline/quasi_test.go
git commit -m "feat(pipeline): lexQuasi tokenizes ~ref and ~(form) in strings

Scans a raw string for ~name, ~param.x, ~(form), and \\~ escape. Returns
ordered parts for the render pipeline to evaluate against a Scope."
```

---

## Task 3: Quasi-string lexer for `~(form)`

**Files:**
- Modify: `internal/pipeline/quasi_test.go`

- [ ] **Step 1: Add failing tests for ~(form)**

Append to `quasi_test.go`:

```go
func TestLexQuasiForm(t *testing.T) {
	parts, err := lexQuasi(`result: ~(step diff)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("want 2 parts, got %d: %+v", len(parts), parts)
	}
	if parts[0].Literal != "result: " {
		t.Errorf("part 0 literal: got %q", parts[0].Literal)
	}
	if parts[1].Kind != partForm || parts[1].Form != "(step diff)" {
		t.Errorf("part 1: got %+v", parts[1])
	}
}

func TestLexQuasiFormNested(t *testing.T) {
	parts, _ := lexQuasi(`~(upper (pick :title param.item))`)
	if len(parts) != 1 || parts[0].Kind != partForm {
		t.Fatalf("want 1 form part, got %+v", parts)
	}
	if parts[0].Form != "(upper (pick :title param.item))" {
		t.Errorf("got form %q", parts[0].Form)
	}
}

func TestLexQuasiFormUnterminated(t *testing.T) {
	_, err := lexQuasi(`oops ~(step diff`)
	if err == nil {
		t.Fatal("expected error on unterminated form")
	}
}

func TestLexQuasiFormWithStringLiteral(t *testing.T) {
	// paren inside a string literal must not confuse depth tracking
	parts, _ := lexQuasi(`~(join ")" xs)`)
	if len(parts) != 1 || parts[0].Kind != partForm {
		t.Fatalf("want 1 form part, got %+v", parts)
	}
	if parts[0].Form != `(join ")" xs)` {
		t.Errorf("got form %q", parts[0].Form)
	}
}
```

- [ ] **Step 2: Run tests, verify pass**

The Task 2 implementation already handles `~(form)`. Confirm:

Run: `go test ./internal/pipeline/ -run TestLexQuasiForm -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/pipeline/quasi_test.go
git commit -m "test(pipeline): lexQuasi covers ~(form) with nesting and literals"
```

---

## Task 4: Builtins dispatcher + step/stepfile/itemfile/branch/or

**Files:**
- Create: `internal/pipeline/builtins.go`
- Create: `internal/pipeline/builtins_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/pipeline/builtins_test.go
package pipeline

import "testing"

func TestBuiltinStep(t *testing.T) {
	s := NewScope()
	s.SetSteps(map[string]string{"fetch": "hello"})
	v, err := callBuiltin("step", []string{"fetch"}, s)
	if err != nil || v != "hello" {
		t.Errorf("got %q err=%v", v, err)
	}
}

func TestBuiltinStepUnknown(t *testing.T) {
	s := NewScope()
	s.SetSteps(map[string]string{})
	_, err := callBuiltin("step", []string{"missing"}, s)
	if err == nil {
		t.Fatal("expected error on unknown step")
	}
}

func TestBuiltinOr(t *testing.T) {
	s := NewScope()
	v, _ := callBuiltin("or", []string{"", "fallback"}, s)
	if v != "fallback" {
		t.Errorf("got %q", v)
	}
	v2, _ := callBuiltin("or", []string{"first", "second"}, s)
	if v2 != "first" {
		t.Errorf("got %q", v2)
	}
}

func TestBuiltinUnknown(t *testing.T) {
	_, err := callBuiltin("doesnotexist", nil, NewScope())
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 2: Run tests, expect failure**

Run: `go test ./internal/pipeline/ -run TestBuiltin -v`
Expected: FAIL (`callBuiltin` undefined).

- [ ] **Step 3: Implement**

```go
// internal/pipeline/builtins.go
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// callBuiltin dispatches a named builtin function with evaluated string args.
// Returns the result string or a clear error.
func callBuiltin(name string, args []string, scope *Scope) (string, error) {
	switch name {
	case "step":
		if len(args) == 0 {
			return "", fmt.Errorf("step: missing id argument")
		}
		v, ok := scope.steps[args[0]]
		if !ok {
			return "", fmt.Errorf("step: unknown step id %q", args[0])
		}
		return v, nil

	case "stepfile":
		if len(args) == 0 {
			return "", fmt.Errorf("stepfile: missing id")
		}
		content, ok := scope.steps[args[0]]
		if !ok {
			return "", fmt.Errorf("stepfile: unknown step id %q", args[0])
		}
		f, err := os.CreateTemp("", "glitch-step-*")
		if err != nil {
			return "", err
		}
		f.WriteString(content)
		f.Close()
		return f.Name(), nil

	case "itemfile":
		if !scope.hasItem {
			return "", fmt.Errorf("itemfile: no ~item in scope")
		}
		f, err := os.CreateTemp("", "glitch-item-*")
		if err != nil {
			return "", err
		}
		f.WriteString(scope.item)
		f.Close()
		return f.Name(), nil

	case "branch":
		if len(args) == 0 {
			return "", fmt.Errorf("branch: missing name")
		}
		for k, v := range scope.steps {
			if strings.HasSuffix(k, "/"+args[0]+"/__output") {
				return v, nil
			}
		}
		if v, ok := scope.steps[args[0]]; ok {
			return v, nil
		}
		return "", fmt.Errorf("branch: unknown branch %q", args[0])

	case "or":
		for _, a := range args {
			if a != "" {
				return a, nil
			}
		}
		return "", nil
	}

	// String helpers
	if v, ok, err := tryStringBuiltin(name, args); ok {
		return v, err
	}
	// JSON helpers
	if v, ok, err := tryJSONBuiltin(name, args); ok {
		return v, err
	}

	return "", fmt.Errorf("unknown builtin %q", name)
}

func tryStringBuiltin(name string, args []string) (string, bool, error) {
	// Implemented in Task 5.
	return "", false, nil
}

func tryJSONBuiltin(name string, args []string) (string, bool, error) {
	_ = json.Valid // keep import for next task
	return "", false, nil
}
```

- [ ] **Step 4: Run tests, verify pass**

Run: `go test ./internal/pipeline/ -run TestBuiltin -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/builtins.go internal/pipeline/builtins_test.go
git commit -m "feat(pipeline): callBuiltin dispatcher with step/stepfile/itemfile/branch/or

Moves the step/branch/stepfile/itemfile helpers from text/template funcMap
into a sexpr-evaluator builtins table. Adds (or ...) for optional-ref
fallback."
```

---

## Task 5: String builtins (11 helpers)

**Files:**
- Modify: `internal/pipeline/builtins.go`
- Modify: `internal/pipeline/builtins_test.go`

- [ ] **Step 1: Add failing tests**

Append to `builtins_test.go`:

```go
func TestStringBuiltins(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"upper", []string{"hello"}, "HELLO"},
		{"lower", []string{"HI"}, "hi"},
		{"trim", []string{"  x  "}, "x"},
		{"trimPrefix", []string{"pre-", "pre-thing"}, "thing"},
		{"trimSuffix", []string{"-end", "thing-end"}, "thing"},
		{"replace", []string{"/", "-", "a/b/c"}, "a-b-c"},
		{"truncate", []string{"5", "abcdefg"}, "abcde"},
		{"truncate noop", []string{"100", "abc"}, "abc"},
		{"contains", []string{"foobar", "foo"}, "true"},
		{"hasPrefix", []string{"elastic/x", "elastic"}, "true"},
		{"hasSuffix", []string{"elastic/x", "/x"}, "true"},
		{"split", []string{"/", "a/b/c"}, "a\nb\nc"},
		{"join", []string{"-", "a\nb\nc"}, "a-b-c"},
		{"first", []string{"a\nb\nc"}, "a"},
		{"last", []string{"a\nb\nc"}, "c"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := callBuiltin(strings_functionName(tc.name), tc.args, NewScope())
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// strings_functionName strips any " noop"/" short" suffix used only for test naming.
func strings_functionName(n string) string {
	if i := indexOf(n, ' '); i >= 0 {
		return n[:i]
	}
	return n
}

func indexOf(s string, ch byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ch {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: Run tests, expect failure**

Run: `go test ./internal/pipeline/ -run TestStringBuiltins -v`
Expected: FAIL (string helpers not implemented yet).

- [ ] **Step 3: Implement `tryStringBuiltin`**

Replace the stub `tryStringBuiltin` in `builtins.go` with:

```go
func tryStringBuiltin(name string, args []string) (string, bool, error) {
	switch name {
	case "upper":
		if len(args) < 1 {
			return "", true, fmt.Errorf("upper: missing arg")
		}
		return strings.ToUpper(args[0]), true, nil
	case "lower":
		if len(args) < 1 {
			return "", true, fmt.Errorf("lower: missing arg")
		}
		return strings.ToLower(args[0]), true, nil
	case "trim":
		if len(args) < 1 {
			return "", true, fmt.Errorf("trim: missing arg")
		}
		return strings.TrimSpace(args[0]), true, nil
	case "trimPrefix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("trimPrefix: need (prefix s)")
		}
		return strings.TrimPrefix(args[1], args[0]), true, nil
	case "trimSuffix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("trimSuffix: need (suffix s)")
		}
		return strings.TrimSuffix(args[1], args[0]), true, nil
	case "replace":
		if len(args) < 3 {
			return "", true, fmt.Errorf("replace: need (old new s)")
		}
		return strings.ReplaceAll(args[2], args[0], args[1]), true, nil
	case "truncate":
		if len(args) < 2 {
			return "", true, fmt.Errorf("truncate: need (n s)")
		}
		n := 0
		fmt.Sscanf(args[0], "%d", &n)
		runes := []rune(args[1])
		if len(runes) <= n {
			return args[1], true, nil
		}
		return string(runes[:n]), true, nil
	case "contains":
		if len(args) < 2 {
			return "", true, fmt.Errorf("contains: need (haystack needle)")
		}
		return boolStr(strings.Contains(args[0], args[1])), true, nil
	case "hasPrefix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("hasPrefix: need (s prefix)")
		}
		return boolStr(strings.HasPrefix(args[0], args[1])), true, nil
	case "hasSuffix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("hasSuffix: need (s suffix)")
		}
		return boolStr(strings.HasSuffix(args[0], args[1])), true, nil
	case "split":
		if len(args) < 2 {
			return "", true, fmt.Errorf("split: need (sep s)")
		}
		return strings.Join(strings.Split(args[1], args[0]), "\n"), true, nil
	case "join":
		if len(args) < 2 {
			return "", true, fmt.Errorf("join: need (sep s)")
		}
		lines := strings.Split(args[1], "\n")
		return strings.Join(lines, args[0]), true, nil
	case "first":
		if len(args) < 1 {
			return "", true, fmt.Errorf("first: missing arg")
		}
		lines := strings.Split(args[0], "\n")
		if len(lines) == 0 {
			return "", true, nil
		}
		return lines[0], true, nil
	case "last":
		if len(args) < 1 {
			return "", true, fmt.Errorf("last: missing arg")
		}
		lines := strings.Split(args[0], "\n")
		if len(lines) == 0 {
			return "", true, nil
		}
		return lines[len(lines)-1], true, nil
	}
	return "", false, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/pipeline/ -run TestStringBuiltins -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/builtins.go internal/pipeline/builtins_test.go
git commit -m "feat(pipeline): string builtins (upper/lower/trim/split/join/...)

Moves the 11 string helpers from text/template funcMap to sexpr builtins.
split/join operate on newline-delimited strings to match the rest of the
DSL (step output is newline-delimited)."
```

---

## Task 6: JSON builtins (pick, assoc)

**Files:**
- Modify: `internal/pipeline/builtins.go`
- Modify: `internal/pipeline/builtins_test.go`

- [ ] **Step 1: Failing tests**

Append:

```go
func TestJSONBuiltins(t *testing.T) {
	obj := `{"title":"bug fix","nested":{"k":"v"}}`
	got, err := callBuiltin("pick", []string{"title", obj}, NewScope())
	if err != nil || got != "bug fix" {
		t.Errorf("pick: got %q err=%v", got, err)
	}
	got2, _ := callBuiltin("pick", []string{"nested.k", obj}, NewScope())
	if got2 != "v" {
		t.Errorf("pick nested: got %q", got2)
	}
	got3, _ := callBuiltin("assoc", []string{"status", "done", `{"a":1}`}, NewScope())
	if got3 != `{"a":1,"status":"done"}` && got3 != `{"status":"done","a":1}` {
		t.Errorf("assoc: got %q", got3)
	}
}
```

- [ ] **Step 2: Run, expect fail**

Run: `go test ./internal/pipeline/ -run TestJSONBuiltins -v`
Expected: FAIL.

- [ ] **Step 3: Implement `tryJSONBuiltin`**

Replace the stub with:

```go
func tryJSONBuiltin(name string, args []string) (string, bool, error) {
	switch name {
	case "pick":
		if len(args) < 2 {
			return "", true, fmt.Errorf("pick: need (key json)")
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[1]), &obj); err != nil {
			return "", true, fmt.Errorf("pick: invalid JSON: %w", err)
		}
		parts := strings.Split(args[0], ".")
		var cur any = obj
		for _, p := range parts {
			m, ok := cur.(map[string]any)
			if !ok {
				return "", true, nil
			}
			cur = m[p]
		}
		switch v := cur.(type) {
		case string:
			return v, true, nil
		case nil:
			return "", true, nil
		default:
			b, _ := json.Marshal(v)
			return string(b), true, nil
		}
	case "assoc":
		if len(args) < 3 {
			return "", true, fmt.Errorf("assoc: need (key val json)")
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[2]), &obj); err != nil {
			return "", true, err
		}
		obj[args[0]] = args[1]
		b, err := json.Marshal(obj)
		if err != nil {
			return "", true, err
		}
		return string(b), true, nil
	}
	return "", false, nil
}
```

- [ ] **Step 4: Run**

Run: `go test ./internal/pipeline/ -run TestJSONBuiltins -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/builtins.go internal/pipeline/builtins_test.go
git commit -m "feat(pipeline): pick and assoc JSON builtins"
```

---

## Task 7: Form evaluator

**Files:**
- Modify: `internal/pipeline/quasi.go`
- Create: failing test in `internal/pipeline/quasi_test.go`

- [ ] **Step 1: Failing tests**

Append to `quasi_test.go`:

```go
func TestEvalForm_Builtin(t *testing.T) {
	scope := NewScope()
	scope.SetSteps(map[string]string{"diff": "a=b\nc=d"})
	got, err := evalForm("(step diff)", scope)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "a=b\nc=d" {
		t.Errorf("got %q", got)
	}
}

func TestEvalForm_Nested(t *testing.T) {
	scope := NewScope()
	scope.SetSteps(map[string]string{"items": "HELLO\nWORLD"})
	got, err := evalForm(`(lower (first (step items)))`, scope)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestEvalForm_BareRef(t *testing.T) {
	scope := NewScope()
	scope.SetLet("model", "qwen2.5:7b")
	got, err := evalForm(`(upper model)`, scope)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "QWEN2.5:7B" {
		t.Errorf("got %q", got)
	}
}

func TestEvalForm_KeywordAsString(t *testing.T) {
	// ~(pick :title json) — :title is a keyword, passed as "title" string arg
	scope := NewScope()
	scope.SetSteps(map[string]string{"data": `{"title":"hi"}`})
	got, err := evalForm(`(pick :title (step data))`, scope)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "hi" {
		t.Errorf("got %q", got)
	}
}

func TestEvalForm_UndefinedRef(t *testing.T) {
	_, err := evalForm(`(upper missing)`, NewScope())
	if err == nil {
		t.Fatal("expected undefined ref error")
	}
}
```

- [ ] **Step 2: Run**

Run: `go test ./internal/pipeline/ -run TestEvalForm -v`
Expected: FAIL (`evalForm` undefined).

- [ ] **Step 3: Implement evalForm**

Append to `quasi.go`:

```go
import (
	"github.com/8op-org/gl1tch/internal/sexpr"
)

// evalForm parses and evaluates a single sexpr form against the scope.
// Returns the stringified result.
func evalForm(src string, scope *Scope) (string, error) {
	nodes, err := sexpr.Parse([]byte(src))
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", src, err)
	}
	if len(nodes) != 1 {
		return "", fmt.Errorf("expected one form, got %d", len(nodes))
	}
	return evalNode(nodes[0], scope)
}

func evalNode(n *sexpr.Node, scope *Scope) (string, error) {
	if n.IsAtom() {
		return evalAtom(n, scope)
	}
	if !n.IsList() || len(n.Children) == 0 {
		return "", nil
	}
	head := n.Children[0]
	name := head.SymbolVal()
	if name == "" {
		name = head.StringVal()
	}
	if name == "" {
		return "", fmt.Errorf("line %d: invalid form head", n.Line)
	}
	args := make([]string, 0, len(n.Children)-1)
	for _, c := range n.Children[1:] {
		v, err := evalNode(c, scope)
		if err != nil {
			return "", err
		}
		args = append(args, v)
	}
	return callBuiltin(name, args, scope)
}

func evalAtom(n *sexpr.Node, scope *Scope) (string, error) {
	switch n.Atom.Type {
	case sexpr.TokenString:
		return n.Atom.Val, nil
	case sexpr.TokenKeyword:
		// :title -> "title"
		return n.KeywordVal(), nil
	case sexpr.TokenSymbol:
		sym := n.Atom.Val
		if strings.Contains(sym, ".") {
			parts := strings.Split(sym, ".")
			return scope.ResolvePath(parts[0], parts[1:])
		}
		return scope.Resolve(sym)
	}
	return "", fmt.Errorf("line %d: unsupported atom type", n.Line)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/pipeline/ -run TestEvalForm -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/quasi.go internal/pipeline/quasi_test.go
git commit -m "feat(pipeline): evalForm evaluates sexpr forms against a Scope

Parses a form source string via internal/sexpr, walks the AST, resolves
symbols against the scope, and dispatches builtin calls. Handles bare
symbols, dotted paths (param.x, env.X), string literals, and keywords
(passed as bare strings to builtins)."
```

---

## Task 8: renderQuasi — full string interpolation

**Files:**
- Modify: `internal/pipeline/quasi.go`
- Modify: `internal/pipeline/quasi_test.go`

- [ ] **Step 1: Failing tests**

Append:

```go
func TestRenderQuasi_Mixed(t *testing.T) {
	scope := NewScope()
	scope.SetParam("repo", "elastic/elasticsearch")
	scope.SetLet("model", "qwen2.5:7b")
	scope.SetSteps(map[string]string{"diff": "hello"})
	got, err := renderQuasi(`repo=~param.repo model=~model diff=~(step diff)`, scope)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := "repo=elastic/elasticsearch model=qwen2.5:7b diff=hello"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestRenderQuasi_UndefinedFails(t *testing.T) {
	scope := NewScope()
	_, err := renderQuasi("~nope", scope)
	if err == nil {
		t.Fatal("expected undefined ref error")
	}
}

func TestRenderQuasi_EscapedTilde(t *testing.T) {
	scope := NewScope()
	got, _ := renderQuasi(`cp file \~/dest`, scope)
	if got != "cp file ~/dest" {
		t.Errorf("got %q", got)
	}
}

func TestRenderQuasi_PlainPassthrough(t *testing.T) {
	got, err := renderQuasi("plain text", NewScope())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "plain text" {
		t.Errorf("got %q", got)
	}
}
```

- [ ] **Step 2: Run, expect fail**

Run: `go test ./internal/pipeline/ -run TestRenderQuasi -v`
Expected: FAIL.

- [ ] **Step 3: Implement renderQuasi**

Append to `quasi.go`:

```go
// renderQuasi interpolates ~name, ~param.x, and ~(form) references in src
// against the given scope. Literal strings with no "~" pass through
// unchanged. Undefined refs return an error.
func renderQuasi(src string, scope *Scope) (string, error) {
	if !strings.ContainsRune(src, '~') {
		return src, nil
	}
	parts, err := lexQuasi(src)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, p := range parts {
		switch p.Kind {
		case partLiteral:
			out.WriteString(p.Literal)
		case partRef:
			var v string
			if len(p.RefPath) == 0 {
				v, err = scope.Resolve(p.RefBase)
			} else {
				v, err = scope.ResolvePath(p.RefBase, p.RefPath)
			}
			if err != nil {
				return "", err
			}
			out.WriteString(v)
		case partForm:
			v, err := evalForm(p.Form, scope)
			if err != nil {
				return "", err
			}
			out.WriteString(v)
		}
	}
	return out.String(), nil
}
```

- [ ] **Step 4: Run**

Run: `go test ./internal/pipeline/ -run TestRenderQuasi -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/quasi.go internal/pipeline/quasi_test.go
git commit -m "feat(pipeline): renderQuasi wires lex+eval into a string renderer"
```

---

## Task 9: Swap `render()` to use renderQuasi

**Files:**
- Modify: `internal/pipeline/runner.go` (lines 752-903)
- Modify: `internal/pipeline/render_test.go`

- [ ] **Step 1: Rewrite render_test.go to new syntax**

Replace `render_test.go` contents:

```go
package pipeline

import "testing"

func TestRenderStringFunctions(t *testing.T) {
	scope := NewScope()
	scope.SetParam("repo", "elastic/elasticsearch")
	scope.SetParam("label", "  Bug Fix  ")

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{"split then join newline", `~(split "/" param.repo)`, "elastic\nelasticsearch"},
		{"first of split", `~(first (split "/" param.repo))`, "elastic"},
		{"last of split", `~(last (split "/" param.repo))`, "elasticsearch"},
		{"join", `~(join "-" (split "/" param.repo))`, "elastic-elasticsearch"},
		{"upper", `~(upper param.repo)`, "ELASTIC/ELASTICSEARCH"},
		{"lower literal", `~(lower "HELLO")`, "hello"},
		{"trim", `~(trim param.label)`, "Bug Fix"},
		{"trimPrefix", `~(trimPrefix "elastic/" param.repo)`, "elasticsearch"},
		{"trimSuffix", `~(trimSuffix "/elasticsearch" param.repo)`, "elastic"},
		{"replace", `~(replace "/" "-" param.repo)`, "elastic-elasticsearch"},
		{"truncate", `~(truncate 7 param.repo)`, "elastic"},
		{"truncate noop", `~(truncate 100 param.repo)`, "elastic/elasticsearch"},
		{"contains", `~(contains param.repo "elastic")`, "true"},
		{"hasPrefix", `~(hasPrefix param.repo "elastic")`, "true"},
		{"hasSuffix", `~(hasSuffix param.repo "search")`, "true"},
		{"nested chain", `~(upper (last (split "/" param.repo)))`, "ELASTICSEARCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := render(tt.tmpl, scope, nil)
			if err != nil {
				t.Fatalf("render(%q): %v", tt.tmpl, err)
			}
			if got != tt.want {
				t.Errorf("render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Replace the render() function body**

In `internal/pipeline/runner.go`, **delete lines 752-903** (the entire `render` function including funcMap, template.New, Execute) and replace with:

```go
// render interpolates sexpr-level unquote (~name, ~param.x, ~(form)) in tmpl
// against the given scope. steps is merged into scope before rendering.
// Signature kept for backwards compat with call sites; steps map is now
// threaded through Scope.
func render(tmpl string, scope *Scope, steps map[string]string) (string, error) {
	if steps != nil {
		scope.SetSteps(steps)
	}
	return renderQuasi(tmpl, scope)
}
```

Remove the `text/template` import from `runner.go` if it's no longer used. Keep `bytes`, `encoding/json`, etc. imports.

- [ ] **Step 3: Fix all call sites (19 of them)**

Each call site today looks like:

```go
rendered, err := render(step.Foo, data, rctx.stepsSnapshot())
```

Where `data` is `map[string]any`. These need to be refactored to pass a `*Scope` instead. Add a helper near the top of the `render` section:

```go
// scopeFromData translates the legacy data map (with "param", "input", keys)
// into a *Scope. Transitional helper; inline callers should build a Scope
// directly once the refactor settles.
func scopeFromData(data map[string]any) *Scope {
	s := NewScope()
	if p, ok := data["param"].(map[string]string); ok {
		for k, v := range p {
			s.SetParam(k, v)
		}
		if item, ok := p["item"]; ok {
			idx := 0
			if ix, ok := p["item_index"]; ok {
				fmt.Sscanf(ix, "%d", &idx)
			}
			s.SetItem(item, idx)
		}
	}
	if in, ok := data["input"].(string); ok {
		s.SetInput(in)
	}
	if defs, ok := data["def"].(map[string]string); ok {
		for k, v := range defs {
			s.SetDef(k, v)
		}
	}
	return s
}
```

Wrap each legacy call site by replacing `render(x, data, steps)` with `render(x, scopeFromData(data), steps)`.

Use search-and-replace on the 19 sites in `runner.go`:

```bash
# Manual fix at each render( call. Pattern:
# render(EXPR, data, stepsSnap)   -->   render(EXPR, scopeFromData(data), stepsSnap)
```

- [ ] **Step 4: Build and run existing tests**

Run:
```bash
go build ./...
go test ./internal/pipeline/ -run TestRender -v
```
Expected: PASS (tests rewritten to new syntax).

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/render_test.go
git commit -m "refactor(pipeline): render() now uses renderQuasi + Scope

Deletes the text/template funcMap/Parse/Execute plumbing and wires the
new quasi renderer in its place. Introduces scopeFromData as a transitional
helper so the 19 existing call sites keep their old data-map signature."
```

---

## Task 10: Delete unused text/template imports and dead code

**Files:**
- Modify: `internal/pipeline/runner.go`

- [ ] **Step 1: Remove unused imports**

Run:
```bash
goimports -w internal/pipeline/runner.go
go build ./...
```
Expected: no errors. `text/template` import should be gone.

- [ ] **Step 2: Grep for dead references**

Run:
```bash
rg -n 'text/template|template\.New|template\.FuncMap' internal/pipeline/
```
Expected: no matches.

- [ ] **Step 3: Commit**

```bash
git add internal/pipeline/
git commit -m "chore(pipeline): drop text/template import after render() rewrite"
```

---

## Task 11: Rewrite script

**Files:**
- Create: `scripts/rewrite-quasi/main.go`

- [ ] **Step 1: Implement the rewrite tool**

```go
// scripts/rewrite-quasi/main.go
// One-shot migration: rewrite {{...}} Go template syntax in .glitch
// workflow files to sexpr-level ~... syntax. Run once, then delete this
// file at merge.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var patterns = []struct {
	re  *regexp.Regexp
	out func(m []string) string
}{
	// {{.input}} -> ~input
	{regexp.MustCompile(`\{\{\s*\.input\s*\}\}`), func(m []string) string { return "~input" }},
	// {{.param.X}} -> ~param.X
	{regexp.MustCompile(`\{\{\s*\.param\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`),
		func(m []string) string { return "~param." + m[1] }},
	// {{.param.item}} already covered above; alias below for item_index
	{regexp.MustCompile(`\{\{\s*\.param\.item_index\s*\}\}`),
		func(m []string) string { return "~item_index" }},
	{regexp.MustCompile(`\{\{\s*\.param\.item\s*\}\}`),
		func(m []string) string { return "~item" }},
	// {{step "X"}} -> ~(step X)
	{regexp.MustCompile(`\{\{\s*step\s+"([^"]+)"\s*\}\}`),
		func(m []string) string { return "~(step " + m[1] + ")" }},
	// {{stepfile "X"}} -> ~(stepfile X)
	{regexp.MustCompile(`\{\{\s*stepfile\s+"([^"]+)"\s*\}\}`),
		func(m []string) string { return "~(stepfile " + m[1] + ")" }},
	// {{itemfile}} -> ~(itemfile)
	{regexp.MustCompile(`\{\{\s*itemfile\s*\}\}`),
		func(m []string) string { return "~(itemfile)" }},
	// {{branch "X"}} -> ~(branch X)
	{regexp.MustCompile(`\{\{\s*branch\s+"([^"]+)"\s*\}\}`),
		func(m []string) string { return "~(branch " + m[1] + ")" }},
}

func rewriteLine(line string) (string, bool) {
	out := line
	changed := false
	for _, p := range patterns {
		if p.re.MatchString(out) {
			out = p.re.ReplaceAllStringFunc(out, func(s string) string {
				m := p.re.FindStringSubmatch(s)
				return p.out(m)
			})
			changed = true
		}
	}
	return out, changed
}

func rewriteFile(path string) (int, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(src), "\n")
	changes := 0
	for i, ln := range lines {
		out, changed := rewriteLine(ln)
		if changed {
			lines[i] = out
			changes++
		}
	}
	if changes == 0 {
		return 0, nil
	}
	return changes, os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func findGlitchFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// skip hidden + build dirs
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".glitch") {
			out = append(out, p)
		}
		return nil
	})
	return out, err
}

func main() {
	root := flag.String("root", ".", "repo root to scan")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	files, err := findGlitchFiles(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	total := 0
	for _, f := range files {
		src, _ := os.ReadFile(f)
		lines := bytes.Split(src, []byte("\n"))
		changes := 0
		for _, ln := range lines {
			_, changed := rewriteLine(string(ln))
			if changed {
				changes++
			}
		}
		if changes == 0 {
			continue
		}
		if *dry {
			fmt.Printf("%s: %d line(s) would change\n", f, changes)
		} else {
			n, err := rewriteFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error rewriting %s: %v\n", f, err)
				continue
			}
			fmt.Printf("%s: rewrote %d line(s)\n", f, n)
			total += n
		}
	}
	fmt.Printf("total: %d lines\n", total)
}
```

- [ ] **Step 2: Dry-run against repo**

Run:
```bash
go run ./scripts/rewrite-quasi/ -root . -dry
```
Expected: list of `.glitch` files with non-zero change counts.

- [ ] **Step 3: Commit the script**

```bash
git add scripts/rewrite-quasi/main.go
git commit -m "tool(scripts): one-shot {{...}} -> ~... rewriter for .glitch

Scans for {{...}} patterns and rewrites them to sexpr-level ~... syntax.
Checked in for reproducibility of the migration PR; deleted at merge."
```

---

## Task 12: Rewrite all .glitch files

**Files:** every `.glitch` file in the repo (listed in the spec's File Structure section).

- [ ] **Step 1: Run the rewriter for real**

```bash
go run ./scripts/rewrite-quasi/ -root .
```
Expected: non-zero change count, no errors.

- [ ] **Step 2: Inspect diffs**

Run:
```bash
git diff --stat -- '*.glitch'
git diff -- '*.glitch' | head -100
```

Expected: clean `{{X}}` → `~...` rewrites. Flag any `{{if}}`, `{{range}}`, `{{with}}`, or pipe-chained templates that the script left behind. Grep for surviving templates:

```bash
rg -n '\{\{' --glob '*.glitch'
```

Expected: no matches. If any survive, fix by hand. Piped expressions like `{{X | f | g}}` need manual conversion to `~(g (f X))`.

- [ ] **Step 3: Build and run full test suite**

```bash
go build ./...
go test ./... 2>&1 | tail -40
```
Expected: tests may still fail in test files that have inline `{{...}}` fixtures not covered by the `.glitch` scan. Fix those in Task 13.

- [ ] **Step 4: Commit workflow rewrites**

```bash
git add '*.glitch' .glitch/workflows/ examples/ test-workspace/workflows/ internal/pipeline/testdata/
git commit -m "migrate(workflows): rewrite all .glitch files to sexpr-unquote syntax

One-shot mechanical conversion via scripts/rewrite-quasi. Every {{...}}
reference becomes ~... . No behavior change expected — new renderer
produces identical output for identical inputs."
```

---

## Task 13: Fix test fixtures using `{{...}}`

**Files:** any `*_test.go` file under `internal/pipeline/` that inlines `{{...}}`.

- [ ] **Step 1: Find them**

Run:
```bash
rg -n '\{\{' internal/pipeline/ --type go
```

- [ ] **Step 2: Convert each by hand**

For each match:
- `{{.input}}` → `~input`
- `{{.param.X}}` → `~param.X`
- `{{step "X"}}` → `~(step X)`
- Pipes → nested calls

- [ ] **Step 3: Run all pipeline tests**

```bash
go test ./internal/pipeline/... -v 2>&1 | tail -80
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/pipeline/
git commit -m "test(pipeline): update fixtures to sexpr-unquote syntax"
```

---

## Task 14: Full build + test

- [ ] **Step 1: Build**

```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 2: Full test run**

```bash
go test ./... 2>&1 | tee /tmp/glitch-test.log
tail -20 /tmp/glitch-test.log
```
Expected: all packages PASS.

- [ ] **Step 3: Vet + staticcheck**

```bash
go vet ./...
```
Expected: no warnings.

- [ ] **Step 4: If any failures, fix and commit per-fix**

- [ ] **Step 5: Delete rewrite script**

```bash
rm -r scripts/rewrite-quasi/
git add -A scripts/
git commit -m "chore(scripts): remove rewrite-quasi after migration"
```

---

## Task 15: Smoke pack baseline

- [ ] **Step 1: Build and install**

```bash
go build -o /tmp/glitch-new .
```

- [ ] **Step 2: Run smoke pack**

```bash
/tmp/glitch-new smoke pack 2>&1 | tail -30
```
Expected: 24/24 baseline against ensemble, kibana, oblt-cli, observability-robots.

- [ ] **Step 3: If regression, bisect and fix**

- [ ] **Step 4: Commit any final fixes**

---

## Self-Review

- **Spec coverage:** each section covered?
  - Forms (`~name`, `~param.x`, `~(form)`, `\~`): Task 2, Task 3, Task 8. ✓
  - Scope rules (let > def > specials): Task 1. ✓
  - Error on undefined: Task 1 (UndefinedRefError), Task 8 (propagated). ✓
  - Builtins migration: Task 4-6. ✓
  - render() rewrite: Task 9-10. ✓
  - Migration script: Task 11. ✓
  - All workflows rewritten: Task 12-13. ✓
  - Smoke pack: Task 15. ✓

- **Placeholder scan:** no TBD / TODO / "add error handling" / abstract "similar to" references.

- **Type consistency:** `Scope`, `UndefinedRefError`, `quasiPart`, `callBuiltin`, `renderQuasi`, `evalForm`, `scopeFromData` are all referenced consistently across tasks.

- **Signature drift:** `render()` keeps `(string, map[string]any, map[string]string) → (string, error)` shape via `scopeFromData`. Call-sites need a one-line change per site. Task 9 includes the pattern.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-16-sexpr-interpolation.md`. Two execution options:

1. **Subagent-Driven (recommended for this plan)** — fresh subagent per task, review between tasks, fast iteration. Best fit for user's "1 subagent per task" preference.

2. **Inline Execution** — execute tasks in this session, batch checkpoints.

User selected: **Subagent-Driven** (per "1 subagent per task, review at end" directive).

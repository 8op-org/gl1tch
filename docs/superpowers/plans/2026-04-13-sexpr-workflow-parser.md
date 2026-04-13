# S-Expression Workflow Parser

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Parse `.glitch` s-expression workflow files into the existing `pipeline.Workflow` struct, as an alternative to YAML.

**Architecture:** A `sexpr` package with three layers: tokenizer (bytes to tokens), parser (tokens to AST), and converter (AST to `pipeline.Workflow`). The converter produces the same structs the YAML loader does, so the runner and router need zero changes.

**Tech Stack:** Pure Go, no dependencies. Standard `text/scanner` not used (we need custom multiline string handling). TDD throughout.

---

## Format Definition

```lisp
;; a gl1tch workflow
(workflow "analyze-issues"
  :description "Fetch and analyze GitHub issues"

  (step "fetch-issues"
    (run "gh issue list --json title,number"))

  (step "analyze"
    (llm
      :provider "ollama"
      :model "qwen2.5:7b"
      :prompt ```
        Summarize these issues:
        {{step "fetch-issues"}}
        ```))

  (step "save-report"
    (save "results/{{.param.repo}}/issues.md"
      :from "analyze")))
```

**Syntax rules:**
- `;` — line comment (to end of line)
- `#_` — discard next form
- `"..."` — string (standard Go escapes: `\n`, `\t`, `\\`, `\"`)
- `` ``` `` — multiline string delimiter (triple backtick), content between opening and closing ``` is literal, leading indentation stripped
- `:keyword` — keyword (self-evaluating identifier)
- `(...)` — list
- Whitespace and commas are separators (commas are whitespace, like EDN)

**Workflow mapping to existing structs:**

| S-expr | `pipeline.Workflow` / `pipeline.Step` |
|--------|---------------------------------------|
| `(workflow "name" ...)` | `Workflow.Name` |
| `:description "..."` | `Workflow.Description` |
| `(step "id" ...)` | `Step.ID` |
| `(run "cmd")` | `Step.Run` |
| `(llm :prompt "..." ...)` | `Step.LLM` with `LLMStep.Prompt` |
| `:provider "ollama"` | `LLMStep.Provider` |
| `:model "qwen2.5:7b"` | `LLMStep.Model` |
| `(save "path")` | `Step.Save` |
| `:from "step-id"` | `Step.SaveStep` |

---

## File Structure

```
internal/sexpr/
  token.go       — Token type, token constants
  lexer.go       — Lexer: []byte → []Token
  lexer_test.go  — Lexer tests
  ast.go         — AST node types (Atom, List)
  parser.go      — Parser: []Token → []Node
  parser_test.go — Parser tests
  workflow.go    — Converter: []Node → *pipeline.Workflow
  workflow_test.go — Converter tests (integration-level)
```

---

### Task 1: Token types and AST nodes

**Files:**
- Create: `internal/sexpr/token.go`
- Create: `internal/sexpr/ast.go`

- [ ] **Step 1: Create token types**

```go
// internal/sexpr/token.go
package sexpr

type TokenType int

const (
	TokenLParen   TokenType = iota // (
	TokenRParen                    // )
	TokenString                    // "..." or ```...```
	TokenKeyword                   // :name
	TokenDiscard                   // #_
)

type Token struct {
	Type TokenType
	Val  string // raw value (strings unescaped, keywords include leading :)
	Pos  int    // byte offset in source
	Line int    // 1-based line number
}
```

- [ ] **Step 2: Create AST node types**

```go
// internal/sexpr/ast.go
package sexpr

// Node is an element in the AST — either an atom or a list.
type Node struct {
	// Exactly one of Atom or Children is set.
	Atom     *Token  // non-nil for leaf nodes (string, keyword)
	Children []*Node // non-nil for list nodes (...)
	Line     int     // source line for error messages
}

// IsAtom returns true if this node is a leaf.
func (n *Node) IsAtom() bool { return n.Atom != nil }

// IsList returns true if this node has children.
func (n *Node) IsList() bool { return n.Children != nil }

// StringVal returns the string value of a string atom, or empty string.
func (n *Node) StringVal() string {
	if n.Atom != nil && n.Atom.Type == TokenString {
		return n.Atom.Val
	}
	return ""
}

// KeywordVal returns the keyword name (without :) or empty string.
func (n *Node) KeywordVal() string {
	if n.Atom != nil && n.Atom.Type == TokenKeyword {
		return n.Atom.Val[1:] // strip leading ':'
	}
	return ""
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./internal/sexpr/`
Expected: clean compile, no errors

- [ ] **Step 4: Commit**

```bash
git add internal/sexpr/token.go internal/sexpr/ast.go
git commit -m "feat(sexpr): add token and AST node types"
```

---

### Task 2: Lexer

**Files:**
- Create: `internal/sexpr/lexer.go`
- Create: `internal/sexpr/lexer_test.go`

- [ ] **Step 1: Write failing tests for the lexer**

```go
// internal/sexpr/lexer_test.go
package sexpr

import "testing"

func TestLex_Parens(t *testing.T) {
	tokens, err := Lex([]byte("()"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenLParen {
		t.Fatalf("expected LParen, got %v", tokens[0].Type)
	}
	if tokens[1].Type != TokenRParen {
		t.Fatalf("expected RParen, got %v", tokens[1].Type)
	}
}

func TestLex_String(t *testing.T) {
	tokens, err := Lex([]byte(`"hello world"`))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenString {
		t.Fatalf("expected String, got %v", tokens[0].Type)
	}
	if tokens[0].Val != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", tokens[0].Val)
	}
}

func TestLex_StringEscapes(t *testing.T) {
	tokens, err := Lex([]byte(`"line1\nline2\t\"quoted\""`))
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Val != "line1\nline2\t\"quoted\"" {
		t.Fatalf("got %q", tokens[0].Val)
	}
}

func TestLex_MultilineString(t *testing.T) {
	src := "```\n  hello\n  world\n  ```"
	tokens, err := Lex([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenString {
		t.Fatalf("expected String, got %v", tokens[0].Type)
	}
	// Leading indentation (2 spaces) should be stripped based on closing delimiter indent
	if tokens[0].Val != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", tokens[0].Val)
	}
}

func TestLex_Keyword(t *testing.T) {
	tokens, err := Lex([]byte(":description"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenKeyword {
		t.Fatalf("expected Keyword, got %v", tokens[0].Type)
	}
	if tokens[0].Val != ":description" {
		t.Fatalf("expected %q, got %q", ":description", tokens[0].Val)
	}
}

func TestLex_Discard(t *testing.T) {
	tokens, err := Lex([]byte("#_"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenDiscard {
		t.Fatalf("expected Discard, got %v", tokens[0].Type)
	}
}

func TestLex_LineComment(t *testing.T) {
	tokens, err := Lex([]byte("; this is a comment\n\"hello\""))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token (comment skipped), got %d", len(tokens))
	}
	if tokens[0].Val != "hello" {
		t.Fatalf("expected %q, got %q", "hello", tokens[0].Val)
	}
}

func TestLex_CommasAreWhitespace(t *testing.T) {
	tokens, err := Lex([]byte(`"a","b","c"`))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
}

func TestLex_FullWorkflow(t *testing.T) {
	src := `
;; example workflow
(workflow "test"
  :description "a test"
  (step "s1"
    (run "echo hello")))
`
	tokens, err := Lex([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// ( workflow "test" :description "a test" ( step "s1" ( run "echo hello" ) ) )
	// 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 = expect specific count
	expected := []TokenType{
		TokenLParen, TokenString, TokenString,
		TokenKeyword, TokenString,
		TokenLParen, TokenString, TokenString,
		TokenLParen, TokenString, TokenString,
		TokenRParen, TokenRParen, TokenRParen,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token %d: expected %v, got %v (val=%q)", i, expected[i], tok.Type, tok.Val)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -v`
Expected: compilation error — `Lex` not defined

- [ ] **Step 3: Implement the lexer**

```go
// internal/sexpr/lexer.go
package sexpr

import (
	"fmt"
	"strings"
)

// Lex tokenizes s-expression source into a slice of tokens.
func Lex(src []byte) ([]Token, error) {
	l := &lexer{src: src, line: 1}
	return l.lexAll()
}

type lexer struct {
	src  []byte
	pos  int
	line int
}

func (l *lexer) lexAll() ([]Token, error) {
	var tokens []Token
	for l.pos < len(l.src) {
		ch := l.src[l.pos]

		switch {
		case ch == '\n':
			l.line++
			l.pos++
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == ',':
			l.pos++
		case ch == ';':
			l.skipLineComment()
		case ch == '(':
			tokens = append(tokens, Token{Type: TokenLParen, Pos: l.pos, Line: l.line})
			l.pos++
		case ch == ')':
			tokens = append(tokens, Token{Type: TokenRParen, Pos: l.pos, Line: l.line})
			l.pos++
		case ch == '"':
			tok, err := l.lexString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
		case ch == '`' && l.peekTripleBacktick():
			tok, err := l.lexMultilineString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
		case ch == ':':
			tokens = append(tokens, l.lexKeyword())
		case ch == '#' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '_':
			tokens = append(tokens, Token{Type: TokenDiscard, Pos: l.pos, Line: l.line})
			l.pos += 2
		default:
			// Bare word — read as string token (for symbols like "workflow", "step", etc.)
			tokens = append(tokens, l.lexBareWord())
		}
	}
	return tokens, nil
}

func (l *lexer) skipLineComment() {
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
}

func (l *lexer) lexString() (Token, error) {
	start := l.pos
	line := l.line
	l.pos++ // skip opening "
	var b strings.Builder
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		if ch == '\\' {
			l.pos++
			if l.pos >= len(l.src) {
				return Token{}, fmt.Errorf("line %d: unexpected end of string escape", line)
			}
			esc := l.src[l.pos]
			switch esc {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				return Token{}, fmt.Errorf("line %d: unknown escape \\%c", line, esc)
			}
			l.pos++
			continue
		}
		if ch == '"' {
			l.pos++ // skip closing "
			return Token{Type: TokenString, Val: b.String(), Pos: start, Line: line}, nil
		}
		if ch == '\n' {
			l.line++
		}
		b.WriteByte(ch)
		l.pos++
	}
	return Token{}, fmt.Errorf("line %d: unterminated string", line)
}

func (l *lexer) peekTripleBacktick() bool {
	return l.pos+2 < len(l.src) && l.src[l.pos+1] == '`' && l.src[l.pos+2] == '`'
}

func (l *lexer) lexMultilineString() (Token, error) {
	start := l.pos
	line := l.line
	l.pos += 3 // skip opening ```

	// Skip optional newline after opening ```
	if l.pos < len(l.src) && l.src[l.pos] == '\n' {
		l.line++
		l.pos++
	}

	// Read until closing ```
	var raw strings.Builder
	for {
		if l.pos+2 < len(l.src) && l.src[l.pos] == '`' && l.src[l.pos+1] == '`' && l.src[l.pos+2] == '`' {
			break
		}
		if l.pos >= len(l.src) {
			return Token{}, fmt.Errorf("line %d: unterminated multiline string", line)
		}
		if l.src[l.pos] == '\n' {
			l.line++
		}
		raw.WriteByte(l.src[l.pos])
		l.pos++
	}

	// Determine indent of closing ``` to strip from all lines
	content := raw.String()
	// Find indent: look backwards from current pos to find leading whitespace on closing line
	closingIndent := 0
	for i := l.pos - 1; i >= 0 && l.src[i] != '\n'; i-- {
		if l.src[i] == ' ' || l.src[i] == '\t' {
			closingIndent++
		} else {
			closingIndent = 0
		}
	}

	l.pos += 3 // skip closing ```

	// Strip trailing newline from content (before the closing delimiter line)
	content = strings.TrimRight(content, "\n")

	// Strip leading indent from each line
	if closingIndent > 0 {
		lines := strings.Split(content, "\n")
		for i, ln := range lines {
			if len(ln) >= closingIndent {
				lines[i] = ln[closingIndent:]
			} else {
				lines[i] = strings.TrimLeft(ln, " \t")
			}
		}
		content = strings.Join(lines, "\n")
	}

	return Token{Type: TokenString, Val: content, Pos: start, Line: line}, nil
}

func (l *lexer) lexKeyword() Token {
	start := l.pos
	line := l.line
	l.pos++ // skip :
	for l.pos < len(l.src) && isWordChar(l.src[l.pos]) {
		l.pos++
	}
	return Token{Type: TokenKeyword, Val: string(l.src[start:l.pos]), Pos: start, Line: line}
}

func (l *lexer) lexBareWord() Token {
	start := l.pos
	line := l.line
	for l.pos < len(l.src) && isWordChar(l.src[l.pos]) {
		l.pos++
	}
	return Token{Type: TokenString, Val: string(l.src[start:l.pos]), Pos: start, Line: line}
}

func isWordChar(ch byte) bool {
	return ch != '(' && ch != ')' && ch != '"' && ch != ';' &&
		ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' && ch != ','
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/sexpr/lexer.go internal/sexpr/lexer_test.go
git commit -m "feat(sexpr): implement lexer with multiline strings and comments"
```

---

### Task 3: Parser

**Files:**
- Create: `internal/sexpr/parser.go`
- Create: `internal/sexpr/parser_test.go`

- [ ] **Step 1: Write failing tests for the parser**

```go
// internal/sexpr/parser_test.go
package sexpr

import "testing"

func TestParse_EmptyList(t *testing.T) {
	nodes, err := Parse([]byte("()"))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if !nodes[0].IsList() {
		t.Fatal("expected list node")
	}
	if len(nodes[0].Children) != 0 {
		t.Fatalf("expected empty list, got %d children", len(nodes[0].Children))
	}
}

func TestParse_NestedList(t *testing.T) {
	nodes, err := Parse([]byte(`(workflow "test" (step "s1" (run "echo hi")))`))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 top-level node, got %d", len(nodes))
	}
	wf := nodes[0]
	if !wf.IsList() || len(wf.Children) != 3 {
		t.Fatalf("expected list with 3 children, got %d", len(wf.Children))
	}
	// wf.Children[0] = "workflow", [1] = "test", [2] = (step ...)
	if wf.Children[0].StringVal() != "workflow" {
		t.Fatalf("expected 'workflow', got %q", wf.Children[0].StringVal())
	}
	step := wf.Children[2]
	if !step.IsList() || len(step.Children) != 3 {
		t.Fatalf("expected step list with 3 children, got %d", len(step.Children))
	}
}

func TestParse_Keywords(t *testing.T) {
	nodes, err := Parse([]byte(`(:name "test")`))
	if err != nil {
		t.Fatal(err)
	}
	list := nodes[0]
	if list.Children[0].KeywordVal() != "name" {
		t.Fatalf("expected keyword 'name', got %q", list.Children[0].KeywordVal())
	}
}

func TestParse_Discard(t *testing.T) {
	nodes, err := Parse([]byte(`("keep" #_"discard" "also-keep")`))
	if err != nil {
		t.Fatal(err)
	}
	list := nodes[0]
	if len(list.Children) != 2 {
		t.Fatalf("expected 2 children (discard removed one), got %d", len(list.Children))
	}
	if list.Children[0].StringVal() != "keep" {
		t.Fatalf("expected 'keep', got %q", list.Children[0].StringVal())
	}
	if list.Children[1].StringVal() != "also-keep" {
		t.Fatalf("expected 'also-keep', got %q", list.Children[1].StringVal())
	}
}

func TestParse_DiscardList(t *testing.T) {
	nodes, err := Parse([]byte(`("a" #_(skip "this" "entire" "thing") "b")`))
	if err != nil {
		t.Fatal(err)
	}
	list := nodes[0]
	if len(list.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(list.Children))
	}
}

func TestParse_UnmatchedRParen(t *testing.T) {
	_, err := Parse([]byte(")"))
	if err == nil {
		t.Fatal("expected error for unmatched )")
	}
}

func TestParse_UnmatchedLParen(t *testing.T) {
	_, err := Parse([]byte("("))
	if err == nil {
		t.Fatal("expected error for unmatched (")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -v -run TestParse`
Expected: compilation error — `Parse` not defined

- [ ] **Step 3: Implement the parser**

```go
// internal/sexpr/parser.go
package sexpr

import "fmt"

// Parse tokenizes and parses s-expression source into an AST.
func Parse(src []byte) ([]*Node, error) {
	tokens, err := Lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	return p.parseAll()
}

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) parseAll() ([]*Node, error) {
	var nodes []*Node
	for p.pos < len(p.tokens) {
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (p *parser) parseNode() (*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}
	tok := p.tokens[p.pos]

	switch tok.Type {
	case TokenDiscard:
		p.pos++ // skip #_
		// Parse and discard the next form
		_, err := p.parseNode()
		if err != nil {
			return nil, fmt.Errorf("line %d: discard (#_) must be followed by a form: %w", tok.Line, err)
		}
		// Return nil — caller skips nil nodes
		return nil, nil

	case TokenLParen:
		return p.parseList()

	case TokenRParen:
		return nil, fmt.Errorf("line %d: unexpected )", tok.Line)

	case TokenString, TokenKeyword:
		p.pos++
		return &Node{Atom: &tok, Line: tok.Line}, nil

	default:
		return nil, fmt.Errorf("line %d: unexpected token %v", tok.Line, tok.Type)
	}
}

func (p *parser) parseList() (*Node, error) {
	open := p.tokens[p.pos]
	p.pos++ // skip (

	var children []*Node
	for {
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("line %d: unterminated list", open.Line)
		}
		if p.tokens[p.pos].Type == TokenRParen {
			p.pos++ // skip )
			return &Node{Children: children, Line: open.Line}, nil
		}
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil { // nil means discarded
			children = append(children, child)
		}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/sexpr/parser.go internal/sexpr/parser_test.go
git commit -m "feat(sexpr): implement parser with discard support"
```

---

### Task 4: Workflow converter

**Files:**
- Create: `internal/sexpr/workflow.go`
- Create: `internal/sexpr/workflow_test.go`

- [ ] **Step 1: Write failing tests for the converter**

```go
// internal/sexpr/workflow_test.go
package sexpr

import (
	"testing"
)

func TestWorkflow_Basic(t *testing.T) {
	src := []byte(`
(workflow "my-pipeline"
  :description "a test pipeline"
  (step "fetch"
    (run "echo hello"))
  (step "analyze"
    (llm :prompt "summarize: {{step \"fetch\"}}")))
`)
	w, err := ParseWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "my-pipeline" {
		t.Fatalf("expected name %q, got %q", "my-pipeline", w.Name)
	}
	if w.Description != "a test pipeline" {
		t.Fatalf("expected description %q, got %q", "a test pipeline", w.Description)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}

	s0 := w.Steps[0]
	if s0.ID != "fetch" {
		t.Fatalf("step 0: expected id %q, got %q", "fetch", s0.ID)
	}
	if s0.Run != "echo hello" {
		t.Fatalf("step 0: expected run %q, got %q", "echo hello", s0.Run)
	}

	s1 := w.Steps[1]
	if s1.ID != "analyze" {
		t.Fatalf("step 1: expected id %q, got %q", "analyze", s1.ID)
	}
	if s1.LLM == nil {
		t.Fatal("step 1: expected LLM step")
	}
	if s1.LLM.Prompt != `summarize: {{step "fetch"}}` {
		t.Fatalf("step 1: expected prompt with template, got %q", s1.LLM.Prompt)
	}
}

func TestWorkflow_LLMWithProviderAndModel(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (llm
      :provider "claude"
      :model "opus"
      :prompt "hello")))
`)
	w, err := ParseWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.LLM.Provider != "claude" {
		t.Fatalf("expected provider %q, got %q", "claude", s.LLM.Provider)
	}
	if s.LLM.Model != "opus" {
		t.Fatalf("expected model %q, got %q", "opus", s.LLM.Model)
	}
}

func TestWorkflow_Save(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "gen"
    (llm :prompt "write something"))
  (step "write"
    (save "output/{{.param.repo}}/result.md" :from "gen")))
`)
	w, err := ParseWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.Save != "output/{{.param.repo}}/result.md" {
		t.Fatalf("expected save path, got %q", s.Save)
	}
	if s.SaveStep != "gen" {
		t.Fatalf("expected save_step %q, got %q", "gen", s.SaveStep)
	}
}

func TestWorkflow_MultilinePrompt(t *testing.T) {
	src := "(workflow \"test\"\n  (step \"s1\"\n    (llm :prompt ```\n      hello\n      world\n      ```)))"
	w, err := ParseWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM.Prompt != "hello\nworld" {
		t.Fatalf("expected dedented multiline, got %q", w.Steps[0].LLM.Prompt)
	}
}

func TestWorkflow_DiscardedStep(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "keep"
    (run "echo yes"))
  #_(step "skip"
    (run "echo no"))
  (step "also-keep"
    (run "echo yes2")))
`)
	w, err := ParseWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps (one discarded), got %d", len(w.Steps))
	}
	if w.Steps[0].ID != "keep" {
		t.Fatalf("expected first step 'keep', got %q", w.Steps[0].ID)
	}
	if w.Steps[1].ID != "also-keep" {
		t.Fatalf("expected second step 'also-keep', got %q", w.Steps[1].ID)
	}
}

func TestWorkflow_NotAWorkflow(t *testing.T) {
	_, err := ParseWorkflow([]byte(`(notworkflow "test")`))
	if err == nil {
		t.Fatal("expected error for non-workflow form")
	}
}

func TestWorkflow_MissingName(t *testing.T) {
	_, err := ParseWorkflow([]byte(`(workflow)`))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -v -run TestWorkflow`
Expected: compilation error — `ParseWorkflow` not defined

- [ ] **Step 3: Implement the converter**

```go
// internal/sexpr/workflow.go
package sexpr

import (
	"fmt"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

// ParseWorkflow parses s-expression source into a pipeline.Workflow.
func ParseWorkflow(src []byte) (*pipeline.Workflow, error) {
	nodes, err := Parse(src)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.IsList() && len(n.Children) > 0 && n.Children[0].StringVal() == "workflow" {
			return convertWorkflow(n)
		}
	}
	return nil, fmt.Errorf("no (workflow ...) form found")
}

func convertWorkflow(n *Node) (*pipeline.Workflow, error) {
	children := n.Children[1:] // skip "workflow" symbol
	if len(children) == 0 {
		return nil, fmt.Errorf("line %d: workflow missing name", n.Line)
	}

	w := &pipeline.Workflow{}

	// First child must be the name
	w.Name = children[0].StringVal()
	if w.Name == "" {
		return nil, fmt.Errorf("line %d: workflow name must be a string", children[0].Line)
	}
	children = children[1:]

	// Process remaining children: keywords for metadata, lists for steps
	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "description":
				w.Description = val.StringVal()
			default:
				return nil, fmt.Errorf("line %d: unknown workflow keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		if child.IsList() && len(child.Children) > 0 && child.Children[0].StringVal() == "step" {
			step, err := convertStep(child)
			if err != nil {
				return nil, err
			}
			w.Steps = append(w.Steps, step)
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in workflow", child.Line)
	}
	return w, nil
}

func convertStep(n *Node) (pipeline.Step, error) {
	children := n.Children[1:] // skip "step"
	if len(children) == 0 {
		return pipeline.Step{}, fmt.Errorf("line %d: step missing id", n.Line)
	}

	s := pipeline.Step{}
	s.ID = children[0].StringVal()
	if s.ID == "" {
		return s, fmt.Errorf("line %d: step id must be a string", children[0].Line)
	}
	children = children[1:]

	for _, child := range children {
		if !child.IsList() || len(child.Children) == 0 {
			return s, fmt.Errorf("line %d: step body must be (run ...), (llm ...), or (save ...)", child.Line)
		}
		head := child.Children[0].StringVal()
		switch head {
		case "run":
			if len(child.Children) < 2 {
				return s, fmt.Errorf("line %d: (run) missing command", child.Line)
			}
			s.Run = child.Children[1].StringVal()
		case "llm":
			llm, err := convertLLM(child)
			if err != nil {
				return s, err
			}
			s.LLM = llm
		case "save":
			if len(child.Children) < 2 {
				return s, fmt.Errorf("line %d: (save) missing path", child.Line)
			}
			s.Save = child.Children[1].StringVal()
			// Check for :from keyword
			rest := child.Children[2:]
			for j := 0; j < len(rest); j++ {
				if rest[j].IsAtom() && rest[j].Atom.Type == TokenKeyword && rest[j].KeywordVal() == "from" {
					j++
					if j < len(rest) {
						s.SaveStep = rest[j].StringVal()
					}
				}
			}
		default:
			return s, fmt.Errorf("line %d: unknown step type %q", child.Line, head)
		}
	}
	return s, nil
}

func convertLLM(n *Node) (*pipeline.LLMStep, error) {
	children := n.Children[1:] // skip "llm"
	llm := &pipeline.LLMStep{}

	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "prompt":
				llm.Prompt = val.StringVal()
			case "provider":
				llm.Provider = val.StringVal()
			case "model":
				llm.Model = val.StringVal()
			default:
				return nil, fmt.Errorf("line %d: unknown llm keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in llm", child.Line)
	}
	if llm.Prompt == "" {
		return nil, fmt.Errorf("line %d: llm missing :prompt", n.Line)
	}
	return llm, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/sexpr/workflow.go internal/sexpr/workflow_test.go
git commit -m "feat(sexpr): workflow converter — sexpr to pipeline.Workflow"
```

---

### Task 5: Wire into LoadDir / LoadFile

**Files:**
- Modify: `internal/pipeline/types.go:34-89`

- [ ] **Step 1: Write failing test for .glitch file loading**

Add to `internal/pipeline/runner_test.go`:

```go
func TestLoadBytes_Sexpr(t *testing.T) {
	src := []byte(`
(workflow "test-sexpr"
  :description "loaded from sexpr"
  (step "s1"
    (run "echo hello")))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "test-sexpr" {
		t.Fatalf("expected name %q, got %q", "test-sexpr", w.Name)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	if w.Steps[0].Run != "echo hello" {
		t.Fatalf("expected run %q, got %q", "echo hello", w.Steps[0].Run)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestLoadBytes_Sexpr`
Expected: FAIL — YAML parser will choke on s-expression syntax

- [ ] **Step 3: Update LoadBytes and LoadDir to dispatch on file extension**

Edit `internal/pipeline/types.go` — update the imports, `LoadFile`, `LoadBytes`, and `LoadDir`:

```go
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
	"gopkg.in/yaml.v3"
)

// ... (Workflow, Step, LLMStep structs unchanged) ...

// LoadFile reads a single workflow file (YAML or s-expression).
func LoadFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBytes(data, filepath.Base(path))
}

// LoadBytes parses a workflow from raw bytes, dispatching on filename extension.
func LoadBytes(data []byte, filename string) (*Workflow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".glitch":
		return sexpr.ParseWorkflow(data)
	default:
		var w Workflow
		if err := yaml.Unmarshal(data, &w); err != nil {
			return nil, fmt.Errorf("parse %s: %w", filename, err)
		}
		if w.Name == "" {
			w.Name = filename
		}
		return &w, nil
	}
}

// LoadDir reads all workflow files from a directory, keyed by workflow name.
func LoadDir(dir string) (map[string]*Workflow, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	workflows := make(map[string]*Workflow)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".yml" && ext != ".glitch" {
			continue
		}
		w, err := LoadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", e.Name(), err)
			continue
		}
		workflows[w.Name] = w
	}
	return workflows, nil
}
```

- [ ] **Step 4: Run all tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ ./internal/sexpr/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/runner_test.go
git commit -m "feat: wire sexpr parser into workflow loader — .glitch extension"
```

---

### Task 6: Example workflow file

**Files:**
- Create: `examples/hello.glitch`

- [ ] **Step 1: Create an example workflow**

```lisp
;; hello.glitch — example gl1tch s-expression workflow
;;
;; Run with: glitch pipeline run examples/hello.glitch

(workflow "hello-sexpr"
  :description "Demo s-expression workflow format"

  (step "gather"
    (run "echo 'hello from a .glitch workflow'"))

  (step "respond"
    (llm
      :prompt ```
        You received this message from a shell command:
        {{step "gather"}}

        Respond with a short, enthusiastic acknowledgment.
        ```))

  ;; disable the save step for now
  #_(step "save-it"
    (save "results/hello.md" :from "respond")))
```

- [ ] **Step 2: Verify it parses**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/sexpr/ -run TestWorkflow -v`
Expected: existing tests still pass (the example is for human reference, not wired into tests)

- [ ] **Step 3: Commit**

```bash
git add examples/hello.glitch
git commit -m "docs: add example .glitch workflow file"
```

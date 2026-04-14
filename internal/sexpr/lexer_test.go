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

func TestLex_Symbol(t *testing.T) {
	tokens, err := Lex([]byte("workflow"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 || tokens[0].Type != TokenSymbol {
		t.Fatalf("expected 1 Symbol token, got %d tokens, type=%v", len(tokens), tokens[0].Type)
	}
	if tokens[0].Val != "workflow" {
		t.Fatalf("expected %q, got %q", "workflow", tokens[0].Val)
	}
}

func TestLex_MapBraces(t *testing.T) {
	tokens, err := Lex([]byte(`{:a "b"}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenLBrace {
		t.Fatalf("expected LBrace, got %v", tokens[0].Type)
	}
	if tokens[3].Type != TokenRBrace {
		t.Fatalf("expected RBrace, got %v", tokens[3].Type)
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
	expected := []TokenType{
		TokenLParen, TokenSymbol, TokenString,
		TokenKeyword, TokenString,
		TokenLParen, TokenSymbol, TokenString,
		TokenLParen, TokenSymbol, TokenString,
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

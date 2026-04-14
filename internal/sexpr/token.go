// internal/sexpr/token.go
package sexpr

type TokenType int

const (
	TokenLParen  TokenType = iota // (
	TokenRParen                   // )
	TokenString                   // "..." or ```...```
	TokenKeyword                  // :name
	TokenDiscard                  // #_
	TokenSymbol                   // bare word (workflow, step, run, ...)
	TokenLBrace                   // {
	TokenRBrace                   // }
)

type Token struct {
	Type TokenType
	Val  string // raw value (strings unescaped, keywords include leading :)
	Pos  int    // byte offset in source
	Line int    // 1-based line number
}

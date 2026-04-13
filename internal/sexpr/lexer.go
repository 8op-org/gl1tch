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

	// Determine indent of closing ``` by scanning backwards from current pos
	// to find the number of whitespace chars on the closing delimiter line.
	closingIndent := 0
	scanPos := l.pos - 1
	for scanPos >= 0 && l.src[scanPos] != '\n' {
		scanPos--
	}
	// scanPos is now at the \n before closing line (or -1 if no newline)
	if scanPos >= 0 {
		for i := scanPos + 1; i < l.pos; i++ {
			if l.src[i] == ' ' || l.src[i] == '\t' {
				closingIndent++
			} else {
				break
			}
		}
	}

	l.pos += 3 // skip closing ```

	content := raw.String()

	// Strip the trailing line (newline + indent whitespace before closing delimiter)
	content = strings.TrimRight(content, " \t\n")

	// Strip leading indent from each line based on closing delimiter indentation
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

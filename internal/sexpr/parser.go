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
		_, err := p.parseNode()
		if err != nil {
			return nil, fmt.Errorf("line %d: discard (#_) must be followed by a form: %w", tok.Line, err)
		}
		return nil, nil

	case TokenLParen:
		return p.parseList()

	case TokenLBrace:
		return p.parseMap()

	case TokenRParen:
		return nil, fmt.Errorf("line %d: unexpected )", tok.Line)

	case TokenRBrace:
		return nil, fmt.Errorf("line %d: unexpected }", tok.Line)

	case TokenString, TokenKeyword, TokenSymbol:
		p.pos++
		return &Node{Atom: &tok, Line: tok.Line}, nil

	default:
		return nil, fmt.Errorf("line %d: unexpected token %v", tok.Line, tok.Type)
	}
}

func (p *parser) parseList() (*Node, error) {
	open := p.tokens[p.pos]
	p.pos++ // skip (

	children := make([]*Node, 0)
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
		if child != nil {
			children = append(children, child)
		}
	}
}

func (p *parser) parseMap() (*Node, error) {
	open := p.tokens[p.pos]
	p.pos++ // skip {

	children := make([]*Node, 0)
	for {
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("line %d: unterminated map", open.Line)
		}
		if p.tokens[p.pos].Type == TokenRBrace {
			p.pos++ // skip }
			return &Node{Children: children, IsMap: true, Line: open.Line}, nil
		}
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			children = append(children, child)
		}
	}
}

// internal/sexpr/ast.go
package sexpr

// Node is an element in the AST — either an atom or a list.
type Node struct {
	// Exactly one of Atom or Children is set.
	Atom     *Token  // non-nil for leaf nodes (string, keyword)
	Children []*Node // non-nil for list nodes (...) or map nodes {...}
	IsMap    bool    // true if this node is a map (children are key/value pairs)
	Line     int     // source line for error messages
}

// IsAtom returns true if this node is a leaf.
func (n *Node) IsAtom() bool { return n.Atom != nil }

// IsList returns true if this node has children.
func (n *Node) IsList() bool { return n.Children != nil }

// StringVal returns the string value of a string or symbol atom, or empty string.
func (n *Node) StringVal() string {
	if n.Atom != nil && (n.Atom.Type == TokenString || n.Atom.Type == TokenSymbol) {
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

// SymbolVal returns the symbol name, or empty string.
func (n *Node) SymbolVal() string {
	if n.Atom != nil && n.Atom.Type == TokenSymbol {
		return n.Atom.Val
	}
	return ""
}

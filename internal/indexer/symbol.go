package indexer

import (
	"crypto/sha256"
	"fmt"
)

// Symbol kind constants.
const (
	KindFunction  = "function"
	KindMethod    = "method"
	KindType      = "type"
	KindInterface = "interface"
	KindClass     = "class"
	KindField     = "field"
	KindConst     = "const"
	KindVar       = "var"
	KindImport    = "import"
	KindExport    = "export"
)

// Edge kind constants.
const (
	EdgeContains   = "contains"
	EdgeImports    = "imports"
	EdgeExports    = "exports"
	EdgeExtends    = "extends"
	EdgeImplements = "implements"
	EdgeCalls      = "calls"
)

// SymbolDoc is a symbol node in the code graph.
type SymbolDoc struct {
	ID        string `json:"id"`
	File      string `json:"file"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Language  string `json:"language"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	ParentID  string `json:"parent_id,omitempty"`
	Docstring string `json:"docstring,omitempty"`
	FileHash  string `json:"file_hash"`
	Repo      string `json:"repo"`
	IndexedAt string `json:"indexed_at"`
}

// EdgeDoc is a directed edge in the code graph.
type EdgeDoc struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Kind     string `json:"kind"`
	File     string `json:"file"`
	Repo     string `json:"repo"`
}

// SymbolID returns a deterministic ID for a symbol: the first 12 bytes of
// SHA-256("file:kind:name:startLine"), hex-encoded (24 characters).
func SymbolID(file, kind, name string, startLine int) string {
	input := fmt.Sprintf("%s:%s:%s:%d", file, kind, name, startLine)
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h[:12])
}

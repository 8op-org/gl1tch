package indexer

import (
	"context"
	"strings"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
)

// SymbolQuery is a tree-sitter query that extracts symbols of a given kind.
// The query must capture @name (symbol name) and @decl (full declaration node).
type SymbolQuery struct {
	Query string
	Kind  string
}

// CallSite is an unresolved call extracted from source.
type CallSite struct {
	CalleeName string
	File       string
	Line       int
}

// UnresolvedImport is an import extracted from source before cross-file resolution.
type UnresolvedImport struct {
	Path  string
	Alias string
	Names []string
	File  string
	Line  int
}

// LanguageExtractor holds tree-sitter queries for a single language.
type LanguageExtractor struct {
	Language      string
	Grammar       *sitter.Language
	Extensions    []string
	SymbolQueries []SymbolQuery
	ImportQuery   string
	ExportQuery   string
	CallQuery     string
	PathResolver  func(importPath, fromFile, repoRoot string) string
}

// Extract parses source with tree-sitter, runs all SymbolQueries, and returns
// a SymbolDoc for each match. @name captures the symbol name; @decl captures
// the full declaration node for line range and signature.
func (le *LanguageExtractor) Extract(source []byte, file, repo, fileHash string) ([]SymbolDoc, error) {
	if len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()
	root := tree.RootNode()

	now := time.Now().UTC().Format(time.RFC3339)
	var symbols []SymbolDoc

	for _, sq := range le.SymbolQueries {
		q, err := sitter.NewQuery([]byte(sq.Query), le.Grammar)
		if err != nil {
			return nil, err
		}
		defer q.Close()

		cursor := sitter.NewQueryCursor()
		defer cursor.Close()
		cursor.Exec(q, root)

		for {
			match, ok := cursor.NextMatch()
			if !ok {
				break
			}

			var name string
			var declNode *sitter.Node

			for _, cap := range match.Captures {
				capName := q.CaptureNameForId(cap.Index)
				switch capName {
				case "name":
					name = cap.Node.Content(source)
				case "decl":
					declNode = cap.Node
				}
			}

			if name == "" || declNode == nil {
				continue
			}

			startLine := int(declNode.StartPoint().Row) + 1
			endLine := int(declNode.EndPoint().Row) + 1

			// Signature is the first line of the declaration text.
			sig := firstLine(declNode.Content(source))

			symbols = append(symbols, SymbolDoc{
				ID:        SymbolID(file, sq.Kind, name, startLine),
				File:      file,
				Kind:      sq.Kind,
				Name:      name,
				Signature: sig,
				Language:  le.Language,
				StartLine: startLine,
				EndLine:   endLine,
				FileHash:  fileHash,
				Repo:      repo,
				IndexedAt: now,
			})
		}
	}

	return symbols, nil
}

// ExtractImports runs the ImportQuery against source and returns unresolved
// imports. If ImportQuery is empty, returns nil.
func (le *LanguageExtractor) ExtractImports(source []byte, file string) ([]UnresolvedImport, error) {
	if le.ImportQuery == "" {
		return nil, nil
	}
	if len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(le.ImportQuery), le.Grammar)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(q, root)

	var imports []UnresolvedImport

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		var path, alias string
		var line int

		for _, cap := range match.Captures {
			capName := q.CaptureNameForId(cap.Index)
			switch capName {
			case "path":
				path = trimQuotes(cap.Node.Content(source))
				line = int(cap.Node.StartPoint().Row) + 1
			case "alias":
				alias = cap.Node.Content(source)
			}
		}

		if path == "" {
			continue
		}

		imports = append(imports, UnresolvedImport{
			Path:  path,
			Alias: alias,
			File:  file,
			Line:  line,
		})
	}

	return imports, nil
}

// ExtractCalls runs the CallQuery against source and returns unresolved call
// sites. If CallQuery is empty, returns nil.
func (le *LanguageExtractor) ExtractCalls(source []byte, file string) ([]CallSite, error) {
	if le.CallQuery == "" {
		return nil, nil
	}
	if len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(le.CallQuery), le.Grammar)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(q, root)

	var calls []CallSite

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, cap := range match.Captures {
			capName := q.CaptureNameForId(cap.Index)
			if capName == "callee" {
				calls = append(calls, CallSite{
					CalleeName: cap.Node.Content(source),
					File:       file,
					Line:       int(cap.Node.StartPoint().Row) + 1,
				})
			}
		}
	}

	return calls, nil
}

// trimQuotes strips surrounding quotes (double, single, or backtick) from s.
func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	first := s[0]
	last := s[len(s)-1]
	if (first == '"' && last == '"') ||
		(first == '\'' && last == '\'') ||
		(first == '`' && last == '`') {
		return s[1 : len(s)-1]
	}
	return s
}

// firstLine returns the first line of s.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

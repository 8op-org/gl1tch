package indexer

import (
	"strings"
)

// Resolver takes the full symbol set and resolves cross-file references into
// EdgeDoc edges (contains, imports, calls, exports).
type Resolver struct {
	symbols    []SymbolDoc
	byFile     map[string][]SymbolDoc
	byName     map[string][]SymbolDoc
	byID       map[string]SymbolDoc
	extractors map[string]*LanguageExtractor
	repo       string
	repoRoot   string
	ModulePath string // Go module path (set externally)
}

// NewResolver builds a Resolver with lookup maps from the symbol list.
// extractors maps language name to its LanguageExtractor (may be nil).
func NewResolver(symbols []SymbolDoc, repo, repoRoot string, extractors map[string]*LanguageExtractor) *Resolver {
	r := &Resolver{
		symbols:    symbols,
		byFile:     make(map[string][]SymbolDoc),
		byName:     make(map[string][]SymbolDoc),
		byID:       make(map[string]SymbolDoc),
		extractors: extractors,
		repo:       repo,
		repoRoot:   repoRoot,
	}
	for _, s := range symbols {
		r.byFile[s.File] = append(r.byFile[s.File], s)
		r.byName[s.Name] = append(r.byName[s.Name], s)
		r.byID[s.ID] = s
	}
	return r
}

// ResolveContains emits a "contains" edge from parent to child for every
// symbol with a non-empty ParentID. Skips if parent ID not found.
func (r *Resolver) ResolveContains() []EdgeDoc {
	var edges []EdgeDoc
	for _, s := range r.symbols {
		if s.ParentID == "" {
			continue
		}
		parent, ok := r.byID[s.ParentID]
		if !ok {
			continue
		}
		edges = append(edges, EdgeDoc{
			SourceID: parent.ID,
			TargetID: s.ID,
			Kind:     EdgeContains,
			File:     s.File,
			Repo:     r.repo,
		})
	}
	return edges
}

// importableKind returns true for symbol kinds that represent importable
// top-level declarations.
func importableKind(kind string) bool {
	switch kind {
	case KindFunction, KindType, KindClass, KindInterface, KindConst, KindVar:
		return true
	}
	return false
}

// ResolveImports resolves import paths to target files/symbols and creates
// "imports" edges from the importing file to each target symbol.
func (r *Resolver) ResolveImports(unresolved []UnresolvedImport) []EdgeDoc {
	var edges []EdgeDoc
	for _, imp := range unresolved {
		relDir := r.resolveImportPath(imp)
		if relDir == "" {
			continue
		}

		sourceID := fileSymbolID(imp.File, r.repo)

		// Find symbols in files under the resolved directory.
		for file, syms := range r.byFile {
			if !strings.HasPrefix(file, relDir+"/") && file != relDir {
				continue
			}
			for _, s := range syms {
				if !importableKind(s.Kind) {
					continue
				}
				edges = append(edges, EdgeDoc{
					SourceID: sourceID,
					TargetID: s.ID,
					Kind:     EdgeImports,
					File:     imp.File,
					Repo:     r.repo,
				})
			}
		}
	}
	return edges
}

// ResolveCalls resolves call sites into "calls" edges. For each call site it
// finds the target symbol by name (local scope first, then global) and the
// enclosing caller function. If either is missing, the call is dropped.
func (r *Resolver) ResolveCalls(calls []CallSite) []EdgeDoc {
	var edges []EdgeDoc
	for _, c := range calls {
		targetID := r.findCallTarget(c)
		if targetID == "" {
			continue
		}
		callerID := r.findEnclosing(c.File, c.Line)
		if callerID == "" {
			continue
		}
		edges = append(edges, EdgeDoc{
			SourceID: callerID,
			TargetID: targetID,
			Kind:     EdgeCalls,
			File:     c.File,
			Repo:     r.repo,
		})
	}
	return edges
}

// resolveImportPath converts an import path to a repo-relative directory.
// It looks up the language-specific PathResolver for the importing file; if
// none is found or the resolver returns empty, falls back to Go module path
// stripping and then the raw path.
func (r *Resolver) resolveImportPath(imp UnresolvedImport) string {
	// Try language-specific resolver first.
	if r.extractors != nil {
		lang := DetectLanguage(imp.File)
		if ext, ok := r.extractors[lang]; ok && ext.PathResolver != nil {
			resolved := ext.PathResolver(imp.Path, imp.File, r.repoRoot)
			if resolved != "" {
				return resolved
			}
		}
	}

	// Fallback: Go module path stripping.
	if r.ModulePath != "" && strings.HasPrefix(imp.Path, r.ModulePath) {
		rel := strings.TrimPrefix(imp.Path, r.ModulePath)
		rel = strings.TrimPrefix(rel, "/")
		return rel
	}
	return imp.Path
}

// findCallTarget finds the target symbol for a call. It checks same-file
// symbols first (local scope), then falls back to any symbol by name.
func (r *Resolver) findCallTarget(call CallSite) string {
	// Local scope: same file, matching name.
	if syms, ok := r.byFile[call.File]; ok {
		for _, s := range syms {
			if s.Name == call.CalleeName && (s.Kind == KindFunction || s.Kind == KindMethod) {
				return s.ID
			}
		}
	}
	// Global: any symbol by name.
	if syms, ok := r.byName[call.CalleeName]; ok {
		for _, s := range syms {
			if s.Kind == KindFunction || s.Kind == KindMethod {
				return s.ID
			}
		}
	}
	return ""
}

// findEnclosing finds the narrowest function or method that contains the
// given line in the given file.
func (r *Resolver) findEnclosing(file string, line int) string {
	syms, ok := r.byFile[file]
	if !ok {
		return ""
	}

	var best SymbolDoc
	bestSpan := -1

	for _, s := range syms {
		if s.Kind != KindFunction && s.Kind != KindMethod {
			continue
		}
		if line < s.StartLine || line > s.EndLine {
			continue
		}
		span := s.EndLine - s.StartLine
		if bestSpan < 0 || span < bestSpan {
			best = s
			bestSpan = span
		}
	}

	if bestSpan < 0 {
		return ""
	}
	return best.ID
}

// ResolveExports creates "exports" edges from a file pseudo-symbol to each
// top-level symbol in that file for languages that have explicit exports
// (JavaScript, TypeScript, TSX, JSX). For other languages exports are
// implicit and this returns nil.
func (r *Resolver) ResolveExports() []EdgeDoc {
	exportLangs := map[string]bool{
		"javascript": true,
		"typescript": true,
		"tsx":        true,
		"jsx":        true,
	}

	var edges []EdgeDoc
	for file, syms := range r.byFile {
		// Only emit export edges for JS/TS family files.
		lang := DetectLanguage(file)
		if !exportLangs[lang] {
			continue
		}

		sourceID := fileSymbolID(file, r.repo)
		for _, s := range syms {
			if !importableKind(s.Kind) {
				continue
			}
			edges = append(edges, EdgeDoc{
				SourceID: sourceID,
				TargetID: s.ID,
				Kind:     EdgeExports,
				File:     file,
				Repo:     r.repo,
			})
		}
	}
	return edges
}

// ResolveExtendsImplements is a placeholder for extends/implements edge
// resolution. Go uses implicit interfaces (requires type-checker, not just
// AST), and other languages would need additional tree-sitter queries for
// superclass/interface capture. Returns nil for now.
//
// TODO: Add tree-sitter queries for explicit inheritance in Python
// (class Foo(Base)), Java/C#/Kotlin (extends/implements clauses), and
// Scala (extends). Go implicit interfaces require go/types analysis.
func (r *Resolver) ResolveExtendsImplements() []EdgeDoc {
	return nil
}

// fileSymbolID returns a deterministic ID for a file-level pseudo-symbol.
func fileSymbolID(file, repo string) string {
	return SymbolID(file, "file", file, 0)
}

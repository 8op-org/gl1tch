package indexer

import (
	"strings"
)

// Resolver takes the full symbol set and resolves cross-file references into
// EdgeDoc edges (contains, imports, calls).
type Resolver struct {
	symbols    []SymbolDoc
	byFile     map[string][]SymbolDoc
	byName     map[string][]SymbolDoc
	byID       map[string]SymbolDoc
	repo       string
	repoRoot   string
	ModulePath string // Go module path (set externally)
}

// NewResolver builds a Resolver with lookup maps from the symbol list.
func NewResolver(symbols []SymbolDoc, repo, repoRoot string) *Resolver {
	r := &Resolver{
		symbols:  symbols,
		byFile:   make(map[string][]SymbolDoc),
		byName:   make(map[string][]SymbolDoc),
		byID:     make(map[string]SymbolDoc),
		repo:     repo,
		repoRoot: repoRoot,
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
// For Go imports, it strips the module path prefix. For others, returns the
// path directly.
func (r *Resolver) resolveImportPath(imp UnresolvedImport) string {
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

// fileSymbolID returns a deterministic ID for a file-level pseudo-symbol.
func fileSymbolID(file, repo string) string {
	return SymbolID(file, "file", file, 0)
}

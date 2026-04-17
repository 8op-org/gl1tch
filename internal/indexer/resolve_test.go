package indexer

import (
	"testing"
)

func TestResolveImportsGoLocal(t *testing.T) {
	// Two symbols in different files under "pkg/util", one import pointing there.
	modPath := "github.com/example/repo"
	repo := "example/repo"
	repoRoot := "/src/repo"

	symbols := []SymbolDoc{
		{
			ID:       SymbolID("pkg/util/helper.go", KindFunction, "HelperFunc", 5),
			File:     "pkg/util/helper.go",
			Kind:     KindFunction,
			Name:     "HelperFunc",
			Language: "go",
			Repo:     repo,
		},
		{
			ID:       SymbolID("pkg/util/types.go", KindType, "Config", 1),
			File:     "pkg/util/types.go",
			Kind:     KindType,
			Name:     "Config",
			Language: "go",
			Repo:     repo,
		},
	}

	r := NewResolver(symbols, repo, repoRoot)
	r.ModulePath = modPath

	imports := []UnresolvedImport{
		{
			Path: modPath + "/pkg/util",
			File: "cmd/main.go",
			Line: 3,
		},
	}

	edges := r.ResolveImports(imports)

	if len(edges) != 2 {
		t.Fatalf("expected 2 import edges, got %d", len(edges))
	}

	sourceID := fileSymbolID("cmd/main.go", repo)
	for _, e := range edges {
		if e.Kind != EdgeImports {
			t.Errorf("expected edge kind %q, got %q", EdgeImports, e.Kind)
		}
		if e.SourceID != sourceID {
			t.Errorf("expected source %q, got %q", sourceID, e.SourceID)
		}
		if e.Repo != repo {
			t.Errorf("expected repo %q, got %q", repo, e.Repo)
		}
	}
}

func TestResolveContainsEdges(t *testing.T) {
	repo := "example/repo"
	repoRoot := "/src/repo"

	parentID := SymbolID("pkg/foo.go", KindType, "Server", 10)
	childID := SymbolID("pkg/foo.go", KindMethod, "Start", 20)

	symbols := []SymbolDoc{
		{
			ID:        parentID,
			File:      "pkg/foo.go",
			Kind:      KindType,
			Name:      "Server",
			StartLine: 10,
			EndLine:   50,
			Repo:      repo,
		},
		{
			ID:        childID,
			File:      "pkg/foo.go",
			Kind:      KindMethod,
			Name:      "Start",
			StartLine: 20,
			EndLine:   30,
			ParentID:  parentID,
			Repo:      repo,
		},
	}

	r := NewResolver(symbols, repo, repoRoot)
	edges := r.ResolveContains()

	if len(edges) != 1 {
		t.Fatalf("expected 1 contains edge, got %d", len(edges))
	}

	e := edges[0]
	if e.SourceID != parentID {
		t.Errorf("expected source %q, got %q", parentID, e.SourceID)
	}
	if e.TargetID != childID {
		t.Errorf("expected target %q, got %q", childID, e.TargetID)
	}
	if e.Kind != EdgeContains {
		t.Errorf("expected kind %q, got %q", EdgeContains, e.Kind)
	}
}

func TestResolveCallsLocalScope(t *testing.T) {
	repo := "example/repo"
	repoRoot := "/src/repo"

	callerID := SymbolID("pkg/foo.go", KindFunction, "main", 1)
	targetID := SymbolID("pkg/foo.go", KindFunction, "helper", 10)

	symbols := []SymbolDoc{
		{
			ID:        callerID,
			File:      "pkg/foo.go",
			Kind:      KindFunction,
			Name:      "main",
			StartLine: 1,
			EndLine:   8,
			Repo:      repo,
		},
		{
			ID:        targetID,
			File:      "pkg/foo.go",
			Kind:      KindFunction,
			Name:      "helper",
			StartLine: 10,
			EndLine:   15,
			Repo:      repo,
		},
	}

	r := NewResolver(symbols, repo, repoRoot)

	calls := []CallSite{
		{
			CalleeName: "helper",
			File:       "pkg/foo.go",
			Line:       5,
		},
	}

	edges := r.ResolveCalls(calls)

	if len(edges) != 1 {
		t.Fatalf("expected 1 calls edge, got %d", len(edges))
	}

	e := edges[0]
	if e.SourceID != callerID {
		t.Errorf("expected source (caller) %q, got %q", callerID, e.SourceID)
	}
	if e.TargetID != targetID {
		t.Errorf("expected target (callee) %q, got %q", targetID, e.TargetID)
	}
	if e.Kind != EdgeCalls {
		t.Errorf("expected kind %q, got %q", EdgeCalls, e.Kind)
	}
}

func TestResolveCallsUnresolvedDropped(t *testing.T) {
	repo := "example/repo"
	repoRoot := "/src/repo"

	callerID := SymbolID("pkg/foo.go", KindFunction, "main", 1)

	symbols := []SymbolDoc{
		{
			ID:        callerID,
			File:      "pkg/foo.go",
			Kind:      KindFunction,
			Name:      "main",
			StartLine: 1,
			EndLine:   8,
			Repo:      repo,
		},
	}

	r := NewResolver(symbols, repo, repoRoot)

	calls := []CallSite{
		{
			CalleeName: "nonExistentFunc",
			File:       "pkg/foo.go",
			Line:       5,
		},
	}

	edges := r.ResolveCalls(calls)

	if len(edges) != 0 {
		t.Fatalf("expected 0 edges for unresolved call, got %d", len(edges))
	}
}

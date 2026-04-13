package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
)

func buildResearchLoop() *research.Loop {
	reg := research.NewRegistry()

	// 1. Load YAML researchers from ~/.config/glitch/researchers/
	if home, err := os.UserHomeDir(); err == nil {
		researcherDir := filepath.Join(home, ".config", "glitch", "researchers")
		research.LoadResearchers(researcherDir, reg, providerReg)
	}

	// 2. Load YAML researchers from .glitch/researchers/ (project-local)
	research.LoadResearchers(".glitch/researchers", reg, providerReg)

	// 3. Add ES researchers if ES is reachable
	es := esearch.NewClient("http://localhost:9200")
	if err := es.Ping(context.Background()); err == nil {
		reg.Register(research.NewESActivityResearcher(es))
		reg.Register(research.NewESCodeResearcher(es))
	}

	if len(reg.Names()) == 0 {
		return nil
	}

	llm := func(ctx context.Context, prompt string) (string, error) {
		return provider.RunOllama("qwen2.5:7b", prompt)
	}

	return research.NewLoop(reg, llm)
}

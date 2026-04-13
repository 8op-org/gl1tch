package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
)

// buildResearchLoop assembles the research loop with all available researchers.
func buildResearchLoop() *research.Loop {
	reg := research.NewRegistry()

	// 1. Native researchers — always available
	reg.Register(&research.GitResearcher{})
	reg.Register(&research.FSResearcher{})

	// 2. ES researchers if ES is reachable
	es := esearch.NewClient("http://localhost:9200")
	if err := es.Ping(context.Background()); err == nil {
		reg.Register(research.NewESActivityResearcher(es))
		reg.Register(research.NewESCodeResearcher(es))
	}

	// 3. YAML researchers from ~/.config/glitch/researchers/
	if home, err := os.UserHomeDir(); err == nil {
		researcherDir := filepath.Join(home, ".config", "glitch", "researchers")
		research.LoadResearchers(researcherDir, reg, providerReg)
	}

	// 4. YAML researchers from .glitch/researchers/ (project-local)
	research.LoadResearchers(".glitch/researchers", reg, providerReg)

	// Local LLM for plan/critique/score
	localLLM := func(ctx context.Context, prompt string) (string, error) {
		return provider.RunOllama("qwen2.5:7b", prompt)
	}

	// Paid LLM for drafting (uses provider system via stdin)
	var draftLLM research.LLMFn
	if _, err := providerReg.RenderCommand("claude", "test"); err == nil {
		draftLLM = func(ctx context.Context, prompt string) (string, error) {
			return providerReg.RunProvider("claude", prompt)
		}
	}

	loop := research.NewLoop(reg, localLLM)
	if draftLLM != nil {
		loop = loop.WithDraftLLM(draftLLM)
	}

	return loop
}

// resultsDir returns the results directory for a given question.
func resultsDir(question string) string {
	org, repo := research.ParseRepoFromQuestion(question)
	if org != "" && repo != "" {
		return filepath.Join("results", org, repo)
	}
	return filepath.Join("results", "general")
}

// printFeedback prints feedback to stderr.
func printFeedback(fb research.Feedback) {
	if fb.Suggestion != "" {
		fmt.Fprintf(os.Stderr, ">> learned: %s\n", fb.Suggestion)
	}
	if fb.Quality != "" {
		fmt.Fprintf(os.Stderr, ">> evidence quality: %s\n", fb.Quality)
	}
	for _, m := range fb.Missing {
		fmt.Fprintf(os.Stderr, ">> missing evidence: %s\n", m)
	}
}

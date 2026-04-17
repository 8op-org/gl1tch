package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/observer"
	"github.com/8op-org/gl1tch/internal/provider"
)

var (
	observeRepo     string
	observeProvider string
	observeModel    string
	observeDepth    int
)

func init() {
	observeCmd.Flags().StringVar(&observeRepo, "repo", "", "scope query to a specific repository (e.g. elastic/kibana)")
	observeCmd.Flags().StringVar(&observeProvider, "provider", "copilot", "LLM provider for query generation and synthesis")
	observeCmd.Flags().StringVar(&observeModel, "model", "", "model name (provider-specific)")
	observeCmd.Flags().IntVar(&observeDepth, "depth", 1, "BFS traversal depth for graph queries")
	rootCmd.AddCommand(observeCmd)
}

var observeCmd = &cobra.Command{
	Use:   "observe [question]",
	Short: "query indexed activity via elasticsearch",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := strings.Join(args, " ")

		es := esearch.NewClient("http://localhost:9200")
		if err := es.Ping(context.Background()); err != nil {
			return fmt.Errorf("elasticsearch is not running — start with: glitch up")
		}

		// Build the LLM function from the provider.
		// Try known agent providers first (copilot, claude, gemini),
		// then fall back to YAML-defined providers, then Ollama.
		llm := buildLLMFunc(observeProvider, observeModel)

		engine := observer.NewQueryEngine(es, llm)
		if observeRepo != "" {
			engine = engine.WithRepo(observeRepo)
		}
		if observeDepth > 0 {
			engine = engine.WithDepth(observeDepth)
		}
		answer, err := engine.Answer(cmd.Context(), question)
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	},
}

func buildLLMFunc(provName, model string) observer.LLMFunc {
	// Check known agent providers (copilot, claude, gemini)
	if agent, ok := provider.KnownAgents[provName]; ok {
		return func(prompt string) (string, error) {
			result, err := agent.Run(model, prompt)
			if err != nil {
				return "", err
			}
			return result.Response, nil
		}
	}

	// Try YAML-defined provider registry
	if providerReg != nil {
		return func(prompt string) (string, error) {
			return providerReg.RunProvider(provName, model, prompt)
		}
	}

	// Fallback: Ollama with the provider name as model
	m := provName
	if model != "" {
		m = model
	}
	return func(prompt string) (string, error) {
		return provider.RunOllama(m, prompt)
	}
}

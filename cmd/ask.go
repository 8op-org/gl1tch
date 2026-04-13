package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/research"
	"github.com/8op-org/gl1tch/internal/router"
)

var reGitHubIssueURL = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/issues/\d+`)
var reGitHubPRURL = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/pull/\d+`)
var reGoogleDocURL = regexp.MustCompile(`https?://docs\.google\.com/document/d/`)

func init() {
	askCmd.Flags().StringVarP(&targetPath, "path", "C", "", "run against this directory instead of cwd")
	rootCmd.AddCommand(askCmd)
}

func isResearchInput(input string) bool {
	return reGitHubIssueURL.MatchString(input) ||
		reGitHubPRURL.MatchString(input) ||
		reGoogleDocURL.MatchString(input)
}

var askCmd = &cobra.Command{
	Use:   "ask [input]",
	Short: "route a question or URL to the best workflow",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetPath != "" {
			if err := os.Chdir(targetPath); err != nil {
				return fmt.Errorf("chdir %s: %w", targetPath, err)
			}
		}
		input := strings.Join(args, " ")

		// Research inputs skip the workflow router
		if isResearchInput(input) {
			return runResearch(input, research.GoalSummarize)
		}

		// Check for "implement" intent
		if strings.Contains(strings.ToLower(input), "implement") {
			if url := reGitHubIssueURL.FindString(input); url != "" {
				return runResearch(url, research.GoalImplement)
			}
		}

		// Tier 1: workflow match
		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}
		w, resolved, params := router.Match(input, workflows, "")
		if w != nil {
			fmt.Printf(">> %s\n", w.Name)
			result, err := pipeline.Run(w, resolved, "", params, providerReg)
			if err != nil {
				return err
			}
			fmt.Println(result.Output)
			return nil
		}

		// Tier 2: research loop for non-URL questions
		return runResearch(input, research.GoalSummarize)
	},
}

func runResearch(input string, goal research.Goal) error {
	fmt.Fprintln(os.Stderr, ">> adapting input...")
	doc, err := research.Adapt(input)
	if err != nil {
		return fmt.Errorf("adapt: %w", err)
	}
	fmt.Fprintf(os.Stderr, ">> source: %s\n", doc.Source)
	if doc.Repo != "" {
		fmt.Fprintf(os.Stderr, ">> repo: %s\n", doc.Repo)
	}

	repoPath := doc.RepoPath
	if repoPath == "" {
		repoPath, _ = os.Getwd()
	}

	fmt.Fprintln(os.Stderr, ">> researching...")
	loop, err := buildToolLoop(repoPath)
	if err != nil {
		return fmt.Errorf("build loop: %w", err)
	}

	result, err := loop.Run(context.Background(), doc, goal)
	if err != nil {
		return fmt.Errorf("research: %w", err)
	}

	fmt.Println(result.Output)

	// Save results
	if saveErr := research.SaveLoopResult("results", result); saveErr != nil {
		fmt.Fprintf(os.Stderr, ">> warning: could not save results: %v\n", saveErr)
	} else {
		fmt.Fprintf(os.Stderr, ">> results saved to results/\n")
	}

	fmt.Fprintf(os.Stderr, ">> %d tool calls, %d LLM calls, ~$%.4f\n",
		len(result.ToolCalls), result.LLMCalls, result.CostUSD)

	return nil
}

func loadWorkflows() (map[string]*pipeline.Workflow, error) {
	workflows := make(map[string]*pipeline.Workflow)
	if home, err := os.UserHomeDir(); err == nil {
		globalDir := home + "/.config/glitch/workflows"
		if m, err := pipeline.LoadDir(globalDir); err == nil {
			for k, v := range m {
				workflows[k] = v
			}
		}
	}
	if m, err := pipeline.LoadDir(".glitch/workflows"); err == nil {
		for k, v := range m {
			workflows[k] = v
		}
	}
	return workflows, nil
}

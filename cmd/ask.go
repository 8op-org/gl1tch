package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/router"
)

// EmbeddedWorkflows is set by main.go before Execute().
var EmbeddedWorkflows embed.FS

func init() {
	rootCmd.AddCommand(askCmd)
}

var askCmd = &cobra.Command{
	Use:   "ask [input]",
	Short: "route a question or URL to the best workflow",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := strings.Join(args, " ")

		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}

		w, resolved := router.Match(input, workflows, "")
		if w == nil {
			fmt.Fprintf(os.Stderr, "no matching workflow for: %s\n", input)
			fmt.Fprintln(os.Stderr, "available workflows:")
			for name := range workflows {
				fmt.Fprintf(os.Stderr, "  - %s\n", name)
			}
			return nil
		}

		fmt.Printf(">> %s\n", w.Name)
		result, err := pipeline.Run(w, resolved, "")
		if err != nil {
			return err
		}
		fmt.Println(result.Output)
		return nil
	},
}

// loadWorkflows merges embedded, user-global, and project-local workflows.
// Later sources override earlier ones.
func loadWorkflows() (map[string]*pipeline.Workflow, error) {
	workflows := make(map[string]*pipeline.Workflow)

	// 1. Embedded defaults.
	entries, err := EmbeddedWorkflows.ReadDir("workflows")
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, err := EmbeddedWorkflows.ReadFile("workflows/" + e.Name())
			if err != nil {
				continue
			}
			w, err := pipeline.LoadBytes(data, e.Name())
			if err != nil {
				continue
			}
			workflows[w.Name] = w
		}
	}

	// 2. User-global: ~/.config/glitch/workflows/
	if home, err := os.UserHomeDir(); err == nil {
		globalDir := filepath.Join(home, ".config", "glitch", "workflows")
		if m, err := pipeline.LoadDir(globalDir); err == nil {
			for k, v := range m {
				workflows[k] = v
			}
		}
	}

	// 3. Project-local: .glitch/workflows/
	if m, err := pipeline.LoadDir(".glitch/workflows"); err == nil {
		for k, v := range m {
			workflows[k] = v
		}
	}

	return workflows, nil
}

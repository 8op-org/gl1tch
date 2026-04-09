package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/router"
)

func init() {
	askCmd.Flags().StringVarP(&targetPath, "path", "C", "", "run against this directory instead of cwd")
	rootCmd.AddCommand(askCmd)
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

// loadWorkflows loads from ~/.config/glitch/workflows/ then .glitch/workflows/.
// Project-local workflows override global ones.
func loadWorkflows() (map[string]*pipeline.Workflow, error) {
	workflows := make(map[string]*pipeline.Workflow)

	// 1. User-global: ~/.config/glitch/workflows/
	if home, err := os.UserHomeDir(); err == nil {
		globalDir := filepath.Join(home, ".config", "glitch", "workflows")
		if m, err := pipeline.LoadDir(globalDir); err == nil {
			for k, v := range m {
				workflows[k] = v
			}
		}
	}

	// 2. Project-local: .glitch/workflows/
	if m, err := pipeline.LoadDir(".glitch/workflows"); err == nil {
		for k, v := range m {
			workflows[k] = v
		}
	}

	return workflows, nil
}

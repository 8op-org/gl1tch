package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

var checkCmd = &cobra.Command{
	Use:   "check <file>",
	Short: "validate a workflow file without running it",
	Long: `Statically analyze a .glitch workflow file for common errors:
  - Unknown functions or forms
  - Bad argument counts
  - Undefined references
  - Definitions that shadow builtins

Example:
  glitch check workflows/deploy.glitch`,
	Args: cobra.ExactArgs(1),
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	path := args[0]
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	nodes, err := sexpr.Parse(src)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: parse error: %s\n", path, err)
		return fmt.Errorf("parse failed")
	}

	diags := pipeline.Check(nodes)
	if len(diags) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: ok\n", path)
		return nil
	}

	hasError := false
	for _, d := range diags {
		prefix := "warning"
		if d.Severity == pipeline.SeverityError {
			prefix = "error"
			hasError = true
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s:%d: %s: %s\n", path, d.Line, prefix, d.Message)
	}

	if hasError {
		return fmt.Errorf("check failed with errors")
	}
	return nil
}

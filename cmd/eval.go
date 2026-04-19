package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

var evalExpr string

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "interactive evaluator for glitch expressions",
	Long: `Start an interactive REPL for experimenting with glitch workflow expressions.

Use -e to evaluate a single expression:
  glitch eval -e '(str "hello" " " "world")'

Without -e, starts an interactive session:
  glitch eval`,
	RunE: runEval,
}

func init() {
	evalCmd.Flags().StringVarP(&evalExpr, "expr", "e", "", "evaluate a single expression and exit")
	rootCmd.AddCommand(evalCmd)
}

func runEval(cmd *cobra.Command, args []string) error {
	ev := pipeline.NewEvaluator()

	if evalExpr != "" {
		env := pipeline.NewEnv(nil)
		ev.RegisterBuiltins(env)
		val, err := ev.RunSourceWithEnv(env, []byte(evalExpr))
		if err != nil {
			return fmt.Errorf("eval: %s", err)
		}
		if val != nil && val.String() != "" {
			fmt.Fprintln(cmd.OutOrStdout(), val.String())
		}
		return nil
	}

	// Interactive mode
	r := pipeline.NewREPL(ev, os.Stdin, cmd.OutOrStdout())
	return r.Run()
}

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/observer"
)

func init() {
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

		engine := observer.NewQueryEngine(es, "")
		answer, err := engine.Answer(cmd.Context(), question)
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	},
}

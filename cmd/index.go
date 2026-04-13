package cmd

import (
	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/indexer"
)

func init() {
	rootCmd.AddCommand(indexCmd)
}

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "index a repository into Elasticsearch for code search",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		es := esearch.NewClient("http://localhost:9200")
		return indexer.IndexRepo(path, es)
	},
}

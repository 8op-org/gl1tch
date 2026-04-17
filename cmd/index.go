package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/indexer"
)

var (
	indexRepo        string
	indexESURL       string
	indexLanguages   string
	indexFull        bool
	indexSymbolsOnly bool
	indexStats       bool
)

func init() {
	indexCmd.Flags().StringVar(&indexRepo, "repo", "", "override repo name (default: directory name)")
	indexCmd.Flags().StringVar(&indexESURL, "es-url", "", "Elasticsearch URL (default: http://localhost:9200)")
	indexCmd.Flags().StringVar(&indexLanguages, "languages", "", "comma-separated language filter")
	indexCmd.Flags().BoolVar(&indexFull, "full", false, "force full re-index")
	indexCmd.Flags().BoolVar(&indexSymbolsOnly, "symbols-only", false, "only index symbols + edges, skip content chunks")
	indexCmd.Flags().BoolVar(&indexStats, "stats", false, "print index stats after completion")
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
		esURL := indexESURL
		if esURL == "" {
			esURL = "http://localhost:9200"
		}
		es := esearch.NewClient(esURL)

		opts := indexer.IndexOpts{
			Repo:        indexRepo,
			ESURL:       esURL,
			Full:        indexFull,
			SymbolsOnly: indexSymbolsOnly,
			Stats:       indexStats,
		}
		if indexLanguages != "" {
			opts.Languages = strings.Split(indexLanguages, ",")
		}

		return indexer.IndexRepoGraph(path, es, opts)
	},
}

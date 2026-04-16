package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set via ldflags at build time (goreleaser):
//
//	-X github.com/8op-org/gl1tch/cmd.Version={{.Version}}
//	-X github.com/8op-org/gl1tch/cmd.Commit={{.ShortCommit}}
//	-X github.com/8op-org/gl1tch/cmd.Date={{.Date}}
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	// Wire cobra's built-in --version / -v flag
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("glitch {{.Version}}\n")
}

// SetVersionString updates the root command version after ldflags are forwarded from main.
func SetVersionString() {
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version, commit hash, and build date",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("glitch %s (commit %s, built %s)\n", Version, Commit, Date)
	},
}

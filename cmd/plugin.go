package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)

	// Discover and register glitch-* plugins from PATH.
	discoverPlugins()
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "manage plugins",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed plugins",
	Run: func(cmd *cobra.Command, args []string) {
		plugins := findPluginBinaries()
		if len(plugins) == 0 {
			fmt.Println("no plugins installed")
			fmt.Println("plugins are binaries named glitch-<name> on your PATH")
			return
		}
		for _, p := range plugins {
			fmt.Printf("  %s\t(%s)\n", p.name, p.path)
		}
	},
}

type pluginInfo struct {
	name string
	path string
}

func findPluginBinaries() []pluginInfo {
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	seen := make(map[string]bool)
	var plugins []pluginInfo

	for _, dir := range pathDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, "glitch-") {
				continue
			}
			pluginName := strings.TrimPrefix(name, "glitch-")
			if seen[pluginName] {
				continue
			}
			seen[pluginName] = true
			plugins = append(plugins, pluginInfo{
				name: pluginName,
				path: filepath.Join(dir, name),
			})
		}
	}
	return plugins
}

func discoverPlugins() {
	for _, p := range findPluginBinaries() {
		binPath := p.path
		rootCmd.AddCommand(&cobra.Command{
			Use:                p.name,
			Short:              fmt.Sprintf("plugin: %s", p.name),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				plugin := exec.Command(binPath, args...)
				plugin.Stdin = os.Stdin
				plugin.Stdout = os.Stdout
				plugin.Stderr = os.Stderr
				return plugin.Run()
			},
		})
	}
}

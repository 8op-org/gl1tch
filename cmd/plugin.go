package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/plugin"
	"github.com/8op-org/gl1tch/internal/workspace"
)

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
}

// pluginDirs returns (localDir, globalDir) for plugin discovery.
func pluginDirs() (string, string) {
	local := filepath.Join(".glitch", "plugins")
	global := ""
	if home, err := os.UserHomeDir(); err == nil {
		global = filepath.Join(home, ".config", "glitch", "plugins")
	}
	return local, global
}

// findPlugin locates a plugin by name, preferring local over global.
// Returns (pluginRoot, pluginInfo, error).
func findPlugin(name string) (string, *plugin.PluginInfo, error) {
	localDir, globalDir := pluginDirs()
	plugins := plugin.DiscoverPlugins(localDir, globalDir)
	info, ok := plugins[name]
	if !ok {
		return "", nil, fmt.Errorf("plugin %q not found, searched: %s, %s", name, localDir, globalDir)
	}
	// pluginRoot is the parent dir of the plugin's dir
	pluginRoot := filepath.Dir(info.Dir)
	return pluginRoot, info, nil
}

var pluginCmd = &cobra.Command{
	Use:                "plugin",
	Short:              "manage and run plugins",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
			return cmd.Help()
		}

		// Delegate "list" to pluginListCmd
		if args[0] == "list" {
			return pluginListCmd.RunE(pluginListCmd, args[1:])
		}

		pluginName := args[0]

		// glitch plugin <name>  or  glitch plugin <name> --help
		if len(args) == 1 || (len(args) == 2 && (args[1] == "--help" || args[1] == "-h")) {
			return showPluginHelp(pluginName)
		}

		subcommand := args[1]

		// glitch plugin <name> <sub> --help
		if len(args) == 3 && (args[2] == "--help" || args[2] == "-h") {
			return showSubcommandHelp(pluginName, subcommand)
		}
		if len(args) == 2 && (subcommand == "--help" || subcommand == "-h") {
			return showPluginHelp(pluginName)
		}

		// glitch plugin <name> <sub> [--flags]
		flags := parsePluginFlags(args[2:])

		pluginRoot, _, err := findPlugin(pluginName)
		if err != nil {
			return err
		}

		wsDir := workspacePath
		if wsDir == "" {
			wsDir, _ = os.Getwd()
		}
		wsName := workspace.ResolveWorkspace(wsDir)
		// Resolve workflows directory so plugin-invoked workflows can call-workflow
		// against the active workspace. Empty string is fine — call-workflow will
		// error cleanly if actually used without a resolvable target.
		workflowsDir := ""
		if resolved := resolveWorkspaceForCommand(); resolved.Path != "" {
			workflowsDir = filepath.Join(resolved.Path, "workflows")
		}
		result, err := pipeline.RunPluginSubcommand(pluginRoot, pluginName, subcommand, flags, providerReg, pipeline.RunOpts{Workspace: wsName, WorkflowsDir: workflowsDir})
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "list available plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		localDir, globalDir := pluginDirs()
		plugins := plugin.DiscoverPlugins(localDir, globalDir)

		names := make([]string, 0, len(plugins))
		for name := range plugins {
			names = append(names, name)
		}
		sort.Strings(names)

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "PLUGIN\tSOURCE\tSUBCOMMANDS")
		for _, name := range names {
			info := plugins[name]
			subs := strings.Join(info.Subcommands, ", ")
			fmt.Fprintf(tw, "%s\t%s\t%s\n", info.Name, info.Source, subs)
		}
		return tw.Flush()
	},
}

// parsePluginFlags converts ["--name", "value", "--flag"] into map[string]string.
func parsePluginFlags(args []string) map[string]string {
	flags := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		name := strings.TrimPrefix(arg, "--")
		// Check if next arg is a value (not another flag or end)
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			flags[name] = args[i+1]
			i++
		} else {
			flags[name] = "true"
		}
	}
	return flags
}

// showPluginHelp prints help for a plugin (description, version, subcommands).
func showPluginHelp(name string) error {
	_, info, err := findPlugin(name)
	if err != nil {
		return err
	}

	manifest, err := plugin.LoadManifest(info.Dir)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	fmt.Printf("Plugin: %s", manifest.Name)
	if manifest.Version != "" {
		fmt.Printf(" (%s)", manifest.Version)
	}
	fmt.Println()

	if manifest.Description != "" {
		fmt.Println()
		fmt.Println(manifest.Description)
	}

	if len(info.Subcommands) > 0 {
		fmt.Println()
		fmt.Println("Subcommands:")
		sort.Strings(info.Subcommands)
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		for _, sub := range info.Subcommands {
			desc := subcommandDescription(info.Dir, sub)
			fmt.Fprintf(tw, "  %s\t%s\n", sub, desc)
		}
		tw.Flush()
	} else {
		fmt.Println("\nNo subcommands available.")
	}

	return nil
}

// showSubcommandHelp prints help for a specific plugin subcommand.
func showSubcommandHelp(pluginName, subcommand string) error {
	_, info, err := findPlugin(pluginName)
	if err != nil {
		return err
	}

	subPath := filepath.Join(info.Dir, subcommand+".glitch")
	data, err := os.ReadFile(subPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin %q has no subcommand %q", pluginName, subcommand)
		}
		return err
	}

	// Parse workflow for description
	w, parseErr := pipeline.LoadBytes(data, subcommand+".glitch")
	if parseErr == nil && w.Description != "" {
		fmt.Println(w.Description)
		fmt.Println()
	}

	fmt.Printf("Usage: glitch plugin %s %s [flags]\n", pluginName, subcommand)

	argDefs, err := plugin.ParseArgs(data)
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	if len(argDefs) > 0 {
		fmt.Println()
		fmt.Println("Flags:")
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		for _, def := range argDefs {
			required := ""
			if def.Required {
				required = " (required)"
			}
			defaultStr := ""
			if def.Default != "" {
				defaultStr = fmt.Sprintf(" [default: %s]", def.Default)
			}
			desc := def.Description
			fmt.Fprintf(tw, "  --%s\t%s%s%s\n", def.Name, desc, required, defaultStr)
		}
		tw.Flush()
	}

	return nil
}

// subcommandDescription reads a subcommand's .glitch file and returns its description.
func subcommandDescription(pluginDir, sub string) string {
	subPath := filepath.Join(pluginDir, sub+".glitch")
	data, err := os.ReadFile(subPath)
	if err != nil {
		return ""
	}
	w, err := pipeline.LoadBytes(data, sub+".glitch")
	if err != nil {
		return ""
	}
	desc := strings.TrimSpace(w.Description)
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}
	return desc
}

// internal/pipeline/plugin_runner.go
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/plugin"
	"github.com/8op-org/gl1tch/internal/provider"
)

// RunPluginSubcommand loads and executes a plugin subcommand from a plugin directory root.
// pluginRoot is the parent directory containing plugin directories (e.g., ~/.config/glitch/plugins).
func RunPluginSubcommand(pluginRoot, pluginName, subcommand string, args map[string]string) (string, error) {
	pluginDir := filepath.Join(pluginRoot, pluginName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return "", fmt.Errorf("plugin %q not found in %s", pluginName, pluginRoot)
	}

	// Load manifest for shared defs and metadata
	manifest, err := plugin.LoadManifest(pluginDir)
	if err != nil {
		return "", fmt.Errorf("plugin %q manifest: %w", pluginName, err)
	}

	// Load subcommand file
	subPath := filepath.Join(pluginDir, subcommand+".glitch")
	data, err := os.ReadFile(subPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("plugin %q has no subcommand %q", pluginName, subcommand)
		}
		return "", err
	}

	// Parse args from subcommand
	argDefs, err := plugin.ParseArgs(data)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q args: %w", pluginName, subcommand, err)
	}

	// Build params: manifest defs + resolved args
	if args == nil {
		args = make(map[string]string)
	}
	params, err := plugin.BuildParams(argDefs, args)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q: %w", pluginName, subcommand, err)
	}

	// Inject manifest defs (subcommand params override)
	for k, v := range manifest.Defs {
		if _, exists := params[k]; !exists {
			params[k] = v
		}
	}

	// Parse and run the workflow
	w, err := parseSexprWorkflow(data)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q parse: %w", pluginName, subcommand, err)
	}

	reg, _ := provider.LoadProviders("")

	result, err := Run(w, "", "", params, reg)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q run: %w", pluginName, subcommand, err)
	}
	return strings.TrimSpace(result.Output), nil
}

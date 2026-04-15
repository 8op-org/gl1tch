// internal/pipeline/plugin_runner.go
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/8op-org/gl1tch/internal/plugin"
	"github.com/8op-org/gl1tch/internal/provider"
)

// pluginCache holds parsed plugin artifacts so repeated calls skip file I/O and parsing.
var pluginCache struct {
	manifests sync.Map // pluginDir → *plugin.Manifest
	subs      sync.Map // subcommandPath → *cachedSubcommand
}

type cachedSubcommand struct {
	argDefs []plugin.ArgDef
	wf      *Workflow
}

// RunPluginSubcommand loads and executes a plugin subcommand from a plugin directory root.
// pluginRoot is the parent directory containing plugin directories (e.g., ~/.config/glitch/plugins).
func RunPluginSubcommand(pluginRoot, pluginName, subcommand string, args map[string]string, reg *provider.ProviderRegistry, opts ...RunOpts) (string, error) {
	pluginDir := filepath.Join(pluginRoot, pluginName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return "", fmt.Errorf("plugin %q not found in %s", pluginName, pluginRoot)
	}

	// Load manifest (cached)
	manifest, err := loadManifestCached(pluginDir, pluginName)
	if err != nil {
		return "", err
	}

	// Load and parse subcommand (cached)
	subPath := filepath.Join(pluginDir, subcommand+".glitch")
	cached, err := loadSubcommandCached(subPath, pluginName, subcommand)
	if err != nil {
		return "", err
	}

	// Build params: manifest defs + resolved args
	if args == nil {
		args = make(map[string]string)
	}
	params, err := plugin.BuildParams(cached.argDefs, args)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q: %w", pluginName, subcommand, err)
	}

	// Inject manifest defs (subcommand params override)
	for k, v := range manifest.Defs {
		if _, exists := params[k]; !exists {
			params[k] = v
		}
	}

	var provReg *provider.ProviderRegistry
	if reg != nil {
		provReg = reg
	} else {
		provReg, _ = provider.LoadProviders("")
	}

	result, err := Run(cached.wf, "", "", params, provReg, opts...)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q run: %w", pluginName, subcommand, err)
	}
	return strings.TrimSpace(result.Output), nil
}

func loadManifestCached(pluginDir, pluginName string) (*plugin.Manifest, error) {
	if v, ok := pluginCache.manifests.Load(pluginDir); ok {
		return v.(*plugin.Manifest), nil
	}
	m, err := plugin.LoadManifest(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("plugin %q manifest: %w", pluginName, err)
	}
	pluginCache.manifests.Store(pluginDir, m)
	return m, nil
}

func loadSubcommandCached(subPath, pluginName, subcommand string) (*cachedSubcommand, error) {
	if v, ok := pluginCache.subs.Load(subPath); ok {
		return v.(*cachedSubcommand), nil
	}

	data, err := os.ReadFile(subPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin %q has no subcommand %q", pluginName, subcommand)
		}
		return nil, err
	}

	argDefs, err := plugin.ParseArgs(data)
	if err != nil {
		return nil, fmt.Errorf("plugin %q %q args: %w", pluginName, subcommand, err)
	}

	wf, err := parseSexprWorkflow(data)
	if err != nil {
		return nil, fmt.Errorf("plugin %q %q parse: %w", pluginName, subcommand, err)
	}

	c := &cachedSubcommand{argDefs: argDefs, wf: wf}
	pluginCache.subs.Store(subPath, c)
	return c, nil
}

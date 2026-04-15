package plugin

import (
	"os"
	"path/filepath"
	"strings"
)

// PluginInfo holds metadata about a discovered plugin.
type PluginInfo struct {
	Name        string
	Source      string // "local" or "global"
	Dir         string // absolute path to plugin directory
	Subcommands []string
}

// DiscoverPlugins scans globalDir then localDir for plugin subdirectories
// containing .glitch files. Local plugins overwrite global ones on conflict.
// An empty string for either dir skips that source.
func DiscoverPlugins(localDir, globalDir string) map[string]*PluginInfo {
	result := make(map[string]*PluginInfo)

	for _, entry := range []struct {
		dir    string
		source string
	}{
		{globalDir, "global"},
		{localDir, "local"},
	} {
		if entry.dir == "" {
			continue
		}
		scanDir(entry.dir, entry.source, result)
	}

	return result
}

func scanDir(dir, source string, result map[string]*PluginInfo) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		pluginDir := filepath.Join(dir, e.Name())
		subcommands := glitchFiles(pluginDir)
		if subcommands == nil {
			// no .glitch files found — not a plugin
			continue
		}

		result[e.Name()] = &PluginInfo{
			Name:        e.Name(),
			Source:      source,
			Dir:         pluginDir,
			Subcommands: subcommands,
		}
	}
}

// glitchFiles returns the list of .glitch file stems (minus extension) inside
// dir, excluding "plugin.glitch". Returns nil if the directory cannot be read
// or contains no .glitch files at all (not even plugin.glitch). Returns an
// empty (non-nil) slice when plugin.glitch is the only .glitch file present.
func glitchFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	hasGlitch := false
	var subs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".glitch") {
			continue
		}
		hasGlitch = true
		if name == "plugin.glitch" {
			continue
		}
		subs = append(subs, strings.TrimSuffix(name, ".glitch"))
	}

	if !hasGlitch {
		return nil
	}
	if subs == nil {
		subs = []string{}
	}
	return subs
}

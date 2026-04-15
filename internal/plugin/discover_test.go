package plugin

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// makePlugin creates a fake plugin directory with the given .glitch files.
func makePlugin(t *testing.T, base, pluginName string, files ...string) {
	t.Helper()
	dir := filepath.Join(base, pluginName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), nil, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
}

func TestDiscoverPlugins_GlobalOnly(t *testing.T) {
	global := t.TempDir()
	makePlugin(t, global, "myplugin", "plugin.glitch", "deploy.glitch", "status.glitch")

	result := DiscoverPlugins("", global)

	if len(result) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result))
	}
	info, ok := result["myplugin"]
	if !ok {
		t.Fatal("expected plugin 'myplugin' not found")
	}
	if info.Source != "global" {
		t.Errorf("expected source 'global', got %q", info.Source)
	}

	subs := append([]string(nil), info.Subcommands...)
	sort.Strings(subs)
	expected := []string{"deploy", "status"}
	if len(subs) != len(expected) {
		t.Fatalf("expected subcommands %v, got %v", expected, subs)
	}
	for i, s := range subs {
		if s != expected[i] {
			t.Errorf("subcommand[%d]: expected %q, got %q", i, expected[i], s)
		}
	}
}

func TestDiscoverPlugins_LocalOverridesGlobal(t *testing.T) {
	global := t.TempDir()
	local := t.TempDir()

	makePlugin(t, global, "shared", "plugin.glitch", "run.glitch")
	makePlugin(t, local, "shared", "plugin.glitch", "run.glitch", "debug.glitch")

	result := DiscoverPlugins(local, global)

	if len(result) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result))
	}
	info := result["shared"]
	if info.Source != "local" {
		t.Errorf("expected source 'local', got %q", info.Source)
	}
	if info.Dir != filepath.Join(local, "shared") {
		t.Errorf("expected dir from local tree, got %q", info.Dir)
	}

	subs := append([]string(nil), info.Subcommands...)
	sort.Strings(subs)
	expected := []string{"debug", "run"}
	if len(subs) != len(expected) {
		t.Fatalf("expected subcommands %v, got %v", expected, subs)
	}
	for i, s := range subs {
		if s != expected[i] {
			t.Errorf("subcommand[%d]: expected %q, got %q", i, expected[i], s)
		}
	}
}

func TestDiscoverPlugins_IgnoresPluginGlitch(t *testing.T) {
	global := t.TempDir()
	// Only plugin.glitch present — should yield an empty subcommands slice,
	// but the plugin directory itself is still discovered (has a .glitch file).
	makePlugin(t, global, "noop", "plugin.glitch")

	result := DiscoverPlugins("", global)

	info, ok := result["noop"]
	if !ok {
		t.Fatal("expected plugin 'noop' to be discovered")
	}
	if len(info.Subcommands) != 0 {
		t.Errorf("expected no subcommands, got %v", info.Subcommands)
	}
}

func TestDiscoverPlugins_Empty(t *testing.T) {
	dir := t.TempDir()
	result := DiscoverPlugins(dir, dir)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

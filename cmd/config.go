package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/sexpr"
	"github.com/8op-org/gl1tch/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

type Config struct {
	DefaultModel    string
	DefaultProvider string
	EvalThreshold   int
	WorkflowsDir    string
	Tiers           []provider.TierConfig
	Providers       map[string]ProviderConfig
}

// ProviderConfig defines a named LLM provider endpoint.
type ProviderConfig struct {
	Type         string
	BaseURL      string
	APIKeyEnv    string
	APIKey       string
	DefaultModel string
}

// ResolveAPIKey returns the API key, checking the environment variable first.
func (pc *ProviderConfig) ResolveAPIKey() (string, error) {
	if pc.APIKeyEnv != "" {
		if v := os.Getenv(pc.APIKeyEnv); v != "" {
			return v, nil
		}
	}
	if pc.APIKey != "" {
		return pc.APIKey, nil
	}
	return "", fmt.Errorf("no API key: set %s env var or api_key in config", pc.APIKeyEnv)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		fmt.Print(string(serializeConfig(cfg)))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			cfg = &Config{}
		}
		switch args[0] {
		case "default_model", "default-model":
			cfg.DefaultModel = args[1]
		case "default_provider", "default-provider":
			cfg.DefaultProvider = args[1]
		case "workflows_dir", "workflows-dir":
			cfg.WorkflowsDir = args[1]
		default:
			return fmt.Errorf("unknown config key: %s", args[0])
		}
		return saveConfig(cfg)
	},
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "glitch", "config.glitch")
}

func legacyYAMLConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "glitch", "config.yaml")
}

var legacyYAMLWarnOnce sync.Once

func defaultConfig() *Config {
	return &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
		Tiers:           provider.DefaultTiers(),
		EvalThreshold:   4,
	}
}

// loadConfigFrom reads and parses an s-expression config file at path.
// Returns defaults when the file is missing.
func loadConfigFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig(), nil
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig()
	// Track whether tiers/providers were explicitly set so defaults don't override.
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "config" {
			continue
		}
		kids := n.Children[1:]
		i := 0
		for i < len(kids) {
			c := kids[i]
			if c.IsAtom() && c.Atom != nil && c.Atom.Type == sexpr.TokenKeyword {
				key := c.KeywordVal()
				if i+1 >= len(kids) {
					return nil, fmt.Errorf("line %d: keyword :%s missing value", c.Line, key)
				}
				val := kids[i+1]
				switch key {
				case "default-model":
					cfg.DefaultModel = val.StringVal()
				case "default-provider":
					cfg.DefaultProvider = val.StringVal()
				case "eval-threshold":
					if nv, ok := atomNumber(val); ok {
						cfg.EvalThreshold = nv
					}
				case "workflows-dir":
					cfg.WorkflowsDir = val.StringVal()
				}
				i += 2
				continue
			}
			if c.IsList() && len(c.Children) > 0 {
				head := c.Children[0].SymbolVal()
				switch head {
				case "providers":
					providers, err := parseProviders(c)
					if err != nil {
						return nil, err
					}
					cfg.Providers = providers
				case "tiers":
					tiers, err := parseTiers(c)
					if err != nil {
						return nil, err
					}
					cfg.Tiers = tiers
				}
				i++
				continue
			}
			i++
		}
	}
	return cfg, nil
}

// atomNumber extracts an integer from an atom regardless of whether it
// was lexed as a symbol (e.g., `4`) or a string (e.g., `"4"`).
func atomNumber(n *sexpr.Node) (int, bool) {
	if n == nil || !n.IsAtom() || n.Atom == nil {
		return 0, false
	}
	v := n.Atom.Val
	if n, err := strconv.Atoi(v); err == nil {
		return n, true
	}
	return 0, false
}

func parseProviders(list *sexpr.Node) (map[string]ProviderConfig, error) {
	out := make(map[string]ProviderConfig)
	if list == nil {
		return out, nil
	}
	// Children[0] is "providers" symbol; rest are (provider ...) forms.
	for _, entry := range list.Children[1:] {
		if !entry.IsList() || len(entry.Children) < 2 {
			return nil, fmt.Errorf("line %d: expected (provider \"name\" ...)", entry.Line)
		}
		if entry.Children[0].SymbolVal() != "provider" {
			return nil, fmt.Errorf("line %d: expected provider form, got %s", entry.Line, entry.Children[0].SymbolVal())
		}
		name := entry.Children[1].StringVal()
		if name == "" {
			return nil, fmt.Errorf("line %d: provider name must be a string", entry.Line)
		}
		pc := ProviderConfig{}
		rest := entry.Children[2:]
		for j := 0; j < len(rest); j++ {
			c := rest[j]
			if !c.IsAtom() || c.Atom == nil || c.Atom.Type != sexpr.TokenKeyword {
				return nil, fmt.Errorf("line %d: expected keyword in provider %q", c.Line, name)
			}
			if j+1 >= len(rest) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", c.Line, c.KeywordVal())
			}
			val := rest[j+1]
			switch c.KeywordVal() {
			case "type":
				pc.Type = val.StringVal()
			case "base-url":
				pc.BaseURL = val.StringVal()
			case "api-key-env":
				pc.APIKeyEnv = val.StringVal()
			case "api-key":
				pc.APIKey = val.StringVal()
			case "default-model":
				pc.DefaultModel = val.StringVal()
			}
			j++
		}
		out[name] = pc
	}
	return out, nil
}

func parseTiers(list *sexpr.Node) ([]provider.TierConfig, error) {
	var out []provider.TierConfig
	if list == nil {
		return out, nil
	}
	for _, entry := range list.Children[1:] {
		if !entry.IsList() || len(entry.Children) == 0 {
			return nil, fmt.Errorf("line %d: expected (tier ...)", entry.Line)
		}
		if entry.Children[0].SymbolVal() != "tier" {
			return nil, fmt.Errorf("line %d: expected tier form, got %s", entry.Line, entry.Children[0].SymbolVal())
		}
		t := provider.TierConfig{}
		rest := entry.Children[1:]
		for j := 0; j < len(rest); j++ {
			c := rest[j]
			if !c.IsAtom() || c.Atom == nil || c.Atom.Type != sexpr.TokenKeyword {
				return nil, fmt.Errorf("line %d: expected keyword in tier", c.Line)
			}
			if j+1 >= len(rest) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", c.Line, c.KeywordVal())
			}
			val := rest[j+1]
			switch c.KeywordVal() {
			case "providers":
				if !val.IsList() {
					return nil, fmt.Errorf("line %d: :providers value must be a list", val.Line)
				}
				names := make([]string, 0, len(val.Children))
				for _, p := range val.Children {
					names = append(names, p.StringVal())
				}
				t.Providers = names
			case "model":
				t.Model = val.StringVal()
			}
			j++
		}
		out = append(out, t)
	}
	return out, nil
}

func loadConfig() (*Config, error) {
	if mergedConfig != nil {
		return mergedConfig, nil
	}
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, yerr := os.Stat(legacyYAMLConfigPath()); yerr == nil {
			legacyYAMLWarnOnce.Do(func() {
				fmt.Fprintln(os.Stderr, "glitch: ~/.config/glitch/config.yaml is no longer read — see https://gl1tch.dev/docs/config — continuing with defaults")
			})
		}
	}
	return loadConfigFrom(path)
}

// LoadConfigForGUI is an exported wrapper for packages that cannot import cmd
// directly (e.g., the GUI avoiding an import cycle). It is intentionally
// permissive — returns zero-value *Config on any unexpected error rather than
// failing the caller.
func LoadConfigForGUI() *Config {
	cfg, err := loadConfig()
	if err != nil || cfg == nil {
		return &Config{}
	}
	return cfg
}

// BuildProviderResolver creates a ResolverFunc from the config's Providers map.
func (cfg *Config) BuildProviderResolver() provider.ResolverFunc {
	return func(name string) (provider.ProviderFunc, bool) {
		pc, ok := cfg.Providers[name]
		if !ok || pc.Type != "openai-compatible" {
			return nil, false
		}
		key, err := pc.ResolveAPIKey()
		if err != nil {
			return nil, false
		}
		p := &provider.OpenAICompatibleProvider{
			Name:         name,
			BaseURL:      pc.BaseURL,
			APIKey:       key,
			DefaultModel: pc.DefaultModel,
		}
		return p.Chat, true
	}
}

// ApplyWorkspace overlays workspace defaults onto the config.
// Call after loadConfig, before CLI flag overrides.
func ApplyWorkspace(ws *workspace.Workspace, cfg *Config) {
	if ws == nil {
		return
	}
	if ws.Defaults.Model != "" {
		cfg.DefaultModel = ws.Defaults.Model
	}
	if ws.Defaults.Provider != "" {
		cfg.DefaultProvider = ws.Defaults.Provider
	}
}

func saveConfig(cfg *Config) error {
	return saveConfigAt(cfg, configPath())
}

func saveConfigAt(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, serializeConfig(cfg), 0o644)
}

// serializeConfig renders a Config as an s-expression.
// Note: APIKey is deliberately omitted from output to avoid persisting secrets.
func serializeConfig(cfg *Config) []byte {
	var b strings.Builder
	b.WriteString("(config")
	if cfg.DefaultModel != "" {
		fmt.Fprintf(&b, "\n  :default-model %q", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "" {
		fmt.Fprintf(&b, "\n  :default-provider %q", cfg.DefaultProvider)
	}
	if cfg.EvalThreshold != 0 {
		fmt.Fprintf(&b, "\n  :eval-threshold %d", cfg.EvalThreshold)
	}
	if cfg.WorkflowsDir != "" {
		fmt.Fprintf(&b, "\n  :workflows-dir %q", cfg.WorkflowsDir)
	}
	if len(cfg.Providers) > 0 {
		b.WriteString("\n  (providers")
		names := make([]string, 0, len(cfg.Providers))
		for n := range cfg.Providers {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, name := range names {
			p := cfg.Providers[name]
			fmt.Fprintf(&b, "\n    (provider %q :type %q", name, p.Type)
			if p.BaseURL != "" {
				fmt.Fprintf(&b, " :base-url %q", p.BaseURL)
			}
			if p.APIKeyEnv != "" {
				fmt.Fprintf(&b, " :api-key-env %q", p.APIKeyEnv)
			}
			if p.DefaultModel != "" {
				fmt.Fprintf(&b, " :default-model %q", p.DefaultModel)
			}
			b.WriteString(")")
		}
		b.WriteString(")")
	}
	if len(cfg.Tiers) > 0 {
		b.WriteString("\n  (tiers")
		for _, t := range cfg.Tiers {
			b.WriteString("\n    (tier :providers (")
			for i, p := range t.Providers {
				if i > 0 {
					b.WriteString(" ")
				}
				fmt.Fprintf(&b, "%q", p)
			}
			b.WriteString(")")
			if t.Model != "" {
				fmt.Fprintf(&b, " :model %q", t.Model)
			}
			b.WriteString(")")
		}
		b.WriteString(")")
	}
	b.WriteString(")\n")
	return []byte(b.String())
}

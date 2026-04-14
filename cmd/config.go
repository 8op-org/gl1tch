package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

type Config struct {
	DefaultModel    string                    `yaml:"default_model"`
	DefaultProvider string                    `yaml:"default_provider"`
	EvalThreshold   int                       `yaml:"eval_threshold,omitempty"`
	Tiers           []provider.TierConfig     `yaml:"tiers,omitempty"`
	Providers       map[string]ProviderConfig `yaml:"providers,omitempty"`
}

// ProviderConfig defines a named LLM provider endpoint.
type ProviderConfig struct {
	Type         string `yaml:"type"`
	BaseURL      string `yaml:"base_url"`
	APIKeyEnv    string `yaml:"api_key_env,omitempty"`
	APIKey       string `yaml:"api_key,omitempty"`
	DefaultModel string `yaml:"default_model,omitempty"`
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
		data, _ := yaml.Marshal(cfg)
		fmt.Print(string(data))
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
		case "default_model":
			cfg.DefaultModel = args[1]
		case "default_provider":
			cfg.DefaultProvider = args[1]
		default:
			return fmt.Errorf("unknown config key: %s", args[0])
		}
		return saveConfig(cfg)
	},
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "glitch", "config.yaml")
}

func loadConfigFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{
			DefaultModel:    "qwen3:8b",
			DefaultProvider: "ollama",
			Tiers:           provider.DefaultTiers(),
		}, nil
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Tiers) == 0 {
		cfg.Tiers = provider.DefaultTiers()
	}
	if cfg.EvalThreshold == 0 {
		cfg.EvalThreshold = 4
	}
	return &cfg, nil
}

func loadConfig() (*Config, error) {
	return loadConfigFrom(configPath())
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

func saveConfig(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

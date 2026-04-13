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
	DefaultModel    string                `yaml:"default_model"`
	DefaultProvider string                `yaml:"default_provider"`
	Tiers           []provider.TierConfig `yaml:"tiers,omitempty"`
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

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
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
	return &cfg, nil
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

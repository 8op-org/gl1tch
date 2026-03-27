package store

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RetentionConfig holds pruning settings for the store.
type RetentionConfig struct {
	MaxAgeDays int `yaml:"max_age_days"`
	MaxRows    int `yaml:"max_rows"`
}

// configFile mirrors the structure of ~/.config/orcai/config.yaml.
type configFile struct {
	Store struct {
		Retention RetentionConfig `yaml:"retention"`
	} `yaml:"store"`
}

// defaultRetention is used when the config file is missing or the key is absent.
var defaultRetention = RetentionConfig{
	MaxAgeDays: 30,
	MaxRows:    1000,
}

// loadRetentionConfig reads ~/.config/orcai/config.yaml and extracts store.retention.
// Returns defaults (30 days, 1000 rows) if the file is missing or the key is absent.
func loadRetentionConfig() RetentionConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultRetention
	}

	cfgPath := filepath.Join(home, ".config", "orcai", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		// File missing or unreadable — use defaults.
		return defaultRetention
	}

	var cfg configFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultRetention
	}

	r := cfg.Store.Retention

	// Fill in zero values with defaults.
	if r.MaxAgeDays == 0 {
		r.MaxAgeDays = defaultRetention.MaxAgeDays
	}
	if r.MaxRows == 0 {
		r.MaxRows = defaultRetention.MaxRows
	}

	return r
}

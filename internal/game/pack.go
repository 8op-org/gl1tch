package game

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PackWeights holds the tuneable coefficients for the XP formula.
// DefaultPackWeights reproduces the original hard-coded formula exactly.
type PackWeights struct {
	BaseMultiplier  float64            `yaml:"base_multiplier"`
	CacheBonusRate  float64            `yaml:"cache_bonus_rate"`
	SpeedBonusCap   int64              `yaml:"speed_bonus_cap"`
	SpeedBonusScale float64            `yaml:"speed_bonus_scale"`
	RetryPenalty    int64              `yaml:"retry_penalty"`
	StreakMultiplier float64            `yaml:"streak_multiplier"`
	ProviderWeights map[string]float64 `yaml:"provider_weights"`
}

// DefaultPackWeights returns coefficients that reproduce the original
// ComputeXP formula exactly:
//
//	base = outputTokens * (output/input ratio) * 10
//	cacheBonus = cacheReadTokens / 2
//	speedBonus = max(0, 1000 - durationMS/100)
//	retryPenalty = retryCount * 50
func DefaultPackWeights() PackWeights {
	return PackWeights{
		BaseMultiplier:  10.0,
		CacheBonusRate:  0.5,
		SpeedBonusCap:   1000,
		SpeedBonusScale: 0.01,
		RetryPenalty:    50,
		StreakMultiplier: 1.0,
		ProviderWeights: map[string]float64{},
	}
}

// GameWorldPack holds the prompts that drive the game engine.
type GameWorldPack struct {
	Name          string
	GameRules     string
	NarratorStyle string
	Weights       PackWeights
}

// WorldPackLoader resolves the active game world pack.
type WorldPackLoader interface {
	ActivePack() GameWorldPack
}

//go:embed packs/cyberspace/pack.yaml
var defaultPackData []byte

// DefaultWorldPackLoader returns the embedded cyberspace pack.
type DefaultWorldPackLoader struct{}

// ActivePack parses and returns the embedded default pack.
func (DefaultWorldPackLoader) ActivePack() GameWorldPack {
	var raw struct {
		Name          string      `yaml:"name"`
		GameRules     string      `yaml:"game_rules"`
		NarratorStyle string      `yaml:"narrator_style"`
		Weights       PackWeights `yaml:"weights"`
	}
	if err := yaml.Unmarshal(defaultPackData, &raw); err != nil {
		return GameWorldPack{Name: "default", Weights: DefaultPackWeights()}
	}
	w := raw.Weights
	if w.BaseMultiplier == 0 {
		w = DefaultPackWeights()
	} else if w.ProviderWeights == nil {
		w.ProviderWeights = map[string]float64{}
	}
	return GameWorldPack{
		Name:          raw.Name,
		GameRules:     raw.GameRules,
		NarratorStyle: raw.NarratorStyle,
		Weights:       w,
	}
}

// APMWorldPackLoader scans the APM agent directory for a kind: game-world
// agent. Falls back to DefaultWorldPackLoader if none is found.
type APMWorldPackLoader struct{}

// ActivePack returns the first installed game-world pack, or the default.
func (APMWorldPackLoader) ActivePack() GameWorldPack {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultWorldPackLoader{}.ActivePack()
	}
	agentDir := filepath.Join(home, ".local", "share", "glitch", "agents")
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return DefaultWorldPackLoader{}.ActivePack()
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".agent.md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(agentDir, entry.Name()))
		if err != nil {
			continue
		}
		pack, ok := parseGameWorldAgent(data)
		if ok {
			return pack
		}
	}
	return DefaultWorldPackLoader{}.ActivePack()
}

// parseGameWorldAgent extracts a GameWorldPack from an .agent.md file if its
// YAML frontmatter contains "kind: game-world".
func parseGameWorldAgent(data []byte) (GameWorldPack, bool) {
	content := string(data)
	// Strip leading BOM or whitespace.
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return GameWorldPack{}, false
	}
	// Extract frontmatter between first and second "---".
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return GameWorldPack{}, false
	}
	frontmatter := rest[:idx]

	var fm struct {
		Kind          string      `yaml:"kind"`
		Name          string      `yaml:"name"`
		GameRules     string      `yaml:"game_rules"`
		NarratorStyle string      `yaml:"narrator_style"`
		Weights       PackWeights `yaml:"weights"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return GameWorldPack{}, false
	}
	if fm.Kind != "game-world" {
		return GameWorldPack{}, false
	}
	w := fm.Weights
	if w.BaseMultiplier == 0 {
		w = DefaultPackWeights()
	} else if w.ProviderWeights == nil {
		w.ProviderWeights = map[string]float64{}
	}
	return GameWorldPack{
		Name:          fm.Name,
		GameRules:     fm.GameRules,
		NarratorStyle: fm.NarratorStyle,
		Weights:       w,
	}, true
}

package game

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultPackLoader_MUDXPEvents(t *testing.T) {
	pack := DefaultWorldPackLoader{}.ActivePack()
	if len(pack.Weights.MUDXPEvents) == 0 {
		t.Error("expected MUDXPEvents to be non-empty in default pack")
	}
	if pack.Weights.MUDXPEvents["mud.room.entered"] == 0 {
		t.Error("expected mud.room.entered to be in MUDXPEvents")
	}
	if pack.Weights.MUDXPEvents["mud.espionage.talked"] == 0 {
		t.Error("expected mud.espionage.talked to be in MUDXPEvents")
	}
	if pack.Weights.MUDXPEvents["mud.hack.success"] == 0 {
		t.Error("expected mud.hack.success to be in MUDXPEvents")
	}
}

func TestPackLoader_BountyContractsField(t *testing.T) {
	rawYAML := `
name: test-pack
game_rules: |
  You are a game engine.
  {"achievements": [], "ice_class": null, "quest_events": []}
  - ach1: output_tokens > 0
  - ach2: cache_read_tokens >= 100
  - ach3: cost_usd == 0
  - ach4: streak_days >= 3
  - ach5: total_runs >= 10
narrator_style: |
  Terse narrator.
weights:
  base_multiplier: 10.0
  cache_bonus_rate: 0.5
  speed_bonus_cap: 1000
  speed_bonus_scale: 0.01
  retry_penalty: 50
  streak_multiplier: 1.0
  provider_weights: {}
  mud_xp_events:
    mud.room.entered: 15
bounty_contracts:
  - id: test-bounty
    description: Hit a cache ratio
    objective_type: cache_ratio
    objective_value: 0.5
    xp_reward: 300
    room_id: bazaar
    valid_until: "2099-01-01T00:00:00Z"
reputation_decay:
  decay_per_day: 3
  floor: 5
  max_decay_days: 10
ice_encounter:
  timeout_hours: 12
`
	var raw struct {
		Name             string                `yaml:"name"`
		GameRules        string                `yaml:"game_rules"`
		NarratorStyle    string                `yaml:"narrator_style"`
		Weights          PackWeights           `yaml:"weights"`
		BountyContracts  []BountyContract      `yaml:"bounty_contracts"`
		ReputationDecay  ReputationDecayConfig `yaml:"reputation_decay"`
		ICEEncounter     ICEEncounterConfig    `yaml:"ice_encounter"`
	}
	if err := yaml.Unmarshal([]byte(rawYAML), &raw); err != nil {
		t.Fatalf("yaml unmarshal: %v", err)
	}

	if raw.Weights.MUDXPEvents["mud.room.entered"] != 15 {
		t.Errorf("MUDXPEvents[mud.room.entered] = %d, want 15", raw.Weights.MUDXPEvents["mud.room.entered"])
	}
	if len(raw.BountyContracts) != 1 {
		t.Fatalf("expected 1 bounty contract, got %d", len(raw.BountyContracts))
	}
	if raw.BountyContracts[0].ID != "test-bounty" {
		t.Errorf("contract ID = %q, want test-bounty", raw.BountyContracts[0].ID)
	}
	if raw.BountyContracts[0].XPReward != 300 {
		t.Errorf("contract XPReward = %d, want 300", raw.BountyContracts[0].XPReward)
	}
	if raw.ReputationDecay.DecayPerDay != 3 {
		t.Errorf("ReputationDecay.DecayPerDay = %d, want 3", raw.ReputationDecay.DecayPerDay)
	}
	if raw.ICEEncounter.TimeoutHours != 12 {
		t.Errorf("ICEEncounter.TimeoutHours = %d, want 12", raw.ICEEncounter.TimeoutHours)
	}
}

func TestTunerValidation_AcceptsBountyContracts(t *testing.T) {
	tuner := &Tuner{}

	// YAML with 5 achievements and valid bounty contracts.
	validYAML := []byte(`
name: test
kind: game-world
game_rules: |
  Rules here.
  - ach1: output_tokens > 0
  - ach2: cache_read_tokens >= 100
  - ach3: cost_usd == 0
  - ach4: streak_days >= 3
  - ach5: total_runs >= 10
narrator_style: |
  Narrator.
weights:
  base_multiplier: 1.0
  streak_multiplier: 1.0
  provider_weights: {}
bounty_contracts:
  - id: blitz-cache
    xp_reward: 500
    room_id: mainframe
  - id: speed-run
    xp_reward: 250
    room_id: bazaar
`)
	if err := tuner.validate(validYAML); err != nil {
		t.Errorf("validate with valid bounty contracts: unexpected error: %v", err)
	}

	// YAML with a bounty contract missing xp_reward (zero value).
	invalidYAML := []byte(`
name: test
kind: game-world
game_rules: |
  Rules here.
  - ach1: output_tokens > 0
  - ach2: cache_read_tokens >= 100
  - ach3: cost_usd == 0
  - ach4: streak_days >= 3
  - ach5: total_runs >= 10
narrator_style: |
  Narrator.
weights:
  base_multiplier: 1.0
  streak_multiplier: 1.0
  provider_weights: {}
bounty_contracts:
  - id: bad-contract
    xp_reward: 0
    room_id: mainframe
`)
	if err := tuner.validate(invalidYAML); err == nil {
		t.Error("validate with zero xp_reward: expected error, got nil")
	}
}

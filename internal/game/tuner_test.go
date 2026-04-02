package game

import (
	"testing"
	"time"

	"github.com/8op-org/gl1tch/internal/store"
)

// ── ShouldTune ────────────────────────────────────────────────────────────────

var zeroState = TunerState{}

func newTunerForTest() *Tuner {
	return &Tuner{}
}

func TestShouldTune_LevelUp(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	if !tuner.ShouldTune(now, zeroState, 1, 2, 0, 0, nil) {
		t.Error("level-up should trigger tune")
	}
}

func TestShouldTune_StreakMilestone(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	for _, m := range []int{3, 7, 14, 30, 60, 90} {
		if !tuner.ShouldTune(now, zeroState, 1, 1, m-1, m, nil) {
			t.Errorf("streak milestone %d should trigger tune", m)
		}
	}
}

func TestShouldTune_NewAchievement(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	if !tuner.ShouldTune(now, zeroState, 1, 1, 2, 2, []string{"ghost-runner"}) {
		t.Error("new achievement should trigger tune")
	}
}

func TestShouldTune_7DayFloor(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	// 8 days since last tune, 5 runs
	state := TunerState{
		LastTunedAt:   now.AddDate(0, 0, -8),
		RunsSinceTune: 5,
	}
	if !tuner.ShouldTune(now, state, 1, 1, 0, 0, nil) {
		t.Error("7-day floor with 5 runs should trigger tune")
	}
}

func TestShouldTune_7DayFloor_NotEnoughRuns(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	state := TunerState{
		LastTunedAt:   now.AddDate(0, 0, -8),
		RunsSinceTune: 4, // only 4, need 5
	}
	if tuner.ShouldTune(now, state, 1, 1, 0, 0, nil) {
		t.Error("7-day floor with only 4 runs should NOT trigger tune")
	}
}

func TestShouldTune_Cooldown(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	// Tuned today — cooldown active.
	state := TunerState{
		LastTunedAt:   now.Add(-1 * time.Hour), // 1 hour ago, same day
		RunsSinceTune: 0,
	}
	// Level-up would normally trigger, but cooldown suppresses it.
	if tuner.ShouldTune(now, state, 1, 2, 0, 0, nil) {
		t.Error("cooldown should suppress tune even when level-up occurs")
	}
}

func TestShouldTune_CooldownExpired(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	// Tuned yesterday — cooldown expired.
	state := TunerState{
		LastTunedAt:   now.AddDate(0, 0, -1).Add(-1 * time.Hour),
		RunsSinceTune: 0,
	}
	if !tuner.ShouldTune(now, state, 1, 2, 0, 0, nil) {
		t.Error("cooldown from yesterday should allow tune today on level-up")
	}
}

func TestShouldTune_NoCondition(t *testing.T) {
	tuner := newTunerForTest()
	now := time.Now()
	if tuner.ShouldTune(now, zeroState, 1, 1, 2, 2, nil) {
		t.Error("no trigger condition should not tune")
	}
}

// ── validate ──────────────────────────────────────────────────────────────────

func TestValidate_ValidPack(t *testing.T) {
	tuner := newTunerForTest()
	yaml := `kind: game-world
name: gl1tch-world-cyberspace-tuned
game_rules: |
  Achievement definitions:
  - ghost-runner: output_tokens > 0
  - cache-warlock: cache_read_tokens >= 10000
  - speed-demon: duration_ms < 2000
  - token-miser: output_tokens > 0
  - cost-cutter: cost_usd == 0
narrator_style: |
  You are Zero Cool.
weights:
  base_multiplier: 10.0
  cache_bonus_rate: 0.5
  speed_bonus_cap: 1000
  speed_bonus_scale: 0.01
  retry_penalty: 50
  streak_multiplier: 1.0
  provider_weights: {}`
	if err := tuner.validate([]byte(yaml)); err != nil {
		t.Errorf("valid pack should pass validation: %v", err)
	}
}

func TestValidate_TooFewAchievements(t *testing.T) {
	tuner := newTunerForTest()
	yaml := `kind: game-world
name: gl1tch-world-cyberspace-tuned
game_rules: |
  - ghost-runner: output_tokens > 0
  - cache-warlock: cache_read_tokens >= 10000
narrator_style: |
  You are Zero Cool.
weights:
  base_multiplier: 10.0
  cache_bonus_rate: 0.5
  speed_bonus_cap: 1000
  speed_bonus_scale: 0.01
  retry_penalty: 50
  streak_multiplier: 1.0
  provider_weights: {}`
	if err := tuner.validate([]byte(yaml)); err == nil {
		t.Error("pack with <5 achievements should fail validation")
	}
}

func TestValidate_WeightOutOfRange(t *testing.T) {
	tuner := newTunerForTest()
	yaml := `kind: game-world
name: gl1tch-world-cyberspace-tuned
game_rules: |
  - ghost-runner: output_tokens > 0
  - cache-warlock: cache_read_tokens >= 10000
  - speed-demon: duration_ms < 2000
  - token-miser: output_tokens > 0
  - cost-cutter: cost_usd == 0
narrator_style: |
  You are Zero Cool.
weights:
  base_multiplier: 10.0
  cache_bonus_rate: 0.5
  speed_bonus_cap: 1000
  speed_bonus_scale: 0.01
  retry_penalty: 50
  streak_multiplier: 99.0
  provider_weights: {}`
	if err := tuner.validate([]byte(yaml)); err == nil {
		t.Error("streak_multiplier out of range should fail validation")
	}
}

func TestValidate_InvalidYAML(t *testing.T) {
	tuner := newTunerForTest()
	if err := tuner.validate([]byte("{not valid yaml: [}")); err == nil {
		t.Error("invalid YAML should fail validation")
	}
}

// ── GameStats stub to ensure type compatibility ───────────────────────────────

func TestGameStats_TypeCompat(t *testing.T) {
	// Ensure store.GameStats is accessible from tuner code.
	var gs store.GameStats
	gs.TotalRuns = 5
	if gs.TotalRuns != 5 {
		t.Error("GameStats type mismatch")
	}
}

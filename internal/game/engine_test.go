package game

import (
	"testing"
	"time"

	"github.com/8op-org/gl1tch/internal/store"
)

func TestComputeXP(t *testing.T) {
	tests := []struct {
		name       string
		usage      TokenUsage
		retryCount int
		wantFinal  int64
	}{
		{
			name: "zero output tokens yields zero base",
			usage: TokenUsage{
				InputTokens:  1000,
				OutputTokens: 0,
				DurationMS:   500,
			},
			wantFinal: 995, // only speed bonus (1000 - 500/100 = 995)
		},
		{
			name: "basic efficiency",
			usage: TokenUsage{
				InputTokens:  100,
				OutputTokens: 50,
				DurationMS:   500,
			},
			// ratio = 50/100 = 0.5; base = 50 * 0.5 * 10 = 250
			// speed = 1000 - 500/100 = 995
			// final = 250 + 0 + 995 = 1245
			wantFinal: 1245,
		},
		{
			name: "cache bonus included",
			usage: TokenUsage{
				InputTokens:     100,
				OutputTokens:    50,
				CacheReadTokens: 200,
				DurationMS:      500,
			},
			// base=250, cache=100, speed=995 → 1345
			wantFinal: 1345,
		},
		{
			name: "retry penalty applied",
			usage: TokenUsage{
				InputTokens:  100,
				OutputTokens: 50,
				DurationMS:   500,
			},
			retryCount: 2,
			// base=250, speed=995, penalty=100 → 1145
			wantFinal: 1145,
		},
		{
			name: "final clamped to zero",
			usage: TokenUsage{
				InputTokens:  100,
				OutputTokens: 1,
				DurationMS:   99000, // speed = 1000 - 990 = 10
			},
			retryCount: 100,
			// base tiny, speed tiny, penalty=5000 → clamped 0
			wantFinal: 0,
		},
		{
			name: "slow run, no speed bonus",
			usage: TokenUsage{
				InputTokens:  100,
				OutputTokens: 50,
				DurationMS:   100000, // speedVal = 1000 - 1000 = 0
			},
			// base=250, cache=0, speed=0, penalty=0
			wantFinal: 250,
		},
		{
			name: "cache creation counts in denominator",
			usage: TokenUsage{
				InputTokens:         0,
				CacheCreationTokens: 100,
				OutputTokens:        50,
				DurationMS:          0,
			},
			// ratio = 50/100 = 0.5; base = 50 * 0.5 * 10 = 250
			// speed = 1000 - 0 = 1000
			wantFinal: 1250,
		},
	}

	dw := DefaultPackWeights()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeXP(tc.usage, tc.retryCount, dw)
			if got.Final != tc.wantFinal {
				t.Errorf("Final = %d, want %d (base=%d cache=%d speed=%d penalty=%d)",
					got.Final, tc.wantFinal,
					got.Base, got.CacheBonus, got.SpeedBonus, got.RetryPenalty)
			}
		})
	}
}

func TestComputeXP_ProviderMultiplier(t *testing.T) {
	w := DefaultPackWeights()
	w.ProviderWeights["providers.claude"] = 1.5
	usage := TokenUsage{
		Provider:     "providers.claude",
		InputTokens:  100,
		OutputTokens: 50,
		DurationMS:   500,
	}
	base := ComputeXP(usage, 0, DefaultPackWeights())
	withMult := ComputeXP(usage, 0, w)
	if withMult.Final != int64(float64(base.Final)*1.5) {
		t.Errorf("provider multiplier: got %d, want %d", withMult.Final, int64(float64(base.Final)*1.5))
	}
}

func TestComputeXP_StreakMultiplier(t *testing.T) {
	w := DefaultPackWeights()
	w.StreakMultiplier = 1.2
	usage := TokenUsage{
		Provider:     "providers.claude",
		InputTokens:  100,
		OutputTokens: 50,
		DurationMS:   500,
		StreakDays:   5,
	}
	base := ComputeXP(usage, 0, DefaultPackWeights())
	withStreak := ComputeXP(usage, 0, w)
	if withStreak.Final != int64(float64(base.Final)*1.2) {
		t.Errorf("streak multiplier: got %d, want %d", withStreak.Final, int64(float64(base.Final)*1.2))
	}
}

func TestComputeXP_StreakMultiplierNotAppliedWhenOne(t *testing.T) {
	w := DefaultPackWeights() // StreakMultiplier == 1.0
	usage := TokenUsage{
		InputTokens:  100,
		OutputTokens: 50,
		DurationMS:   500,
		StreakDays:   10,
	}
	base := ComputeXP(usage, 0, DefaultPackWeights())
	withStreak := ComputeXP(usage, 0, w)
	if withStreak.Final != base.Final {
		t.Errorf("streak=1.0 should be no-op: base=%d withStreak=%d", base.Final, withStreak.Final)
	}
}

func TestLevelForXP(t *testing.T) {
	tests := []struct {
		xp          int64
		wantLevel   int
		wantTitle   string
		wantNextXP  int64
	}{
		{0, 1, "Apprentice of the Shell", 500},
		{499, 1, "Apprentice of the Shell", 500},
		{500, 2, "Journeyman Prompter", 1500},
		{1500, 3, "Adept of the Token", 3000},
		{150000, 13, "The Tokenless One", 150000}, // max level stays at own XP
		{200000, 13, "The Tokenless One", 150000},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			level, title, nextXP := LevelForXP(tc.xp)
			if level != tc.wantLevel {
				t.Errorf("XP=%d: level=%d, want %d", tc.xp, level, tc.wantLevel)
			}
			if title != tc.wantTitle {
				t.Errorf("XP=%d: title=%q, want %q", tc.xp, title, tc.wantTitle)
			}
			if nextXP != tc.wantNextXP {
				t.Errorf("XP=%d: nextLevelXP=%d, want %d", tc.xp, nextXP, tc.wantNextXP)
			}
		})
	}
}

func TestUpdateStreak(t *testing.T) {
	base := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	t.Run("first run", func(t *testing.T) {
		us := store.UserScore{}
		got := UpdateStreak(us, base)
		if got.StreakDays != 1 {
			t.Errorf("StreakDays = %d, want 1", got.StreakDays)
		}
		if got.LastRunDate != "2026-04-01" {
			t.Errorf("LastRunDate = %q, want 2026-04-01", got.LastRunDate)
		}
	})

	t.Run("same day does not increment", func(t *testing.T) {
		us := store.UserScore{StreakDays: 3, LastRunDate: "2026-04-01"}
		got := UpdateStreak(us, base)
		if got.StreakDays != 3 {
			t.Errorf("StreakDays = %d, want 3 (no change same day)", got.StreakDays)
		}
	})

	t.Run("consecutive day increments", func(t *testing.T) {
		us := store.UserScore{StreakDays: 3, LastRunDate: "2026-03-31"}
		got := UpdateStreak(us, base)
		if got.StreakDays != 4 {
			t.Errorf("StreakDays = %d, want 4", got.StreakDays)
		}
	})

	t.Run("gap resets streak", func(t *testing.T) {
		us := store.UserScore{StreakDays: 5, LastRunDate: "2026-03-29"}
		got := UpdateStreak(us, base)
		if got.StreakDays != 1 {
			t.Errorf("StreakDays = %d, want 1 (gap resets)", got.StreakDays)
		}
	})

	t.Run("invalid date resets streak", func(t *testing.T) {
		us := store.UserScore{StreakDays: 5, LastRunDate: "not-a-date"}
		got := UpdateStreak(us, base)
		if got.StreakDays != 1 {
			t.Errorf("StreakDays = %d, want 1 (invalid date resets)", got.StreakDays)
		}
	})
}

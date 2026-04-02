package game

import (
	"time"

	"github.com/8op-org/gl1tch/internal/store"
)

// XPResult holds the breakdown of XP earned in a single run.
type XPResult struct {
	Base         int64
	CacheBonus   int64
	SpeedBonus   int64
	RetryPenalty int64
	Final        int64
}

// ComputeXP calculates XP earned from a run based on token efficiency metrics.
// weights controls all formula coefficients; pass DefaultPackWeights() for
// identical behaviour to the original hard-coded formula.
func ComputeXP(usage TokenUsage, retryCount int, weights PackWeights) XPResult {
	var r XPResult
	denom := usage.InputTokens + usage.CacheCreationTokens
	if denom > 0 && usage.OutputTokens > 0 {
		ratio := float64(usage.OutputTokens) / float64(denom)
		r.Base = int64(float64(usage.OutputTokens) * ratio * weights.BaseMultiplier)
	}
	r.CacheBonus = int64(float64(usage.CacheReadTokens) * weights.CacheBonusRate)
	speedVal := weights.SpeedBonusCap - int64(float64(usage.DurationMS)*weights.SpeedBonusScale)
	if speedVal > 0 {
		r.SpeedBonus = speedVal
	}
	r.RetryPenalty = int64(retryCount) * weights.RetryPenalty
	r.Final = r.Base + r.CacheBonus + r.SpeedBonus - r.RetryPenalty
	if r.Final < 0 {
		r.Final = 0
	}
	// Apply provider multiplier.
	if mult, ok := weights.ProviderWeights[usage.Provider]; ok && mult != 0 {
		r.Final = int64(float64(r.Final) * mult)
	}
	// Apply streak multiplier when streak > 1.
	if usage.StreakDays > 1 && weights.StreakMultiplier != 0 && weights.StreakMultiplier != 1.0 {
		r.Final = int64(float64(r.Final) * weights.StreakMultiplier)
	}
	return r
}

// LevelEntry pairs an XP threshold with a title.
type LevelEntry struct {
	XP    int64
	Title string
}

// levelTable defines the progression ladder.
var levelTable = []LevelEntry{
	{0, "Apprentice of the Shell"},
	{500, "Journeyman Prompter"},
	{1500, "Adept of the Token"},
	{3000, "Conjurer of Context"},
	{5000, "Mage of Efficient Context"},
	{8000, "Sorcerer of the Stream"},
	{12000, "Warlock of the Window"},
	{15000, "Archon of Cache"},
	{22000, "Void Walker"},
	{30000, "Wizard of the Sparse Prompt"},
	{50000, "Elder of the Token"},
	{75000, "Grand Necromancer of Parsimony"},
	{150000, "The Tokenless One"},
}

// LevelForXP resolves the player's current level, title, and next threshold.
func LevelForXP(totalXP int64) (level int, title string, nextLevelXP int64) {
	level = 1
	title = levelTable[0].Title
	nextLevelXP = levelTable[1].XP
	for i, e := range levelTable {
		if totalXP >= e.XP {
			level = i + 1
			title = e.Title
			if i+1 < len(levelTable) {
				nextLevelXP = levelTable[i+1].XP
			} else {
				nextLevelXP = e.XP
			}
		}
	}
	return
}

// UpdateStreak computes the new streak day count for the player.
func UpdateStreak(us store.UserScore, now time.Time) store.UserScore {
	today := now.Format("2006-01-02")
	if us.LastRunDate == "" {
		us.StreakDays = 1
		us.LastRunDate = today
		return us
	}
	if us.LastRunDate == today {
		return us
	}
	last, err := time.Parse("2006-01-02", us.LastRunDate)
	if err != nil {
		us.StreakDays = 1
		us.LastRunDate = today
		return us
	}
	diff := now.Truncate(24*time.Hour).Sub(last.Truncate(24 * time.Hour))
	if diff == 24*time.Hour {
		us.StreakDays++
	} else {
		us.StreakDays = 1
	}
	us.LastRunDate = today
	return us
}

// XPBreakdown is the JSON-serializable form of an XPResult.
type XPBreakdown struct {
	Base         int64 `json:"base"`
	CacheBonus   int64 `json:"cache_bonus"`
	SpeedBonus   int64 `json:"speed_bonus"`
	RetryPenalty int64 `json:"retry_penalty"`
}

// GameRunScoredPayload is published on the game.run.scored BUSD topic.
type GameRunScoredPayload struct {
	XP           int64       `json:"xp"`
	XPBreakdown  XPBreakdown `json:"xp_breakdown"`
	TotalXP      int64       `json:"total_xp"`
	Level        int         `json:"level"`
	LevelTitle   string      `json:"level_title"`
	PrevLevel    int         `json:"prev_level"`
	NextLevelXP  int64       `json:"next_level_xp"`
	StreakDays   int         `json:"streak_days"`
	StepFailures int         `json:"step_failures"`
	Achievements []string    `json:"achievements"`
	ICEClass     string      `json:"ice_class"`
	Usage        TokenUsage  `json:"usage"`
}

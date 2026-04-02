package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestGameStatsQuery_EmptyWindow(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	gs, err := s.GameStatsQuery(ctx, 30)
	if err != nil {
		t.Fatalf("GameStatsQuery on empty db: %v", err)
	}
	if gs.TotalRuns != 0 {
		t.Errorf("TotalRuns = %d, want 0", gs.TotalRuns)
	}
	if gs.AvgOutputRatio != 0 {
		t.Errorf("AvgOutputRatio = %f, want 0", gs.AvgOutputRatio)
	}
	if gs.ProviderRunCounts == nil {
		t.Error("ProviderRunCounts should be non-nil map")
	}
	if len(gs.ProviderRunCounts) != 0 {
		t.Errorf("ProviderRunCounts = %v, want empty", gs.ProviderRunCounts)
	}
	if gs.UnlockedAchievementIDs == nil {
		t.Error("UnlockedAchievementIDs should be non-nil slice")
	}
}

func TestGameStatsQuery_AggregatesOverWindow(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	now := time.Now().UnixMilli()
	old := time.Now().AddDate(0, 0, -60).UnixMilli() // outside 30-day window

	events := []ScoreEvent{
		// Within window
		{RunID: 1, XP: 100, InputTokens: 100, OutputTokens: 50, CacheReadTokens: 200, CostUSD: 0.01, Provider: "providers.claude", CreatedAt: now - 1000},
		{RunID: 2, XP: 200, InputTokens: 200, OutputTokens: 100, CacheReadTokens: 400, CostUSD: 0.02, Provider: "providers.claude", CreatedAt: now - 2000},
		{RunID: 3, XP: 0, InputTokens: 50, OutputTokens: 10, CacheReadTokens: 0, CostUSD: 0.0, Provider: "providers.codex", CreatedAt: now - 3000},
		// Outside window — should not count
		{RunID: 4, XP: 999, InputTokens: 999, OutputTokens: 999, CostUSD: 9.99, Provider: "providers.claude", CreatedAt: old},
	}
	for _, e := range events {
		if err := s.RecordScoreEvent(ctx, e); err != nil {
			t.Fatalf("RecordScoreEvent: %v", err)
		}
	}

	// Record an achievement.
	if err := s.RecordAchievement(ctx, "ghost-runner"); err != nil {
		t.Fatalf("RecordAchievement: %v", err)
	}

	gs, err := s.GameStatsQuery(ctx, 30)
	if err != nil {
		t.Fatalf("GameStatsQuery: %v", err)
	}

	if gs.TotalRuns != 3 {
		t.Errorf("TotalRuns = %d, want 3", gs.TotalRuns)
	}
	if gs.RunsSinceDate != 3 {
		t.Errorf("RunsSinceDate = %d, want 3", gs.RunsSinceDate)
	}
	// StepFailureRate = 1 zero-xp run / 3 total = ~0.333
	const wantFR = 1.0 / 3.0
	if diff := gs.StepFailureRate - wantFR; diff > 0.001 || diff < -0.001 {
		t.Errorf("StepFailureRate = %f, want ~%f", gs.StepFailureRate, wantFR)
	}
	if gs.AvgCostUSD <= 0 {
		t.Errorf("AvgCostUSD = %f, want > 0", gs.AvgCostUSD)
	}
	if len(gs.UnlockedAchievementIDs) != 1 || gs.UnlockedAchievementIDs[0] != "ghost-runner" {
		t.Errorf("UnlockedAchievementIDs = %v, want [ghost-runner]", gs.UnlockedAchievementIDs)
	}
}

func TestGameStatsQuery_ProviderMixPopulated(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	now := time.Now().UnixMilli()

	events := []ScoreEvent{
		{RunID: 1, XP: 10, Provider: "providers.claude", CreatedAt: now - 100},
		{RunID: 2, XP: 10, Provider: "providers.claude", CreatedAt: now - 200},
		{RunID: 3, XP: 10, Provider: "providers.codex", CreatedAt: now - 300},
	}
	for _, e := range events {
		if err := s.RecordScoreEvent(ctx, e); err != nil {
			t.Fatalf("RecordScoreEvent: %v", err)
		}
	}

	gs, err := s.GameStatsQuery(ctx, 30)
	if err != nil {
		t.Fatalf("GameStatsQuery: %v", err)
	}

	if gs.ProviderRunCounts["providers.claude"] != 2 {
		t.Errorf("providers.claude count = %d, want 2", gs.ProviderRunCounts["providers.claude"])
	}
	if gs.ProviderRunCounts["providers.codex"] != 1 {
		t.Errorf("providers.codex count = %d, want 1", gs.ProviderRunCounts["providers.codex"])
	}
}

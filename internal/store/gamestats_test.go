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

func TestInsertAchievement_Idempotent(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	// First insert — should succeed.
	if err := s.RecordAchievement(ctx, "speed-demon"); err != nil {
		t.Fatalf("RecordAchievement first: %v", err)
	}
	// Second insert — should be a no-op, not an error.
	if err := s.RecordAchievement(ctx, "speed-demon"); err != nil {
		t.Fatalf("RecordAchievement second (idempotent): %v", err)
	}

	has, err := s.HasAchievement("speed-demon")
	if err != nil {
		t.Fatalf("HasAchievement: %v", err)
	}
	if !has {
		t.Error("expected HasAchievement to return true after insert")
	}

	// Unknown achievement should return false.
	has2, err := s.HasAchievement("never-inserted")
	if err != nil {
		t.Fatalf("HasAchievement unknown: %v", err)
	}
	if has2 {
		t.Error("expected HasAchievement to return false for unknown ID")
	}
}

func TestPersonalBests_UpdateLogic(t *testing.T) {
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	// First record — always inserted.
	if err := s.InsertOrUpdatePersonalBest("highest_xp", 100, "run-1"); err != nil {
		t.Fatalf("InsertOrUpdatePersonalBest first: %v", err)
	}
	// Better value — should replace.
	if err := s.InsertOrUpdatePersonalBest("highest_xp", 200, "run-2"); err != nil {
		t.Fatalf("InsertOrUpdatePersonalBest better: %v", err)
	}
	// Worse value — should NOT replace.
	if err := s.InsertOrUpdatePersonalBest("highest_xp", 50, "run-3"); err != nil {
		t.Fatalf("InsertOrUpdatePersonalBest worse: %v", err)
	}

	bests, err := s.GetPersonalBests()
	if err != nil {
		t.Fatalf("GetPersonalBests: %v", err)
	}
	if len(bests) != 1 {
		t.Fatalf("expected 1 best, got %d", len(bests))
	}
	if bests[0].Value != 200 {
		t.Errorf("expected value=200 (best), got %f", bests[0].Value)
	}
	if bests[0].RunID != "run-2" {
		t.Errorf("expected run_id=run-2, got %q", bests[0].RunID)
	}

	// Lower-is-better: fastest_run_ms.
	_ = s.InsertOrUpdatePersonalBest("fastest_run_ms", 5000, "run-a")
	_ = s.InsertOrUpdatePersonalBest("fastest_run_ms", 3000, "run-b") // better (lower)
	_ = s.InsertOrUpdatePersonalBest("fastest_run_ms", 4000, "run-c") // worse (higher)

	bests2, _ := s.GetPersonalBests()
	for _, pb := range bests2 {
		if pb.Metric == "fastest_run_ms" {
			if pb.Value != 3000 {
				t.Errorf("fastest_run_ms: expected 3000, got %f", pb.Value)
			}
			if pb.RunID != "run-b" {
				t.Errorf("fastest_run_ms run_id: expected run-b, got %q", pb.RunID)
			}
		}
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

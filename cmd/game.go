package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/game"
	"github.com/8op-org/gl1tch/internal/store"
)

func init() {
	rootCmd.AddCommand(gameCmd)
	gameCmd.AddCommand(gameTuneCmd)
}

var gameCmd = &cobra.Command{
	Use:   "game",
	Short: "Game system commands",
}

var gameTuneCmd = &cobra.Command{
	Use:   "tune",
	Short: "Manually trigger the self-evolving game pack tuner",
	Long: `Calls local Ollama to analyze your usage patterns and generate an evolved
game pack. Writes the result to ~/.local/share/glitch/agents/game-world-tuned.agent.md.
The APMWorldPackLoader picks it up automatically on the next run.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Open()
		if err != nil {
			return fmt.Errorf("game tune: open store: %w", err)
		}
		defer st.Close()

		engine := game.NewGameEngine()
		loader := game.APMWorldPackLoader{}
		tuner := game.NewTuner(st, engine, loader)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Gather stats.
		stats, err := st.GameStatsQuery(ctx, 30)
		if err != nil {
			return fmt.Errorf("game tune: query stats: %w", err)
		}

		// Read the current pack so we can diff what changed.
		currentPack := loader.ActivePack()

		// Dummy payload for the manual tune — we don't have a live run.
		payload := game.GameRunScoredPayload{}

		fmt.Fprintln(os.Stdout, "Running game tuner...")
		fmt.Fprintf(os.Stdout, "  Stats window: last 30 days, %d runs\n", stats.TotalRuns)
		fmt.Fprintf(os.Stdout, "  Current pack: %s\n", currentPack.Name)

		start := time.Now()
		if err := tuner.Tune(ctx, stats, payload); err != nil {
			fmt.Fprintf(os.Stderr, "Tuner failed: %v\n", err)
			os.Exit(1)
		}
		elapsed := time.Since(start).Round(time.Second)

		// Read the new pack and print a summary of what changed.
		newPack := game.APMWorldPackLoader{}.ActivePack()
		printTuneSummary(currentPack, newPack, elapsed)

		return nil
	},
}

// printTuneSummary prints a human-readable diff of what the tuner changed.
func printTuneSummary(old, new_ game.GameWorldPack, elapsed time.Duration) {
	fmt.Printf("\nGame pack evolved in %s\n", elapsed)
	fmt.Printf("  Old pack: %s\n", old.Name)
	fmt.Printf("  New pack: %s\n", new_.Name)

	// Weight deltas.
	oldW := old.Weights
	newW := new_.Weights
	fmt.Println("\nWeight changes:")
	printWeightDelta("base_multiplier", oldW.BaseMultiplier, newW.BaseMultiplier)
	printWeightDelta("cache_bonus_rate", oldW.CacheBonusRate, newW.CacheBonusRate)
	printWeightDelta("speed_bonus_scale", oldW.SpeedBonusScale, newW.SpeedBonusScale)
	printWeightDelta("retry_penalty", float64(oldW.RetryPenalty), float64(newW.RetryPenalty))
	printWeightDelta("streak_multiplier", oldW.StreakMultiplier, newW.StreakMultiplier)

	// Provider weight changes.
	allProviders := map[string]struct{}{}
	for k := range oldW.ProviderWeights {
		allProviders[k] = struct{}{}
	}
	for k := range newW.ProviderWeights {
		allProviders[k] = struct{}{}
	}
	for p := range allProviders {
		oldV := oldW.ProviderWeights[p]
		if oldV == 0 {
			oldV = 1.0
		}
		newV := newW.ProviderWeights[p]
		if newV == 0 {
			newV = 1.0
		}
		printWeightDelta("provider_weights."+p, oldV, newV)
	}

	// Quick game_rules line count diff.
	oldLines := countLines(old.GameRules)
	newLines := countLines(new_.GameRules)
	if newLines != oldLines {
		fmt.Printf("\nGame rules: %d → %d lines (%+d)\n", oldLines, newLines, newLines-oldLines)
	} else {
		fmt.Printf("\nGame rules: %d lines (unchanged length)\n", newLines)
	}

	fmt.Println("\nPack written to ~/.local/share/glitch/agents/game-world-tuned.agent.md")
}

func printWeightDelta(name string, old, new_ float64) {
	if old == new_ {
		return
	}
	diff := new_ - old
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	fmt.Printf("  %s: %.4f → %.4f (%s%.4f)\n", name, old, new_, sign, diff)
}

func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

// tuneStatsJSON returns a compact JSON representation of GameStats for display.
func tuneStatsJSON(gs store.GameStats) string {
	b, _ := json.Marshal(gs)
	return string(b)
}

// suppress unused warning
var _ = tuneStatsJSON

package game

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/8op-org/gl1tch/internal/store"
)

// ApplyMUDReputationDecay checks the last scored run date against today. For
// each full day of inactivity (up to pack.ReputationDecay.MaxDecayDays), it
// reduces every faction's reputation in the MUD's world.db by DecayPerDay,
// bounded at the Floor minimum.
//
// If the MUD database does not exist (plugin not installed), the function
// returns nil without error.
func ApplyMUDReputationDecay(ctx context.Context, st *store.Store, pack GameWorldPack) error {
	cfg := pack.ReputationDecay
	if cfg.DecayPerDay <= 0 {
		cfg = DefaultReputationDecay()
	}

	// Determine the last run date from score_events.
	lastRunAt, err := lastScoreEventTime(ctx, st)
	if err != nil || lastRunAt.IsZero() {
		return nil // no runs yet — nothing to decay
	}

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	lastDay := lastRunAt.Truncate(24 * time.Hour)
	if !lastDay.Before(today) {
		return nil // ran today — no decay
	}

	inactiveDays := int(today.Sub(lastDay).Hours() / 24)
	if inactiveDays > cfg.MaxDecayDays {
		inactiveDays = cfg.MaxDecayDays
	}
	totalDecay := cfg.DecayPerDay * inactiveDays
	if totalDecay <= 0 {
		return nil
	}

	mudDBPath, err := mudWorldDBPath()
	if err != nil {
		return nil
	}
	if _, err := os.Stat(mudDBPath); os.IsNotExist(err) {
		return nil // MUD not installed
	}

	db, err := sql.Open("sqlite", "file:"+mudDBPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(3000)")
	if err != nil {
		return fmt.Errorf("reputation decay: open mud db: %w", err)
	}
	defer db.Close()

	// Fetch all factions, apply decay bounded by floor.
	rows, err := db.QueryContext(ctx, `SELECT faction, value FROM player_reputation`)
	if err != nil {
		return nil // table may not exist yet — skip silently
	}
	defer rows.Close()

	type factionRep struct {
		faction string
		value   int
	}
	var factions []factionRep
	for rows.Next() {
		var fr factionRep
		if err := rows.Scan(&fr.faction, &fr.value); err == nil {
			factions = append(factions, fr)
		}
	}
	_ = rows.Err()

	for _, fr := range factions {
		newVal := fr.value - totalDecay
		if newVal < cfg.Floor {
			newVal = cfg.Floor
		}
		if newVal == fr.value {
			continue
		}
		_, _ = db.ExecContext(ctx,
			`UPDATE player_reputation SET value = ? WHERE faction = ?`,
			newVal, fr.faction,
		)
	}
	return nil
}

// lastScoreEventTime returns the most recent score_event created_at timestamp.
func lastScoreEventTime(ctx context.Context, st *store.Store) (time.Time, error) {
	var ms int64
	err := st.DB().QueryRowContext(ctx, `SELECT MAX(created_at) FROM score_events`).Scan(&ms)
	if err != nil || ms == 0 {
		return time.Time{}, nil
	}
	return time.UnixMilli(ms), nil
}

// mudWorldDBPath returns the standard path for the gl1tch-mud world database.
func mudWorldDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "gl1tch-mud", "world.db"), nil
}

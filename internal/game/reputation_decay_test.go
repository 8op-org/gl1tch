package game

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/8op-org/gl1tch/internal/store"
)

// withMUDDB creates a temporary MUD database and patches HOME so
// mudWorldDBPath() resolves to it. It returns the db path and a no-op cleanup.
// writing a file at the expected path location. Since mudWorldDBPath reads
// os.UserHomeDir(), we override HOME to point at a temp dir containing the DB.
func withMUDDB(t *testing.T, factions map[string]int) (dbPath string, cleanup func()) {
	t.Helper()
	tmpHome := t.TempDir()
	mudDir := filepath.Join(tmpHome, ".local", "share", "gl1tch-mud")
	if err := os.MkdirAll(mudDir, 0o755); err != nil {
		t.Fatalf("mkdir mud dir: %v", err)
	}
	dbPath = filepath.Join(mudDir, "world.db")

	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open test mud db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE player_reputation (faction TEXT PRIMARY KEY, value INTEGER NOT NULL DEFAULT 0)`)
	if err != nil {
		t.Fatalf("create player_reputation: %v", err)
	}
	for faction, val := range factions {
		if _, err := db.Exec(`INSERT INTO player_reputation (faction, value) VALUES (?, ?)`, faction, val); err != nil {
			t.Fatalf("insert faction %s: %v", faction, err)
		}
	}
	t.Setenv("HOME", tmpHome)
	return dbPath, func() {}
}

func readRep(t *testing.T, dbPath, faction string) int {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open mud db for read: %v", err)
	}
	defer db.Close()
	var val int
	db.QueryRow(`SELECT value FROM player_reputation WHERE faction = ?`, faction).Scan(&val) //nolint:errcheck
	return val
}

func TestApplyMUDReputationDecay_OneDay(t *testing.T) {
	st, err := store.OpenAt(filepath.Join(t.TempDir(), "gl.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	// Record a run 2 days ago.
	ctx := context.Background()
	twoDaysAgo := time.Now().AddDate(0, 0, -2).UnixMilli()
	_ = st.RecordScoreEvent(ctx, store.ScoreEvent{XP: 10, Provider: "test", CreatedAt: twoDaysAgo})

	dbPath, _ := withMUDDB(t, map[string]int{"netrunners": 50, "fixers": 30})
	pack := GameWorldPack{
		ReputationDecay: ReputationDecayConfig{DecayPerDay: 5, Floor: 10, MaxDecayDays: 7},
	}

	if err := ApplyMUDReputationDecay(ctx, st, pack); err != nil {
		t.Fatalf("ApplyMUDReputationDecay: %v", err)
	}

	// 2 days × 5 decay = 10 reduction.
	netRep := readRep(t, dbPath, "netrunners")
	if netRep != 40 { // 50 - 10
		t.Errorf("netrunners rep = %d, want 40", netRep)
	}
	fixRep := readRep(t, dbPath, "fixers")
	if fixRep != 20 { // 30 - 10
		t.Errorf("fixers rep = %d, want 20", fixRep)
	}
}

func TestApplyMUDReputationDecay_CapAtMaxDays(t *testing.T) {
	st, err := store.OpenAt(filepath.Join(t.TempDir(), "gl.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	// Run 30 days ago — should only apply max_decay_days (7).
	oldRun := time.Now().AddDate(0, 0, -30).UnixMilli()
	_ = st.RecordScoreEvent(ctx, store.ScoreEvent{XP: 10, Provider: "test", CreatedAt: oldRun})

	dbPath, _ := withMUDDB(t, map[string]int{"netrunners": 100})
	pack := GameWorldPack{
		ReputationDecay: ReputationDecayConfig{DecayPerDay: 5, Floor: 10, MaxDecayDays: 7},
	}

	if err := ApplyMUDReputationDecay(ctx, st, pack); err != nil {
		t.Fatalf("ApplyMUDReputationDecay: %v", err)
	}

	// 7 days × 5 = 35 reduction (capped at max_decay_days=7, not 30).
	rep := readRep(t, dbPath, "netrunners")
	if rep != 65 { // 100 - 35
		t.Errorf("netrunners rep = %d, want 65 (7-day cap applied)", rep)
	}
}

func TestApplyMUDReputationDecay_FloorClamping(t *testing.T) {
	st, err := store.OpenAt(filepath.Join(t.TempDir(), "gl.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	twoDaysAgo := time.Now().AddDate(0, 0, -2).UnixMilli()
	_ = st.RecordScoreEvent(ctx, store.ScoreEvent{XP: 10, Provider: "test", CreatedAt: twoDaysAgo})

	dbPath, _ := withMUDDB(t, map[string]int{"netrunners": 12})
	pack := GameWorldPack{
		ReputationDecay: ReputationDecayConfig{DecayPerDay: 10, Floor: 10, MaxDecayDays: 7},
	}

	if err := ApplyMUDReputationDecay(ctx, st, pack); err != nil {
		t.Fatalf("ApplyMUDReputationDecay: %v", err)
	}

	// 2 × 10 = 20 decay would bring 12 → -8, but floor is 10.
	rep := readRep(t, dbPath, "netrunners")
	if rep != 10 {
		t.Errorf("netrunners rep = %d, want 10 (floor clamped)", rep)
	}
}

func TestApplyMUDReputationDecay_SameDayNoDecay(t *testing.T) {
	st, err := store.OpenAt(filepath.Join(t.TempDir(), "gl.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	// Run happened today.
	_ = st.RecordScoreEvent(ctx, store.ScoreEvent{XP: 10, Provider: "test", CreatedAt: time.Now().UnixMilli()})

	dbPath, _ := withMUDDB(t, map[string]int{"netrunners": 50})
	pack := GameWorldPack{
		ReputationDecay: ReputationDecayConfig{DecayPerDay: 5, Floor: 10, MaxDecayDays: 7},
	}

	if err := ApplyMUDReputationDecay(ctx, st, pack); err != nil {
		t.Fatalf("ApplyMUDReputationDecay: %v", err)
	}

	rep := readRep(t, dbPath, "netrunners")
	if rep != 50 {
		t.Errorf("netrunners rep = %d, want 50 (no decay on same day)", rep)
	}
}

package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	_ "modernc.org/sqlite"
)

// Store is the SQLite-backed persistence layer for run history and research hints.
type Store struct {
	db *sql.DB
}

// ResearchEvent holds a single research loop hint record.
type ResearchEvent struct {
	QueryID        string
	Question       string
	Researchers    string
	Reason         string
	CompositeScore float64
}

// Open opens the store at the default path (~/.local/share/glitch/glitch.db).
func Open() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return OpenAt(filepath.Join(home, ".local", "share", "glitch", "glitch.db"))
}

// OpenAt opens the store at the given path, creating parent directories as needed.
func OpenAt(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(createSchema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// RecordRun inserts a new run record and returns the new row ID.
func (s *Store) RecordRun(kind, name, input string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO runs (kind, name, input, started_at) VALUES (?, ?, ?, ?)`,
		kind, name, input, time.Now().UnixMilli(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FinishRun updates an existing run with its output and exit status.
func (s *Store) FinishRun(id int64, output string, exitStatus int) error {
	_, err := s.db.Exec(
		`UPDATE runs SET output = ?, exit_status = ?, finished_at = ? WHERE id = ?`,
		output, exitStatus, time.Now().UnixMilli(), id,
	)
	return err
}

// RecordStep inserts or replaces a step record for a given run.
func (s *Store) RecordStep(runID int64, stepID, prompt, output, model string, durationMs int64) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO steps (run_id, step_id, prompt, output, model, duration_ms)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		runID, stepID, prompt, output, model, durationMs,
	)
	return err
}

// RecordResearchEvent inserts a new research event hint.
func (s *Store) RecordResearchEvent(evt ResearchEvent) error {
	_, err := s.db.Exec(
		`INSERT INTO research_events (query_id, question, researchers, composite_score, reason, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		evt.QueryID, evt.Question, evt.Researchers, evt.CompositeScore, evt.Reason, time.Now().UnixMilli(),
	)
	return err
}

// SimilarResearchEvents returns up to limit events whose questions are most
// similar to the given question, using token-level Jaccard similarity over the
// last 200 stored events.
func (s *Store) SimilarResearchEvents(question string, limit int) ([]ResearchEvent, error) {
	rows, err := s.db.Query(
		`SELECT query_id, question, researchers, composite_score, reason
		 FROM research_events
		 ORDER BY id DESC
		 LIMIT 200`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		evt   ResearchEvent
		score float64
	}

	qTokens := tokenise(question)
	var candidates []scored

	for rows.Next() {
		var evt ResearchEvent
		if err := rows.Scan(&evt.QueryID, &evt.Question, &evt.Researchers, &evt.CompositeScore, &evt.Reason); err != nil {
			return nil, err
		}
		j := jaccard(qTokens, tokenise(evt.Question))
		candidates = append(candidates, scored{evt: evt, score: j})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if limit > len(candidates) {
		limit = len(candidates)
	}
	result := make([]ResearchEvent, limit)
	for i := range result {
		result[i] = candidates[i].evt
	}
	return result, nil
}

// stopwords is a small set of common English words excluded from token similarity.
var stopwords = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "but": {},
	"in": {}, "on": {}, "at": {}, "to": {}, "for": {}, "of": {},
	"with": {}, "by": {}, "from": {}, "is": {}, "are": {}, "was": {},
	"were": {}, "be": {}, "been": {}, "has": {}, "have": {}, "had": {},
	"do": {}, "does": {}, "did": {}, "not": {}, "no": {}, "it": {},
	"its": {}, "this": {}, "that": {}, "as": {}, "if": {}, "so": {},
	"how": {}, "what": {}, "when": {}, "where": {}, "who": {}, "why": {},
	"can": {}, "will": {}, "would": {}, "could": {}, "should": {}, "may": {},
	"about": {}, "into": {}, "than": {}, "then": {}, "also": {},
}

// tokenise lowercases s, splits on non-alphanumeric characters, drops stopwords
// and single-character tokens, collapses trailing-s plurals, and deduplicates.
func tokenise(s string) []string {
	s = strings.ToLower(s)
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	seen := make(map[string]struct{}, len(words))
	out := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) <= 1 {
			continue
		}
		if _, ok := stopwords[w]; ok {
			continue
		}
		// Collapse trailing-s plurals (simple heuristic)
		if len(w) > 2 && strings.HasSuffix(w, "s") {
			w = w[:len(w)-1]
		}
		if _, dup := seen[w]; dup {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
	}
	return out
}

// jaccard computes the Jaccard similarity coefficient for two token slices.
func jaccard(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	setA := make(map[string]struct{}, len(a))
	for _, t := range a {
		setA[t] = struct{}{}
	}
	var intersection int
	setB := make(map[string]struct{}, len(b))
	for _, t := range b {
		setB[t] = struct{}{}
		if _, ok := setA[t]; ok {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

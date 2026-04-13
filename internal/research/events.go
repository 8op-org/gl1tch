package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType identifies what kind of research event occurred.
type EventType string

const (
	EventTypeAttempt  EventType = "research_attempt"
	EventTypeScore    EventType = "research_score"
	EventTypeFeedback EventType = "research_feedback"
)

// Event captures a single research loop occurrence.
type Event struct {
	Type      EventType       `json:"type"`
	Timestamp string          `json:"timestamp"`
	QueryID   string          `json:"query_id"`
	Question  string          `json:"question"`
	Iteration int             `json:"iteration"`
	Reason    Reason          `json:"reason,omitempty"`
	Score     Score           `json:"score,omitempty"`
	Bundle    *EvidenceBundle `json:"bundle,omitempty"`
}

// EventSink receives research events.
type EventSink interface {
	Emit(Event) error
}

// nopSink discards all events.
type nopSink struct{}

func (nopSink) Emit(Event) error { return nil }

// MemoryEventSink collects events in memory. Thread-safe. Exported for tests.
type MemoryEventSink struct {
	mu     sync.Mutex
	events []Event
}

// Emit appends the event.
func (m *MemoryEventSink) Emit(e Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}

// Events returns a copy of collected events.
func (m *MemoryEventSink) Events() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}

// FileEventSink appends JSONL to a file.
type FileEventSink struct {
	mu   sync.Mutex
	path string
}

// NewFileEventSink creates a sink that writes to the given path.
// If path is empty, defaults to ~/.glitch/research_events.jsonl.
func NewFileEventSink(path string) *FileEventSink {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		path = filepath.Join(home, ".glitch", "research_events.jsonl")
	}
	return &FileEventSink{path: path}
}

// Emit appends the event as a single JSON line.
func (f *FileEventSink) Emit(e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("fileEventSink: marshal: %w", err)
	}
	data = append(data, '\n')

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return fmt.Errorf("fileEventSink: mkdir: %w", err)
	}
	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("fileEventSink: open: %w", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

// emitAttempt emits both an attempt and a score event.
func emitAttempt(sink EventSink, question, queryID string, iter int, score Score, bundle *EvidenceBundle, reason Reason) {
	ts := time.Now().UTC().Format(time.RFC3339)
	_ = sink.Emit(Event{
		Type:      EventTypeAttempt,
		Timestamp: ts,
		QueryID:   queryID,
		Question:  question,
		Iteration: iter,
		Reason:    reason,
		Bundle:    bundle,
	})
	_ = sink.Emit(Event{
		Type:      EventTypeScore,
		Timestamp: ts,
		QueryID:   queryID,
		Question:  question,
		Iteration: iter,
		Score:     score,
	})
}

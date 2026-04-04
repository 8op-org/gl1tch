package picker

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sahilm/fuzzy"
)

// PickerItem is a single selectable row in the fuzzy picker.
type PickerItem struct {
	Kind         string // "session" | "pipeline" | "skill" | "agent" | "provider"
	Name         string
	Description  string
	SourceTag    string // "[global]" "[project]" "[copilot]" — empty for providers/sessions
	ProviderID   string // for kind=provider; also set after skill/agent picks a CLI
	ModelID      string // for kind=provider with a pre-selected model
	InjectText   string // for kind=skill|agent — sent to CLI after launch via tmux send-keys
	PipelineFile string // for kind=pipeline
	SessionIndex string // for kind=session — tmux window index to focus
	// internal — populated by ApplyFuzzy
	matchIndexes []int
}

// Filter returns the string used for fuzzy matching.
func (p PickerItem) Filter() string {
	if p.Description == "" {
		return p.Name
	}
	return p.Name + " " + p.Description
}

// SetMatchIndexes stores which character positions were matched by the fuzzy algorithm.
func (p *PickerItem) SetMatchIndexes(indexes []int) { p.matchIndexes = indexes }

// MatchIndexes returns the stored fuzzy match positions (nil when no filter active).
func (p PickerItem) MatchIndexes() []int { return p.matchIndexes }

// itemsSource implements fuzzy.Source over a []PickerItem.
type itemsSource []PickerItem

func (s itemsSource) Len() int            { return len(s) }
func (s itemsSource) String(i int) string { return s[i].Filter() }

// ApplyFuzzy filters items using sahilm/fuzzy.
// Returns all items (group order preserved) when query is empty.
// Returns matched items sorted by score when query is non-empty.
func ApplyFuzzy(query string, items []PickerItem) []PickerItem {
	if query == "" {
		out := make([]PickerItem, len(items))
		for i, item := range items {
			item.matchIndexes = nil
			out[i] = item
		}
		return out
	}
	matches := fuzzy.FindFrom(query, itemsSource(items))
	out := make([]PickerItem, len(matches))
	for i, m := range matches {
		item := items[m.Index]
		item.matchIndexes = m.MatchedIndexes
		out[i] = item
	}
	return out
}

// GlitchConfigDir returns ~/.config/glitch, or "" on error.
func GlitchConfigDir() string { return glitchConfigDir() }

// glitchConfigDir returns ~/.config/glitch, or "" on error.
func glitchConfigDir() string {
	if d := os.Getenv("GLITCH_CONFIG_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "glitch")
}

// MarshalPickerItem serialises item to JSON bytes for use in GLITCH_PICKER_SELECTION.
func MarshalPickerItem(item PickerItem) ([]byte, error) {
	return json.Marshal(item)
}

// UnmarshalPickerItem deserialises a PickerItem from JSON bytes.
func UnmarshalPickerItem(data []byte) (PickerItem, error) {
	var item PickerItem
	err := json.Unmarshal(data, &item)
	return item, err
}


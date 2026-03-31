package picker

import (
	"sort"
	"testing"
)

// TestProviderPriority_Order verifies the canonical ranking of known providers.
func TestProviderPriority_Order(t *testing.T) {
	rank := make(map[string]int, len(providerPriority))
	for i, name := range providerPriority {
		rank[name] = i
	}

	cases := []struct {
		a, b string // a must rank before b
	}{
		{"claude", "copilot"},
		{"copilot", "codex"},
		{"codex", "gemini"},
		{"gemini", "opencode"},
		{"opencode", "ollama"},
		{"ollama", "shell"},
	}
	for _, tc := range cases {
		ra, aOk := rank[tc.a]
		rb, bOk := rank[tc.b]
		if !aOk {
			t.Errorf("%q not found in providerPriority", tc.a)
			continue
		}
		if !bOk {
			t.Errorf("%q not found in providerPriority", tc.b)
			continue
		}
		if ra >= rb {
			t.Errorf("expected %q (rank %d) before %q (rank %d)", tc.a, ra, tc.b, rb)
		}
	}
}

// TestExtrasSort_KnownBeforeUnknown verifies that the sort used in buildProviders
// places priority providers before unknown ones.
func TestExtrasSort_KnownBeforeUnknown(t *testing.T) {
	type entry struct{ name string }
	extras := []entry{
		{"my-custom-agent"},
		{"gemini"},
		{"codex"},
		{"claude"},
		{"another-unknown"},
	}

	priorityRank := make(map[string]int, len(providerPriority))
	for i, name := range providerPriority {
		priorityRank[name] = i
	}
	sort.SliceStable(extras, func(i, j int) bool {
		ri, iOk := priorityRank[extras[i].name]
		rj, jOk := priorityRank[extras[j].name]
		if iOk && jOk {
			return ri < rj
		}
		if iOk {
			return true
		}
		if jOk {
			return false
		}
		return false
	})

	wantOrder := []string{"claude", "codex", "gemini", "my-custom-agent", "another-unknown"}
	for i, want := range wantOrder {
		if extras[i].name != want {
			t.Errorf("position %d: got %q, want %q", i, extras[i].name, want)
		}
	}
}

// TestExtrasSort_UnknownPreservesRelativeOrder verifies that unknown providers
// keep their original relative order after sorted priority providers.
func TestExtrasSort_UnknownPreservesRelativeOrder(t *testing.T) {
	type entry struct{ name string }
	extras := []entry{
		{"zebra-agent"},
		{"alpha-agent"},
		{"claude"},
	}

	priorityRank := make(map[string]int, len(providerPriority))
	for i, name := range providerPriority {
		priorityRank[name] = i
	}
	sort.SliceStable(extras, func(i, j int) bool {
		ri, iOk := priorityRank[extras[i].name]
		rj, jOk := priorityRank[extras[j].name]
		if iOk && jOk {
			return ri < rj
		}
		if iOk {
			return true
		}
		if jOk {
			return false
		}
		return false
	})

	// claude first, then unknown providers in original order: zebra-agent, alpha-agent
	if extras[0].name != "claude" {
		t.Errorf("position 0: got %q, want claude", extras[0].name)
	}
	if extras[1].name != "zebra-agent" {
		t.Errorf("position 1: got %q, want zebra-agent (original order preserved)", extras[1].name)
	}
	if extras[2].name != "alpha-agent" {
		t.Errorf("position 2: got %q, want alpha-agent (original order preserved)", extras[2].name)
	}
}

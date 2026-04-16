package workspace

import (
	"fmt"
	"sort"
	"strings"
)

// Serialize writes a Workspace back to s-expression format.
func Serialize(w *Workspace) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("(workspace %q", w.Name))

	if w.Description != "" {
		b.WriteString(fmt.Sprintf("\n  :description %q", w.Description))
	}
	if w.Owner != "" {
		b.WriteString(fmt.Sprintf("\n  :owner %q", w.Owner))
	}

	if len(w.Repos) > 0 {
		b.WriteString("\n  (repos")
		for _, r := range w.Repos {
			b.WriteString(fmt.Sprintf("\n    %q", r))
		}
		b.WriteString(")")
	}

	hasDefaults := w.Defaults.Model != "" || w.Defaults.Provider != "" ||
		w.Defaults.Elasticsearch != "" || len(w.Defaults.Params) > 0
	if hasDefaults {
		b.WriteString("\n  (defaults")
		if w.Defaults.Model != "" {
			b.WriteString(fmt.Sprintf("\n    :model %q", w.Defaults.Model))
		}
		if w.Defaults.Provider != "" {
			b.WriteString(fmt.Sprintf("\n    :provider %q", w.Defaults.Provider))
		}
		if w.Defaults.Elasticsearch != "" {
			b.WriteString(fmt.Sprintf("\n    :elasticsearch %q", w.Defaults.Elasticsearch))
		}
		if len(w.Defaults.Params) > 0 {
			b.WriteString("\n    (params")
			// Sort keys for deterministic output
			keys := make([]string, 0, len(w.Defaults.Params))
			for k := range w.Defaults.Params {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				b.WriteString(fmt.Sprintf("\n      :%s %q", k, w.Defaults.Params[k]))
			}
			b.WriteString(")")
		}
		b.WriteString(")")
	}

	b.WriteString(")\n")
	return []byte(b.String())
}

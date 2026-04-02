package console

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/8op-org/gl1tch/internal/executor"
)

// WidgetConfig is a loaded widget-capable sidecar with everything dispatch
// code needs to activate a widget mode without re-reading disk.
type WidgetConfig struct {
	Schema executor.SidecarSchema
}

// WidgetRegistry holds all widget-capable sidecars loaded at startup.
type WidgetRegistry struct {
	widgets []WidgetConfig
}

// matchTopic returns true when topic matches pattern. Pattern may end with ".*"
// as a wildcard prefix (e.g. "mud.*" matches "mud.room.entered").
func matchTopic(pattern, topic string) bool {
	if pattern == topic {
		return true
	}
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return strings.HasPrefix(topic, prefix+".")
	}
	return false
}

// LoadWidgetRegistry scans wrappersDir for sidecar YAML files and loads
// those that declare a non-zero mode block. Invalid or duplicate triggers
// are skipped with a warning log.
func LoadWidgetRegistry(wrappersDir string) *WidgetRegistry {
	reg := &WidgetRegistry{}
	entries, err := os.ReadDir(wrappersDir)
	if err != nil {
		// Directory may not exist yet — not an error.
		return reg
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(wrappersDir, e.Name()))
		if err != nil {
			log.Printf("widget_registry: read %s: %v", e.Name(), err)
			continue
		}
		var schema executor.SidecarSchema
		if err := yaml.Unmarshal(data, &schema); err != nil {
			log.Printf("widget_registry: parse %s: %v", e.Name(), err)
			continue
		}
		if schema.Mode.IsZero() {
			continue
		}
		// Validate required fields.
		if schema.Mode.Logo == "" || schema.Mode.Speaker == "" || schema.Mode.ExitCommand == "" {
			log.Printf("widget_registry: WARN %s: mode block missing required fields (logo/speaker/exit_command) — skipping", e.Name())
			continue
		}
		if seen[schema.Mode.Trigger] {
			log.Printf("widget_registry: WARN %s: duplicate trigger %q — first loaded wins", e.Name(), schema.Mode.Trigger)
			continue
		}
		seen[schema.Mode.Trigger] = true
		reg.widgets = append(reg.widgets, WidgetConfig{Schema: schema})
	}
	return reg
}

// FindByTrigger returns the WidgetConfig whose Mode.Trigger matches trigger,
// or nil if no match.
func (r *WidgetRegistry) FindByTrigger(trigger string) *WidgetConfig {
	for i := range r.widgets {
		if r.widgets[i].Schema.Mode.Trigger == trigger {
			return &r.widgets[i]
		}
	}
	return nil
}

// AllSignalTopics returns a deduplicated slice of all topic values declared
// across all loaded sidecars' signals blocks.
func (r *WidgetRegistry) AllSignalTopics() []string {
	seen := map[string]bool{}
	var topics []string
	for _, w := range r.widgets {
		for _, sig := range w.Schema.Signals {
			if sig.Topic != "" && !seen[sig.Topic] {
				seen[sig.Topic] = true
				topics = append(topics, sig.Topic)
			}
		}
	}
	return topics
}

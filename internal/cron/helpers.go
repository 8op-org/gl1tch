package cron

import (
	"fmt"
	"time"

	robfigcron "github.com/robfig/cron/v3"
)

// cronParser is the standard 5-field parser shared across the package.
var cronParser = robfigcron.NewParser(
	robfigcron.Minute | robfigcron.Hour | robfigcron.Dom | robfigcron.Month | robfigcron.Dow,
)

// NextRun parses entry.Schedule and returns the next scheduled fire time after
// time.Now(). Returns an error if the schedule expression is invalid.
func NextRun(entry Entry) (time.Time, error) {
	sched, err := cronParser.Parse(entry.Schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("cron: parse schedule %q: %w", entry.Schedule, err)
	}
	return sched.Next(time.Now()), nil
}

// FormatRelative returns a concise human-readable relative duration from now
// to t, e.g. "in 4m", "in 2h 30m", "in 3d". Returns "now" if t is in the
// past or within 1 second.
func FormatRelative(t time.Time) string {
	d := time.Until(t)
	if d <= time.Second {
		return "now"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("in %dd %dh", days, hours)
	case days > 0:
		return fmt.Sprintf("in %dd", days)
	case hours > 0 && mins > 0:
		return fmt.Sprintf("in %dh %dm", hours, mins)
	case hours > 0:
		return fmt.Sprintf("in %dh", hours)
	default:
		return fmt.Sprintf("in %dm", max1(mins, 1))
	}
}

// max1 returns the larger of a and b (avoids importing slices for a trivial op).
func max1(a, b int) int {
	if a > b {
		return a
	}
	return b
}

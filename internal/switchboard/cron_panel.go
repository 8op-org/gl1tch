package switchboard

import (
	"fmt"
	"sort"

	orcaicron "github.com/adam-stokes/orcai/internal/cron"
)

// CronPanel holds display state for the CRON JOBS panel.
type CronPanel struct {
	selectedIdx  int
	scrollOffset int
	focused      bool
}

// buildCronSection renders the CRON JOBS box as a slice of ANSI lines.
func (m Model) buildCronSection(w int) []string {
	pal := m.ansiPalette()
	borderColor := pal.Border
	if m.cronPanel.focused {
		borderColor = pal.Accent
	}

	var rows []string
	if sprite := PanelHeader(m.activeBundle(), "cron", w); sprite != nil {
		rows = append(rows, sprite...)
	} else {
		rows = append(rows, boxTop(w, "CRON JOBS", borderColor, pal.Accent))
	}

	entries, err := orcaicron.LoadConfig()
	if err != nil || len(entries) == 0 {
		rows = append(rows, boxRow(pal.Dim+"  no scheduled jobs"+aRst, w, borderColor))
		rows = append(rows, boxBot(w, borderColor))
		return rows
	}

	// Sort entries by next run time ascending.
	type entryWithNext struct {
		e    orcaicron.Entry
		next string
	}
	sorted := make([]entryWithNext, 0, len(entries))
	for _, e := range entries {
		t, err := orcaicron.NextRun(e)
		rel := ""
		if err == nil {
			rel = orcaicron.FormatRelative(t)
		}
		sorted = append(sorted, entryWithNext{e: e, next: rel})
	}
	sort.Slice(sorted, func(i, j int) bool {
		ti, ei := orcaicron.NextRun(sorted[i].e)
		tj, ej := orcaicron.NextRun(sorted[j].e)
		if ei != nil || ej != nil {
			return false
		}
		return ti.Before(tj)
	})

	for i, item := range sorted {
		name := item.e.Name
		kind := item.e.Kind
		sched := item.e.Schedule
		rel := item.next

		// Build compact display: name  kind  schedule  relative
		content := fmt.Sprintf("  %s  %s  %s  %s%s%s",
			pal.FG+name+aRst,
			pal.Dim+kind+aRst,
			pal.Dim+sched+aRst,
			pal.Accent, rel, aRst,
		)

		if m.cronPanel.focused && i == m.cronPanel.selectedIdx {
			// Highlight selected row.
			highlightContent := fmt.Sprintf("  %s%s%s  %s  %s  %s%s%s",
				pal.Accent, name, aRst,
				pal.Dim+kind+aRst,
				pal.Dim+sched+aRst,
				pal.Accent, rel, aRst,
			)
			rows = append(rows, boxRow(highlightContent, w, borderColor))
		} else {
			rows = append(rows, boxRow(content, w, borderColor))
		}
	}

	rows = append(rows, boxBot(w, borderColor))
	return rows
}

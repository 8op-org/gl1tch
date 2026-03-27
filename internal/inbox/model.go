// Package inbox implements the ORCAI inbox panel — a BubbleTea component that
// displays a scrollable list of recorded pipeline and agent run results with
// polling refresh and a full-screen detail modal.
package inbox

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/adam-stokes/orcai/internal/store"
	"github.com/adam-stokes/orcai/internal/themes"
)

// pollInterval is how often the inbox polls the store for new runs.
const pollInterval = 5 * time.Second

// RunCompletedMsg triggers an immediate inbox refresh when a run completes in-process.
type RunCompletedMsg struct {
	RunID int64
}

// RerunMsg requests that a pipeline or agent be re-run.
type RerunMsg struct {
	// Kind is "pipeline" or "agent".
	Kind string
	// Target is the name of the pipeline or agent to re-run.
	Target string
}

// tickMsg triggers a poll-based refresh.
type tickMsg struct{}

// item wraps a store.Run for display in the bubbles list.
type item struct {
	run    store.Run
	bundle *themes.Bundle
}

// Title returns the run name with a kind badge for display in the list.
func (i item) Title() string {
	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color(i.bundle.Palette.Dim)).
		Render("[" + i.run.Kind + "]")
	return badge + " " + i.run.Name
}

// Description returns elapsed/finished time and a status indicator.
func (i item) Description() string {
	return statusIndicator(i.run, i.bundle) + "  " + elapsedStr(i.run)
}

// FilterValue returns the run name for list filtering.
func (i item) FilterValue() string { return i.run.Name }

// statusIndicator returns a colored dot string representing run status.
func statusIndicator(run store.Run, bundle *themes.Bundle) string {
	switch {
	case run.ExitStatus == nil: // in-flight
		return lipgloss.NewStyle().Foreground(lipgloss.Color(bundle.Palette.Accent)).Render("◉")
	case *run.ExitStatus == 0: // success
		return lipgloss.NewStyle().Foreground(lipgloss.Color(bundle.Palette.Success)).Render("●")
	default: // error
		return lipgloss.NewStyle().Foreground(lipgloss.Color(bundle.Palette.Error)).Render("●")
	}
}

// elapsedStr formats how long a run took or how long it has been running.
func elapsedStr(run store.Run) string {
	if run.FinishedAt != nil {
		dur := time.Duration((*run.FinishedAt - run.StartedAt) * int64(time.Millisecond))
		return fmt.Sprintf("%s", dur.Round(time.Second))
	}
	dur := time.Since(time.UnixMilli(run.StartedAt))
	return fmt.Sprintf("running %s", dur.Round(time.Second))
}

// Model is the BubbleTea model for the inbox panel.
type Model struct {
	store   *store.Store
	bundle  *themes.Bundle
	list    list.Model
	runs    []store.Run // parallel slice to list items for modal access
	width   int
	height  int
	focused bool
}

// New creates an inbox Model. s may be nil (renders an empty state).
func New(s *store.Store, bundle *themes.Bundle) Model {
	if bundle == nil {
		bundle = &themes.Bundle{
			Palette: themes.Palette{
				BG:      "#282a36",
				FG:      "#f8f8f2",
				Accent:  "#bd93f9",
				Dim:     "#6272a4",
				Border:  "#44475a",
				Error:   "#ff5555",
				Success: "#50fa7b",
			},
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color(bundle.Palette.Accent)).
		BorderForeground(lipgloss.Color(bundle.Palette.Accent))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color(bundle.Palette.Dim)).
		BorderForeground(lipgloss.Color(bundle.Palette.Accent))

	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)

	return Model{
		store:  s,
		bundle: bundle,
		list:   l,
	}
}


// Init returns the initial command that starts the poll tick.
func (m Model) Init() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// scheduleNextTick returns a command that fires the next poll tick.
func scheduleNextTick() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// refreshRuns queries the store and updates the list items.
// Returns the updated model and the next tick command.
func (m Model) refreshRuns() (Model, tea.Cmd) {
	if m.store == nil {
		return m, scheduleNextTick()
	}
	runs, err := m.store.QueryRuns(50)
	if err != nil {
		return m, scheduleNextTick()
	}
	m.runs = runs
	items := make([]list.Item, len(runs))
	for i, r := range runs {
		items[i] = item{run: r, bundle: m.bundle}
	}
	cmd := m.list.SetItems(items)
	return m, tea.Batch(cmd, scheduleNextTick())
}

// Update handles BubbleTea messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {

	case tickMsg, RunCompletedMsg:
		return m.refreshRuns()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the inbox panel.
func (m Model) View() string {
	return m.panelOnlyView()
}

// PanelView renders just the list panel at the given dimensions, ignoring any
// open modal. Used by the switchboard to embed the inbox in the left column.
func (m Model) PanelView(w, h int) string {
	m.width = w
	m.height = h
	return m.panelOnlyView()
}

// panelOnlyView renders the list panel without checking for an open modal.
func (m Model) panelOnlyView() string {
	borderColor := lipgloss.Color(m.bundle.Palette.Border)
	if m.focused {
		borderColor = lipgloss.Color(m.bundle.Palette.Accent)
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.bundle.Palette.Accent)).
		Bold(true).
		Render("INBOX")

	// Reserve 2 rows for the border (top + bottom) and 1 for the title bar.
	innerW := m.width - 2
	innerH := m.height - 3
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	var body string
	if len(m.runs) == 0 {
		body = lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.bundle.Palette.Dim)).
			Width(innerW).
			Height(innerH).
			Align(lipgloss.Center, lipgloss.Center).
			Render("(empty)")
	} else {
		m.list.SetSize(innerW, innerH)
		body = m.list.View()
	}

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerW).
		Padding(0, 1)

	return panel.Render(title + "\n" + body)
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	innerW := w - 4 // border + padding
	innerH := h - 3
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	m.list.SetSize(innerW, innerH)
}

// SetTheme updates the active theme bundle.
func (m *Model) SetTheme(bundle *themes.Bundle) {
	if bundle == nil {
		return
	}
	m.bundle = bundle
}

// SetFocused sets the focus state of the panel.
func (m *Model) SetFocused(v bool) {
	m.focused = v
}

// Runs returns the current slice of recorded runs. Used by the switchboard to
// render the inbox section using its own ANSI box drawing functions.
func (m Model) Runs() []store.Run { return m.runs }

// SelectedIndex returns the index of the currently selected list item.
func (m Model) SelectedIndex() int { return m.list.Index() }


package crontui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/log"

	"github.com/adam-stokes/orcai/internal/cron"
)

// New creates the cron TUI model, wiring up the scheduler with a LogSink so
// that all scheduler log output is captured to the in-app log pane.
func New() (Model, error) {
	logCh := make(chan tea.Msg, 256)
	sink := NewLogSink(logCh)
	logger := log.NewWithOptions(sink, log.Options{
		ReportTimestamp: true,
		Prefix:          "orcai-cron",
	})
	sched := cron.New(logger, nil)

	entries, _ := cron.LoadConfig()

	fi := textinput.New()
	fi.Placeholder = "/ filter..."
	fi.CharLimit = 64

	m := Model{
		scheduler:   sched,
		entries:     entries,
		filtered:    entries,
		logCh:       logCh,
		filterInput: fi,
	}
	return m, nil
}

// Init starts the scheduler, begins the 30-second tick, and starts listening
// for log lines.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if err := m.scheduler.Start(context.Background()); err != nil {
				return logLineMsg{line: "ERROR: failed to start scheduler: " + err.Error()}
			}
			return logLineMsg{line: "INFO: scheduler started"}
		},
		tick(),
		listenLogs(m.logCh),
	)
}

// tick schedules a reload every 30 seconds.
func tick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// listenLogs blocks until a message arrives on ch, then returns it.
// The Update handler re-schedules this after each message so the listener
// stays active for the lifetime of the program.
func listenLogs(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

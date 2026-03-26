// Package switchboard implements the ORCAI Switchboard — a full-screen BubbleTea
// TUI that merges the sysop panel and welcome dashboard into a single control
// surface with a Pipeline Launcher, Agent Runner, and Activity Feed.
package switchboard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/plugin"
)

// ── ANSI palette — Dracula BBS aesthetic ─────────────────────────────────────

const (
	aBlu   = "\x1b[34m"  // blue
	aPur   = "\x1b[35m"  // purple
	aCyn   = "\x1b[36m"  // cyan (unused but defined per spec)
	aBrC   = "\x1b[96m"  // bright cyan
	aPnk   = "\x1b[95m"  // pink
	aGrn   = "\x1b[32m"  // green
	aRed   = "\x1b[31m"  // red
	aDim   = "\x1b[2m"   // dim
	aBld   = "\x1b[1m"   // bold
	aWht   = "\x1b[97m"  // white
	aSelBg = "\x1b[44m"  // blue bg (selection)
	aRst   = "\x1b[0m"   // reset
	aBC    = "\x1b[36m"  // cyan borders (alias)
	aBlu2  = "\x1b[34m"  // blue alias (unused var prevention)
)

// suppress unused-const warnings at compile time
var _ = aBlu
var _ = aCyn
var _ = aBlu2

// ── Feed types ────────────────────────────────────────────────────────────────

// FeedStatus is the lifecycle state of an activity feed entry.
type FeedStatus int

const (
	FeedRunning FeedStatus = iota
	FeedDone
	FeedFailed
)

// agentInnerHeight is the fixed number of body rows inside the AGENT RUNNER box.
const agentInnerHeight = 8

// maxParallelJobs is the maximum number of jobs that can run concurrently.
const maxParallelJobs = 8

type feedEntry struct {
	id         string
	title      string
	status     FeedStatus
	ts         time.Time
	lines      []string
	tmuxWindow string // fully-qualified target "session:orcai-<feedID>", empty if no window
	logFile    string // /tmp/orcai-<feedID>.log
	doneFile   string // non-empty for window-mode jobs; written by the shell when the command exits
}

// ── Section types ─────────────────────────────────────────────────────────────

type launcherSection struct {
	pipelines []string
	selected  int
	focused   bool
}

type agentSection struct {
	formStep         int // 0=provider, 1=model, 2=prompt
	providers        []picker.ProviderDef
	selectedProvider int
	selectedModel    int
	prompt           textarea.Model
	focused          bool
	agentScrollOffset int
}

type jobHandle struct {
	id         string
	cancel     context.CancelFunc
	ch         chan tea.Msg
	tmuxWindow string
	logFile    string // /tmp/orcai-<feedID>.log — tailed in the tmux window
}

// ── Tea messages ──────────────────────────────────────────────────────────────

// FeedLineMsg is a tea.Msg carrying one line of output from a running job.
// Exported so test packages can assert on it.
type FeedLineMsg struct {
	ID   string
	Line string
}

type jobDoneMsg struct {
	id string
}

type jobFailedMsg struct {
	id  string
	err error
}

type tickMsg time.Time

// ── Window / telemetry types (preserved from sidebar for backwards compat) ────

// Window represents a tmux window (excluding window 0).
type Window struct {
	Index  int
	Name   string
	Active bool
}

// ParseWindows parses output of:
//
//	tmux list-windows -t orcai -F "#{window_index} #{window_name} #{window_active}"
//
// Skips window 0 (the ORCAI home window).
func ParseWindows(output string) []Window {
	var windows []Window
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil || idx == 0 {
			continue
		}
		windows = append(windows, Window{
			Index:  idx,
			Name:   parts[1],
			Active: parts[2] == "1",
		})
	}
	return windows
}

// TelemetryMsg carries a parsed telemetry event from the bus.
type TelemetryMsg struct {
	SessionID    string
	WindowName   string
	Provider     string
	Status       string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
}

// ── Model ─────────────────────────────────────────────────────────────────────

// Model is the BubbleTea model for the Switchboard.
type Model struct {
	width              int
	height             int
	feed               []feedEntry // ring buffer, cap 200
	launcher           launcherSection
	agent              agentSection
	activeJobs         map[string]*jobHandle
	feedSelected       int // index into feed for expanded view
	confirmQuit        bool
	feedScrollOffset   int
	feedFocused        bool
	signalBoard        SignalBoard
	signalBoardFocused bool
}

// New creates a new Switchboard Model, discovering pipelines and providers.
func New() Model {
	ta := textarea.New()
	ta.Placeholder = "Enter prompt… (ctrl+s to submit)"
	ta.CharLimit = 4096
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(4)

	return Model{
		launcher: launcherSection{
			pipelines: ScanPipelines(pipelinesDir()),
			focused:   true,
		},
		agent: agentSection{
			providers: picker.BuildProviders(),
			prompt:    ta,
		},
		signalBoard: SignalBoard{activeFilter: "all"},
		activeJobs:  make(map[string]*jobHandle),
	}
}

// NewWithWindows is kept for backward-compat with sidebar-based callers.
// It ignores the window list and calls New().
func NewWithWindows(_ []Window) Model { return New() }

// NewWithPipelines creates a Model with a fixed pipeline list — used in tests.
func NewWithPipelines(pipelines []string) Model {
	m := New()
	m.launcher.pipelines = pipelines
	m.launcher.selected = 0
	m.launcher.focused = true
	return m
}

// NewWithTestProviders creates a Model with synthetic providers for testing.
func NewWithTestProviders() Model {
	m := New()
	m.agent.providers = []picker.ProviderDef{
		{
			ID:    "test-provider",
			Label: "Test Provider",
			Models: []picker.ModelOption{
				{ID: "model-a", Label: "Model A"},
				{ID: "model-b", Label: "Model B"},
			},
		},
	}
	return m
}

// Cursor returns the launcher cursor position — used in tests.
func (m Model) Cursor() int { return m.launcher.selected }

// AgentFormStep returns the current agent form step — used in tests.
func (m Model) AgentFormStep() int { return m.agent.formStep }

// FeedScrollOffset returns the current feed scroll offset — used in tests.
func (m Model) FeedScrollOffset() int { return m.feedScrollOffset }

// BuildAgentSection is an exported wrapper for tests.
func (m Model) BuildAgentSection(w int) []string { return m.buildAgentSection(w) }

// BuildSignalBoard is an exported wrapper for tests.
func (m Model) BuildSignalBoard(height, width int) []string { return m.buildSignalBoard(height, width) }

// SignalBoardBlinkOn returns the current blink state — used in tests.
func (m Model) SignalBoardBlinkOn() bool { return m.signalBoard.blinkOn }

// ActiveJobsCount returns the number of currently active jobs — used in tests.
func (m Model) ActiveJobsCount() int { return len(m.activeJobs) }

// AddActiveJob injects a fake job handle for testing purposes.
func (m Model) AddActiveJob(id string) Model {
	if m.activeJobs == nil {
		m.activeJobs = make(map[string]*jobHandle)
	}
	m.activeJobs[id] = &jobHandle{id: id}
	return m
}

// MaxParallelJobs returns the parallel job cap constant — used in tests.
func MaxParallelJobs() int { return maxParallelJobs }

// MakeTickMsg returns a tickMsg for use in tests.
func MakeTickMsg() tea.Msg { return tickMsg(time.Now()) }

// SignalBoardFocused returns the signal board focus state — used in tests.
func (m Model) SignalBoardFocused() bool { return m.signalBoardFocused }

// SetSignalBoardFocused sets the signal board focus state — used in tests.
func (m Model) SetSignalBoardFocused(v bool) Model {
	m.signalBoardFocused = v
	m.launcher.focused = false
	m.agent.focused = false
	m.feedFocused = false
	return m
}

// SetFeedFocused sets the feed focus state — used in tests.
func (m Model) SetFeedFocused(v bool) Model {
	m.feedFocused = v
	m.launcher.focused = false
	m.agent.focused = false
	m.signalBoardFocused = false
	return m
}

// SetFeedSelected sets the selected feed entry index — used in tests.
func (m Model) SetFeedSelected(idx int) Model {
	m.feedSelected = idx
	return m
}


// AddFeedEntry adds a feed entry — used in tests.
func (m Model) AddFeedEntry(id, title string, status FeedStatus, lines []string) Model {
	entry := feedEntry{
		id:     id,
		title:  title,
		status: status,
		ts:     time.Now(),
		lines:  lines,
	}
	m.feed = append([]feedEntry{entry}, m.feed...)
	return m
}

// AddFeedEntryWithTmux adds a feed entry with a tmux window — used in tests.
func (m Model) AddFeedEntryWithTmux(id, title string, status FeedStatus, tmuxWindow string) Model {
	entry := feedEntry{
		id:         id,
		title:      title,
		status:     status,
		ts:         time.Now(),
		tmuxWindow: tmuxWindow,
	}
	m.feed = append([]feedEntry{entry}, m.feed...)
	return m
}

// ── Init ──────────────────────────────────────────────────────────────────────

// Init starts the tick command.
func (m Model) Init() tea.Cmd { return tickCmd() }

func tickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ── Pipeline helpers ──────────────────────────────────────────────────────────

func pipelinesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "orcai", "pipelines")
}

// ScanPipelines lists *.pipeline.yaml basenames (without extension) from dir.
// Exported so tests can call it directly.
// Returns an empty slice if dir is missing or empty.
func ScanPipelines(dir string) []string {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".pipeline.yaml") {
			names = append(names, strings.TrimSuffix(n, ".pipeline.yaml"))
		}
	}
	return names
}

// ── ChanPublisher ─────────────────────────────────────────────────────────────

// ChanPublisher implements pipeline.EventPublisher and forwards events as
// FeedLineMsg values through a channel consumed by the BubbleTea update loop.
// Exported so tests can construct and verify it.
type ChanPublisher struct {
	id string
	ch chan<- tea.Msg
}

// NewChanPublisher creates a ChanPublisher for the given feed entry id and channel.
func NewChanPublisher(id string, ch chan<- tea.Msg) *ChanPublisher {
	return &ChanPublisher{id: id, ch: ch}
}

// Publish converts a pipeline lifecycle event to a FeedLineMsg and sends it.
func (p *ChanPublisher) Publish(_ context.Context, topic string, payload []byte) error {
	line := fmt.Sprintf("[%s] %s", topic, strings.TrimSpace(string(payload)))
	select {
	case p.ch <- FeedLineMsg{ID: p.id, Line: line}:
	default:
	}
	return nil
}

// lineWriter is an io.Writer that buffers lines and sends FeedLineMsg per line.
type lineWriter struct {
	id  string
	ch  chan<- tea.Msg
	buf bytes.Buffer
}

func (w *lineWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	for {
		data := w.buf.Bytes()
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimRight(string(data[:idx]), "\r")
		w.buf.Next(idx + 1)
		if line != "" {
			select {
			case w.ch <- FeedLineMsg{ID: w.id, Line: line}:
			default:
			}
		}
	}
	return n, err
}

func (w *lineWriter) flush() {
	if remaining := strings.TrimSpace(w.buf.String()); remaining != "" {
		select {
		case w.ch <- FeedLineMsg{ID: w.id, Line: remaining}:
		default:
		}
	}
}

var _ io.Writer = (*lineWriter)(nil)

// ── Update ────────────────────────────────────────────────────────────────────

// Update handles tea.Msg values.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		leftW := m.leftColWidth()
		m.agent.prompt.SetWidth(max(leftW-4, 10))
		m.clampFeedScroll()
		return m, nil

	case tickMsg:
		// Toggle blink if any job is running.
		for _, e := range m.feed {
			if e.status == FeedRunning {
				m.signalBoard.blinkOn = !m.signalBoard.blinkOn
				break
			}
		}
		return m, tickCmd()

	case TelemetryMsg:
		line := fmt.Sprintf("telemetry: window=%s provider=%s status=%s", msg.WindowName, msg.Provider, msg.Status)
		entry := feedEntry{
			id:     "tel-" + msg.SessionID,
			title:  "tmux/" + msg.WindowName,
			status: FeedDone,
			ts:     time.Now(),
			lines:  []string{line},
		}
		m.feed = append([]feedEntry{entry}, m.feed...)
		if len(m.feed) > 200 {
			m.feed = m.feed[:200]
		}
		m.feedScrollOffset = 0
		return m, nil

	case FeedLineMsg:
		m = m.appendFeedLine(msg.ID, msg.Line)
		// For in-process (agent) jobs the log file is written here.
		// Window-mode (pipeline) jobs write via tee in the shell — skip.
		for _, e := range m.feed {
			if e.id == msg.ID && e.logFile != "" && e.doneFile == "" {
				appendToFile(e.logFile, stripANSI(msg.Line)+"\n")
				break
			}
		}
		// Re-issue drain only for the job that produced this message.
		// Draining all jobs would accumulate goroutines and starve channels.
		if jh, ok := m.activeJobs[msg.ID]; ok {
			return m, drainChan(jh.ch)
		}
		return m, nil

	case jobDoneMsg:
		// Drain any remaining lines buffered in the channel before marking done.
		if jh, ok := m.activeJobs[msg.id]; ok {
		drainDone:
			for {
				select {
				case buffered, ok := <-jh.ch:
					if !ok {
						break drainDone
					}
					if fl, ok2 := buffered.(FeedLineMsg); ok2 {
						m = m.appendFeedLine(fl.ID, fl.Line)
					}
				default:
					break drainDone
				}
			}
		}
		m = m.setFeedStatus(msg.id, FeedDone)
		delete(m.activeJobs, msg.id)
		return m, nil

	case jobFailedMsg:
		if jh, ok := m.activeJobs[msg.id]; ok {
		drainFailed:
			for {
				select {
				case buffered, ok := <-jh.ch:
					if !ok {
						break drainFailed
					}
					if fl, ok2 := buffered.(FeedLineMsg); ok2 {
						m = m.appendFeedLine(fl.ID, fl.Line)
					}
				default:
					break drainFailed
				}
			}
		}
		m = m.setFeedStatus(msg.id, FeedFailed)
		if msg.err != nil {
			m = m.appendFeedLine(msg.id, "error: "+msg.err.Error())
		}
		delete(m.activeJobs, msg.id)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Confirm quit when a job is running.
	if m.confirmQuit {
		switch key {
		case "y", "Y", "enter":
			for _, jh := range m.activeJobs {
				jh.cancel()
			}
			return m, tea.Quit
		default:
			m.confirmQuit = false
			return m, nil
		}
	}

	// When at the prompt step, forward keys to textarea first,
	// except for submit (ctrl+s) and navigation (esc, ctrl+c, tab, q).
	if m.agent.focused && m.agent.formStep == 2 {
		switch key {
		case "ctrl+s":
			return m.handleEnter()
		case "esc":
			m.agent.prompt.Blur()
			m.agent.formStep = 1
			// If going back to step 1 but no models, go to step 0.
			prov := m.currentProvider()
			if prov == nil || len(nonSepModels(prov.Models)) == 0 {
				m.agent.formStep = 0
			}
			return m, nil
		case "ctrl+c":
			if len(m.activeJobs) > 0 {
				m.confirmQuit = true
				return m, nil
			}
			return m, tea.Quit
		case "tab":
			// Wrap back to launcher.
			m.agent.focused = false
			m.launcher.focused = true
			m.agent.formStep = 0
			m.agent.prompt.Blur()
			return m, nil
		case "q":
			if len(m.activeJobs) > 0 {
				m.confirmQuit = true
				return m, nil
			}
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.agent.prompt, cmd = m.agent.prompt.Update(msg)
			return m, cmd
		}
	}

	switch key {
	case "ctrl+c":
		if len(m.activeJobs) > 0 {
			m.confirmQuit = true
			return m, nil
		}
		return m, tea.Quit

	case "q":
		if len(m.activeJobs) > 0 {
			m.confirmQuit = true
			return m, nil
		}
		return m, tea.Quit

	case "tab":
		if m.feedFocused {
			// feed → launcher
			m.feedFocused = false
			m.launcher.focused = true
		} else if m.signalBoardFocused {
			// signalBoard → launcher
			m.signalBoardFocused = false
			m.launcher.focused = true
		} else if m.agent.focused {
			if m.agent.formStep < 2 {
				m = m.agentAdvanceStep()
			} else {
				// agent (step 2) → signalBoard
				m.agent.focused = false
				m.agent.formStep = 0
				m.agent.prompt.Blur()
				m.signalBoardFocused = true
			}
		} else if m.launcher.focused {
			m.launcher.focused = false
			m.agent.focused = true
		}
		return m, nil

	case "f":
		if m.signalBoardFocused {
			m.signalBoard.cycleFilter()
		} else {
			// Toggle activity feed focus so ↑↓ scrolls through output lines.
			m.launcher.focused = false
			m.agent.focused = false
			m.signalBoardFocused = false
			m.feedFocused = !m.feedFocused
		}
		return m, nil

	case "a":
		m.launcher.focused = false
		m.agent.focused = true
		m.feedFocused = false
		return m, nil

	case "s":
		m.launcher.focused = false
		m.agent.focused = false
		m.feedFocused = false
		m.signalBoardFocused = true
		return m, nil

	case "r":
		m.launcher.pipelines = ScanPipelines(pipelinesDir())
		if m.launcher.selected >= len(m.launcher.pipelines) && m.launcher.selected > 0 {
			m.launcher.selected = max(len(m.launcher.pipelines)-1, 0)
		}
		m.agent.providers = picker.BuildProviders()
		m.agent.selectedProvider = 0
		m.agent.selectedModel = 0
		return m, nil

	case "j", "down":
		return m.handleDown(), nil

	case "k", "up":
		return m.handleUp(), nil

	case "left":
		if m.agent.focused && m.agent.formStep > 0 {
			m.agent.formStep--
			// If going back to step 1 but provider has no models, skip to 0.
			if m.agent.formStep == 1 {
				prov := m.currentProvider()
				if prov != nil && len(nonSepModels(prov.Models)) == 0 {
					m.agent.formStep = 0
				}
			}
		}
		return m, nil

	case "esc":
		if m.feedFocused {
			m.feedFocused = false
			m.launcher.focused = true
			return m, nil
		} else if m.signalBoardFocused {
			m.signalBoardFocused = false
			m.launcher.focused = true
		} else if m.agent.focused {
			if m.agent.formStep > 0 {
				m.agent.formStep--
				m.agent.agentScrollOffset = 0
			} else {
				m.agent.focused = false
				m.launcher.focused = true
			}
		}
		return m, nil

	case "right":
		if m.agent.focused && m.agent.formStep < 2 {
			m = m.agentAdvanceStep()
		}
		return m, nil

	case "enter":
		return m.handleEnter()
	}

	return m, nil
}

func (m Model) handleDown() Model {
	if m.feedFocused {
		m.feedScrollOffset++
		m.clampFeedScroll()
		return m
	}
	if m.signalBoardFocused {
		filtered := m.filteredFeed()
		if m.signalBoard.selectedIdx < len(filtered)-1 {
			m.signalBoard.selectedIdx++
		}
		return m
	}
	if m.launcher.focused {
		if m.launcher.selected < len(m.launcher.pipelines)-1 {
			m.launcher.selected++
		}
		return m
	}
	if m.agent.focused {
		switch m.agent.formStep {
		case 0:
			if m.agent.selectedProvider < len(m.agent.providers)-1 {
				m.agent.selectedProvider++
			}
		case 1:
			prov := m.currentProvider()
			if prov != nil {
				models := nonSepModels(prov.Models)
				if m.agent.selectedModel < len(models)-1 {
					m.agent.selectedModel++
				}
			}
		}
	}
	return m
}

func (m Model) handleUp() Model {
	if m.feedFocused {
		if m.feedScrollOffset > 0 {
			m.feedScrollOffset--
		}
		m.clampFeedScroll()
		return m
	}
	if m.signalBoardFocused {
		if m.signalBoard.selectedIdx > 0 {
			m.signalBoard.selectedIdx--
		}
		return m
	}
	if m.launcher.focused {
		if m.launcher.selected > 0 {
			m.launcher.selected--
		}
		return m
	}
	if m.agent.focused {
		switch m.agent.formStep {
		case 0:
			if m.agent.selectedProvider > 0 {
				m.agent.selectedProvider--
			}
		case 1:
			if m.agent.selectedModel > 0 {
				m.agent.selectedModel--
			}
		}
	}
	return m
}

func (m Model) agentAdvanceStep() Model {
	if m.agent.formStep == 0 {
		prov := m.currentProvider()
		if prov == nil || len(nonSepModels(prov.Models)) == 0 {
			m.agent.formStep = 2
			m.agent.prompt.Focus()
		} else {
			m.agent.formStep = 1
		}
		m.agent.agentScrollOffset = 0
	} else if m.agent.formStep == 1 {
		m.agent.formStep = 2
		m.agent.prompt.Focus()
		m.agent.agentScrollOffset = 0
	}
	return m
}

func (m Model) handleEnter() (Model, tea.Cmd) {
	// Signal board: navigate directly into the job's tmux window.
	if m.signalBoardFocused {
		filtered := m.filteredFeed()
		if len(filtered) > 0 && m.signalBoard.selectedIdx < len(filtered) {
			tw := filtered[m.signalBoard.selectedIdx].tmuxWindow
			if tw != "" {
				exec.Command("tmux", "select-window", "-t", tw).Run() //nolint:errcheck
			}
		}
		return m, nil
	}

	// Launcher: launch selected pipeline.
	if m.launcher.focused {
		if len(m.launcher.pipelines) == 0 {
			return m, nil
		}
		// Enforce parallel job cap.
		if len(m.activeJobs) >= maxParallelJobs {
			feedID := fmt.Sprintf("warn-%d", time.Now().UnixNano())
			warnEntry := feedEntry{
				id:     feedID,
				title:  "warning",
				status: FeedFailed,
				ts:     time.Now(),
				lines:  []string{"max parallel jobs reached (8)"},
			}
			m.feed = append([]feedEntry{warnEntry}, m.feed...)
			if len(m.feed) > 200 {
				m.feed = m.feed[:200]
			}
			return m, nil
		}

		name := m.launcher.pipelines[m.launcher.selected]
		yamlPath := filepath.Join(pipelinesDir(), name+".pipeline.yaml")

		feedID := fmt.Sprintf("pipe-%d", time.Now().UnixNano())
		entry := feedEntry{
			id:     feedID,
			title:  "pipeline: " + name,
			status: FeedRunning,
			ts:     time.Now(),
		}
		m.feed = append([]feedEntry{entry}, m.feed...)
		if len(m.feed) > 200 {
			m.feed = m.feed[:200]
		}
		m.feedSelected = 0
		m.feedScrollOffset = 0

		// Run the pipeline directly in a background tmux window so the user
		// gets real shell history and scrollback.
		orcaiBin := orcaiBinaryPath()
		shellCmd := orcaiBin + " pipeline run " + yamlPath
		windowName, logFile, doneFile := createJobWindow(feedID, shellCmd, name)
		entry.tmuxWindow = windowName
		entry.logFile = logFile
		entry.doneFile = doneFile
		m.feed[0] = entry

		ch := make(chan tea.Msg, 256)
		_, cancel := context.WithCancel(context.Background())
		m.activeJobs[feedID] = &jobHandle{id: feedID, cancel: cancel, ch: ch, tmuxWindow: windowName, logFile: logFile}

		// Watch the log file for output; detect completion via the done file.
		startLogWatcher(feedID, logFile, doneFile, ch)
		return m, drainChan(ch)
	}

	// Agent section: advance form or submit.
	if m.agent.focused {
		if m.agent.formStep < 2 {
			m = m.agentAdvanceStep()
			return m, nil
		}
		// Step 2: submit.
		// Enforce parallel job cap.
		if len(m.activeJobs) >= maxParallelJobs {
			feedID := fmt.Sprintf("warn-%d", time.Now().UnixNano())
			warnEntry := feedEntry{
				id:     feedID,
				title:  "warning",
				status: FeedFailed,
				ts:     time.Now(),
				lines:  []string{"max parallel jobs reached (8)"},
			}
			m.feed = append([]feedEntry{warnEntry}, m.feed...)
			if len(m.feed) > 200 {
				m.feed = m.feed[:200]
			}
			return m, nil
		}
		input := strings.TrimSpace(m.agent.prompt.Value())
		if input == "" {
			return m, nil
		}
		prov := m.currentProvider()
		if prov == nil {
			return m, nil
		}

		var modelID string
		models := nonSepModels(prov.Models)
		if len(models) > 0 && m.agent.selectedModel < len(models) {
			modelID = models[m.agent.selectedModel].ID
		}

		title := "agent: " + prov.ID
		if modelID != "" {
			title += "/" + modelID
		}

		feedID := fmt.Sprintf("agent-%d", time.Now().UnixNano())
		entry := feedEntry{
			id:     feedID,
			title:  title,
			status: FeedRunning,
			ts:     time.Now(),
		}
		m.feed = append([]feedEntry{entry}, m.feed...)
		if len(m.feed) > 200 {
			m.feed = m.feed[:200]
		}
		m.feedSelected = 0
		m.feedScrollOffset = 0

		// Agent runs in-process; window shows live output via tail.
		windowName, logFile, _ := createJobWindow(feedID, "", title)
		entry.tmuxWindow = windowName
		entry.logFile = logFile
		m.feed[0] = entry

		provArgs := picker.PipelineLaunchArgs(prov.ID)
		binary := prov.Command
		if binary == "" {
			binary = prov.ID
		}
		adapter := plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, provArgs...)
		vars := map[string]string{}
		if modelID != "" {
			vars["model"] = modelID
		}

		ch := make(chan tea.Msg, 256)
		_, cancel := context.WithCancel(context.Background())
		m.activeJobs[feedID] = &jobHandle{id: feedID, cancel: cancel, ch: ch, tmuxWindow: windowName, logFile: logFile}

		cmd := runAgentCmdCh(adapter, input, vars, feedID, ch, cancel)
		drain := drainChan(ch)

		// Reset form after submission.
		m.agent.prompt.SetValue("")
		m.agent.formStep = 0
		m.agent.prompt.Blur()

		return m, tea.Batch(cmd, drain)
	}

	return m, nil
}

// runAgentCmdCh starts CliAdapter.Execute in a goroutine, streaming output to ch.
func runAgentCmdCh(adapter *plugin.CliAdapter, input string, vars map[string]string, feedID string, ch chan tea.Msg, cancel context.CancelFunc) tea.Cmd {
	return func() tea.Msg {
		defer cancel()
		w := &lineWriter{id: feedID, ch: ch}
		err := adapter.Execute(context.Background(), input, vars, w)
		w.flush()
		if err != nil {
			ch <- jobFailedMsg{id: feedID, err: err}
		} else {
			ch <- jobDoneMsg{id: feedID}
		}
		return nil
	}
}

// orcaiBinaryPath returns the path to the running orcai binary, falling back
// to a PATH lookup and finally the bare name.
func orcaiBinaryPath() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	if p, err := exec.LookPath("orcai"); err == nil {
		return p
	}
	return "orcai"
}

// drainChan returns a tea.Cmd that blocks until a message arrives on ch.
func drainChan(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// ── Feed helpers ──────────────────────────────────────────────────────────────

func (m Model) appendFeedLine(id, line string) Model {
	// Strip carriage returns — progress bars use \r for in-place updates.
	// Keep only the last "frame" so the displayed text is clean.
	if idx := strings.LastIndexByte(line, '\r'); idx >= 0 {
		line = line[idx+1:]
	}
	if line == "" {
		return m
	}
	for i := range m.feed {
		if m.feed[i].id == id {
			m.feed[i].lines = append(m.feed[i].lines, line)
			return m
		}
	}
	return m
}

func (m Model) setFeedStatus(id string, status FeedStatus) Model {
	for i := range m.feed {
		if m.feed[i].id == id {
			m.feed[i].status = status
			return m
		}
	}
	return m
}

// clampFeedScroll clamps feedScrollOffset to valid range.
func (m *Model) clampFeedScroll() {
	h := m.height
	if h <= 0 {
		h = 40
	}
	contentH := max(h-1, 5)
	sbHeight := min(len(m.feed)+2, 8)
	if sbHeight < 2 {
		sbHeight = 2
	}
	feedH := max(contentH-sbHeight, 3)
	visibleH := feedH - 2
	if visibleH <= 0 {
		visibleH = 1
	}
	total := totalFeedLines(m.feed)
	maxOffset := max(0, total-visibleH)
	if m.feedScrollOffset > maxOffset {
		m.feedScrollOffset = maxOffset
	}
	if m.feedScrollOffset < 0 {
		m.feedScrollOffset = 0
	}
}

// filteredFeed returns feed entries matching the current signal board filter.
func (m Model) filteredFeed() []feedEntry {
	filter := m.signalBoard.activeFilter
	if filter == "all" || filter == "" {
		return m.feed
	}
	var out []feedEntry
	for _, e := range m.feed {
		switch filter {
		case "running":
			if e.status == FeedRunning {
				out = append(out, e)
			}
		case "done":
			if e.status == FeedDone {
				out = append(out, e)
			}
		case "failed":
			if e.status == FeedFailed {
				out = append(out, e)
			}
		}
	}
	return out
}

func (m Model) currentProvider() *picker.ProviderDef {
	if len(m.agent.providers) == 0 {
		return nil
	}
	if m.agent.selectedProvider >= len(m.agent.providers) {
		return &m.agent.providers[0]
	}
	return &m.agent.providers[m.agent.selectedProvider]
}

// nonSepModels filters separator entries from a model list.
func nonSepModels(models []picker.ModelOption) []picker.ModelOption {
	var out []picker.ModelOption
	for _, mo := range models {
		if !mo.Separator {
			out = append(out, mo)
		}
	}
	return out
}

// ── View ──────────────────────────────────────────────────────────────────────

// View renders the full-screen switchboard layout.
func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 120
	}
	h := m.height
	if h <= 0 {
		h = 40
	}

	leftW := m.leftColWidth()
	feedW := max(w-leftW-1, 20)
	contentH := max(h-1, 5) // reserve one line for bottom bar

	// Signal board: fixed height above the feed.
	// Minimum 5 rows so the box is always visible (top+3body+bottom).
	sbHeight := min(len(m.feed)+4, 8)
	if sbHeight < 5 {
		sbHeight = 5
	}
	// Clamp sbHeight so feedH is at least 3.
	if sbHeight > contentH-3 {
		sbHeight = max(contentH-3, 5)
	}
	feedH := max(contentH-sbHeight, 3)

	left := m.viewLeftColumn(contentH, leftW)
	sb := m.buildSignalBoard(sbHeight, feedW)
	feed := m.viewActivityFeed(feedH, feedW)

	// Right column: signal board lines followed by feed lines, clipped to contentH.
	rightLines := append(sb, strings.Split(feed, "\n")...)
	if len(rightLines) > contentH {
		rightLines = rightLines[:contentH]
	}

	leftLines := strings.Split(left, "\n")
	totalRows := max(len(leftLines), len(rightLines))

	var rows []string
	for i := range totalRows {
		l := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		f := ""
		if i < len(rightLines) {
			f = rightLines[i]
		}
		rows = append(rows, padToVis(l, leftW)+" "+f)
	}

	body := strings.Join(rows, "\n")

	if m.confirmQuit {
		return body + "\n" + aPnk + aBld + " Job is running. Quit anyway? (y/N) " + aRst
	}

	return body + "\n" + m.viewBottomBar(w)
}

func (m Model) leftColWidth() int {
	w := m.width
	if w <= 0 {
		w = 120
	}
	lw := w * 30 / 100
	if lw < 28 {
		lw = 28
	}
	return lw
}

// viewLeftColumn renders the left column: banner + launcher + agent sections.
func (m Model) viewLeftColumn(height, width int) string {
	var lines []string

	banner := m.buildBanner(width)
	lines = append(lines, strings.Split(banner, "\n")...)
	lines = append(lines, "")

	lines = append(lines, m.buildLauncherSection(width)...)
	lines = append(lines, "")

	lines = append(lines, m.buildAgentSection(width)...)

	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// buildBanner renders the ORCAI BBS header banner.
func (m Model) buildBanner(w int) string {
	if w < 10 {
		w = 28
	}
	inner := w - 2
	top := aPur + "╔" + strings.Repeat("═", inner) + "╗" + aRst

	const logoPrefixLen = 14
	logoPad := max(inner-logoPrefixLen, 0)
	logoLine := aPur + "║" + aPnk + " ░▒▓ " + aBld + "ORCAI" + aRst + aPnk + " ▓▒░" +
		strings.Repeat(" ", logoPad) + aPur + "║" + aRst

	const subtitleLen = 20
	subPad := max(inner-subtitleLen, 0)
	subLine := aPur + "║" + aBrC + "  ABBS Switchboard  " +
		strings.Repeat(" ", subPad) + aPur + "║" + aRst

	bot := aPur + "╚" + strings.Repeat("═", inner) + "╝" + aRst
	return strings.Join([]string{top, logoLine, subLine, bot}, "\n")
}

// buildLauncherSection renders the Pipeline Launcher box.
func (m Model) buildLauncherSection(w int) []string {
	borderColor := aBC
	if m.launcher.focused {
		borderColor = aBrC
	}

	header := "PIPELINES"
	if n := len(m.activeJobs); n > 0 {
		header += fmt.Sprintf(" [%d running]", n)
	}
	rows := []string{boxTop(w, header, borderColor)}

	if len(m.launcher.pipelines) == 0 {
		rows = append(rows, boxRow(aDim+"  no pipelines saved"+aRst, 20, w))
	} else {
		for i, name := range m.launcher.pipelines {
			maxNameLen := max(w-4, 1)
			displayName := name
			if len(displayName) > maxNameLen {
				displayName = displayName[:maxNameLen-1] + "…"
			}
			contentVis := 2 + len(displayName)
			if i == m.launcher.selected && m.launcher.focused {
				content := aSelBg + aWht + "  " + displayName + aRst
				rows = append(rows, aBC+"│"+content+strings.Repeat(" ", max(w-2-contentVis, 0))+aBC+"│"+aRst)
			} else {
				content := aBrC + "  " + aBC + displayName + aRst
				rows = append(rows, boxRow(content, contentVis, w))
			}
		}
	}

	rows = append(rows, boxBot(w))
	return rows
}

// buildAgentSection renders the Agent Runner inline form with fixed height.
// Always returns agentInnerHeight + 2 lines (top border + body + bottom border).
func (m Model) buildAgentSection(w int) []string {
	borderColor := aBC
	if m.agent.focused {
		borderColor = aBrC
	}

	rows := []string{boxTop(w, "AGENT RUNNER", borderColor)}

	var bodyRows []string

	if len(m.agent.providers) == 0 {
		bodyRows = append(bodyRows, boxRow(aDim+"  no providers available"+aRst, 23, w))
	} else {
		switch m.agent.formStep {
		case 0:
			// Window size for provider list (full agentInnerHeight rows).
			windowSize := agentInnerHeight
			offset := m.agent.agentScrollOffset
			// Ensure selected item scrolls into view.
			if m.agent.selectedProvider < offset {
				offset = m.agent.selectedProvider
			} else if m.agent.selectedProvider >= offset+windowSize {
				offset = m.agent.selectedProvider - windowSize + 1
			}
			end := offset + windowSize
			if end > len(m.agent.providers) {
				end = len(m.agent.providers)
			}
			for i := offset; i < end; i++ {
				prov := m.agent.providers[i]
				label := prov.Label
				if label == "" {
					label = prov.ID
				}
				maxLen := max(w-5, 1)
				if len(label) > maxLen {
					label = label[:maxLen-1] + "…"
				}
				contentVis := 4 + len(label)
				if i == m.agent.selectedProvider {
					sel := ""
					if m.agent.focused {
						sel = aSelBg + aWht
					} else {
						sel = aBrC
					}
					content := sel + "  > " + label + aRst
					bodyRows = append(bodyRows, aBC+"│"+content+strings.Repeat(" ", max(w-2-contentVis, 0))+aBC+"│"+aRst)
				} else {
					content := aDim + "    " + aBC + label + aRst
					bodyRows = append(bodyRows, boxRow(content, contentVis, w))
				}
			}

		case 1:
			prov := m.currentProvider()
			var models []picker.ModelOption
			if prov != nil {
				models = nonSepModels(prov.Models)
			}
			// Breadcrumb: show which provider was selected.
			provLabel := ""
			if prov != nil {
				provLabel = prov.Label
				if provLabel == "" {
					provLabel = prov.ID
				}
			}
			crumb := "  " + aDim + provLabel + aRst + aBrC + " > model" + aRst
			bodyRows = append(bodyRows, boxRow(crumb, 2+len(provLabel)+9, w))

			// Window size for model list (agentInnerHeight - 1 row for breadcrumb).
			windowSize := agentInnerHeight - 1
			if len(models) == 0 {
				bodyRows = append(bodyRows, boxRow(aDim+"  no models"+aRst, 11, w))
			} else {
				offset := m.agent.agentScrollOffset
				if m.agent.selectedModel < offset {
					offset = m.agent.selectedModel
				} else if m.agent.selectedModel >= offset+windowSize {
					offset = m.agent.selectedModel - windowSize + 1
				}
				end := offset + windowSize
				if end > len(models) {
					end = len(models)
				}
				for i := offset; i < end; i++ {
					mo := models[i]
					label := mo.Label
					if label == "" {
						label = mo.ID
					}
					maxLen := max(w-5, 1)
					if len(label) > maxLen {
						label = label[:maxLen-1] + "…"
					}
					contentVis := 4 + len(label)
					if i == m.agent.selectedModel && m.agent.focused {
						content := aSelBg + aWht + "  > " + label + aRst
						bodyRows = append(bodyRows, aBC+"│"+content+strings.Repeat(" ", max(w-2-contentVis, 0))+aBC+"│"+aRst)
					} else {
						content := aDim + "    " + aBC + label + aRst
						bodyRows = append(bodyRows, boxRow(content, contentVis, w))
					}
				}
			}

		case 2:
			// Breadcrumb: show provider and model selection.
			prov := m.currentProvider()
			provLabel := ""
			if prov != nil {
				provLabel = prov.Label
				if provLabel == "" {
					provLabel = prov.ID
				}
			}
			models := []picker.ModelOption{}
			if prov != nil {
				models = nonSepModels(prov.Models)
			}
			modelLabel := ""
			if len(models) > 0 && m.agent.selectedModel < len(models) {
				modelLabel = models[m.agent.selectedModel].Label
				if modelLabel == "" {
					modelLabel = models[m.agent.selectedModel].ID
				}
			}
			crumb := "  " + aDim + provLabel
			if modelLabel != "" {
				crumb += " > " + modelLabel
			}
			crumb += aRst + aBrC + " > prompt" + aRst
			crumbVis := 2 + len(provLabel) + len(modelLabel) + 9
			if modelLabel != "" {
				crumbVis += 3 // " > "
			}
			bodyRows = append(bodyRows, boxRow(crumb, crumbVis, w))
			if len(m.activeJobs) > 0 {
				warn := aPnk + fmt.Sprintf("  ⚠ %d job(s) running — ctrl+c to cancel", len(m.activeJobs)) + aRst
				bodyRows = append(bodyRows, boxRow(warn, visLen(warn), w))
			} else {
				bodyRows = append(bodyRows, boxRow(aBrC+"  Prompt: (ctrl+s to submit)"+aRst, 27, w))
				for _, pLine := range strings.Split(m.agent.prompt.View(), "\n") {
					// textarea may use \r\n; strip \r to avoid cursor-reset corruption.
					pLine = strings.TrimRight(pLine, "\r")
					padded := "  " + pLine
					bodyRows = append(bodyRows, boxRow(padded, visLen(padded), w))
				}
			}
		}
	}

	// Pad or clip body rows to exactly agentInnerHeight.
	for len(bodyRows) < agentInnerHeight {
		bodyRows = append(bodyRows, boxRow("", 0, w))
	}
	if len(bodyRows) > agentInnerHeight {
		bodyRows = bodyRows[:agentInnerHeight]
	}

	rows = append(rows, bodyRows...)
	rows = append(rows, boxBot(w))
	return rows
}

// totalFeedLines computes the total number of content lines for a feed (not counting borders).
func totalFeedLines(feed []feedEntry) int {
	n := 0
	for _, entry := range feed {
		n++ // title line
		n += len(entry.lines)
	}
	return n
}

// viewActivityFeed renders the center activity feed with scroll support.
func (m Model) viewActivityFeed(height, width int) string {
	visibleH := height - 2 // minus top and bottom border

	// Flatten all feed entries into content lines.
	var allLines []string
	if len(m.feed) == 0 {
		allLines = append(allLines, boxRow(aDim+"  no activity yet"+aRst, 17, width))
	} else {
		for _, entry := range m.feed {
			badge, badgeColor := statusBadge(entry.status)
			ts := entry.ts.Format("15:04:05")
			titleLine := fmt.Sprintf("  %s%s%s %s%s%s  %s",
				badgeColor, badge, aRst,
				aDim, ts, aRst,
				aBrC+entry.title+aRst)
			allLines = append(allLines, boxRow(titleLine, visLen(titleLine), width))

			// Cap output lines per entry: show the last feedLinesPerEntry lines only.
			const feedLinesPerEntry = 10
			entryLines := entry.lines
			skipped := 0
			if len(entryLines) > feedLinesPerEntry {
				skipped = len(entryLines) - feedLinesPerEntry
				entryLines = entryLines[skipped:]
			}
			if skipped > 0 {
				skipMsg := fmt.Sprintf("    … %d earlier lines (press f to scroll)", skipped)
				skipFull := aDim + skipMsg + aRst
				allLines = append(allLines, boxRow(skipFull, visLen(skipFull), width))
			}
			for _, outLine := range entryLines {
				// Strip ANSI codes — feed renders with its own dim style.
				outLine = stripANSI(outLine)
				maxLen := max(width-6, 1)
				if len(outLine) > maxLen {
					outLine = outLine[:maxLen-1] + "…"
				}
				content := aDim + "    " + outLine + aRst
				allLines = append(allLines, boxRow(content, visLen(content), width))
			}
		}
	}

	// Clamp offset and slice visible window.
	offset := m.feedScrollOffset
	total := len(allLines)
	if visibleH <= 0 {
		visibleH = 1
	}
	maxOffset := max(0, total-visibleH)
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + visibleH
	if end > total {
		end = total
	}
	visible := allLines[offset:end]

	var lines []string
	borderColor := aBC
	if m.feedFocused {
		borderColor = aBrC
	}
	lines = append(lines, boxTop(width, "ACTIVITY FEED", borderColor))
	lines = append(lines, visible...)

	// Pad to fill the box body.
	for len(lines) < height-1 {
		lines = append(lines, boxRow("", 0, width))
	}
	lines = append(lines, boxBot(width))

	// Trim to exact height.
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// viewBottomBar renders the one-line keybinding hint strip.
func (m Model) viewBottomBar(width int) string {
	hint := func(key, desc string) string {
		return aBrC + key + aDim + " " + desc + aRst
	}
	sep := aDim + " · " + aRst

	var parts []string
	switch {
	case m.signalBoardFocused:
		parts = []string{
			hint("↑↓", "nav"),
			hint("f", "filter"),
			hint("enter", "go to window"),
			hint("tab", "focus"),
			hint("q", "quit"),
		}
	case m.feedFocused:
		parts = []string{
			hint("↑↓", "scroll"),
			hint("tab", "focus"),
			hint("q", "quit"),
		}
	default:
		parts = []string{
			hint("enter", "launch"),
			hint("ctrl+s", "submit"),
			hint("tab", "focus"),
			hint("a", "agent"),
			hint("s", "signals"),
			hint("f", "feed"),
			hint("r", "refresh"),
			hint("↑↓", "nav"),
			hint("q", "quit"),
		}
	}

	bar := "  " + strings.Join(parts, sep)
	if visLen(bar) < width {
		bar += strings.Repeat(" ", width-visLen(bar))
	}
	return bar + aRst
}

// ── Box drawing helpers ────────────────────────────────────────────────────────

func boxTop(w int, title, borderColor string) string {
	if title == "" {
		return borderColor + "┌" + strings.Repeat("─", max(w-2, 0)) + "┐" + aRst
	}
	label := " " + title + " "
	dashes := max(w-2-len(label), 0)
	left := dashes / 2
	right := dashes - left
	return borderColor + "┌" + strings.Repeat("─", left) + aBrC + label + borderColor + strings.Repeat("─", right) + "┐" + aRst
}

func boxBot(w int) string {
	return aBC + "└" + strings.Repeat("─", max(w-2, 0)) + "┘" + aRst
}

func boxRow(content string, contentVis, w int) string {
	inner := w - 2
	pad := max(inner-contentVis, 0)
	return aBC + "│" + aRst + content + strings.Repeat(" ", pad) + aBC + "│" + aRst
}

func statusBadge(s FeedStatus) (string, string) {
	switch s {
	case FeedRunning:
		return "▶ running", aPnk
	case FeedDone:
		return "✓ done", aGrn
	case FeedFailed:
		return "✗ failed", aRed
	default:
		return "? unknown", aDim
	}
}

// visLen returns the number of visible terminal columns in s.
// Handles CSI sequences (\x1b[...X), skips all control characters.
func visLen(s string) int {
	n := 0
	esc := false
	for _, r := range s {
		if r == '\x1b' {
			esc = true
			continue
		}
		if esc {
			// CSI/ESC sequences end on any ASCII letter.
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				esc = false
			}
			continue
		}
		// Skip control characters (CR, LF, tab, etc.) — not visible columns.
		if r < 0x20 || r == 0x7f {
			continue
		}
		n++
	}
	return n
}

// padToVis right-pads s with spaces until its visible length equals w.
func padToVis(s string, w int) string {
	vl := visLen(s)
	if vl >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vl)
}

// ── Run ───────────────────────────────────────────────────────────────────────

// Run starts the Switchboard as a full-screen BubbleTea program.
func Run() {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("switchboard error: %v\n", err)
	}
}

// RunToggle opens the switchboard as a tmux popup.
func RunToggle() {
	bin := resolveSwitchboardBin()
	exec.Command("tmux", "display-popup", "-E", "-w", "100%", "-h", "100%", bin).Run() //nolint:errcheck
}

func resolveSwitchboardBin() string {
	if bin, err := exec.LookPath("orcai-sysop"); err == nil {
		return bin
	}
	self, _ := os.Executable()
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}
	return filepath.Join(filepath.Dir(self), "orcai-sysop")
}

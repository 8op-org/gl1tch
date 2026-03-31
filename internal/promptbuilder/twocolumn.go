package promptbuilder

// TwoColumnModel is the new two-column prompt builder TUI.
// Layout:
//   Left  column : Sidebar (saved prompts)
//   Right column : EditorPanel (provider/name/content) + RunnerPanel + ChatInput

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/adam-stokes/orcai/internal/buildershared"
	"github.com/adam-stokes/orcai/internal/busd/topics"
	"github.com/adam-stokes/orcai/internal/panelrender"
	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/pipeline"
	"github.com/adam-stokes/orcai/internal/plugin"
	"github.com/adam-stokes/orcai/internal/styles"
)

// tcFocus constants for TwoColumnModel outer focus.
const (
	tcFocusSidebar = 0
	tcFocusEditor  = 1
	tcFocusRunner  = 2
	tcFocusChat    = 3
)

// TwoColumnModel implements tea.Model for the prompt builder TUI.
type TwoColumnModel struct {
	sidebar   buildershared.Sidebar
	editor    buildershared.EditorPanel
	runner    buildershared.RunnerPanel
	chatInput textinput.Model

	focus int

	// Persistence
	promptsDir string
	pluginMgr  *plugin.Manager

	// Feedback loop
	firstPrompt string
	sentOnce    bool

	// Status
	statusMsg string
	statusErr bool

	width, height int
	pal           styles.ANSIPalette
}

// promptEntry is stored on disk as JSON.
type promptEntry struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Content  string `json:"content"`
}

// NewTwoColumn creates a new TwoColumnModel.
func NewTwoColumn(promptsDir string, providers []picker.ProviderDef, mgr *plugin.Manager) *TwoColumnModel {
	chatIn := textinput.New()
	chatIn.Placeholder = "send a message to run against this prompt…"
	chatIn.CharLimit = 4000

	pal := styles.ANSIPalette{
		Accent:  "\x1b[35m",
		Dim:     "\x1b[2m",
		Success: "\x1b[32m",
		Error:   "\x1b[31m",
		Warn:    "\x1b[33m",
		FG:      "\x1b[97m",
		BG:      "\x1b[40m",
		Border:  "\x1b[36m",
		SelBG:   "\x1b[44m",
	}

	m := &TwoColumnModel{
		sidebar:    buildershared.NewSidebar("PROMPTS", nil),
		editor:     buildershared.NewEditorPanel(providers),
		runner:     buildershared.NewRunnerPanel(),
		chatInput:  chatIn,
		promptsDir: promptsDir,
		pluginMgr:  mgr,
		pal:        pal,
	}
	m.sidebar = m.sidebar.SetItems(m.loadPromptNames())
	m.sidebar = m.sidebar.SetFocused(true)
	return m
}

// SetPalette updates the color palette.
func (m *TwoColumnModel) SetPalette(pal styles.ANSIPalette) {
	m.pal = pal
}

// Init implements tea.Model.
func (m *TwoColumnModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *TwoColumnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		return m, nil

	case buildershared.RunLineMsg, buildershared.RunDoneMsg:
		var cmd tea.Cmd
		m.runner, cmd = m.runner.Update(v)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(v)
	}
	return m, nil
}

func (m *TwoColumnModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys.
	switch key {
	case "J":
		// Open jump window as a tmux popup.
		if os.Getenv("TMUX") != "" {
			return m, func() tea.Msg {
				self, _ := os.Executable()
				exec.Command("tmux", "display-popup", "-E", "-w", "80%", "-h", "70%",
					filepath.Clean(self)+" widget jump-window").Run() //nolint:errcheck
				return nil
			}
		}
		return m, nil

	case "ctrl+s":
		if err := m.saveCurrentPrompt(); err != nil {
			m.statusMsg = "save error: " + err.Error()
			m.statusErr = true
		} else {
			m.statusMsg = "saved"
			m.statusErr = false
			m.sidebar = m.sidebar.SetItems(m.loadPromptNames())
		}
		return m, nil

	case "ctrl+r":
		if m.firstPrompt != "" {
			m.runner = m.runner.Clear()
			return m, m.startRun(m.firstPrompt)
		}
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}

	switch m.focus {
	case tcFocusSidebar:
		return m.handleSidebarKey(msg)
	case tcFocusEditor:
		return m.handleEditorKey(msg)
	case tcFocusRunner:
		return m.handleRunnerKey(msg)
	case tcFocusChat:
		return m.handleChatKey(msg)
	}
	return m, nil
}

func (m *TwoColumnModel) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)

	if cmd != nil {
		innerMsg := cmd()
		switch v := innerMsg.(type) {
		case buildershared.SidebarSelectMsg:
			m.loadPromptIntoEditor(v.Name)
			m.sidebar = m.sidebar.SetFocused(false)
			m.editor = m.editor.SetFocused(true)
			m.focus = tcFocusEditor
			return m, nil
		case buildershared.SidebarDeleteMsg:
			os.Remove(filepath.Join(m.promptsDir, v.Name+".json")) //nolint:errcheck
			m.sidebar = m.sidebar.SetItems(m.loadPromptNames())
			return m, nil
		}
	}

	if key == "n" {
		m.editor = m.editor.SetName("new-prompt")
		m.editor = m.editor.SetContent("")
		m.sidebar = m.sidebar.SetFocused(false)
		m.editor = m.editor.SetFocused(true)
		m.focus = tcFocusEditor
		return m, nil
	}

	if key == "tab" {
		m.sidebar = m.sidebar.SetFocused(false)
		m.editor = m.editor.SetFocused(true)
		m.focus = tcFocusEditor
		return m, nil
	}

	return m, nil
}

func (m *TwoColumnModel) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)

	if cmd != nil {
		innerMsg := cmd()
		switch innerMsg.(type) {
		case buildershared.EditorTabOutMsg:
			m.editor = m.editor.SetFocused(false)
			m.focus = tcFocusChat
			m.chatInput.Focus()
			return m, nil
		case buildershared.EditorShiftTabOutMsg:
			m.editor = m.editor.SetFocused(false)
			m.sidebar = m.sidebar.SetFocused(true)
			m.focus = tcFocusSidebar
			return m, nil
		}
	}

	if key == "esc" {
		m.editor = m.editor.SetFocused(false)
		m.sidebar = m.sidebar.SetFocused(true)
		m.focus = tcFocusSidebar
		return m, nil
	}

	return m, cmd
}

func (m *TwoColumnModel) handleRunnerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	var cmd tea.Cmd
	m.runner, cmd = m.runner.Update(msg)

	switch key {
	case "shift+tab", "esc":
		m.runner = m.runner.SetFocused(false)
		m.editor = m.editor.SetFocused(true)
		m.focus = tcFocusEditor
	case "tab":
		m.runner = m.runner.SetFocused(false)
		m.focus = tcFocusChat
		m.chatInput.Focus()
	}
	return m, cmd
}

func (m *TwoColumnModel) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		prompt := strings.TrimSpace(m.chatInput.Value())
		m.chatInput.SetValue("")
		if prompt == "" {
			return m, nil
		}
		if !m.sentOnce {
			m.firstPrompt = prompt
			m.sentOnce = true
		}
		m.runner = m.runner.Clear()
		m.runner = m.runner.SetFocused(true)
		m.focus = tcFocusRunner
		return m, m.startRun(prompt)

	case "shift+tab", "esc":
		m.chatInput.Blur()
		m.editor = m.editor.SetFocused(true)
		m.focus = tcFocusEditor
		return m, nil

	case "tab":
		m.chatInput.Blur()
		m.sidebar = m.sidebar.SetFocused(true)
		m.focus = tcFocusSidebar
		return m, nil
	}

	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m *TwoColumnModel) View() string {
	w, h := m.width, m.height
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 40
	}

	pal := m.pal
	leftW := w / 4
	if leftW < 20 {
		leftW = 20
	}
	rightW := w - leftW
	bodyH := h - 1

	topBar := pal.Accent + "\x1b[1m PROMPT BUILDER\x1b[0m"
	if m.statusMsg != "" {
		sep := pal.Dim + " · \x1b[0m"
		if m.statusErr {
			topBar += sep + pal.Error + m.statusMsg + "\x1b[0m"
		} else {
			topBar += sep + pal.Success + m.statusMsg + "\x1b[0m"
		}
	}

	leftLines := m.sidebar.SetFocused(m.focus == tcFocusSidebar).View(leftW, bodyH, pal)
	rightLines := m.buildRight(rightW, bodyH)

	var rows []string
	rows = append(rows, topBar)
	for i := range bodyH {
		var l, r string
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		lv := lipgloss.Width(l)
		if lv < leftW {
			l = l + strings.Repeat(" ", leftW-lv)
		}
		rows = append(rows, l+r)
	}
	return strings.Join(rows, "\n")
}

func (m *TwoColumnModel) buildRight(w, h int) []string {
	chatH := 4
	if h < 20 {
		chatH = 3
	}
	remaining := h - chatH
	editorH := remaining * 55 / 100
	if editorH < 10 {
		editorH = 10
	}
	runnerH := remaining - editorH
	if runnerH < 5 {
		runnerH = 5
	}

	ed := m.editor.SetFocused(m.focus == tcFocusEditor)
	rn := m.runner.SetFocused(m.focus == tcFocusRunner)

	var rows []string
	rows = append(rows, ed.View(w, editorH, m.pal)...)
	rows = append(rows, rn.View(w, runnerH, m.pal)...)
	rows = append(rows, m.buildChatBox(w, chatH)...)
	return rows
}

func (m *TwoColumnModel) buildChatBox(w, h int) []string {
	pal := m.pal
	borderColor := pal.Border
	if m.focus == tcFocusChat {
		borderColor = pal.Accent
	}

	var rows []string
	rows = append(rows, panelrender.BoxTop(w, "SEND", borderColor, pal.Accent))

	m.chatInput.Width = w - 6
	if m.chatInput.Width < 10 {
		m.chatInput.Width = 10
	}
	rows = append(rows, panelrender.BoxRow("  "+m.chatInput.View(), w, borderColor))

	for len(rows) < h-2 {
		rows = append(rows, panelrender.BoxRow("", w, borderColor))
	}

	hints := []panelrender.Hint{
		{Key: "enter", Desc: "send"},
		{Key: "ctrl+r", Desc: "re-run"},
		{Key: "ctrl+s", Desc: "save"},
		{Key: "shift+tab", Desc: "editor"},
	}
	rows = append(rows, panelrender.BoxRow(panelrender.HintBar(hints, w-2, pal), w, borderColor))
	rows = append(rows, panelrender.BoxBot(w, borderColor))
	return rows
}

// ── Persistence ───────────────────────────────────────────────────────────────

func (m *TwoColumnModel) loadPromptNames() []string {
	entries, err := os.ReadDir(m.promptsDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".json") {
			names = append(names, strings.TrimSuffix(n, ".json"))
		}
	}
	return names
}

func (m *TwoColumnModel) loadPromptIntoEditor(name string) {
	data, err := os.ReadFile(filepath.Join(m.promptsDir, name+".json"))
	if err != nil {
		return
	}
	var p promptEntry
	if err := json.Unmarshal(data, &p); err != nil {
		return
	}
	m.editor = m.editor.SetName(p.Name)
	m.editor = m.editor.SetContent(p.Content)
	if p.Provider != "" || p.Model != "" {
		m.editor = m.editor.SelectBySlug(p.Provider + "/" + p.Model)
	}
}

func (m *TwoColumnModel) saveCurrentPrompt() error {
	name := strings.TrimSpace(m.editor.Name())
	if name == "" {
		return fmt.Errorf("prompt name is required")
	}
	if err := os.MkdirAll(m.promptsDir, 0o755); err != nil {
		return err
	}
	p := promptEntry{
		Name:     name,
		Provider: m.editor.SelectedProviderID(),
		Model:    m.editor.SelectedModelID(),
		Content:  m.editor.Content(),
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.promptsDir, name+".json"), data, 0o644)
}

// ── Run logic ─────────────────────────────────────────────────────────────────

func (m *TwoColumnModel) startRun(userMsg string) tea.Cmd {
	content := strings.TrimSpace(m.editor.Content())
	if content == "" && userMsg == "" {
		return nil
	}

	executorID := m.editor.SelectedProviderID()
	if executorID == "" {
		executorID = "claude"
	}
	modelID := m.editor.SelectedModelID()

	// Build a single-step pipeline: system=content, user=userMsg.
	fullPrompt := content
	if userMsg != "" && content != "" {
		fullPrompt = content + "\n\n" + userMsg
	} else if userMsg != "" {
		fullPrompt = userMsg
	}

	yamlContent := buildPromptYAML("run-prompt", executorID, modelID, fullPrompt)

	ch := make(chan string, 200)
	ctx, cancel := context.WithCancel(context.Background())

	mgr := m.pluginMgr
	providers := picker.BuildProviders()

	go func() {
		defer close(ch)

		if mgr == nil {
			var err error
			mgr, err = buildPromptPluginManager(providers)
			if err != nil {
				ch <- "error: " + err.Error()
				return
			}
		}

		p, err := pipeline.Load(strings.NewReader(yamlContent))
		if err != nil {
			ch <- "error: " + err.Error()
			return
		}

		pub := &tcLinePublisher{ch: ch}
		_, runErr := pipeline.Run(ctx, p, mgr, "", pipeline.WithEventPublisher(pub))
		if runErr != nil {
			if ctx.Err() != nil {
				ch <- "cancelled"
			} else {
				ch <- "error: " + runErr.Error()
			}
		}
	}()

	m.runner, _ = m.runner.StartRun(ch, cancel)
	return buildershared.WaitForLine(ch)
}

func buildPromptYAML(name, executorID, modelID, prompt string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("name: %s\nversion: \"1\"\nsteps:\n", name))
	sb.WriteString(fmt.Sprintf("  - id: run\n    executor: %s\n", executorID))
	if modelID != "" {
		sb.WriteString(fmt.Sprintf("    model: %s\n", modelID))
	}
	if prompt != "" {
		sb.WriteString("    prompt: |\n")
		for _, line := range strings.Split(prompt, "\n") {
			sb.WriteString("      " + line + "\n")
		}
	}
	return sb.String()
}

func buildPromptPluginManager(providers []picker.ProviderDef) (*plugin.Manager, error) {
	mgr := plugin.NewManager()
	for _, prov := range providers {
		if prov.SidecarPath != "" {
			continue
		}
		binary := prov.Command
		if binary == "" {
			binary = prov.ID
		}
		_ = mgr.Register(plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...))
	}
	configDir := picker.OrcaiConfigDir()
	if configDir != "" {
		_ = mgr.LoadWrappersFromDir(filepath.Join(configDir, "wrappers"))
	}
	return mgr, nil
}

// tcLinePublisher implements pipeline.EventPublisher for the TwoColumnModel.
type tcLinePublisher struct {
	ch chan<- string
}

func (p *tcLinePublisher) Publish(_ context.Context, topic string, payload []byte) error {
	switch topic {
	case topics.StepDone, topics.StepFailed:
		var evt struct {
			Output string `json:"output"`
			StepID string `json:"step_id"`
		}
		if err := json.Unmarshal(payload, &evt); err == nil {
			if topic == topics.StepFailed {
				p.ch <- fmt.Sprintf("[fail] %s", evt.StepID)
			}
			if evt.Output != "" {
				for _, line := range strings.Split(evt.Output, "\n") {
					line = strings.TrimRight(line, "\r")
					if line != "" {
						p.ch <- line
					}
				}
			}
		}
	case topics.StepStarted:
		var evt struct {
			StepID string `json:"step_id"`
		}
		if err := json.Unmarshal(payload, &evt); err == nil && evt.StepID != "" {
			p.ch <- fmt.Sprintf("[running via %s…]", evt.StepID)
		}
	case topics.RunCompleted:
		p.ch <- "[done]"
	case topics.RunFailed:
		var evt struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(payload, &evt); err == nil && evt.Error != "" {
			p.ch <- "[fail] " + evt.Error
		} else {
			p.ch <- "[fail] run failed"
		}
	}
	return nil
}

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/pipeline"
	"github.com/adam-stokes/orcai/internal/pipelineeditor"
	"github.com/adam-stokes/orcai/internal/plugin"
	"github.com/adam-stokes/orcai/internal/promptbuilder"
	"github.com/adam-stokes/orcai/internal/store"
	"github.com/adam-stokes/orcai/internal/styles"
	"github.com/adam-stokes/orcai/internal/themes"
)

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.AddCommand(pipelineBuildCmd)
	pipelineCmd.AddCommand(pipelineRunCmd)
	pipelineRunCmd.Flags().StringVar(&pipelineRunInput, "input", "", "user input passed to the pipeline as {{param.input}}")
	pipelineCmd.AddCommand(pipelineResumeCmd)
	pipelineResumeCmd.Flags().Int64Var(&pipelineResumeRunID, "run-id", 0, "Store run ID to resume")
	_ = pipelineResumeCmd.MarkFlagRequired("run-id")

	rootCmd.AddCommand(pipelineBuilderCmd)
	pipelineBuilderCmd.AddCommand(pipelineBuilderStartCmd)
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage and run AI pipelines",
}

var pipelineBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Open the interactive pipeline builder",
	RunE: func(cmd *cobra.Command, args []string) error {
		providers := picker.BuildProviders()

		mgr := plugin.NewManager()
		for _, prov := range providers {
			mgr.Register(plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", prov.ID, prov.PipelineArgs...))
		}

		m := promptbuilder.New(mgr)
		m.SetName("new-pipeline")
		m.AddStep(pipeline.Step{ID: "input", Type: "input", Prompt: "Enter your prompt:"})
		m.AddStep(pipeline.Step{ID: "step1", Executor: "claude", Model: "claude-sonnet-4-6"})
		m.AddStep(pipeline.Step{ID: "output", Type: "output"})

		bubble := promptbuilder.NewBubble(m, providers)
		prog := tea.NewProgram(bubble, tea.WithAltScreen())
		if _, err := prog.Run(); err != nil {
			return fmt.Errorf("pipeline builder: %w", err)
		}
		return nil
	},
}

var pipelineRunCmd = &cobra.Command{
	Use:   "run <name|file>",
	Short: "Run a saved pipeline by name or file path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := args[0]

		// Accept either an absolute/relative file path or a bare name.
		// If the arg contains a path separator or ends in .yaml, treat it as a file path.
		var yamlPath string
		if strings.Contains(arg, string(filepath.Separator)) || strings.HasSuffix(arg, ".yaml") {
			yamlPath = arg
		} else {
			configDir, err := orcaiConfigDir()
			if err != nil {
				return err
			}
			yamlPath = filepath.Join(configDir, "pipelines", arg+".pipeline.yaml")
		}

		f, err := os.Open(yamlPath)
		if err != nil {
			return fmt.Errorf("pipeline %q not found: %w", arg, err)
		}
		defer f.Close()

		p, err := pipeline.Load(f)
		if err != nil {
			return err
		}

		fmt.Printf("[pipeline] starting: %s\n", p.Name)

		runProviders := picker.BuildProviders()
		mgr := plugin.NewManager()
		for _, prov := range runProviders {
			// Sidecar-backed providers are fully registered by LoadWrappersFromDir below.
			if prov.SidecarPath != "" {
				continue
			}
			binary := prov.Command
			if binary == "" {
				binary = prov.ID
			}
			if err := mgr.Register(plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...)); err != nil {
				fmt.Fprintf(os.Stderr, "pipeline: register provider %q: %v\n", prov.ID, err)
			}
		}

		// Load sidecar plugins from ~/.config/orcai/wrappers/.
		wrappersConfigDir, _ := orcaiConfigDir()
		if wrappersConfigDir != "" {
			wrappersDir := filepath.Join(wrappersConfigDir, "wrappers")
			if errs := mgr.LoadWrappersFromDir(wrappersDir); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "pipeline: sidecar load warning: %v\n", e)
				}
			}
		}

		// Open the result store so this run is recorded in the inbox.
		// A failure to open the store is non-fatal — the pipeline still runs.
		var storeOpts []pipeline.RunOption
		if s, serr := store.Open(); serr == nil {
			defer s.Close()
			storeOpts = append(storeOpts, pipeline.WithRunStore(s))
			// Wire brain context injection: use_brain / write_brain flags on pipeline
			// steps will prepend DB context and parse <brain> notes from responses.
			storeOpts = append(storeOpts, pipeline.WithBrainInjector(pipeline.NewStoreBrainInjector(s)))
		} else {
			fmt.Fprintf(os.Stderr, "pipeline: store unavailable: %v\n", serr)
		}

		// Wire busd publisher if the daemon is reachable.
		if pub := newBusPublisher(); pub != nil {
			storeOpts = append(storeOpts, pipeline.WithEventPublisher(pub))
		}

		result, err := pipeline.Run(cmd.Context(), p, mgr, pipelineRunInput, storeOpts...)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func orcaiConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "orcai"), nil
}

var pipelineRunInput string

var pipelineResumeRunID int64

var pipelineResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a pipeline that is paused waiting for a clarification answer",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open()
		if err != nil {
			return fmt.Errorf("resume: open store: %w", err)
		}
		defer s.Close()

		runIDStr := strconv.FormatInt(pipelineResumeRunID, 10)
		clarif, err := s.LoadClarificationForRun(runIDStr)
		if err != nil {
			return fmt.Errorf("resume: load clarification: %w", err)
		}
		if clarif == nil {
			return fmt.Errorf("resume: no pending clarification found for run %d", pipelineResumeRunID)
		}
		if clarif.Answer == "" {
			return fmt.Errorf("resume: clarification for run %d has no answer yet", pipelineResumeRunID)
		}

		run, err := s.GetRun(pipelineResumeRunID)
		if err != nil {
			return fmt.Errorf("resume: load run: %w", err)
		}

		// Parse pipeline_file and cwd from run metadata.
		type runMeta struct {
			PipelineFile string `json:"pipeline_file"`
			CWD          string `json:"cwd"`
		}
		var meta runMeta
		_ = json.Unmarshal([]byte(run.Metadata), &meta)
		if meta.PipelineFile == "" {
			return fmt.Errorf("resume: run %d has no pipeline_file in metadata", pipelineResumeRunID)
		}

		f, err := os.Open(meta.PipelineFile)
		if err != nil {
			return fmt.Errorf("resume: open pipeline %q: %w", meta.PipelineFile, err)
		}
		defer f.Close()

		p, err := pipeline.Load(f)
		if err != nil {
			return fmt.Errorf("resume: load pipeline: %w", err)
		}

		// Build the executor manager (same as pipelineRunCmd).
		runProviders := picker.BuildProviders()
		mgr := plugin.NewManager()
		for _, prov := range runProviders {
			if prov.SidecarPath != "" {
				continue
			}
			binary := prov.Command
			if binary == "" {
				binary = prov.ID
			}
			if err := mgr.Register(plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...)); err != nil {
				fmt.Fprintf(os.Stderr, "resume: register provider %q: %v\n", prov.ID, err)
			}
		}
		wrappersConfigDir, _ := orcaiConfigDir()
		if wrappersConfigDir != "" {
			if errs := mgr.LoadWrappersFromDir(filepath.Join(wrappersConfigDir, "wrappers")); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "resume: sidecar load warning: %v\n", e)
				}
			}
		}

		followUp := pipeline.BuildClarificationFollowUp(clarif.Output, clarif.Answer)

		// Delete clarification before running to prevent re-entrant resumes.
		_ = s.DeleteClarification(runIDStr)

		var runOpts []pipeline.RunOption
		runOpts = append(runOpts, pipeline.WithRunStore(s))
		runOpts = append(runOpts, pipeline.WithResumeFrom(pipelineResumeRunID, clarif.StepID, followUp))
		if pub := newBusPublisher(); pub != nil {
			runOpts = append(runOpts, pipeline.WithEventPublisher(pub))
		}

		fmt.Printf("[pipeline] resuming run %d from step %q\n", pipelineResumeRunID, clarif.StepID)
		result, err := pipeline.Run(cmd.Context(), p, mgr, "", runOpts...)
		if err != nil {
			return err
		}

		fmt.Println(result)
		return nil
	},
}

// pipelineBuilderCmd opens the two-column pipeline builder TUI.
var pipelineBuilderCmd = &cobra.Command{
	Use:   "pipeline-builder",
	Short: "Open the pipeline builder TUI (two-column layout)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If inside tmux and not already in the pipeline-builder window, open a new window.
		if os.Getenv("TMUX") != "" {
			out, _ := exec.Command("tmux", "display-message", "-p", "#W").Output()
			if strings.TrimSpace(string(out)) != "orcai-pipeline-builder" {
				self, err := os.Executable()
				if err != nil {
					return fmt.Errorf("pipeline-builder: resolve executable: %w", err)
				}
				return exec.Command("tmux", "new-window", "-n", "orcai-pipeline-builder",
					filepath.Clean(self)+" pipeline-builder").Run()
			}
		}

		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "pipeline-builder: panic: %v\n", r)
				os.Exit(2)
			}
		}()

		home, _ := os.UserHomeDir()
		userThemesDir := filepath.Join(home, ".config", "orcai", "themes")

		var pal styles.ANSIPalette
		if reg, err := themes.NewRegistry(userThemesDir); err == nil {
			themes.SetGlobalRegistry(reg)
			if bundle := reg.Active(); bundle != nil {
				pal = styles.BundleANSI(bundle)
			}
		}

		providers := picker.BuildProviders()

		st, err := store.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pipeline-builder: store unavailable: %v\n", err)
			st = nil
		} else {
			defer st.Close()
		}

		configDir, _ := orcaiConfigDir()
		pipelinesDir := filepath.Join(configDir, "pipelines")

		m := pipelineeditor.New(providers, pipelinesDir, st)
		if pal.Accent != "" {
			m.SetPalette(pal)
		}

		prog := tea.NewProgram(pipelineBuilderTeaModel{m: m}, tea.WithAltScreen())
		_, err = prog.Run()
		return err
	},
}

// pipelineBuilderTeaModel wraps pipelineeditor.Model to satisfy tea.Model.
type pipelineBuilderTeaModel struct {
	m      pipelineeditor.Model
	width  int
	height int
}

func (t pipelineBuilderTeaModel) Init() tea.Cmd { return t.m.Init() }

func (t pipelineBuilderTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width, t.height = msg.Width, msg.Height
		t.m.SetSize(msg.Width, msg.Height)
		return t, nil
	case pipelineeditor.CloseMsg:
		// Editor requested close — quit the program.
		return t, tea.Quit
	case tea.KeyMsg:
		updated, cmd := t.m.HandleKey(msg)
		t.m = updated
		return t, cmd
	default:
		updated, cmd := t.m.HandleMsg(msg)
		t.m = updated
		return t, cmd
	}
}

func (t pipelineBuilderTeaModel) View() string {
	return t.m.View(t.width, t.height)
}

// pipelineBuilderStartCmd opens pipeline-builder in a new tmux window.
var pipelineBuilderStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Open pipeline-builder in a new tmux window",
	RunE: func(cmd *cobra.Command, args []string) error {
		self, err := os.Executable()
		if err != nil {
			return fmt.Errorf("pipeline-builder: resolve executable: %w", err)
		}
		self = filepath.Clean(self)
		newArgs := []string{
			"new-window", "-n", "orcai-pipeline-builder",
			self + " pipeline-builder",
		}
		if err := exec.Command("tmux", newArgs...).Run(); err != nil {
			return fmt.Errorf("pipeline-builder: open window: %w", err)
		}
		return nil
	},
}

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/executor"
	"github.com/8op-org/gl1tch/internal/picker"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/store"
)

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.AddCommand(pipelineRunCmd)
	pipelineRunCmd.Flags().StringVar(&pipelineRunInput, "input", "", "user input passed to the pipeline as {{param.input}}")
	pipelineCmd.AddCommand(pipelineResumeCmd)
	pipelineResumeCmd.Flags().Int64Var(&pipelineResumeRunID, "run-id", 0, "Store run ID to resume")
	_ = pipelineResumeCmd.MarkFlagRequired("run-id")
	pipelineCmd.AddCommand(pipelineResultsCmd)
	pipelineResultsCmd.Flags().IntVar(&pipelineResultsLimit, "limit", 1, "number of matching runs to show")
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage and run AI pipelines",
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
			configDir, err := glitchConfigDir()
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

		if os.Getenv("FORCE_COLOR") == "1" {
			toANSI := func(envKey, fallback string) string {
				hex := os.Getenv(envKey)
				if len(hex) != 6 {
					hex = fallback
				}
				var r, g, b uint64
				fmt.Sscanf(hex[0:2], "%x", &r)
				fmt.Sscanf(hex[2:4], "%x", &g)
				fmt.Sscanf(hex[4:6], "%x", &b)
				return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
			}
			dim := toANSI("GLITCH_COL_DIM", "6272a4")
			accent := toANSI("GLITCH_COL_ACCENT", "bd93f9")
			reset := "\033[0m"
			fmt.Printf("%s[pipeline]%s starting: %s%s%s\n", dim, reset, accent, p.Name, reset)
		} else {
			fmt.Printf("[pipeline] starting: %s\n", p.Name)
		}

		runProviders := picker.BuildProviders()
		mgr := executor.NewManager()
		for _, prov := range runProviders {
			// Sidecar-backed providers are fully registered by LoadWrappersFromDir below.
			if prov.SidecarPath != "" {
				continue
			}
			binary := prov.Command
			if binary == "" {
				binary = prov.ID
			}
			if err := mgr.Register(executor.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...)); err != nil {
				fmt.Fprintf(os.Stderr, "pipeline: register provider %q: %v\n", prov.ID, err)
			}
		}

		// Load sidecar plugins from ~/.config/glitch/wrappers/.
		wrappersConfigDir, _ := glitchConfigDir()
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

func glitchConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "glitch"), nil
}

var pipelineRunInput string

var pipelineResumeRunID int64

var pipelineResultsLimit int

// extractStreamResult scans NDJSON stdout from a Claude/Ollama executor run and
// returns the human-readable result text. Falls back to raw stdout if no
// stream-json result line is found (e.g. shell executor output).
func extractStreamResult(stdout string) string {
	if stdout == "" {
		return ""
	}
	// Fast path: if the output doesn't look like NDJSON, return as-is.
	if len(stdout) == 0 || stdout[0] != '{' {
		return stdout
	}
	type resultLine struct {
		Type   string `json:"type"`
		Result string `json:"result"`
	}
	var last string
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		var r resultLine
		if err := json.Unmarshal([]byte(line), &r); err == nil && r.Type == "result" && r.Result != "" {
			last = r.Result
		}
	}
	if last != "" {
		return last
	}
	return stdout
}

var pipelineResultsCmd = &cobra.Command{
	Use:   "results [name]",
	Short: "Show the output of the most recent pipeline run, optionally filtered by name",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open()
		if err != nil {
			return fmt.Errorf("results: open store: %w", err)
		}
		defer s.Close()

		runs, err := s.QueryRuns(200)
		if err != nil {
			return fmt.Errorf("results: query store: %w", err)
		}

		// Filter by name if given.
		nameFilter := ""
		if len(args) > 0 {
			nameFilter = args[0]
		}

		var matched []store.Run
		for _, r := range runs {
			if nameFilter == "" || r.Name == nameFilter {
				matched = append(matched, r)
				if len(matched) >= pipelineResultsLimit {
					break
				}
			}
		}

		if len(matched) == 0 {
			if nameFilter != "" {
				return fmt.Errorf("no runs found for pipeline %q", nameFilter)
			}
			return fmt.Errorf("no pipeline runs recorded yet")
		}

		for i, r := range matched {
			if i > 0 {
				fmt.Println("---")
			}
			status := "in-flight"
			if r.ExitStatus != nil {
				if *r.ExitStatus == 0 {
					status = "ok"
				} else {
					status = fmt.Sprintf("exit %d", *r.ExitStatus)
				}
			}
			startedAt := time.UnixMilli(r.StartedAt).Format(time.DateTime)
			fmt.Printf("pipeline: %s  |  run: %d  |  started: %s  |  status: %s\n", r.Name, r.ID, startedAt, status)
			if len(r.Steps) > 0 {
				for _, step := range r.Steps {
					dur := ""
					if step.DurationMs > 0 {
						dur = fmt.Sprintf("  %dms", step.DurationMs)
					}
					fmt.Printf("  %-20s %s%s\n", step.ID, step.Status, dur)
				}
			}
			fmt.Println()
			if r.Stdout != "" {
				fmt.Println(extractStreamResult(r.Stdout))
			}
			if r.Stderr != "" {
				fmt.Fprintf(os.Stderr, "%s\n", r.Stderr)
			}
		}
		return nil
	},
}

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
		mgr := executor.NewManager()
		for _, prov := range runProviders {
			if prov.SidecarPath != "" {
				continue
			}
			binary := prov.Command
			if binary == "" {
				binary = prov.ID
			}
			if err := mgr.Register(executor.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...)); err != nil {
				fmt.Fprintf(os.Stderr, "resume: register provider %q: %v\n", prov.ID, err)
			}
		}
		wrappersConfigDir, _ := glitchConfigDir()
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

package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/plugin"
)

// Workflow is a named sequence of steps loaded from YAML.
type Workflow struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Tags        []string       `yaml:"tags,omitempty"`
	Author      string         `yaml:"author,omitempty"`
	Version     string         `yaml:"version,omitempty"`
	Created     string         `yaml:"created,omitempty"`
	Actions     []string       `yaml:"actions,omitempty"`
	Steps       []Step         `yaml:"steps"`
	Items       []WorkflowItem `yaml:"-"`
	SourceFile  string         `yaml:"-"`
	Source      []byte         `yaml:"-"` // raw .glitch source for evaluator
	Args        []plugin.ArgDef `yaml:"-"`
	Input       *InputDef       `yaml:"-"`
}

// WorkflowItem is a union type for the ordered sequence of workflow elements.
// Exactly one of Step or Phase is non-nil.
type WorkflowItem struct {
	Step  *Step
	Phase *Phase
}

// Phase groups steps and gates into a retriable unit of work.
type Phase struct {
	ID      string
	Retries int
	Steps   []Step // work steps
	Gates   []Step // verification steps (IsGate = true)
}

// Step is a single unit of work — either a shell command, an LLM call, or a save-to-file.
// Control-flow forms (retry, timeout, catch, cond, map) wrap or replace regular steps.
type Step struct {
	ID       string   `yaml:"id"`
	Run      string   `yaml:"run,omitempty"`       // shell command
	LLM      *LLMStep `yaml:"llm,omitempty"`       // LLM call
	Save     string   `yaml:"save,omitempty"`      // write to file path (template-rendered)
	SaveStep string   `yaml:"save_step,omitempty"` // which step's output to save (default: previous)

	// Control-flow modifiers (applied to any step type)
	Retry   int    `yaml:"retry,omitempty"`   // max retries on failure (0 = no retry)
	Timeout string `yaml:"timeout,omitempty"` // deadline duration, e.g. "30s", "2m"

	// Compound forms — Form discriminates the variant.
	Form     string `yaml:"-"` // "cond", "map", "map-resources", "catch", or "" for regular steps
	Fallback *Step  `yaml:"-"` // catch: step to run on failure
	MapOver  string `yaml:"-"` // map: step ID whose output to iterate (newline-split)
	MapBody  *Step        `yaml:"-"` // map: template step executed per item (single step or first of chain)
	MapSteps []Step       `yaml:"-"` // map: extra body steps executed after MapBody per iteration

	// Conditional execution
	WhenPred string `yaml:"-"` // when: step ID or shell command predicate
	WhenBody *Step  `yaml:"-"` // when: step to execute if predicate is true
	WhenNot  bool   `yaml:"-"` // when-not: invert the predicate

	// Filter collection form
	FilterOver string `yaml:"-"` // filter: step ID whose output to iterate
	FilterBody *Step  `yaml:"-"` // filter: predicate step (truthy output = keep)

	// Reduce collection form
	ReduceOver string `yaml:"-"` // reduce: step ID whose output to iterate
	ReduceBody *Step  `yaml:"-"` // reduce: body step (receives item + accumulator)

	// map-resources: iterate over RunOpts.Resources, binding each to .resource.item.
	MapResourcesType string `yaml:"-"` // optional type filter ("git" / "local" / "tracker" or empty)
	MapResourcesBody *Step  `yaml:"-"` // body step executed per resource

	// Parallel execution
	ParSteps []Step `yaml:"-"` // par: concurrent child steps

	// Compare execution
	CompareBranches []CompareBranch `yaml:"-"` // compare: named alternative branches
	CompareReview   *ReviewConfig   `yaml:"-"` // compare: review config (nil = default judge)
	CompareID        string          `yaml:"-"` // compare: id for top-level compare blocks
	CompareObjective string          `yaml:"-"` // compare: required objective statement

	// Display hint — shown in CLI output next to the step name
	Hint string `yaml:"-"`

	// Phase/gate marker
	IsGate bool `yaml:"-"`

	// SDK forms
	Lines   string   `yaml:"-"` // step ID to split by newlines
	Flatten string   `yaml:"-"` // step ID whose JSON array output to flatten to NDJSON
	Merge   []string `yaml:"-"` // step IDs to merge
	ReadFile string  `yaml:"-"` // file path to read

	// call-workflow: invokes a sibling workflow by name as a nested run.
	CallWorkflow string            `yaml:"-"` // workflow name
	CallInput    string            `yaml:"-"` // template-rendered input for child
	CallSet      map[string]string `yaml:"-"` // :set key=value params

	// Source location (populated from sexpr nodes; zero for YAML-loaded steps).
	Line int `yaml:"-"`
	Col  int `yaml:"-"`
}

// CompareBranch is one named alternative in a (compare ...) form.
type CompareBranch struct {
	Name  string // branch name
	Steps []Step // steps to execute in this branch
}

// ReviewConfig configures the judge for a (compare ...) form.
type ReviewConfig struct {
	Criteria []string // scoring criteria names (criteria mode)
	Prompt   string   // custom review prompt (prompt mode)
	Model    string   // model override for the judge
	Provider string   // provider override for the judge (e.g. "claude", "copilot")
}

// LLMStep configures an LLM invocation.
type LLMStep struct {
	Provider string `yaml:"provider,omitempty"` // "ollama", "claude", "copilot", "gemini" (default: config)
	Model    string `yaml:"model,omitempty"`
	Prompt   string `yaml:"prompt"`
	Skill    string `yaml:"skill,omitempty"` // skill name or path — prepended to prompt as system context
	Tier     *int   `yaml:"tier,omitempty"`
	Format   string `yaml:"format,omitempty"`
}

// LoadFile reads a single workflow file (YAML or sexpr).
func LoadFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBytes(data, path)
}

// LoadBytes parses a workflow from raw bytes. Only .glitch sexpr files are supported.
func LoadBytes(data []byte, filename string) (*Workflow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".glitch" {
		return nil, fmt.Errorf("unsupported workflow format %q — only .glitch files are supported", ext)
	}
	args, err := plugin.ParseArgs(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	input, err := ParseInput(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	w, err := parseSexprWorkflow(data)
	if err != nil {
		return nil, err
	}
	w.Args = args
	w.Input = input
	w.SourceFile = filename
	for _, warn := range MergeImplicitArgs(w) {
		fmt.Fprintf(os.Stderr, "warning: %s: %s\n", filename, warn)
	}
	return w, nil
}

// LoadDir reads all .glitch workflow files from a directory, keyed by workflow name.
// Later entries overwrite earlier ones (allows local overrides).
func LoadDir(dir string) (map[string]*Workflow, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	// Resolve symlinks so WalkDir traverses the real target
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		resolved = dir
	}
	workflows := make(map[string]*Workflow)
	filepath.WalkDir(resolved, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != resolved && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(d.Name()) != ".glitch" {
			return nil
		}
		w, loadErr := LoadFile(path)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, loadErr)
			return nil
		}
		workflows[w.Name] = w
		return nil
	})
	return workflows, nil
}

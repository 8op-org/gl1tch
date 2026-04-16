package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
	Form     string       `yaml:"-"` // "cond", "map", "map-resources", "catch", or "" for regular steps
	Fallback *Step        `yaml:"-"` // catch: step to run on failure
	Branches []CondBranch `yaml:"-"` // cond: ordered predicate→step pairs
	MapOver  string       `yaml:"-"` // map: step ID whose output to iterate (newline-split)
	MapBody  *Step        `yaml:"-"` // map: template step executed per item

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
	CompareID       string          `yaml:"-"` // compare: id for top-level compare blocks

	// Plugin invocation
	PluginCall *PluginCallStep `yaml:"-"`

	// Display hint — shown in CLI output next to the step name
	Hint string `yaml:"-"`

	// Phase/gate marker
	IsGate bool `yaml:"-"`

	// SDK forms
	JsonPick  *JsonPickStep  `yaml:"-"`
	Lines     string         `yaml:"-"` // step ID to split by newlines
	Flatten   string         `yaml:"-"` // step ID whose JSON array output to flatten to NDJSON
	Merge     []string       `yaml:"-"` // step IDs to merge
	HttpCall  *HttpCallStep  `yaml:"-"`
	ReadFile  string         `yaml:"-"` // file path to read
	WriteFile *WriteFileStep `yaml:"-"`
	GlobPat   *GlobStep      `yaml:"-"`

	// ES forms
	Search *SearchStep `yaml:"-"`
	Index  *IndexStep  `yaml:"-"`
	Delete *DeleteStep `yaml:"-"`

	// Embedding
	Embed *EmbedStep `yaml:"-"`

	// call-workflow: invokes a sibling workflow by name as a nested run.
	CallWorkflow string            `yaml:"-"` // workflow name
	CallInput    string            `yaml:"-"` // template-rendered input for child
	CallSet      map[string]string `yaml:"-"` // :set key=value params

	// Source location (populated from sexpr nodes; zero for YAML-loaded steps).
	Line int `yaml:"-"`
	Col  int `yaml:"-"`
}

// CondBranch is one arm of a (cond ...) form.
type CondBranch struct {
	Pred string // shell command predicate (exit 0 = true), or "else"
	Step Step   // step to execute if predicate succeeds
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

// PluginCallStep invokes a plugin subcommand as a sub-workflow.
type PluginCallStep struct {
	Plugin     string            // plugin name
	Subcommand string            // subcommand name
	Args       map[string]string // keyword args
}

// JsonPickStep runs a jq expression against a step's output.
type JsonPickStep struct {
	Expr string // jq expression
	From string // step ID
}

// HttpCallStep performs an HTTP request.
type HttpCallStep struct {
	Method  string            // "GET" or "POST"
	URL     string            // template-rendered
	Body    string            // template-rendered (POST only)
	Headers map[string]string // template-rendered
}

// WriteFileStep writes a step's output to a file.
type WriteFileStep struct {
	Path string // template-rendered file path
	From string // step ID whose output to write
}

// GlobStep matches files against a pattern.
type GlobStep struct {
	Pattern string // glob pattern
	Dir     string // optional base directory
}

// SearchStep queries Elasticsearch and returns hits as JSON.
type SearchStep struct {
	IndexName string
	Query     string   // raw JSON query body
	Size      int      // max hits (default 10)
	Fields    []string // _source field filter
	ESURL     string   // override ES URL (empty = workspace default)
	Sort      string   // raw JSON sort clause
	NDJSON    bool     // output as NDJSON instead of JSON array
}

// IndexStep indexes a single document into Elasticsearch.
type IndexStep struct {
	IndexName  string
	Doc        string // template-rendered JSON document
	DocID      string // optional explicit _id
	ESURL      string
	EmbedField string // field in doc to embed (empty = no embedding)
	EmbedModel string
	Upsert     *bool // nil = default (upsert), false = skip if exists (op_type=create)
}

// DeleteStep deletes documents matching a query from Elasticsearch.
type DeleteStep struct {
	IndexName string
	Query     string // raw JSON query body
	ESURL     string
}

// EmbedStep generates an embedding vector from text.
type EmbedStep struct {
	Input string // template-rendered text to embed
	Model string
}

// LoadFile reads a single workflow file (YAML or sexpr).
func LoadFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBytes(data, path)
}

// LoadBytes parses a workflow from raw bytes, dispatching on file extension.
func LoadBytes(data []byte, filename string) (*Workflow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".glitch":
		w, err := parseSexprWorkflow(data)
		if err != nil {
			return nil, err
		}
		w.SourceFile = filename
		return w, nil
	default:
		var w Workflow
		if err := yaml.Unmarshal(data, &w); err != nil {
			return nil, fmt.Errorf("parse %s: %w", filename, err)
		}
		if w.Name == "" {
			w.Name = filename
		}
		w.SourceFile = filename
		return &w, nil
	}
}

// LoadDir reads all .yaml files from a directory, keyed by workflow name.
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
		ext := filepath.Ext(d.Name())
		if ext != ".yaml" && ext != ".yml" && ext != ".glitch" {
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

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
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
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
	Form     string       `yaml:"-"` // "cond", "map", "catch", or "" for regular steps
	Fallback *Step        `yaml:"-"` // catch: step to run on failure
	Branches []CondBranch `yaml:"-"` // cond: ordered predicate→step pairs
	MapOver  string       `yaml:"-"` // map: step ID whose output to iterate (newline-split)
	MapBody  *Step        `yaml:"-"` // map: template step executed per item
}

// CondBranch is one arm of a (cond ...) form.
type CondBranch struct {
	Pred string // shell command predicate (exit 0 = true), or "else"
	Step Step   // step to execute if predicate succeeds
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
	return LoadBytes(data, filepath.Base(path))
}

// LoadBytes parses a workflow from raw bytes, dispatching on file extension.
func LoadBytes(data []byte, filename string) (*Workflow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".glitch":
		return parseSexprWorkflow(data)
	default:
		var w Workflow
		if err := yaml.Unmarshal(data, &w); err != nil {
			return nil, fmt.Errorf("parse %s: %w", filename, err)
		}
		if w.Name == "" {
			w.Name = filename
		}
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
		if err != nil || d.IsDir() {
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

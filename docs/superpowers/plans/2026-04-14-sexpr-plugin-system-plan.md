# Sexpr Plugin System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace compiled Go binary plugins with user-authored sexpr workflow directories, adding SDK forms for JSON, HTTP, and file I/O.

**Architecture:** Plugins are directories of `.glitch` files discovered from `.glitch/plugins/` (project-local) and `~/.config/glitch/plugins/` (user-global). Each file is a subcommand with `(arg ...)` declarations. A new `(plugin ...)` step form invokes them from workflows. New SDK forms (`json-pick`, `lines`, `merge`, `http-get`, `http-post`, `read-file`, `write-file`, `glob`) are added as step-level forms in the pipeline evaluator.

**Tech Stack:** Go, Cobra, existing sexpr parser (no changes needed), `net/http`, `os`, `os/exec` (for jq)

**Spec:** `docs/superpowers/specs/2026-04-14-sexpr-plugin-system-design.md`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/plugin/discover.go` | Plugin directory discovery and subcommand resolution |
| `internal/plugin/discover_test.go` | Tests for discovery |
| `internal/plugin/manifest.go` | Parse `plugin.glitch` manifest — extract `(plugin ...)` metadata and shared `(def ...)` bindings |
| `internal/plugin/manifest_test.go` | Tests for manifest parsing |
| `internal/plugin/args.go` | Parse `(arg ...)` forms from subcommand files, build params map from CLI flags or sexpr keywords |
| `internal/plugin/args_test.go` | Tests for arg parsing |
| `internal/pipeline/sexpr.go` | Add `convertPlugin`, `convertJsonPick`, `convertLines`, `convertMerge`, `convertHttpGet`, `convertHttpPost`, `convertReadFile`, `convertWriteFile`, `convertGlob` converter functions; add cases to `convertForm` and `convertStep` |
| `internal/pipeline/types.go` | Add `PluginCall`, `JsonPick`, `HttpCall`, `FileOp` fields to `Step` struct |
| `internal/pipeline/runner.go` | Add execution logic for all new step forms in `runSingleStep` |
| `internal/pipeline/sexpr_test.go` | Tests for parsing all new forms |
| `internal/pipeline/runner_test.go` | Tests for executing new forms |
| `cmd/plugin.go` | `glitch plugin list`, `glitch plugin <name> --help`, `glitch plugin <name> <subcommand> [--flags]` |
| `cmd/plugin_test.go` | Tests for CLI plugin commands |

---

### Task 1: Plugin Discovery

Discover plugin directories from the two-layer search path and enumerate subcommands.

**Files:**
- Create: `internal/plugin/discover.go`
- Create: `internal/plugin/discover_test.go`

- [ ] **Step 1: Write failing test for DiscoverPlugins**

```go
// internal/plugin/discover_test.go
package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPlugins_GlobalOnly(t *testing.T) {
	dir := t.TempDir()
	ghDir := filepath.Join(dir, "github")
	os.MkdirAll(ghDir, 0o755)
	os.WriteFile(filepath.Join(ghDir, "prs.glitch"), []byte(`(workflow "prs" (step "s" (run "echo hi")))`), 0o644)
	os.WriteFile(filepath.Join(ghDir, "reviews.glitch"), []byte(`(workflow "reviews" (step "s" (run "echo hi")))`), 0o644)

	plugins := DiscoverPlugins("", dir)
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	p := plugins["github"]
	if p.Name != "github" {
		t.Fatalf("expected name %q, got %q", "github", p.Name)
	}
	if p.Source != "global" {
		t.Fatalf("expected source %q, got %q", "global", p.Source)
	}
	if len(p.Subcommands) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(p.Subcommands))
	}
}

func TestDiscoverPlugins_LocalOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	localDir := t.TempDir()

	// Same plugin in both
	for _, d := range []string{globalDir, localDir} {
		ghDir := filepath.Join(d, "github")
		os.MkdirAll(ghDir, 0o755)
		os.WriteFile(filepath.Join(ghDir, "prs.glitch"), []byte(`(workflow "prs" (step "s" (run "echo")))`), 0o644)
	}

	plugins := DiscoverPlugins(localDir, globalDir)
	p := plugins["github"]
	if p.Source != "local" {
		t.Fatalf("expected local to win, got %q", p.Source)
	}
}

func TestDiscoverPlugins_IgnoresPluginGlitch(t *testing.T) {
	dir := t.TempDir()
	ghDir := filepath.Join(dir, "github")
	os.MkdirAll(ghDir, 0o755)
	os.WriteFile(filepath.Join(ghDir, "plugin.glitch"), []byte(`(plugin "github")`), 0o644)
	os.WriteFile(filepath.Join(ghDir, "prs.glitch"), []byte(`(workflow "prs" (step "s" (run "echo")))`), 0o644)

	plugins := DiscoverPlugins("", dir)
	p := plugins["github"]
	if len(p.Subcommands) != 1 {
		t.Fatalf("expected 1 subcommand (plugin.glitch excluded), got %d", len(p.Subcommands))
	}
	if p.Subcommands[0] != "prs" {
		t.Fatalf("expected subcommand %q, got %q", "prs", p.Subcommands[0])
	}
}

func TestDiscoverPlugins_Empty(t *testing.T) {
	plugins := DiscoverPlugins("", t.TempDir())
	if len(plugins) != 0 {
		t.Fatalf("expected 0 plugins, got %d", len(plugins))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/plugin/ -v -run TestDiscover`
Expected: FAIL — package does not exist yet

- [ ] **Step 3: Implement DiscoverPlugins**

```go
// internal/plugin/discover.go
package plugin

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PluginInfo holds discovered plugin metadata.
type PluginInfo struct {
	Name        string
	Source      string // "local" or "global"
	Dir         string // absolute path to plugin directory
	Subcommands []string
}

// DiscoverPlugins finds plugins from local and global directories.
// Local plugins override global plugins with the same name.
// Pass empty string for localDir or globalDir to skip that source.
func DiscoverPlugins(localDir, globalDir string) map[string]*PluginInfo {
	plugins := make(map[string]*PluginInfo)

	// Global first, then local overwrites
	if globalDir != "" {
		discoverFrom(globalDir, "global", plugins)
	}
	if localDir != "" {
		discoverFrom(localDir, "local", plugins)
	}
	return plugins
}

func discoverFrom(dir, source string, plugins map[string]*PluginInfo) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		pluginDir := filepath.Join(dir, name)
		subs := listSubcommands(pluginDir)
		if len(subs) == 0 {
			continue
		}
		plugins[name] = &PluginInfo{
			Name:        name,
			Source:      source,
			Dir:         pluginDir,
			Subcommands: subs,
		}
	}
}

func listSubcommands(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var subs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".glitch") {
			continue
		}
		if name == "plugin.glitch" {
			continue
		}
		subs = append(subs, strings.TrimSuffix(name, ".glitch"))
	}
	sort.Strings(subs)
	return subs
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/plugin/ -v -run TestDiscover`
Expected: PASS (all 4 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/discover.go internal/plugin/discover_test.go
git commit -m "feat: plugin discovery from local and global directories"
```

---

### Task 2: Manifest Parsing

Parse `plugin.glitch` to extract `(plugin ...)` metadata and shared `(def ...)` bindings.

**Files:**
- Create: `internal/plugin/manifest.go`
- Create: `internal/plugin/manifest_test.go`

- [ ] **Step 1: Write failing tests for LoadManifest**

```go
// internal/plugin/manifest_test.go
package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest_Full(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin.glitch"), []byte(`
(plugin "github"
  :description "GitHub activity queries"
  :version "0.1.0")

(def username "adam-stokes")
(def timezone "US/Eastern")
`), 0o644)

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "github" {
		t.Fatalf("expected name %q, got %q", "github", m.Name)
	}
	if m.Description != "GitHub activity queries" {
		t.Fatalf("expected description %q, got %q", "GitHub activity queries", m.Description)
	}
	if m.Version != "0.1.0" {
		t.Fatalf("expected version %q, got %q", "0.1.0", m.Version)
	}
	if m.Defs["username"] != "adam-stokes" {
		t.Fatalf("expected def username, got %q", m.Defs["username"])
	}
	if m.Defs["timezone"] != "US/Eastern" {
		t.Fatalf("expected def timezone, got %q", m.Defs["timezone"])
	}
}

func TestLoadManifest_NoManifest(t *testing.T) {
	dir := t.TempDir()
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != filepath.Base(dir) {
		t.Fatalf("expected dir name as plugin name, got %q", m.Name)
	}
	if len(m.Defs) != 0 {
		t.Fatalf("expected no defs, got %d", len(m.Defs))
	}
}

func TestLoadManifest_DefsOnly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin.glitch"), []byte(`
(def x "hello")
(def y "world")
`), 0o644)

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	// No (plugin ...) form — name falls back to directory name
	if m.Name != filepath.Base(dir) {
		t.Fatalf("expected dir name, got %q", m.Name)
	}
	if m.Defs["x"] != "hello" {
		t.Fatalf("expected x=hello, got %q", m.Defs["x"])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/plugin/ -v -run TestLoadManifest`
Expected: FAIL — `LoadManifest` undefined

- [ ] **Step 3: Implement LoadManifest**

```go
// internal/plugin/manifest.go
package plugin

import (
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Manifest holds parsed plugin.glitch metadata.
type Manifest struct {
	Name        string
	Description string
	Version     string
	Defs        map[string]string
}

// LoadManifest parses plugin.glitch from dir.
// Returns a default manifest (name = dir basename) if no manifest file exists.
func LoadManifest(dir string) (*Manifest, error) {
	m := &Manifest{
		Name: filepath.Base(dir),
		Defs: make(map[string]string),
	}

	data, err := os.ReadFile(filepath.Join(dir, "plugin.glitch"))
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	nodes, err := sexpr.Parse(data)
	if err != nil {
		return nil, err
	}

	// Collect defs first (they may reference earlier defs)
	for _, n := range nodes {
		if n.IsList() && len(n.Children) >= 3 && n.Children[0].SymbolVal() == "def" {
			name := n.Children[1].SymbolVal()
			if name == "" {
				name = n.Children[1].StringVal()
			}
			val := n.Children[2].StringVal()
			// Resolve def references
			if n.Children[2].Atom != nil && n.Children[2].Atom.Type == sexpr.TokenSymbol {
				if v, ok := m.Defs[val]; ok {
					val = v
				}
			}
			m.Defs[name] = val
		}
	}

	// Find (plugin ...) form
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) < 2 {
			continue
		}
		if n.Children[0].SymbolVal() != "plugin" {
			continue
		}
		m.Name = n.Children[1].StringVal()
		// Parse keywords
		children := n.Children[2:]
		for i := 0; i < len(children); i++ {
			child := children[i]
			if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
				key := child.KeywordVal()
				i++
				if i >= len(children) {
					break
				}
				val := children[i].StringVal()
				switch key {
				case "description":
					m.Description = val
				case "version":
					m.Version = val
				}
			}
		}
		break
	}

	return m, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/plugin/ -v -run TestLoadManifest`
Expected: PASS (all 3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/manifest.go internal/plugin/manifest_test.go
git commit -m "feat: parse plugin.glitch manifest for metadata and shared defs"
```

---

### Task 3: Argument Parsing

Parse `(arg ...)` forms from subcommand files and map CLI flags or sexpr keywords to params.

**Files:**
- Create: `internal/plugin/args.go`
- Create: `internal/plugin/args_test.go`

- [ ] **Step 1: Write failing tests for ParseArgs and BuildParams**

```go
// internal/plugin/args_test.go
package plugin

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	src := []byte(`
(arg "since" :default "yesterday" :description "Time range")
(arg "authored" :type :flag :description "PRs you authored")
(arg "count" :type :number :description "Max results")

(workflow "prs"
  (step "s" (run "echo")))
`)
	args, err := ParseArgs(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}

	since := args[0]
	if since.Name != "since" {
		t.Fatalf("expected name %q, got %q", "since", since.Name)
	}
	if since.Default != "yesterday" {
		t.Fatalf("expected default %q, got %q", "yesterday", since.Default)
	}
	if since.Type != "string" {
		t.Fatalf("expected type %q, got %q", "string", since.Type)
	}
	if since.Description != "Time range" {
		t.Fatalf("expected description %q, got %q", "Time range", since.Description)
	}

	authored := args[1]
	if authored.Type != "flag" {
		t.Fatalf("expected type %q, got %q", "flag", authored.Type)
	}

	count := args[2]
	if count.Type != "number" {
		t.Fatalf("expected type %q, got %q", "number", count.Type)
	}
}

func TestParseArgs_NoArgs(t *testing.T) {
	src := []byte(`(workflow "test" (step "s" (run "echo")))`)
	args, err := ParseArgs(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 0 {
		t.Fatalf("expected 0 args, got %d", len(args))
	}
}

func TestBuildParams_FromFlags(t *testing.T) {
	argDefs := []ArgDef{
		{Name: "since", Default: "yesterday", Type: "string"},
		{Name: "authored", Type: "flag"},
	}
	flags := map[string]string{"since": "week", "authored": "true"}

	params, err := BuildParams(argDefs, flags)
	if err != nil {
		t.Fatal(err)
	}
	if params["since"] != "week" {
		t.Fatalf("expected since=week, got %q", params["since"])
	}
	if params["authored"] != "true" {
		t.Fatalf("expected authored=true, got %q", params["authored"])
	}
}

func TestBuildParams_Defaults(t *testing.T) {
	argDefs := []ArgDef{
		{Name: "since", Default: "yesterday", Type: "string"},
		{Name: "authored", Type: "flag"},
	}
	flags := map[string]string{}

	params, err := BuildParams(argDefs, flags)
	if err != nil {
		t.Fatal(err)
	}
	if params["since"] != "yesterday" {
		t.Fatalf("expected default since=yesterday, got %q", params["since"])
	}
	if params["authored"] != "" {
		t.Fatalf("expected empty authored, got %q", params["authored"])
	}
}

func TestBuildParams_RequiredMissing(t *testing.T) {
	argDefs := []ArgDef{
		{Name: "repo", Type: "string"}, // no default = required
	}
	flags := map[string]string{}

	_, err := BuildParams(argDefs, flags)
	if err == nil {
		t.Fatal("expected error for missing required arg")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/plugin/ -v -run "TestParseArgs|TestBuildParams"`
Expected: FAIL — `ParseArgs`, `ArgDef`, `BuildParams` undefined

- [ ] **Step 3: Implement ParseArgs and BuildParams**

```go
// internal/plugin/args.go
package plugin

import (
	"fmt"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// ArgDef describes a declared plugin argument.
type ArgDef struct {
	Name        string
	Default     string // empty string means required (unless type is flag)
	Type        string // "string", "flag", "number"
	Description string
	Required    bool
}

// ParseArgs extracts (arg ...) forms from a subcommand source file.
func ParseArgs(src []byte) ([]ArgDef, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}

	var args []ArgDef
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) < 2 {
			continue
		}
		if n.Children[0].SymbolVal() != "arg" {
			continue
		}

		ad := ArgDef{
			Name: n.Children[1].StringVal(),
			Type: "string",
		}

		children := n.Children[2:]
		for i := 0; i < len(children); i++ {
			child := children[i]
			if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
				key := child.KeywordVal()
				i++
				if i >= len(children) {
					break
				}
				val := children[i]
				switch key {
				case "default":
					ad.Default = val.StringVal()
				case "type":
					// :type :flag → keyword val is ":flag", strip the colon
					if val.Atom != nil && val.Atom.Type == sexpr.TokenKeyword {
						ad.Type = val.KeywordVal()
					} else {
						ad.Type = val.StringVal()
					}
				case "description":
					ad.Description = val.StringVal()
				}
			}
		}

		// Required if no default and not a flag
		ad.Required = ad.Default == "" && ad.Type != "flag"
		args = append(args, ad)
	}
	return args, nil
}

// BuildParams maps provided flag values against arg definitions, applying defaults.
// Returns error if a required arg is missing.
func BuildParams(argDefs []ArgDef, flags map[string]string) (map[string]string, error) {
	params := make(map[string]string)
	for _, ad := range argDefs {
		if val, ok := flags[ad.Name]; ok {
			params[ad.Name] = val
		} else if ad.Default != "" {
			params[ad.Name] = ad.Default
		} else if ad.Required {
			return nil, fmt.Errorf("required argument %q not provided", ad.Name)
		}
	}
	return params, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/plugin/ -v -run "TestParseArgs|TestBuildParams"`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/args.go internal/plugin/args_test.go
git commit -m "feat: parse (arg ...) forms and build params with defaults/validation"
```

---

### Task 4: Step Types for New Forms

Add fields to the `Step` struct and new types for plugin calls, JSON operations, HTTP calls, and file operations.

**Files:**
- Modify: `internal/pipeline/types.go:20-38`

- [ ] **Step 1: Add new fields to Step struct and new types**

Add these fields to the `Step` struct in `types.go` after the existing `SaveStep` field (line 27):

```go
// Plugin invocation
PluginCall *PluginCallStep `yaml:"-"`

// SDK forms
JsonPick  *JsonPickStep  `yaml:"-"`
Lines     string         `yaml:"-"` // step ID to split by newlines
Merge     []string       `yaml:"-"` // step IDs to merge
HttpCall  *HttpCallStep  `yaml:"-"`
ReadFile  string         `yaml:"-"` // file path to read
WriteFile *WriteFileStep `yaml:"-"`
GlobPat   *GlobStep      `yaml:"-"`
```

Add these new types after `LLMStep`:

```go
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
```

- [ ] **Step 2: Verify existing tests still pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: PASS — new fields are zero-valued, no behavior change

- [ ] **Step 3: Commit**

```bash
git add internal/pipeline/types.go
git commit -m "feat: add Step fields and types for plugin calls and SDK forms"
```

---

### Task 5: Sexpr Converters for SDK Forms

Add converter functions for all new forms and wire them into the dispatch.

**Files:**
- Modify: `internal/pipeline/sexpr.go:103-146` (convertForm switch) and `internal/pipeline/sexpr.go:329-365` (convertStep switch)
- Modify: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write failing tests for all new form parsers**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_JsonPick(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "fetch" (run "echo '{\"a\":1}'"))
  (step "pick" (json-pick ".a" :from "fetch")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.JsonPick == nil {
		t.Fatal("expected json-pick step")
	}
	if s.JsonPick.Expr != ".a" {
		t.Fatalf("expected expr %q, got %q", ".a", s.JsonPick.Expr)
	}
	if s.JsonPick.From != "fetch" {
		t.Fatalf("expected from %q, got %q", "fetch", s.JsonPick.From)
	}
}

func TestSexprWorkflow_Lines(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "list" (run "echo -e 'a\nb\nc'"))
  (step "split" (lines "list")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.Lines != "list" {
		t.Fatalf("expected lines ref %q, got %q", "list", s.Lines)
	}
}

func TestSexprWorkflow_Merge(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "a" (run "echo '{\"x\":1}'"))
  (step "b" (run "echo '{\"y\":2}'"))
  (step "combined" (merge "a" "b")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[2]
	if len(s.Merge) != 2 || s.Merge[0] != "a" || s.Merge[1] != "b" {
		t.Fatalf("expected merge [a b], got %v", s.Merge)
	}
}

func TestSexprWorkflow_HttpGet(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "fetch"
    (http-get "https://api.example.com/data"
      :headers {"Authorization" "Bearer token123"})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected http-call step")
	}
	if s.HttpCall.Method != "GET" {
		t.Fatalf("expected method GET, got %q", s.HttpCall.Method)
	}
	if s.HttpCall.URL != "https://api.example.com/data" {
		t.Fatalf("expected url, got %q", s.HttpCall.URL)
	}
	if s.HttpCall.Headers["Authorization"] != "Bearer token123" {
		t.Fatalf("expected auth header, got %v", s.HttpCall.Headers)
	}
}

func TestSexprWorkflow_HttpPost(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "submit"
    (http-post "https://api.example.com/submit"
      :body "{\"key\":\"value\"}"
      :headers {"Content-Type" "application/json"})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected http-call step")
	}
	if s.HttpCall.Method != "POST" {
		t.Fatalf("expected method POST, got %q", s.HttpCall.Method)
	}
	if s.HttpCall.Body != `{"key":"value"}` {
		t.Fatalf("expected body, got %q", s.HttpCall.Body)
	}
}

func TestSexprWorkflow_ReadFile(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "config" (read-file "config.json")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].ReadFile != "config.json" {
		t.Fatalf("expected read-file path, got %q", w.Steps[0].ReadFile)
	}
}

func TestSexprWorkflow_WriteFile(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "gen" (run "echo data"))
  (step "save" (write-file "output.json" :from "gen")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.WriteFile == nil {
		t.Fatal("expected write-file step")
	}
	if s.WriteFile.Path != "output.json" {
		t.Fatalf("expected path %q, got %q", "output.json", s.WriteFile.Path)
	}
	if s.WriteFile.From != "gen" {
		t.Fatalf("expected from %q, got %q", "gen", s.WriteFile.From)
	}
}

func TestSexprWorkflow_Glob(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "files" (glob "*.yaml" :dir "configs/")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.GlobPat == nil {
		t.Fatal("expected glob step")
	}
	if s.GlobPat.Pattern != "*.yaml" {
		t.Fatalf("expected pattern %q, got %q", "*.yaml", s.GlobPat.Pattern)
	}
	if s.GlobPat.Dir != "configs/" {
		t.Fatalf("expected dir %q, got %q", "configs/", s.GlobPat.Dir)
	}
}

func TestSexprWorkflow_PluginCall(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "prs"
    (plugin "github" "prs" :since "yesterday" :authored)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.PluginCall == nil {
		t.Fatal("expected plugin call")
	}
	if s.PluginCall.Plugin != "github" {
		t.Fatalf("expected plugin %q, got %q", "github", s.PluginCall.Plugin)
	}
	if s.PluginCall.Subcommand != "prs" {
		t.Fatalf("expected subcommand %q, got %q", "prs", s.PluginCall.Subcommand)
	}
	if s.PluginCall.Args["since"] != "yesterday" {
		t.Fatalf("expected since=yesterday, got %q", s.PluginCall.Args["since"])
	}
	if s.PluginCall.Args["authored"] != "true" {
		t.Fatalf("expected authored=true, got %q", s.PluginCall.Args["authored"])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run "TestSexprWorkflow_Json|TestSexprWorkflow_Lines|TestSexprWorkflow_Merge|TestSexprWorkflow_Http|TestSexprWorkflow_ReadFile|TestSexprWorkflow_WriteFile|TestSexprWorkflow_Glob|TestSexprWorkflow_PluginCall"`
Expected: FAIL — unknown step types

- [ ] **Step 3: Implement converter functions**

Add these functions to `internal/pipeline/sexpr.go`:

```go
func convertJsonPick(n *sexpr.Node, defs map[string]string) (*JsonPickStep, error) {
	children := n.Children[1:] // skip "json-pick"
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (json-pick) missing expression", n.Line)
	}
	jp := &JsonPickStep{Expr: resolveVal(children[0], defs)}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword && child.KeywordVal() == "from" {
			i++
			if i < len(children) {
				jp.From = resolveVal(children[i], defs)
			}
		}
	}
	return jp, nil
}

func convertLines(n *sexpr.Node, defs map[string]string) (string, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return "", fmt.Errorf("line %d: (lines) missing step ID", n.Line)
	}
	return resolveVal(children[0], defs), nil
}

func convertMerge(n *sexpr.Node, defs map[string]string) ([]string, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return nil, fmt.Errorf("line %d: (merge) needs at least 2 step IDs", n.Line)
	}
	var ids []string
	for _, c := range children {
		ids = append(ids, resolveVal(c, defs))
	}
	return ids, nil
}

func convertHttpCall(n *sexpr.Node, method string, defs map[string]string) (*HttpCallStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (%s) missing URL", n.Line, n.Children[0].StringVal())
	}
	h := &HttpCallStep{
		Method:  method,
		URL:     resolveVal(children[0], defs),
		Headers: make(map[string]string),
	}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				break
			}
			val := children[i]
			switch key {
			case "body":
				h.Body = resolveVal(val, defs)
			case "headers":
				if val.IsMap || (val.IsList() && !val.IsMap) {
					// Parse map children as key-value pairs
					src := val.Children
					for j := 0; j+1 < len(src); j += 2 {
						h.Headers[src[j].StringVal()] = resolveVal(src[j+1], defs)
					}
				}
			}
		}
	}
	return h, nil
}

func convertReadFile(n *sexpr.Node, defs map[string]string) (string, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return "", fmt.Errorf("line %d: (read-file) missing path", n.Line)
	}
	return resolveVal(children[0], defs), nil
}

func convertWriteFile(n *sexpr.Node, defs map[string]string) (*WriteFileStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (write-file) missing path", n.Line)
	}
	wf := &WriteFileStep{Path: resolveVal(children[0], defs)}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword && child.KeywordVal() == "from" {
			i++
			if i < len(children) {
				wf.From = resolveVal(children[i], defs)
			}
		}
	}
	return wf, nil
}

func convertGlobStep(n *sexpr.Node, defs map[string]string) (*GlobStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (glob) missing pattern", n.Line)
	}
	g := &GlobStep{Pattern: resolveVal(children[0], defs)}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword && child.KeywordVal() == "dir" {
			i++
			if i < len(children) {
				g.Dir = resolveVal(children[i], defs)
			}
		}
	}
	return g, nil
}

func convertPluginCall(n *sexpr.Node, defs map[string]string) (*PluginCallStep, error) {
	children := n.Children[1:] // skip "plugin"
	if len(children) < 2 {
		return nil, fmt.Errorf("line %d: (plugin) needs plugin name and subcommand", n.Line)
	}
	pc := &PluginCallStep{
		Plugin:     resolveVal(children[0], defs),
		Subcommand: resolveVal(children[1], defs),
		Args:       make(map[string]string),
	}
	for i := 2; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			// Check if next token is another keyword or end — then it's a flag
			if i+1 >= len(children) || (children[i+1].IsAtom() && children[i+1].Atom.Type == sexpr.TokenKeyword) {
				pc.Args[key] = "true"
			} else {
				i++
				pc.Args[key] = resolveVal(children[i], defs)
			}
		}
	}
	return pc, nil
}
```

- [ ] **Step 4: Wire converters into convertStep**

In `convertStep` in `internal/pipeline/sexpr.go`, add these cases to the switch on `head` inside the `for _, child := range children` loop (after the existing `"save"` case):

```go
case "json-pick":
	jp, err := convertJsonPick(child, defs)
	if err != nil {
		return s, err
	}
	s.JsonPick = jp
case "lines":
	ref, err := convertLines(child, defs)
	if err != nil {
		return s, err
	}
	s.Lines = ref
case "merge":
	ids, err := convertMerge(child, defs)
	if err != nil {
		return s, err
	}
	s.Merge = ids
case "http-get":
	hc, err := convertHttpCall(child, "GET", defs)
	if err != nil {
		return s, err
	}
	s.HttpCall = hc
case "http-post":
	hc, err := convertHttpCall(child, "POST", defs)
	if err != nil {
		return s, err
	}
	s.HttpCall = hc
case "read-file":
	path, err := convertReadFile(child, defs)
	if err != nil {
		return s, err
	}
	s.ReadFile = path
case "write-file":
	wf, err := convertWriteFile(child, defs)
	if err != nil {
		return s, err
	}
	s.WriteFile = wf
case "glob":
	g, err := convertGlobStep(child, defs)
	if err != nil {
		return s, err
	}
	s.GlobPat = g
case "plugin":
	pc, err := convertPluginCall(child, defs)
	if err != nil {
		return s, err
	}
	s.PluginCall = pc
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run "TestSexprWorkflow_Json|TestSexprWorkflow_Lines|TestSexprWorkflow_Merge|TestSexprWorkflow_Http|TestSexprWorkflow_ReadFile|TestSexprWorkflow_WriteFile|TestSexprWorkflow_Glob|TestSexprWorkflow_PluginCall"`
Expected: PASS (all 9 tests)

- [ ] **Step 6: Run ALL existing tests to verify no regressions**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: PASS — all existing + new tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat: sexpr converters for plugin call and SDK forms (json-pick, lines, merge, http, file, glob)"
```

---

### Task 6: Runner Execution for SDK Forms

Add execution logic in `runSingleStep` for all new step types.

**Files:**
- Modify: `internal/pipeline/runner.go:518-693`
- Modify: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write failing tests for SDK form execution**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRunSingleStep_Lines(t *testing.T) {
	rctx := &runCtx{
		steps:  map[string]string{"list": "alpha\nbeta\ngamma"},
		params: map[string]string{},
	}
	step := Step{ID: "split", Lines: "list"}
	outcome, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	expected := `["alpha","beta","gamma"]`
	if outcome.output != expected {
		t.Fatalf("expected %q, got %q", expected, outcome.output)
	}
}

func TestRunSingleStep_Merge(t *testing.T) {
	rctx := &runCtx{
		steps:  map[string]string{"a": `{"x":1}`, "b": `{"y":2}`},
		params: map[string]string{},
	}
	step := Step{ID: "combined", Merge: []string{"a", "b"}}
	outcome, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	// Should contain both keys
	if !strings.Contains(outcome.output, `"x"`) || !strings.Contains(outcome.output, `"y"`) {
		t.Fatalf("expected merged JSON, got %q", outcome.output)
	}
}

func TestRunSingleStep_ReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("file contents"), 0o644)

	rctx := &runCtx{
		steps:  map[string]string{},
		params: map[string]string{},
	}
	step := Step{ID: "read", ReadFile: path}
	outcome, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.output != "file contents" {
		t.Fatalf("expected %q, got %q", "file contents", outcome.output)
	}
}

func TestRunSingleStep_WriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out", "result.txt")

	rctx := &runCtx{
		steps:  map[string]string{"gen": "hello world"},
		params: map[string]string{},
	}
	step := Step{ID: "save", WriteFile: &WriteFileStep{Path: path, From: "gen"}}
	outcome, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.output != path {
		t.Fatalf("expected path as output, got %q", outcome.output)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello world" {
		t.Fatalf("expected file contents %q, got %q", "hello world", string(data))
	}
}

func TestRunSingleStep_Glob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "b.yaml"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte(""), 0o644)

	rctx := &runCtx{
		steps:  map[string]string{},
		params: map[string]string{},
	}
	step := Step{ID: "files", GlobPat: &GlobStep{Pattern: "*.yaml", Dir: dir}}
	outcome, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(outcome.output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 matches, got %d: %q", len(lines), outcome.output)
	}
}

func TestRunSingleStep_JsonPick(t *testing.T) {
	// Skip if jq not available
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH")
	}

	rctx := &runCtx{
		steps:  map[string]string{"data": `{"name":"alice","age":30}`},
		params: map[string]string{},
	}
	step := Step{ID: "pick", JsonPick: &JsonPickStep{Expr: ".name", From: "data"}}
	outcome, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(outcome.output) != `"alice"` {
		t.Fatalf("expected %q, got %q", `"alice"`, outcome.output)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run "TestRunSingleStep_Lines|TestRunSingleStep_Merge|TestRunSingleStep_ReadFile|TestRunSingleStep_WriteFile|TestRunSingleStep_Glob|TestRunSingleStep_JsonPick"`
Expected: FAIL — unhandled step types

- [ ] **Step 3: Implement execution for all new forms**

Add these cases to `runSingleStep` in `internal/pipeline/runner.go`, before the final error return at line 692:

```go
if step.JsonPick != nil {
	from := rctx.steps[step.JsonPick.From]
	rendered, err := render(step.JsonPick.Expr, data, rctx.steps)
	if err != nil {
		return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
	}
	fmt.Printf("  > %s (json-pick)\n", step.ID)
	cmd := exec.CommandContext(ctx, "jq", rendered)
	cmd.Stdin = strings.NewReader(from)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("step %s: jq: %w", step.ID, err)
	}
	return &stepOutcome{output: strings.TrimSpace(string(out))}, nil
}

if step.Lines != "" {
	from := rctx.steps[step.Lines]
	lines := strings.Split(strings.TrimSpace(from), "\n")
	// Build JSON array
	var parts []string
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%q", l))
	}
	out := "[" + strings.Join(parts, ",") + "]"
	fmt.Printf("  > %s (lines)\n", step.ID)
	return &stepOutcome{output: out}, nil
}

if len(step.Merge) > 0 {
	merged := make(map[string]any)
	for _, id := range step.Merge {
		var obj map[string]any
		if err := json.Unmarshal([]byte(rctx.steps[id]), &obj); err != nil {
			return nil, fmt.Errorf("step %s: merge %q: %w", step.ID, id, err)
		}
		for k, v := range obj {
			merged[k] = v
		}
	}
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("step %s: merge marshal: %w", step.ID, err)
	}
	fmt.Printf("  > %s (merge)\n", step.ID)
	return &stepOutcome{output: string(out)}, nil
}

if step.HttpCall != nil {
	renderedURL, err := render(step.HttpCall.URL, data, rctx.steps)
	if err != nil {
		return nil, fmt.Errorf("step %s: template url: %w", step.ID, err)
	}
	var bodyReader *strings.Reader
	if step.HttpCall.Body != "" {
		renderedBody, err := render(step.HttpCall.Body, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template body: %w", step.ID, err)
		}
		bodyReader = strings.NewReader(renderedBody)
	}
	fmt.Printf("  > %s (%s %s)\n", step.ID, step.HttpCall.Method, renderedURL)
	var req *http.Request
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, step.HttpCall.Method, renderedURL, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, step.HttpCall.Method, renderedURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("step %s: http request: %w", step.ID, err)
	}
	for k, v := range step.HttpCall.Headers {
		rendered, err := render(v, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template header %q: %w", step.ID, k, err)
		}
		req.Header.Set(k, rendered)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("step %s: http: %w", step.ID, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("step %s: http %d: %s", step.ID, resp.StatusCode, string(body))
	}
	return &stepOutcome{output: string(body)}, nil
}

if step.ReadFile != "" {
	renderedPath, err := render(step.ReadFile, data, rctx.steps)
	if err != nil {
		return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
	}
	fmt.Printf("  > %s (read-file %s)\n", step.ID, renderedPath)
	content, err := os.ReadFile(renderedPath)
	if err != nil {
		return nil, fmt.Errorf("step %s: read-file: %w", step.ID, err)
	}
	return &stepOutcome{output: string(content)}, nil
}

if step.WriteFile != nil {
	renderedPath, err := render(step.WriteFile.Path, data, rctx.steps)
	if err != nil {
		return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
	}
	content := rctx.steps[step.WriteFile.From]
	fmt.Printf("  > %s (write-file %s)\n", step.ID, renderedPath)
	if err := os.MkdirAll(filepath.Dir(renderedPath), 0o755); err != nil {
		return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
	}
	if err := os.WriteFile(renderedPath, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("step %s: write-file: %w", step.ID, err)
	}
	return &stepOutcome{output: renderedPath}, nil
}

if step.GlobPat != nil {
	pattern := step.GlobPat.Pattern
	dir := step.GlobPat.Dir
	if dir != "" {
		renderedDir, err := render(dir, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template dir: %w", step.ID, err)
		}
		pattern = filepath.Join(renderedDir, pattern)
	}
	fmt.Printf("  > %s (glob %s)\n", step.ID, pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("step %s: glob: %w", step.ID, err)
	}
	return &stepOutcome{output: strings.Join(matches, "\n")}, nil
}

if step.PluginCall != nil {
	fmt.Printf("  > %s (plugin %s %s)\n", step.ID, step.PluginCall.Plugin, step.PluginCall.Subcommand)
	out, err := executePluginCall(ctx, rctx, step)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", step.ID, err)
	}
	return &stepOutcome{output: out}, nil
}
```

Add required imports to `runner.go`:

```go
"encoding/json"
"io"
"net/http"
```

- [ ] **Step 4: Add executePluginCall stub**

Add to `internal/pipeline/runner.go` (this will be fully implemented in Task 7):

```go
// executePluginCall resolves and runs a plugin subcommand as a sub-workflow.
func executePluginCall(ctx context.Context, rctx *runCtx, step Step) (string, error) {
	return "", fmt.Errorf("plugin execution not yet implemented")
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run "TestRunSingleStep_Lines|TestRunSingleStep_Merge|TestRunSingleStep_ReadFile|TestRunSingleStep_WriteFile|TestRunSingleStep_Glob|TestRunSingleStep_JsonPick"`
Expected: PASS (all 6 tests)

- [ ] **Step 6: Run ALL tests to verify no regressions**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat: execute SDK forms (json-pick, lines, merge, http, read-file, write-file, glob)"
```

---

### Task 7: Plugin Execution (executePluginCall)

Wire up the full plugin execution path: discover plugin, load manifest, parse args, merge params, run sub-workflow.

**Files:**
- Modify: `internal/pipeline/runner.go` (replace `executePluginCall` stub)
- Create: `internal/pipeline/plugin_runner.go`
- Create: `internal/pipeline/plugin_runner_test.go`

- [ ] **Step 1: Write failing test for plugin execution**

```go
// internal/pipeline/plugin_runner_test.go
package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecutePluginCall_Basic(t *testing.T) {
	// Set up a plugin directory
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "greeter")
	os.MkdirAll(pluginDir, 0o755)

	// Manifest with shared def
	os.WriteFile(filepath.Join(pluginDir, "plugin.glitch"), []byte(`
(plugin "greeter" :description "test plugin")
(def greeting "hello")
`), 0o644)

	// Subcommand
	os.WriteFile(filepath.Join(pluginDir, "say.glitch"), []byte(`
(arg "name" :default "world")

(workflow "say"
  (step "out" (run "echo {{.param.greeting}} {{.param.name}}")))
`), 0o644)

	result, err := RunPluginSubcommand(dir, "greeter", "say", map[string]string{"name": "alice"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "hello alice"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestExecutePluginCall_MissingPlugin(t *testing.T) {
	dir := t.TempDir()
	_, err := RunPluginSubcommand(dir, "nope", "sub", nil)
	if err == nil {
		t.Fatal("expected error for missing plugin")
	}
}

func TestExecutePluginCall_MissingSubcommand(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "test")
	os.MkdirAll(pluginDir, 0o755)
	os.WriteFile(filepath.Join(pluginDir, "only.glitch"), []byte(`(workflow "only" (step "s" (run "echo")))`), 0o644)

	_, err := RunPluginSubcommand(dir, "test", "nope", nil)
	if err == nil {
		t.Fatal("expected error for missing subcommand")
	}
}

func TestExecutePluginCall_RequiredArgMissing(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "test")
	os.MkdirAll(pluginDir, 0o755)
	os.WriteFile(filepath.Join(pluginDir, "cmd.glitch"), []byte(`
(arg "repo")
(workflow "cmd" (step "s" (run "echo {{.param.repo}}")))
`), 0o644)

	_, err := RunPluginSubcommand(dir, "test", "cmd", nil)
	if err == nil {
		t.Fatal("expected error for missing required arg")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestExecutePluginCall`
Expected: FAIL — `RunPluginSubcommand` undefined

- [ ] **Step 3: Implement RunPluginSubcommand**

```go
// internal/pipeline/plugin_runner.go
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/plugin"
	"github.com/8op-org/gl1tch/internal/provider"
)

// RunPluginSubcommand loads and executes a plugin subcommand from a plugin directory root.
// pluginRoot is the parent directory containing plugin directories (e.g., ~/.config/glitch/plugins).
func RunPluginSubcommand(pluginRoot, pluginName, subcommand string, args map[string]string) (string, error) {
	pluginDir := filepath.Join(pluginRoot, pluginName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return "", fmt.Errorf("plugin %q not found in %s", pluginName, pluginRoot)
	}

	// Load manifest
	manifest, err := plugin.LoadManifest(pluginDir)
	if err != nil {
		return "", fmt.Errorf("plugin %q manifest: %w", pluginName, err)
	}

	// Load subcommand file
	subPath := filepath.Join(pluginDir, subcommand+".glitch")
	data, err := os.ReadFile(subPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("plugin %q has no subcommand %q", pluginName, subcommand)
		}
		return "", err
	}

	// Parse args from subcommand
	argDefs, err := plugin.ParseArgs(data)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q args: %w", pluginName, subcommand, err)
	}

	// Build params: manifest defs + resolved args
	if args == nil {
		args = make(map[string]string)
	}
	params, err := plugin.BuildParams(argDefs, args)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q: %w", pluginName, subcommand, err)
	}

	// Inject manifest defs (subcommand params override)
	for k, v := range manifest.Defs {
		if _, exists := params[k]; !exists {
			params[k] = v
		}
	}

	// Parse and run the workflow
	w, err := parseSexprWorkflow(data)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q parse: %w", pluginName, subcommand, err)
	}

	reg, _ := provider.LoadProviders("")

	result, err := Run(w, "", "", params, reg)
	if err != nil {
		return "", fmt.Errorf("plugin %q %q run: %w", pluginName, subcommand, err)
	}
	return strings.TrimSpace(result.Output), nil
}
```

- [ ] **Step 4: Replace executePluginCall stub in runner.go**

Replace the stub in `internal/pipeline/runner.go`:

```go
// executePluginCall resolves and runs a plugin subcommand as a sub-workflow.
func executePluginCall(ctx context.Context, rctx *runCtx, step Step) (string, error) {
	pc := step.PluginCall

	// Search order: project-local, then user-global
	searchDirs := []string{".glitch/plugins"}
	if home, err := os.UserHomeDir(); err == nil {
		searchDirs = append(searchDirs, filepath.Join(home, ".config", "glitch", "plugins"))
	}

	for _, dir := range searchDirs {
		pluginDir := filepath.Join(dir, pc.Plugin)
		if _, err := os.Stat(pluginDir); err == nil {
			return RunPluginSubcommand(dir, pc.Plugin, pc.Subcommand, pc.Args)
		}
	}

	return "", fmt.Errorf("plugin %q not found, searched: %s", pc.Plugin, strings.Join(searchDirs, ", "))
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestExecutePluginCall`
Expected: PASS (all 4 tests)

- [ ] **Step 6: Run ALL tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/plugin_runner.go internal/pipeline/plugin_runner_test.go internal/pipeline/runner.go
git commit -m "feat: plugin execution — load manifest, parse args, run sub-workflow"
```

---

### Task 8: CLI Command (`glitch plugin`)

Add `glitch plugin list`, `glitch plugin <name> --help`, and `glitch plugin <name> <subcommand> [--flags]`.

**Files:**
- Create: `cmd/plugin.go`
- Create: `cmd/plugin_test.go`

- [ ] **Step 1: Write failing test for plugin CLI**

```go
// cmd/plugin_test.go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginListCmd(t *testing.T) {
	// Create a temp global plugin dir
	dir := t.TempDir()
	ghDir := filepath.Join(dir, "github")
	os.MkdirAll(ghDir, 0o755)
	os.WriteFile(filepath.Join(ghDir, "plugin.glitch"), []byte(`
(plugin "github" :description "GitHub queries" :version "1.0")
`), 0o644)
	os.WriteFile(filepath.Join(ghDir, "prs.glitch"), []byte(`
(workflow "prs" (step "s" (run "echo prs")))
`), 0o644)

	// Override global dir for test
	plugins := discoverAllPlugins("", dir)
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins["github"] == nil {
		t.Fatal("expected github plugin")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -v -run TestPluginListCmd`
Expected: FAIL — `discoverAllPlugins` undefined

- [ ] **Step 3: Implement cmd/plugin.go**

```go
// cmd/plugin.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/plugin"
)

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "manage and run sexpr plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return nil
	},
	// Allow unknown subcommands (dynamic plugin dispatch)
	DisableFlagParsing: true,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "list discovered plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		localDir := ".glitch/plugins"
		globalDir := ""
		if home, err := os.UserHomeDir(); err == nil {
			globalDir = filepath.Join(home, ".config", "glitch", "plugins")
		}

		plugins := discoverAllPlugins(localDir, globalDir)
		if len(plugins) == 0 {
			fmt.Println("No plugins found.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(tw, "PLUGIN\tSOURCE\tSUBCOMMANDS\n")
		names := make([]string, 0, len(plugins))
		for name := range plugins {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			p := plugins[name]
			fmt.Fprintf(tw, "%s\t%s\t%s\n", p.Name, p.Source, strings.Join(p.Subcommands, ", "))
		}
		return tw.Flush()
	},
}

func discoverAllPlugins(localDir, globalDir string) map[string]*plugin.PluginInfo {
	return plugin.DiscoverPlugins(localDir, globalDir)
}

func init() {
	// Dynamic plugin dispatch — handles "glitch plugin <name> <subcommand> [--flags]"
	pluginCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		// First arg after flag parsing disabled is the subcommand or plugin name
		// Check if it's "list" — already handled by subcommand
		if args[0] == "list" {
			return pluginListCmd.RunE(pluginListCmd, args[1:])
		}

		pluginName := args[0]

		// Handle --help for plugin
		if len(args) == 1 || (len(args) == 2 && args[1] == "--help") {
			return showPluginHelp(pluginName)
		}

		subcommand := args[1]

		// Handle --help for subcommand
		if len(args) == 3 && args[2] == "--help" {
			return showSubcommandHelp(pluginName, subcommand)
		}

		// Parse remaining args as flags
		flagArgs := args[2:]
		flags := parsePluginFlags(flagArgs)

		// Resolve plugin directory
		localDir := ".glitch/plugins"
		globalDir := ""
		if home, err := os.UserHomeDir(); err == nil {
			globalDir = filepath.Join(home, ".config", "glitch", "plugins")
		}

		// Try local first, then global
		for _, dir := range []string{localDir, globalDir} {
			if dir == "" {
				continue
			}
			pluginDir := filepath.Join(dir, pluginName)
			if _, err := os.Stat(pluginDir); err == nil {
				result, err := pipeline.RunPluginSubcommand(dir, pluginName, subcommand, flags)
				if err != nil {
					return err
				}
				fmt.Println(result)
				return nil
			}
		}

		return fmt.Errorf("plugin %q not found, searched: %s, %s", pluginName, localDir, globalDir)
	}
}

func parsePluginFlags(args []string) map[string]string {
	flags := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		name := strings.TrimPrefix(arg, "--")
		// Check if next arg is a value or another flag
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			flags[name] = args[i+1]
			i++
		} else {
			flags[name] = "true"
		}
	}
	return flags
}

func showPluginHelp(name string) error {
	localDir := ".glitch/plugins"
	globalDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		globalDir = filepath.Join(home, ".config", "glitch", "plugins")
	}

	for _, dir := range []string{localDir, globalDir} {
		if dir == "" {
			continue
		}
		pluginDir := filepath.Join(dir, name)
		if _, err := os.Stat(pluginDir); err != nil {
			continue
		}
		manifest, err := plugin.LoadManifest(pluginDir)
		if err != nil {
			return err
		}
		desc := manifest.Description
		if desc == "" {
			desc = name
		}
		ver := ""
		if manifest.Version != "" {
			ver = fmt.Sprintf(" (v%s)", manifest.Version)
		}
		fmt.Printf("%s — %s%s\n\n", name, desc, ver)

		info := plugin.DiscoverPlugins("", dir)
		if p, ok := info[name]; ok {
			fmt.Println("Subcommands:")
			for _, sub := range p.Subcommands {
				// Try to get description from workflow
				subPath := filepath.Join(pluginDir, sub+".glitch")
				subDesc := sub
				if data, err := os.ReadFile(subPath); err == nil {
					if w, err := pipeline.LoadBytes(data, sub+".glitch"); err == nil && w.Description != "" {
						subDesc = w.Description
					}
				}
				fmt.Printf("  %-12s %s\n", sub, subDesc)
			}
		}
		return nil
	}
	return fmt.Errorf("plugin %q not found", name)
}

func showSubcommandHelp(pluginName, subcommand string) error {
	localDir := ".glitch/plugins"
	globalDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		globalDir = filepath.Join(home, ".config", "glitch", "plugins")
	}

	for _, dir := range []string{localDir, globalDir} {
		if dir == "" {
			continue
		}
		subPath := filepath.Join(dir, pluginName, subcommand+".glitch")
		data, err := os.ReadFile(subPath)
		if err != nil {
			continue
		}

		w, err := pipeline.LoadBytes(data, subcommand+".glitch")
		if err != nil {
			return err
		}
		desc := w.Description
		if desc == "" {
			desc = subcommand
		}
		fmt.Printf("%s %s — %s\n\n", pluginName, subcommand, desc)

		argDefs, err := plugin.ParseArgs(data)
		if err != nil {
			return err
		}
		if len(argDefs) > 0 {
			fmt.Println("Flags:")
			for _, ad := range argDefs {
				def := ""
				if ad.Default != "" {
					def = fmt.Sprintf(" (default: %s)", ad.Default)
				}
				fmt.Printf("  --%-12s %s%s\n", ad.Name, ad.Description, def)
			}
		}
		return nil
	}
	return fmt.Errorf("plugin %q subcommand %q not found", pluginName, subcommand)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -v -run TestPlugin`
Expected: PASS

- [ ] **Step 5: Run ALL tests to verify no regressions**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -20`
Expected: PASS (or only pre-existing failures)

- [ ] **Step 6: Commit**

```bash
git add cmd/plugin.go cmd/plugin_test.go
git commit -m "feat: glitch plugin CLI — list, help, dynamic subcommand dispatch"
```

---

### Task 9: Integration Test — End-to-End Plugin

Create a real plugin and verify the full path: discovery → manifest → args → execution → output.

**Files:**
- Create: `internal/pipeline/plugin_integration_test.go`

- [ ] **Step 1: Write integration test**

```go
// internal/pipeline/plugin_integration_test.go
package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginIntegration_FullPath(t *testing.T) {
	// Set up a complete plugin
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "echo-tool")
	os.MkdirAll(pluginDir, 0o755)

	os.WriteFile(filepath.Join(pluginDir, "plugin.glitch"), []byte(`
(plugin "echo-tool"
  :description "Echo utility"
  :version "1.0.0")

(def prefix "[echo]")
`), 0o644)

	os.WriteFile(filepath.Join(pluginDir, "say.glitch"), []byte(`
(arg "message" :default "hello")
(arg "loud" :type :flag)

(workflow "say"
  :description "Echo a message"
  (step "out"
    (run "echo {{.param.prefix}} {{.param.message}}")))
`), 0o644)

	os.WriteFile(filepath.Join(pluginDir, "count.glitch"), []byte(`
(arg "items")

(workflow "count"
  :description "Count items"
  (step "list"
    (run "echo {{.param.items}}"))
  (step "result"
    (lines "list")))
`), 0o644)

	// Test 1: Basic subcommand with manifest defs
	result, err := RunPluginSubcommand(dir, "echo-tool", "say", map[string]string{"message": "world"})
	if err != nil {
		t.Fatalf("say: %v", err)
	}
	if !strings.Contains(result, "[echo] world") {
		t.Fatalf("expected '[echo] world', got %q", result)
	}

	// Test 2: Default arg value
	result, err = RunPluginSubcommand(dir, "echo-tool", "say", nil)
	if err != nil {
		t.Fatalf("say default: %v", err)
	}
	if !strings.Contains(result, "[echo] hello") {
		t.Fatalf("expected '[echo] hello', got %q", result)
	}

	// Test 3: SDK form (lines) in plugin
	result, err = RunPluginSubcommand(dir, "echo-tool", "count", map[string]string{"items": "a b c"})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if !strings.Contains(result, "[") {
		t.Fatalf("expected JSON array from lines, got %q", result)
	}
}

func TestPluginIntegration_WorkflowInvokesPlugin(t *testing.T) {
	// Set up plugin
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "greet")
	os.MkdirAll(pluginDir, 0o755)

	os.WriteFile(filepath.Join(pluginDir, "hello.glitch"), []byte(`
(arg "name" :default "world")

(workflow "hello"
  (step "out" (run "echo hello {{.param.name}}")))
`), 0o644)

	// Set up a workflow that calls the plugin
	// We need to be in a directory where .glitch/plugins resolves to our temp dir
	// For this test, directly test RunPluginSubcommand which is what executePluginCall uses
	result, err := RunPluginSubcommand(dir, "greet", "hello", map[string]string{"name": "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result) != "hello alice" {
		t.Fatalf("expected %q, got %q", "hello alice", result)
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestPluginIntegration`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -20`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/pipeline/plugin_integration_test.go
git commit -m "test: end-to-end plugin integration tests"
```

---

### Task 10: Build Verification and Cleanup

Verify the binary builds, run smoke tests, clean up any issues.

- [ ] **Step 1: Build the binary**

Run: `cd /Users/stokes/Projects/gl1tch && go build -o gl1tch .`
Expected: Clean build, no errors

- [ ] **Step 2: Verify glitch plugin list works**

Run: `cd /Users/stokes/Projects/gl1tch && ./gl1tch plugin list`
Expected: Either "No plugins found." or lists any plugins in ~/.config/glitch/plugins/

- [ ] **Step 3: Run full test suite one final time**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -30`
Expected: All new tests PASS, no regressions

- [ ] **Step 4: Clean up build artifact**

Run: `rm -f /Users/stokes/Projects/gl1tch/gl1tch`

- [ ] **Step 5: Final commit if any cleanup was needed**

Only if changes were made during cleanup.

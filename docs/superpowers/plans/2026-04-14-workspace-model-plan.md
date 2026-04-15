# Workspace Model Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `--workspace` flag to glitch that makes a directory the container for workflow discovery and result storage, with consistent `<org>/<repo>/<issue|pr>-<number>` result paths and a README.md rollup artifact.

**Architecture:** A persistent `--workspace` flag on the root cobra command threads a workspace path through to workflow loading (`loadWorkflows`) and result storage (`resultDir`, `saveResults`). When set, workflows resolve from `<workspace>/workflows/` and results write to `<workspace>/results/` with the new path convention. A `WorkflowsDir` config field provides an override escape hatch. A new `WriteReadme` function generates the rollup artifact.

**Tech Stack:** Go, cobra, existing pipeline/research/batch packages

---

### Task 1: Add `--workspace` persistent flag to root command

**Files:**
- Modify: `cmd/root.go:13-24`

- [ ] **Step 1: Write the failing test**

Create `cmd/root_test.go`:

```go
package cmd

import (
	"testing"
)

func TestWorkspaceFlag(t *testing.T) {
	// Reset for test
	workspacePath = ""

	rootCmd.SetArgs([]string{"--workspace", "/tmp/test-ws", "version"})
	if err := rootCmd.Execute(); err != nil {
		// version command may not exist yet, that's ok — we just check the flag was parsed
	}

	if workspacePath != "/tmp/test-ws" {
		t.Fatalf("workspacePath: got %q, want /tmp/test-ws", workspacePath)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run TestWorkspaceFlag -v`
Expected: FAIL — `workspacePath` undefined

- [ ] **Step 3: Write minimal implementation**

In `cmd/root.go`, add the global variable and persistent flag:

```go
var (
	targetPath    string
	workspacePath string
)

var providerReg *provider.ProviderRegistry

func init() {
	rootCmd.PersistentFlags().StringVar(&workspacePath, "workspace", "", "workspace directory for workflows and results")

	if home, err := os.UserHomeDir(); err == nil {
		providerReg, _ = provider.LoadProviders(filepath.Join(home, ".config", "glitch", "providers"))
	}
	if providerReg == nil {
		providerReg, _ = provider.LoadProviders("")
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run TestWorkspaceFlag -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/root.go cmd/root_test.go
git commit -m "feat: add --workspace persistent flag to root command"
```

---

### Task 2: Add `WorkflowsDir` to config

**Files:**
- Modify: `cmd/config.go:19-25`
- Modify: `cmd/config.go:68-87` (configSetCmd)

- [ ] **Step 1: Write the failing test**

Add to `cmd/config_test.go` (create if needed):

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_WorkflowsDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("default_model: qwen3:8b\nworkflows_dir: /custom/workflows\n"), 0o644)

	cfg, err := loadConfigFrom(cfgPath)
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if cfg.WorkflowsDir != "/custom/workflows" {
		t.Fatalf("WorkflowsDir: got %q, want /custom/workflows", cfg.WorkflowsDir)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run TestLoadConfig_WorkflowsDir -v`
Expected: FAIL — `WorkflowsDir` field doesn't exist

- [ ] **Step 3: Write minimal implementation**

Add the field to the Config struct in `cmd/config.go`:

```go
type Config struct {
	DefaultModel    string                    `yaml:"default_model"`
	DefaultProvider string                    `yaml:"default_provider"`
	EvalThreshold   int                       `yaml:"eval_threshold,omitempty"`
	WorkflowsDir    string                    `yaml:"workflows_dir,omitempty"`
	Tiers           []provider.TierConfig     `yaml:"tiers,omitempty"`
	Providers       map[string]ProviderConfig `yaml:"providers,omitempty"`
}
```

Also add `workflows_dir` to the `configSetCmd` switch in `cmd/config.go`:

```go
switch args[0] {
case "default_model":
	cfg.DefaultModel = args[1]
case "default_provider":
	cfg.DefaultProvider = args[1]
case "workflows_dir":
	cfg.WorkflowsDir = args[1]
default:
	return fmt.Errorf("unknown config key: %s", args[0])
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run TestLoadConfig_WorkflowsDir -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/config.go cmd/config_test.go
git commit -m "feat: add workflows_dir config field"
```

---

### Task 3: Update `loadWorkflows()` to respect workspace and config override

**Files:**
- Modify: `cmd/ask.go:266-282`

- [ ] **Step 1: Write the failing test**

Add to `cmd/root_test.go`:

```go
func TestLoadWorkflows_Workspace(t *testing.T) {
	// Set up a workspace with a workflow
	wsDir := t.TempDir()
	wfDir := filepath.Join(wsDir, "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "test-wf.yaml"), []byte("name: test-wf\ndescription: test workflow\nsteps: []\n"), 0o644)

	// Set workspace
	workspacePath = wsDir
	defer func() { workspacePath = "" }()

	workflows, err := loadWorkflows()
	if err != nil {
		t.Fatalf("loadWorkflows: %v", err)
	}
	if _, ok := workflows["test-wf"]; !ok {
		t.Fatal("expected test-wf workflow from workspace")
	}
}

func TestLoadWorkflows_ConfigOverride(t *testing.T) {
	// Set up a custom workflows dir
	customDir := t.TempDir()
	os.WriteFile(filepath.Join(customDir, "custom-wf.yaml"), []byte("name: custom-wf\ndescription: custom\nsteps: []\n"), 0o644)

	// Set workspace (its workflows/ dir is empty — config override should win)
	wsDir := t.TempDir()
	os.MkdirAll(filepath.Join(wsDir, "workflows"), 0o755)
	workspacePath = wsDir
	defer func() { workspacePath = "" }()

	// Write temp config with workflows_dir
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	os.WriteFile(cfgPath, []byte("default_model: qwen3:8b\nworkflows_dir: "+customDir+"\n"), 0o644)

	cfg, _ := loadConfigFrom(cfgPath)
	// We need to test the resolution logic, so call the helper directly
	wfDir := resolveWorkflowsDir(cfg)
	if wfDir != customDir {
		t.Fatalf("resolveWorkflowsDir: got %q, want %q", wfDir, customDir)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run "TestLoadWorkflows_Workspace|TestLoadWorkflows_ConfigOverride" -v`
Expected: FAIL — `resolveWorkflowsDir` undefined, workspace logic not implemented

- [ ] **Step 3: Write minimal implementation**

Replace `loadWorkflows()` in `cmd/ask.go` and add `resolveWorkflowsDir`:

```go
// resolveWorkflowsDir returns the workflows directory based on workspace and config.
// Priority: config.WorkflowsDir > <workspace>/workflows/ > global (~/.config/glitch/workflows)
func resolveWorkflowsDir(cfg *Config) string {
	if cfg != nil && cfg.WorkflowsDir != "" {
		return cfg.WorkflowsDir
	}
	if workspacePath != "" {
		return filepath.Join(workspacePath, "workflows")
	}
	return ""
}

func loadWorkflows() (map[string]*pipeline.Workflow, error) {
	cfg, _ := loadConfig()
	wfDir := resolveWorkflowsDir(cfg)

	// If workspace mode: single source only
	if workspacePath != "" {
		if wfDir == "" {
			wfDir = filepath.Join(workspacePath, "workflows")
		}
		workflows, err := pipeline.LoadDir(wfDir)
		if err != nil {
			return nil, err
		}
		if workflows == nil {
			workflows = make(map[string]*pipeline.Workflow)
		}
		return workflows, nil
	}

	// Non-workspace mode: global then local (existing behavior)
	workflows := make(map[string]*pipeline.Workflow)
	if home, err := os.UserHomeDir(); err == nil {
		globalDir := home + "/.config/glitch/workflows"
		if m, err := pipeline.LoadDir(globalDir); err == nil {
			for k, v := range m {
				workflows[k] = v
			}
		}
	}
	if m, err := pipeline.LoadDir(".glitch/workflows"); err == nil {
		for k, v := range m {
			workflows[k] = v
		}
	}
	return workflows, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run "TestLoadWorkflows_Workspace|TestLoadWorkflows_ConfigOverride" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/ask.go cmd/root_test.go
git commit -m "feat: workspace-aware workflow resolution with config override"
```

---

### Task 4: Update `resultDir()` to use `issue-`/`pr-` prefix from Source field

**Files:**
- Modify: `internal/research/results.go:60-81`
- Modify: `internal/research/results_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/research/results_test.go`:

```go
func TestResultDir_IssuePrefix(t *testing.T) {
	result := LoopResult{
		Document: ResearchDocument{
			Source:   "github_issue",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "872"},
		},
	}
	got := resultDir("/base", result)
	want := "/base/elastic/ensemble/issue-872"
	if got != want {
		t.Fatalf("resultDir: got %q, want %q", got, want)
	}
}

func TestResultDir_PRPrefix(t *testing.T) {
	result := LoopResult{
		Document: ResearchDocument{
			Source:   "github_pr",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "100"},
		},
	}
	got := resultDir("/base", result)
	want := "/base/elastic/ensemble/pr-100"
	if got != want {
		t.Fatalf("resultDir: got %q, want %q", got, want)
	}
}

func TestResultDir_NoSourceFallback(t *testing.T) {
	result := LoopResult{
		Document: ResearchDocument{
			Source:   "text",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "50"},
		},
	}
	got := resultDir("/base", result)
	want := "/base/elastic/ensemble/issue-50"
	if got != want {
		t.Fatalf("resultDir: got %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run "TestResultDir_IssuePrefix|TestResultDir_PRPrefix|TestResultDir_NoSourceFallback" -v`
Expected: FAIL — current output is `elastic/ensemble/872` without prefix

- [ ] **Step 3: Write minimal implementation**

Update `resultDir()` in `internal/research/results.go`:

```go
// resultDir computes the output directory path from the document metadata.
// Pattern: baseDir/<org>/<repo>/<issue|pr>-<number>
func resultDir(baseDir string, result LoopResult) string {
	repo := result.Document.Repo
	number := result.Document.Metadata["number"]

	if repo == "" {
		return filepath.Join(baseDir, "general")
	}

	parts := strings.SplitN(repo, "/", 2)
	org := parts[0]
	repoName := "general"
	if len(parts) > 1 {
		repoName = parts[1]
	}

	if number == "" {
		return filepath.Join(baseDir, org, repoName)
	}

	prefix := "issue"
	if result.Document.Source == "github_pr" {
		prefix = "pr"
	}
	return filepath.Join(baseDir, org, repoName, prefix+"-"+number)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run "TestResultDir_IssuePrefix|TestResultDir_PRPrefix|TestResultDir_NoSourceFallback" -v`
Expected: PASS

- [ ] **Step 5: Update existing tests for new path convention**

The existing `TestSaveLoopResult` and `TestSaveLoopResultImplement` tests hardcode `elastic/ensemble/872` and `elastic/ensemble/100`. Update them to expect `elastic/ensemble/issue-872` and `elastic/ensemble/issue-100`:

In `TestSaveLoopResult` (line 68):
```go
dir := filepath.Join(base, "elastic", "ensemble", "issue-872")
```

In `TestSaveLoopResultImplement` (line 124):
```go
dir := filepath.Join(base, "elastic", "ensemble", "issue-100")
```

- [ ] **Step 6: Run all research tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add internal/research/results.go internal/research/results_test.go
git commit -m "feat: use issue-/pr- prefix in result directory paths"
```

---

### Task 5: Update `run.json` schema with standardized fields

**Files:**
- Modify: `internal/research/results.go:44-58`

- [ ] **Step 1: Write the failing test**

Add to `internal/research/results_test.go`:

```go
func TestRunJSON_StandardFields(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-std-001",
		Document: ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/50",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "50"},
		},
		Goal:     GoalSummarize,
		Output:   "summary text here, long enough to be substantive" + strings.Repeat(" content", 100),
		LLMCalls: 2,
		Duration: 5 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	dir := filepath.Join(base, "elastic", "ensemble", "issue-50")
	data, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatalf("read run.json: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Check new standardized fields
	if raw["repo"] != "elastic/ensemble" {
		t.Fatalf("repo: got %v", raw["repo"])
	}
	if raw["ref_type"] != "issue" {
		t.Fatalf("ref_type: got %v", raw["ref_type"])
	}
	if raw["ref_number"] != float64(50) {
		t.Fatalf("ref_number: got %v", raw["ref_number"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run TestRunJSON_StandardFields -v`
Expected: FAIL — `repo`, `ref_type`, `ref_number` fields missing from run.json

- [ ] **Step 3: Write minimal implementation**

Update the `runJSON` struct and its population in `internal/research/results.go`:

```go
// runJSON is the metadata structure written to run.json.
type runJSON struct {
	RunID       string  `json:"run_id"`
	Repo        string  `json:"repo"`
	RefType     string  `json:"ref_type"`
	RefNumber   int     `json:"ref_number"`
	Source      string  `json:"source"`
	SourceURL   string  `json:"source_url"`
	Goal        Goal    `json:"goal"`
	ToolCalls   int     `json:"tool_calls"`
	LLMCalls    int     `json:"llm_calls"`
	TokensIn    int     `json:"tokens_in"`
	TokensOut   int     `json:"tokens_out"`
	CostUSD     float64 `json:"cost_usd"`
	MaxTier     int     `json:"max_tier"`
	Escalations int     `json:"escalations"`
	DurationMS  int64   `json:"duration_ms"`
}
```

Update the `SaveLoopResult` function where `meta` is built (around line 101). Add a helper to parse ref info and update the struct population:

```go
	refType := "issue"
	if result.Document.Source == "github_pr" {
		refType = "pr"
	}
	refNumber := 0
	if n := result.Document.Metadata["number"]; n != "" {
		fmt.Sscanf(n, "%d", &refNumber)
	}

	meta := runJSON{
		RunID:       result.RunID,
		Repo:        result.Document.Repo,
		RefType:     refType,
		RefNumber:   refNumber,
		Source:      result.Document.Source,
		SourceURL:   result.Document.SourceURL,
		Goal:        result.Goal,
		ToolCalls:   len(result.ToolCalls),
		LLMCalls:    result.LLMCalls,
		TokensIn:    result.TokensIn,
		TokensOut:   result.TokensOut,
		CostUSD:     result.CostUSD,
		MaxTier:     result.MaxTier,
		Escalations: result.Escalations,
		DurationMS:  result.Duration.Milliseconds(),
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run TestRunJSON_StandardFields -v`
Expected: PASS

- [ ] **Step 5: Run all research tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/research/results.go internal/research/results_test.go
git commit -m "feat: add repo, ref_type, ref_number to run.json schema"
```

---

### Task 6: Generate README.md rollup artifact

**Files:**
- Modify: `internal/research/results.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/research/results_test.go`:

```go
func TestSaveLoopResult_WritesReadme(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-readme-001",
		Document: ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/42",
			Title:     "Fix flaky CI test",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "42"},
		},
		Goal:   GoalSummarize,
		Output: "# Summary\n\nThe CI test is flaky because of a race condition.\n\n## Recommendation\n\nAdd a mutex around the shared state.\n\n## Response Draft\n\nI investigated the flaky CI test and found a race condition in the shared state handler.",
		ToolCalls: []ToolResult{
			{Tool: "grep_code", Output: "found race"},
		},
		LLMCalls: 2,
		Duration: 3 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	dir := filepath.Join(base, "elastic", "ensemble", "issue-42")
	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal("README.md not created")
	}

	content := string(readme)
	// Check frontmatter fields
	if !strings.Contains(content, "repo: elastic/ensemble") {
		t.Fatal("README.md missing repo frontmatter")
	}
	if !strings.Contains(content, "ref: issue-42") {
		t.Fatal("README.md missing ref frontmatter")
	}
	if !strings.Contains(content, "title: \"Fix flaky CI test\"") {
		t.Fatal("README.md missing title frontmatter")
	}
	// Check body content
	if !strings.Contains(content, "race condition") {
		t.Fatal("README.md missing output content")
	}
	// Check evidence index
	if !strings.Contains(content, "001-grep_code.txt") {
		t.Fatal("README.md missing evidence index")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run TestSaveLoopResult_WritesReadme -v`
Expected: FAIL — README.md not created

- [ ] **Step 3: Write minimal implementation**

Add `writeReadme` function and call it from `SaveLoopResult` in `internal/research/results.go`:

```go
// writeReadme generates a README.md rollup artifact with frontmatter and content.
func writeReadme(dir string, result LoopResult) error {
	refType := "issue"
	if result.Document.Source == "github_pr" {
		refType = "pr"
	}
	number := result.Document.Metadata["number"]
	ref := refType + "-" + number

	status := "researched"
	if result.Goal == GoalImplement {
		status = "planned"
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	fmt.Fprintf(&buf, "repo: %s\n", result.Document.Repo)
	fmt.Fprintf(&buf, "ref: %s\n", ref)
	fmt.Fprintf(&buf, "title: %q\n", result.Document.Title)
	fmt.Fprintf(&buf, "status: %s\n", status)
	fmt.Fprintf(&buf, "source_url: %s\n", result.Document.SourceURL)
	buf.WriteString("---\n\n")

	buf.WriteString(result.Output)

	if len(result.ToolCalls) > 0 {
		buf.WriteString("\n\n## Evidence Index\n\n")
		for i, tc := range result.ToolCalls {
			name := fmt.Sprintf("%03d-%s.txt", i+1, tc.Tool)
			fmt.Fprintf(&buf, "- [%s](evidence/%s)\n", name, name)
		}
	}

	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(buf.String()), 0o644)
}
```

Call it at the end of `SaveLoopResult`, just before `return nil`:

```go
	if err := writeReadme(dir, result); err != nil {
		return fmt.Errorf("results: write README.md: %w", err)
	}

	return nil
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run TestSaveLoopResult_WritesReadme -v`
Expected: PASS

- [ ] **Step 5: Run all research tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/research/results.go internal/research/results_test.go
git commit -m "feat: generate README.md rollup artifact in result directories"
```

---

### Task 7: Wire workspace into `ask` command result paths

**Files:**
- Modify: `cmd/ask.go:114-122` (batch result path)
- Modify: `cmd/ask.go:188-197` (research loop result path)
- Modify: `cmd/ask.go:235-245` (single issue result path)

- [ ] **Step 1: Add helper to resolve results dir**

Add `resolveResultsDir` to `cmd/ask.go`:

```go
// resolveResultsDir returns the results directory based on flags and workspace.
// Priority: --results-dir flag > <workspace>/results/ > CWD/.glitch/results
func resolveResultsDir() string {
	if askResultsDir != "" {
		return askResultsDir
	}
	if workspacePath != "" {
		return filepath.Join(workspacePath, "results")
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".glitch", "results")
}
```

- [ ] **Step 2: Replace all inline resultsDir computations**

In the batch result path (around line 114):
```go
rdir := resolveResultsDir()
fmt.Printf("\nResults ready:\n")
```

Pass to batch.Run (around line 92):
```go
ResultsDir: resolveResultsDir(),
```

In the research loop path (around line 188):
```go
if research.IsSubstantive(result.Output) {
	resultsBase := resolveResultsDir()
	if err := research.SaveLoopResult(resultsBase, result); err != nil {
```

In the single issue path (around line 235):
```go
rdir := resolveResultsDir()
```

- [ ] **Step 3: Run existing tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -v`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/ask.go
git commit -m "feat: wire workspace into ask command result path resolution"
```

---

### Task 8: Update batch result paths to use `<org>/<repo>/<issue|pr>-<number>` convention

**Files:**
- Modify: `internal/batch/batch.go:140-177`

- [ ] **Step 1: Write the failing test**

Create `internal/batch/batch_test.go`:

```go
package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResultPath_Convention(t *testing.T) {
	dir := t.TempDir()
	repo := "elastic/observability-robots"
	issue := "3920"
	variant := "claude"
	iter := 1

	resultPath := resultPath(dir, repo, issue, variant, iter)
	want := filepath.Join(dir, "elastic", "observability-robots", "issue-3920", "iteration-1", "claude")
	if resultPath != want {
		t.Fatalf("resultPath: got %q, want %q", resultPath, want)
	}
}

func TestResultPath_NoVariant(t *testing.T) {
	dir := t.TempDir()
	resultPath := resultPath(dir, "elastic/ensemble", "100", "", 1)
	want := filepath.Join(dir, "elastic", "ensemble", "issue-100", "iteration-1")
	if resultPath != want {
		t.Fatalf("resultPath: got %q, want %q", resultPath, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/batch/ -run "TestResultPath" -v`
Expected: FAIL — `resultPath` undefined

- [ ] **Step 3: Write minimal implementation**

Add `resultPath` function and update `saveResults` in `internal/batch/batch.go`:

```go
// resultPath computes the result directory for a batch issue run.
// Pattern: baseDir/<org>/<repo>/issue-<number>/iteration-<n>[/<variant>]
func resultPath(baseDir, repo, issue, variant string, iter int) string {
	parts := strings.SplitN(repo, "/", 2)
	org := parts[0]
	repoName := "general"
	if len(parts) > 1 {
		repoName = parts[1]
	}

	dir := filepath.Join(baseDir, org, repoName, "issue-"+issue, fmt.Sprintf("iteration-%d", iter))
	if variant != "" {
		dir = filepath.Join(dir, variant)
	}
	return dir
}

func saveResults(resultsBase, issue, variant string, iter int, repo string, result *pipeline.Result) {
	dir := resultPath(resultsBase, repo, issue, variant, iter)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: mkdir %s: %v\n", dir, err)
		return
	}

	// ... rest of saveResults unchanged ...
```

Also update the cross-review path (around line 110):

```go
	crParts := strings.SplitN(opts.Repo, "/", 2)
	crOrg := crParts[0]
	crRepo := "general"
	if len(crParts) > 1 {
		crRepo = crParts[1]
	}
	crDir := filepath.Join(resultsBase, crOrg, crRepo, "issue-"+issue, fmt.Sprintf("iteration-%d", iter))
	os.MkdirAll(crDir, 0o755)
	os.WriteFile(filepath.Join(crDir, "cross-review.md"), []byte(cr), 0o644)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/batch/ -run "TestResultPath" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/batch/batch.go internal/batch/batch_test.go
git commit -m "feat: use org/repo/issue-N convention in batch result paths"
```

---

### Task 9: End-to-end integration test

**Files:**
- Create: `cmd/workspace_test.go`

- [ ] **Step 1: Write integration test**

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/research"
)

func TestWorkspaceIntegration(t *testing.T) {
	wsDir := t.TempDir()

	// Create workspace structure
	wfDir := filepath.Join(wsDir, "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "test.yaml"), []byte("name: test\ndescription: test\nsteps: []\n"), 0o644)

	// Set workspace
	workspacePath = wsDir
	defer func() { workspacePath = "" }()

	// Verify workflow resolution
	wfs, err := loadWorkflows()
	if err != nil {
		t.Fatalf("loadWorkflows: %v", err)
	}
	if _, ok := wfs["test"]; !ok {
		t.Fatal("workspace workflow not found")
	}

	// Verify result path resolution
	rdir := resolveResultsDir()
	expected := filepath.Join(wsDir, "results")
	if rdir != expected {
		t.Fatalf("resolveResultsDir: got %q, want %q", rdir, expected)
	}

	// Verify SaveLoopResult writes to workspace
	result := research.LoopResult{
		RunID: "ws-test-001",
		Document: research.ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/99",
			Title:     "Test issue",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "99"},
		},
		Goal:   research.GoalSummarize,
		Output: "Test summary with enough content to be substantive." + string(make([]byte, 500)),
	}

	if err := research.SaveLoopResult(rdir, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	// Check result landed in workspace
	resultDir := filepath.Join(wsDir, "results", "elastic", "ensemble", "issue-99")
	if _, err := os.Stat(filepath.Join(resultDir, "README.md")); err != nil {
		t.Fatal("README.md not in workspace results")
	}
	if _, err := os.Stat(filepath.Join(resultDir, "run.json")); err != nil {
		t.Fatal("run.json not in workspace results")
	}
}
```

- [ ] **Step 2: Run integration test**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run TestWorkspaceIntegration -v`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -20`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/workspace_test.go
git commit -m "test: add workspace integration test"
```

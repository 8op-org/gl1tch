# Work-on-Issue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an end-to-end `work-on-issue` workflow and a YAML-based provider registry so `glitch ask work on issue 3442` fetches context, runs any AI coding tool, and produces a results folder ready for PR creation.

**Architecture:** Replace hardcoded provider functions with a YAML registry loaded from `~/.config/glitch/providers/`. Add structured params to the pipeline runner so the router can pass repo + issue number. Add a `work-on-issue` workflow in stokagent that chains 8 steps: fetch → classify → enrich → branch → implement → results.

**Tech Stack:** Go, text/template, YAML, gh CLI, Elasticsearch (optional)

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/provider/provider.go` | Modify | Add `ProviderRegistry`, `LoadProviders()`, `RunProvider()`. Remove `RunClaude`. |
| `internal/provider/provider_test.go` | Create | Tests for provider loading and command rendering |
| `internal/pipeline/runner.go` | Modify | Use provider registry instead of switch. Accept `params` map. |
| `internal/pipeline/runner_test.go` | Create | Tests for param template rendering |
| `internal/router/router.go` | Modify | Add `work on issue` fast-path, issue ref parsing, return params |
| `internal/router/router_test.go` | Create | Tests for issue reference parsing |
| `cmd/ask.go` | Modify | Thread params from router to pipeline.Run |
| `cmd/plugin.go` | Delete | Plugin system removed |
| `~/.config/glitch/providers/claude.yaml` | Create | Claude provider command template |
| `~/.config/glitch/providers/codex.yaml` | Create | Codex provider command template |
| `~/.config/glitch/providers/copilot.yaml` | Create | Copilot provider command template |
| `~/.config/glitch/providers/gemini.yaml` | Create | Gemini provider command template |
| `~/.config/glitch/providers/opencode.yaml` | Create | OpenCode provider command template |
| `~/Projects/stokagent/workflows/work-on-issue.yaml` | Create | The workflow |

---

### Task 1: Provider Registry

**Files:**
- Modify: `internal/provider/provider.go`
- Create: `internal/provider/provider_test.go`

- [ ] **Step 1: Write failing test for LoadProviders**

```go
// internal/provider/provider_test.go
package provider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProviders_ReadsYAMLFiles(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "claude.yaml"), []byte(`
name: claude
command: claude -p --output-format text "{{.prompt}}"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "codex.yaml"), []byte(`
name: codex
command: codex -p --full-auto "{{.prompt}}"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	reg, err := LoadProviders(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(reg.providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(reg.providers))
	}

	if reg.providers["claude"].Command == "" {
		t.Fatal("claude provider command is empty")
	}
	if reg.providers["codex"].Command == "" {
		t.Fatal("codex provider command is empty")
	}
}

func TestLoadProviders_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	reg, err := LoadProviders(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.providers) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(reg.providers))
	}
}

func TestLoadProviders_MissingDir(t *testing.T) {
	reg, err := LoadProviders("/nonexistent/path")
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.providers) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(reg.providers))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -v -run TestLoadProviders`
Expected: FAIL — `LoadProviders` not defined

- [ ] **Step 3: Implement provider types and LoadProviders**

```go
// Add to internal/provider/provider.go

// Provider is a YAML-defined command template for an LLM tool.
type Provider struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// ProviderRegistry holds loaded provider definitions.
type ProviderRegistry struct {
	providers map[string]*Provider
}

// LoadProviders reads all .yaml files from a directory into a registry.
func LoadProviders(dir string) (*ProviderRegistry, error) {
	reg := &ProviderRegistry{providers: make(map[string]*Provider)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var p Provider
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		reg.providers[p.Name] = &p
	}

	return reg, nil
}
```

Add these imports to the import block: `"os"`, `"path/filepath"`, `"gopkg.in/yaml.v3"`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -v -run TestLoadProviders`
Expected: PASS

- [ ] **Step 5: Write failing test for RunProvider**

```go
// Add to internal/provider/provider_test.go

func TestRenderProviderCommand(t *testing.T) {
	reg := &ProviderRegistry{
		providers: map[string]*Provider{
			"echo-test": {
				Name:    "echo-test",
				Command: `echo "{{.prompt}}"`,
			},
		},
	}

	cmd, err := reg.RenderCommand("echo-test", "hello world")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != `echo "hello world"` {
		t.Fatalf("expected 'echo \"hello world\"', got %q", cmd)
	}
}

func TestRenderProviderCommand_NotFound(t *testing.T) {
	reg := &ProviderRegistry{providers: make(map[string]*Provider)}

	_, err := reg.RenderCommand("missing", "hello")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestRunProvider_ExecsCommand(t *testing.T) {
	reg := &ProviderRegistry{
		providers: map[string]*Provider{
			"echo-test": {
				Name:    "echo-test",
				Command: `echo "{{.prompt}}"`,
			},
		},
	}

	out, err := reg.RunProvider("echo-test", "hello world")
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello world" {
		t.Fatalf("expected 'hello world', got %q", out)
	}
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -v -run "TestRenderProviderCommand|TestRunProvider"`
Expected: FAIL — `RenderCommand` and `RunProvider` not defined

- [ ] **Step 7: Implement RenderCommand and RunProvider**

```go
// Add to internal/provider/provider.go

// RenderCommand renders a provider's command template with the given prompt.
func (r *ProviderRegistry) RenderCommand(name, prompt string) (string, error) {
	p, ok := r.providers[name]
	if !ok {
		available := make([]string, 0, len(r.providers))
		for k := range r.providers {
			available = append(available, k)
		}
		return "", fmt.Errorf("provider %q not found (available: %s)", name, strings.Join(available, ", "))
	}

	t, err := template.New("").Parse(p.Command)
	if err != nil {
		return "", fmt.Errorf("provider %s: template: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{"prompt": prompt}); err != nil {
		return "", fmt.Errorf("provider %s: render: %w", name, err)
	}
	return buf.String(), nil
}

// RunProvider renders and executes a provider command, returning stdout.
func (r *ProviderRegistry) RunProvider(name, prompt string) (string, error) {
	cmd, err := r.RenderCommand(name, prompt)
	if err != nil {
		return "", err
	}
	return RunShell(cmd)
}
```

Add `"bytes"`, `"strings"`, `"text/template"` to imports.

- [ ] **Step 8: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -v`
Expected: PASS

- [ ] **Step 9: Remove RunClaude**

Delete the `RunClaude` function from `internal/provider/provider.go` (lines 54-66).

- [ ] **Step 10: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./... && go test ./...`
Expected: Build may fail if `RunClaude` is referenced elsewhere. Check and fix in next task.

- [ ] **Step 11: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/provider/provider.go internal/provider/provider_test.go
git commit -m "feat: add YAML-based provider registry, remove RunClaude"
```

---

### Task 2: Pipeline Runner — Provider Registry + Params

**Files:**
- Modify: `internal/pipeline/runner.go`
- Create: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write failing test for param support in templates**

```go
// internal/pipeline/runner_test.go
package pipeline

import (
	"testing"
)

func TestRender_WithParams(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"input": "work on issue 3442",
		"param": map[string]string{
			"repo":  "elastic/observability-robots",
			"issue": "3442",
		},
	}

	result, err := render(`gh issue view {{.param.issue}} --repo {{.param.repo}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}

	expected := "gh issue view 3442 --repo elastic/observability-robots"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestRender_WithStepRefs(t *testing.T) {
	steps := map[string]string{
		"fetch": `{"title": "fix bug"}`,
	}
	data := map[string]any{
		"input": "test",
	}

	result, err := render(`Issue: {{step "fetch"}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}

	expected := `Issue: {"title": "fix bug"}`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}
```

- [ ] **Step 2: Run test to verify it passes (render already supports map data)**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -run TestRender`
Expected: PASS — the existing `render` function already uses `map[string]any` data, so `{{.param.issue}}` should work if the data map contains a `"param"` key. Verify this.

- [ ] **Step 3: Update Run signature to accept params and registry**

Change `Run` in `internal/pipeline/runner.go` from:

```go
func Run(w *Workflow, input string, defaultModel string) (*Result, error) {
```

to:

```go
func Run(w *Workflow, input string, defaultModel string, params map[string]string, reg *provider.ProviderRegistry) (*Result, error) {
```

Update the data map construction:

```go
data := map[string]any{
    "input": input,
    "param": params,
}
```

Replace the provider switch (lines 47-63) with:

```go
prov := strings.ToLower(step.LLM.Provider)
model := step.LLM.Model
if model == "" {
    model = defaultModel
}

var out string
switch prov {
case "ollama", "":
    if model == "" {
        model = "qwen2.5:7b"
    }
    out, err = provider.RunOllama(model, rendered)
default:
    out, err = reg.RunProvider(prov, rendered)
}
```

Add `"github.com/8op-org/gl1tch/internal/provider"` import (it's already imported but may need the registry type).

- [ ] **Step 4: Update all callers of pipeline.Run**

In `cmd/ask.go` line 49, change:

```go
result, err := pipeline.Run(w, resolved, "")
```

to:

```go
result, err := pipeline.Run(w, resolved, "", params, providerReg)
```

Where `params` and `providerReg` come from the router (wired in Task 4).

In `cmd/workflow.go` line 84, change:

```go
result, err := pipeline.Run(w, input, "")
```

to:

```go
result, err := pipeline.Run(w, input, "", nil, providerReg)
```

The `providerReg` needs to be loaded once at startup. Add to `cmd/root.go`:

```go
var providerReg *provider.ProviderRegistry

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		providerReg, _ = provider.LoadProviders(filepath.Join(home, ".config", "glitch", "providers"))
	}
	if providerReg == nil {
		providerReg, _ = provider.LoadProviders("")
	}
}
```

Add imports: `"path/filepath"`, `"github.com/8op-org/gl1tch/internal/provider"`.

- [ ] **Step 5: Build and verify**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./... && go test ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/pipeline/runner.go internal/pipeline/runner_test.go cmd/ask.go cmd/workflow.go cmd/root.go
git commit -m "feat: pipeline runner uses provider registry and supports params"
```

---

### Task 3: Router — Issue Reference Parsing

**Files:**
- Modify: `internal/router/router.go`
- Create: `internal/router/router_test.go`

- [ ] **Step 1: Write failing tests for issue reference parsing**

```go
// internal/router/router_test.go
package router

import (
	"testing"
)

func TestParseIssueRef_BareNumber(t *testing.T) {
	repo, issue, ok := ParseIssueRef("3442")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "" {
		t.Fatalf("expected empty repo, got %q", repo)
	}
	if issue != "3442" {
		t.Fatalf("expected issue 3442, got %q", issue)
	}
}

func TestParseIssueRef_ShortForm(t *testing.T) {
	repo, issue, ok := ParseIssueRef("observability-robots#3442")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "elastic/observability-robots" {
		t.Fatalf("expected elastic/observability-robots, got %q", repo)
	}
	if issue != "3442" {
		t.Fatalf("expected issue 3442, got %q", issue)
	}
}

func TestParseIssueRef_FullForm(t *testing.T) {
	repo, issue, ok := ParseIssueRef("elastic/ensemble#1281")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "elastic/ensemble" {
		t.Fatalf("expected elastic/ensemble, got %q", repo)
	}
	if issue != "1281" {
		t.Fatalf("expected issue 1281, got %q", issue)
	}
}

func TestParseIssueRef_Invalid(t *testing.T) {
	_, _, ok := ParseIssueRef("not an issue")
	if ok {
		t.Fatal("expected not ok")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/router/ -v -run TestParseIssueRef`
Expected: FAIL — `ParseIssueRef` not defined

- [ ] **Step 3: Implement ParseIssueRef**

```go
// Add to internal/router/router.go

const defaultOrg = "elastic"

// reIssueRef matches: "owner/repo#123", "repo#123", or "123"
var reIssueRef = regexp.MustCompile(`^(?:(?:([a-zA-Z0-9_.-]+)/)?([a-zA-Z0-9_.-]+)#)?(\d+)$`)

// ParseIssueRef parses an issue reference into repo and issue number.
// Returns ("", issue, true) for bare numbers — caller resolves repo from git remote.
// Returns ("org/repo", issue, true) for qualified refs.
func ParseIssueRef(ref string) (repo, issue string, ok bool) {
	ref = strings.TrimSpace(ref)
	m := reIssueRef.FindStringSubmatch(ref)
	if m == nil {
		return "", "", false
	}
	owner := m[1]
	repoName := m[2]
	issue = m[3]

	if repoName == "" {
		// Bare number: "3442"
		return "", issue, true
	}
	if owner == "" {
		// Short form: "observability-robots#3442"
		return defaultOrg + "/" + repoName, issue, true
	}
	// Full form: "elastic/ensemble#1281"
	return owner + "/" + repoName, issue, true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/router/ -v -run TestParseIssueRef`
Expected: PASS

- [ ] **Step 5: Write failing test for work-on-issue routing**

```go
// Add to internal/router/router_test.go

func TestMatchWorkOnIssue(t *testing.T) {
	tests := []struct {
		input     string
		wantMatch bool
		wantRef   string
	}{
		{"work on issue 3442", true, "3442"},
		{"work on issue observability-robots#3442", true, "observability-robots#3442"},
		{"work on issue elastic/ensemble#1281", true, "elastic/ensemble#1281"},
		{"what issues are open", false, ""},
		{"list prs", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, ok := MatchWorkOnIssue(tt.input)
			if ok != tt.wantMatch {
				t.Fatalf("match=%v, want %v", ok, tt.wantMatch)
			}
			if ok && ref != tt.wantRef {
				t.Fatalf("ref=%q, want %q", ref, tt.wantRef)
			}
		})
	}
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/router/ -v -run TestMatchWorkOnIssue`
Expected: FAIL — `MatchWorkOnIssue` not defined

- [ ] **Step 7: Implement MatchWorkOnIssue**

```go
// Add to internal/router/router.go

var reWorkOnIssue = regexp.MustCompile(`(?i)work on issue\s+(.+)`)

// MatchWorkOnIssue checks if input matches "work on issue <ref>" and returns the ref string.
func MatchWorkOnIssue(input string) (ref string, ok bool) {
	m := reWorkOnIssue.FindStringSubmatch(input)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/router/ -v -run TestMatchWorkOnIssue`
Expected: PASS

- [ ] **Step 9: Write failing test for ResolveRepo (git remote fallback)**

```go
// Add to internal/router/router_test.go

func TestResolveRepo_FromGitRemote(t *testing.T) {
	// This test requires being in a git repo with a github remote.
	// Use the gl1tch repo itself as the test fixture.
	repo, err := ResolveRepo("")
	if err != nil {
		t.Skipf("not in a git repo with github remote: %v", err)
	}
	if repo == "" {
		t.Fatal("expected non-empty repo")
	}
	// Should be owner/repo format
	if !strings.Contains(repo, "/") {
		t.Fatalf("expected owner/repo format, got %q", repo)
	}
}
```

- [ ] **Step 10: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/router/ -v -run TestResolveRepo`
Expected: FAIL — `ResolveRepo` not defined

- [ ] **Step 11: Implement ResolveRepo**

```go
// Add to internal/router/router.go

// ResolveRepo resolves a repo string. If empty, reads the current git remote origin.
func ResolveRepo(repo string) (string, error) {
	if repo != "" {
		return repo, nil
	}
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("no repo specified and git remote failed: %w", err)
	}
	remote := strings.TrimSpace(string(out))
	// Parse github.com/owner/repo from SSH or HTTPS URL
	remote = strings.TrimSuffix(remote, ".git")
	if i := strings.Index(remote, "github.com"); i >= 0 {
		parts := strings.Split(remote[i:], "/")
		if len(parts) >= 3 {
			return parts[1] + "/" + parts[2], nil
		}
	}
	return "", fmt.Errorf("could not parse owner/repo from remote: %s", remote)
}
```

Add `"os/exec"` to imports.

- [ ] **Step 12: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/router/ -v -run TestResolveRepo`
Expected: PASS (or SKIP if not in a github repo)

- [ ] **Step 13: Update Match to handle work-on-issue and return params**

Change the `Match` function signature from:

```go
func Match(input string, workflows map[string]*pipeline.Workflow, model string) (*pipeline.Workflow, string)
```

to:

```go
func Match(input string, workflows map[string]*pipeline.Workflow, model string) (*pipeline.Workflow, string, map[string]string)
```

Add the work-on-issue fast path at the top of `Match`, before the GitHub URL checks:

```go
// Fast path: work on issue <ref>
if ref, ok := MatchWorkOnIssue(input); ok {
    if w, ok := workflows["work-on-issue"]; ok {
        repo, issue, ok := ParseIssueRef(ref)
        if ok {
            resolved, err := ResolveRepo(repo)
            if err != nil {
                // Fall through to LLM routing
            } else {
                return w, input, map[string]string{
                    "repo":  resolved,
                    "issue": issue,
                }
            }
        }
    }
}
```

Update all return statements in Match to include the third `nil` param value:
- Line with `return w, url` → `return w, url, nil`
- Line with `return nil, input` → `return nil, input, nil`
- Line with `return workflows[name], input` → `return workflows[name], input, nil`

- [ ] **Step 14: Run all tests**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./... && go test ./...`
Expected: PASS

- [ ] **Step 15: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/router/router.go internal/router/router_test.go
git commit -m "feat: add issue reference parsing and work-on-issue routing"
```

---

### Task 4: Wire Router Params Through ask Command

**Files:**
- Modify: `cmd/ask.go`

- [ ] **Step 1: Update ask command to pass params**

Change `cmd/ask.go` to handle the new 3-return Match:

```go
w, resolved, params := router.Match(input, workflows, "")
if w == nil {
    fmt.Fprintf(os.Stderr, "no matching workflow for: %s\n", input)
    fmt.Fprintln(os.Stderr, "available workflows:")
    for name := range workflows {
        fmt.Fprintf(os.Stderr, "  - %s\n", name)
    }
    return nil
}

fmt.Printf(">> %s\n", w.Name)
result, err := pipeline.Run(w, resolved, "", params, providerReg)
```

- [ ] **Step 2: Build and verify**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add cmd/ask.go
git commit -m "feat: thread router params through ask command to pipeline"
```

---

### Task 5: Remove Plugin System

**Files:**
- Delete: `cmd/plugin.go`
- Modify: `cmd/root.go` (if plugin.go's init registers with rootCmd)

- [ ] **Step 1: Delete plugin.go**

```bash
cd /Users/stokes/Projects/gl1tch
rm cmd/plugin.go
```

- [ ] **Step 2: Build and verify no compile errors**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: PASS — plugin.go's init() registered commands on rootCmd, but removing it just means those commands no longer exist.

- [ ] **Step 3: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add -u cmd/plugin.go
git commit -m "feat: remove plugin system, providers replace it"
```

---

### Task 6: Create Provider YAML Files

**Files:**
- Create: `~/.config/glitch/providers/claude.yaml`
- Create: `~/.config/glitch/providers/codex.yaml`
- Create: `~/.config/glitch/providers/copilot.yaml`
- Create: `~/.config/glitch/providers/gemini.yaml`
- Create: `~/.config/glitch/providers/opencode.yaml`

- [ ] **Step 1: Create providers directory and files**

```bash
mkdir -p ~/.config/glitch/providers
```

```yaml
# ~/.config/glitch/providers/claude.yaml
name: claude
command: claude -p --output-format text "{{.prompt}}"
```

```yaml
# ~/.config/glitch/providers/codex.yaml
name: codex
command: codex exec "{{.prompt}}"
```

```yaml
# ~/.config/glitch/providers/copilot.yaml
name: copilot
command: gh copilot suggest "{{.prompt}}"
```

```yaml
# ~/.config/glitch/providers/gemini.yaml
name: gemini
command: gemini -p "{{.prompt}}"
```

```yaml
# ~/.config/glitch/providers/opencode.yaml
name: opencode
command: opencode run "{{.prompt}}"
```

- [ ] **Step 2: Verify providers load**

Run: `cd /Users/stokes/Projects/gl1tch && go run . config show`
Expected: No errors. Then test with a simple echo provider:

```bash
cat > ~/.config/glitch/providers/echo-test.yaml << 'EOF'
name: echo-test
command: echo "{{.prompt}}"
EOF
```

Create a test workflow:
```bash
cat > /tmp/test-provider.yaml << 'EOF'
name: test-provider
description: test provider registry
steps:
  - id: test
    llm:
      provider: echo-test
      prompt: hello from provider registry
EOF
```

Run: `cd /Users/stokes/Projects/gl1tch && go run . workflow run test-provider` (after copying to workflows dir)
Expected: Output includes "hello from provider registry"

- [ ] **Step 3: Clean up test provider**

```bash
rm ~/.config/glitch/providers/echo-test.yaml
rm /tmp/test-provider.yaml
```

- [ ] **Step 4: Commit (nothing to commit in gl1tch repo — these are user config files)**

No git commit needed. Provider YAMLs live in user config, not the repo.

---

### Task 7: Create work-on-issue Workflow

**Files:**
- Create: `~/Projects/stokagent/workflows/work-on-issue.yaml`
- Create symlink: `~/.config/glitch/workflows/work-on-issue.yaml` → `~/Projects/stokagent/workflows/work-on-issue.yaml`

- [ ] **Step 1: Write the workflow**

```yaml
# ~/Projects/stokagent/workflows/work-on-issue.yaml
name: work-on-issue
description: Solve a GitHub issue end-to-end — fetch, analyze, implement, and prepare PR artifacts

steps:
  - id: fetch-issue
    run: |
      gh issue view {{.param.issue}} --repo {{.param.repo}} \
        --json number,title,body,labels,comments,assignees

  - id: fetch-repo-context
    run: |
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      if [ ! -d "$REPO_PATH" ]; then
        echo "ERROR: repo not found at $REPO_PATH"
        exit 1
      fi
      echo "=== DIRECTORY STRUCTURE ==="
      find "$REPO_PATH" -maxdepth 3 -type f \
        -not -path '*/node_modules/*' \
        -not -path '*/.git/*' \
        -not -path '*/vendor/*' \
        -not -path '*/.next/*' \
        -not -path '*/dist/*' | head -200
      echo ""
      echo "=== RECENT COMMITS ==="
      git -C "$REPO_PATH" log --oneline -20

  - id: classify
    llm:
      provider: ollama
      model: qwen2.5:7b
      prompt: |
        Analyze this GitHub issue and extract structured information.

        Issue JSON:
        {{step "fetch-issue"}}

        Respond with ONLY valid JSON, no markdown fences:
        {
          "type": "code|documentation|test|refactor|infrastructure",
          "affected_files": ["path/to/file1", "path/to/file2"],
          "acceptance_criteria": ["criterion 1", "criterion 2"],
          "related_issues": [],
          "summary": "one line summary of what needs to happen"
        }

  - id: gather-context
    run: |
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      curl -sf "http://localhost:9200/glitch-code-${REPO_NAME}/_search" \
        -H 'Content-Type: application/json' \
        -d '{
          "query": {
            "multi_match": {
              "query": "issue {{.param.issue}}",
              "fields": ["content", "path", "symbols"]
            }
          },
          "size": 10
        }' 2>/dev/null \
        | jq -r '.hits.hits[]._source | "\(.path):\n\(.content)\n---"' 2>/dev/null \
        || echo "ES not available — skipping code index lookup"

  - id: create-branch
    run: |
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      ISSUE={{.param.issue}}
      cd "$REPO_PATH"
      git fetch origin
      git checkout -b "fix/${ISSUE}" origin/main 2>/dev/null \
        || git checkout "fix/${ISSUE}"
      echo "fix/${ISSUE}"

  - id: build-prompt
    llm:
      provider: ollama
      model: qwen2.5:7b
      prompt: |
        You are a prompt engineer. Build a detailed implementation prompt for an AI coding assistant that will solve this GitHub issue.

        ISSUE:
        {{step "fetch-issue"}}

        CLASSIFICATION:
        {{step "classify"}}

        REPOSITORY STRUCTURE:
        {{step "fetch-repo-context"}}

        RELATED CODE FROM INDEX:
        {{step "gather-context"}}

        Write a complete, self-contained prompt that tells the AI assistant:
        1. The full path to the repository on disk (~/Projects/<repo-name>)
        2. They are on branch fix/{{.param.issue}}
        3. Exactly what files to modify or create
        4. What changes to make with specific guidance
        5. Acceptance criteria to verify against
        6. Constraints: do NOT commit, do NOT push, do NOT open a PR, do NOT modify unrelated files

        The prompt must be detailed enough that the assistant can work autonomously.
        Output ONLY the prompt text. No wrapper, no explanation, no markdown fences.

  - id: implement
    llm:
      provider: claude
      model: claude-sonnet-4-20250514
      prompt: |
        {{step "build-prompt"}}

  - id: build-results
    run: |
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      ISSUE={{.param.issue}}
      RESULTS_DIR="$REPO_PATH/.glitch/results/$ISSUE"

      mkdir -p "$RESULTS_DIR"

      # Record branch
      echo "fix/${ISSUE}" > "$RESULTS_DIR/branch.txt"

      # Capture changes
      cd "$REPO_PATH"
      echo "# Changes for Issue #${ISSUE}" > "$RESULTS_DIR/changes.md"
      echo "" >> "$RESULTS_DIR/changes.md"
      echo "## Files Changed" >> "$RESULTS_DIR/changes.md"
      git diff origin/main --stat >> "$RESULTS_DIR/changes.md"
      echo "" >> "$RESULTS_DIR/changes.md"
      echo "## Full Diff" >> "$RESULTS_DIR/changes.md"
      git diff origin/main >> "$RESULTS_DIR/changes.md"

      # Ensure gitignore
      if ! grep -q '.glitch/results' "$REPO_PATH/.gitignore" 2>/dev/null; then
        echo '.glitch/results/' >> "$REPO_PATH/.gitignore"
      fi

      # Pass diff summary for LLM to generate PR artifacts
      git diff origin/main --stat

  - id: build-summary
    llm:
      provider: ollama
      model: qwen2.5:7b
      prompt: |
        Generate a PR summary and next steps for this completed work.

        ORIGINAL ISSUE:
        {{step "fetch-issue"}}

        CHANGES MADE:
        {{step "build-results"}}

        IMPLEMENTATION NOTES:
        {{step "implement"}}

        Output TWO sections with exact delimiters:

        ---SUMMARY_START---
        ## <PR title — short, imperative>

        <2-3 sentence description of what changed and why>

        ### Test Plan
        <bulleted list of how to verify the changes>

        Closes #<issue-number>
        ---SUMMARY_END---

        ---NEXT_START---
        <numbered list of what the developer needs to do next to get this up for review, including specific commands to run>
        ---NEXT_END---

  - id: write-artifacts
    run: |
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      ISSUE={{.param.issue}}
      RESULTS_DIR="$REPO_PATH/.glitch/results/$ISSUE"

      CONTENT=$(cat <<'INNEREOF'
      {{step "build-summary"}}
      INNEREOF
      )

      echo "$CONTENT" | sed -n '/---SUMMARY_START---/,/---SUMMARY_END---/{//!p;}' > "$RESULTS_DIR/summary.md"
      echo "$CONTENT" | sed -n '/---NEXT_START---/,/---NEXT_END---/{//!p;}' > "$RESULTS_DIR/next-steps.md"

      echo ""
      echo "=== Issue #$ISSUE — Ready for Review ==="
      echo ""
      echo "Branch: fix/$ISSUE"
      echo "Repo:   $REPO_PATH"
      echo "Results: $RESULTS_DIR/"
      echo ""
      echo "--- Next Steps ---"
      cat "$RESULTS_DIR/next-steps.md"
```

- [ ] **Step 2: Create the symlink**

```bash
ln -sf ~/Projects/stokagent/workflows/work-on-issue.yaml ~/.config/glitch/workflows/work-on-issue.yaml
```

- [ ] **Step 3: Verify the workflow loads**

Run: `cd /Users/stokes/Projects/gl1tch && go run . workflow list`
Expected: `work-on-issue` appears in the list with its description.

- [ ] **Step 4: Commit the workflow to stokagent**

```bash
cd ~/Projects/stokagent
git add workflows/work-on-issue.yaml
git commit -m "feat: add work-on-issue workflow for automated issue solving"
```

---

### Task 8: End-to-End Smoke Test

- [ ] **Step 1: Build glitch**

Run: `cd /Users/stokes/Projects/gl1tch && go build -o glitch . && go install .`
Expected: PASS

- [ ] **Step 2: Verify provider registry loads**

Run: `ls ~/.config/glitch/providers/`
Expected: claude.yaml, codex.yaml, copilot.yaml listed

- [ ] **Step 3: Verify workflow routing**

Run: `glitch workflow list | grep work-on-issue`
Expected: `work-on-issue` with description

- [ ] **Step 4: Dry run against a real issue**

Run from the ensemble repo:
```bash
cd ~/Projects/ensemble
glitch ask work on issue 1281
```

Expected behavior:
1. Router detects "work on issue 1281"
2. Resolves repo from git remote → `elastic/ensemble`
3. Runs `work-on-issue` workflow
4. Fetches issue, classifies, gathers context, creates branch, builds prompt, implements, writes results
5. Outputs summary with next steps
6. `.glitch/results/1281/` contains summary.md, changes.md, next-steps.md, branch.txt

- [ ] **Step 5: Verify results folder**

```bash
ls ~/Projects/ensemble/.glitch/results/1281/
cat ~/Projects/ensemble/.glitch/results/1281/next-steps.md
```

Expected: All four files present with content.

- [ ] **Step 6: Run all tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./...`
Expected: PASS

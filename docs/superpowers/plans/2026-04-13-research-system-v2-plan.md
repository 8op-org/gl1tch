# Research System v2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current researcher-menu research loop with a tool-use loop where the LLM iteratively calls tools against the codebase, with tiered LLM escalation and full ES telemetry.

**Architecture:** Input adapters normalize any source (GitHub issue, PR, Google Doc) into a ResearchDocument. A tool-use loop lets the LLM call shell-backed tools (grep, read, git, ES) iteratively until it has enough evidence. LLM calls escalate through tiers (local → free paid → paid). Every action is indexed to ES for Kibana dashboards.

**Tech Stack:** Go, Ollama (qwen3:8b), provider CLI tools (claude, copilot, codex, gemini), Elasticsearch 8.17, Kibana 8.17

---

## File Structure

### New Files
| File | Responsibility |
|------|---------------|
| `internal/research/document.go` | ResearchDocument type and Link type |
| `internal/research/adapter.go` | Input adapters (GitHub issue, PR, Google Doc) + URL detection |
| `internal/research/adapter_test.go` | Adapter tests |
| `internal/research/tools.go` | Tool interface, tool definitions, tool execution |
| `internal/research/tools_test.go` | Tool tests |
| `internal/research/toolloop.go` | Tool-use loop engine (replaces loop.go) |
| `internal/research/toolloop_test.go` | Tool-use loop tests |
| `internal/provider/tiers.go` | Tiered provider chain with escalation |
| `internal/provider/tiers_test.go` | Tier tests |
| `internal/provider/tokens.go` | Token counting, cost estimation, LLMResult type |
| `internal/provider/tokens_test.go` | Token/cost tests |
| `internal/esearch/telemetry.go` | Research run / tool call / LLM call indexing |
| `internal/esearch/telemetry_test.go` | Telemetry tests |
| `deploy/kibana/research-overview.ndjson` | Research Overview dashboard |
| `deploy/kibana/cost-tokens.ndjson` | Cost & Tokens dashboard |
| `deploy/kibana/escalation-funnel.ndjson` | Escalation Funnel dashboard |
| `deploy/kibana/tool-effectiveness.ndjson` | Tool Effectiveness dashboard |
| `deploy/kibana/provider-comparison.ndjson` | Provider Comparison dashboard |

### Modified Files
| File | Changes |
|------|---------|
| `internal/esearch/mappings.go` | Replace unused index mappings with 3 new indices |
| `internal/provider/provider.go` | Add `LLMResult` return from `RunOllama`, expose token fields |
| `cmd/ask.go` | Rewrite to use adapters + tool-use loop |
| `cmd/research_helpers.go` | Rewrite to build tool-use loop instead of researcher registry |
| `cmd/config.go` | Add `tiers` config field |
| `internal/research/results.go` | Update for new directory structure (evidence/, run.json) |
| `internal/research/repo.go` | Unchanged (EnsureRepo, ParseRepoFromQuestion) |
| `docker-compose.yml` | Add dashboard import init container |

### Removed Files
| File | Reason |
|------|--------|
| `internal/research/researcher.go` | Replaced by tools.go |
| `internal/research/registry.go` | No longer needed |
| `internal/research/registry_test.go` | No longer needed |
| `internal/research/git_researcher.go` | Replaced by git_log/git_diff tools |
| `internal/research/git_researcher_test.go` | Replaced |
| `internal/research/fs_researcher.go` | Replaced by grep_code/read_file/list_files tools |
| `internal/research/fs_researcher_test.go` | Replaced |
| `internal/research/es_researcher.go` | Replaced by search_es tool |
| `internal/research/yaml_researcher.go` | Removed entirely |
| `internal/research/prompts.go` | Replaced by toolloop system prompt |
| `internal/research/score.go` | Removed (confidence checks replace scoring) |
| `internal/research/score_test.go` | Removed |
| `internal/research/loop.go` | Replaced by toolloop.go |
| `internal/research/loop_test.go` | Replaced by toolloop_test.go |
| `internal/research/events.go` | Replaced by esearch/telemetry.go |
| `internal/research/feedback.go` | Integrated into toolloop output parsing |
| `internal/research/feedback_test.go` | Removed |
| `researchers/` | YAML researchers removed |

---

## Task 1: ResearchDocument and Input Adapters

**Files:**
- Create: `internal/research/document.go`
- Create: `internal/research/adapter.go`
- Create: `internal/research/adapter_test.go`

- [ ] **Step 1: Write ResearchDocument types**

```go
// internal/research/document.go
package research

// ResearchDocument is the normalized input to the research loop.
// Every input source (issue, PR, doc) gets converted to this before research begins.
type ResearchDocument struct {
	Source   string            `json:"source"`    // "github-issue", "github-pr", "google-doc", "text"
	SourceURL string          `json:"source_url"` // original URL
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Repo     string            `json:"repo"`      // "elastic/ensemble"
	RepoPath string            `json:"repo_path"` // local clone path
	Metadata map[string]string `json:"metadata,omitempty"`
	Links    []Link            `json:"links,omitempty"`
}

// Link is a reference extracted from the input source.
type Link struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}
```

- [ ] **Step 2: Write adapter detection and interface**

```go
// internal/research/adapter.go
package research

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	reGHIssue = regexp.MustCompile(`https?://github\.com/([^/]+)/([^/]+)/issues/(\d+)`)
	reGHPR    = regexp.MustCompile(`https?://github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	reGDoc    = regexp.MustCompile(`https?://docs\.google\.com/document/d/([^/]+)`)
)

// Adapt detects the input source and normalizes it into a ResearchDocument.
// Falls back to a plain text document if no URL pattern matches.
func Adapt(input string) (ResearchDocument, error) {
	input = strings.TrimSpace(input)

	if m := reGHIssue.FindStringSubmatch(input); m != nil {
		return adaptGitHubIssue(m[1], m[2], m[3], input)
	}
	if m := reGHPR.FindStringSubmatch(input); m != nil {
		return adaptGitHubPR(m[1], m[2], m[3], input)
	}
	if m := reGDoc.FindStringSubmatch(input); m != nil {
		return adaptGoogleDoc(m[1], input)
	}

	// Plain text fallback
	return ResearchDocument{
		Source: "text",
		Title:  input,
		Body:   input,
	}, nil
}
```

- [ ] **Step 3: Write GitHub issue adapter**

```go
// in internal/research/adapter.go

func adaptGitHubIssue(org, repo, number, url string) (ResearchDocument, error) {
	out, err := exec.Command("gh", "issue", "view", url,
		"--json", "number,title,body,comments,labels,assignees").Output()
	if err != nil {
		return ResearchDocument{}, fmt.Errorf("adapter: gh issue view: %w", err)
	}

	var issue struct {
		Number   int    `json:"number"`
		Title    string `json:"title"`
		Body     string `json:"body"`
		Labels   []struct{ Name string } `json:"labels"`
		Comments []struct {
			Body   string `json:"body"`
			Author struct{ Login string } `json:"author"`
		} `json:"comments"`
		Assignees []struct{ Login string } `json:"assignees"`
	}
	if err := json.Unmarshal(out, &issue); err != nil {
		return ResearchDocument{}, fmt.Errorf("adapter: parse issue: %w", err)
	}

	// Build full body with comments
	var body strings.Builder
	body.WriteString(issue.Body)
	for _, c := range issue.Comments {
		fmt.Fprintf(&body, "\n\n---\n**@%s:**\n%s", c.Author.Login, c.Body)
	}

	// Extract linked URLs from body
	links := extractLinks(issue.Body)

	meta := map[string]string{
		"number": fmt.Sprintf("%d", issue.Number),
	}
	if len(issue.Labels) > 0 {
		names := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			names[i] = l.Name
		}
		meta["labels"] = strings.Join(names, ",")
	}
	if len(issue.Assignees) > 0 {
		logins := make([]string, len(issue.Assignees))
		for i, a := range issue.Assignees {
			logins[i] = a.Login
		}
		meta["assignees"] = strings.Join(logins, ",")
	}

	// Resolve repo locally
	repoPath, _ := EnsureRepo(org, repo, "")

	return ResearchDocument{
		Source:    "github-issue",
		SourceURL: url,
		Title:     issue.Title,
		Body:      body.String(),
		Repo:      org + "/" + repo,
		RepoPath:  repoPath,
		Metadata:  meta,
		Links:     links,
	}, nil
}

// extractLinks finds GitHub URLs in text and returns them as Links.
func extractLinks(text string) []Link {
	re := regexp.MustCompile(`https?://github\.com/[^\s)\]]+`)
	matches := re.FindAllString(text, -1)
	var links []Link
	for _, u := range matches {
		links = append(links, Link{URL: u, Label: u})
	}
	return links
}
```

- [ ] **Step 4: Write GitHub PR adapter**

```go
// in internal/research/adapter.go

func adaptGitHubPR(org, repo, number, url string) (ResearchDocument, error) {
	out, err := exec.Command("gh", "pr", "view", url,
		"--json", "number,title,body,comments,files,additions,deletions,reviews,state").Output()
	if err != nil {
		return ResearchDocument{}, fmt.Errorf("adapter: gh pr view: %w", err)
	}

	var pr struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		State     string `json:"state"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Files     []struct{ Path string } `json:"files"`
		Comments  []struct {
			Body   string `json:"body"`
			Author struct{ Login string } `json:"author"`
		} `json:"comments"`
		Reviews []struct {
			Body   string `json:"body"`
			State  string `json:"state"`
			Author struct{ Login string } `json:"author"`
		} `json:"reviews"`
	}
	if err := json.Unmarshal(out, &pr); err != nil {
		return ResearchDocument{}, fmt.Errorf("adapter: parse PR: %w", err)
	}

	var body strings.Builder
	body.WriteString(pr.Body)
	for _, c := range pr.Comments {
		fmt.Fprintf(&body, "\n\n---\n**@%s:**\n%s", c.Author.Login, c.Body)
	}
	for _, r := range pr.Reviews {
		if r.Body != "" {
			fmt.Fprintf(&body, "\n\n---\n**@%s (%s):**\n%s", r.Author.Login, r.State, r.Body)
		}
	}

	meta := map[string]string{
		"number":    fmt.Sprintf("%d", pr.Number),
		"state":     pr.State,
		"additions": fmt.Sprintf("%d", pr.Additions),
		"deletions": fmt.Sprintf("%d", pr.Deletions),
	}
	files := make([]string, len(pr.Files))
	for i, f := range pr.Files {
		files[i] = f.Path
	}
	meta["files"] = strings.Join(files, ",")

	links := extractLinks(pr.Body)
	repoPath, _ := EnsureRepo(org, repo, "")

	return ResearchDocument{
		Source:    "github-pr",
		SourceURL: url,
		Title:     pr.Title,
		Body:      body.String(),
		Repo:      org + "/" + repo,
		RepoPath:  repoPath,
		Metadata:  meta,
		Links:     links,
	}, nil
}
```

- [ ] **Step 5: Write Google Doc adapter**

```go
// in internal/research/adapter.go

func adaptGoogleDoc(docID, url string) (ResearchDocument, error) {
	out, err := exec.Command("gws", "docs", "get", docID).Output()
	if err != nil {
		return ResearchDocument{}, fmt.Errorf("adapter: gws docs get: %w", err)
	}

	content := strings.TrimSpace(string(out))
	// Use first line as title
	title := content
	if idx := strings.IndexByte(content, '\n'); idx > 0 {
		title = content[:idx]
	}

	return ResearchDocument{
		Source:    "google-doc",
		SourceURL: url,
		Title:     title,
		Body:      content,
		Metadata:  map[string]string{"doc_id": docID},
	}, nil
}
```

- [ ] **Step 6: Write adapter tests**

```go
// internal/research/adapter_test.go
package research

import "testing"

func TestAdaptDetectsGitHubIssue(t *testing.T) {
	// Test URL detection only — adapter calls gh which needs network
	m := reGHIssue.FindStringSubmatch("https://github.com/elastic/ensemble/issues/872")
	if m == nil {
		t.Fatal("expected match")
	}
	if m[1] != "elastic" || m[2] != "ensemble" || m[3] != "872" {
		t.Errorf("got org=%s repo=%s num=%s", m[1], m[2], m[3])
	}
}

func TestAdaptDetectsGitHubPR(t *testing.T) {
	m := reGHPR.FindStringSubmatch("https://github.com/elastic/ensemble/pull/747")
	if m == nil {
		t.Fatal("expected match")
	}
	if m[1] != "elastic" || m[2] != "ensemble" || m[3] != "747" {
		t.Errorf("got org=%s repo=%s num=%s", m[1], m[2], m[3])
	}
}

func TestAdaptDetectsGoogleDoc(t *testing.T) {
	m := reGDoc.FindStringSubmatch("https://docs.google.com/document/d/1aBcDeFgHiJk/edit")
	if m == nil {
		t.Fatal("expected match")
	}
	if m[1] != "1aBcDeFgHiJk" {
		t.Errorf("got doc_id=%s", m[1])
	}
}

func TestAdaptFallsBackToText(t *testing.T) {
	doc, err := Adapt("what is the meaning of life")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Source != "text" {
		t.Errorf("expected source=text, got %s", doc.Source)
	}
	if doc.Body != "what is the meaning of life" {
		t.Errorf("unexpected body: %s", doc.Body)
	}
}

func TestExtractLinks(t *testing.T) {
	text := "See https://github.com/elastic/ensemble/pull/747 for details"
	links := extractLinks(text)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "https://github.com/elastic/ensemble/pull/747" {
		t.Errorf("unexpected URL: %s", links[0].URL)
	}
}
```

- [ ] **Step 7: Run tests**

Run: `go test ./internal/research/ -run TestAdapt -v`
Expected: all 5 tests PASS

- [ ] **Step 8: Commit**

```bash
git add internal/research/document.go internal/research/adapter.go internal/research/adapter_test.go
git commit -m "feat: add input adapters — normalize GitHub issues, PRs, and Google Docs into ResearchDocument"
```

---

## Task 2: Research Tools

**Files:**
- Create: `internal/research/tools.go`
- Create: `internal/research/tools_test.go`

- [ ] **Step 1: Write Tool interface and registry**

```go
// internal/research/tools.go
package research

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/8op-org/gl1tch/internal/esearch"
)

// Tool is a function the LLM can call during research.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Params      string `json:"params"` // human-readable param description
}

// ToolResult is what a tool returns.
type ToolResult struct {
	Tool   string `json:"tool"`
	Output string `json:"output"`
	Err    string `json:"error,omitempty"`
}

// ToolSet holds all available tools and the context they operate in.
type ToolSet struct {
	repoPath string
	es       *esearch.Client
}

// NewToolSet creates a ToolSet bound to a repo path and optional ES client.
func NewToolSet(repoPath string, es *esearch.Client) *ToolSet {
	return &ToolSet{repoPath: repoPath, es: es}
}

// Definitions returns the tool descriptions for the LLM system prompt.
func (ts *ToolSet) Definitions() []Tool {
	return []Tool{
		{Name: "grep_code", Description: "search file contents for a regex pattern", Params: "pattern (required), path (optional, relative), glob (optional, e.g. '*.go')"},
		{Name: "read_file", Description: "read a file's contents", Params: "path (required, relative), start_line (optional), end_line (optional)"},
		{Name: "git_log", Description: "search git commit history", Params: "query (optional, grep pattern), path (optional), limit (optional, default 20)"},
		{Name: "git_diff", Description: "show diff between refs or working tree", Params: "ref1 (optional, default HEAD~10), ref2 (optional), path (optional)"},
		{Name: "search_es", Description: "search Elasticsearch indices for activity, code, or events", Params: "query (required, search text), index (optional, e.g. 'glitch-events')"},
		{Name: "list_files", Description: "list directory tree", Params: "path (optional, relative, default '.'), depth (optional, default 3)"},
		{Name: "fetch_issue", Description: "fetch a GitHub issue", Params: "repo (required, e.g. 'elastic/ensemble'), number (required)"},
		{Name: "fetch_pr", Description: "fetch a GitHub pull request with diff stats", Params: "repo (required), number (required)"},
	}
}
```

- [ ] **Step 2: Write Execute dispatcher**

```go
// in internal/research/tools.go

// Execute runs a tool by name with the given params.
func (ts *ToolSet) Execute(ctx context.Context, name string, params map[string]string) ToolResult {
	switch name {
	case "grep_code":
		return ts.grepCode(params)
	case "read_file":
		return ts.readFile(params)
	case "git_log":
		return ts.gitLog(params)
	case "git_diff":
		return ts.gitDiff(params)
	case "search_es":
		return ts.searchES(ctx, params)
	case "list_files":
		return ts.listFiles(params)
	case "fetch_issue":
		return ts.fetchIssue(params)
	case "fetch_pr":
		return ts.fetchPR(params)
	default:
		return ToolResult{Tool: name, Err: fmt.Sprintf("unknown tool: %s", name)}
	}
}

// ValidTool returns true if the name is a known tool.
func (ts *ToolSet) ValidTool(name string) bool {
	for _, t := range ts.Definitions() {
		if t.Name == name {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: Write tool implementations**

```go
// in internal/research/tools.go

func (ts *ToolSet) grepCode(p map[string]string) ToolResult {
	pattern := p["pattern"]
	if pattern == "" {
		return ToolResult{Tool: "grep_code", Err: "pattern is required"}
	}
	args := []string{"-rn", "--max-count=5", pattern}
	if glob := p["glob"]; glob != "" {
		args = append(args, "--include="+glob)
	}
	searchPath := ts.repoPath
	if rel := p["path"]; rel != "" {
		searchPath = searchPath + "/" + rel
	}
	args = append(args, searchPath)
	out, _ := exec.Command("grep", args...).Output()
	result := strings.TrimSpace(string(out))
	if result == "" {
		return ToolResult{Tool: "grep_code", Output: "no matches found"}
	}
	return ToolResult{Tool: "grep_code", Output: truncateOutput(result, 8000)}
}

func (ts *ToolSet) readFile(p map[string]string) ToolResult {
	path := p["path"]
	if path == "" {
		return ToolResult{Tool: "read_file", Err: "path is required"}
	}
	full := ts.repoPath + "/" + path
	data, err := exec.Command("head", "-n", "200", full).Output()
	if err != nil {
		return ToolResult{Tool: "read_file", Err: fmt.Sprintf("read %s: %v", path, err)}
	}
	return ToolResult{Tool: "read_file", Output: string(data)}
}

func (ts *ToolSet) gitLog(p map[string]string) ToolResult {
	args := []string{"-C", ts.repoPath, "log", "--oneline"}
	if q := p["query"]; q != "" {
		args = append(args, "--all", "--grep="+q)
	}
	limit := p["limit"]
	if limit == "" {
		limit = "20"
	}
	args = append(args, "-"+limit)
	if path := p["path"]; path != "" {
		args = append(args, "--", path)
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ToolResult{Tool: "git_log", Err: fmt.Sprintf("git log: %v", err)}
	}
	return ToolResult{Tool: "git_log", Output: strings.TrimSpace(string(out))}
}

func (ts *ToolSet) gitDiff(p map[string]string) ToolResult {
	args := []string{"-C", ts.repoPath, "diff"}
	ref1 := p["ref1"]
	if ref1 == "" {
		ref1 = "HEAD~10"
	}
	args = append(args, ref1)
	if ref2 := p["ref2"]; ref2 != "" {
		args = append(args, ref2)
	}
	args = append(args, "--stat")
	if path := p["path"]; path != "" {
		args = append(args, "--", path)
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ToolResult{Tool: "git_diff", Err: fmt.Sprintf("git diff: %v", err)}
	}
	return ToolResult{Tool: "git_diff", Output: strings.TrimSpace(string(out))}
}

func (ts *ToolSet) searchES(ctx context.Context, p map[string]string) ToolResult {
	if ts.es == nil {
		return ToolResult{Tool: "search_es", Err: "elasticsearch not available"}
	}
	q := p["query"]
	if q == "" {
		return ToolResult{Tool: "search_es", Err: "query is required"}
	}
	indices := []string{"glitch-events", "glitch-code-*"}
	if idx := p["index"]; idx != "" {
		indices = []string{idx}
	}
	body := fmt.Sprintf(`{"query":{"multi_match":{"query":%q,"fields":["content","message","body","path","symbols"]}},"size":10}`, q)
	resp, err := ts.es.Search(ctx, indices, json.RawMessage(body))
	if err != nil {
		return ToolResult{Tool: "search_es", Err: fmt.Sprintf("search: %v", err)}
	}
	var results []string
	for _, hit := range resp.Results {
		results = append(results, string(hit.Source))
	}
	if len(results) == 0 {
		return ToolResult{Tool: "search_es", Output: "no results found"}
	}
	return ToolResult{Tool: "search_es", Output: truncateOutput(strings.Join(results, "\n---\n"), 8000)}
}

func (ts *ToolSet) listFiles(p map[string]string) ToolResult {
	path := ts.repoPath
	if rel := p["path"]; rel != "" {
		path = path + "/" + rel
	}
	depth := p["depth"]
	if depth == "" {
		depth = "3"
	}
	out, err := exec.Command("find", path, "-maxdepth", depth, "-type", "f",
		"-not", "-path", "*/.git/*",
		"-not", "-path", "*/node_modules/*",
		"-not", "-path", "*/vendor/*",
	).Output()
	if err != nil {
		return ToolResult{Tool: "list_files", Err: fmt.Sprintf("list: %v", err)}
	}
	result := strings.ReplaceAll(strings.TrimSpace(string(out)), path+"/", "")
	return ToolResult{Tool: "list_files", Output: truncateOutput(result, 8000)}
}

func (ts *ToolSet) fetchIssue(p map[string]string) ToolResult {
	repo := p["repo"]
	num := p["number"]
	if repo == "" || num == "" {
		return ToolResult{Tool: "fetch_issue", Err: "repo and number are required"}
	}
	out, err := exec.Command("gh", "issue", "view", num, "--repo", repo,
		"--json", "number,title,body,comments,labels").Output()
	if err != nil {
		return ToolResult{Tool: "fetch_issue", Err: fmt.Sprintf("gh issue view: %v", err)}
	}
	return ToolResult{Tool: "fetch_issue", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) fetchPR(p map[string]string) ToolResult {
	repo := p["repo"]
	num := p["number"]
	if repo == "" || num == "" {
		return ToolResult{Tool: "fetch_pr", Err: "repo and number are required"}
	}
	out, err := exec.Command("gh", "pr", "view", num, "--repo", repo,
		"--json", "number,title,body,files,additions,deletions,state,reviews").Output()
	if err != nil {
		return ToolResult{Tool: "fetch_pr", Err: fmt.Sprintf("gh pr view: %v", err)}
	}
	return ToolResult{Tool: "fetch_pr", Output: truncateOutput(string(out), 8000)}
}

func truncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... (truncated)"
}
```

- [ ] **Step 4: Write tool tests**

```go
// internal/research/tools_test.go
package research

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolSetValidTool(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	if !ts.ValidTool("grep_code") {
		t.Error("grep_code should be valid")
	}
	if ts.ValidTool("hack_mainframe") {
		t.Error("hack_mainframe should not be valid")
	}
}

func TestToolSetDefinitions(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	defs := ts.Definitions()
	if len(defs) != 8 {
		t.Errorf("expected 8 tool definitions, got %d", len(defs))
	}
	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}
	for _, want := range []string{"grep_code", "read_file", "git_log", "git_diff", "search_es", "list_files", "fetch_issue", "fetch_pr"} {
		if !names[want] {
			t.Errorf("missing tool definition: %s", want)
		}
	}
}

func TestGrepCodeRequiresPattern(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	result := ts.Execute(context.Background(), "grep_code", map[string]string{})
	if result.Err == "" {
		t.Error("expected error for missing pattern")
	}
}

func TestReadFileWorks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world\n"), 0644)

	ts := NewToolSet(dir, nil)
	result := ts.Execute(context.Background(), "read_file", map[string]string{"path": "hello.txt"})
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Output != "hello world\n" {
		t.Errorf("unexpected output: %q", result.Output)
	}
}

func TestListFilesWorks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "b.go"), []byte("package b"), 0644)

	ts := NewToolSet(dir, nil)
	result := ts.Execute(context.Background(), "list_files", map[string]string{"depth": "2"})
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if !strings.Contains(result.Output, "a.go") || !strings.Contains(result.Output, "sub/b.go") {
		t.Errorf("expected both files in output: %s", result.Output)
	}
}

func TestUnknownToolReturnsError(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	result := ts.Execute(context.Background(), "hack_mainframe", nil)
	if result.Err == "" {
		t.Error("expected error for unknown tool")
	}
}

func TestSearchESWithoutClient(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	result := ts.Execute(context.Background(), "search_es", map[string]string{"query": "test"})
	if result.Err == "" || result.Err != "elasticsearch not available" {
		t.Errorf("expected ES unavailable error, got: %s", result.Err)
	}
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/research/ -run TestToolSet -v && go test ./internal/research/ -run TestGrep -v && go test ./internal/research/ -run TestRead -v && go test ./internal/research/ -run TestList -v && go test ./internal/research/ -run TestUnknown -v && go test ./internal/research/ -run TestSearch -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/research/tools.go internal/research/tools_test.go
git commit -m "feat: add research tools — grep, read, git, ES, list, fetch for tool-use loop"
```

---

## Task 3: LLM Token Tracking and Tiered Escalation

**Files:**
- Create: `internal/provider/tokens.go`
- Create: `internal/provider/tokens_test.go`
- Create: `internal/provider/tiers.go`
- Create: `internal/provider/tiers_test.go`
- Modify: `internal/provider/provider.go` (RunOllama return type)
- Modify: `cmd/config.go` (add tiers config)

- [ ] **Step 1: Write LLMResult type and token estimation**

```go
// internal/provider/tokens.go
package provider

import "time"

// LLMResult wraps an LLM response with metadata for telemetry.
type LLMResult struct {
	Provider  string        `json:"provider"`
	Model     string        `json:"model"`
	Response  string        `json:"response"`
	TokensIn  int           `json:"tokens_in"`
	TokensOut int           `json:"tokens_out"`
	Latency   time.Duration `json:"latency"`
	CostUSD   float64       `json:"cost_usd"`
}

// EstimateTokens returns a rough token count (chars / 4).
// Used when the provider doesn't return token counts.
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// Known pricing per 1M tokens (input, output) in USD.
var pricing = map[string][2]float64{
	"claude":  {3.00, 15.00},
	"copilot": {0.00, 0.00}, // included in subscription
	"codex":   {0.00, 0.00}, // free tier
	"gemini":  {0.00, 0.00}, // vertex free tier
	"ollama":  {0.00, 0.00}, // local
}

// EstimateCost returns estimated cost in USD given provider and token counts.
func EstimateCost(providerName string, tokensIn, tokensOut int) float64 {
	p, ok := pricing[providerName]
	if !ok {
		return 0
	}
	return (float64(tokensIn) * p[0] / 1_000_000) + (float64(tokensOut) * p[1] / 1_000_000)
}
```

- [ ] **Step 2: Write token tests**

```go
// internal/provider/tokens_test.go
package provider

import "testing"

func TestEstimateTokens(t *testing.T) {
	// 100 chars → ~25 tokens
	text := "aaaaaaaaaa" // 10 chars repeated
	got := EstimateTokens(text)
	if got < 2 || got > 4 {
		t.Errorf("EstimateTokens(%d chars) = %d, want ~3", len(text), got)
	}
}

func TestEstimateCostClaude(t *testing.T) {
	cost := EstimateCost("claude", 1000, 500)
	if cost < 0.001 || cost > 0.01 {
		t.Errorf("EstimateCost(claude, 1000, 500) = %f, want ~0.003-0.01", cost)
	}
}

func TestEstimateCostFreeProvider(t *testing.T) {
	cost := EstimateCost("ollama", 10000, 5000)
	if cost != 0 {
		t.Errorf("EstimateCost(ollama) = %f, want 0", cost)
	}
}

func TestEstimateCostUnknownProvider(t *testing.T) {
	cost := EstimateCost("unknown", 1000, 500)
	if cost != 0 {
		t.Errorf("EstimateCost(unknown) = %f, want 0", cost)
	}
}
```

- [ ] **Step 3: Run token tests**

Run: `go test ./internal/provider/ -run TestEstimate -v`
Expected: all PASS

- [ ] **Step 4: Update RunOllama to return LLMResult**

Modify `internal/provider/provider.go` — add `RunOllamaWithResult` alongside existing `RunOllama` (don't break existing callers):

```go
// Add to internal/provider/provider.go after RunOllama

// RunOllamaWithResult is like RunOllama but returns full LLMResult with token counts.
func RunOllamaWithResult(model, prompt string) (LLMResult, error) {
	start := time.Now()
	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return LLMResult{}, fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResult{}, fmt.Errorf("ollama: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return LLMResult{}, fmt.Errorf("ollama: %s\n%s", resp.Status, data)
	}

	var raw struct {
		Response           string `json:"response"`
		PromptEvalCount    int    `json:"prompt_eval_count"`
		EvalCount          int    `json:"eval_count"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return LLMResult{}, fmt.Errorf("ollama: parse: %w", err)
	}

	tokIn := raw.PromptEvalCount
	if tokIn == 0 {
		tokIn = EstimateTokens(prompt)
	}
	tokOut := raw.EvalCount
	if tokOut == 0 {
		tokOut = EstimateTokens(raw.Response)
	}

	return LLMResult{
		Provider:  "ollama",
		Model:     model,
		Response:  strings.TrimSpace(raw.Response),
		TokensIn:  tokIn,
		TokensOut: tokOut,
		Latency:   time.Since(start),
		CostUSD:   0,
	}, nil
}
```

- [ ] **Step 5: Write tiered provider chain**

```go
// internal/provider/tiers.go
package provider

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TierConfig defines one escalation tier.
type TierConfig struct {
	Providers []string `yaml:"providers"`
	Model     string   `yaml:"model,omitempty"`
}

// TieredRunner tries providers in tier order, escalating on failure.
type TieredRunner struct {
	tiers []TierConfig
	reg   *ProviderRegistry
}

// NewTieredRunner creates a runner with the given tier configuration.
func NewTieredRunner(tiers []TierConfig, reg *ProviderRegistry) *TieredRunner {
	return &TieredRunner{tiers: tiers, reg: reg}
}

// EscalationReason describes why a tier was skipped.
type EscalationReason string

const (
	ReasonMalformed      EscalationReason = "malformed_output"
	ReasonEmpty          EscalationReason = "empty_response"
	ReasonHallucinated   EscalationReason = "hallucinated_tool"
	ReasonProviderError  EscalationReason = "provider_error"
)

// RunResult is a single LLM call result with escalation metadata.
type RunResult struct {
	LLMResult
	Tier             int              `json:"tier"`
	Escalated        bool             `json:"escalated"`
	EscalationReason EscalationReason `json:"escalation_reason,omitempty"`
}

// Run sends a prompt through the tier chain. validate checks the response
// and returns an EscalationReason if the response is unacceptable.
// If validate is nil, any non-error response is accepted.
func (tr *TieredRunner) Run(ctx context.Context, prompt string, validate func(string) EscalationReason) (RunResult, error) {
	var lastErr error

	for tierIdx, tier := range tr.tiers {
		for _, provName := range tier.Providers {
			result, err := tr.callProvider(provName, tier.Model, prompt)
			if err != nil {
				lastErr = err
				continue
			}

			result.Tier = tierIdx

			// If no validator, accept any response
			if validate == nil {
				return result, nil
			}

			// Check confidence
			if reason := validate(result.Response); reason != "" {
				result.Escalated = true
				result.EscalationReason = reason
				lastErr = fmt.Errorf("tier %d provider %s: %s", tierIdx, provName, reason)
				break // escalate to next tier
			}

			return result, nil
		}
	}

	return RunResult{}, fmt.Errorf("all tiers exhausted: %w", lastErr)
}

func (tr *TieredRunner) callProvider(name, model, prompt string) (RunResult, error) {
	start := time.Now()

	if name == "ollama" {
		m := model
		if m == "" {
			m = "qwen3:8b"
		}
		lr, err := RunOllamaWithResult(m, prompt)
		if err != nil {
			return RunResult{}, err
		}
		return RunResult{LLMResult: lr}, nil
	}

	// External provider via registry
	resp, err := tr.reg.RunProvider(name, prompt)
	if err != nil {
		return RunResult{}, fmt.Errorf("provider %s: %w", name, err)
	}

	tokIn := EstimateTokens(prompt)
	tokOut := EstimateTokens(resp)

	return RunResult{
		LLMResult: LLMResult{
			Provider:  name,
			Model:     model,
			Response:  strings.TrimSpace(resp),
			TokensIn:  tokIn,
			TokensOut: tokOut,
			Latency:   time.Since(start),
			CostUSD:   EstimateCost(name, tokIn, tokOut),
		},
	}, nil
}

// DefaultTiers returns the default tier configuration.
func DefaultTiers() []TierConfig {
	return []TierConfig{
		{Providers: []string{"ollama"}, Model: "qwen3:8b"},
		{Providers: []string{"codex", "gemini"}},
		{Providers: []string{"copilot", "claude"}},
	}
}
```

- [ ] **Step 6: Write tier tests**

```go
// internal/provider/tiers_test.go
package provider

import (
	"context"
	"testing"
)

func TestDefaultTiersHasThreeLevels(t *testing.T) {
	tiers := DefaultTiers()
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	if tiers[0].Providers[0] != "ollama" {
		t.Errorf("tier 0 should start with ollama, got %s", tiers[0].Providers[0])
	}
	if tiers[0].Model != "qwen3:8b" {
		t.Errorf("tier 0 model should be qwen3:8b, got %s", tiers[0].Model)
	}
}

func TestTieredRunnerAllExhausted(t *testing.T) {
	// No providers configured, all should fail
	reg, _ := LoadProviders(t.TempDir())
	runner := NewTieredRunner([]TierConfig{
		{Providers: []string{"nonexistent"}},
	}, reg)

	_, err := runner.Run(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error when all tiers exhausted")
	}
}

func TestEscalationReasonConstants(t *testing.T) {
	reasons := []EscalationReason{ReasonMalformed, ReasonEmpty, ReasonHallucinated, ReasonProviderError}
	for _, r := range reasons {
		if r == "" {
			t.Error("escalation reason should not be empty")
		}
	}
}
```

- [ ] **Step 7: Add tiers to config**

Modify `cmd/config.go` — add Tiers field to Config struct and support loading it:

```go
// In Config struct, add:
type Config struct {
	DefaultModel    string              `yaml:"default_model"`
	DefaultProvider string              `yaml:"default_provider"`
	Tiers           []provider.TierConfig `yaml:"tiers,omitempty"`
}

// In loadConfig(), update the default:
func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return &Config{
			DefaultModel:    "qwen3:8b",
			DefaultProvider: "ollama",
			Tiers:           provider.DefaultTiers(),
		}, nil
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Tiers) == 0 {
		cfg.Tiers = provider.DefaultTiers()
	}
	return &cfg, nil
}
```

Also update the import to include `"github.com/8op-org/gl1tch/internal/provider"`.

- [ ] **Step 8: Run all tests**

Run: `go test ./internal/provider/ -v && go build ./...`
Expected: all PASS, build succeeds

- [ ] **Step 9: Commit**

```bash
git add internal/provider/tokens.go internal/provider/tokens_test.go internal/provider/tiers.go internal/provider/tiers_test.go internal/provider/provider.go cmd/config.go
git commit -m "feat: add tiered LLM escalation — local → free paid → paid with token tracking"
```

---

## Task 4: ES Telemetry

**Files:**
- Create: `internal/esearch/telemetry.go`
- Create: `internal/esearch/telemetry_test.go`
- Modify: `internal/esearch/mappings.go`

- [ ] **Step 1: Replace unused index mappings**

Replace the contents of `internal/esearch/mappings.go` with:

```go
package esearch

const (
	IndexEvents       = "glitch-events"
	IndexResearchRuns = "glitch-research-runs"
	IndexToolCalls    = "glitch-tool-calls"
	IndexLLMCalls     = "glitch-llm-calls"
)

const EventsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "type":      { "type": "keyword" },
      "source":    { "type": "keyword" },
      "repo":      { "type": "keyword" },
      "author":    { "type": "keyword" },
      "message":   { "type": "text" },
      "body":      { "type": "text" },
      "metadata":  { "type": "object", "enabled": false },
      "timestamp": { "type": "date" }
    }
  }
}`

const ResearchRunsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":           { "type": "keyword" },
      "input_source":     { "type": "keyword" },
      "source_url":       { "type": "keyword" },
      "goal":             { "type": "keyword" },
      "total_tool_calls": { "type": "integer" },
      "total_llm_calls":  { "type": "integer" },
      "total_tokens_in":  { "type": "long" },
      "total_tokens_out": { "type": "long" },
      "total_cost_usd":   { "type": "float" },
      "duration_ms":      { "type": "long" },
      "final_tier_used":  { "type": "integer" },
      "escalation_count": { "type": "integer" },
      "confidence_pass":  { "type": "boolean" },
      "timestamp":        { "type": "date" }
    }
  }
}`

const ToolCallsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":            { "type": "keyword" },
      "tool_name":         { "type": "keyword" },
      "input_summary":     { "type": "text" },
      "output_size_bytes": { "type": "integer" },
      "latency_ms":        { "type": "long" },
      "success":           { "type": "boolean" },
      "timestamp":         { "type": "date" }
    }
  }
}`

const LLMCallsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":             { "type": "keyword" },
      "step":               { "type": "keyword" },
      "tier":               { "type": "integer" },
      "provider":           { "type": "keyword" },
      "model":              { "type": "keyword" },
      "tokens_in":          { "type": "long" },
      "tokens_out":         { "type": "long" },
      "cost_usd":           { "type": "float" },
      "latency_ms":         { "type": "long" },
      "escalated":          { "type": "boolean" },
      "escalation_reason":  { "type": "keyword" },
      "timestamp":          { "type": "date" }
    }
  }
}`

// AllIndices returns a map of index name → mapping JSON for all managed indices.
func AllIndices() map[string]string {
	return map[string]string{
		IndexEvents:       EventsMapping,
		IndexResearchRuns: ResearchRunsMapping,
		IndexToolCalls:    ToolCallsMapping,
		IndexLLMCalls:     LLMCallsMapping,
	}
}
```

- [ ] **Step 2: Write telemetry indexer**

```go
// internal/esearch/telemetry.go
package esearch

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// Telemetry indexes research run data into Elasticsearch.
type Telemetry struct {
	client *Client
}

// NewTelemetry creates a Telemetry instance. Returns nil if client is nil.
func NewTelemetry(client *Client) *Telemetry {
	if client == nil {
		return nil
	}
	return &Telemetry{client: client}
}

// EnsureIndices creates all telemetry indices if they don't exist.
func (t *Telemetry) EnsureIndices(ctx context.Context) error {
	if t == nil {
		return nil
	}
	for name, mapping := range AllIndices() {
		if err := t.client.EnsureIndex(ctx, name, mapping); err != nil {
			return fmt.Errorf("ensure %s: %w", name, err)
		}
	}
	return nil
}

// IndexResearchRun indexes a completed research run.
func (t *Telemetry) IndexResearchRun(ctx context.Context, doc ResearchRunDoc) error {
	if t == nil {
		return nil
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return t.client.BulkIndex(ctx, IndexResearchRuns, []BulkDoc{
		{ID: doc.RunID, Body: body},
	})
}

// IndexToolCall indexes a single tool invocation.
func (t *Telemetry) IndexToolCall(ctx context.Context, doc ToolCallDoc) error {
	if t == nil {
		return nil
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s-%s-%d", doc.RunID, doc.ToolName, doc.Timestamp.UnixNano())
	return t.client.BulkIndex(ctx, IndexToolCalls, []BulkDoc{
		{ID: id, Body: body},
	})
}

// IndexLLMCall indexes a single LLM invocation.
func (t *Telemetry) IndexLLMCall(ctx context.Context, doc LLMCallDoc) error {
	if t == nil {
		return nil
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s-%s-%d", doc.RunID, doc.Provider, doc.Timestamp.UnixNano())
	return t.client.BulkIndex(ctx, IndexLLMCalls, []BulkDoc{
		{ID: id, Body: body},
	})
}

// ResearchRunDoc is the ES document for a research run.
type ResearchRunDoc struct {
	RunID           string    `json:"run_id"`
	InputSource     string    `json:"input_source"`
	SourceURL       string    `json:"source_url"`
	Goal            string    `json:"goal"`
	TotalToolCalls  int       `json:"total_tool_calls"`
	TotalLLMCalls   int       `json:"total_llm_calls"`
	TotalTokensIn   int       `json:"total_tokens_in"`
	TotalTokensOut  int       `json:"total_tokens_out"`
	TotalCostUSD    float64   `json:"total_cost_usd"`
	DurationMS      int64     `json:"duration_ms"`
	FinalTierUsed   int       `json:"final_tier_used"`
	EscalationCount int       `json:"escalation_count"`
	ConfidencePass  bool      `json:"confidence_pass"`
	Timestamp       time.Time `json:"timestamp"`
}

// ToolCallDoc is the ES document for a tool call.
type ToolCallDoc struct {
	RunID           string    `json:"run_id"`
	ToolName        string    `json:"tool_name"`
	InputSummary    string    `json:"input_summary"`
	OutputSizeBytes int       `json:"output_size_bytes"`
	LatencyMS       int64     `json:"latency_ms"`
	Success         bool      `json:"success"`
	Timestamp       time.Time `json:"timestamp"`
}

// LLMCallDoc is the ES document for an LLM call.
type LLMCallDoc struct {
	RunID            string    `json:"run_id"`
	Step             string    `json:"step"`
	Tier             int       `json:"tier"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	TokensIn         int       `json:"tokens_in"`
	TokensOut        int       `json:"tokens_out"`
	CostUSD          float64   `json:"cost_usd"`
	LatencyMS        int64     `json:"latency_ms"`
	Escalated        bool      `json:"escalated"`
	EscalationReason string    `json:"escalation_reason,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// NewRunID generates a unique run ID.
func NewRunID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("run-%d-%x", time.Now().UnixNano(), b)
}
```

- [ ] **Step 3: Write telemetry tests**

```go
// internal/esearch/telemetry_test.go
package esearch

import (
	"testing"
	"time"
)

func TestNewTelemetryNilClient(t *testing.T) {
	tel := NewTelemetry(nil)
	if tel != nil {
		t.Error("expected nil telemetry for nil client")
	}
}

func TestNilTelemetryMethodsSafe(t *testing.T) {
	var tel *Telemetry
	// All methods should be nil-safe
	if err := tel.EnsureIndices(nil); err != nil {
		t.Errorf("EnsureIndices on nil: %v", err)
	}
	if err := tel.IndexResearchRun(nil, ResearchRunDoc{}); err != nil {
		t.Errorf("IndexResearchRun on nil: %v", err)
	}
	if err := tel.IndexToolCall(nil, ToolCallDoc{}); err != nil {
		t.Errorf("IndexToolCall on nil: %v", err)
	}
	if err := tel.IndexLLMCall(nil, LLMCallDoc{}); err != nil {
		t.Errorf("IndexLLMCall on nil: %v", err)
	}
}

func TestNewRunIDUnique(t *testing.T) {
	a := NewRunID()
	b := NewRunID()
	if a == b {
		t.Errorf("expected unique run IDs, got %s twice", a)
	}
	if len(a) < 10 {
		t.Errorf("run ID too short: %s", a)
	}
}

func TestResearchRunDocJSON(t *testing.T) {
	doc := ResearchRunDoc{
		RunID:       "run-test",
		InputSource: "github-issue",
		Goal:        "summarize",
		Timestamp:   time.Now(),
	}
	if doc.RunID != "run-test" {
		t.Error("unexpected RunID")
	}
}
```

- [ ] **Step 4: Fix any references to old index constants**

Search for `IndexSummaries`, `IndexPipelines`, `IndexInsights` in the codebase and update them. Key files:
- `internal/research/es_researcher.go` — will be deleted in Task 6, but must compile until then
- `internal/observer/query.go` — update to use new index names or remove stale references

Run: `grep -rn "IndexSummaries\|IndexPipelines\|IndexInsights\|SummariesMapping\|PipelinesMapping\|InsightsMapping" internal/ cmd/`

Update any references found to use the new constants, or remove the references if the code using them is being deleted.

- [ ] **Step 5: Run tests and build**

Run: `go test ./internal/esearch/ -v && go build ./...`
Expected: all PASS, build succeeds

- [ ] **Step 6: Commit**

```bash
git add internal/esearch/mappings.go internal/esearch/telemetry.go internal/esearch/telemetry_test.go
git commit -m "feat: add ES telemetry — research runs, tool calls, and LLM calls indexed"
```

---

## Task 5: Tool-Use Loop Engine

**Files:**
- Create: `internal/research/toolloop.go`
- Create: `internal/research/toolloop_test.go`

- [ ] **Step 1: Write the system prompt builder**

```go
// internal/research/toolloop.go
package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
)

// Goal represents what the research loop should produce.
type Goal string

const (
	GoalSummarize  Goal = "summarize"
	GoalImplement  Goal = "implement"
)

// ToolLoop is the v2 research engine using iterative tool-use.
type ToolLoop struct {
	tools     *ToolSet
	runner    *provider.TieredRunner
	telemetry *esearch.Telemetry
}

// NewToolLoop creates a ToolLoop.
func NewToolLoop(tools *ToolSet, runner *provider.TieredRunner, tel *esearch.Telemetry) *ToolLoop {
	return &ToolLoop{tools: tools, runner: runner, telemetry: tel}
}

func buildSystemPrompt(doc ResearchDocument, goal Goal, tools []Tool) string {
	var b strings.Builder
	b.WriteString("You are a research assistant. You have tools to investigate a codebase.\n\n")

	b.WriteString("## Input\n")
	fmt.Fprintf(&b, "Source: %s\n", doc.Source)
	fmt.Fprintf(&b, "Title: %s\n", doc.Title)
	if doc.Repo != "" {
		fmt.Fprintf(&b, "Repo: %s\n", doc.Repo)
	}
	fmt.Fprintf(&b, "\n%s\n\n", doc.Body)

	b.WriteString("## Goal\n")
	switch goal {
	case GoalSummarize:
		b.WriteString("Produce a research summary with key findings, file references, and links.\n")
	case GoalImplement:
		b.WriteString("Produce an implementation plan with specific file changes and a patch.\n")
	}

	b.WriteString("\n## Tools\n")
	b.WriteString("Call a tool by responding with ONLY a JSON object:\n")
	b.WriteString(`{"tool": "tool_name", "params": {"key": "value"}}`)
	b.WriteString("\n\nWhen you have enough evidence, respond with your final output (not a JSON tool call).\n\n")

	b.WriteString("Available tools:\n")
	for _, t := range tools {
		fmt.Fprintf(&b, "- **%s**: %s. Params: %s\n", t.Name, t.Description, t.Params)
	}

	b.WriteString("\n## Rules\n")
	b.WriteString("- Call tools to gather evidence before answering\n")
	b.WriteString("- Cite specific files and line numbers from tool output\n")
	b.WriteString("- Do not guess — if a tool returned no results, say so\n")
	b.WriteString("- You have a budget of 15 tool calls. Use them wisely.\n")

	return b.String()
}
```

- [ ] **Step 2: Write the loop execution**

```go
// in internal/research/toolloop.go

// ToolCall is a parsed tool invocation from LLM output.
type ToolCall struct {
	Tool   string            `json:"tool"`
	Params map[string]string `json:"params"`
}

// LoopResult is the output of a tool-use research run.
type LoopResult struct {
	RunID      string       `json:"run_id"`
	Document   ResearchDocument `json:"document"`
	Goal       Goal         `json:"goal"`
	Output     string       `json:"output"`
	ToolCalls  []ToolResult `json:"tool_calls"`
	LLMCalls   int          `json:"llm_calls"`
	TokensIn   int          `json:"tokens_in"`
	TokensOut  int          `json:"tokens_out"`
	CostUSD    float64      `json:"cost_usd"`
	MaxTier    int          `json:"max_tier"`
	Escalations int         `json:"escalations"`
	Duration   time.Duration `json:"duration"`
}

const maxToolCalls = 15

// Run executes the tool-use research loop.
func (tl *ToolLoop) Run(ctx context.Context, doc ResearchDocument, goal Goal) (LoopResult, error) {
	runID := esearch.NewRunID()
	start := time.Now()

	sysPrompt := buildSystemPrompt(doc, goal, tl.tools.Definitions())
	var conversation []string
	conversation = append(conversation, sysPrompt)

	result := LoopResult{
		RunID:    runID,
		Document: doc,
		Goal:     goal,
	}

	for i := 0; i < maxToolCalls+5; i++ { // +5 for LLM overhead calls
		prompt := strings.Join(conversation, "\n\n")

		// Validate tool calls: must be valid JSON with known tool name
		validate := func(resp string) provider.EscalationReason {
			resp = strings.TrimSpace(resp)
			// If it's not a tool call, it's the final output — accept it
			if !strings.HasPrefix(resp, "{") {
				return ""
			}
			var tc ToolCall
			if err := json.Unmarshal([]byte(resp), &tc); err != nil {
				return provider.ReasonMalformed
			}
			if tc.Tool != "" && !tl.tools.ValidTool(tc.Tool) {
				return provider.ReasonHallucinated
			}
			return ""
		}

		llmResult, err := tl.runner.Run(ctx, prompt, validate)
		if err != nil {
			// If all tiers fail, return what we have
			result.Output = "Research incomplete — all LLM tiers failed: " + err.Error()
			break
		}

		result.LLMCalls++
		result.TokensIn += llmResult.TokensIn
		result.TokensOut += llmResult.TokensOut
		result.CostUSD += llmResult.CostUSD
		if llmResult.Tier > result.MaxTier {
			result.MaxTier = llmResult.Tier
		}
		if llmResult.Escalated {
			result.Escalations++
		}

		// Index LLM call
		tl.telemetry.IndexLLMCall(ctx, esearch.LLMCallDoc{
			RunID:            runID,
			Step:             "tool_select",
			Tier:             llmResult.Tier,
			Provider:         llmResult.Provider,
			Model:            llmResult.Model,
			TokensIn:         llmResult.TokensIn,
			TokensOut:        llmResult.TokensOut,
			CostUSD:          llmResult.CostUSD,
			LatencyMS:        llmResult.Latency.Milliseconds(),
			Escalated:        llmResult.Escalated,
			EscalationReason: string(llmResult.EscalationReason),
			Timestamp:        time.Now(),
		})

		resp := strings.TrimSpace(llmResult.Response)

		// Try to parse as tool call
		var tc ToolCall
		if err := json.Unmarshal([]byte(resp), &tc); err == nil && tc.Tool != "" {
			if len(result.ToolCalls) >= maxToolCalls {
				conversation = append(conversation, "Tool budget exhausted. Produce your final output now.")
				continue
			}

			fmt.Fprintf(printErr, "  > %s\n", tc.Tool)
			toolStart := time.Now()
			toolResult := tl.tools.Execute(ctx, tc.Tool, tc.Params)
			toolLatency := time.Since(toolStart)

			result.ToolCalls = append(result.ToolCalls, toolResult)

			// Index tool call
			inputJSON, _ := json.Marshal(tc.Params)
			tl.telemetry.IndexToolCall(ctx, esearch.ToolCallDoc{
				RunID:           runID,
				ToolName:        tc.Tool,
				InputSummary:    truncateOutput(string(inputJSON), 200),
				OutputSizeBytes: len(toolResult.Output) + len(toolResult.Err),
				LatencyMS:       toolLatency.Milliseconds(),
				Success:         toolResult.Err == "",
				Timestamp:       time.Now(),
			})

			// Add tool result to conversation
			if toolResult.Err != "" {
				conversation = append(conversation, fmt.Sprintf("Tool %s error: %s", tc.Tool, toolResult.Err))
			} else {
				conversation = append(conversation, fmt.Sprintf("Tool %s result:\n%s", tc.Tool, toolResult.Output))
			}
			continue
		}

		// Not a tool call — this is the final output
		result.Output = resp
		break
	}

	result.Duration = time.Since(start)

	// Index research run
	tl.telemetry.IndexResearchRun(ctx, esearch.ResearchRunDoc{
		RunID:           runID,
		InputSource:     doc.Source,
		SourceURL:       doc.SourceURL,
		Goal:            string(goal),
		TotalToolCalls:  len(result.ToolCalls),
		TotalLLMCalls:   result.LLMCalls,
		TotalTokensIn:   result.TokensIn,
		TotalTokensOut:  result.TokensOut,
		TotalCostUSD:    result.CostUSD,
		DurationMS:      result.Duration.Milliseconds(),
		FinalTierUsed:   result.MaxTier,
		EscalationCount: result.Escalations,
		ConfidencePass:  result.Output != "" && !strings.HasPrefix(result.Output, "Research incomplete"),
		Timestamp:       start,
	})

	return result, nil
}

// printErr is os.Stderr, extracted for testability.
var printErr = stderrWriter{}

type stderrWriter struct{}

func (stderrWriter) Write(p []byte) (int, error) {
	return fmt.Fprint(stderr, string(p))
}

// Injected at init to avoid import cycle; set in cmd/ask.go.
var stderr = devNull{}

type devNull struct{}

func (devNull) Write(p []byte) (int, error) { return len(p), nil }
```

Wait — that stderr pattern is too complex. Simplify:

```go
// Replace the stderr section with:
import "os"

// printToStderr writes status to stderr.
func printToStderr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}
```

And change the `fmt.Fprintf(printErr, ...)` line to `printToStderr("  > %s\n", tc.Tool)`.

- [ ] **Step 3: Write tool-use loop tests**

```go
// internal/research/toolloop_test.go
package research

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	doc := ResearchDocument{
		Source: "github-issue",
		Title:  "Fix error handling",
		Body:   "Errors are not user-friendly",
		Repo:   "elastic/ensemble",
	}
	tools := NewToolSet("/tmp", nil).Definitions()
	prompt := buildSystemPrompt(doc, GoalSummarize, tools)

	if !strings.Contains(prompt, "github-issue") {
		t.Error("expected source in prompt")
	}
	if !strings.Contains(prompt, "Fix error handling") {
		t.Error("expected title in prompt")
	}
	if !strings.Contains(prompt, "grep_code") {
		t.Error("expected tool definitions in prompt")
	}
	if !strings.Contains(prompt, "summarize") && !strings.Contains(prompt, "summary") {
		t.Error("expected goal description in prompt")
	}
}

func TestBuildSystemPromptImplementGoal(t *testing.T) {
	doc := ResearchDocument{Source: "text", Title: "test", Body: "test"}
	tools := NewToolSet("/tmp", nil).Definitions()
	prompt := buildSystemPrompt(doc, GoalImplement, tools)

	if !strings.Contains(prompt, "implementation") {
		t.Error("expected implementation goal in prompt")
	}
}

func TestToolCallParsing(t *testing.T) {
	input := `{"tool": "grep_code", "params": {"pattern": "error"}}`
	var tc ToolCall
	if err := json.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if tc.Tool != "grep_code" {
		t.Errorf("expected grep_code, got %s", tc.Tool)
	}
	if tc.Params["pattern"] != "error" {
		t.Errorf("expected pattern=error, got %s", tc.Params["pattern"])
	}
}

func TestToolCallParsingRejectsNonJSON(t *testing.T) {
	input := "Here is my analysis of the codebase..."
	var tc ToolCall
	err := json.Unmarshal([]byte(input), &tc)
	if err == nil && tc.Tool != "" {
		t.Error("expected non-JSON text to not parse as tool call")
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/research/ -run TestBuildSystem -v && go test ./internal/research/ -run TestToolCall -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/research/toolloop.go internal/research/toolloop_test.go
git commit -m "feat: add tool-use research loop — LLM calls tools iteratively with tiered escalation"
```

---

## Task 6: Update Results Output

**Files:**
- Modify: `internal/research/results.go`
- Modify: `internal/research/results_test.go`

- [ ] **Step 1: Rewrite results.go for new directory structure**

```go
// internal/research/results.go
package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveLoopResult saves a LoopResult to the results directory.
// Structure: results/<org>/<repo>/<number>/
//   summary.md, evidence/, run.json, implementation/ (if goal=implement)
func SaveLoopResult(baseDir string, result LoopResult) error {
	dir := resultDir(baseDir, result)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("results: mkdir: %w", err)
	}

	// Save run.json
	runJSON, _ := json.MarshalIndent(struct {
		RunID      string  `json:"run_id"`
		Source     string  `json:"source"`
		SourceURL  string  `json:"source_url"`
		Goal       string  `json:"goal"`
		ToolCalls  int     `json:"tool_calls"`
		LLMCalls   int     `json:"llm_calls"`
		TokensIn   int     `json:"tokens_in"`
		TokensOut  int     `json:"tokens_out"`
		CostUSD    float64 `json:"cost_usd"`
		MaxTier    int     `json:"max_tier"`
		Escalations int    `json:"escalations"`
		DurationMS int64   `json:"duration_ms"`
	}{
		RunID:       result.RunID,
		Source:      result.Document.Source,
		SourceURL:   result.Document.SourceURL,
		Goal:        string(result.Goal),
		ToolCalls:   len(result.ToolCalls),
		LLMCalls:    result.LLMCalls,
		TokensIn:    result.TokensIn,
		TokensOut:   result.TokensOut,
		CostUSD:     result.CostUSD,
		MaxTier:     result.MaxTier,
		Escalations: result.Escalations,
		DurationMS:  result.Duration.Milliseconds(),
	}, "", "  ")
	os.WriteFile(filepath.Join(dir, "run.json"), runJSON, 0o644)

	// Save evidence
	evidenceDir := filepath.Join(dir, "evidence")
	os.MkdirAll(evidenceDir, 0o755)
	for i, tc := range result.ToolCalls {
		content := tc.Output
		if tc.Err != "" {
			content = "ERROR: " + tc.Err
		}
		name := fmt.Sprintf("%03d-%s.txt", i+1, tc.Tool)
		os.WriteFile(filepath.Join(evidenceDir, name), []byte(content), 0o644)
	}

	// Save summary or implementation
	if result.Goal == GoalImplement {
		implDir := filepath.Join(dir, "implementation")
		os.MkdirAll(implDir, 0o755)
		os.WriteFile(filepath.Join(implDir, "plan.md"), []byte(result.Output), 0o644)
	} else {
		os.WriteFile(filepath.Join(dir, "summary.md"), []byte(result.Output), 0o644)
	}

	return nil
}

func resultDir(baseDir string, result LoopResult) string {
	doc := result.Document
	if doc.Repo != "" {
		parts := strings.SplitN(doc.Repo, "/", 2)
		if len(parts) == 2 {
			num := doc.Metadata["number"]
			if num != "" {
				return filepath.Join(baseDir, parts[0], parts[1], num)
			}
			return filepath.Join(baseDir, parts[0], parts[1])
		}
	}
	return filepath.Join(baseDir, "general")
}
```

- [ ] **Step 2: Write result tests**

```go
// internal/research/results_test.go
package research

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoopResult(t *testing.T) {
	dir := t.TempDir()
	result := LoopResult{
		RunID: "run-test-123",
		Document: ResearchDocument{
			Source:    "github-issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/872",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "872"},
		},
		Goal:    GoalSummarize,
		Output:  "# Summary\n\nThis is a test summary.",
		ToolCalls: []ToolResult{
			{Tool: "grep_code", Output: "cmd/run.go:45: fmt.Errorf(\"failed\")"},
			{Tool: "read_file", Output: "package cmd\n..."},
		},
		LLMCalls:  3,
		TokensIn:  1000,
		TokensOut: 500,
		CostUSD:   0.003,
		Duration:  5 * time.Second,
	}

	err := SaveLoopResult(dir, result)
	if err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	// Check directory structure
	base := filepath.Join(dir, "elastic", "ensemble", "872")
	if _, err := os.Stat(filepath.Join(base, "summary.md")); err != nil {
		t.Error("missing summary.md")
	}
	if _, err := os.Stat(filepath.Join(base, "run.json")); err != nil {
		t.Error("missing run.json")
	}
	if _, err := os.Stat(filepath.Join(base, "evidence", "001-grep_code.txt")); err != nil {
		t.Error("missing evidence/001-grep_code.txt")
	}
	if _, err := os.Stat(filepath.Join(base, "evidence", "002-read_file.txt")); err != nil {
		t.Error("missing evidence/002-read_file.txt")
	}
}

func TestSaveLoopResultImplement(t *testing.T) {
	dir := t.TempDir()
	result := LoopResult{
		RunID:    "run-impl-123",
		Document: ResearchDocument{Source: "text", Repo: "elastic/ensemble", Metadata: map[string]string{"number": "872"}},
		Goal:     GoalImplement,
		Output:   "## Plan\n\nChange cmd/run.go",
		Duration: time.Second,
	}

	err := SaveLoopResult(dir, result)
	if err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	base := filepath.Join(dir, "elastic", "ensemble", "872")
	if _, err := os.Stat(filepath.Join(base, "implementation", "plan.md")); err != nil {
		t.Error("missing implementation/plan.md")
	}
	// summary.md should NOT exist for implement goal
	if _, err := os.Stat(filepath.Join(base, "summary.md")); err == nil {
		t.Error("summary.md should not exist for implement goal")
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/research/ -run TestSaveLoop -v`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/research/results.go internal/research/results_test.go
git commit -m "feat: update results output — structured evidence dir, run.json, implementation support"
```

---

## Task 7: Wire Everything into cmd/ask.go

**Files:**
- Modify: `cmd/ask.go`
- Modify: `cmd/research_helpers.go`

- [ ] **Step 1: Rewrite research_helpers.go**

```go
// cmd/research_helpers.go
package cmd

import (
	"context"
	"fmt"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
)

// buildToolLoop assembles the v2 tool-use research loop.
func buildToolLoop(repoPath string) (*research.ToolLoop, error) {
	cfg, _ := loadConfig()

	// ES client (optional)
	var es *esearch.Client
	esClient := esearch.NewClient("http://localhost:9200")
	if err := esClient.Ping(context.Background()); err == nil {
		es = esClient
	}

	// Telemetry
	tel := esearch.NewTelemetry(es)
	if tel != nil {
		tel.EnsureIndices(context.Background())
	}

	// Tools
	tools := research.NewToolSet(repoPath, es)

	// Tiered runner
	tiers := cfg.Tiers
	if len(tiers) == 0 {
		tiers = provider.DefaultTiers()
	}
	runner := provider.NewTieredRunner(tiers, providerReg)

	return research.NewToolLoop(tools, runner, tel), nil
}
```

- [ ] **Step 2: Rewrite ask.go to use adapters + tool loop**

```go
// cmd/ask.go
package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/research"
	"github.com/8op-org/gl1tch/internal/router"
)

var reGitHubIssueURL = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/issues/\d+`)
var reGitHubPRURL = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/pull/\d+`)
var reGoogleDocURL = regexp.MustCompile(`https?://docs\.google\.com/document/d/`)

func init() {
	askCmd.Flags().StringVarP(&targetPath, "path", "C", "", "run against this directory instead of cwd")
	rootCmd.AddCommand(askCmd)
}

// isResearchInput returns true if the input is a URL that should go to the research loop.
func isResearchInput(input string) bool {
	return reGitHubIssueURL.MatchString(input) ||
		reGitHubPRURL.MatchString(input) ||
		reGoogleDocURL.MatchString(input)
}

var askCmd = &cobra.Command{
	Use:   "ask [input]",
	Short: "route a question or URL to the best workflow",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetPath != "" {
			if err := os.Chdir(targetPath); err != nil {
				return fmt.Errorf("chdir %s: %w", targetPath, err)
			}
		}

		input := strings.Join(args, " ")

		// Research inputs skip the workflow router entirely
		if isResearchInput(input) {
			return runResearch(input, research.GoalSummarize)
		}

		// Check for "implement" intent
		if strings.Contains(strings.ToLower(input), "implement") {
			// Extract the URL or issue ref from the input
			if url := reGitHubIssueURL.FindString(input); url != "" {
				return runResearch(url, research.GoalImplement)
			}
		}

		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}

		// Tier 1: workflow match
		w, resolved, params := router.Match(input, workflows, "")
		if w != nil {
			fmt.Printf(">> %s\n", w.Name)
			result, err := pipeline.Run(w, resolved, "", params, providerReg)
			if err != nil {
				return err
			}
			fmt.Println(result.Output)
			return nil
		}

		// Tier 2: research loop for non-URL questions
		return runResearch(input, research.GoalSummarize)
	},
}

func runResearch(input string, goal research.Goal) error {
	fmt.Fprintln(os.Stderr, ">> adapting input...")
	doc, err := research.Adapt(input)
	if err != nil {
		return fmt.Errorf("adapt: %w", err)
	}
	fmt.Fprintf(os.Stderr, ">> source: %s\n", doc.Source)
	if doc.Repo != "" {
		fmt.Fprintf(os.Stderr, ">> repo: %s\n", doc.Repo)
	}

	repoPath := doc.RepoPath
	if repoPath == "" {
		repoPath, _ = os.Getwd()
	}

	fmt.Fprintln(os.Stderr, ">> researching...")
	loop, err := buildToolLoop(repoPath)
	if err != nil {
		return fmt.Errorf("build loop: %w", err)
	}

	result, err := loop.Run(context.Background(), doc, goal)
	if err != nil {
		return fmt.Errorf("research: %w", err)
	}

	fmt.Println(result.Output)

	// Save results
	dir := "results"
	if saveErr := research.SaveLoopResult(dir, result); saveErr != nil {
		fmt.Fprintf(os.Stderr, ">> warning: could not save results: %v\n", saveErr)
	} else {
		fmt.Fprintf(os.Stderr, ">> results saved to results/\n")
	}

	// Print cost summary
	fmt.Fprintf(os.Stderr, ">> %d tool calls, %d LLM calls, ~$%.4f\n",
		len(result.ToolCalls), result.LLMCalls, result.CostUSD)

	return nil
}

// loadWorkflows loads from ~/.config/glitch/workflows/ then .glitch/workflows/.
func loadWorkflows() (map[string]*pipeline.Workflow, error) {
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

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: build succeeds with no errors

- [ ] **Step 4: Commit**

```bash
git add cmd/ask.go cmd/research_helpers.go
git commit -m "feat: wire tool-use loop into glitch ask — adapters, tiered LLM, ES telemetry"
```

---

## Task 8: Remove Old Research System

**Files:**
- Delete: `internal/research/researcher.go`
- Delete: `internal/research/registry.go`, `registry_test.go`
- Delete: `internal/research/git_researcher.go`, `git_researcher_test.go`
- Delete: `internal/research/fs_researcher.go`, `fs_researcher_test.go`
- Delete: `internal/research/es_researcher.go`
- Delete: `internal/research/yaml_researcher.go`
- Delete: `internal/research/prompts.go`
- Delete: `internal/research/score.go`, `score_test.go`
- Delete: `internal/research/loop.go`, `loop_test.go`
- Delete: `internal/research/events.go`
- Delete: `internal/research/feedback.go`, `feedback_test.go`
- Delete: `researchers/` directory

- [ ] **Step 1: Delete old research files**

```bash
rm internal/research/researcher.go
rm internal/research/registry.go internal/research/registry_test.go
rm internal/research/git_researcher.go internal/research/git_researcher_test.go
rm internal/research/fs_researcher.go internal/research/fs_researcher_test.go
rm internal/research/es_researcher.go
rm internal/research/yaml_researcher.go
rm internal/research/prompts.go
rm internal/research/score.go internal/research/score_test.go
rm internal/research/loop.go internal/research/loop_test.go
rm internal/research/events.go
rm internal/research/feedback.go internal/research/feedback_test.go
rm -rf researchers/
```

- [ ] **Step 2: Fix any remaining import references**

Run: `go build ./... 2>&1`

Fix any compilation errors from removed types/functions. Key places to check:
- `internal/observer/query.go` — may reference old index constants
- Any remaining references to `Registry`, `Researcher`, `Evidence`, `EvidenceBundle`

The `types.go` file should be updated to remove `Evidence`, `EvidenceBundle`, `Score`, `Critique`, `Feedback`, `Result` types that are no longer used. Keep `ResearchQuery` only if still referenced; otherwise remove it too.

- [ ] **Step 3: Run all tests**

Run: `go test ./... 2>&1`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove old researcher system — registry, prompts, scoring, YAML researchers"
```

---

## Task 9: Kibana Dashboards

**Files:**
- Create: `deploy/kibana/` directory
- Create: `deploy/kibana/import.sh`

- [ ] **Step 1: Create dashboard import script**

```bash
#!/usr/bin/env bash
# deploy/kibana/import.sh — imports saved objects into Kibana
set -euo pipefail

KIBANA_URL="${KIBANA_URL:-http://localhost:5601}"

echo "Waiting for Kibana..."
for i in $(seq 1 30); do
  if curl -sf "$KIBANA_URL/api/status" > /dev/null 2>&1; then
    break
  fi
  sleep 2
done

for f in deploy/kibana/*.ndjson; do
  [ -f "$f" ] || continue
  echo "Importing $(basename "$f")..."
  curl -sf -X POST "$KIBANA_URL/api/saved_objects/_import?overwrite=true" \
    -H "kbn-xsrf: true" \
    --form file=@"$f" > /dev/null
done

echo "Dashboards imported."
```

- [ ] **Step 2: Create Research Overview dashboard NDJSON**

This is generated via Kibana's export API. For the initial version, create a minimal index pattern + dashboard. The full dashboard panels will be refined after real data flows in.

Create `deploy/kibana/index-patterns.ndjson`:

```json
{"type":"index-pattern","id":"glitch-research-runs","attributes":{"title":"glitch-research-runs","timeFieldName":"timestamp"}}
{"type":"index-pattern","id":"glitch-tool-calls","attributes":{"title":"glitch-tool-calls","timeFieldName":"timestamp"}}
{"type":"index-pattern","id":"glitch-llm-calls","attributes":{"title":"glitch-llm-calls","timeFieldName":"timestamp"}}
```

- [ ] **Step 3: Wire import into glitch up**

Check how `glitch up` works and add dashboard import after docker compose up succeeds. Find the `up` command:

Run: `grep -rn "glitch up\|upCmd\|\"up\"" cmd/`

Add the import script call after compose up completes.

- [ ] **Step 4: Make import script executable and commit**

```bash
chmod +x deploy/kibana/import.sh
git add deploy/kibana/
git commit -m "feat: add Kibana dashboard scaffolding — index patterns and import script"
```

---

## Task 10: Integration Test

- [ ] **Step 1: Verify full flow end-to-end**

Run against the ensemble issue to verify the full pipeline works:

```bash
go build -o /tmp/glitch-v2 .
/tmp/glitch-v2 ask "https://github.com/elastic/ensemble/issues/872" 2>&1
```

Expected:
- `>> adapting input...`
- `>> source: github-issue`
- `>> repo: elastic/ensemble`
- `>> researching...`
- Tool calls printed (e.g., `> grep_code`, `> read_file`, `> git_log`)
- Summary output with file references
- `>> results saved to results/`
- `>> N tool calls, N LLM calls, ~$X.XXXX`

- [ ] **Step 2: Check results directory**

```bash
ls -la results/elastic/ensemble/872/
cat results/elastic/ensemble/872/summary.md
cat results/elastic/ensemble/872/run.json
ls results/elastic/ensemble/872/evidence/
```

Expected: summary.md with real findings, run.json with telemetry, evidence/ with tool outputs

- [ ] **Step 3: Check ES telemetry (if ES running)**

```bash
curl -s localhost:9200/glitch-research-runs/_search?pretty | head -30
curl -s localhost:9200/glitch-tool-calls/_search?pretty | head -30
curl -s localhost:9200/glitch-llm-calls/_search?pretty | head -30
```

Expected: documents present in all three indices

- [ ] **Step 4: Install and run smoke test**

```bash
go install .
glitch ask "https://github.com/elastic/ensemble/issues/872"
```

- [ ] **Step 5: Commit any fixes**

```bash
git add -A
git commit -m "fix: integration test fixes for research system v2"
```

---

## Summary

| Task | What it builds | Key files |
|------|---------------|-----------|
| 1 | Input adapters | document.go, adapter.go |
| 2 | Research tools | tools.go |
| 3 | Tiered LLM escalation | tiers.go, tokens.go |
| 4 | ES telemetry | telemetry.go, mappings.go |
| 5 | Tool-use loop engine | toolloop.go |
| 6 | Results output | results.go |
| 7 | Wire into cmd/ask | ask.go, research_helpers.go |
| 8 | Remove old system | delete 15+ files |
| 9 | Kibana dashboards | deploy/kibana/ |
| 10 | Integration test | end-to-end verification |

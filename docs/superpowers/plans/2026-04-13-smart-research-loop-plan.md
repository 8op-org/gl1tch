# Smart Research Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the research loop so `glitch ask "..."` handles any question intelligently — native git/fs researchers, paid model feedback, auto-clone/index, results extraction, and missing-researcher guidance.

**Architecture:** Four native Go researchers (git, fs, es-activity, es-code) always available. YAML researchers for tool-specific commands. Paid model drafts with full autonomy and provides feedback that trains the local planner. Results auto-saved with mirrored repo structure when substantive.

**Tech Stack:** Go 1.26, Elasticsearch 8.17, SQLite (modernc.org/sqlite), Ollama (qwen2.5:7b), Claude (paid provider via stdin)

---

## File Structure

```
internal/research/
├── git_researcher.go          (new — native git researcher)
├── git_researcher_test.go     (new)
├── fs_researcher.go           (new — native filesystem researcher)
├── fs_researcher_test.go      (new)
├── repo.go                    (new — auto-clone + auto-index logic)
├── repo_test.go               (new)
├── feedback.go                (new — parse paid model feedback, store in SQLite)
├── feedback_test.go           (new)
├── results.go                 (new — extract files from draft, write to results/)
├── results_test.go            (new)
├── loop.go                    (modify — missing researcher guidance, feedback stage, auto-save)
├── prompts.go                 (modify — update draft prompt to request feedback)
├── events.go                  (modify — add EventTypeFeedback)
├── types.go                   (modify — add Feedback field to Result)
├── es_researcher.go           (existing — no changes)
├── yaml_researcher.go         (existing — no changes)
├── registry.go                (existing — no changes)
├── score.go                   (existing — no changes)
└── researcher.go              (existing — no changes)

cmd/
├── ask.go                     (modify — handle results saving + feedback printing)
├── research_helpers.go        (modify — register native researchers, wire paid model + feedback)

researchers/                   (existing YAML — update github-issue.yaml)
├── git-log.yaml               (delete — replaced by native git researcher)
├── github-prs.yaml            (existing — keep)
├── github-issues.yaml         (existing — keep)
└── github-issue.yaml          (new — single issue fetch)
```

---

### Task 1: Auto-Clone + Auto-Index (`internal/research/repo.go`)

**Files:**
- Create: `internal/research/repo.go`
- Create: `internal/research/repo_test.go`

- [ ] **Step 1: Write repo_test.go**

```go
package research

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepoDir(t *testing.T) {
	dir := RepoDir("elastic", "observability-robots")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "share", "glitch", "repos", "elastic", "observability-robots")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
}

func TestEnsureRepoLocal(t *testing.T) {
	// Test with a path that already exists (cwd)
	cwd, _ := os.Getwd()
	got, err := EnsureRepo("", "", cwd)
	if err != nil {
		t.Fatalf("EnsureRepo with existing path: %v", err)
	}
	if got != cwd {
		t.Fatalf("got %q, want %q", got, cwd)
	}
}

func TestParseRepoFromQuestion(t *testing.T) {
	cases := []struct {
		input    string
		wantOrg  string
		wantRepo string
	}{
		{"fix observability-robots#3928", "elastic", "observability-robots"},
		{"what's in elastic/ensemble", "elastic", "ensemble"},
		{"check the gl1tch code", "", ""},
	}
	for _, tc := range cases {
		org, repo := ParseRepoFromQuestion(tc.input)
		if org != tc.wantOrg || repo != tc.wantRepo {
			t.Errorf("ParseRepoFromQuestion(%q) = %q/%q, want %q/%q", tc.input, org, repo, tc.wantOrg, tc.wantRepo)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/research/... -run TestRepo -v`
Expected: FAIL — functions don't exist

- [ ] **Step 3: Create repo.go**

```go
package research

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var reRepoRef = regexp.MustCompile(`(?:([a-zA-Z0-9_.-]+)/)?([a-zA-Z0-9_.-]+)(?:#\d+)?`)

// RepoDir returns the standard clone path for a repo.
func RepoDir(org, repo string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "glitch", "repos", org, repo)
}

// EnsureRepo ensures a repo is available locally. Checks:
// 1. Explicit localPath (if provided and exists, use it)
// 2. ~/Projects/<repo>
// 3. ~/.local/share/glitch/repos/<org>/<repo> (clone if missing)
// Returns the local path to the repo.
func EnsureRepo(org, repo, localPath string) (string, error) {
	if localPath != "" {
		if info, err := os.Stat(localPath); err == nil && info.IsDir() {
			return localPath, nil
		}
	}

	// Check ~/Projects/<repo>
	if home, err := os.UserHomeDir(); err == nil && repo != "" {
		candidate := filepath.Join(home, "Projects", repo)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	if org == "" || repo == "" {
		return "", fmt.Errorf("repo: cannot clone without org/repo")
	}

	// Check standard clone dir
	dir := RepoDir(org, repo)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		// Freshen with pull
		cmd := exec.Command("git", "-C", dir, "pull", "--ff-only", "-q")
		cmd.Run() // best-effort
		return dir, nil
	}

	// Clone
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return "", fmt.Errorf("repo: mkdir: %w", err)
	}
	remote := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)
	cmd := exec.Command("git", "clone", "--depth=1", remote, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("repo: clone %s: %w\n%s", remote, err, out)
	}
	return dir, nil
}

// ParseRepoFromQuestion extracts org/repo from a question string.
// Returns empty strings if no repo reference found.
func ParseRepoFromQuestion(question string) (org, repo string) {
	// Look for org/repo or repo#number patterns
	m := reRepoRef.FindStringSubmatch(question)
	if m == nil {
		return "", ""
	}
	org = m[1]
	repo = m[2]
	// Skip common words that aren't repos
	skip := map[string]bool{"the": true, "this": true, "that": true, "what": true, "fix": true, "check": true}
	if skip[strings.ToLower(repo)] {
		return "", ""
	}
	if org == "" {
		org = "elastic" // default org
	}
	return org, repo
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/research/... -run TestRepo -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/research/repo.go internal/research/repo_test.go
git commit -m "feat: add repo auto-clone and resolution for research loop"
```

---

### Task 2: Native Git Researcher

**Files:**
- Create: `internal/research/git_researcher.go`
- Create: `internal/research/git_researcher_test.go`
- Delete: `researchers/git-log.yaml`

- [ ] **Step 1: Write git_researcher_test.go**

```go
package research

import (
	"context"
	"strings"
	"testing"
)

func TestGitResearcherName(t *testing.T) {
	g := &GitResearcher{}
	if g.Name() != "git" {
		t.Fatalf("got %q, want %q", g.Name(), "git")
	}
}

func TestGitResearcherGather(t *testing.T) {
	// This test runs in the gl1tch repo itself
	g := &GitResearcher{}
	q := ResearchQuery{Question: "what recent commits were made?"}
	ev, err := g.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if ev.Source != "git" {
		t.Fatalf("source: got %q, want %q", ev.Source, "git")
	}
	if ev.Body == "" {
		t.Fatal("expected non-empty body")
	}
	// Should contain commit hashes (7+ hex chars)
	if !strings.ContainsAny(ev.Body, "0123456789abcdef") {
		t.Fatal("expected git log output with commit hashes")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/research/... -run TestGitResearcher -v`
Expected: FAIL

- [ ] **Step 3: Create git_researcher.go**

```go
package research

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitResearcher is a native researcher that queries git history, diffs, and content.
type GitResearcher struct {
	// RepoPath overrides the working directory. Empty means cwd.
	RepoPath string
}

func (g *GitResearcher) Name() string { return "git" }
func (g *GitResearcher) Describe() string {
	return "git history, diffs, remotes, and blame for the current or target repository"
}

func (g *GitResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	var sections []string

	// Always include recent log
	if out := g.run("log", "--oneline", "-30"); out != "" {
		sections = append(sections, "=== Recent Commits ===\n"+out)
	}

	// Include remote info
	if out := g.run("remote", "-v"); out != "" {
		sections = append(sections, "=== Remotes ===\n"+out)
	}

	question := strings.ToLower(q.Question)

	// If question is about changes/diffs
	if containsAny(question, "change", "diff", "modif", "broke", "break", "fix") {
		if out := g.run("diff", "--stat", "HEAD~10"); out != "" {
			sections = append(sections, "=== Recent Diff Stats ===\n"+out)
		}
	}

	// If question is about a specific term, search log
	keywords := extractKeywords(question)
	for _, kw := range keywords {
		if out := g.run("log", "--oneline", "--all", "--grep="+kw, "-10"); out != "" {
			sections = append(sections, fmt.Sprintf("=== Commits mentioning %q ===\n%s", kw, out))
		}
	}

	body := strings.Join(sections, "\n\n")
	if body == "" {
		return Evidence{}, fmt.Errorf("git: no output")
	}

	return Evidence{
		Source: "git",
		Title:  "git history and context",
		Body:   body,
	}, nil
}

func (g *GitResearcher) run(args ...string) string {
	if g.RepoPath != "" {
		args = append([]string{"-C", g.RepoPath}, args...)
	}
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// containsAny returns true if s contains any of the substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// extractKeywords pulls non-stopword tokens > 3 chars from a question.
func extractKeywords(question string) []string {
	stops := map[string]bool{
		"what": true, "when": true, "where": true, "which": true, "that": true,
		"this": true, "have": true, "been": true, "from": true, "with": true,
		"were": true, "there": true, "their": true, "about": true, "does": true,
	}
	var kws []string
	for _, word := range strings.Fields(question) {
		w := strings.ToLower(strings.Trim(word, "?.,!\"'"))
		if len(w) > 3 && !stops[w] {
			kws = append(kws, w)
		}
	}
	if len(kws) > 3 {
		kws = kws[:3]
	}
	return kws
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/research/... -run TestGitResearcher -v`
Expected: PASS

- [ ] **Step 5: Delete old git-log.yaml**

```bash
rm researchers/git-log.yaml
```

- [ ] **Step 6: Commit**

```bash
git add internal/research/git_researcher.go internal/research/git_researcher_test.go
git rm researchers/git-log.yaml
git commit -m "feat: add native git researcher, remove git-log YAML"
```

---

### Task 3: Native FS Researcher

**Files:**
- Create: `internal/research/fs_researcher.go`
- Create: `internal/research/fs_researcher_test.go`

- [ ] **Step 1: Write fs_researcher_test.go**

```go
package research

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFSResearcherName(t *testing.T) {
	f := &FSResearcher{}
	if f.Name() != "fs" {
		t.Fatalf("got %q, want %q", f.Name(), "fs")
	}
}

func TestFSResearcherGatherScan(t *testing.T) {
	// Create temp dir with a file containing TBC
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "docs"), 0o755)
	os.WriteFile(filepath.Join(dir, "docs", "test.md"), []byte("# Title\n\n> TBC\n\nSome content"), 0o644)

	f := &FSResearcher{RootPath: dir}
	q := ResearchQuery{Question: "what placeholders are in the docs?"}
	ev, err := f.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if !strings.Contains(ev.Body, "TBC") {
		t.Fatal("expected body to contain TBC scan results")
	}
}

func TestFSResearcherGatherReadFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("hello world"), 0o644)

	f := &FSResearcher{RootPath: dir}
	q := ResearchQuery{Question: "read readme.md"}
	ev, err := f.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if !strings.Contains(ev.Body, "hello world") {
		t.Fatalf("expected file content, got: %s", ev.Body)
	}
}

func TestFSResearcherGatherTree(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src", "pkg"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "pkg", "main.go"), []byte("package main"), 0o644)

	f := &FSResearcher{RootPath: dir}
	q := ResearchQuery{Question: "what is the project structure?"}
	ev, err := f.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if !strings.Contains(ev.Body, "main.go") {
		t.Fatalf("expected tree output, got: %s", ev.Body)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/research/... -run TestFSResearcher -v`
Expected: FAIL

- [ ] **Step 3: Create fs_researcher.go**

```go
package research

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var reFilePath = regexp.MustCompile(`(?:^|\s)([\w./-]+\.\w{1,6})`)

// FSResearcher reads files, lists directories, and scans for patterns.
type FSResearcher struct {
	RootPath string // empty means cwd
}

func (f *FSResearcher) Name() string { return "fs" }
func (f *FSResearcher) Describe() string {
	return "read files, list directories, search file contents, and scan for patterns like TBC/TODO"
}

func (f *FSResearcher) Gather(ctx context.Context, q ResearchQuery, prior EvidenceBundle) (Evidence, error) {
	root := f.RootPath
	if root == "" {
		root, _ = os.Getwd()
	}

	var sections []string
	question := strings.ToLower(q.Question)

	// If question mentions specific file paths, read them
	paths := extractFilePaths(q.Question)
	// Also check evidence from other researchers for file paths
	for _, ev := range prior.Items {
		paths = append(paths, extractFilePaths(ev.Body)...)
	}
	paths = dedupStrings(paths)
	for _, p := range paths {
		full := filepath.Join(root, p)
		data, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		// Cap at 5000 bytes per file
		content := string(data)
		if len(content) > 5000 {
			content = content[:5000] + "\n... (truncated)"
		}
		sections = append(sections, fmt.Sprintf("=== FILE: %s ===\n%s\n=== END: %s ===", p, content, p))
	}

	// If question is about placeholders/missing content, scan
	if containsAny(question, "tbc", "tbd", "todo", "fixme", "placeholder", "missing", "incomplete") {
		if out := grepDir(root, `TBC\|TBD\|TODO\|FIXME`, "*.md"); out != "" {
			sections = append(sections, "=== Placeholder Scan ===\n"+out)
		}
	}

	// If question is about structure/layout/what's in the repo
	if containsAny(question, "structure", "layout", "tree", "directory", "what's in", "files in", "project") {
		if out := listTree(root, 3); out != "" {
			sections = append(sections, "=== Directory Tree ===\n"+out)
		}
	}

	// If we found nothing specific, do a general content search with keywords
	if len(sections) == 0 {
		kws := extractKeywords(question)
		for _, kw := range kws {
			if out := grepDir(root, kw, "*.md"); out != "" {
				sections = append(sections, fmt.Sprintf("=== Files matching %q ===\n%s", kw, out))
			}
		}
	}

	// If still nothing, list the tree
	if len(sections) == 0 {
		if out := listTree(root, 2); out != "" {
			sections = append(sections, "=== Directory Tree ===\n"+out)
		}
	}

	body := strings.Join(sections, "\n\n")
	if body == "" {
		return Evidence{}, fmt.Errorf("fs: no results found")
	}

	return Evidence{
		Source: "fs",
		Title:  "filesystem scan",
		Body:   body,
	}, nil
}

func extractFilePaths(s string) []string {
	matches := reFilePath.FindAllStringSubmatch(s, -1)
	var out []string
	for _, m := range matches {
		p := m[1]
		// Filter out things that aren't paths
		if strings.HasPrefix(p, "http") || strings.HasPrefix(p, "//") {
			continue
		}
		out = append(out, p)
	}
	return out
}

func dedupStrings(in []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func grepDir(dir, pattern, glob string) string {
	cmd := exec.Command("grep", "-rn", pattern, dir, "--include="+glob)
	out, _ := cmd.Output()
	s := strings.TrimSpace(string(out))
	// Cap at 50 lines
	lines := strings.Split(s, "\n")
	if len(lines) > 50 {
		s = strings.Join(lines[:50], "\n") + fmt.Sprintf("\n... (%d more lines)", len(lines)-50)
	}
	return s
}

func listTree(dir string, depth int) string {
	cmd := exec.Command("find", dir, "-maxdepth", fmt.Sprintf("%d", depth),
		"-type", "f",
		"-not", "-path", "*/.git/*",
		"-not", "-path", "*/node_modules/*",
		"-not", "-path", "*/vendor/*",
		"-not", "-path", "*/.worktrees/*",
	)
	out, _ := cmd.Output()
	s := strings.TrimSpace(string(out))
	// Make paths relative
	s = strings.ReplaceAll(s, dir+"/", "")
	lines := strings.Split(s, "\n")
	if len(lines) > 100 {
		s = strings.Join(lines[:100], "\n") + fmt.Sprintf("\n... (%d more files)", len(lines)-100)
	}
	return s
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/research/... -run TestFSResearcher -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/research/fs_researcher.go internal/research/fs_researcher_test.go
git commit -m "feat: add native fs researcher — read files, scan patterns, list trees"
```

---

### Task 4: Feedback Parsing + Storage

**Files:**
- Create: `internal/research/feedback.go`
- Create: `internal/research/feedback_test.go`
- Modify: `internal/research/types.go`
- Modify: `internal/research/events.go`

- [ ] **Step 1: Write feedback_test.go**

```go
package research

import (
	"testing"
)

func TestParseFeedback(t *testing.T) {
	raw := `Here are the fixed files.

--- FEEDBACK ---
- evidence_quality: good
- missing: ["directory listing for docs/teams/ci/macos/", "sibling file orka.md"]
- useful: ["repo-scan found all TBC placeholders"]
- suggestion: "for doc fixes, include sibling files in the same directory"
`
	fb, err := ParseFeedback(raw)
	if err != nil {
		t.Fatalf("ParseFeedback: %v", err)
	}
	if fb.Quality != "good" {
		t.Fatalf("quality: got %q, want %q", fb.Quality, "good")
	}
	if len(fb.Missing) != 2 {
		t.Fatalf("missing: got %d, want 2", len(fb.Missing))
	}
	if fb.Suggestion == "" {
		t.Fatal("expected non-empty suggestion")
	}
}

func TestParseFeedbackNoSection(t *testing.T) {
	raw := "Just a plain answer with no feedback section."
	fb, err := ParseFeedback(raw)
	if err != nil {
		t.Fatalf("ParseFeedback: %v", err)
	}
	if fb.Quality != "" {
		t.Fatalf("expected empty quality, got %q", fb.Quality)
	}
}

func TestSplitDraftAndFeedback(t *testing.T) {
	raw := "The fixes are:\n\nblah blah\n\n--- FEEDBACK ---\n- evidence_quality: adequate\n- suggestion: test"
	draft, fb := SplitDraftAndFeedback(raw)
	if !containsAny(draft, "blah") {
		t.Fatalf("draft should contain content, got: %s", draft)
	}
	if fb == "" {
		t.Fatal("expected non-empty feedback section")
	}
}
```

- [ ] **Step 2: Add Feedback type to types.go**

Append to `internal/research/types.go`:

```go
// Feedback captures the paid model's assessment of evidence quality.
type Feedback struct {
	Quality    string   `json:"quality"`    // good, adequate, insufficient
	Missing    []string `json:"missing"`    // what evidence was missing
	Useful     []string `json:"useful"`     // what evidence was helpful
	Suggestion string   `json:"suggestion"` // how to improve next time
}
```

And add `Feedback` field to `Result`:

```go
type Result struct {
	Query      ResearchQuery  `json:"query"`
	Draft      string         `json:"draft"`
	Bundle     EvidenceBundle `json:"bundle"`
	Score      Score          `json:"score"`
	Reason     Reason         `json:"reason"`
	Iterations int            `json:"iterations"`
	Feedback   Feedback       `json:"feedback,omitempty"`
}
```

- [ ] **Step 3: Add EventTypeFeedback to events.go**

Add to the const block in `internal/research/events.go`:

```go
const (
	EventTypeAttempt  EventType = "research_attempt"
	EventTypeScore    EventType = "research_score"
	EventTypeFeedback EventType = "research_feedback"
)
```

- [ ] **Step 4: Create feedback.go**

```go
package research

import (
	"encoding/json"
	"regexp"
	"strings"
)

var reFeedbackSection = regexp.MustCompile(`(?s)---\s*FEEDBACK\s*---\s*(.+)`)

// SplitDraftAndFeedback separates the draft content from the feedback section.
func SplitDraftAndFeedback(raw string) (draft, feedbackRaw string) {
	loc := reFeedbackSection.FindStringIndex(raw)
	if loc == nil {
		return strings.TrimSpace(raw), ""
	}
	return strings.TrimSpace(raw[:loc[0]]), strings.TrimSpace(raw[loc[0]:])
}

// ParseFeedback extracts structured feedback from the feedback section.
func ParseFeedback(raw string) (Feedback, error) {
	_, section := SplitDraftAndFeedback(raw)
	if section == "" {
		return Feedback{}, nil
	}

	var fb Feedback
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")

		if strings.HasPrefix(line, "evidence_quality:") {
			fb.Quality = cleanValue(strings.TrimPrefix(line, "evidence_quality:"))
		} else if strings.HasPrefix(line, "missing:") {
			fb.Missing = parseJSONList(strings.TrimPrefix(line, "missing:"))
		} else if strings.HasPrefix(line, "useful:") {
			fb.Useful = parseJSONList(strings.TrimPrefix(line, "useful:"))
		} else if strings.HasPrefix(line, "suggestion:") {
			fb.Suggestion = cleanValue(strings.TrimPrefix(line, "suggestion:"))
		}
	}
	return fb, nil
}

func cleanValue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	return s
}

func parseJSONList(s string) []string {
	s = strings.TrimSpace(s)
	var items []string
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		// Fallback: split by comma
		for _, item := range strings.Split(s, ",") {
			item = cleanValue(item)
			item = strings.Trim(item, "[]")
			item = strings.TrimSpace(item)
			if item != "" {
				items = append(items, item)
			}
		}
	}
	return items
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/research/... -run TestParseFeedback -v && go test ./internal/research/... -run TestSplitDraft -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/research/feedback.go internal/research/feedback_test.go internal/research/types.go internal/research/events.go
git commit -m "feat: add feedback parsing — paid model rates evidence quality"
```

---

### Task 5: Results Extraction

**Files:**
- Create: `internal/research/results.go`
- Create: `internal/research/results_test.go`

- [ ] **Step 1: Write results_test.go**

```go
package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFiles(t *testing.T) {
	draft := `Here are the fixes.

--- FILE: docs/teams/ci/macos/index.md ---
# macOS Runners

Some content here.
--- END FILE ---

--- FILE: docs/teams/ci/dependencies/updatecli.md ---
# Updatecli

More content.
--- END FILE ---
`
	files := ExtractFiles(draft)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "docs/teams/ci/macos/index.md" {
		t.Fatalf("path: got %q", files[0].Path)
	}
	if !strings.Contains(files[0].Content, "macOS Runners") {
		t.Fatal("expected content")
	}
}

func TestSaveResults(t *testing.T) {
	dir := t.TempDir()
	result := Result{
		Draft: "--- FILE: docs/test.md ---\n# Test\n--- END FILE ---",
		Feedback: Feedback{
			Quality:    "good",
			Suggestion: "test suggestion",
		},
	}
	err := SaveResults(dir, result)
	if err != nil {
		t.Fatalf("SaveResults: %v", err)
	}

	// Check drafts.md exists
	if _, err := os.Stat(filepath.Join(dir, "drafts.md")); err != nil {
		t.Fatal("drafts.md not created")
	}
	// Check extracted file exists
	if _, err := os.Stat(filepath.Join(dir, "docs", "test.md")); err != nil {
		t.Fatal("extracted file not created")
	}
	// Check feedback.md exists
	if _, err := os.Stat(filepath.Join(dir, "feedback.md")); err != nil {
		t.Fatal("feedback.md not created")
	}
}

func TestIsSubstantive(t *testing.T) {
	if IsSubstantive("short answer") {
		t.Fatal("short answer should not be substantive")
	}
	if !IsSubstantive(strings.Repeat("x", 501)) {
		t.Fatal("long answer should be substantive")
	}
	if !IsSubstantive("--- FILE: foo.md ---\ncontent\n--- END FILE ---") {
		t.Fatal("answer with files should be substantive")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/research/... -run TestExtract -v`
Expected: FAIL

- [ ] **Step 3: Create results.go**

```go
package research

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reFileBlock = regexp.MustCompile(`(?ms)^---\s*FILE:\s*(.+?)\s*---\s*\n(.*?)^---\s*END FILE\s*---`)

// ExtractedFile is a single file extracted from draft output.
type ExtractedFile struct {
	Path    string
	Content string
}

// ExtractFiles pulls file blocks from draft output.
func ExtractFiles(draft string) []ExtractedFile {
	matches := reFileBlock.FindAllStringSubmatch(draft, -1)
	var files []ExtractedFile
	for _, m := range matches {
		files = append(files, ExtractedFile{
			Path:    strings.TrimSpace(m[1]),
			Content: m[2],
		})
	}
	return files
}

// IsSubstantive returns true if the draft warrants saving to results/.
func IsSubstantive(draft string) bool {
	if len(draft) > 500 {
		return true
	}
	if strings.Contains(draft, "--- FILE:") {
		return true
	}
	return false
}

// SaveResults writes the full result to a results directory.
func SaveResults(dir string, result Result) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("results: mkdir: %w", err)
	}

	// Write full draft
	if err := os.WriteFile(filepath.Join(dir, "drafts.md"), []byte(result.Draft), 0o644); err != nil {
		return fmt.Errorf("results: write drafts: %w", err)
	}

	// Extract and write individual files
	files := ExtractFiles(result.Draft)
	for _, f := range files {
		target := filepath.Join(dir, f.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("results: mkdir %s: %w", f.Path, err)
		}
		if err := os.WriteFile(target, []byte(f.Content), 0o644); err != nil {
			return fmt.Errorf("results: write %s: %w", f.Path, err)
		}
	}

	// Write feedback if present
	if result.Feedback.Quality != "" || result.Feedback.Suggestion != "" {
		var fb strings.Builder
		fmt.Fprintf(&fb, "# Research Feedback\n\n")
		if result.Feedback.Quality != "" {
			fmt.Fprintf(&fb, "**Evidence Quality:** %s\n\n", result.Feedback.Quality)
		}
		if len(result.Feedback.Missing) > 0 {
			fb.WriteString("**Missing:**\n")
			for _, m := range result.Feedback.Missing {
				fmt.Fprintf(&fb, "- %s\n", m)
			}
			fb.WriteString("\n")
		}
		if len(result.Feedback.Useful) > 0 {
			fb.WriteString("**Useful:**\n")
			for _, u := range result.Feedback.Useful {
				fmt.Fprintf(&fb, "- %s\n", u)
			}
			fb.WriteString("\n")
		}
		if result.Feedback.Suggestion != "" {
			fmt.Fprintf(&fb, "**Suggestion:** %s\n", result.Feedback.Suggestion)
		}
		if err := os.WriteFile(filepath.Join(dir, "feedback.md"), []byte(fb.String()), 0o644); err != nil {
			return fmt.Errorf("results: write feedback: %w", err)
		}
	}

	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/research/... -run "TestExtract|TestSaveResults|TestIsSubstantive" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/research/results.go internal/research/results_test.go
git commit -m "feat: add results extraction — parse draft files, save mirrored structure"
```

---

### Task 6: Update Draft Prompt + Missing Researcher Guidance

**Files:**
- Modify: `internal/research/prompts.go`
- Modify: `internal/research/loop.go`

- [ ] **Step 1: Update DraftPrompt in prompts.go to request feedback**

Replace the existing `DraftPrompt` function in `internal/research/prompts.go`:

```go
// DraftPrompt builds a prompt that instructs the LLM to answer using only evidence,
// and to provide structured feedback on evidence quality.
func DraftPrompt(question string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString("You are a research assistant with full autonomy. Answer the question using the evidence provided below.\n\n")
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\n\nEvidence:\n")
	for i, e := range bundle.Items {
		fmt.Fprintf(&b, "\n--- Evidence %d (source: %s) ---\n", i+1, e.Source)
		if e.Title != "" {
			fmt.Fprintf(&b, "Title: %s\n", e.Title)
		}
		b.WriteString(e.Body)
		b.WriteString("\n")
	}
	b.WriteString("\nRules:\n")
	b.WriteString("- If the answer involves file changes, output each changed file as:\n")
	b.WriteString("  --- FILE: path/to/file.md ---\n")
	b.WriteString("  (complete file content)\n")
	b.WriteString("  --- END FILE ---\n")
	b.WriteString("- Cite only verbatim identifiers from the evidence\n")
	b.WriteString("- Say \"I don't have enough evidence\" if nothing relevant is provided\n")
	b.WriteString("- Never suggest commands to run — just produce the output\n\n")
	b.WriteString("After your answer, append a feedback section rating the evidence you received:\n\n")
	b.WriteString("--- FEEDBACK ---\n")
	b.WriteString("- evidence_quality: good|adequate|insufficient\n")
	b.WriteString("- missing: [\"what evidence you needed but didn't get\"]\n")
	b.WriteString("- useful: [\"what evidence was most helpful\"]\n")
	b.WriteString("- suggestion: \"how to improve evidence gathering next time\"\n")
	return b.String()
}
```

- [ ] **Step 2: Update loop.go — missing researcher guidance + feedback parsing**

Replace the `plan` method in `internal/research/loop.go` to print missing researchers:

```go
// plan calls the LLM with PlanPrompt, parses researcher names, validates against registry.
// Prints missing researcher guidance to stderr.
func (l *Loop) plan(ctx context.Context, question string, already map[string]struct{}) ([]string, error) {
	researchers := l.registry.List()
	if len(researchers) == 0 {
		return nil, fmt.Errorf("no researchers registered")
	}

	prompt := PlanPrompt(question, researchers, l.hints)
	raw, err := l.llm(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("plan LLM call: %w", err)
	}

	names, err := ParsePlan(raw)
	if err != nil {
		return nil, fmt.Errorf("plan parse: %w", err)
	}

	// Validate against registry, report missing ones.
	var valid []string
	for _, name := range names {
		if _, ok := l.registry.Lookup(name); !ok {
			fmt.Fprintf(os.Stderr, ">> missing researcher: %s — add ~/.config/glitch/researchers/%s.yaml\n", name, name)
			continue
		}
		valid = append(valid, name)
	}

	if len(valid) > 0 {
		fmt.Fprintf(os.Stderr, ">> using: %s\n", strings.Join(valid, ", "))
	}

	return filterPicked(valid, already), nil
}
```

Add `"os"` to the imports in loop.go.

Update the `Run` method to parse feedback from the draft and set it on the Result. After the draft call:

```go
		// Parse feedback from draft
		draftContent, _ := SplitDraftAndFeedback(draft)
		feedback, _ := ParseFeedback(draft)
		draft = draftContent // strip feedback from the user-visible draft
```

And set the feedback on the result:

```go
		result := Result{
			Query:      q,
			Draft:      draft,
			Bundle:     bundle,
			Score:      score,
			Iterations: iter,
			Feedback:   feedback,
		}
```

- [ ] **Step 3: Run all tests**

Run: `go test ./internal/research/... -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/research/prompts.go internal/research/loop.go
git commit -m "feat: draft prompt requests feedback, loop prints missing researchers"
```

---

### Task 7: Wire Everything in cmd/

**Files:**
- Modify: `cmd/research_helpers.go`
- Modify: `cmd/ask.go`
- Create: `researchers/github-issue.yaml`
- Delete: `researchers/doc-fix.yaml` (if exists in workflows)

- [ ] **Step 1: Update cmd/research_helpers.go**

Replace the entire file:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
)

// buildResearchLoop assembles the research loop with all available researchers.
func buildResearchLoop() *research.Loop {
	reg := research.NewRegistry()

	// 1. Native researchers — always available
	reg.Register(&research.GitResearcher{})
	reg.Register(&research.FSResearcher{})

	// 2. ES researchers if ES is reachable
	es := esearch.NewClient("http://localhost:9200")
	if err := es.Ping(context.Background()); err == nil {
		reg.Register(research.NewESActivityResearcher(es))
		reg.Register(research.NewESCodeResearcher(es))
	}

	// 3. YAML researchers from ~/.config/glitch/researchers/
	if home, err := os.UserHomeDir(); err == nil {
		researcherDir := filepath.Join(home, ".config", "glitch", "researchers")
		research.LoadResearchers(researcherDir, reg, providerReg)
	}

	// 4. YAML researchers from .glitch/researchers/ (project-local)
	research.LoadResearchers(".glitch/researchers", reg, providerReg)

	// Local LLM for plan/critique/score
	localLLM := func(ctx context.Context, prompt string) (string, error) {
		return provider.RunOllama("qwen2.5:7b", prompt)
	}

	// Paid LLM for drafting (uses provider system via stdin)
	var draftLLM research.LLMFn
	if _, err := providerReg.RenderCommand("claude", "test"); err == nil {
		draftLLM = func(ctx context.Context, prompt string) (string, error) {
			return providerReg.RunProvider("claude", prompt)
		}
	}

	loop := research.NewLoop(reg, localLLM)
	if draftLLM != nil {
		loop = loop.WithDraftLLM(draftLLM)
	}

	return loop
}

// resultsDir returns the results directory for a given question.
func resultsDir(question string) string {
	org, repo := research.ParseRepoFromQuestion(question)
	if org != "" && repo != "" {
		return filepath.Join("results", org, repo)
	}
	return filepath.Join("results", "general")
}

// printFeedback prints feedback to stderr.
func printFeedback(fb research.Feedback) {
	if fb.Suggestion != "" {
		fmt.Fprintf(os.Stderr, ">> learned: %s\n", fb.Suggestion)
	}
	if fb.Quality != "" {
		fmt.Fprintf(os.Stderr, ">> evidence quality: %s\n", fb.Quality)
	}
	for _, m := range fb.Missing {
		fmt.Fprintf(os.Stderr, ">> missing evidence: %s\n", m)
	}
}
```

- [ ] **Step 2: Update cmd/ask.go tier 2**

Replace the tier 2 section in `cmd/ask.go`:

```go
		// Tier 2: research loop
		if loop := buildResearchLoop(); loop != nil {
			fmt.Fprintln(os.Stderr, ">> researching...")
			ctx := context.Background()
			q := research.ResearchQuery{Question: input}
			res, err := loop.Run(ctx, q, research.DefaultBudget())
			if err == nil && res.Draft != "" {
				fmt.Println(res.Draft)

				// Print feedback
				printFeedback(res.Feedback)

				// Save results if substantive
				if research.IsSubstantive(res.Draft) {
					dir := resultsDir(input)
					if err := research.SaveResults(dir, res); err != nil {
						fmt.Fprintf(os.Stderr, ">> warning: could not save results: %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, ">> results saved to %s/\n", dir)
					}
				}

				return nil
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, ">> research error: %v\n", err)
			}
		}
```

- [ ] **Step 3: Create researchers/github-issue.yaml**

```yaml
name: github-issue
description: fetch a specific GitHub issue with full body and comments
steps:
  - id: gather
    run: gh issue view {{.input}} --json number,title,body,comments,labels
```

- [ ] **Step 4: Delete the doc-fix workflow**

```bash
rm -f ~/.config/glitch/workflows/doc-fix.yaml 2>/dev/null
rm -f /Users/stokes/Projects/stokagent/workflows/doc-fix.yaml 2>/dev/null
```

- [ ] **Step 5: Build and verify**

Run: `go build -o /tmp/glitch . && go test ./... -count=1`
Expected: All tests pass, binary builds.

- [ ] **Step 6: Commit**

```bash
git add cmd/research_helpers.go cmd/ask.go researchers/github-issue.yaml
git commit -m "feat: wire smart research loop — native researchers, paid feedback, auto-save results"
```

---

### Task 8: Clean Up + Remove Router Fix Pattern

**Files:**
- Modify: `internal/router/router.go` — remove the `fix issue` pattern and `doc-fix` routing
- Delete: `internal/pipeline/runner.go` — remove `save` step and `stepfile` template function (no longer needed)
- Delete: `internal/pipeline/types.go` — remove `Save`/`SaveStep` fields from Step

Actually, the `save` step and `stepfile` may still be useful for other workflows. Keep them — they're clean additions. Only remove the router fix pattern.

- [ ] **Step 1: Remove fix-issue routing from router.go**

Remove the `reFixIssue` regex, the `MatchFixIssue` function, and the "fix issue" fast path block in `Match()` that routes to `doc-fix`.

- [ ] **Step 2: Run tests**

Run: `go test ./internal/router/... -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/router/router.go
git commit -m "refactor: remove fix-issue router pattern — research loop handles it generically"
```

---

### Task 9: Integration Test

- [ ] **Step 1: Build**

Run: `go build -o /tmp/glitch .`

- [ ] **Step 2: Install default YAML researchers**

```bash
mkdir -p ~/.config/glitch/researchers
cp researchers/*.yaml ~/.config/glitch/researchers/
```

- [ ] **Step 3: Test with a simple question (should use git + fs)**

Run: `/tmp/glitch ask "what packages does this project have?"`
Expected: ">> researching..." → uses git + fs → prints answer about the project packages.

- [ ] **Step 4: Test with a repo question (should auto-clone if needed)**

Run: `/tmp/glitch ask "what TBC placeholders are in the observability-robots CI docs?"`
Expected: ">> researching..." → uses fs (scans for TBC) + github-issues → prints findings → saves results if substantive.

- [ ] **Step 5: Verify results folder**

```bash
ls results/
```
Expected: results directory with findings if the answer was substantive.

- [ ] **Step 6: Test missing researcher guidance**

If the planner picks a researcher that doesn't exist, verify you see:
```
>> missing researcher: <name> — add ~/.config/glitch/researchers/<name>.yaml
```

- [ ] **Step 7: Commit final state**

```bash
git add -A
git commit -m "feat: smart research loop complete — native researchers, feedback, auto-save"
```

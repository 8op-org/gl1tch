# glitch-github Plugin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a personal Go binary plugin that replaces inline GraphQL/jq plumbing in stokagent workflows with clean subcommands that output JSON.

**Architecture:** Single Go binary (`glitch-github`) with cobra subcommands. Each subcommand shells out to `gh api graphql`, parses the JSON response in Go, filters to target repos, and prints shaped JSON to stdout. Baked-in defaults for username, repos, timezone.

**Tech Stack:** Go 1.26.1, cobra, `os/exec` (shelling out to `gh`), `encoding/json`, `time`

**Spec:** `docs/superpowers/specs/2026-04-10-glitch-github-plugin-design.md`

---

## File Structure

```
~/Projects/gl1tch-github/
├── main.go              # cobra root command + subcommand wiring
├── github.go            # GraphQL queries, gh execution, JSON shaping, repo filtering
├── dateparse.go         # --since flag parsing (yesterday, week, Nd, YYYY-MM-DD)
├── dateparse_test.go    # tests for date parsing
├── github_test.go       # tests for JSON shaping and repo filtering
├── Makefile
├── .gitignore
├── go.mod
├── .goreleaser.yml
└── .github/workflows/release.yml
```

---

### Task 1: Scaffold repo and release plumbing

**Files:**
- Create: `~/Projects/gl1tch-github/go.mod`
- Create: `~/Projects/gl1tch-github/main.go`
- Create: `~/Projects/gl1tch-github/Makefile`
- Create: `~/Projects/gl1tch-github/.gitignore`
- Create: `~/Projects/gl1tch-github/.goreleaser.yml`
- Create: `~/Projects/gl1tch-github/.github/workflows/release.yml`

- [ ] **Step 1: Create repo directory and init git**

```bash
mkdir -p ~/Projects/gl1tch-github
cd ~/Projects/gl1tch-github
git init
```

- [ ] **Step 2: Create go.mod**

```
module github.com/8op-org/gl1tch-github

go 1.26.1

require github.com/spf13/cobra v1.10.2
```

Then run:
```bash
cd ~/Projects/gl1tch-github && go mod tidy
```

- [ ] **Step 3: Create main.go with root command**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "glitch-github",
	Short: "personal GitHub activity feed",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create Makefile**

```makefile
BINARY      := glitch-github
INSTALL_BIN := $(or $(shell test -w /usr/local/bin && echo /usr/local/bin),$(HOME)/.local/bin)

.PHONY: build install test clean

build:
	go build -o $(BINARY) .

install: build
	install -m 0755 $(BINARY) $(INSTALL_BIN)/$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
```

- [ ] **Step 5: Create .gitignore**

```
glitch-github
*.db
```

- [ ] **Step 6: Create .goreleaser.yml**

```yaml
version: 2

project_name: glitch-github

builds:
  - id: glitch-github
    binary: glitch-github
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ignore:
      - goos: windows
        goarch: arm64
    ldflags: [-s -w]

archives:
  - id: default
    format_overrides:
      - goos: windows
        formats: [zip]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

brews:
  - name: glitch-github
    directory: Formula
    repository:
      owner: 8op-org
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: "https://github.com/8op-org/gl1tch-github"
    description: "Personal GitHub activity feed for gl1tch"
    license: "MIT"
    install: |
      bin.install "glitch-github"
    test: |
      assert_predicate bin/"glitch-github", :exist?

changelog:
  use: git
  sort: asc
  filters:
    exclude: ["^chore", "^docs", "^ci", "^test"]
```

- [ ] **Step 7: Create .github/workflows/release.yml**

```yaml
name: Release

on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

- [ ] **Step 8: Verify build compiles**

```bash
cd ~/Projects/gl1tch-github && go build -o glitch-github . && ./glitch-github --help
```

Expected: help text showing `glitch-github` with no subcommands yet.

- [ ] **Step 9: Commit**

```bash
git add go.mod go.sum main.go Makefile .gitignore .goreleaser.yml .github/
git commit -m "feat: scaffold repo with cobra root and release plumbing"
```

---

### Task 2: Date parsing

**Files:**
- Create: `~/Projects/gl1tch-github/dateparse.go`
- Create: `~/Projects/gl1tch-github/dateparse_test.go`

- [ ] **Step 1: Write failing tests for date parsing**

```go
package main

import (
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	// Fix "now" so tests are deterministic.
	// Thursday 2026-04-09 15:00 Eastern
	loc, _ := time.LoadLocation("US/Eastern")
	now := time.Date(2026, 4, 9, 15, 0, 0, 0, loc)

	tests := []struct {
		input string
		want  string // expected YYYY-MM-DD
	}{
		{"yesterday", "2026-04-08"},
		{"week", "2026-04-06"},   // Monday of current week
		{"7d", "2026-04-02"},
		{"30d", "2026-03-10"},
		{"1d", "2026-04-08"},
		{"2026-03-15", "2026-03-15"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSince(tt.input, now)
			if err != nil {
				t.Fatalf("parseSince(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseSince(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSinceInvalid(t *testing.T) {
	loc, _ := time.LoadLocation("US/Eastern")
	now := time.Date(2026, 4, 9, 15, 0, 0, 0, loc)

	bad := []string{"", "bogus", "0d", "-3d", "2026-13-01"}
	for _, input := range bad {
		t.Run(input, func(t *testing.T) {
			_, err := parseSince(input, now)
			if err == nil {
				t.Errorf("parseSince(%q) should have failed", input)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd ~/Projects/gl1tch-github && go test -run TestParseSince -v
```

Expected: compilation error — `parseSince` not defined.

- [ ] **Step 3: Implement dateparse.go**

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const tz = "US/Eastern"

// parseSince converts a --since flag value into a YYYY-MM-DD start date.
// Accepts: "yesterday", "week", "Nd" (e.g. "7d"), or "YYYY-MM-DD".
func parseSince(s string, now time.Time) (string, error) {
	switch {
	case s == "yesterday":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil

	case s == "week":
		// Roll back to Monday.
		wd := now.Weekday()
		offset := int(wd) - int(time.Monday)
		if offset < 0 {
			offset += 7
		}
		return now.AddDate(0, 0, -offset).Format("2006-01-02"), nil

	case strings.HasSuffix(s, "d"):
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || n < 1 {
			return "", fmt.Errorf("invalid duration %q — use Nd (e.g. 7d)", s)
		}
		return now.AddDate(0, 0, -n).Format("2006-01-02"), nil

	default:
		// Try YYYY-MM-DD.
		_, err := time.Parse("2006-01-02", s)
		if err != nil {
			return "", fmt.Errorf("invalid --since %q — use yesterday, week, Nd, or YYYY-MM-DD", s)
		}
		return s, nil
	}
}

// nowEastern returns the current time in US/Eastern.
func nowEastern() time.Time {
	loc, _ := time.LoadLocation(tz)
	return time.Now().In(loc)
}

// todayEastern returns today's date as YYYY-MM-DD in US/Eastern.
func todayEastern() string {
	return nowEastern().Format("2006-01-02")
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd ~/Projects/gl1tch-github && go test -run TestParseSince -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add dateparse.go dateparse_test.go
git commit -m "feat: add --since date parsing (yesterday, week, Nd, YYYY-MM-DD)"
```

---

### Task 3: GitHub query execution and JSON shaping

**Files:**
- Create: `~/Projects/gl1tch-github/github.go`
- Create: `~/Projects/gl1tch-github/github_test.go`

- [ ] **Step 1: Write failing tests for JSON shaping and repo filtering**

These tests verify the pure functions that parse `gh` output and shape it — they don't call `gh` itself.

```go
package main

import (
	"encoding/json"
	"testing"
)

func TestFilterRepos(t *testing.T) {
	items := []repoItem{
		{Repo: "elastic/ensemble"},
		{Repo: "elastic/kibana"},
		{Repo: "elastic/oblt-cli"},
	}
	got := filterRepos(items)
	if len(got) != 2 {
		t.Fatalf("filterRepos: got %d items, want 2", len(got))
	}
	if got[0].Repo != "elastic/ensemble" {
		t.Errorf("got[0].Repo = %q, want elastic/ensemble", got[0].Repo)
	}
	if got[1].Repo != "elastic/oblt-cli" {
		t.Errorf("got[1].Repo = %q, want elastic/oblt-cli", got[1].Repo)
	}
}

func TestParseAuthoredPRs(t *testing.T) {
	raw := `{
		"data": {
			"new": {
				"nodes": [
					{
						"number": 1,
						"title": "Add feature",
						"url": "https://github.com/elastic/ensemble/pull/1",
						"repository": {"nameWithOwner": "elastic/ensemble"},
						"additions": 10,
						"deletions": 5,
						"createdAt": "2026-04-09T10:00:00Z",
						"mergedAt": null
					}
				]
			},
			"merged": {
				"nodes": [
					{
						"number": 1,
						"title": "Add feature",
						"url": "https://github.com/elastic/ensemble/pull/1",
						"repository": {"nameWithOwner": "elastic/ensemble"},
						"additions": 10,
						"deletions": 5,
						"createdAt": "2026-04-09T10:00:00Z",
						"mergedAt": "2026-04-09T12:00:00Z"
					},
					{
						"number": 2,
						"title": "Fix bug",
						"url": "https://github.com/elastic/kibana/pull/2",
						"repository": {"nameWithOwner": "elastic/kibana"},
						"additions": 3,
						"deletions": 1,
						"createdAt": "2026-04-09T11:00:00Z",
						"mergedAt": "2026-04-09T14:00:00Z"
					}
				]
			}
		}
	}`

	prs, err := parseAuthoredPRs([]byte(raw))
	if err != nil {
		t.Fatalf("parseAuthoredPRs: %v", err)
	}
	// PR #1 appears in both new and merged — should be deduplicated.
	// PR #2 is in elastic/kibana — should be filtered out.
	if len(prs) != 1 {
		t.Fatalf("got %d PRs, want 1", len(prs))
	}
	if prs[0].Number != 1 {
		t.Errorf("pr.Number = %d, want 1", prs[0].Number)
	}
	if prs[0].State != "MERGED" {
		t.Errorf("pr.State = %q, want MERGED", prs[0].State)
	}
}

func TestParseReviewsExcludesSelf(t *testing.T) {
	raw := `{
		"data": {
			"search": {
				"nodes": [
					{
						"number": 10,
						"title": "My own PR",
						"url": "https://github.com/elastic/ensemble/pull/10",
						"repository": {"nameWithOwner": "elastic/ensemble"},
						"author": {"login": "adam-stokes"},
						"reviews": {"nodes": []}
					},
					{
						"number": 11,
						"title": "Teammate PR",
						"url": "https://github.com/elastic/oblt-cli/pull/11",
						"repository": {"nameWithOwner": "elastic/oblt-cli"},
						"author": {"login": "coworker"},
						"reviews": {
							"nodes": [
								{"author": {"login": "adam-stokes"}, "state": "APPROVED", "submittedAt": "2026-04-09T10:00:00Z"},
								{"author": {"login": "someone-else"}, "state": "COMMENTED", "submittedAt": "2026-04-09T09:00:00Z"}
							]
						}
					}
				]
			}
		}
	}`

	reviews, err := parseReviews([]byte(raw), "2026-04-09")
	if err != nil {
		t.Fatalf("parseReviews: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("got %d reviews, want 1", len(reviews))
	}
	if reviews[0].Author != "coworker" {
		t.Errorf("review.Author = %q, want coworker", reviews[0].Author)
	}
	if reviews[0].ReviewState != "APPROVED" {
		t.Errorf("review.ReviewState = %q, want APPROVED", reviews[0].ReviewState)
	}
}

func TestParseClosedIssues(t *testing.T) {
	raw := `{
		"data": {
			"search": {
				"nodes": [
					{
						"number": 50,
						"title": "Fix CI",
						"url": "https://github.com/elastic/ensemble/issues/50",
						"repository": {"nameWithOwner": "elastic/ensemble"}
					}
				]
			}
		}
	}`

	issues, err := parseClosedIssues([]byte(raw))
	if err != nil {
		t.Fatalf("parseClosedIssues: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].Number != 50 {
		t.Errorf("issue.Number = %d, want 50", issues[0].Number)
	}
}

func TestActivityJSON(t *testing.T) {
	a := ActivityFeed{
		Authored: []AuthoredPR{{Number: 1, Title: "X", URL: "u", Repo: "elastic/ensemble", State: "OPEN"}},
		Reviews:  []Review{},
		Closed:   []ClosedIssue{},
	}
	data, _ := json.Marshal(a)
	var got map[string]json.RawMessage
	json.Unmarshal(data, &got)
	if _, ok := got["authored"]; !ok {
		t.Error("missing authored key")
	}
	if _, ok := got["reviews"]; !ok {
		t.Error("missing reviews key")
	}
	if _, ok := got["closed"]; !ok {
		t.Error("missing closed key")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd ~/Projects/gl1tch-github && go test -run "TestFilter|TestParse|TestActivity" -v
```

Expected: compilation error — types and functions not defined.

- [ ] **Step 3: Implement github.go**

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const username = "adam-stokes"

var targetRepos = map[string]bool{
	"elastic/observability-test-environments": true,
	"elastic/observability-robots":            true,
	"elastic/oblt-cli":                        true,
	"elastic/ensemble":                        true,
}

// --- Output types ---

type AuthoredPR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Repo      string `json:"repo"`
	State     string `json:"state"`
	Additions int    `json:"additions,omitempty"`
	Deletions int    `json:"deletions,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	MergedAt  string `json:"merged_at,omitempty"`
}

type ReviewingPR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Repo      string `json:"repo"`
	Author    string `json:"author"`
	Additions int    `json:"additions,omitempty"`
	Deletions int    `json:"deletions,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type Review struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Repo        string `json:"repo"`
	Author      string `json:"author"`
	ReviewState string `json:"review_state"`
}

type AssignedIssue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Repo      string   `json:"repo"`
	CreatedAt string   `json:"created_at,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

type ClosedIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Repo   string `json:"repo"`
}

type Mention struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Repo      string `json:"repo"`
	Type      string `json:"type"`
	UpdatedAt string `json:"updated_at"`
}

type ActivityFeed struct {
	Authored []AuthoredPR  `json:"authored"`
	Reviews  []Review      `json:"reviews"`
	Closed   []ClosedIssue `json:"closed"`
}

// --- Repo filtering ---

type repoItem struct {
	Repo string
}

func filterRepos[T any](items []T, repo func(T) string) []T {
	var out []T
	for _, item := range items {
		if targetRepos[repo(item)] {
			out = append(out, item)
		}
	}
	return out
}

// --- gh execution ---

func runGH(query string) ([]byte, error) {
	cmd := exec.Command("gh", "api", "graphql", "-f", "query="+query)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh api graphql: %w", err)
	}
	return out, nil
}

func repoSearchFilter() string {
	var parts []string
	for r := range targetRepos {
		parts = append(parts, "repo:"+r)
	}
	return strings.Join(parts, " ")
}

// --- GraphQL queries ---

func queryAuthoredPRs(since, until string) string {
	repos := repoSearchFilter()
	return fmt.Sprintf(`{
		new: search(query: "type:pr author:%s created:%s..%s %s", type: ISSUE, first: 50) {
			nodes { ... on PullRequest { number title url repository { nameWithOwner } additions deletions createdAt mergedAt } }
		}
		merged: search(query: "type:pr author:%s merged:%s..%s %s", type: ISSUE, first: 50) {
			nodes { ... on PullRequest { number title url repository { nameWithOwner } additions deletions createdAt mergedAt } }
		}
	}`, username, since, until, repos, username, since, until, repos)
}

func queryReviewingPRs() string {
	repos := repoSearchFilter()
	return fmt.Sprintf(`{
		search(query: "type:pr state:open review-requested:%s %s", type: ISSUE, first: 50) {
			nodes { ... on PullRequest { number title url repository { nameWithOwner } createdAt author { login } additions deletions } }
		}
	}`, username, repos)
}

func queryReviewsGiven(since string) string {
	repos := repoSearchFilter()
	return fmt.Sprintf(`{
		search(query: "type:pr reviewed-by:%s updated:>=%s %s", type: ISSUE, first: 50) {
			nodes { ... on PullRequest { number title url repository { nameWithOwner } author { login } reviews(last: 10) { nodes { author { login } state submittedAt } } } }
		}
	}`, username, since, repos)
}

func queryAssignedIssues() string {
	repos := repoSearchFilter()
	return fmt.Sprintf(`{
		search(query: "type:issue state:open assignee:%s %s", type: ISSUE, first: 50) {
			nodes { ... on Issue { number title url repository { nameWithOwner } createdAt labels(first: 10) { nodes { name } } } }
		}
	}`, username, repos)
}

func queryClosedIssues(since, until string) string {
	repos := repoSearchFilter()
	return fmt.Sprintf(`{
		search(query: "type:issue assignee:%s closed:%s..%s %s", type: ISSUE, first: 50) {
			nodes { ... on Issue { number title url repository { nameWithOwner } } }
		}
	}`, username, since, until, repos)
}

func queryMentions(since string) string {
	repos := repoSearchFilter()
	return fmt.Sprintf(`{
		search(query: "mentions:%s updated:>%s %s", type: ISSUE, first: 20) {
			nodes {
				... on Issue { number title url repository { nameWithOwner } updatedAt __typename }
				... on PullRequest { number title url repository { nameWithOwner } updatedAt __typename }
			}
		}
	}`, username, since, repos)
}

// --- Response parsing ---

// Raw GraphQL response shapes for json.Unmarshal.

type gqlPRNode struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	CreatedAt string `json:"createdAt"`
	MergedAt  string `json:"mergedAt"`
	State     string `json:"state"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Reviews struct {
		Nodes []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State       string `json:"state"`
			SubmittedAt string `json:"submittedAt"`
		} `json:"nodes"`
	} `json:"reviews"`
}

type gqlIssueNode struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	CreatedAt string `json:"createdAt"`
	ClosedAt  string `json:"closedAt"`
	UpdatedAt string `json:"updatedAt"`
	Typename  string `json:"__typename"`
	Labels    struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
}

func parseAuthoredPRs(data []byte) ([]AuthoredPR, error) {
	var resp struct {
		Data struct {
			New    struct{ Nodes []gqlPRNode } `json:"new"`
			Merged struct{ Nodes []gqlPRNode } `json:"merged"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	seen := map[int]bool{}
	var prs []AuthoredPR

	// Merged PRs first — if a PR appears in both, we want the MERGED state.
	for _, n := range resp.Data.Merged.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] {
			continue
		}
		seen[n.Number] = true
		prs = append(prs, AuthoredPR{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			Repo:      n.Repository.NameWithOwner,
			State:     "MERGED",
			Additions: n.Additions,
			Deletions: n.Deletions,
			CreatedAt: n.CreatedAt,
			MergedAt:  n.MergedAt,
		})
	}
	for _, n := range resp.Data.New.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] || seen[n.Number] {
			continue
		}
		prs = append(prs, AuthoredPR{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			Repo:      n.Repository.NameWithOwner,
			State:     "OPEN",
			Additions: n.Additions,
			Deletions: n.Deletions,
			CreatedAt: n.CreatedAt,
		})
	}
	return prs, nil
}

func parseReviewingPRs(data []byte) ([]ReviewingPR, error) {
	var resp struct {
		Data struct {
			Search struct{ Nodes []gqlPRNode } `json:"search"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var prs []ReviewingPR
	for _, n := range resp.Data.Search.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] {
			continue
		}
		prs = append(prs, ReviewingPR{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			Repo:      n.Repository.NameWithOwner,
			Author:    n.Author.Login,
			Additions: n.Additions,
			Deletions: n.Deletions,
			CreatedAt: n.CreatedAt,
		})
	}
	return prs, nil
}

func parseReviews(data []byte, since string) ([]Review, error) {
	var resp struct {
		Data struct {
			Search struct{ Nodes []gqlPRNode } `json:"search"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var reviews []Review
	for _, n := range resp.Data.Search.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] {
			continue
		}
		// Exclude self-authored PRs.
		if n.Author.Login == username {
			continue
		}
		// Find the latest review by us in the time range.
		var latestState string
		for _, r := range n.Reviews.Nodes {
			if r.Author.Login == username && r.SubmittedAt >= since {
				latestState = r.State
			}
		}
		if latestState == "" {
			continue
		}
		reviews = append(reviews, Review{
			Number:      n.Number,
			Title:       n.Title,
			URL:         n.URL,
			Repo:        n.Repository.NameWithOwner,
			Author:      n.Author.Login,
			ReviewState: latestState,
		})
	}
	return reviews, nil
}

func parseAssignedIssues(data []byte) ([]AssignedIssue, error) {
	var resp struct {
		Data struct {
			Search struct{ Nodes []gqlIssueNode } `json:"search"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var issues []AssignedIssue
	for _, n := range resp.Data.Search.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] {
			continue
		}
		var labels []string
		for _, l := range n.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		issues = append(issues, AssignedIssue{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			Repo:      n.Repository.NameWithOwner,
			CreatedAt: n.CreatedAt,
			Labels:    labels,
		})
	}
	return issues, nil
}

func parseClosedIssues(data []byte) ([]ClosedIssue, error) {
	var resp struct {
		Data struct {
			Search struct{ Nodes []gqlIssueNode } `json:"search"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var issues []ClosedIssue
	for _, n := range resp.Data.Search.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] {
			continue
		}
		issues = append(issues, ClosedIssue{
			Number: n.Number,
			Title:  n.Title,
			URL:    n.URL,
			Repo:   n.Repository.NameWithOwner,
		})
	}
	return issues, nil
}

func parseMentions(data []byte) ([]Mention, error) {
	var resp struct {
		Data struct {
			Search struct{ Nodes []gqlIssueNode } `json:"search"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var mentions []Mention
	for _, n := range resp.Data.Search.Nodes {
		if !targetRepos[n.Repository.NameWithOwner] {
			continue
		}
		typ := "issue"
		if n.Typename == "PullRequest" {
			typ = "pr"
		}
		mentions = append(mentions, Mention{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			Repo:      n.Repository.NameWithOwner,
			Type:      typ,
			UpdatedAt: n.UpdatedAt,
		})
	}
	return mentions, nil
}

// --- JSON output helper ---

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ensureSlice returns v if non-nil, or an empty slice so JSON encodes as []
// instead of null.
func ensureSlice[T any](v []T) []T {
	if v == nil {
		return []T{}
	}
	return v
}
```

**Note on `filterRepos`:** The generic `filterRepos` helper defined in the test is replaced by inline `targetRepos` checks in each parse function — simpler and avoids the accessor function overhead. Update the test to test repo filtering through `parseAuthoredPRs` instead (it already does this — the kibana PR gets filtered out). Remove the standalone `filterRepos` test and the `repoItem` type from the test file:

Replace the `TestFilterRepos` test with:

```go
func TestRepoFiltering(t *testing.T) {
	// Verified through TestParseAuthoredPRs — elastic/kibana PR #2 is filtered out.
	// This test verifies the targetRepos map directly.
	if !targetRepos["elastic/ensemble"] {
		t.Error("elastic/ensemble should be a target repo")
	}
	if targetRepos["elastic/kibana"] {
		t.Error("elastic/kibana should not be a target repo")
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd ~/Projects/gl1tch-github && go test -run "TestRepo|TestParse|TestActivity" -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add github.go github_test.go
git commit -m "feat: add GraphQL queries, JSON parsing, and repo filtering"
```

---

### Task 4: Wire up cobra subcommands

**Files:**
- Modify: `~/Projects/gl1tch-github/main.go`

- [ ] **Step 1: Replace main.go with full subcommand wiring**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "glitch-github",
	Short: "personal GitHub activity feed",
}

// --- prs subcommand ---

var prsCmd = &cobra.Command{
	Use:   "prs",
	Short: "fetch pull requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		authored, _ := cmd.Flags().GetBool("authored")
		reviewing, _ := cmd.Flags().GetBool("reviewing")

		if !authored && !reviewing {
			return fmt.Errorf("specify --authored or --reviewing")
		}

		if reviewing {
			data, err := runGH(queryReviewingPRs())
			if err != nil {
				return err
			}
			prs, err := parseReviewingPRs(data)
			if err != nil {
				return err
			}
			return printJSON(ensureSlice(prs))
		}

		// --authored
		since, _ := cmd.Flags().GetString("since")
		now := nowEastern()
		start, err := parseSince(since, now)
		if err != nil {
			return err
		}
		end := todayEastern()

		data, err := runGH(queryAuthoredPRs(start, end))
		if err != nil {
			return err
		}
		prs, err := parseAuthoredPRs(data)
		if err != nil {
			return err
		}
		return printJSON(ensureSlice(prs))
	},
}

// --- reviews subcommand ---

var reviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "fetch reviews you gave",
	RunE: func(cmd *cobra.Command, args []string) error {
		since, _ := cmd.Flags().GetString("since")
		now := nowEastern()
		start, err := parseSince(since, now)
		if err != nil {
			return err
		}

		data, err := runGH(queryReviewsGiven(start))
		if err != nil {
			return err
		}
		reviews, err := parseReviews(data, start)
		if err != nil {
			return err
		}
		return printJSON(ensureSlice(reviews))
	},
}

// --- issues subcommand ---

var issuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "fetch issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		assigned, _ := cmd.Flags().GetBool("assigned")
		closed, _ := cmd.Flags().GetBool("closed")

		if !assigned && !closed {
			return fmt.Errorf("specify --assigned or --closed")
		}

		if assigned {
			data, err := runGH(queryAssignedIssues())
			if err != nil {
				return err
			}
			issues, err := parseAssignedIssues(data)
			if err != nil {
				return err
			}
			return printJSON(ensureSlice(issues))
		}

		// --closed
		since, _ := cmd.Flags().GetString("since")
		now := nowEastern()
		start, err := parseSince(since, now)
		if err != nil {
			return err
		}
		end := todayEastern()

		data, err := runGH(queryClosedIssues(start, end))
		if err != nil {
			return err
		}
		issues, err := parseClosedIssues(data)
		if err != nil {
			return err
		}
		return printJSON(ensureSlice(issues))
	},
}

// --- mentions subcommand ---

var mentionsCmd = &cobra.Command{
	Use:   "mentions",
	Short: "fetch mentions",
	RunE: func(cmd *cobra.Command, args []string) error {
		since, _ := cmd.Flags().GetString("since")
		now := nowEastern()
		start, err := parseSince(since, now)
		if err != nil {
			return err
		}

		data, err := runGH(queryMentions(start))
		if err != nil {
			return err
		}
		mentions, err := parseMentions(data)
		if err != nil {
			return err
		}
		return printJSON(ensureSlice(mentions))
	},
}

// --- activity subcommand ---

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "combined dashboard feed (authored + reviews + closed)",
	RunE: func(cmd *cobra.Command, args []string) error {
		since, _ := cmd.Flags().GetString("since")
		now := nowEastern()
		start, err := parseSince(since, now)
		if err != nil {
			return err
		}
		end := todayEastern()

		// Authored PRs
		authoredData, err := runGH(queryAuthoredPRs(start, end))
		if err != nil {
			return err
		}
		authored, err := parseAuthoredPRs(authoredData)
		if err != nil {
			return err
		}

		// Reviews given
		reviewsData, err := runGH(queryReviewsGiven(start))
		if err != nil {
			return err
		}
		reviews, err := parseReviews(reviewsData, start)
		if err != nil {
			return err
		}

		// Closed issues
		closedData, err := runGH(queryClosedIssues(start, end))
		if err != nil {
			return err
		}
		closed, err := parseClosedIssues(closedData)
		if err != nil {
			return err
		}

		feed := ActivityFeed{
			Authored: ensureSlice(authored),
			Reviews:  ensureSlice(reviews),
			Closed:   ensureSlice(closed),
		}
		return printJSON(feed)
	},
}

func init() {
	prsCmd.Flags().Bool("authored", false, "PRs you authored")
	prsCmd.Flags().Bool("reviewing", false, "open PRs requesting your review")
	prsCmd.Flags().String("since", "yesterday", "time range: yesterday, week, Nd, YYYY-MM-DD")

	reviewsCmd.Flags().String("since", "yesterday", "time range: yesterday, week, Nd, YYYY-MM-DD")

	issuesCmd.Flags().Bool("assigned", false, "open issues assigned to you")
	issuesCmd.Flags().Bool("closed", false, "issues you closed")
	issuesCmd.Flags().String("since", "yesterday", "time range: yesterday, week, Nd, YYYY-MM-DD")

	mentionsCmd.Flags().String("since", "7d", "time range: yesterday, week, Nd, YYYY-MM-DD")

	activityCmd.Flags().String("since", "yesterday", "time range: yesterday, week, Nd, YYYY-MM-DD")

	rootCmd.AddCommand(prsCmd, reviewsCmd, issuesCmd, mentionsCmd, activityCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify build and help output**

```bash
cd ~/Projects/gl1tch-github && go build -o glitch-github . && ./glitch-github --help
```

Expected: help showing `prs`, `reviews`, `issues`, `mentions`, `activity` subcommands.

```bash
./glitch-github prs --help
```

Expected: help showing `--authored`, `--reviewing`, `--since` flags.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: wire up all cobra subcommands"
```

---

### Task 5: Smoke test with live GitHub data

**Files:** none — manual verification only

- [ ] **Step 1: Build and install**

```bash
cd ~/Projects/gl1tch-github && make install
```

- [ ] **Step 2: Verify plugin discovery**

```bash
glitch plugin list
```

Expected: `github` appears in the list.

- [ ] **Step 3: Test each subcommand**

```bash
glitch-github prs --authored --since week
glitch-github prs --reviewing
glitch-github reviews --since week
glitch-github issues --assigned
glitch-github issues --closed --since week
glitch-github mentions --since 7d
glitch-github activity --since yesterday
```

For each: verify valid JSON output, no errors, repos are all in the target set.

- [ ] **Step 4: Test through glitch plugin interface**

```bash
glitch github activity --since yesterday
```

Expected: same output as calling `glitch-github` directly.

- [ ] **Step 5: Test error cases**

```bash
glitch-github prs              # should error: specify --authored or --reviewing
glitch-github issues           # should error: specify --assigned or --closed
glitch-github activity --since bogus  # should error: invalid --since
```

- [ ] **Step 6: Compare activity output against existing workflow**

```bash
cd ~/Projects/stokagent && glitch workflow run dashboard-activity > /tmp/old.json
glitch-github activity --since yesterday > /tmp/new.json
diff <(jq -S . /tmp/old.json) <(jq -S . /tmp/new.json)
```

Verify the shapes match. Field names may differ slightly (`review_state` vs `reviewState`) — the important thing is the same PRs/reviews/issues appear.

---

### Task 6: Push repo and cut first release

- [ ] **Step 1: Create GitHub repo**

```bash
cd ~/Projects/gl1tch-github
gh repo create 8op-org/gl1tch-github --public \
  --description "Personal GitHub activity feed plugin for gl1tch" \
  --source=. --remote=origin --push
```

- [ ] **Step 2: Set tap secret**

```bash
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo 8op-org/gl1tch-github
```

Paste the PAT when prompted (same token used by other gl1tch plugin repos).

- [ ] **Step 3: Tag and release**

```bash
git tag v0.1.0
git push origin v0.1.0
```

- [ ] **Step 4: Watch release workflow**

```bash
gh run list --repo 8op-org/gl1tch-github
gh run watch --repo 8op-org/gl1tch-github
```

Expected: GoReleaser builds binaries, creates release, updates `8op-org/homebrew-tap/Formula/glitch-github.rb`.

- [ ] **Step 5: Verify Homebrew install**

```bash
brew update
brew install 8op-org/tap/glitch-github
glitch-github activity --since yesterday
```

---

### Task 7: Update stokagent workflows to use the plugin

**Files:**
- Modify: `~/Projects/stokagent/workflows/dashboard-activity.yaml`
- Modify: `~/Projects/stokagent/.glitch/workflows/activity-report.yaml`
- Modify: `~/Projects/stokagent/.glitch/workflows/review-dashboard.yaml`
- Modify: `~/Projects/stokagent/.glitch/workflows/morning-briefing.yaml`

- [ ] **Step 1: Rewrite dashboard-activity.yaml**

```yaml
name: dashboard-activity
description: JSON feed — yesterday's authored PRs, reviews given, closed issues
steps:
  - id: fetch
    run: glitch-github activity --since yesterday
```

- [ ] **Step 2: Run the rewritten workflow and compare**

```bash
cd ~/Projects/stokagent && glitch workflow run dashboard-activity > /tmp/new-dashboard.json
jq . /tmp/new-dashboard.json
```

Verify it produces the same `{"authored":[], "reviews":[], "closed":[]}` shape.

- [ ] **Step 3: Rewrite review-dashboard.yaml**

```yaml
name: review-dashboard
description: pending reviews, assigned issues, and mentions across elastic repos
steps:
  - id: fetch-reviews
    run: glitch-github prs --reviewing

  - id: fetch-issues
    run: glitch-github issues --assigned

  - id: fetch-mentions
    run: glitch-github mentions --since 7d

  - id: format
    llm:
      prompt: |
        You are a concise report formatter. Given three JSON arrays of GitHub data, produce a markdown review dashboard.

        PRs awaiting review:
        {{step "fetch-reviews"}}

        Assigned issues:
        {{step "fetch-issues"}}

        Recent mentions:
        {{step "fetch-mentions"}}

        Format rules:
        - Group by repository, oldest first within each group
        - PRs: `- #NUMBER — Title (AGE days, +ADDS/-DELS, by @AUTHOR) URL`
        - Issues: `- #NUMBER — Title (AGE days) [labels] URL`
        - Mentions: `- repo#NUMBER — Title (updated DATE) URL`
        - Calculate age in days from createdAt to today
        - End with a summary line: N PRs, N issues, N mentions
        - No emoji, no tables, bullet lists only
        - If a section is empty say "None"
```

- [ ] **Step 4: Rewrite morning-briefing.yaml**

```yaml
name: morning-briefing
description: combined review dashboard + yesterday's activity in one briefing
steps:
  - id: fetch-reviews
    run: glitch-github prs --reviewing

  - id: fetch-issues
    run: glitch-github issues --assigned

  - id: fetch-mentions
    run: glitch-github mentions --since 7d

  - id: fetch-yesterday
    run: glitch-github activity --since yesterday

  - id: format
    llm:
      prompt: |
        Produce a morning briefing with two sections.

        ## What needs attention

        PRs awaiting review:
        {{step "fetch-reviews"}}

        Assigned issues:
        {{step "fetch-issues"}}

        Recent mentions:
        {{step "fetch-mentions"}}

        ## What I did yesterday

        {{step "fetch-yesterday"}}

        Format rules:
        - Group "needs attention" items by repo, oldest first
        - PRs: `- #NUMBER — Title (AGE days, +ADDS/-DELS, by @AUTHOR) URL`
        - Issues: `- #NUMBER — Title (AGE days) [labels] URL`
        - Yesterday's activity as a flat bullet list: `- **[Status]** [Title](URL) - short description`
        - Write like telling a coworker what happened, one short sentence per item
        - End with a one-line summary of totals
        - No emoji, no tables, bullet lists only
        - If a section is empty say "None"
```

- [ ] **Step 5: Rewrite activity-report.yaml**

```yaml
name: activity-report
description: weekly or daily contribution report — authored PRs, reviews, closed issues
steps:
  - id: date-range
    run: |
      INPUT="{{.input}}"
      if echo "$INPUT" | grep -qi "daily\|standup\|yesterday"; then
        echo "yesterday"
      else
        echo "week"
      fi

  - id: fetch-authored
    run: glitch-github prs --authored --since {{step "date-range"}}

  - id: fetch-reviews
    run: glitch-github reviews --since {{step "date-range"}}

  - id: fetch-closed
    run: glitch-github issues --closed --since {{step "date-range"}}

  - id: format
    llm:
      prompt: |
        Generate a contribution report as a flat markdown bullet list under "Adam:".

        Date range: {{step "date-range"}}

        Authored PRs:
        {{step "fetch-authored"}}

        Reviews given (excluding self-authored):
        {{step "fetch-reviews"}}

        Closed issues:
        {{step "fetch-closed"}}

        Format each item as:
        - **[Status]** [Title](URL) - One short sentence describing what changed

        Status tags: [Merged], [Open], [Approved], [Commented], [Changes Requested], [Closed]

        For reviews, map the review state: APPROVED→[Approved], CHANGES_REQUESTED→[Changes Requested], COMMENTED or DISMISSED→[Commented]. Use the reviewer's latest review state for adam-stokes.

        Writing rules:
        - Write like telling a coworker what you did
        - One short specific sentence, no -ing phrases tacked on
        - No promotional language (streamlined, enhanced, robust)
        - For reviews, describe what the PR does, not "reviewed X's PR"
        - No emoji, no tables, no summary stats
```

- [ ] **Step 6: Test each rewritten workflow**

```bash
cd ~/Projects/stokagent
glitch workflow run review-dashboard
glitch workflow run morning-briefing
glitch workflow run activity-report "daily standup"
```

Verify each produces reasonable output.

- [ ] **Step 7: Commit workflow updates**

```bash
cd ~/Projects/stokagent
git add workflows/dashboard-activity.yaml .glitch/workflows/activity-report.yaml .glitch/workflows/review-dashboard.yaml .glitch/workflows/morning-briefing.yaml
git commit -m "refactor: replace inline GraphQL with glitch-github plugin calls"
```

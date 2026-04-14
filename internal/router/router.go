package router

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
)

const defaultOrg = "elastic"

var reIssueRef = regexp.MustCompile(`^(?:(?:([a-zA-Z0-9_.-]+)/)?([a-zA-Z0-9_.-]+)#)?(\d+)$`)
var reWorkOnIssue = regexp.MustCompile(`(?i)work on issue\s+(.+)`)

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
		return "", issue, true
	}
	if owner == "" {
		return defaultOrg + "/" + repoName, issue, true
	}
	return owner + "/" + repoName, issue, true
}

// MatchWorkOnIssue checks if input matches "work on issue <ref>".
func MatchWorkOnIssue(input string) (ref string, ok bool) {
	m := reWorkOnIssue.FindStringSubmatch(input)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

// ResolveRepo returns the repo as-is if non-empty, otherwise infers from git remote.
func ResolveRepo(repo string) (string, error) {
	if repo != "" {
		return repo, nil
	}
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("no repo specified and git remote failed: %w", err)
	}
	remote := strings.TrimSpace(string(out))
	remote = strings.TrimSuffix(remote, ".git")
	// Handle SSH format: git@github.com:owner/repo
	if i := strings.Index(remote, "github.com:"); i >= 0 {
		path := remote[i+len("github.com:"):]
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1], nil
		}
	}
	// Handle HTTPS format: https://github.com/owner/repo
	if i := strings.Index(remote, "github.com/"); i >= 0 {
		path := remote[i+len("github.com/"):]
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1], nil
		}
	}
	return "", fmt.Errorf("could not parse owner/repo from remote: %s", remote)
}

// ParseMultiIssue parses space-separated issue refs.
// Returns the list of issue numbers and a repo (from the first qualified ref, or empty).
// Returns ok=false if the input doesn't look like issue refs at all.
func ParseMultiIssue(input string) (issues []string, repo string, ok bool) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, "", false
	}
	for _, part := range parts {
		r, issue, matched := ParseIssueRef(part)
		if !matched {
			return nil, "", false
		}
		issues = append(issues, issue)
		if r != "" && repo == "" {
			repo = r
		}
	}
	return issues, repo, len(issues) > 0
}

var reGitHubPR = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/pull/\d+`)
var reGitHubIssueURL = regexp.MustCompile(`https?://github\.com/([^/]+)/([^/]+)/issues/(\d+)`)

// ParseIssueURL extracts org, repo, and issue number from a GitHub issue URL.
func ParseIssueURL(input string) (repo, issue string, ok bool) {
	m := reGitHubIssueURL.FindStringSubmatch(input)
	if m == nil {
		return "", "", false
	}
	return m[1] + "/" + m[2], m[3], true
}

// Match picks the best workflow for the user's input.
// It tries fast URL-based matching first, then falls back to Ollama.
func Match(input string, workflows map[string]*pipeline.Workflow, model string) (*pipeline.Workflow, string, map[string]string) {
	// Fast path: work on issue <ref>
	if ref, ok := MatchWorkOnIssue(input); ok {
		if w, ok := workflows["work-on-issue"]; ok {
			repo, issue, ok := ParseIssueRef(ref)
			if ok {
				resolved, err := ResolveRepo(repo)
				if err == nil {
					return w, input, map[string]string{
						"repo":  resolved,
						"issue": issue,
					}
				}
			}
		}
	}

	// Fast path: detect GitHub URLs.
	if url := reGitHubPR.FindString(input); url != "" {
		if w, ok := workflows["github-pr-review"]; ok {
			return w, url, nil
		}
	}

	// Fast path: GitHub issue URLs → issue-to-pr workflow.
	if repo, issue, ok := ParseIssueURL(input); ok {
		if w, wOK := workflows["issue-to-pr"]; wOK {
			return w, input, map[string]string{
				"repo":  repo,
				"issue": issue,
			}
		}
	}

	// Fast path: bare issue ref (e.g. "3339", "observability-robots#3339", "elastic/observability-robots#3339")
	if repo, issue, ok := ParseIssueRef(input); ok {
		if w, wOK := workflows["issue-to-pr"]; wOK {
			resolved, err := ResolveRepo(repo)
			if err == nil {
				return w, input, map[string]string{
					"repo":  resolved,
					"issue": issue,
				}
			}
		}
	}

	// Build a numbered menu for the LLM.
	var menu []string
	var keys []string
	i := 1
	for name, w := range workflows {
		desc := w.Description
		if desc == "" {
			desc = name
		}
		menu = append(menu, fmt.Sprintf("%d. %s — %s", i, name, strings.TrimSpace(desc)))
		keys = append(keys, name)
		i++
	}

	prompt := fmt.Sprintf(`Given these available workflows:
%s
0. none — the question does not match any workflow above

The user asked: %q

Which workflow number best matches? Reply with ONLY the number, nothing else. Reply 0 if the question is general or does not match a specific workflow.`, strings.Join(menu, "\n"), input)

	if model == "" {
		model = "qwen2.5:7b"
	}
	out, err := provider.RunOllama(model, prompt)
	if err != nil {
		return nil, input, nil
	}

	// Parse the number from the response.
	out = strings.TrimSpace(out)
	// Extract first number found in response.
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(out)
	if numStr == "" {
		return nil, input, nil
	}
	n, err := strconv.Atoi(numStr)
	if err != nil || n < 1 || n > len(keys) {
		return nil, input, nil // 0 or out of range = no match, fall through to research loop
	}

	name := keys[n-1]
	return workflows[name], input, nil
}

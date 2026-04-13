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
	reAnyLink = regexp.MustCompile(`https?://github\.com/[^\s\)>\]]+`)
)

// Adapt detects the input type and returns a normalized ResearchDocument.
func Adapt(input string) (ResearchDocument, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return ResearchDocument{}, fmt.Errorf("adapt: empty input")
	}

	if m := reGHIssue.FindStringSubmatch(input); m != nil {
		return adaptGitHubIssue(m[1], m[2], m[3], m[0])
	}
	if m := reGHPR.FindStringSubmatch(input); m != nil {
		return adaptGitHubPR(m[1], m[2], m[3], m[0])
	}
	if m := reGDoc.FindStringSubmatch(input); m != nil {
		return adaptGoogleDoc(m[1], m[0])
	}

	// Fallback: plain text input.
	return ResearchDocument{
		Source: "text",
		Title:  firstLine(input),
		Body:   input,
		Links:  extractLinks(input),
	}, nil
}

// adaptGitHubIssue fetches a GitHub issue via `gh` and builds a ResearchDocument.
func adaptGitHubIssue(org, repo, number, url string) (ResearchDocument, error) {
	out, err := exec.Command("gh", "issue", "view", url,
		"--json", "number,title,body,comments,labels,assignees").Output()
	if err != nil {
		return ResearchDocument{}, fmt.Errorf("adapt: gh issue view: %w", err)
	}

	var raw struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		Comments  []struct {
			Body string `json:"body"`
		} `json:"comments"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Assignees []struct {
			Login string `json:"login"`
		} `json:"assignees"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return ResearchDocument{}, fmt.Errorf("adapt: parse issue json: %w", err)
	}

	var body strings.Builder
	body.WriteString(raw.Body)
	for i, c := range raw.Comments {
		fmt.Fprintf(&body, "\n\n--- comment %d ---\n%s", i+1, c.Body)
	}

	meta := make(map[string]string)
	meta["number"] = fmt.Sprintf("%d", raw.Number)
	if len(raw.Labels) > 0 {
		names := make([]string, len(raw.Labels))
		for i, l := range raw.Labels {
			names[i] = l.Name
		}
		meta["labels"] = strings.Join(names, ",")
	}
	if len(raw.Assignees) > 0 {
		logins := make([]string, len(raw.Assignees))
		for i, a := range raw.Assignees {
			logins[i] = a.Login
		}
		meta["assignees"] = strings.Join(logins, ",")
	}

	fullBody := body.String()
	repoSlug := org + "/" + repo
	repoPath, _ := EnsureRepo(org, repo, "")

	return ResearchDocument{
		Source:    "github_issue",
		SourceURL: url,
		Title:     raw.Title,
		Body:      fullBody,
		Repo:      repoSlug,
		RepoPath:  repoPath,
		Metadata:  meta,
		Links:     extractLinks(fullBody),
	}, nil
}

// adaptGitHubPR fetches a GitHub PR via `gh` and builds a ResearchDocument.
func adaptGitHubPR(org, repo, number, url string) (ResearchDocument, error) {
	out, err := exec.Command("gh", "pr", "view", url,
		"--json", "number,title,body,comments,files,additions,deletions,reviews,state").Output()
	if err != nil {
		return ResearchDocument{}, fmt.Errorf("adapt: gh pr view: %w", err)
	}

	var raw struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		State     string `json:"state"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Comments  []struct {
			Body string `json:"body"`
		} `json:"comments"`
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
		Reviews []struct {
			Body  string `json:"body"`
			State string `json:"state"`
		} `json:"reviews"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return ResearchDocument{}, fmt.Errorf("adapt: parse pr json: %w", err)
	}

	var body strings.Builder
	body.WriteString(raw.Body)
	for i, c := range raw.Comments {
		fmt.Fprintf(&body, "\n\n--- comment %d ---\n%s", i+1, c.Body)
	}
	for i, r := range raw.Reviews {
		if r.Body != "" {
			fmt.Fprintf(&body, "\n\n--- review %d (%s) ---\n%s", i+1, r.State, r.Body)
		}
	}

	meta := make(map[string]string)
	meta["number"] = fmt.Sprintf("%d", raw.Number)
	meta["state"] = raw.State
	meta["additions"] = fmt.Sprintf("%d", raw.Additions)
	meta["deletions"] = fmt.Sprintf("%d", raw.Deletions)
	if len(raw.Files) > 0 {
		paths := make([]string, len(raw.Files))
		for i, f := range raw.Files {
			paths[i] = f.Path
		}
		meta["files"] = strings.Join(paths, ",")
	}

	fullBody := body.String()
	repoSlug := org + "/" + repo
	repoPath, _ := EnsureRepo(org, repo, "")

	return ResearchDocument{
		Source:    "github_pr",
		SourceURL: url,
		Title:     raw.Title,
		Body:      fullBody,
		Repo:      repoSlug,
		RepoPath:  repoPath,
		Metadata:  meta,
		Links:     extractLinks(fullBody),
	}, nil
}

// adaptGoogleDoc fetches a Google Doc via `gws` and builds a ResearchDocument.
func adaptGoogleDoc(docID, url string) (ResearchDocument, error) {
	out, err := exec.Command("gws", "docs", "get", docID).Output()
	if err != nil {
		return ResearchDocument{}, fmt.Errorf("adapt: gws docs get: %w", err)
	}

	text := strings.TrimSpace(string(out))
	title := firstLine(text)

	return ResearchDocument{
		Source:    "google_doc",
		SourceURL: url,
		Title:     title,
		Body:      text,
		Links:     extractLinks(text),
	}, nil
}

// extractLinks finds all GitHub URLs in text and returns them as Links.
func extractLinks(text string) []Link {
	matches := reAnyLink.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var links []Link
	for _, u := range matches {
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		links = append(links, Link{URL: u, Label: labelForURL(u)})
	}
	return links
}

// labelForURL generates a short label from a GitHub URL.
func labelForURL(u string) string {
	if m := reGHIssue.FindStringSubmatch(u); m != nil {
		return fmt.Sprintf("%s/%s#%s", m[1], m[2], m[3])
	}
	if m := reGHPR.FindStringSubmatch(u); m != nil {
		return fmt.Sprintf("%s/%s#%s (PR)", m[1], m[2], m[3])
	}
	return u
}

// firstLine returns the first non-empty line of text.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return s
}

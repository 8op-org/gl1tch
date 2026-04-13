package research

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitResearcher is a native researcher that queries git history, diffs, and content.
type GitResearcher struct {
	RepoPath string
}

func (g *GitResearcher) Name() string { return "git" }
func (g *GitResearcher) Describe() string {
	return "git history, diffs, remotes, and blame for the current or target repository"
}

func (g *GitResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	var sections []string

	if out := g.run("log", "--oneline", "-30"); out != "" {
		sections = append(sections, "=== Recent Commits ===\n"+out)
	}
	if out := g.run("remote", "-v"); out != "" {
		sections = append(sections, "=== Remotes ===\n"+out)
	}

	question := strings.ToLower(q.Question)
	if containsAny(question, "change", "diff", "modif", "broke", "break", "fix") {
		if out := g.run("diff", "--stat", "HEAD~10"); out != "" {
			sections = append(sections, "=== Recent Diff Stats ===\n"+out)
		}
	}

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

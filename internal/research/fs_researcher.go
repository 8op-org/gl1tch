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
	RootPath string
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

	// Read specific file paths mentioned in question or prior evidence
	paths := extractFilePaths(q.Question)
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
		content := string(data)
		if len(content) > 5000 {
			content = content[:5000] + "\n... (truncated)"
		}
		sections = append(sections, fmt.Sprintf("=== FILE: %s ===\n%s\n=== END: %s ===", p, content, p))
	}

	// Scan for placeholders
	if containsAny(question, "tbc", "tbd", "todo", "fixme", "placeholder", "missing", "incomplete") {
		if out := grepDir(root, `TBC\|TBD\|TODO\|FIXME`, "*.md"); out != "" {
			sections = append(sections, "=== Placeholder Scan ===\n"+out)
		}
	}

	// Directory structure
	if containsAny(question, "structure", "layout", "tree", "directory", "what's in", "files in", "project") {
		if out := listTree(root, 3); out != "" {
			sections = append(sections, "=== Directory Tree ===\n"+out)
		}
	}

	// Fallback: keyword search
	if len(sections) == 0 {
		kws := extractKeywords(question)
		for _, kw := range kws {
			if out := grepDir(root, kw, "*.md"); out != "" {
				sections = append(sections, fmt.Sprintf("=== Files matching %q ===\n%s", kw, out))
			}
		}
	}

	// Last resort: tree
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
	s = strings.ReplaceAll(s, dir+"/", "")
	lines := strings.Split(s, "\n")
	if len(lines) > 100 {
		s = strings.Join(lines[:100], "\n") + fmt.Sprintf("\n... (%d more files)", len(lines)-100)
	}
	return s
}

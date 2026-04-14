package batch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest holds the result of scanning all iterations/variants for one issue.
type Manifest struct {
	Issue         string
	BestVariant   string
	BestIteration int
	BestScore     int
	BestTotal     int
	Scores        []IterationScores
}

// IterationScores holds per-variant scores for one iteration.
type IterationScores struct {
	Iteration int
	Variants  map[string]Score
}

// Score holds parsed PASS/FAIL counts from a review.
type Score struct {
	Passed int
	Total  int
	Pass   bool
}

// ParseReview reads a review.md and counts PASS/FAIL lines.
func ParseReview(path string) (Score, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Score{}, err
	}
	upper := strings.ToUpper(strings.ReplaceAll(string(data), "*", ""))
	var passed, total int
	for _, line := range strings.Split(upper, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "OVERALL") {
			continue
		}
		hasPass := strings.Contains(line, "PASS")
		hasFail := strings.Contains(line, "FAIL")
		if hasPass && !hasFail {
			passed++
			total++
		} else if hasFail {
			total++
		}
	}
	overallPass := strings.Contains(upper, "OVERALL: PASS") || strings.Contains(upper, "OVERALL PASS")
	return Score{Passed: passed, Total: total, Pass: overallPass}, nil
}

// GenerateManifest scans results and picks the best variant/iteration.
func GenerateManifest(issueDir, issue string, variants []string, iterations int) (*Manifest, error) {
	m := &Manifest{Issue: issue}

	for iter := 1; iter <= iterations; iter++ {
		is := IterationScores{Iteration: iter, Variants: make(map[string]Score)}
		for _, variant := range variants {
			reviewPath := filepath.Join(issueDir, fmt.Sprintf("iteration-%d", iter), variant, "review.md")
			score, err := ParseReview(reviewPath)
			if err != nil {
				continue
			}
			is.Variants[variant] = score
			if score.Passed > m.BestScore || (score.Passed == m.BestScore && score.Total < m.BestTotal) {
				m.BestScore = score.Passed
				m.BestTotal = score.Total
				m.BestVariant = variant
				m.BestIteration = iter
			}
		}
		m.Scores = append(m.Scores, is)
	}
	return m, nil
}

// WriteManifest writes manifest.md to the issue results directory.
func WriteManifest(issueDir string, m *Manifest, variants []string, iterations int) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# Issue #%s — Results Manifest\n\n", m.Issue)

	b.WriteString("## Confidence Scores\n\n")
	b.WriteString("| Iteration |")
	for _, v := range variants {
		fmt.Fprintf(&b, " %s |", v)
	}
	b.WriteString("\n|-----------|")
	for range variants {
		b.WriteString("-------|")
	}
	b.WriteString("\n")

	for _, is := range m.Scores {
		fmt.Fprintf(&b, "| %d |", is.Iteration)
		for _, v := range variants {
			if s, ok := is.Variants[v]; ok && s.Total > 0 {
				fmt.Fprintf(&b, " %d/%d |", s.Passed, s.Total)
			} else {
				b.WriteString(" - |")
			}
		}
		b.WriteString("\n")
	}

	if m.BestVariant != "" {
		fmt.Fprintf(&b, "\n## Best Result\n\n")
		fmt.Fprintf(&b, "**Winner:** %s (iteration %d, %d/%d)\n\n", m.BestVariant, m.BestIteration, m.BestScore, m.BestTotal)

		bestDir := filepath.Join(issueDir, fmt.Sprintf("iteration-%d", m.BestIteration), m.BestVariant)

		b.WriteString("### PR Title\n\n")
		if data, err := os.ReadFile(filepath.Join(bestDir, "pr-title.txt")); err == nil {
			b.WriteString(strings.TrimSpace(string(data)))
		}

		b.WriteString("\n\n### PR Body\n\n")
		if data, err := os.ReadFile(filepath.Join(bestDir, "pr-body.md")); err == nil {
			b.WriteString(strings.TrimSpace(string(data)))
		}

		fmt.Fprintf(&b, "\n\n### Implementation Plan\n\nSee: `%s/plan.md`\n", bestDir)

		b.WriteString("\n\n### Next Steps\n\n")
		if data, err := os.ReadFile(filepath.Join(bestDir, "next-steps.md")); err == nil {
			b.WriteString(strings.TrimSpace(string(data)))
		}
	}

	b.WriteString("\n\n## Cross-Review Summary\n\n")
	for iter := 1; iter <= iterations; iter++ {
		crPath := filepath.Join(issueDir, fmt.Sprintf("iteration-%d", iter), "cross-review.md")
		if data, err := os.ReadFile(crPath); err == nil {
			fmt.Fprintf(&b, "### Iteration %d\n\n%s\n\n", iter, strings.TrimSpace(string(data)))
		}
	}

	return os.WriteFile(filepath.Join(issueDir, "manifest.md"), []byte(b.String()), 0o644)
}

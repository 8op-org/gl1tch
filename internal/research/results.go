package research

import (
	"encoding/json"
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

// resultDir computes the output directory path from the document metadata.
// Pattern: baseDir/<org>/<repo>/<number>  — falls back to "general" for missing parts.
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

// SaveLoopResult saves a LoopResult to a structured directory layout:
//
//	results/<org>/<repo>/<number>/
//	├── summary.md          # when goal=summarize
//	├── evidence/           # raw tool call outputs
//	│   ├── 001-grep_code.txt
//	│   ├── 002-read_file.txt
//	├── run.json            # run metadata
//	└── implementation/     # when goal=implement
//	    └── plan.md
func SaveLoopResult(baseDir string, result LoopResult) error {
	dir := resultDir(baseDir, result)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("results: mkdir: %w", err)
	}

	// Write run.json
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
	raw, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("results: marshal run.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.json"), raw, 0o644); err != nil {
		return fmt.Errorf("results: write run.json: %w", err)
	}

	// Write evidence files
	if len(result.ToolCalls) > 0 {
		evidenceDir := filepath.Join(dir, "evidence")
		if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
			return fmt.Errorf("results: mkdir evidence: %w", err)
		}
		for i, tc := range result.ToolCalls {
			name := fmt.Sprintf("%03d-%s.txt", i+1, tc.Tool)
			content := tc.Output
			if tc.Err != "" {
				content = "ERROR: " + tc.Err
			}
			if err := os.WriteFile(filepath.Join(evidenceDir, name), []byte(content), 0o644); err != nil {
				return fmt.Errorf("results: write evidence %s: %w", name, err)
			}
		}
	}

	// Write output based on goal
	switch result.Goal {
	case GoalImplement:
		implDir := filepath.Join(dir, "implementation")
		if err := os.MkdirAll(implDir, 0o755); err != nil {
			return fmt.Errorf("results: mkdir implementation: %w", err)
		}
		if err := os.WriteFile(filepath.Join(implDir, "plan.md"), []byte(result.Output), 0o644); err != nil {
			return fmt.Errorf("results: write plan.md: %w", err)
		}
	default:
		if err := os.WriteFile(filepath.Join(dir, "summary.md"), []byte(result.Output), 0o644); err != nil {
			return fmt.Errorf("results: write summary.md: %w", err)
		}
	}

	if err := writeReadme(dir, result); err != nil {
		return fmt.Errorf("results: write README.md: %w", err)
	}

	return nil
}

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

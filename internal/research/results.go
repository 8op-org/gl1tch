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

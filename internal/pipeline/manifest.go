package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ManifestThreshold is the maximum output size (in bytes) that will be stored
// inline. Outputs larger than this are written to a file.
const ManifestThreshold = 4096

// Section describes a named region of an output file.
type Section struct {
	Name      string
	StartLine int // 1-based
	EndLine   int // 1-based, inclusive
	Summary   string
}

// OutputManifest describes the disposition of a step's output.
type OutputManifest struct {
	Path     string    // empty if inline
	Inline   string    // non-empty if output fits below threshold
	Size     int64
	Lines    int
	Format   string    // "text", "json", "yaml", "markdown"
	Sections []Section
}

// BuildManifest creates an OutputManifest for output.
// If len(output) <= ManifestThreshold: returns inline manifest, no file written.
// Otherwise: writes content to filepath.Join(dir, id+".out"), detects format,
// generates sections, returns manifest with Path set.
func BuildManifest(output, id, dir string) (OutputManifest, error) {
	if len(output) <= ManifestThreshold {
		return OutputManifest{
			Inline: output,
			Size:   int64(len(output)),
			Lines:  manifestCountLines(output),
			Format: detectFormat(output),
		}, nil
	}

	// Write to file.
	path := filepath.Join(dir, id+".out")
	if err := os.WriteFile(path, []byte(output), 0o644); err != nil {
		return OutputManifest{}, fmt.Errorf("manifest: write output file: %w", err)
	}

	lines := manifestCountLines(output)
	format := detectFormat(output)
	sections := buildSections(output, format, lines)

	return OutputManifest{
		Path:     path,
		Size:     int64(len(output)),
		Lines:    lines,
		Format:   format,
		Sections: sections,
	}, nil
}

// Summary returns a compact string (<= 1024 bytes) describing the manifest,
// suitable for inclusion in an agent prompt.
// For inline manifests, returns the inline content.
// For file manifests, returns path + section index.
func (m OutputManifest) Summary() string {
	if m.Inline != "" {
		if len(m.Inline) > 1024 {
			return m.Inline[:1024]
		}
		return m.Inline
	}

	var sb strings.Builder
	sb.WriteString(m.Path)
	fmt.Fprintf(&sb, " (%s, %d lines)\n", m.Format, m.Lines)
	for _, s := range m.Sections {
		line := fmt.Sprintf("  [%d-%d] %s\n", s.StartLine, s.EndLine, s.Name)
		if sb.Len()+len(line) > 1024 {
			break
		}
		sb.WriteString(line)
	}

	result := sb.String()
	if len(result) > 1024 {
		return result[:1024]
	}
	return result
}

// manifestCountLines counts the number of lines in a string.
func manifestCountLines(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	// If the string doesn't end with a newline, count the last partial line.
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

// detectFormat inspects the first non-whitespace content to determine the format.
func detectFormat(output string) string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return "text"
	}

	// JSON: starts with { or [
	if trimmed[0] == '{' || trimmed[0] == '[' {
		return "json"
	}

	// Markdown: contains ## or ### headings
	if strings.Contains(output, "## ") || strings.Contains(output, "### ") {
		return "markdown"
	}

	// YAML: starts with --- or has key: value structure on first non-empty line
	if strings.HasPrefix(trimmed, "---") {
		return "yaml"
	}
	// Check for YAML-like key: structure on first line
	firstLine := strings.SplitN(trimmed, "\n", 2)[0]
	if isYAMLLine(firstLine) {
		return "yaml"
	}

	return "text"
}

// isYAMLLine returns true if the line looks like a YAML key: value pair.
func isYAMLLine(line string) bool {
	// Must contain a colon not at the start and have something before it
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return false
	}
	key := strings.TrimSpace(line[:idx])
	// Key must be a simple identifier (no spaces, no special chars at start)
	if key == "" || strings.ContainsAny(key, " \t{[") {
		return false
	}
	return true
}

// buildSections generates sections based on the detected format.
func buildSections(output, format string, totalLines int) []Section {
	switch format {
	case "markdown":
		return buildMarkdownSections(output, totalLines)
	case "json":
		return buildJSONSections(output, totalLines)
	default:
		return buildTextChunks(output, totalLines)
	}
}

// buildMarkdownSections splits output into sections based on ## and ### headings.
func buildMarkdownSections(output string, totalLines int) []Section {
	lines := strings.Split(output, "\n")
	var sections []Section
	var current *Section

	for i, line := range lines {
		lineNum := i + 1
		if strings.HasPrefix(line, "##") {
			// Close previous section.
			if current != nil {
				current.EndLine = max(lineNum-1, current.StartLine)
				sections = append(sections, *current)
			}
			// Start new section.
			name := strings.TrimLeft(line, "#")
			name = strings.TrimSpace(name)
			current = &Section{
				Name:      name,
				StartLine: lineNum,
			}
		}
	}

	// Close final section.
	if current != nil {
		current.EndLine = totalLines
		sections = append(sections, *current)
	}

	// If no headings found, fall back to a single text chunk.
	if len(sections) == 0 {
		return buildTextChunks(output, totalLines)
	}

	// Ensure first section starts at 1 if there's preamble content.
	if len(sections) > 0 && sections[0].StartLine > 1 {
		preamble := Section{
			Name:      "Preamble",
			StartLine: 1,
			EndLine:   sections[0].StartLine - 1,
		}
		sections = append([]Section{preamble}, sections...)
	}

	return sections
}

// buildJSONSections detects top-level keys in a JSON object as sections.
func buildJSONSections(output string, totalLines int) []Section {
	// Try to parse as JSON object and extract top-level keys with line numbers.
	var obj map[string]json.RawMessage
	firstBlock := extractFirstJSONObject(output)
	if firstBlock != "" {
		if err := json.Unmarshal([]byte(firstBlock), &obj); err != nil {
			// Fall back to text chunks if parse fails.
			return buildTextChunks(output, totalLines)
		}
	} else {
		return buildTextChunks(output, totalLines)
	}

	if len(obj) == 0 {
		return buildTextChunks(output, totalLines)
	}

	// Scan lines to find key positions.
	lines := strings.Split(output, "\n")
	var sections []Section
	chunkSize := max(totalLines/max(len(obj), 1), 1)

	// Build sections by finding key occurrences in the lines.
	for key := range obj {
		quoted := `"` + key + `"`
		for i, line := range lines {
			if strings.Contains(line, quoted+":") || strings.Contains(line, quoted+" :") {
				startLine := i + 1
				endLine := min(startLine+chunkSize-1, totalLines)
				sections = append(sections, Section{
					Name:      key,
					StartLine: startLine,
					EndLine:   endLine,
					Summary:   "JSON key: " + key,
				})
				break
			}
		}
	}

	if len(sections) == 0 {
		return buildTextChunks(output, totalLines)
	}

	// Sort sections by StartLine.
	sortSections(sections)

	// Fix coverage: first section starts at 1, last ends at totalLines.
	if sections[0].StartLine > 1 {
		sections[0].StartLine = 1
	}
	sections[len(sections)-1].EndLine = totalLines

	return sections
}

// extractFirstJSONObject finds the first complete JSON object in output.
func extractFirstJSONObject(output string) string {
	start := strings.Index(output, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(output); i++ {
		ch := output[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inStr {
			escaped = true
			continue
		}
		if ch == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return output[start : i+1]
			}
		}
	}
	return ""
}

// buildTextChunks splits output into 50-line chunks.
func buildTextChunks(output string, totalLines int) []Section {
	if totalLines == 0 {
		return nil
	}
	const chunkSize = 50
	var sections []Section
	for start := 1; start <= totalLines; start += chunkSize {
		end := min(start+chunkSize-1, totalLines)
		sections = append(sections, Section{
			Name:      fmt.Sprintf("Lines %d-%d", start, end),
			StartLine: start,
			EndLine:   end,
		})
	}
	return sections
}

// sortSections sorts sections by StartLine ascending.
func sortSections(sections []Section) {
	n := len(sections)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sections[j].StartLine < sections[i].StartLine {
				sections[i], sections[j] = sections[j], sections[i]
			}
		}
	}
}


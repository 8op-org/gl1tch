package pipeline

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// --- OutputManifest unit tests ---

func TestBuildManifest_SmallOutput_IsInline(t *testing.T) {
	output := "hello world"
	m, err := BuildManifest(output, "step-a", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Inline != output {
		t.Errorf("expected Inline=%q, got %q", output, m.Inline)
	}
	if m.Path != "" {
		t.Errorf("expected no file path for small output, got %q", m.Path)
	}
}

func TestBuildManifest_LargeOutput_WritesFile(t *testing.T) {
	output := strings.Repeat("x", ManifestThreshold+1)
	dir := t.TempDir()

	m, err := BuildManifest(output, "step-big", dir)
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Path == "" {
		t.Fatal("expected file path for large output")
	}
	if m.Inline != "" {
		t.Errorf("expected no inline content for large output, got %q", m.Inline)
	}
	if m.Lines == 0 {
		t.Error("expected non-zero line count")
	}
	if m.Size == 0 {
		t.Error("expected non-zero size")
	}
}

func TestBuildManifest_FileIsReadable(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strings.Repeat("line content ", 30)
	}
	output := strings.Join(lines, "\n")
	dir := t.TempDir()

	m, err := BuildManifest(output, "step-read", dir)
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}

	if m.Path == "" {
		t.Skip("output below threshold, no file written")
	}
	if !filepath.IsAbs(m.Path) {
		t.Errorf("expected absolute path, got %q", m.Path)
	}
}

func TestBuildManifest_Markdown_DetectsSections(t *testing.T) {
	output := strings.Repeat("preamble line\n", 5) +
		"## Section One\n" + strings.Repeat("content one\n", 10) +
		"## Section Two\n" + strings.Repeat("content two\n", 10) +
		"### Subsection\n" + strings.Repeat("sub content\n", 10)
	// Ensure it's large enough to trigger file write
	for len(output) <= ManifestThreshold {
		output += strings.Repeat("padding\n", 20)
	}

	m, err := BuildManifest(output, "step-md", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Format != "markdown" {
		t.Errorf("expected format=markdown, got %q", m.Format)
	}
	if len(m.Sections) < 2 {
		t.Errorf("expected at least 2 sections detected, got %d: %+v", len(m.Sections), m.Sections)
	}

	names := make([]string, len(m.Sections))
	for i, s := range m.Sections {
		names[i] = s.Name
	}
	found := false
	for _, n := range names {
		if strings.Contains(n, "Section One") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Section One' in sections, got: %v", names)
	}
}

func TestBuildManifest_JSON_DetectsTopLevelKeys(t *testing.T) {
	obj := map[string]any{
		"summary":      strings.Repeat("s", 100),
		"results":      strings.Repeat("r", 100),
		"metadata":     strings.Repeat("m", 100),
		"dependencies": strings.Repeat("d", 100),
	}
	raw, _ := json.MarshalIndent(obj, "", "  ")
	output := string(raw)
	for len(output) <= ManifestThreshold {
		output += "\n" + string(raw)
	}

	m, err := BuildManifest(output, "step-json", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Format != "json" {
		t.Errorf("expected format=json, got %q", m.Format)
	}
	if len(m.Sections) == 0 {
		t.Error("expected sections from JSON top-level keys")
	}
}

func TestBuildManifest_PlainText_ProducesChunks(t *testing.T) {
	var sb strings.Builder
	for i := range 200 {
		sb.WriteString(strings.Repeat("line ", 10))
		sb.WriteString("\n")
		_ = i
	}
	output := sb.String()

	m, err := BuildManifest(output, "step-plain", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Format != "text" {
		t.Errorf("expected format=text, got %q", m.Format)
	}
	if len(m.Sections) == 0 {
		t.Error("expected chunk sections for plain text")
	}
	// Each section should have a start and end line
	for i, s := range m.Sections {
		if s.StartLine <= 0 {
			t.Errorf("section %d: StartLine must be > 0, got %d", i, s.StartLine)
		}
		if s.EndLine < s.StartLine {
			t.Errorf("section %d: EndLine (%d) must be >= StartLine (%d)", i, s.EndLine, s.StartLine)
		}
	}
}

func TestOutputManifest_Summary_IsSmall(t *testing.T) {
	output := strings.Repeat("large content line\n", 300)
	m, err := BuildManifest(output, "step-sum", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	summary := m.Summary()
	if len(summary) > 1024 {
		t.Errorf("manifest summary should be <1KB for context efficiency, got %d bytes", len(summary))
	}
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestOutputManifest_Summary_ContainsPath(t *testing.T) {
	output := strings.Repeat("large content line\n", 300)
	m, err := BuildManifest(output, "step-path", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Path == "" {
		t.Skip("output below threshold")
	}
	summary := m.Summary()
	if !strings.Contains(summary, m.Path) {
		t.Errorf("expected path %q in summary; got:\n%s", m.Path, summary)
	}
}

func TestOutputManifest_Summary_ContainsSectionNames(t *testing.T) {
	output := "## Analysis\n" + strings.Repeat("analysis content\n", 30) +
		"## Recommendations\n" + strings.Repeat("recommendation content\n", 30)
	for len(output) <= ManifestThreshold {
		output += strings.Repeat("padding\n", 20)
	}

	m, err := BuildManifest(output, "step-snames", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Path == "" {
		t.Skip("output below threshold")
	}
	summary := m.Summary()
	if !strings.Contains(summary, "Analysis") {
		t.Errorf("expected section name 'Analysis' in summary; got:\n%s", summary)
	}
}

func TestBuildManifest_SectionLineNumbers_CoverFullOutput(t *testing.T) {
	var sb strings.Builder
	for i := range 150 {
		sb.WriteString(strings.Repeat("x", 20))
		sb.WriteString("\n")
		_ = i
	}
	output := sb.String()

	m, err := BuildManifest(output, "step-cov", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if len(m.Sections) == 0 {
		t.Skip("no sections generated")
	}

	first := m.Sections[0]
	if first.StartLine != 1 {
		t.Errorf("first section should start at line 1, got %d", first.StartLine)
	}
	last := m.Sections[len(m.Sections)-1]
	if last.EndLine < m.Lines {
		t.Errorf("last section end (%d) should cover all %d lines", last.EndLine, m.Lines)
	}
}

func TestBuildManifest_ExactThreshold_IsInline(t *testing.T) {
	output := strings.Repeat("a", ManifestThreshold)
	m, err := BuildManifest(output, "step-thresh", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Path != "" {
		t.Errorf("output at exactly threshold should be inline, got path %q", m.Path)
	}
}

func TestBuildManifest_OneByteOverThreshold_WritesFile(t *testing.T) {
	output := strings.Repeat("a", ManifestThreshold+1)
	m, err := BuildManifest(output, "step-over", t.TempDir())
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.Path == "" {
		t.Error("output one byte over threshold should write file")
	}
}

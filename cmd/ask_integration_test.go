//go:build integration

package cmd

// Integration test that verifies gl1tch end-to-end by building the binary and
// running it as a subprocess.
//
// Definition of done: `gl1tch ask --pipeline pr-review --input <PR URL>` produces
// output that addresses all reviewer issues. No PR is posted — output is written
// to /tmp/gl1tch-pr-fix-output.md for manual inspection before any follow-up.
//
// Run with:
//
//	go test ./cmd/... -tags integration -run TestIntegration_GLitchAsk_PRReview -v -timeout 300s
//
// Requires:
//   - gh auth login  (read access to elastic/ensemble)
//   - copilot binary in PATH

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegration_GLitchAsk_PRReview builds the gl1tch binary and runs
// `gl1tch ask --pipeline pr-review --input <PR URL>`, then asserts the output
// addresses all four reviewer issues without posting a PR.
func TestIntegration_GLitchAsk_PRReview(t *testing.T) {
	// ── build gl1tch binary ───────────────────────────────────────────────────
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "gl1tch")

	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = filepath.Join("..", "") // repo root relative to cmd/
	// go build needs the module root, not cmd/
	build.Dir = repoRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	// ── run gl1tch ask ────────────────────────────────────────────────────────
	const prURL = "https://github.com/elastic/ensemble/pull/1246"

	run := exec.Command(binPath, "ask",
		"--pipeline", "pr-review",
		"-y",
		prURL,
	)
	run.Env = append(os.Environ(), "NO_COLOR=1")

	out, err := run.CombinedOutput()
	result := string(out)

	if err != nil {
		t.Logf("gl1tch output:\n%s", result)
		t.Fatalf("gl1tch ask failed: %v", err)
	}

	// ── write output for manual review ───────────────────────────────────────
	outFile := "/tmp/gl1tch-pr-fix-output.md"
	if werr := os.WriteFile(outFile, out, 0o644); werr != nil {
		t.Logf("warning: could not write output file: %v", werr)
	} else {
		t.Logf("fix output written to: %s", outFile)
	}
	t.Logf("response (first 2000 chars):\n%s", prTruncate(result, 2000))

	assertPRFixResponse(t, result)
}

// ── shared assertions ─────────────────────────────────────────────────────────

// assertPRFixResponse verifies a response addresses all four issues from the
// elastic/ensemble#1246 review and does not contain any PR-creation commands.
func assertPRFixResponse(t *testing.T, response string) {
	t.Helper()
	lower := strings.ToLower(response)

	checks := []struct {
		issue   string
		needles []string // any one match is sufficient
	}{
		{
			"subprocess replaced with Command wrapper",
			[]string{"command", "binds.base", "commandnotfound", "errorreturncode", "timeoutexception"},
		},
		{
			"CI installer replaces hardcoded linux_amd64 script",
			[]string{"tools install", "tools env", "ensemble tools", "linux_amd64", "oblt-cli", "oblt"},
		},
		{
			"regex tightened to avoid cross-block DOTALL matching",
			[]string{"re.multiline", "dotall", "anchor", "stack_version", "- name:", "state machine", "line-by-line", "pyyaml", "yaml.safe_load"},
		},
		{
			"assert replaced with raise ValueError",
			[]string{"raise valueerror", "valueerror", "raise value"},
		},
	}

	for _, c := range checks {
		found := false
		for _, needle := range c.needles {
			if strings.Contains(lower, needle) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("response missing fix for: %s\n  (checked needles: %v)", c.issue, c.needles)
		}
	}

	// Must not contain PR creation or push commands.
	forbidden := []string{
		"gh pr create",
		"create pull request",
		"git push",
		"push --set-upstream",
		"push origin",
	}
	for _, f := range forbidden {
		if strings.Contains(lower, f) {
			t.Errorf("response must not contain %q — no auto-posting until manual review", f)
		}
	}
}

func prTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// repoRoot returns the repository root by walking up from the cmd package.
func repoRoot(t *testing.T) string {
	t.Helper()
	// The test runs with working directory = cmd/, so the repo root is one level up.
	abs, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	return abs
}

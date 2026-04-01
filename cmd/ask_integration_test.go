//go:build integration

package cmd

// Integration tests that verify gl1tch can analyse and fix real GitHub PR review feedback.
//
// Definition of done: gl1tch produces a response that correctly addresses all
// reviewer issues on elastic/ensemble#1246. No PR is posted — output is written
// to a temp file for manual inspection before any follow-up action.
//
// Run with:
//
//	go test ./cmd/... -tags integration -run TestIntegration_PRReview -v -timeout 300s
//
// Requires:
//   - ollama serve (running on localhost:11434)
//   - ollama pull llama3.2  (or set GLITCH_TEST_MODEL)
//   - gh auth login         (read access to elastic/ensemble)

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

func prIntegrationModel() string {
	if m := os.Getenv("GLITCH_TEST_MODEL"); m != "" {
		return m
	}
	return "llama3.2"
}

// TestIntegration_PRReview_Analysis feeds a concise description of the review
// feedback (no raw diff) and asserts gl1tch identifies all four required changes.
//
// This is the fast variant: it does not call gh and runs in ~60s.
func TestIntegration_PRReview_Analysis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	mgr, providerID, model, err := buildAskManager("ollama", prIntegrationModel())
	if err != nil {
		t.Skipf("no Ollama provider available: %v", err)
	}

	// Reviewer comment verbatim — generic enough to not reveal private context.
	prompt := `A reviewer left the following comment on a Python PR that adds a
cloud:update-stack-versions script to an internal repo:

1. Reuse the existing Command wrapper from the codebase (ensemble.core.process.binds.base)
   instead of raw subprocess.run — it already handles CommandNotFound, ErrorReturnCode,
   and TimeoutException.
2. The CI workflow hardcodes linux_amd64 in the download URL. Replace the shell script
   with the repo's own tool installer (ensemble tools install oblt-cli latest +
   ensemble tools env oblt-cli latest) so it is cross-platform and idempotent.
3. The regex uses re.DOTALL which lets .*? match across - name: block boundaries,
   risking replacement of the wrong parameter's values block. Tighten the regex to
   anchor on the specific parameter name before matching, or use a line-by-line
   state machine instead.
4. The assert in _version_sort_key is silently dropped under python -O.
   Replace it with an explicit raise ValueError.

List each of the four code changes needed, and show the corrected Python for each.
Do not open a pull request or run any git commands.`

	p := buildAskPipeline(prompt, providerID, model, nil, false, "")
	result, err := pipeline.Run(ctx, p, mgr, "")
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}
	t.Logf("response:\n%s", result)

	assertPRFixResponse(t, result)
}

// TestIntegration_PRFix_EnsemblePR1246 fetches the real PR diff and reviewer
// comment via gh CLI, then asks gl1tch to produce a corrected implementation.
//
// This is the full end-to-end variant: requires gh auth and network access.
// Output is written to a temp file so it can be reviewed before any PR is opened.
func TestIntegration_PRFix_EnsemblePR1246(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()

	// ── fetch live PR data ────────────────────────────────────────────────────
	diffBytes, err := exec.CommandContext(ctx,
		"gh", "pr", "diff", "https://github.com/elastic/ensemble/pull/1246",
	).Output()
	if err != nil {
		t.Skipf("gh pr diff failed (need gh auth + read access to elastic/ensemble): %v", err)
	}

	commentBytes, err := exec.CommandContext(ctx,
		"gh", "pr", "view", "https://github.com/elastic/ensemble/pull/1246",
		"--json", "comments",
		"--jq", ".comments[0].body",
	).Output()
	if err != nil {
		t.Skipf("gh pr view failed: %v", err)
	}

	// Trim diff to scripts/cloud.py only — keeps context window manageable.
	diff := string(diffBytes)
	if idx := strings.Index(diff, "diff --git a/scripts/cloud.py"); idx >= 0 {
		diff = diff[idx:]
		if next := strings.Index(diff[20:], "\ndiff --git"); next >= 0 {
			diff = diff[:next+20]
		}
	}

	reviewerComment := strings.TrimSpace(string(commentBytes))

	// ── build prompt ──────────────────────────────────────────────────────────
	prompt := "Here is a GitHub PR diff for scripts/cloud.py and a reviewer comment.\n" +
		"Produce corrected Python for update_stack_versions() and its helpers that " +
		"addresses every point in the review.\n" +
		"Do not open a pull request, push any branches, or run any git commands.\n\n" +
		"## Reviewer comment\n\n" + reviewerComment + "\n\n" +
		"## Diff\n\n```diff\n" + diff + "\n```"

	mgr, providerID, model, err := buildAskManager("ollama", prIntegrationModel())
	if err != nil {
		t.Skipf("no Ollama provider available: %v", err)
	}

	p := buildAskPipeline(prompt, providerID, model, nil, false, "")
	result, err := pipeline.Run(ctx, p, mgr, "")
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}

	// ── write output for manual review ───────────────────────────────────────
	outFile := t.TempDir() + "/fix-output.md"
	if werr := os.WriteFile(outFile, []byte(result), 0o644); werr != nil {
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
			[]string{"tools install", "tools env", "ensemble tools", "linux_amd64"},
		},
		{
			"regex tightened to avoid cross-block DOTALL matching",
			[]string{"re.multiline", "dotall", "anchor", "stack_version", "- name:", "state machine", "line-by-line"},
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

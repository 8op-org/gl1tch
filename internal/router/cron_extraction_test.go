//go:build !integration

package router

// Tests for validateCron and sanitizeFocus — the two extraction helpers
// that clean LLM output before it reaches RouteResult.
// Cases are derived from real cron scheduling prompts in the session archives.

import "testing"

func TestValidateCron_RealSchedules(t *testing.T) {
	cases := []struct {
		label string
		input string // what the LLM returns
		want  string // what validateCron produces
	}{
		// ── valid 5-field expressions (from "every X" prompts) ─────────────────
		{"every morning at 9",      "0 9 * * *",   "0 9 * * *"},
		{"every weekday at 9am",    "0 9 * * 1-5", "0 9 * * 1-5"},
		{"every 2 hours",           "0 */2 * * *", "0 */2 * * *"},
		{"every hour",              "0 * * * *",   "0 * * * *"},
		{"every hour on weekdays",  "0 * * * 1-5", "0 * * * 1-5"},
		{"daily at midnight",       "0 0 * * *",   "0 0 * * *"},
		{"1st of every month",      "0 9 1 * *",   "0 9 1 * *"},
		{"every monday at 9",       "0 9 * * 1",   "0 9 * * 1"},
		{"every 30 minutes",        "*/30 * * * *","*/30 * * * *"},

		// ── LLM returns NONE or empty ──────────────────────────────────────────
		{"NONE uppercase",    "NONE",  ""},
		{"none lowercase",    "none",  ""},
		{"None mixed",        "None",  ""},
		{"empty string",      "",      ""},
		{"whitespace only",   "   ",   ""},

		// ── invalid field counts — all must return "" ──────────────────────────
		{"4 fields (missing dow)",  "0 9 * *",       ""},
		{"6 fields (with seconds)", "0 0 9 * * *",   ""},
		{"3 fields",                "9 * *",          ""},
		{"1 field",                 "0",              ""},
		{"7 fields",                "0 0 9 * * * *", ""},

		// ── extra whitespace around valid expression ───────────────────────────
		{"leading spaces",   "  0 9 * * *  ", "0 9 * * *"},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			got := validateCron(tc.input)
			if got != tc.want {
				t.Errorf("validateCron(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeFocus_ArchiveInputs(t *testing.T) {
	// Inputs are what the LLM emits for the "input" field —
	// often quoted or punctuated. sanitizeFocus must strip outer quotes/periods.
	cases := []struct {
		label string
		input string
		want  string
	}{
		// ── real extracted inputs from routing scenarios ───────────────────────
		{`double-quoted`,           `"support queue"`,    "support queue"},
		{`single-quoted`,           `'acme corp'`,        "acme corp"},
		{`period suffix`,           `executor package.`,  "executor package"},
		{`quoted with period`,      `"birthday card".`,   "birthday card"},
		{`apostrophe wrapped`,      `'my project'`,       "my project"},
		{`clean value`,             `opencode jq`,        "opencode jq"},
		{`codebase indexing`,       `codebase indexing`,  "codebase indexing"},

		// ── NONE variants — all must return "" ────────────────────────────────
		{"NONE uppercase",  "NONE",  ""},
		{"none lowercase",  "none",  ""},
		{"None mixed",      "None",  ""},
		{"empty string",    "",      ""},
		{"whitespace only", "   ",   ""},

		// ── multi-word inputs with internal punctuation preserved ─────────────
		{"url-style value", "gl1tch ask", "gl1tch ask"},
		{"hyphenated",      "support-digest", "support-digest"},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			got := sanitizeFocus(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeFocus(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

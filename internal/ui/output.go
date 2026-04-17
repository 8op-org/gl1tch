// Package ui provides styled terminal output for workflow execution.
// Uses lipgloss for color/formatting. No BubbleTea, no interactivity.
package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	teal   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68"))
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("#737aa2"))
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
	bold   = lipgloss.NewStyle().Bold(true)

	// Column width for right-aligned annotations
	colWidth = 40
)

// WorkflowStart prints the workflow header.
func WorkflowStart(name string) {
	fmt.Fprintf(os.Stderr, "\n%s %s\n", teal.Bold(true).Render(">>"), bold.Foreground(lipgloss.Color("#c0caf5")).Render(name))
}

// StepShell prints a shell step.
func StepShell(id string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", dim.Render(">"), dim.Render(id))
}

// StepRunning prints a long-running step with a hint that it's blocking.
func StepRunning(id string, hint string) {
	line := fmt.Sprintf("  %s %s", teal.Render("▸"), teal.Render(id))
	if hint != "" {
		line = padRight(line, colWidth) + dim.Render(hint)
	}
	fmt.Fprintln(os.Stderr, line)
}

// StepLLM prints an LLM step with provider and model annotation.
func StepLLM(id string, provider string, model string) {
	line := fmt.Sprintf("  %s %s", yellow.Render(">"), yellow.Render(id))
	anno := providerModel(provider, model)
	if anno != "" {
		line = padRight(line, colWidth) + dim.Render(anno)
	}
	fmt.Fprintln(os.Stderr, line)
}

// StepLLMDone annotates a completed LLM step with provider, model, token count and timing.
func StepLLMDone(id string, provider string, model string, tokensIn, tokensOut int64, elapsed time.Duration) {
	parts := []string{}
	anno := providerModel(provider, model)
	if anno != "" {
		parts = append(parts, anno)
	}
	total := tokensIn + tokensOut
	if total > 0 {
		parts = append(parts, fmt.Sprintf("%d tok", total))
	}
	if elapsed > time.Second {
		parts = append(parts, formatDuration(elapsed))
	}
	if len(parts) > 0 {
		annotation := strings.Join(parts, " · ")
		line := fmt.Sprintf("  %s %s", yellow.Render("✓"), yellow.Render(id))
		line = padRight(line, colWidth) + dim.Render(annotation)
		// Overwrite the "starting" line
		fmt.Fprintf(os.Stderr, "\033[1A\033[2K%s\n", line)
	}
}

func providerModel(provider, model string) string {
	if provider == "" && model == "" {
		return ""
	}
	if provider != "" && model != "" {
		return provider + "/" + model
	}
	if provider != "" {
		return provider
	}
	return model
}

// StepSave prints a save step with the file path.
func StepSave(id string, path string) {
	line := fmt.Sprintf("  %s %s", dim.Render(">"), dim.Render(id))
	if path != "" {
		line = padRight(line, colWidth) + dim.Render("→ "+path)
	}
	fmt.Fprintln(os.Stderr, line)
}

// StepPlugin prints a plugin call.
func StepPlugin(id string, plugin, subcommand string) {
	annotation := fmt.Sprintf("%s/%s", plugin, subcommand)
	line := fmt.Sprintf("  %s %s", teal.Render(">"), teal.Render(id))
	line = padRight(line, colWidth) + dim.Render(annotation)
	fmt.Fprintln(os.Stderr, line)
}

// StepSDK prints an SDK form step (json-pick, lines, merge, etc).
func StepSDK(id string, form string) {
	line := fmt.Sprintf("  %s %s", dim.Render(">"), dim.Render(id))
	if form != "" {
		line = padRight(line, colWidth) + dim.Render(form)
	}
	fmt.Fprintln(os.Stderr, line)
}

// StepRetry prints a retry annotation.
func StepRetry(id string, attempt, max int) {
	fmt.Fprintf(os.Stderr, "  %s %s %s\n",
		yellow.Render(">"), yellow.Render(id),
		dim.Render(fmt.Sprintf("(retry %d/%d)", attempt, max)))
}

// StepFallback prints a fallback step.
func StepFallback(id string, fallbackID string) {
	fmt.Fprintf(os.Stderr, "  %s %s %s\n",
		yellow.Render(">"), yellow.Render(id),
		dim.Render(fmt.Sprintf("(fallback → %s)", fallbackID)))
}

// PhaseStart prints a phase header.
func PhaseStart(name string, attempt, retries int) {
	fmt.Fprintln(os.Stderr)
	if attempt > 0 {
		fmt.Fprintf(os.Stderr, "%s Phase %s %s\n",
			teal.Bold(true).Render(">>>"),
			bold.Foreground(lipgloss.Color("#c0caf5")).Render(fmt.Sprintf("%q", name)),
			dim.Render(fmt.Sprintf("retry %d/%d", attempt, retries)))
	} else {
		fmt.Fprintf(os.Stderr, "%s Phase %s\n",
			teal.Bold(true).Render(">>>"),
			bold.Foreground(lipgloss.Color("#c0caf5")).Render(fmt.Sprintf("%q", name)))
	}
}

// PhaseStep prints a step within a phase.
func PhaseStep(id string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", dim.Render(">"), dim.Render(id))
}

// GateStart prints a gate being evaluated.
func GateStart(id string) {
	fmt.Fprintf(os.Stderr, "  %s gate %s\n", dim.Render(">"), dim.Render(id))
}

// GatePass prints a gate that passed.
func GatePass(id string) {
	line := fmt.Sprintf("  %s gate %s", green.Render("✓"), lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5")).Render(id))
	line = padRight(line, colWidth) + green.Render("PASS")
	// Overwrite the "starting" line
	fmt.Fprintf(os.Stderr, "\033[1A\033[2K%s\n", line)
}

// GateFail prints a gate that failed.
func GateFail(id string, reason string) {
	line := fmt.Sprintf("  %s gate %s", red.Render("✗"), red.Render(id))
	line = padRight(line, colWidth) + red.Render("FAIL")
	// Overwrite the "starting" line
	fmt.Fprintf(os.Stderr, "\033[1A\033[2K%s\n", line)
	if reason != "" {
		// First line of gate output as detail
		first := strings.SplitN(reason, "\n", 2)[0]
		if len(first) > 60 {
			first = first[:57] + "..."
		}
		fmt.Fprintf(os.Stderr, "    %s\n", dim.Render(first))
	}
}

// GateRetry prints a gate failure triggering phase retry.
func GateRetry(id string) {
	fmt.Fprintf(os.Stderr, "  %s gate %s %s\n",
		yellow.Render("↻"), yellow.Render(id),
		dim.Render("retrying phase"))
}

// WorkflowDone prints the final summary.
func WorkflowDone(name string, elapsed time.Duration, totalTokens int64, totalCost float64) {
	parts := []string{}
	if elapsed > 0 {
		parts = append(parts, formatDuration(elapsed))
	}
	if totalTokens > 0 {
		parts = append(parts, fmt.Sprintf("%d tokens", totalTokens))
	}
	if totalCost > 0 {
		parts = append(parts, fmt.Sprintf("$%.4f", totalCost))
	}
	if len(parts) > 0 {
		fmt.Fprintf(os.Stderr, "\n%s %s\n",
			dim.Render("──"),
			dim.Render(strings.Join(parts, " · ")))
	}
}

// TierLog prints a styled tier escalation message.
func TierLog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "    %s\n", dim.Render(msg))
}

// CompareWarn prints a warning about compare branch failures or skipped reviews.
func CompareWarn(msg string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", yellow.Render("WARN"), dim.Render(msg))
}

// SavedTo prints a file-saved confirmation (used by save steps that write to stdout).
func SavedTo(path string) {
	fmt.Fprintf(os.Stderr, "%s\n", dim.Render("saved "+path))
}

func padRight(s string, width int) string {
	// Strip ANSI for length calculation
	visible := lipgloss.Width(s)
	if visible >= width {
		return s + "  "
	}
	return s + strings.Repeat(" ", width-visible)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// Package clarify provides the Detector interface and registry for reactive
// agent clarification. Each executor type registers a Detector that recognises
// its output convention for requesting user input. The pipeline runner injects
// Instruction into prompts for registered executors, and the switchboard
// intercepts matching log lines to surface the clarification overlay.
package clarify

import "strings"

// Instruction is appended to the prompt for every registered executor.
// The executor (Claude, OpenCode, etc.) is expected to output the marker line
// and then stop generating, allowing the user to reply before a follow-up run.
const Instruction = `

---
SYSTEM INSTRUCTION — READ BEFORE RESPONDING:

If anything about this task is unclear and you need user input before you can do a good job, you MUST use this exact format — output the line below and STOP immediately, outputting nothing else:

ORCAI_CLARIFY: <your single most important question here>

Rules:
- ONE question only — pick the most critical unknown
- Output the ORCAI_CLARIFY line first, before any other text
- After that line, output NOTHING — no explanation, no numbered lists, no "I need to know X and Y and Z"
- The system will pause, ask the user, and re-run you with the answer
- If you DO have enough information, skip this entirely and proceed with the task

DO NOT ask questions in normal prose. The ORCAI_CLARIFY format is the ONLY way to request input.
---`

// Detector inspects a single log line for a clarification request.
type Detector interface {
	Detect(line string) (question string, found bool)
}

// NoOp is the default Detector — it never matches. Pipelines using executors
// without a registered Detector run to completion without interruption.
type NoOp struct{}

func (NoOp) Detect(string) (string, bool) { return "", false }

// StructuredDetector matches "ORCAI_CLARIFY: <question>" anywhere in a line,
// after stripping ANSI escape sequences. This is the standard convention for
// all AI executor types (claude, opencode, ollama, gemini, etc.).
type StructuredDetector struct{}

const marker = "ORCAI_CLARIFY:"

func (StructuredDetector) Detect(line string) (string, bool) {
	clean := stripANSI(line)
	idx := strings.Index(clean, marker)
	if idx < 0 {
		return "", false
	}
	q := strings.TrimSpace(clean[idx+len(marker):])
	if q == "" {
		return "", false
	}
	return q, true
}

// stripANSI removes ANSI CSI escape sequences from s.
func stripANSI(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // consume 'm'
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

// registry maps executor IDs to their Detector.
var registry = map[string]Detector{}

// Register associates executorID with d in the global registry.
// Call from init() or package startup; not goroutine-safe after init.
func Register(executorID string, d Detector) { registry[executorID] = d }

// Get returns the Detector for executorID, or NoOp{} if unregistered.
func Get(executorID string) Detector {
	if d, ok := registry[executorID]; ok {
		return d
	}
	return NoOp{}
}

// IsReactive reports whether executorID has an active Detector registered.
func IsReactive(executorID string) bool {
	_, ok := registry[executorID]
	return ok
}

func init() {
	// All standard AI executor types use the ORCAI_CLARIFY: structured protocol.
	for _, id := range []string{"claude", "opencode", "ollama", "gemini", "github-copilot"} {
		Register(id, StructuredDetector{})
	}
}

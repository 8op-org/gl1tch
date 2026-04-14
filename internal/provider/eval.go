package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BuildEvalPrompt constructs a self-evaluation prompt for the local model.
func BuildEvalPrompt(task, response string) string {
	return fmt.Sprintf(`Rate the quality of this response 1-5:
- 5: Complete, accurate, well-structured
- 3: Partially correct but missing key details
- 1: Wrong, irrelevant, or incoherent

Task: %s

Response: %s

Reply with only a number.`, task, response)
}

var scoreRe = regexp.MustCompile(`[1-5]`)

// ParseEvalScore extracts a 1-5 score from an LLM self-eval response.
// Returns 0 if no valid score is found.
func ParseEvalScore(response string) int {
	trimmed := strings.TrimSpace(response)

	// Try first character for bare number responses
	if len(trimmed) >= 1 {
		if n, err := strconv.Atoi(trimmed[:1]); err == nil && n >= 1 && n <= 5 {
			return n
		}
	}

	// Scan for first 1-5 digit
	match := scoreRe.FindString(trimmed)
	if match != "" {
		n, _ := strconv.Atoi(match)
		return n
	}

	return 0
}

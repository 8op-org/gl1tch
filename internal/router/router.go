package router

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
)

var reGitHubPR = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/pull/\d+`)
var reGitHubIssue = regexp.MustCompile(`https?://github\.com/[^/]+/[^/]+/issues/\d+`)

// Match picks the best workflow for the user's input.
// It tries fast URL-based matching first, then falls back to Ollama.
func Match(input string, workflows map[string]*pipeline.Workflow, model string) (*pipeline.Workflow, string) {
	// Fast path: detect GitHub URLs.
	if url := reGitHubPR.FindString(input); url != "" {
		if w, ok := workflows["github-pr-review"]; ok {
			return w, url
		}
	}
	if url := reGitHubIssue.FindString(input); url != "" {
		if w, ok := workflows["github-issues"]; ok {
			return w, url
		}
	}

	// Build a numbered menu for the LLM.
	var menu []string
	var keys []string
	i := 1
	for name, w := range workflows {
		desc := w.Description
		if desc == "" {
			desc = name
		}
		menu = append(menu, fmt.Sprintf("%d. %s — %s", i, name, strings.TrimSpace(desc)))
		keys = append(keys, name)
		i++
	}

	prompt := fmt.Sprintf(`Given these available workflows:
%s

The user asked: %q

Which workflow number best matches? Reply with ONLY the number, nothing else.`, strings.Join(menu, "\n"), input)

	if model == "" {
		model = "qwen2.5:7b"
	}
	out, err := provider.RunOllama(model, prompt)
	if err != nil {
		return nil, input
	}

	// Parse the number from the response.
	out = strings.TrimSpace(out)
	// Extract first number found in response.
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(out)
	if numStr == "" {
		return nil, input
	}
	n, err := strconv.Atoi(numStr)
	if err != nil || n < 1 || n > len(keys) {
		return nil, input
	}

	name := keys[n-1]
	return workflows[name], input
}

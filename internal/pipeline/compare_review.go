package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

// buildReviewPrompt constructs the judge prompt for a compare block.
// If cfg is nil, uses default generic criteria.
// If cfg.Prompt is set, uses custom prompt with branch outputs appended.
// If cfg.Criteria is set, generates structured scoring prompt.
func buildReviewPrompt(cfg *ReviewConfig, branchOutputs map[string]string) string {
	names := make([]string, 0, len(branchOutputs))
	for name := range branchOutputs {
		names = append(names, name)
	}
	sort.Strings(names)

	var outputsSection strings.Builder
	for _, name := range names {
		fmt.Fprintf(&outputsSection, "--- %s ---\n%s\n\n", strings.ToUpper(name), branchOutputs[name])
	}

	if cfg != nil && cfg.Prompt != "" {
		return fmt.Sprintf(`%s

Here are the outputs from each variant:

%s
You MUST end your response with the structured scoring format:

For each variant, output:
VARIANT: <name>
total: <score>/<max>

Then on the final line:
WINNER: <name>`, cfg.Prompt, outputsSection.String())
	}

	criteria := []string{"coherence", "completeness", "relevance"}
	if cfg != nil && len(cfg.Criteria) > 0 {
		criteria = cfg.Criteria
	}

	var criteriaLines strings.Builder
	for _, c := range criteria {
		fmt.Fprintf(&criteriaLines, "- %s\n", c)
	}

	return fmt.Sprintf(`You are a judge comparing outputs from different AI variants on the same task.

Score each variant on the following criteria (1-10 scale):
%s
Here are the outputs:

%s
For each variant, output EXACTLY this format (one block per variant):

VARIANT: <name>
%stotal: <sum>/%d

Then on the final line:
WINNER: <name of best variant>

Be objective. Score based on quality, not length.`,
		criteriaLines.String(),
		outputsSection.String(),
		buildCriteriaFormat(criteria),
		len(criteria)*10,
	)
}

func buildCriteriaFormat(criteria []string) string {
	var b strings.Builder
	for _, c := range criteria {
		fmt.Fprintf(&b, "%s: <score>/10\n", c)
	}
	return b.String()
}

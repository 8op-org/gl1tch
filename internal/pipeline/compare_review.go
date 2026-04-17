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
func buildReviewPrompt(cfg *ReviewConfig, branchOutputs map[string]string, objective string) string {
	names := make([]string, 0, len(branchOutputs))
	for name := range branchOutputs {
		names = append(names, name)
	}
	sort.Strings(names)

	var outputsSection strings.Builder
	for _, name := range names {
		fmt.Fprintf(&outputsSection, "--- %s ---\n%s\n\n", strings.ToUpper(name), branchOutputs[name])
	}

	objectiveClause := ""
	if objective != "" {
		objectiveClause = fmt.Sprintf("The stated objective for this comparison is:\n\"%s\"\n\nScore each variant based on how well it achieves this objective.\n\n", objective)
	}

	if cfg != nil && cfg.Prompt != "" {
		return fmt.Sprintf(`%s%s

Here are the outputs from each variant:

%s
You MUST end your response with the structured scoring format:

For each variant, output:
VARIANT: <name>
total: <score>/<max>

Then on the final line:
WINNER: <name>`, objectiveClause, cfg.Prompt, outputsSection.String())
	}

	criteria := []string{"coherence", "completeness", "relevance"}
	if cfg != nil && len(cfg.Criteria) > 0 {
		criteria = cfg.Criteria
	}

	var criteriaLines strings.Builder
	for _, c := range criteria {
		fmt.Fprintf(&criteriaLines, "- %s\n", c)
	}

	return fmt.Sprintf(`%sYou are a judge comparing outputs from different AI variants on the same task.

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
		objectiveClause,
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

// buildReflectionPrompt constructs the post-judge reflection prompt for a compare block.
func buildReflectionPrompt(objective string, scores string, branchOutputs map[string]string, winner string) string {
	names := make([]string, 0, len(branchOutputs))
	for name := range branchOutputs {
		names = append(names, name)
	}
	sort.Strings(names)

	var branches strings.Builder
	for _, name := range names {
		output := branchOutputs[name]
		if len(output) > 500 {
			output = output[:500] + "..."
		}
		fmt.Fprintf(&branches, "--- %s ---\n%s\n\n", strings.ToUpper(name), output)
	}

	return fmt.Sprintf(`You are reviewing the results of a model comparison.

Objective: "%s"

Judge scores:
%s

Winner: %s

Branch outputs (truncated):
%s
Produce a structured learning with EXACTLY this format:

FINDING: <one sentence — what did this comparison prove?>
MODEL_INSIGHT: <for each model, one sentence about strengths/weaknesses>
CONFIDENCE: <high|medium|low — how decisive was the result?>
RECOMMENDATION: <one sentence — what should future runs do differently?>`, objective, scores, winner, branches.String())
}

// ReflectionResult holds parsed fields from a reflection LLM response.
type ReflectionResult struct {
	Finding        string
	ModelInsight   map[string]string
	Confidence     string
	Recommendation string
}

// ParseReflection extracts structured fields from a reflection LLM response.
func ParseReflection(output string) ReflectionResult {
	r := ReflectionResult{ModelInsight: make(map[string]string)}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FINDING:") {
			r.Finding = strings.TrimSpace(strings.TrimPrefix(line, "FINDING:"))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			r.Confidence = strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
		} else if strings.HasPrefix(line, "RECOMMENDATION:") {
			r.Recommendation = strings.TrimSpace(strings.TrimPrefix(line, "RECOMMENDATION:"))
		} else if strings.HasPrefix(line, "MODEL_INSIGHT:") {
			r.ModelInsight["_raw"] = strings.TrimSpace(strings.TrimPrefix(line, "MODEL_INSIGHT:"))
		}
	}
	return r
}

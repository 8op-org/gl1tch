package research

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// PlanPrompt builds a prompt that instructs the LLM to select researcher names.
func PlanPrompt(question string, researchers []Researcher, hints string) string {
	var b strings.Builder
	b.WriteString("You are a research planner. Given a question and a menu of researchers, pick which researchers to use.\n\n")
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\n\nAvailable researchers:\n")
	for _, r := range researchers {
		fmt.Fprintf(&b, "- %s: %s\n", r.Name(), r.Describe())
	}
	if hints != "" {
		b.WriteString("\nHints: ")
		b.WriteString(hints)
		b.WriteString("\n")
	}
	b.WriteString("\nRules:\n")
	b.WriteString("- Output ONLY a JSON array of researcher names, e.g. [\"name1\", \"name2\"]\n")
	b.WriteString("- Pick only names from the list above\n")
	b.WriteString("- Never invent names that are not in the list\n")
	b.WriteString("- If no researcher is relevant, output []\n")
	return b.String()
}

// DraftPrompt builds a prompt that instructs the LLM to answer using only evidence,
// and to provide structured feedback on evidence quality.
func DraftPrompt(question string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString("You are a research assistant with full autonomy. Answer the question using the evidence provided below.\n\n")
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\n\nEvidence:\n")
	for i, e := range bundle.Items {
		fmt.Fprintf(&b, "\n--- Evidence %d (source: %s) ---\n", i+1, e.Source)
		if e.Title != "" {
			fmt.Fprintf(&b, "Title: %s\n", e.Title)
		}
		b.WriteString(e.Body)
		b.WriteString("\n")
	}
	b.WriteString("\nRules:\n")
	b.WriteString("- If the answer involves file changes, output each changed file as:\n")
	b.WriteString("  --- FILE: path/to/file.md ---\n")
	b.WriteString("  (complete file content)\n")
	b.WriteString("  --- END FILE ---\n")
	b.WriteString("- Cite only verbatim identifiers from the evidence\n")
	b.WriteString("- Say \"I don't have enough evidence\" if nothing relevant is provided\n")
	b.WriteString("- Never suggest commands to run — just produce the output\n\n")
	b.WriteString("After your answer, append a feedback section rating the evidence you received:\n\n")
	b.WriteString("--- FEEDBACK ---\n")
	b.WriteString("- evidence_quality: good|adequate|insufficient\n")
	b.WriteString("- missing: [\"what evidence you needed but didn't get\"]\n")
	b.WriteString("- useful: [\"what evidence was most helpful\"]\n")
	b.WriteString("- suggestion: \"how to improve evidence gathering next time\"\n")
	return b.String()
}

// CritiquePrompt builds a prompt that instructs the LLM to extract and label claims.
func CritiquePrompt(draft string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString("You are a fact-checker. Extract every factual claim from the draft below and label each as \"grounded\", \"partial\", or \"ungrounded\" based on the evidence.\n\n")
	b.WriteString("Draft:\n")
	b.WriteString(draft)
	b.WriteString("\n\nEvidence:\n")
	for i, e := range bundle.Items {
		fmt.Fprintf(&b, "\n--- Evidence %d (source: %s) ---\n", i+1, e.Source)
		if e.Title != "" {
			fmt.Fprintf(&b, "Title: %s\n", e.Title)
		}
		b.WriteString(e.Body)
		b.WriteString("\n")
	}
	b.WriteString("\nOutput ONLY a JSON array of objects with \"text\" and \"label\" fields.\n")
	b.WriteString("Example: [{\"text\": \"some claim\", \"label\": \"grounded\"}]\n")
	return b.String()
}

// JudgePrompt builds a prompt that asks for a single 0.0-1.0 score.
func JudgePrompt(question, draft string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString("You are a judge evaluating how well a draft answers a question given the available evidence.\n\n")
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\n\nDraft:\n")
	b.WriteString(draft)
	b.WriteString("\n\nEvidence:\n")
	for i, e := range bundle.Items {
		fmt.Fprintf(&b, "\n--- Evidence %d (source: %s) ---\n", i+1, e.Source)
		if e.Title != "" {
			fmt.Fprintf(&b, "Title: %s\n", e.Title)
		}
		b.WriteString(e.Body)
		b.WriteString("\n")
	}
	b.WriteString("\nOutput a single decimal number between 0.0 and 1.0 representing the quality score.\n")
	b.WriteString("0.0 = completely wrong or unsupported, 1.0 = perfectly answered with full evidence.\n")
	return b.String()
}

// ParsePlan extracts researcher names from LLM output with three tolerance passes.
func ParsePlan(raw string) ([]string, error) {
	// Pass 1: find JSON array
	if names, err := parsePlanJSON(raw); err == nil {
		return dedup(names), nil
	}
	// Pass 2: strip backslash-escaped quotes, retry
	cleaned := strings.ReplaceAll(raw, `\"`, `"`)
	if names, err := parsePlanJSON(cleaned); err == nil {
		return dedup(names), nil
	}
	// Pass 3: lex bare identifiers
	if names, err := parsePlanBare(raw); err == nil && len(names) > 0 {
		return dedup(names), nil
	}
	return nil, fmt.Errorf("parsePlan: could not extract names from: %s", raw)
}

// parsePlanJSON finds the first [ and matching ], then unmarshals a JSON string array.
func parsePlanJSON(raw string) ([]string, error) {
	start := strings.IndexByte(raw, '[')
	if start < 0 {
		return nil, fmt.Errorf("no opening bracket")
	}
	depth := 0
	end := -1
	for i := start; i < len(raw); i++ {
		switch raw[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				end = i
				goto found
			}
		}
	}
	return nil, fmt.Errorf("no matching closing bracket")
found:
	var names []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &names); err != nil {
		return nil, err
	}
	return names, nil
}

// parsePlanBare lexes bare identifiers (alphanumeric + hyphen) between [ and ].
func parsePlanBare(raw string) ([]string, error) {
	start := strings.IndexByte(raw, '[')
	if start < 0 {
		return nil, fmt.Errorf("no opening bracket")
	}
	end := strings.IndexByte(raw[start:], ']')
	if end < 0 {
		return nil, fmt.Errorf("no closing bracket")
	}
	inner := raw[start+1 : start+end]

	var names []string
	var current strings.Builder
	for _, r := range inner {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				names = append(names, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		names = append(names, current.String())
	}
	return names, nil
}

func dedup(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

package research

import (
	"encoding/json"
	"regexp"
	"strings"
)

var reFeedbackSection = regexp.MustCompile(`(?s)---\s*FEEDBACK\s*---\s*(.+)`)

// SplitDraftAndFeedback separates the draft content from the feedback section.
func SplitDraftAndFeedback(raw string) (draft, feedbackRaw string) {
	loc := reFeedbackSection.FindStringIndex(raw)
	if loc == nil {
		return strings.TrimSpace(raw), ""
	}
	return strings.TrimSpace(raw[:loc[0]]), strings.TrimSpace(raw[loc[0]:])
}

// ParseFeedback extracts structured feedback from the feedback section.
func ParseFeedback(raw string) (Feedback, error) {
	_, section := SplitDraftAndFeedback(raw)
	if section == "" {
		return Feedback{}, nil
	}

	var fb Feedback
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")

		if strings.HasPrefix(line, "evidence_quality:") {
			fb.Quality = cleanValue(strings.TrimPrefix(line, "evidence_quality:"))
		} else if strings.HasPrefix(line, "missing:") {
			fb.Missing = parseJSONList(strings.TrimPrefix(line, "missing:"))
		} else if strings.HasPrefix(line, "useful:") {
			fb.Useful = parseJSONList(strings.TrimPrefix(line, "useful:"))
		} else if strings.HasPrefix(line, "suggestion:") {
			fb.Suggestion = cleanValue(strings.TrimPrefix(line, "suggestion:"))
		}
	}
	return fb, nil
}

func cleanValue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	return s
}

func parseJSONList(s string) []string {
	s = strings.TrimSpace(s)
	var items []string
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		for _, item := range strings.Split(s, ",") {
			item = cleanValue(item)
			item = strings.Trim(item, "[]")
			item = strings.TrimSpace(item)
			if item != "" {
				items = append(items, item)
			}
		}
	}
	return items
}

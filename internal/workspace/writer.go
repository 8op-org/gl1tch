package workspace

import (
	"fmt"
	"regexp"
	"strings"
)

// UpdatePin rewrites a single :pin value for the named resource, preserving
// comments and whitespace elsewhere. If the :pin key is missing, it is added
// immediately before the closing `)` of the matching (resource ...) form.
func UpdatePin(src []byte, name, pin string) ([]byte, error) {
	s := string(src)
	block, err := findResourceBlock(s, name)
	if err != nil {
		return nil, err
	}

	pinRe := regexp.MustCompile(`:pin\s+"[^"]*"`)
	updated := pinRe.ReplaceAllStringFunc(block.content, func(match string) string {
		return fmt.Sprintf(`:pin %q`, pin)
	})
	if updated == block.content {
		// No :pin key — insert just before the closing paren of the block.
		idx := strings.LastIndex(updated, ")")
		if idx < 0 {
			return nil, fmt.Errorf("malformed resource block for %q", name)
		}
		updated = updated[:idx] + fmt.Sprintf(" :pin %q", pin) + updated[idx:]
	}
	return []byte(s[:block.start] + updated + s[block.end:]), nil
}

type resourceBlock struct {
	start, end int
	content    string
}

// findResourceBlock locates the (resource "<name>" ...) s-expression by
// scanning for the opening token, then counting parens to find the matching
// close.
func findResourceBlock(s, name string) (resourceBlock, error) {
	header := fmt.Sprintf(`(resource %q`, name)
	start := strings.Index(s, header)
	if start < 0 {
		return resourceBlock{}, fmt.Errorf("resource %q not found", name)
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '"':
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' && i+1 < len(s) {
					i++
				}
				i++
			}
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end := i + 1
				return resourceBlock{start: start, end: end, content: s[start:end]}, nil
			}
		}
	}
	return resourceBlock{}, fmt.Errorf("unbalanced parens in resource %q", name)
}

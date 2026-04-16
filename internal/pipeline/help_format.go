package pipeline

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/8op-org/gl1tch/internal/plugin"
)

// FormatHelp renders cobra-idiomatic help text for a workflow. Output goes
// to stdout via cmd/run.go's SetHelpFunc hook.
func FormatHelp(w *Workflow) string {
	var b strings.Builder

	// Header
	if w.Description != "" {
		fmt.Fprintf(&b, "%s - %s\n\n", w.Name, w.Description)
	} else {
		fmt.Fprintf(&b, "%s\n\n", w.Name)
	}

	// Usage
	b.WriteString("Usage:\n")
	fmt.Fprintf(&b, "  glitch run %s%s%s\n\n", w.Name, usageInputFragment(w), usageArgsFragment(w.Args))

	// Arguments (positional input)
	if w.Input != nil {
		b.WriteString("Arguments:\n")
		tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
		desc := w.Input.Description
		if w.Input.Implicit {
			desc = `(undocumented — add (input :description "...") to annotate)`
		}
		fmt.Fprintf(tw, "  input\t%s\n", desc)
		if w.Input.Example != "" {
			fmt.Fprintf(tw, "\tExample: %q\n", w.Input.Example)
		}
		tw.Flush()
		b.WriteString("\n")
	}

	// Flags
	if len(w.Args) > 0 {
		b.WriteString("Flags:\n")
		tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
		for _, a := range w.Args {
			tag := flagTag(a)
			desc := a.Description
			if a.Implicit {
				desc = fmt.Sprintf(`(undocumented — add (arg "%s" :description "...") to annotate)`, a.Name)
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\n", a.Name, tag, desc)
			if a.Example != "" {
				fmt.Fprintf(tw, "\t\tExample: --set %s=%q\n", a.Name, a.Example)
			}
		}
		tw.Flush()
		b.WriteString("\n")
	}

	// Source
	if w.SourceFile != "" {
		line := 1
		if len(w.Steps) > 0 && w.Steps[0].Line > 0 {
			line = w.Steps[0].Line
		}
		fmt.Fprintf(&b, "Defined in: %s:%d\n", w.SourceFile, line)
	}

	return b.String()
}

func usageInputFragment(w *Workflow) string {
	if w.Input == nil {
		return ""
	}
	return " [<input>]"
}

func usageArgsFragment(args []plugin.ArgDef) string {
	if len(args) == 0 {
		return ""
	}
	var parts []string
	for _, a := range args {
		frag := fmt.Sprintf("--set %s=<%s>", a.Name, a.Name)
		if !a.Required {
			frag = "[" + frag + "]"
		}
		parts = append(parts, frag)
	}
	return " " + strings.Join(parts, " ")
}

func flagTag(a plugin.ArgDef) string {
	if a.Required {
		return "(required)"
	}
	if a.Default != "" {
		return fmt.Sprintf("(optional, default: %s)", a.Default)
	}
	return "(optional)"
}

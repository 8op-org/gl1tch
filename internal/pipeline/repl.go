package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// REPL is a read-eval-print loop driven by an io.Reader/io.Writer pair.
// It maintains a persistent environment across expressions so that
// definitions made in one input line are visible in subsequent ones.
type REPL struct {
	ev  *Evaluator
	env *Env
	in  io.Reader
	out io.Writer
}

// NewREPL creates a REPL with a persistent env and registered builtins.
func NewREPL(ev *Evaluator, in io.Reader, out io.Writer) *REPL {
	env := NewEnv(nil)
	ev.RegisterBuiltins(env)
	return &REPL{ev: ev, env: env, in: in, out: out}
}

// Run executes the main loop: read lines, detect incomplete input
// (unbalanced parens), eval complete expressions, print results or errors.
func (r *REPL) Run() error {
	scanner := bufio.NewScanner(r.in)
	var buf strings.Builder
	prompt := "gl1tch> "

	r.writePrompt(prompt)

	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line)
		buf.WriteString("\n")

		src := buf.String()
		if !balanced(src) {
			prompt = "  ...   "
			r.writePrompt(prompt)
			continue
		}

		trimmed := strings.TrimSpace(src)
		buf.Reset()
		prompt = "gl1tch> "

		if trimmed == "" {
			r.writePrompt(prompt)
			continue
		}

		val, err := r.ev.RunSourceWithEnv(r.env, []byte(trimmed))
		if err != nil {
			fmt.Fprintf(r.out, "error: %s\n", err)
		} else if val != nil && val.String() != "" {
			fmt.Fprintln(r.out, val.String())
		}

		r.writePrompt(prompt)
	}
	// If we hit EOF with a non-empty buffer, the input was incomplete.
	if remaining := strings.TrimSpace(buf.String()); remaining != "" {
		fmt.Fprintf(r.out, "error: unexpected end of input (unbalanced expression)\n")
	}
	return scanner.Err()
}

func (r *REPL) writePrompt(p string) {
	fmt.Fprint(r.out, p)
}

// balanced reports whether src has balanced parentheses, accounting for
// strings ("..."), triple-backtick strings (```...```), and comments (;).
func balanced(src string) bool {
	depth := 0
	inString := false
	inTriple := false
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if inTriple {
			if ch == '`' && i+2 < len(src) && src[i+1] == '`' && src[i+2] == '`' {
				inTriple = false
				i += 2
			}
			continue
		}
		if inString {
			if ch == '\\' {
				i++
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '`':
			if i+2 < len(src) && src[i+1] == '`' && src[i+2] == '`' {
				inTriple = true
				i += 2
			}
		case '"':
			inString = true
		case ';':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case '(':
			depth++
		case ')':
			depth--
		}
	}
	return depth <= 0 && !inString && !inTriple
}

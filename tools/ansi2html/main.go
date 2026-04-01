// ansi2html converts ANSI-coloured text from stdin to an HTML fragment
// using the bbs.css colour classes. Wraps output in <pre class="ansi-logo">.
//
// Usage: tdfrender | ansi2html > site/src/generated/glitch-logo.html
package main

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"
)

// Maps ANSI colour codes to bbs.css helper classes.
var ansiClass = map[int]string{
	31: "fg-red",
	32: "fg-green",
	33: "fg-yellow",
	34: "fg-purple",
	35: "fg-pink",
	36: "fg-cyan",
	37: "fg-fg",
	90: "fg-comment",
	91: "fg-red",
	92: "fg-green",
	93: "fg-yellow",
	94: "fg-purple",
	95: "fg-pink",
	96: "fg-cyan",
	97: "fg-fg",
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(`<pre class="ansi-logo">`)

	first := true
	for scanner.Scan() {
		if !first {
			fmt.Print("\n")
		}
		first = false
		emitLine(scanner.Text())
	}

	fmt.Print("</pre>\n")
}

func emitLine(line string) {
	inSpan := false

	closeSpan := func() {
		if inSpan {
			fmt.Print("</span>")
			inSpan = false
		}
	}

	i := 0
	for i < len(line) {
		// Escape sequence?
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			j := i + 2
			for j < len(line) && line[j] != 'm' {
				j++
			}
			if j >= len(line) {
				break
			}
			seq := line[i+2 : j]
			i = j + 1

			closeSpan()

			if seq == "0" || seq == "" {
				continue
			}

			class := ""
			for _, p := range strings.Split(seq, ";") {
				n, err := strconv.Atoi(p)
				if err != nil {
					continue
				}
				if c, ok := ansiClass[n]; ok && c != "" {
					class = c
					break
				}
			}
			if class != "" {
				fmt.Printf(`<span class="%s">`, class)
				inSpan = true
			}
			continue
		}

		// Regular character(s) — collect until next escape.
		j := i
		for j < len(line) && line[j] != '\x1b' {
			j++
		}
		fmt.Print(html.EscapeString(line[i:j]))
		i = j
	}

	closeSpan()
}

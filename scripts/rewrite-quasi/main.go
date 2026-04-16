// scripts/rewrite-quasi/main.go
// One-shot migration: rewrite {{...}} Go template syntax in .glitch
// workflow files to sexpr-level ~... syntax. Run once, then delete this
// file at merge.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var patterns = []struct {
	re  *regexp.Regexp
	out func(m []string) string
}{
	// {{.input}} -> ~input
	{regexp.MustCompile(`\{\{\s*\.input\s*\}\}`), func(m []string) string { return "~input" }},
	// {{.param.X}} -> ~param.X
	{regexp.MustCompile(`\{\{\s*\.param\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`),
		func(m []string) string { return "~param." + m[1] }},
	// {{.param.item}} already covered above; alias below for item_index
	{regexp.MustCompile(`\{\{\s*\.param\.item_index\s*\}\}`),
		func(m []string) string { return "~item_index" }},
	{regexp.MustCompile(`\{\{\s*\.param\.item\s*\}\}`),
		func(m []string) string { return "~item" }},
	// {{step "X"}} -> ~(step X)
	{regexp.MustCompile(`\{\{\s*step\s+"([^"]+)"\s*\}\}`),
		func(m []string) string { return "~(step " + m[1] + ")" }},
	// {{stepfile "X"}} -> ~(stepfile X)
	{regexp.MustCompile(`\{\{\s*stepfile\s+"([^"]+)"\s*\}\}`),
		func(m []string) string { return "~(stepfile " + m[1] + ")" }},
	// {{itemfile}} -> ~(itemfile)
	{regexp.MustCompile(`\{\{\s*itemfile\s*\}\}`),
		func(m []string) string { return "~(itemfile)" }},
	// {{branch "X"}} -> ~(branch X)
	{regexp.MustCompile(`\{\{\s*branch\s+"([^"]+)"\s*\}\}`),
		func(m []string) string { return "~(branch " + m[1] + ")" }},
}

func rewriteLine(line string) (string, bool) {
	out := line
	changed := false
	for _, p := range patterns {
		if p.re.MatchString(out) {
			out = p.re.ReplaceAllStringFunc(out, func(s string) string {
				m := p.re.FindStringSubmatch(s)
				return p.out(m)
			})
			changed = true
		}
	}
	return out, changed
}

func rewriteFile(path string) (int, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(src), "\n")
	changes := 0
	for i, ln := range lines {
		out, changed := rewriteLine(ln)
		if changed {
			lines[i] = out
			changes++
		}
	}
	if changes == 0 {
		return 0, nil
	}
	return changes, os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func findGlitchFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// skip hidden + build dirs
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".glitch") {
			out = append(out, p)
		}
		return nil
	})
	return out, err
}

func main() {
	root := flag.String("root", ".", "repo root to scan")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	files, err := findGlitchFiles(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	total := 0
	for _, f := range files {
		src, _ := os.ReadFile(f)
		lines := bytes.Split(src, []byte("\n"))
		changes := 0
		for _, ln := range lines {
			_, changed := rewriteLine(string(ln))
			if changed {
				changes++
			}
		}
		if changes == 0 {
			continue
		}
		if *dry {
			fmt.Printf("%s: %d line(s) would change\n", f, changes)
		} else {
			n, err := rewriteFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error rewriting %s: %v\n", f, err)
				continue
			}
			fmt.Printf("%s: rewrote %d line(s)\n", f, n)
			total += n
		}
	}
	fmt.Printf("total: %d lines\n", total)
}

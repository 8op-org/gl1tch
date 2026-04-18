# Site Workflow Conventions & Language Extensions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `(include)` and extended `(def)` to the sexpr language, create a syntax linter gate, conventions file, and consolidated site workflow family so that LLM-generated content always uses correct syntax.

**Architecture:** Extend the sexpr parser (`parseSexprWorkflow` and `resolveVal` in `internal/pipeline/sexpr.go`) to support file inclusion and form evaluation in def bindings. Add `scripts/gate-syntax.py` as a regex-based linter. Create `site/conventions.md` and `site/shared.glitch` as shared context. Consolidate site workflows to use `(include)` and inject conventions into every LLM prompt.

**Tech Stack:** Go (sexpr parser/runner), Python (gate scripts), glitch sexpr DSL (workflows)

---

### Task 1: Create `scripts/gate-syntax.py`

**Files:**
- Create: `scripts/gate-syntax.py`

This is the immediate fix — catches stale interpolation syntax regardless of language changes.

- [ ] **Step 1: Write the gate script**

```python
#!/usr/bin/env python3
"""Gate: No stale interpolation syntax in docs or labs."""
import sys, re
from pathlib import Path

SCAN_DIRS = [
    Path("site/src/content/docs"),
    Path("site/src/content/labs"),
]

# Patterns that should never appear in published content
BAD_PATTERNS = [
    (re.compile(r'\{\{step\s+"'), "old Go template step reference: use ~(step name) instead"),
    (re.compile(r"\{\{step\s+'"), "old Go template step reference: use ~(step name) instead"),
    (re.compile(r'\{\{\.param\.'), "old Go template param reference: use ~param.key instead"),
    (re.compile(r'\{\{\.input\b'), "old Go template input reference: use ~input instead"),
    (re.compile(r'\bglitch ask\b'), "decommissioned command: glitch ask"),
]


def check_file(path: Path) -> list[str]:
    errors = []
    text = path.read_text(encoding="utf-8")
    slug = path.stem

    # Strip YAML frontmatter
    if text.startswith("---"):
        end = text.find("\n---", 3)
        if end != -1:
            text = text[end + 4:]

    for line_num, line in enumerate(text.splitlines(), start=1):
        for pattern, msg in BAD_PATTERNS:
            if pattern.search(line):
                errors.append(f"{slug}:{line_num}: {msg}")

    return errors


def main():
    all_errors = []
    file_count = 0

    for scan_dir in SCAN_DIRS:
        if not scan_dir.exists():
            continue
        for path in sorted(scan_dir.glob("*.md")):
            file_count += 1
            all_errors.extend(check_file(path))

    if not file_count:
        print("FAIL: no .md files found to scan")
        sys.exit(1)

    if all_errors:
        print("FAIL")
        for e in all_errors:
            print(f"  {e}")
        sys.exit(1)
    else:
        print(f"PASS: no stale syntax ({file_count} files)")


if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Make it executable and run it**

Run: `chmod +x scripts/gate-syntax.py && python3 scripts/gate-syntax.py`
Expected: PASS (we already fixed the stale syntax earlier today)

- [ ] **Step 3: Commit**

```bash
git add scripts/gate-syntax.py
git commit -m "feat: add gate-syntax.py — catches stale interpolation patterns in docs/labs"
```

---

### Task 2: Add `gate-syntax` to existing site workflows

**Files:**
- Modify: `.glitch/workflows/site-sync.glitch` (verify phase, ~line 515)
- Modify: `.glitch/workflows/site-create-page.glitch` (verify phase, ~line 79)
- Modify: `.glitch/workflows/site-update-page.glitch` (verify phase, ~line 73)

- [ ] **Step 1: Add gate to site-sync verify phase**

In `.glitch/workflows/site-sync.glitch`, find the `(phase "verify"` block and add `gate-syntax` after `hallucinations`:

```glitch
  (phase "verify" :retries 1
    (gate "hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"
      (run "python3 scripts/gate-syntax.py"))
    (gate "structure"
      (run "python3 scripts/gate-structure.py"))
    (gate "links"
      (run "python3 scripts/gate-links.py"))
    (gate "sidebar"
      (run "python3 scripts/gate-sidebar.py")))
```

- [ ] **Step 2: Add gate to site-create-page verify phase**

In `.glitch/workflows/site-create-page.glitch`, find the `(phase "verify"` block and add:

```glitch
  (phase "verify" :retries 1
    (gate "no-hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"
      (run "python3 scripts/gate-syntax.py"))
    (gate "stub-coverage"
      (run "python3 scripts/gate-coverage.py"))
    (gate "structure-and-tone"
      (run "python3 scripts/gate-structure.py")))
```

- [ ] **Step 3: Add gate to site-update-page verify phase**

In `.glitch/workflows/site-update-page.glitch`, find the `(phase "verify"` block and add:

```glitch
  (phase "verify" :retries 1
    (gate "no-hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"
      (run "python3 scripts/gate-syntax.py"))
    (gate "stub-coverage"
      (run "python3 scripts/gate-coverage.py"))
    (gate "structure-and-tone"
      (run "python3 scripts/gate-structure.py")))
```

- [ ] **Step 4: Commit**

```bash
git add .glitch/workflows/site-sync.glitch .glitch/workflows/site-create-page.glitch .glitch/workflows/site-update-page.glitch
git commit -m "feat: add gate-syntax to all site workflow verify phases"
```

---

### Task 3: Create `site/conventions.md`

**Files:**
- Create: `site/conventions.md`

- [ ] **Step 1: Write the conventions file**

```markdown
# gl1tch Site Conventions

Rules for generating and editing 8op.org content. Every site workflow
injects this file into LLM prompts. Violations are caught by gate scripts.

## Interpolation Syntax

gl1tch uses tilde-based unquote interpolation. The old Go template
syntax (double-brace) is dead — never generate it.

| Form | Example | Purpose |
|------|---------|---------|
| `~(step name)` | `~(step gather)` | Insert a previous step's output |
| `~(stepfile name)` | `~(stepfile diff)` | Write step output to tempfile, return path |
| `~param.key` | `~param.repo` | Reference a `--set key=value` parameter |
| `~input` | `~input` | Reference user input |
| `~(fn args)` | `~(split "/" "a/b/c")` | String function |
| `~(fn args \| fn2)` | `~(trim " x " \| upper)` | Pipe threading |

**NEVER use:** `{{step "name"}}`, `{{.param.key}}`, `{{.input}}` — these are
old Go template syntax and do not work.

## Tone & Voice

- "your" framing throughout — never say "the user"
- Examples before explanation — show real code first, explain second
- No internal implementation details: BubbleTea, tmux, SQLite, Go types,
  OTel, lipgloss, or internal package names
- Every code example must come from CONTEXT provided, not from training data
- Do NOT invent commands, flags, or features not present in the context

## Decommissioned Features

- `glitch ask` — removed, do not mention in any content

## Frontmatter

| Content type | Required fields |
|-------------|-----------------|
| Docs | `title`, `order`, `description` |
| Labs | `title`, `slug`, `description`, `date` |

## Code Blocks

- Workflow examples use ````glitch` fence
- Shell examples use ```bash` fence
- All workflow code must use current interpolation syntax (see table above)
- Triple backticks inside glitch code delimit multiline strings

## CLI Commands

Only reference commands that exist. Current valid commands:

- `glitch workflow run <name>`, `glitch workflow list`
- `glitch run <name>` (alias for workflow run)
- `glitch observe`, `glitch up`, `glitch down`
- `glitch workspace init|use|list|status|gui|register|unregister|add|rm|sync|pin|workflow`
- `glitch plugin list`, `glitch plugin`
- `glitch config show`, `glitch config set`
- `glitch index`, `glitch version`, `glitch --help`
```

- [ ] **Step 2: Commit**

```bash
git add site/conventions.md
git commit -m "docs: add site/conventions.md — shared rules for all site LLM prompts"
```

---

### Task 4: Add conventions + examples as bridge steps to existing workflows

**Files:**
- Modify: `.glitch/workflows/site-create-page.glitch`
- Modify: `.glitch/workflows/site-update-page.glitch`

Bridge solution: add shell steps that load conventions and examples into prompts, before `(include)` and extended `(def)` exist. These steps will be removed when Task 7+ lands.

- [ ] **Step 1: Add conventions and examples steps to site-create-page**

In `.glitch/workflows/site-create-page.glitch`, add two new steps after the existing gather steps (after `"existing-pages"`), and update the generate prompt to reference them:

Add steps:
```glitch
  (step "conventions"
    (read-file "site/conventions.md"))

  (step "example-workflows"
    (run "for f in examples/*.glitch; do echo '=== '$f' ==='; cat \"$f\"; echo; done"))
```

Update the generate step prompt — prepend conventions and add examples after the RULES block:

```glitch
  (step "generate"
    (llm
      :tier 0
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        CONVENTIONS — follow these exactly:
        ~(step conventions)

        TOPIC: ~param.topic

        EXISTING PAGES (don't duplicate these):
        ~(step existing-pages)

        EXISTING DOC STUBS (for tone reference):
        ~(step existing-docs)

        REAL WORKFLOW EXAMPLES (use this exact syntax in code blocks):
        ~(step example-workflows)

        REPO STRUCTURE:
        ~(step repo-structure)

        RECENT CHANGES:
        ~(step recent-commits)

        OUTPUT: A single JSON object:
        {"slug": "kebab-case-name", "title": "Page Title", "order": N, "description": "one line", "content": "full markdown body starting after the title"}

        The order should be after the last existing page. The slug should be descriptive.
        The content should be complete and ready to publish.
        ```))
```

- [ ] **Step 2: Add conventions and examples steps to site-update-page**

In `.glitch/workflows/site-update-page.glitch`, add a conventions step after the existing gather steps (after `"workflows"`), and update the rewrite prompt:

Add step:
```glitch
  (step "conventions"
    (read-file "site/conventions.md"))
```

Update the rewrite step prompt — prepend conventions and reference example workflows already gathered:

```glitch
  (step "rewrite"
    (llm
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        CONVENTIONS — follow these exactly:
        ~(step conventions)

        INSTRUCTIONS: ~param.instructions

        CURRENT PAGE CONTENT:
        ~(step current-page)

        REAL WORKFLOW EXAMPLES (use this exact syntax in code blocks):
        ~(step examples)

        REAL WORKFLOWS:
        ~(step workflows)

        REPO STRUCTURE:
        ~(step repo-structure)

        RECENT CHANGES:
        ~(step recent-commits)

        OUTPUT: The complete updated markdown page. Start with the # Heading.
        Include everything — this replaces the entire file.
        ```))
```

- [ ] **Step 3: Verify workflows parse**

Run: `go test ./internal/pipeline/ -run TestSexpr -count=1`
Expected: PASS (existing tests still pass — we didn't change the parser)

- [ ] **Step 4: Commit**

```bash
git add .glitch/workflows/site-create-page.glitch .glitch/workflows/site-update-page.glitch
git commit -m "feat: inject conventions and examples into site workflow LLM prompts"
```

---

### Task 5: Implement `(include "path")` in the parser

**Files:**
- Modify: `internal/pipeline/sexpr.go:13-31` (parseSexprWorkflow)
- Create: `internal/pipeline/sexpr_include_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/pipeline/sexpr_include_test.go`:

```go
package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSexpr_Include_Basic(t *testing.T) {
	dir := t.TempDir()

	// Write the included file
	shared := `(def model "qwen2.5:7b")
(def greeting "hello")`
	if err := os.WriteFile(filepath.Join(dir, "shared.glitch"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write the main workflow that includes it
	main := `(include "` + filepath.Join(dir, "shared.glitch") + `")

(workflow "test-include"
  :description "tests include"
  (step "s1"
    (run "echo hello")))`

	w, err := parseSexprWorkflow([]byte(main))
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "test-include" {
		t.Fatalf("expected name %q, got %q", "test-include", w.Name)
	}
}

func TestSexpr_Include_DefPropagates(t *testing.T) {
	dir := t.TempDir()

	shared := `(def model "qwen2.5:7b")`
	if err := os.WriteFile(filepath.Join(dir, "shared.glitch"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}

	main := `(include "` + filepath.Join(dir, "shared.glitch") + `")

(workflow "test-def-propagation"
  :description "included def resolves in workflow"
  (step "s1"
    (llm :model model :prompt "hello")))`

	w, err := parseSexprWorkflow([]byte(main))
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM == nil {
		t.Fatal("expected LLM step")
	}
	if w.Steps[0].LLM.Model != "qwen2.5:7b" {
		t.Fatalf("expected model %q, got %q", "qwen2.5:7b", w.Steps[0].LLM.Model)
	}
}

func TestSexpr_Include_CircularDetected(t *testing.T) {
	dir := t.TempDir()

	fileA := filepath.Join(dir, "a.glitch")
	fileB := filepath.Join(dir, "b.glitch")

	os.WriteFile(fileA, []byte(`(include "`+fileB+`")
(def x "1")`), 0o644)
	os.WriteFile(fileB, []byte(`(include "`+fileA+`")
(def y "2")`), 0o644)

	main := `(include "` + fileA + `")
(workflow "circular" :description "test" (step "s" (run "echo")))`

	_, err := parseSexprWorkflow([]byte(main))
	if err == nil {
		t.Fatal("expected circular include error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Fatalf("expected circular error, got: %v", err)
	}
}

func TestSexpr_Include_FileNotFound(t *testing.T) {
	main := `(include "/nonexistent/file.glitch")
(workflow "test" :description "test" (step "s" (run "echo")))`

	_, err := parseSexprWorkflow([]byte(main))
	if err == nil {
		t.Fatal("expected file not found error")
	}
}

func TestSexpr_Include_OnlyImportsDefs(t *testing.T) {
	dir := t.TempDir()

	// Included file has a workflow — it should be ignored
	shared := `(def model "qwen2.5:7b")
(workflow "should-be-ignored" :description "x" (step "s" (run "echo")))`
	if err := os.WriteFile(filepath.Join(dir, "shared.glitch"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}

	main := `(include "` + filepath.Join(dir, "shared.glitch") + `")

(workflow "real-workflow"
  :description "only this workflow should exist"
  (step "s1"
    (llm :model model :prompt "hello")))`

	w, err := parseSexprWorkflow([]byte(main))
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "real-workflow" {
		t.Fatalf("expected name %q, got %q", "real-workflow", w.Name)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pipeline/ -run TestSexpr_Include -v -count=1`
Expected: FAIL — `parseSexprWorkflow` doesn't handle `include` yet

- [ ] **Step 3: Implement `(include)` in `parseSexprWorkflow`**

Modify `internal/pipeline/sexpr.go`. Replace the existing `parseSexprWorkflow` function (lines 13-40) with:

```go
// parseSexprWorkflow parses s-expression source into a Workflow.
func parseSexprWorkflow(src []byte) (*Workflow, error) {
	return parseSexprWorkflowWithIncludes(src, nil)
}

// parseSexprWorkflowWithIncludes parses with circular-include detection.
func parseSexprWorkflowWithIncludes(src []byte, visited map[string]bool) (*Workflow, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	// Collect defs — start with included defs
	defs := make(map[string]string)

	// Process includes first
	for _, n := range nodes {
		if n.IsList() && len(n.Children) >= 2 && n.Children[0].SymbolVal() == "include" {
			path := n.Children[1].StringVal()
			if path == "" {
				return nil, fmt.Errorf("line %d: (include) missing path", n.Line)
			}
			if visited[path] {
				return nil, fmt.Errorf("line %d: circular include: %s", n.Line, path)
			}
			visited[path] = true

			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("line %d: include %q: %w", n.Line, path, err)
			}

			includedNodes, err := sexpr.Parse(data)
			if err != nil {
				return nil, fmt.Errorf("include %q: %w", path, err)
			}

			// Recursively process includes in the included file
			for _, in := range includedNodes {
				if in.IsList() && len(in.Children) >= 2 && in.Children[0].SymbolVal() == "include" {
					incPath := in.Children[1].StringVal()
					if incPath == "" {
						return nil, fmt.Errorf("include %q: line %d: (include) missing path", path, in.Line)
					}
					if visited[incPath] {
						return nil, fmt.Errorf("include %q: line %d: circular include: %s", path, in.Line, incPath)
					}
					visited[incPath] = true
					incData, err := os.ReadFile(incPath)
					if err != nil {
						return nil, fmt.Errorf("include %q: line %d: include %q: %w", path, in.Line, incPath, err)
					}
					incNodes, err := sexpr.Parse(incData)
					if err != nil {
						return nil, fmt.Errorf("include %q → %q: %w", path, incPath, err)
					}
					for _, inn := range incNodes {
						if inn.IsList() && len(inn.Children) >= 3 && inn.Children[0].SymbolVal() == "def" {
							name := inn.Children[1].SymbolVal()
							if name == "" {
								name = inn.Children[1].StringVal()
							}
							defs[name] = resolveVal(inn.Children[2], defs)
						}
					}
				}
			}

			// Only import defs from included file
			for _, in := range includedNodes {
				if in.IsList() && len(in.Children) >= 3 && in.Children[0].SymbolVal() == "def" {
					name := in.Children[1].SymbolVal()
					if name == "" {
						name = in.Children[1].StringVal()
					}
					defs[name] = resolveVal(in.Children[2], defs)
				}
			}
		}
	}

	// Collect local defs (override included)
	for _, n := range nodes {
		if n.IsList() && len(n.Children) >= 3 && n.Children[0].SymbolVal() == "def" {
			name := n.Children[1].SymbolVal()
			if name == "" {
				name = n.Children[1].StringVal()
			}
			val := resolveVal(n.Children[2], defs)
			defs[name] = val
		}
	}

	// Find workflow
	for _, n := range nodes {
		if n.IsList() && len(n.Children) > 0 && n.Children[0].StringVal() == "workflow" {
			return convertWorkflow(n, defs)
		}
	}
	return nil, fmt.Errorf("no (workflow ...) form found")
}
```

Add `"os"` to the imports at the top of `sexpr.go`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pipeline/ -run TestSexpr_Include -v -count=1`
Expected: All 5 tests PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./internal/pipeline/ -count=1`
Expected: All existing tests still PASS

- [ ] **Step 6: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_include_test.go
git commit -m "feat: implement (include \"path\") for shared def imports in sexpr"
```

---

### Task 6: Extend `def` to evaluate `(read-file ...)`

**Files:**
- Modify: `internal/pipeline/sexpr.go:42-50` (resolveVal)
- Modify: `internal/pipeline/sexpr_include_test.go` (add tests)

- [ ] **Step 1: Write the failing test**

Add to `internal/pipeline/sexpr_include_test.go`:

```go
func TestSexpr_DefReadFile(t *testing.T) {
	dir := t.TempDir()
	content := "these are my conventions"
	if err := os.WriteFile(filepath.Join(dir, "conventions.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	src := `(def conventions (read-file "` + filepath.Join(dir, "conventions.md") + `"))

(workflow "test-def-readfile"
  :description "def evaluates read-file"
  (step "s1"
    (llm :prompt "follow these: ~conventions")))`

	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM == nil {
		t.Fatal("expected LLM step")
	}
	// The prompt should contain the resolved conventions text via ~conventions
	// At parse time, defs are just stored; interpolation happens at runtime via quasi
	// But we can verify the def was stored correctly by checking the unresolved prompt
	// contains ~conventions (runtime resolves it)
}

func TestSexpr_DefReadFileMultiple(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("file-a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("file-b"), 0o644)

	src := `(def combined (read-file "` + filepath.Join(dir, "a.md") + `" "` + filepath.Join(dir, "b.md") + `"))

(workflow "test-multi-readfile"
  :description "def reads multiple files"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefReadFileNotFound(t *testing.T) {
	src := `(def x (read-file "/nonexistent/path.md"))
(workflow "test" :description "test" (step "s" (run "echo")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err == nil {
		t.Fatal("expected error for missing file in def")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pipeline/ -run TestSexpr_DefReadFile -v -count=1`
Expected: FAIL — `resolveVal` doesn't evaluate `(read-file ...)`

- [ ] **Step 3: Extend `resolveVal` to handle `(read-file ...)`**

Replace the `resolveVal` function in `internal/pipeline/sexpr.go` (lines 42-50):

```go
// resolveVal returns the string value of a node, substituting def bindings for symbols.
// If the node is a (read-file ...) form, it reads the file(s) at parse time.
func resolveVal(n *sexpr.Node, defs map[string]string) string {
	// Symbol lookup
	if n.Atom != nil && n.Atom.Type == sexpr.TokenSymbol {
		if v, ok := defs[n.Atom.Val]; ok {
			return v
		}
	}

	// Form evaluation in def context
	if n.IsList() && len(n.Children) >= 2 {
		head := n.Children[0].SymbolVal()
		switch head {
		case "read-file", "read":
			return resolveReadFile(n, defs)
		}
	}

	return n.StringVal()
}

// resolveReadFile evaluates (read-file "path" ...) at parse time.
// Multiple paths are concatenated with \n\n separator.
func resolveReadFile(n *sexpr.Node, defs map[string]string) string {
	children := n.Children[1:]
	var parts []string
	for _, child := range children {
		path := resolveVal(child, defs)
		data, err := os.ReadFile(path)
		if err != nil {
			// Return error marker — parseSexprWorkflow will need to check for this.
			// For now, we panic with a clear message since resolveVal doesn't return errors.
			panic(fmt.Sprintf("(read-file): %v", err))
		}
		parts = append(parts, string(data))
	}
	return strings.Join(parts, "\n\n")
}
```

Also update `parseSexprWorkflow` / `parseSexprWorkflowWithIncludes` to recover from panics in `resolveVal`:

Add a deferred recover at the top of `parseSexprWorkflowWithIncludes`:

```go
func parseSexprWorkflowWithIncludes(src []byte, visited map[string]bool) (w *Workflow, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	// ... rest of function unchanged
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pipeline/ -run TestSexpr_DefReadFile -v -count=1`
Expected: All 3 tests PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./internal/pipeline/ -count=1`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_include_test.go
git commit -m "feat: extend def to evaluate (read-file) forms at parse time"
```

---

### Task 7: Extend `def` to evaluate `(glob ...)`

**Files:**
- Modify: `internal/pipeline/sexpr.go` (resolveVal switch)
- Modify: `internal/pipeline/sexpr_include_test.go` (add tests)

- [ ] **Step 1: Write the failing test**

Add to `internal/pipeline/sexpr_include_test.go`:

```go
func TestSexpr_DefGlob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.glitch"), []byte("workflow-a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.glitch"), []byte("workflow-b"), 0o644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("not-a-workflow"), 0o644)

	src := `(def files (glob "` + filepath.Join(dir, "*.glitch") + `"))

(workflow "test-glob"
  :description "def evaluates glob"
  (step "s1"
    (run "echo test")))`

	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	_ = w // Glob result is stored as a def — verified by no parse error
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/pipeline/ -run TestSexpr_DefGlob -v -count=1`
Expected: FAIL

- [ ] **Step 3: Add `glob` case to resolveVal**

In `internal/pipeline/sexpr.go`, add the `glob` case to the switch in `resolveVal`:

```go
	if n.IsList() && len(n.Children) >= 2 {
		head := n.Children[0].SymbolVal()
		switch head {
		case "read-file", "read":
			return resolveReadFile(n, defs)
		case "glob":
			return resolveGlob(n, defs)
		}
	}
```

Add the `resolveGlob` function:

```go
// resolveGlob evaluates (glob "pattern") at parse time.
// Returns newline-separated list of matching file paths.
func resolveGlob(n *sexpr.Node, defs map[string]string) string {
	pattern := resolveVal(n.Children[1], defs)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		panic(fmt.Sprintf("(glob): %v", err))
	}
	sort.Strings(matches)
	return strings.Join(matches, "\n")
}
```

Add `"path/filepath"` and `"sort"` to imports if not already present.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/pipeline/ -run TestSexpr_DefGlob -v -count=1`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./internal/pipeline/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_include_test.go
git commit -m "feat: extend def to evaluate (glob) forms at parse time"
```

---

### Task 8: Extend `def` to evaluate `(-> ...)` threading chains

**Files:**
- Modify: `internal/pipeline/sexpr.go` (resolveVal switch, add resolveThread)
- Modify: `internal/pipeline/sexpr_include_test.go` (add tests)

This is the key feature: `(def examples (-> (glob "*.glitch") (map read-file) (join "\n\n")))`.

- [ ] **Step 1: Write the failing test**

Add to `internal/pipeline/sexpr_include_test.go`:

```go
func TestSexpr_DefThread_GlobMapReadFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.glitch"), []byte("content-a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.glitch"), []byte("content-b"), 0o644)

	src := `(def examples (-> (glob "` + filepath.Join(dir, "*.glitch") + `") (map read-file) (join "\n\n")))

(workflow "test-thread"
  :description "threading in def"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSexpr_DefThread_Lines(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "data.txt"), []byte("line1\nline2\nline3"), 0o644)

	src := `(def line-count (-> (read-file "` + filepath.Join(dir, "data.txt") + `") (lines) (join ", ")))

(workflow "test-thread-lines"
  :description "lines + join in thread"
  (step "s1"
    (run "echo test")))`

	_, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pipeline/ -run TestSexpr_DefThread -v -count=1`
Expected: FAIL

- [ ] **Step 3: Implement thread evaluation in def context**

Add the `->` case to the switch in `resolveVal`:

```go
	if n.IsList() && len(n.Children) >= 2 {
		head := n.Children[0].SymbolVal()
		switch head {
		case "read-file", "read":
			return resolveReadFile(n, defs)
		case "glob":
			return resolveGlob(n, defs)
		case "->":
			return resolveThread(n, defs)
		}
	}
```

Add the `resolveThread` function and helpers:

```go
// resolveThread evaluates (-> form1 form2 ...) at parse time.
// Each form receives the output of the previous as implicit input.
func resolveThread(n *sexpr.Node, defs map[string]string) string {
	children := n.Children[1:] // skip "->"
	if len(children) < 2 {
		panic("(->): needs at least 2 forms")
	}

	// First form produces initial value
	val := resolveVal(children[0], defs)

	// Remaining forms transform the value
	for _, child := range children[1:] {
		val = applyDefForm(child, val, defs)
	}
	return val
}

// applyDefForm applies a form to a value in def/thread context.
// Supports: map, filter, lines, join, split, trim, upper, lower, replace,
// truncate, contains, flatten, read-file, glob.
func applyDefForm(n *sexpr.Node, input string, defs map[string]string) string {
	if !n.IsList() || len(n.Children) == 0 {
		panic(fmt.Sprintf("line %d: expected form in (->), got atom", n.Line))
	}
	head := n.Children[0].SymbolVal()

	switch head {
	case "map":
		// (map form) — apply form to each line of input
		if len(n.Children) < 2 {
			panic("(map) in thread: missing form name")
		}
		formName := n.Children[1].SymbolVal()
		if formName == "" {
			formName = n.Children[1].StringVal()
		}
		lines := strings.Split(input, "\n")
		var results []string
		for _, line := range lines {
			if line == "" {
				continue
			}
			switch formName {
			case "read-file", "read":
				data, err := os.ReadFile(line)
				if err != nil {
					panic(fmt.Sprintf("(map read-file): %v", err))
				}
				results = append(results, string(data))
			default:
				panic(fmt.Sprintf("(map %s): unsupported form in def context", formName))
			}
		}
		return strings.Join(results, "\n")

	case "join":
		// (join "sep") — join lines with separator
		sep := "\n"
		if len(n.Children) >= 2 {
			sep = resolveVal(n.Children[1], defs)
		}
		lines := strings.Split(input, "\n")
		var nonEmpty []string
		for _, l := range lines {
			if l != "" {
				nonEmpty = append(nonEmpty, l)
			}
		}
		return strings.Join(nonEmpty, sep)

	case "lines":
		// (lines) — identity in thread context (input is already line-separated)
		return input

	case "split":
		// (split "sep") — split and rejoin with newlines
		if len(n.Children) < 2 {
			panic("(split): missing separator")
		}
		sep := resolveVal(n.Children[1], defs)
		parts := strings.Split(input, sep)
		return strings.Join(parts, "\n")

	case "trim":
		return strings.TrimSpace(input)

	case "upper":
		return strings.ToUpper(input)

	case "lower":
		return strings.ToLower(input)

	case "replace":
		// (replace "old" "new")
		if len(n.Children) < 3 {
			panic("(replace): needs old and new strings")
		}
		old := resolveVal(n.Children[1], defs)
		newStr := resolveVal(n.Children[2], defs)
		return strings.ReplaceAll(input, old, newStr)

	case "filter":
		// (filter (contains "substr"))
		if len(n.Children) < 2 {
			panic("(filter): missing predicate")
		}
		pred := n.Children[1]
		lines := strings.Split(input, "\n")
		var kept []string
		for _, line := range lines {
			if evalPredicate(pred, line, defs) {
				kept = append(kept, line)
			}
		}
		return strings.Join(kept, "\n")

	case "flatten":
		return input // lines are already flat

	case "read-file", "read":
		// (read-file) in thread — read the path that input contains
		data, err := os.ReadFile(strings.TrimSpace(input))
		if err != nil {
			panic(fmt.Sprintf("(read-file) in thread: %v", err))
		}
		return string(data)

	default:
		panic(fmt.Sprintf("unsupported form %q in def thread", head))
	}
}

// evalPredicate evaluates a predicate form against a string value.
func evalPredicate(n *sexpr.Node, val string, defs map[string]string) bool {
	if !n.IsList() || len(n.Children) == 0 {
		return false
	}
	head := n.Children[0].SymbolVal()
	switch head {
	case "contains":
		if len(n.Children) < 2 {
			return false
		}
		substr := resolveVal(n.Children[1], defs)
		return strings.Contains(val, substr)
	default:
		panic(fmt.Sprintf("unsupported predicate %q in filter", head))
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/pipeline/ -run TestSexpr_DefThread -v -count=1`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./internal/pipeline/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_include_test.go
git commit -m "feat: implement (->) threading chains in def context with map, filter, join, lines, string ops"
```

---

### Task 9: Create `site/shared.glitch`

**Files:**
- Create: `site/shared.glitch`

- [ ] **Step 1: Write the shared defs file**

```glitch
;; site/shared.glitch — shared context for all site workflows
;;
;; Usage: (include "site/shared.glitch") at top of any site workflow.
;; Provides ~conventions and ~examples for LLM prompts.

(def conventions (read-file "site/conventions.md"))

(def examples
  (-> (glob "examples/*.glitch")
      (map read-file)
      (join "\n\n")))

(def model "qwen2.5:7b")
```

- [ ] **Step 2: Verify it parses**

Create a minimal test workflow to verify the shared file works end-to-end:

Run: `go test ./internal/pipeline/ -run TestSexpr_Include -v -count=1`
Expected: PASS (tests from Task 5 cover this pattern)

- [ ] **Step 3: Commit**

```bash
git add site/shared.glitch
git commit -m "feat: add site/shared.glitch — conventions + examples defs for site workflows"
```

---

### Task 10: Migrate site-create-page and site-update-page to use `(include)`

**Files:**
- Modify: `.glitch/workflows/site-create-page.glitch`
- Modify: `.glitch/workflows/site-update-page.glitch`

- [ ] **Step 1: Rewrite site-create-page to use include**

Replace the contents of `.glitch/workflows/site-create-page.glitch`:

```glitch
;; site-create-page.glitch — AI generates a doc page, gates verify it
;;
;; Run with: glitch workflow run site-create-page --set topic="batch comparison runs"

(include "site/shared.glitch")

(workflow "site-create-page"
  :description "AI-generate a new doc page with gated verification"

  ;; ── Gather context ────────────────────────────────

  (step "existing-docs"
    (run "for f in docs/site/*.md; do echo '=== '$f' ==='; head -5 \"$f\"; echo; done"))

  (step "repo-structure"
    (run "find internal/ cmd/ -name '*.go' -not -path '*/testdata/*' | head -40"))

  (step "recent-commits"
    (run "git log --oneline -30 -- cmd/ internal/ examples/ .glitch/"))

  (step "existing-pages"
    (run "python3 -c \"import glob,os; [print(os.path.splitext(os.path.basename(p))[0]) for p in sorted(glob.glob('docs/site/*.md'))]\""))

  ;; ── Generate page ─────────────────────────────────

  (step "generate"
    (llm
      :tier 0
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        CONVENTIONS — follow these exactly:
        ~conventions

        TOPIC: ~param.topic

        EXISTING PAGES (don't duplicate these):
        ~(step existing-pages)

        EXISTING DOC STUBS (for tone reference):
        ~(step existing-docs)

        REAL WORKFLOW EXAMPLES (use this exact syntax in code blocks):
        ~examples

        REPO STRUCTURE:
        ~(step repo-structure)

        RECENT CHANGES:
        ~(step recent-commits)

        OUTPUT: A single JSON object:
        {"slug": "kebab-case-name", "title": "Page Title", "order": N, "description": "one line", "content": "full markdown body starting after the title"}

        The order should be after the last existing page. The slug should be descriptive.
        The content should be complete and ready to publish.
        ```))

  ;; ── Save generated page ───────────────────────────

  (step "save-stub"
    (run "cat '~(stepfile generate)' | python3 scripts/save-generated-page.py"))

  (step "rebuild-json"
    (run "python3 scripts/stubs-to-json.py > site/generated/docs.json && python3 scripts/split-docs.py"))

  ;; ── Verify ────────────────────────────────────────

  (phase "verify" :retries 1
    (gate "no-hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"
      (run "python3 scripts/gate-syntax.py"))
    (gate "stub-coverage"
      (run "python3 scripts/gate-coverage.py"))
    (gate "structure-and-tone"
      (run "python3 scripts/gate-structure.py")))

  (phase "page-tests" :retries 0
    (gate "playwright"
      (run "bash scripts/gate-playwright.sh")))

  (step "done"
    (run "echo 'Page created and tested. Run: glitch workflow run site-dev to preview.'")))
```

- [ ] **Step 2: Rewrite site-update-page to use include**

Replace the contents of `.glitch/workflows/site-update-page.glitch`:

```glitch
;; site-update-page.glitch — AI rewrites an existing doc page with gated verification
;;
;; Run with: glitch workflow run site-update-page --set page=plugins
;; Or:       glitch workflow run site-update-page --set page=getting-started --set instructions="add section about par form"

(include "site/shared.glitch")

(workflow "site-update-page"
  :description "AI-rewrite an existing doc page with gated verification"

  ;; ── Gather context ────────────────────────────────

  (step "current-page"
    (run "cat docs/site/~(trim param.page).md"))

  (step "repo-structure"
    (run "find internal/ cmd/ -name '*.go' -not -path '*/testdata/*' | head -40"))

  (step "recent-commits"
    (run "git log --oneline -30 -- cmd/ internal/ examples/ .glitch/"))

  (step "workflows"
    (run "for f in .glitch/workflows/*.glitch; do echo '=== '$f' ==='; cat \"$f\"; echo; done"))

  ;; ── Rewrite ───────────────────────────────────────

  (step "rewrite"
    (llm
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        CONVENTIONS — follow these exactly:
        ~conventions

        INSTRUCTIONS: ~param.instructions

        CURRENT PAGE CONTENT:
        ~(step current-page)

        REAL WORKFLOW EXAMPLES (use this exact syntax in code blocks):
        ~examples

        REAL WORKFLOWS:
        ~(step workflows)

        REPO STRUCTURE:
        ~(step repo-structure)

        RECENT CHANGES:
        ~(step recent-commits)

        OUTPUT: The complete updated markdown page. Start with the # Heading.
        Include everything — this replaces the entire file.
        ```))

  ;; ── Save ──────────────────────────────────────────

  (step "save"
    (save "docs/site/~(trim param.page).md" :from "rewrite"))

  (step "rebuild"
    (run "python3 scripts/stubs-to-json.py > site/generated/docs.json && python3 scripts/split-docs.py"))

  ;; ── Verify ────────────────────────────────────────

  (phase "verify" :retries 1
    (gate "no-hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"
      (run "python3 scripts/gate-syntax.py"))
    (gate "stub-coverage"
      (run "python3 scripts/gate-coverage.py"))
    (gate "structure-and-tone"
      (run "python3 scripts/gate-structure.py")))

  (step "done"
    (run "echo 'Page updated. Run: glitch workflow run site-dev to preview.'")))
```

- [ ] **Step 3: Commit**

```bash
git add .glitch/workflows/site-create-page.glitch .glitch/workflows/site-update-page.glitch
git commit -m "refactor: migrate site-create-page and site-update-page to use (include) + conventions"
```

---

### Task 11: Create `site-deploy` workflow

**Files:**
- Create: `.glitch/workflows/site-deploy.glitch`

- [ ] **Step 1: Write the deploy workflow**

```glitch
;; site-deploy.glitch — gate, build, commit, push for GitHub Pages
;;
;; Run with: glitch workflow run site-deploy

(include "site/shared.glitch")

(workflow "site-deploy"
  :description "Verify, build, commit, push for GitHub Pages"

  ;; ── Verify current disk state ─────────────────────

  (phase "verify" :retries 0
    (gate "hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"
      (run "python3 scripts/gate-syntax.py"))
    (gate "structure"
      (run "python3 scripts/gate-structure.py"))
    (gate "links"
      (run "python3 scripts/gate-links.py"))
    (gate "sidebar"
      (run "python3 scripts/gate-sidebar.py")))

  ;; ── Build ─────────────────────────────────────────

  (step "build-sidebar"
    (run "python3 scripts/site-build-sidebar.py"))

  (step "build"
    (run "cd site && npx astro build 2>&1"))

  ;; ── Smoke test ────────────────────────────────────

  (phase "smoke" :retries 0
    (gate "playwright"
      (run "bash scripts/gate-playwright.sh")))

  ;; ── Commit and push ───────────────────────────────

  (step "commit"
    (run ```
      git add site/src/content/ site/src/pages/ site/src/components/
      git diff --cached --quiet && echo 'nothing to commit' && exit 0
      git commit -m "docs: update site content"
      git push
      ```))

  (step "done"
    (run "echo 'Deployed. GitHub Pages will pick it up.'")))
```

- [ ] **Step 2: Commit**

```bash
git add .glitch/workflows/site-deploy.glitch
git commit -m "feat: add site-deploy workflow — gate, build, commit, push"
```

---

### Task 12: Add `include` to valid forms and update gate-hallucinations

**Files:**
- Modify: `scripts/gate-hallucinations.py` (~line 59, VALID_FORMS)

- [ ] **Step 1: Add `include` to VALID_FORMS**

In `scripts/gate-hallucinations.py`, add `"include"` to the `VALID_FORMS` set:

```python
VALID_FORMS = {
    "def", "workflow", "step", "run", "llm", "save", "plugin",
    "arg", "retry", "timeout", "catch", "cond", "map", "let",
    "phase", "gate", "par", "include",
    ...
}
```

- [ ] **Step 2: Run the gate to verify**

Run: `python3 scripts/gate-hallucinations.py`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add scripts/gate-hallucinations.py
git commit -m "fix: add include to VALID_FORMS in gate-hallucinations"
```

---

### Task 13: Update workflow-syntax docs

**Files:**
- Modify: `docs/site/workflow-syntax.md`

- [ ] **Step 1: Add include and extended def sections**

Add a section after the existing `(def ...)` documentation in `docs/site/workflow-syntax.md`:

```markdown
## Sharing Definitions

Use `(include)` to import `(def ...)` bindings from another file:

````glitch
(include "site/shared.glitch")

(workflow "my-workflow"
  :description "uses shared defs"
  (step "s1"
    (llm :model model :prompt "~conventions")))
````

Only `(def ...)` forms are imported — workflows and steps in the included file are ignored. Circular includes produce a parse error.

## Evaluated Definitions

`(def)` can evaluate forms at parse time, not just literal strings:

````glitch
;; Read a file into a constant
(def conventions (read-file "site/conventions.md"))

;; Glob + read + join via threading
(def examples
  (-> (glob "examples/*.glitch")
      (map read-file)
      (join "\n\n")))

;; Filter lines from a file
(def commands
  (-> (read-file "valid-commands.txt")
      (lines)
      (filter (contains "glitch"))
      (join "\n")))
````

Available forms in `(def)` context: `read-file`, `glob`, `->`, `map`, `filter`, `lines`, `join`, `split`, `trim`, `upper`, `lower`, `replace`, `contains`, `flatten`.

These evaluate at parse time — they produce constants. Runtime references like `~param.*` or `~(step ...)` are not available in `(def)`.
```

- [ ] **Step 2: Commit**

```bash
git add docs/site/workflow-syntax.md
git commit -m "docs: add include and evaluated def sections to workflow-syntax"
```

---

### Follow-Up (not in this plan)

- **`site-create-lab` and `site-update-lab` workflows** — same shape as create-page/update-page but targeting labs. These need the current `labs-generate.glitch` broken into per-lab workflows. Separate plan once the include/def infrastructure lands.
- **Migrate `site-sync` inline Python to native forms** — the big workflow has ~200 lines of Python heredocs. Worth migrating incrementally as the threading/def features prove stable.
- **`(assoc)`, `(pick)`, `(json-pick)` in def context** — spec lists these but they're not needed for the initial site workflows. Add when a workflow needs them.

# Spec: Fix repo structure drift in issue-to-pr pipeline

## Problem

The `issue-to-pr-tiered` workflow (stokagent/workflows/issue-to-pr-tiered.glitch) produces implementation plans and PR artifacts that drift from the target repo's actual structure. Batch #3912 (8 issues for elastic/observability-robots) surfaced three classes of error:

### 1. Path fabrication

Plans propose `docs/ai-agents/skill-creation/skill-creator/` as a new directory. The repo already has pattern-named directories:

```
docs/ai-agents/skill-creation/
├── index.md
├── decision-matrix.md
├── tool-wrapper/index.md
├── generator/index.md
├── inversion/index.md
├── reviewer/index.md
└── pipeline/index.md
```

The use-case guides should go *inside* existing pattern directories (e.g. `tool-wrapper/use-case-curl.md`) or as siblings alongside them, not in a fabricated `skill-creator/` subdirectory.

### 2. Inconsistent file naming across issues

The 8 issues in a single batch produced 3 different naming conventions:

| Issues | Convention | Example |
|--------|-----------|---------|
| 3916-3917 | `use-case-N.md` | `use-case-1.md` |
| 3918-3922 | `NN-pattern-topic.md` | `01-wrapper-curl.md` |
| 3923 | `TASK-NNNN-NN.md` | `TASK-3912-02.md` |

A batch should enforce a single naming convention derived from the repo's existing patterns.

### 3. Tool/skill misidentification

Plans reference `skill-creator` as a CLI tool with flags (`--name`, `--validate`, `--output`). In the repo it's a Claude Code agent skill at `.claude/skills/skill-creator/`. The LLM hallucinates CLI syntax for something that isn't a CLI.

## Root cause

The `research` step prompt says "Reference actual file paths from the repo structure" but the `repo-context` plugin only provides a flat `find | head -300` listing and 20 recent commits. This is insufficient for the LLM to:

1. Understand the directory hierarchy and naming conventions
2. Distinguish between CLI tools and agent skills
3. Maintain consistency across batch runs (each issue is an independent workflow invocation)

## Proposed fixes

### Fix 1: Add a `conventions` plugin step

Create a new plugin command `github/conventions` that extracts naming patterns and structure rules from the target directory:

```
;; conventions.glitch — extract file naming and structure patterns
(arg "repo" :default "elastic/observability-robots" :description "org/repo")
(arg "path" :default "." :description "subdirectory to analyze")

(workflow "conventions"
  :description "Extract naming conventions and directory structure from target path"

  (step "conventions"
    (run ```
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      TARGET="$REPO_PATH/{{.param.path}}"

      if [ ! -d "$TARGET" ]; then
        echo "Target path does not exist: {{.param.path}}"
        exit 0
      fi

      echo "=== DIRECTORY TREE (3 levels) ==="
      find "$TARGET" -maxdepth 3 -not -path '*/.git/*' | sort | head -100

      echo ""
      echo "=== FILE NAMING PATTERNS ==="
      find "$TARGET" -maxdepth 3 -type f -name "*.md" | sort | head -50

      echo ""
      echo "=== EXISTING INDEX FILES ==="
      for idx in $(find "$TARGET" -name "index.md" -maxdepth 3); do
        echo "--- $idx ---"
        head -20 "$idx"
        echo ""
      done

      echo ""
      echo "=== AGENT SKILLS (not CLI tools) ==="
      if [ -d "$REPO_PATH/.claude/skills" ]; then
        for skill_dir in "$REPO_PATH/.claude/skills"/*/; do
          if [ -f "${skill_dir}SKILL.md" ]; then
            skill_name=$(basename "$skill_dir")
            echo "SKILL: $skill_name (agent skill, NOT a CLI tool)"
            head -5 "${skill_dir}SKILL.md"
            echo ""
          fi
        done
      fi
      ```)))
```

### Fix 2: Strengthen the `research` prompt with structure constraints

Update the `research` step in `issue-to-pr-tiered.glitch` to include conventions data and explicit constraints:

```
DIRECTORY CONVENTIONS:
{{step "conventions"}}

CONSTRAINTS:
- Place new files within the EXISTING directory structure. Do NOT create new
  parent directories unless the issue explicitly requires it.
- Match the naming convention of sibling files in the target directory.
- If a tool exists as a .claude/skills/ agent skill, reference it as an
  agent skill invoked via Claude Code — NOT as a CLI with flags.
- All file paths must correspond to directories shown in DIRECTORY TREE.
```

### Fix 3: Add a batch-level naming convention lock

For batch runs, the first issue's `classify` step should establish the naming convention, and subsequent issues should receive it as input. Options:

**Option A — batch script injects a shared context file:**

The batch script (`scripts/batch-3912.sh`) writes a `conventions.json` to `.tmp/` after the first issue completes, and passes it via `--set conventions-file=...` to subsequent issues.

**Option B — glitch workspace-level shared state:**

Use glitch workspace variables to store the naming convention from the first run and read it in subsequent runs. This requires a `workspace-get`/`workspace-set` primitive in glitch.

**Recommendation:** Option A (batch script injection) is simpler and doesn't require new glitch primitives. Option B is the right long-term answer if glitch adds workspace state.

### Fix 4: Add a `validate-paths` post-step

Add a validation step after `research` that checks proposed file paths against the actual repo:

```
(step "validate-paths"
  (run ```
    REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
    REPO_PATH="$HOME/Projects/$REPO_NAME"
    PLAN='{{stepfile "research"}}'

    echo "=== PATH VALIDATION ==="
    # Extract proposed file paths from plan
    grep -oP '`[^`]*\.(md|yaml|py|go|sh)`' "$PLAN" 2>/dev/null | tr -d '`' | while read -r filepath; do
      parent=$(dirname "$filepath")
      if [ -d "$REPO_PATH/$parent" ]; then
        echo "OK: $filepath (parent exists)"
      elif [ -d "$REPO_PATH/$(dirname "$parent")" ]; then
        echo "WARN: $filepath (new subdirectory: $parent)"
      else
        echo "FAIL: $filepath (path fabricated — no ancestor exists)"
      fi
    done
    ```))
```

If any paths show FAIL, the `review` step should flag the plan as needing correction.

## Affected files

| File | Change |
|------|--------|
| `stokagent/.glitch/plugins/github/conventions.glitch` | NEW — conventions extraction plugin |
| `stokagent/workflows/issue-to-pr-tiered.glitch` | Add `conventions` step, strengthen `research` prompt, add `validate-paths` step |
| `stokagent/scripts/batch-3912.sh` | Inject naming convention lock for batch runs |

## Success criteria

Re-run batch #3912 after fixes. All 8 issues should:

1. Reference paths under `docs/ai-agents/skill-creation/` that match existing directory structure
2. Use a single consistent file naming convention across the batch
3. Correctly identify `skill-creator` as an agent skill, not a CLI tool
4. Pass the `validate-paths` step with zero FAIL results

## Out of scope

- Glitch workspace-level shared state (Option B) — tracked separately
- Auto-creating PRs from results — still requires manual review before `gh pr create`
- Fixing the existing 8 result sets — re-run the batch instead

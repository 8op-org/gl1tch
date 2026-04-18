#!/usr/bin/env python3
"""Gate 1: No hallucinated commands, keywords, or forms in docs."""
import sys, re
from pathlib import Path

DOCS_DIR = Path("site/src/content/docs")

VALID_COMMANDS = {
    "glitch run",
    "glitch workflow run", "glitch workflow list",
    "glitch wf list", "glitch plugin list", "glitch plugin",
    "glitch observe", "glitch up", "glitch down", "glitch index",
    "glitch config show", "glitch config set", "glitch config",
    "glitch version", "glitch --help", "glitch --workspace",
    "glitch workspace init", "glitch workspace use", "glitch workspace list",
    "glitch workspace status", "glitch workspace gui",
    "glitch workspace register", "glitch workspace unregister",
    "glitch workspace add", "glitch workspace rm",
    "glitch workspace sync", "glitch workspace pin",
    "glitch workspace workflow", "glitch workspace",
}

# Every valid keyword in the sexpr language
VALID_KEYWORDS = {
    ":prompt", ":provider", ":model", ":skill", ":format", ":tier",
    ":description", ":from", ":default", ":type", ":dir", ":headers",
    ":body", ":retries", ":version", ":flag", ":string", ":number",
    ":since", ":authored", ":reviewing", ":assigned",
    # workspace mechanics
    ":owner", ":url", ":ref", ":pin", ":path", ":repo", ":as",
    ":set", ":elasticsearch", ":fetched", ":active-workspace",
    # global config
    ":default-model", ":default-provider", ":eval-threshold",
    ":base-url", ":api-key-env",
    # compare / review forms
    ":id", ":criteria",
    # Elasticsearch forms
    ":index", ":query", ":size", ":fields", ":sort", ":ndjson",
    ":es", ":doc", ":upsert",
    # embed form
    ":input", ":embed", ":field",
    # websearch form
    ":websearch", ":engines", ":results", ":lang",
    # assoc / pick (key names used as keyword args)
    ":key", ":status", ":title", ":author",
    # workspace defaults / params
    ":env", ":providers",
    # openrouter free tier suffix (appears in config examples)
    ":free",
    # phase / gate / site manifest
    ":template", ":sections",
    # config.glitch keywords
    ":workflows-dir",
    # resource-state
    ":active",
    # plugin arg keyword (issue number)
    ":issue",
}

# Every valid form name
VALID_FORMS = {
    "def", "workflow", "step", "run", "llm", "save", "plugin", "include",
    "arg", "retry", "timeout", "catch", "cond", "map", "let",
    "phase", "gate", "par",
    "json-pick", "lines", "merge",
    "http-get", "http-post", "read-file", "write-file", "glob", "websearch",
    # aliases documented in workflow-syntax
    "fetch", "send", "write",
    # def-context transform forms
    "join", "split", "trim", "upper", "lower", "replace", "contains",
    # workspace mechanics
    "workspace", "workspaces", "defaults", "params", "repos",
    "resource", "resources", "resource-state", "map-resources",
    "call-workflow",
    # global config
    "config", "providers", "provider", "tiers", "tier", "state",
    # compare / branch
    "compare", "branch", "review",
    # threading / transforms
    "->", "filter", "reduce", "flatten", "embed", "assoc", "pick",
    "thread", "site",
    # Elasticsearch native forms
    "search", "index", "delete",
    # conditionals
    "when", "when-not",
    # iteration alias
    "each",
    # file read form
    "read",
    # call-workflow :set sub-forms
    "repo", "pr",
}

# English words / non-form tokens that can appear after "(" in prose, table cells,
# LLM prompt strings, or let-binding variable names
_ENGLISH_SKIP = frozenset({
    "or", "not", "and", "if", "the", "a", "an", "is",
    "no", "can", "has", "e", "in", "on", "at", "to",
    "respects", "requires", "returns", "creates",
    "supports", "expects", "like", "see", "used",
    "default", "optional", "only", "must", "exit", "exits",
    "one", "this", "else",
    "e.g", "i.e",
    # appear in table cell annotations
    "unquote", "escaped", "op_type", "unquote-rendered",
    # common words that appear in LLM prompt strings inside code blocks
    "features", "fixes", "chores", "themes", "under",
    # shell / comment references that appear in parentheses
    "shell", "bash",
    # common English words that appear in parenthetical prose inside code blocks
    "free", "expensive",
    # let-binding variable names that look like form names
    "token", "endpoint",
    # plugin subcommand namespace prefix — (name/sub ...) handled separately
})


def parse_md(path: Path) -> tuple[str, str]:
    """Return (slug, content_without_frontmatter) for a .md file."""
    text = path.read_text(encoding="utf-8")
    slug = path.stem
    # Strip YAML frontmatter between --- markers
    if text.startswith("---"):
        end = text.find("\n---", 3)
        if end != -1:
            text = text[end + 4:]
    return slug, text


def check_doc(slug: str, content: str) -> list[str]:
    errors = []
    in_code_block = False
    fence_char = ""   # the backtick sequence that opened the fence
    code_lang = ""

    for line in content.splitlines():
        stripped = line.strip()

        # Track fenced code blocks — handle 3 or 4 backtick fences
        # Match any sequence of 3+ backticks at the start of a line
        fence_match = re.match(r'^(`{3,})', stripped)
        if fence_match:
            fence_seq = fence_match.group(1)
            if not in_code_block:
                in_code_block = True
                fence_char = fence_seq
                code_lang = stripped[len(fence_seq):].strip().lower()
            elif fence_seq == fence_char:
                # Only close on same fence width
                in_code_block = False
                fence_char = ""
                code_lang = ""
            continue

        if not in_code_block:
            continue

        # ── CLI command check ─────────────────────────────────────────────
        if stripped.startswith("glitch ") or stripped.startswith("$ glitch "):
            cmd_part = stripped.lstrip("$ ").split("--")[0].strip()
            words = cmd_part.split()
            if len(words) >= 2:
                prefix = " ".join(words[:3])
                if not any(prefix.startswith(v) for v in VALID_COMMANDS):
                    errors.append(f"{slug}: invalid command: {cmd_part[:60]}")

        # ── Sexpr checks (lisp/glitch/clojure/scheme or unlabelled blocks) ─
        if code_lang in ("", "lisp", "glitch", "clojure", "scheme"):
            # Skip comment lines — these are not executable sexpr
            if stripped.startswith(";"):
                continue
            is_meta = any(w in stripped for w in ["form arg", "form-name", "keyword value"])
            if not is_meta:
                for kw in re.findall(r':[a-z][-a-z_]*', stripped):
                    if kw not in VALID_KEYWORDS:
                        errors.append(f"{slug}: invalid keyword {kw}")

                for form in re.findall(r'(?:^|(?<=\s))\(([a-z>][-a-z_>/]*)', stripped):
                    if form in _ENGLISH_SKIP:
                        continue
                    if form not in VALID_FORMS and not form.startswith("step"):
                        if "/" not in form:
                            errors.append(f"{slug}: invalid form ({form} ...)")

    return errors


def main():
    if not DOCS_DIR.exists():
        print(f"FAIL: docs directory not found: {DOCS_DIR}")
        sys.exit(1)

    md_files = sorted(DOCS_DIR.glob("*.md"))
    if not md_files:
        print(f"FAIL: no .md files found in {DOCS_DIR}")
        sys.exit(1)

    all_errors = []
    for path in md_files:
        slug, content = parse_md(path)
        all_errors.extend(check_doc(slug, content))

    # Deduplicate while preserving order
    seen = set()
    unique = []
    for e in all_errors:
        if e not in seen:
            seen.add(e)
            unique.append(e)

    if unique:
        print("FAIL")
        for e in unique:
            print(f"  {e}")
        sys.exit(1)
    else:
        print(f"PASS: no hallucinated commands, keywords, or forms ({len(md_files)} docs)")


if __name__ == "__main__":
    main()

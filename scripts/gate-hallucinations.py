#!/usr/bin/env python3
"""Gate 1: No hallucinated commands, keywords, or forms in generated docs."""
import json, sys, re

with open("site/generated/docs.json") as f:
    raw = f.read()
start = raw.index("[")
end = raw.rindex("]")
docs = json.loads(raw[start : end + 1])

VALID_COMMANDS = {
    "glitch ask", "glitch run", "glitch batch",
    "glitch workflow run", "glitch workflow list",  # legacy, still referenced in older docs
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
    ":since", ":authored", ":reviewing", ":assigned",  # common plugin args
    # workspace mechanics
    ":owner", ":url", ":ref", ":pin", ":path", ":repo", ":as",
    ":set", ":elasticsearch", ":fetched", ":active-workspace",
    # global config
    ":default-model", ":default-provider", ":eval-threshold",
    ":base-url", ":api-key-env",
}

# Every valid form name
VALID_FORMS = {
    "def", "workflow", "step", "run", "llm", "save", "plugin",
    "arg", "retry", "timeout", "catch", "cond", "map", "let",
    "phase", "gate", "par",
    "json-pick", "lines", "merge",
    "http-get", "http-post", "read-file", "write-file", "glob",
    # workspace mechanics
    "workspace", "workspaces", "defaults", "params", "repos",
    "resource", "resources", "resource-state", "map-resources",
    "call-workflow",
    # global config
    "config", "providers", "provider", "tiers", "tier", "state",
}

errors = []
for doc in docs:
    content = doc.get("content", "")
    in_code_block = False
    code_lang = ""
    for line in content.splitlines():
        stripped = line.strip()
        if stripped.startswith("```"):
            if in_code_block:
                in_code_block = False
                code_lang = ""
            else:
                in_code_block = True
                code_lang = stripped[3:].strip().lower()
            continue
        if not in_code_block:
            continue

        # Check glitch CLI commands
        if stripped.startswith("glitch ") or (stripped.startswith("$ glitch ") and len(stripped) > 9):
            cmd_part = stripped.lstrip("$ ").split("--")[0].strip()
            words = cmd_part.split()
            if len(words) >= 2:
                prefix = " ".join(words[:3])
                if not any(prefix.startswith(v) for v in VALID_COMMANDS):
                    errors.append(f'{doc["slug"]}: invalid command: {cmd_part[:60]}')

        # Check sexpr keywords in lisp/glitch/clojure code blocks
        if code_lang in ("", "lisp", "glitch", "clojure", "scheme"):
            is_meta = any(w in stripped for w in ["form arg", "form-name", "keyword value"])
            if not is_meta:
                for kw in re.findall(r':[a-z][-a-z_]*', stripped):
                    if kw not in VALID_KEYWORDS:
                        errors.append(f'{doc["slug"]}: invalid keyword {kw}')

            # Check form names — (form-name ...)
            # Skip lines that are clearly syntax documentation examples
            is_meta = any(w in stripped for w in ["form arg", "form-name", "keyword value"])
            if not is_meta:
                # Only match forms at start of line (sexpr style)
                for form in re.findall(r'(?:^|(?<=\s))\(([a-z][-a-z_]*)', stripped):
                    if form not in VALID_FORMS and not form.startswith("step"):
                        if "/" not in form:
                            # Skip English words that happen to follow a paren
                            if form in ("or", "not", "and", "if", "the", "a", "an", "is",
                                        "no", "can", "has", "e", "in", "on", "at", "to",
                                        "respects", "requires", "returns", "creates",
                                        "supports", "expects", "like", "see", "used",
                                        "default", "optional", "only", "must", "exit"):
                                continue
                            errors.append(f'{doc["slug"]}: invalid form ({form} ...)')

if errors:
    # Dedupe
    seen = set()
    unique = []
    for e in errors:
        if e not in seen:
            seen.add(e)
            unique.append(e)
    print("FAIL")
    for e in unique:
        print(f"  {e}")
    sys.exit(1)
else:
    print("PASS: no hallucinated commands, keywords, or forms")

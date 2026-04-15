#!/usr/bin/env python3
"""Read LLM-generated JSON from stdin, extract page, write as stub."""
import json, sys, re

raw = sys.stdin.read()

match = re.search(r'\{[\s\S]*\}', raw)
if not match:
    print("ERROR: no JSON found in LLM output", file=sys.stderr)
    sys.exit(1)

try:
    doc = json.loads(match.group())
except json.JSONDecodeError as e:
    print(f"ERROR: invalid JSON: {e}", file=sys.stderr)
    sys.exit(1)

slug = doc.get("slug", "untitled")
title = doc.get("title", slug.replace("-", " ").title())
content = doc.get("content", "")

path = f"docs/site/{slug}.md"
with open(path, "w") as f:
    f.write(f"# {title}\n\n{content}\n")
print(f"wrote stub: {path}")

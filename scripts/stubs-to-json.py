#!/usr/bin/env python3
"""Convert docs/site/*.md stubs to the JSON array build.sh expects."""
import os, json, glob

docs = []
order = 1
for path in sorted(glob.glob("docs/site/*.md")):
    slug = os.path.splitext(os.path.basename(path))[0]
    with open(path) as f:
        content = f.read()

    # Extract title from first H1
    title = slug.replace("-", " ").title()
    for line in content.splitlines():
        if line.startswith("# "):
            title = line[2:].strip()
            break

    # Strip the H1 from content (build.sh adds frontmatter with title)
    lines = content.splitlines()
    if lines and lines[0].startswith("# "):
        content = "\n".join(lines[1:]).lstrip("\n")

    # Description from first non-empty paragraph
    desc = ""
    for line in content.splitlines():
        stripped = line.strip()
        if stripped and not stripped.startswith("#") and not stripped.startswith("`"):
            desc = stripped[:120]
            break

    docs.append({
        "slug": slug,
        "title": title,
        "order": order,
        "description": desc,
        "content": content,
        "provenance": "stub \u2192 json \u2192 gate\u00d73 \u2192 astro  \u00b7  0 tokens",
    })
    order += 1

print(json.dumps(docs, indent=2))

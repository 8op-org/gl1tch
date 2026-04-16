#!/usr/bin/env python3
"""Parse LLM JSON output on stdin into site/generated/docs.json.

Falls back to raw stubs in docs/site/ if the JSON fails to parse or
doesn't have the required shape. Used by the site-update workflow so
step output never has to round-trip through a shell heredoc.
"""

import glob
import json
import os
import re
import sys


def load_fallback_pages():
    pages = []
    order = 1
    for path in sorted(glob.glob("docs/site/*.md")):
        slug = os.path.splitext(os.path.basename(path))[0]
        with open(path) as f:
            content = f.read()
        title = slug.replace("-", " ").title()
        for line in content.splitlines():
            if line.startswith("# "):
                title = line[2:].strip()
                break
        lines = content.splitlines()
        if lines and lines[0].startswith("# "):
            content = "\n".join(lines[1:]).lstrip("\n")
        desc = ""
        for line in content.splitlines():
            s = line.strip()
            if s and not s.startswith("#") and not s.startswith("`"):
                desc = s[:120]
                break
        pages.append(
            {
                "slug": slug,
                "title": title,
                "order": order,
                "description": desc,
                "content": content,
            }
        )
        order += 1
    return pages


def main():
    raw = sys.stdin.read()

    pages = []
    match = re.search(r"\[[\s\S]*\]", raw)
    if match:
        try:
            pages = json.loads(match.group())
        except json.JSONDecodeError:
            pages = []

    if not pages or not all("slug" in p and "content" in p for p in pages):
        print(
            "WARN: LLM output failed to parse, falling back to raw stubs",
            file=sys.stderr,
        )
        pages = load_fallback_pages()

    pages.sort(key=lambda p: p.get("order", 99))

    with open("site/generated/docs.json", "w") as f:
        json.dump(pages, f, indent=2)
    print(f"wrote {len(pages)} pages to site/generated/docs.json")


if __name__ == "__main__":
    main()

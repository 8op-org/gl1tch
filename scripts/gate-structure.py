#!/usr/bin/env python3
"""Gate 3: Valid JSON structure, no internals leaked, no empty docs."""
import json, sys

with open("site/generated/docs.json") as f:
    raw = f.read()
start = raw.index("[")
end = raw.rindex("]")
docs = json.loads(raw[start : end + 1])

errors = []
INTERNALS = [
    "BubbleTea", "bubbletea", "tea.Model", "tea.Cmd",
    "internal/tui", "lipgloss", "sqlite3", "SQLite",
]

for doc in docs:
    # Required fields
    for field in ["slug", "title", "order", "description", "content"]:
        if field not in doc:
            errors.append(f'{doc.get("slug", "?")}: missing field {field}')

    # No internals leaked
    content = doc.get("content", "")
    for term in INTERNALS:
        if term in content:
            errors.append(f'{doc["slug"]}: leaked internal: {term}')

    # Has actual content (not just frontmatter)
    if len(content) < 200:
        errors.append(f'{doc["slug"]}: suspiciously short ({len(content)} chars)')

if errors:
    print("FAIL")
    for e in errors:
        print(f"  {e}")
    sys.exit(1)
else:
    print(f"PASS: {len(docs)} docs valid, no internals leaked")

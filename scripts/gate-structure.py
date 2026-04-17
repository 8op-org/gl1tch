#!/usr/bin/env python3
"""Gate 3: Valid frontmatter, no internals leaked, no empty docs, "your" framing."""
import sys, re
from pathlib import Path

DOCS_DIR = Path("site/src/content/docs")

INTERNALS = [
    "BubbleTea", "bubbletea", "tea.Model", "tea.Cmd",
    "internal/tui", "lipgloss", "sqlite3", "SQLite",
]

REQUIRED_FRONTMATTER = ["title", "order", "description"]


def parse_frontmatter(text: str) -> tuple[dict, str]:
    """Return (frontmatter_dict, body_text). Frontmatter is between --- markers."""
    fm = {}
    body = text
    if text.startswith("---"):
        end = text.find("\n---", 3)
        if end != -1:
            fm_block = text[4:end]  # skip opening "---\n"
            body = text[end + 4:]
            for line in fm_block.splitlines():
                if ":" in line:
                    key, _, val = line.partition(":")
                    fm[key.strip()] = val.strip().strip('"')
    return fm, body


def main():
    if not DOCS_DIR.exists():
        print(f"FAIL: docs directory not found: {DOCS_DIR}")
        sys.exit(1)

    md_files = sorted(DOCS_DIR.glob("*.md"))
    if not md_files:
        print(f"FAIL: no .md files found in {DOCS_DIR}")
        sys.exit(1)

    errors = []
    warnings = []

    for path in md_files:
        slug = path.stem
        text = path.read_text(encoding="utf-8")
        fm, body = parse_frontmatter(text)

        # Required frontmatter fields
        for field in REQUIRED_FRONTMATTER:
            if field not in fm or not fm[field]:
                errors.append(f"{slug}: missing frontmatter field: {field}")

        # No leaked internals (check full file text)
        for term in INTERNALS:
            if term in text:
                errors.append(f"{slug}: leaked internal: {term}")

        # Content length — body only (exclude frontmatter)
        if len(body.strip()) < 200:
            errors.append(f"{slug}: suspiciously short body ({len(body.strip())} chars)")

        # "Your" framing — warn if "the user" appears
        if re.search(r'\bthe user\b', text, re.IGNORECASE):
            warnings.append(f"{slug}: found 'the user' — prefer 'you'/'your' framing")

    if warnings:
        print("WARN")
        for w in warnings:
            print(f"  [warn] {w}")

    if errors:
        print("FAIL")
        for e in errors:
            print(f"  {e}")
        sys.exit(1)
    else:
        count = len(md_files)
        warn_note = f", {len(warnings)} warning(s)" if warnings else ""
        print(f"PASS: {count} docs valid, no internals leaked{warn_note}")


if __name__ == "__main__":
    main()

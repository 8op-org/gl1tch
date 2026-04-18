#!/usr/bin/env python3
"""Gate: No stale interpolation syntax in docs or labs."""
import sys, re
from pathlib import Path

SCAN_DIRS = [
    Path("site/src/content/docs"),
    Path("site/src/content/labs"),
]

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

    fm_lines = 0
    if text.startswith("---"):
        end = text.find("\n---", 3)
        if end != -1:
            fm_lines = text[:end + 4].count("\n")
            text = text[end + 4:]

    for line_num, line in enumerate(text.splitlines(), start=fm_lines + 1):
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

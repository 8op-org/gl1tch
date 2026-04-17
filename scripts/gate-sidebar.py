#!/usr/bin/env python3
"""Gate: Sidebar sections/pages match manifest declarations in the correct order."""
import sys, re
from pathlib import Path

MANIFEST = Path("site-manifest.glitch")
SIDEBAR = Path("site/src/components/DocSidebar.astro")

# Manifest parsing: extract sections with their ordered pages
RE_SECTION_OPEN = re.compile(r'^\s*\(section\s+"([^"]+)"')
RE_PAGE_IN_SECTION = re.compile(r'^\s*\(page\s+"([^"]+)"')
RE_TITLE = re.compile(r':title\s+"([^"]+)"')

# Sidebar parsing: extract sections array from the frontmatter script block
RE_SIDEBAR_SECTION = re.compile(r'\{\s*label:\s*"([^"]+)"')
RE_SIDEBAR_SLUG = re.compile(r'\{\s*slug:\s*"([^"]+)"')


def parse_manifest_sections(text: str) -> list[tuple[str, list[str]]]:
    """
    Return ordered list of (section_label, [slug, ...]) from the manifest.
    Only includes pages that are inside (section ...) blocks (not top-level pages).
    """
    sections = []
    current_section = None
    current_pages: list[str] = []
    depth = 0  # paren depth within a section block

    for line in text.splitlines():
        section_match = RE_SECTION_OPEN.match(line)
        if section_match:
            # Save previous section if any
            if current_section is not None:
                sections.append((current_section, current_pages))
            current_section = section_match.group(1)
            current_pages = []
            depth = line.count("(") - line.count(")")
            continue

        if current_section is not None:
            depth += line.count("(") - line.count(")")
            page_match = RE_PAGE_IN_SECTION.match(line)
            if page_match:
                current_pages.append(page_match.group(1))
            if depth <= 0:
                sections.append((current_section, current_pages))
                current_section = None
                current_pages = []
                depth = 0

    # Flush last section if file ends without closing paren
    if current_section is not None and current_pages:
        sections.append((current_section, current_pages))

    return sections


def parse_sidebar_sections(text: str) -> list[tuple[str, list[str]]]:
    """
    Return ordered list of (section_label, [slug, ...]) from DocSidebar.astro.
    Parses the const sections = [...] array in the frontmatter script block.
    """
    # Extract the script frontmatter (between --- ... ---)
    fm_match = re.search(r'^---\n(.*?)^---', text, re.MULTILINE | re.DOTALL)
    if not fm_match:
        return []
    script = fm_match.group(1)

    # Find the sections array: everything between "const sections = [" and the matching "]"
    arr_start = script.find("const sections = [")
    if arr_start == -1:
        return []
    arr_start = script.index("[", arr_start)
    # Find matching closing bracket
    depth = 0
    arr_end = arr_start
    for i, ch in enumerate(script[arr_start:], arr_start):
        if ch == "[":
            depth += 1
        elif ch == "]":
            depth -= 1
            if depth == 0:
                arr_end = i
                break
    sections_str = script[arr_start: arr_end + 1]

    # Split into per-section chunks by finding "{ label:" patterns
    sections: list[tuple[str, list[str]]] = []
    label_positions = [m.start() for m in RE_SIDEBAR_SECTION.finditer(sections_str)]

    for idx, pos in enumerate(label_positions):
        label_match = RE_SIDEBAR_SECTION.search(sections_str, pos)
        if not label_match:
            continue
        label = label_match.group(1)

        # Slug region: between this label position and the next label (or end)
        region_end = label_positions[idx + 1] if idx + 1 < len(label_positions) else len(sections_str)
        region = sections_str[pos:region_end]

        slugs = [m.group(1) for m in RE_SIDEBAR_SLUG.finditer(region)]
        sections.append((label, slugs))

    return sections


def main():
    errors = []

    if not MANIFEST.exists():
        print(f"FAIL: manifest not found: {MANIFEST}")
        sys.exit(1)
    if not SIDEBAR.exists():
        print(f"FAIL: sidebar not found: {SIDEBAR}")
        sys.exit(1)

    manifest_text = MANIFEST.read_text(encoding="utf-8")
    sidebar_text = SIDEBAR.read_text(encoding="utf-8")

    manifest_sections = parse_manifest_sections(manifest_text)
    sidebar_sections = parse_sidebar_sections(sidebar_text)

    if not manifest_sections:
        print("FAIL: no sections parsed from manifest")
        sys.exit(1)
    if not sidebar_sections:
        print("FAIL: no sections parsed from DocSidebar.astro")
        sys.exit(1)

    # Check section count
    if len(manifest_sections) != len(sidebar_sections):
        errors.append(
            f"section count mismatch: manifest has {len(manifest_sections)}, "
            f"sidebar has {len(sidebar_sections)}"
        )
        # Still compare what we can
        count = min(len(manifest_sections), len(sidebar_sections))
    else:
        count = len(manifest_sections)

    for i in range(count):
        m_label, m_slugs = manifest_sections[i]
        s_label, s_slugs = sidebar_sections[i]

        if m_label != s_label:
            errors.append(
                f"section {i + 1} label mismatch: "
                f"manifest '{m_label}' vs sidebar '{s_label}'"
            )

        if m_slugs != s_slugs:
            m_set = set(m_slugs)
            s_set = set(s_slugs)
            missing = m_set - s_set
            extra = s_set - m_set
            if missing:
                errors.append(
                    f"section '{m_label}': pages in manifest but missing from sidebar: "
                    + ", ".join(sorted(missing))
                )
            if extra:
                errors.append(
                    f"section '{m_label}': pages in sidebar but not in manifest: "
                    + ", ".join(sorted(extra))
                )
            # Also check ordering if sets match
            if not missing and not extra and m_slugs != s_slugs:
                errors.append(
                    f"section '{m_label}': page order differs — "
                    f"manifest: {m_slugs}, sidebar: {s_slugs}"
                )

    # Check for manifest sections beyond what sidebar covers
    for j in range(count, len(manifest_sections)):
        label, _ = manifest_sections[j]
        errors.append(f"manifest section '{label}' missing from sidebar entirely")

    if errors:
        print("FAIL")
        for e in errors:
            print(f"  {e}")
        sys.exit(1)
    else:
        total_pages = sum(len(slugs) for _, slugs in manifest_sections)
        print(
            f"PASS: sidebar matches manifest "
            f"({len(manifest_sections)} sections, {total_pages} pages)"
        )


if __name__ == "__main__":
    main()

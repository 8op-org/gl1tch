#!/usr/bin/env python3
"""
site-read-manifest.py — parse site-manifest.glitch into JSON.

Output JSON structure:
{
  "site": "gl1tch",
  "url": "https://...",
  "homepage": {
    "slug": "index",
    "title": "...",
    "template": "homepage",
    "sections": [...],
    "context_query": "...",
    "context_paths": [...]
  },
  "sections": [
    {
      "label": "Section Name",
      "pages": [
        {
          "slug": "page-slug",
          "title": "...",
          "order": N,
          "description": "...",
          "context_query": "...",
          "context_paths": [...]
        },
        ...
      ]
    },
    ...
  ],
  "all_slugs": ["index", "getting-started", ...]
}
"""
import sys, re, json
from pathlib import Path

MANIFEST = Path("site-manifest.glitch")


def parse_string(s: str) -> str:
    """Strip surrounding quotes from a manifest string token."""
    s = s.strip()
    if (s.startswith('"') and s.endswith('"')) or \
       (s.startswith("'") and s.endswith("'")):
        return s[1:-1]
    return s


def parse_list(text: str) -> list[str]:
    """
    Parse a Lisp-style list  ("a" "b" "c")  into a Python list of strings.
    Does not handle nested parens — only used for flat string lists.
    """
    inner = re.sub(r'^\s*\(\s*', '', text)
    inner = re.sub(r'\s*\)\s*$', '', inner)
    return [parse_string(t) for t in re.findall(r'"[^"]*"', inner)]


# ── per-line attribute regexes ────────────────────────────────────────────────
RE_SITE        = re.compile(r'^\(site\s+"([^"]+)"')
RE_URL         = re.compile(r':url\s+"([^"]+)"')
RE_SECTION     = re.compile(r'^\s*\(section\s+"([^"]+)"')
RE_SIDEBAR     = re.compile(r':sidebar\s+(true|false)')
RE_PAGE        = re.compile(r'^\s*\(page\s+"([^"]+)"')
RE_TITLE       = re.compile(r':title\s+"([^"]+)"')
RE_TEMPLATE    = re.compile(r':template\s+"([^"]+)"')
RE_ORDER       = re.compile(r':order\s+(\d+)')
RE_DESC        = re.compile(r':description\s+"([^"]+)"')
RE_CQ          = re.compile(r':context-query\s+"([^"]+)"')
RE_CP          = re.compile(r':context-paths\s+(\([^)]*\))')


def parse_manifest(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    lines = text.splitlines()

    result = {
        "site": "",
        "url": "",
        "homepage": None,
        "sections": [],
        "all_slugs": [],
    }

    # ── state machine ─────────────────────────────────────────────────────────
    # We walk line-by-line, accumulating multi-line page/section blocks.
    # depth tracks open parens so we know when a block closes.

    in_section = False
    in_page    = False
    section_depth = 0
    page_depth    = 0

    current_section: dict | None = None
    current_page:    dict | None = None

    def flush_page():
        nonlocal in_page, current_page, page_depth
        if current_page is None:
            return
        slug = current_page["slug"]
        result["all_slugs"].append(slug)
        if current_page.get("template") == "homepage" or slug == "index":
            result["homepage"] = current_page
        elif in_section and current_section is not None:
            current_section["pages"].append(current_page)
        current_page  = None
        in_page       = False
        page_depth    = 0

    def flush_section():
        nonlocal in_section, current_section, section_depth
        if current_section is None:
            return
        result["sections"].append(current_section)
        current_section = None
        in_section      = False
        section_depth   = 0

    for line in lines:
        # ── site declaration ─────────────────────────────────────────────────
        m = RE_SITE.match(line)
        if m:
            result["site"] = m.group(1)
            continue

        # ── site-level :url ──────────────────────────────────────────────────
        m = RE_URL.search(line)
        if m and not in_section and not in_page:
            result["url"] = m.group(1)
            continue

        # ── section open ─────────────────────────────────────────────────────
        m = RE_SECTION.match(line)
        if m:
            flush_page()
            flush_section()
            in_section      = True
            section_depth   = line.count("(") - line.count(")")
            sidebar = True
            sm = RE_SIDEBAR.search(line)
            if sm:
                sidebar = sm.group(1) == "true"
            current_section = {"label": m.group(1), "pages": [], "sidebar": sidebar}
            continue

        # ── page open ────────────────────────────────────────────────────────
        m = RE_PAGE.match(line)
        if m:
            flush_page()
            in_page    = True
            page_depth = line.count("(") - line.count(")")
            current_page = {
                "slug":          m.group(1),
                "title":         "",
                "order":         0,
                "description":   "",
                "context_query": "",
                "context_paths": [],
                "template":      "",
                "sections":      [],
            }
            # Attributes may be on the same line
            for attr, regex, cast in [
                ("title",         RE_TITLE,    str),
                ("template",      RE_TEMPLATE, str),
                ("order",         RE_ORDER,    int),
                ("description",   RE_DESC,     str),
                ("context_query", RE_CQ,       str),
            ]:
                am = regex.search(line)
                if am:
                    current_page[attr] = cast(am.group(1))
            cpm = RE_CP.search(line)
            if cpm:
                current_page["context_paths"] = parse_list(cpm.group(1))
            continue

        # ── inside a page block ──────────────────────────────────────────────
        if in_page and current_page is not None:
            page_depth += line.count("(") - line.count(")")

            for attr, regex, cast in [
                ("title",         RE_TITLE,    str),
                ("template",      RE_TEMPLATE, str),
                ("order",         RE_ORDER,    int),
                ("description",   RE_DESC,     str),
                ("context_query", RE_CQ,       str),
            ]:
                am = regex.search(line)
                if am:
                    current_page[attr] = cast(am.group(1))

            # :sections ("hero" "features" ...)
            sm = re.search(r':sections\s+(\([^)]*\))', line)
            if sm:
                current_page["sections"] = parse_list(sm.group(1))

            cpm = RE_CP.search(line)
            if cpm:
                current_page["context_paths"] = parse_list(cpm.group(1))

            if page_depth <= 0:
                flush_page()
            continue

        # ── inside a section block (but not a page) ──────────────────────────
        if in_section and current_section is not None:
            section_depth += line.count("(") - line.count(")")
            if section_depth <= 0:
                flush_section()

    # Flush anything still open at EOF
    flush_page()
    flush_section()

    return result


def main():
    if not MANIFEST.exists():
        print(f"ERROR: manifest not found: {MANIFEST}", file=sys.stderr)
        sys.exit(1)

    data = parse_manifest(MANIFEST)
    print(json.dumps(data, indent=2))


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Gate: Internal /docs/ links resolve to declared page slugs in site-manifest.glitch."""
import sys, re
from pathlib import Path

DOCS_DIR = Path("site/src/content/docs")
MANIFEST = Path("site-manifest.glitch")
INDEX_ASTRO = Path("site/src/pages/index.astro")
BASE_ASTRO = Path("site/src/layouts/Base.astro")

# Pattern: (page "slug-name" ...) in the manifest
RE_PAGE_SLUG = re.compile(r'\(page\s+"([^"]+)"')

# Pattern: href="/docs/slug" or href='/docs/slug' in .astro files,
# and [text](/docs/slug) in markdown
RE_MD_LINK = re.compile(r'\[([^\]]+)\]\(/docs/([^)#\s]+)\)')
RE_ASTRO_LINK = re.compile(r'href=["\']\/docs\/([^"\'#\s]+)["\']')


def load_manifest_slugs(manifest_path: Path) -> set[str]:
    """Extract all page slugs declared in the manifest."""
    if not manifest_path.exists():
        return set()
    text = manifest_path.read_text(encoding="utf-8")
    return set(RE_PAGE_SLUG.findall(text))


def collect_links_from_md(path: Path) -> list[tuple[str, str]]:
    """Return list of (source_label, target_slug) for /docs/ links in a .md file."""
    text = path.read_text(encoding="utf-8")
    return [(m.group(1), m.group(2)) for m in RE_MD_LINK.finditer(text)]


def collect_links_from_astro(path: Path) -> list[tuple[str, str]]:
    """Return list of (source_file, target_slug) for /docs/ hrefs in an .astro file."""
    if not path.exists():
        return []
    text = path.read_text(encoding="utf-8")
    return [(path.name, m.group(1)) for m in RE_ASTRO_LINK.finditer(text)]


def main():
    errors = []

    # Load manifest slugs
    if not MANIFEST.exists():
        print(f"FAIL: manifest not found: {MANIFEST}")
        sys.exit(1)

    valid_slugs = load_manifest_slugs(MANIFEST)
    if not valid_slugs:
        print(f"FAIL: no page slugs found in {MANIFEST}")
        sys.exit(1)

    # Check docs .md files
    if not DOCS_DIR.exists():
        print(f"FAIL: docs directory not found: {DOCS_DIR}")
        sys.exit(1)

    md_files = sorted(DOCS_DIR.glob("*.md"))
    for path in md_files:
        slug = path.stem
        for link_text, target in collect_links_from_md(path):
            if target not in valid_slugs:
                errors.append(
                    f"{slug}: broken link [{link_text}](/docs/{target}) — "
                    f"'{target}' not in manifest"
                )

    # Check index.astro and Base.astro
    for astro_path in (INDEX_ASTRO, BASE_ASTRO):
        for source, target in collect_links_from_astro(astro_path):
            if target not in valid_slugs:
                errors.append(
                    f"{source}: broken link /docs/{target} — "
                    f"'{target}' not in manifest"
                )

    if errors:
        print("FAIL")
        for e in errors:
            print(f"  {e}")
        sys.exit(1)
    else:
        print(
            f"PASS: all /docs/ links valid "
            f"({len(md_files)} docs, {len(valid_slugs)} manifest slugs)"
        )


if __name__ == "__main__":
    main()

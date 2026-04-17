#!/usr/bin/env python3
"""
site-diff-disk.py — compare manifest vs site/src/content/docs/*.md

Reads manifest JSON from stdin (output of site-read-manifest.py) and
compares against the files on disk.

Output JSON:
{
  "create": [
    {"slug": "...", "title": "...", "order": N, "description": "...",
     "context_query": "...", "context_paths": [...], "disk_path": null}
  ],
  "update": [
    {"slug": "...", "title": "...", "order": N, "description": "...",
     "context_query": "...", "context_paths": [...],
     "disk_path": "site/src/content/docs/slug.md",
     "existing_content": "..."}
  ],
  "ok": ["slug1", "slug2"],
  "orphan": ["slug-on-disk-not-in-manifest"]
}

An "update" is triggered when:
  - The manifest title/description/order differs from the page's frontmatter, OR
  - Context-path files have been modified more recently than the .md file
    (upstream code changed — page may be stale).

A "create" is triggered when the manifest declares a page that has no .md file.
An "orphan" is a .md file on disk with no corresponding manifest entry.
"""
import sys, json
from pathlib import Path

DOCS_DIR = Path("site/src/content/docs")

# ── helpers ───────────────────────────────────────────────────────────────────

def read_frontmatter(text: str) -> dict:
    """Return dict of frontmatter key→value from a markdown file."""
    fm = {}
    if not text.startswith("---"):
        return fm
    end = text.find("\n---", 3)
    if end == -1:
        return fm
    block = text[4:end]
    for line in block.splitlines():
        if ":" in line:
            key, _, val = line.partition(":")
            fm[key.strip()] = val.strip().strip('"')
    return fm


def mtime(path: Path) -> float:
    """Return mtime as a float; 0.0 if path does not exist."""
    try:
        return path.stat().st_mtime
    except FileNotFoundError:
        return 0.0


def context_path_mtime(context_paths: list[str]) -> float:
    """
    Return the newest mtime across all files/dirs in context_paths.
    Glob-expands dirs (shallow — first level of .go files).
    """
    newest = 0.0
    for cp in context_paths:
        p = Path(cp)
        if p.is_dir():
            for f in p.rglob("*.go"):
                newest = max(newest, mtime(f))
        elif p.exists():
            newest = max(newest, mtime(p))
        else:
            # Glob pattern like cmd/workspace*.go
            for f in Path(".").glob(cp):
                newest = max(newest, mtime(f))
    return newest


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    manifest = json.load(sys.stdin)

    # Build a flat list of all non-homepage pages from the manifest.
    # Homepage (index) is handled separately via sync-homepage step.
    manifest_pages: dict[str, dict] = {}
    for section in manifest.get("sections", []):
        # Skip non-sidebar sections (e.g. labs) — they have their own pipeline
        if not section.get("sidebar", True):
            continue
        for page in section.get("pages", []):
            manifest_pages[page["slug"]] = page

    # Collect on-disk slugs
    disk_slugs: set[str] = set()
    if DOCS_DIR.exists():
        for md_path in DOCS_DIR.glob("*.md"):
            disk_slugs.add(md_path.stem)

    create: list[dict] = []
    update: list[dict] = []
    ok:     list[str]  = []
    orphan: list[str]  = []

    for slug, page in manifest_pages.items():
        md_path = DOCS_DIR / f"{slug}.md"

        if not md_path.exists():
            # Page declared in manifest but not on disk → create
            entry = dict(page)
            entry["disk_path"] = None
            create.append(entry)
            continue

        text = md_path.read_text(encoding="utf-8")
        fm   = read_frontmatter(text)

        # Compare frontmatter fields
        fm_title = fm.get("title", "").strip('"')
        fm_order = fm.get("order", "")
        fm_desc  = fm.get("description", "").strip('"')

        try:
            fm_order_int = int(fm_order)
        except (ValueError, TypeError):
            fm_order_int = -1

        frontmatter_changed = (
            fm_title != page.get("title", "") or
            fm_order_int != page.get("order", 0) or
            fm_desc != page.get("description", "")
        )

        # Check if context sources are newer than the doc
        doc_mtime     = mtime(md_path)
        context_mtime = context_path_mtime(page.get("context_paths", []))
        code_changed  = context_mtime > doc_mtime

        if frontmatter_changed or code_changed:
            entry = dict(page)
            entry["disk_path"]        = str(md_path)
            entry["existing_content"] = text
            entry["reason"] = []
            if frontmatter_changed:
                entry["reason"].append("frontmatter_changed")
            if code_changed:
                entry["reason"].append("code_changed")
            update.append(entry)
        else:
            ok.append(slug)

    # Orphans: on disk but not in manifest (and not "index")
    for slug in sorted(disk_slugs):
        if slug not in manifest_pages and slug != "index":
            orphan.append(slug)

    result = {
        "create": create,
        "update": update,
        "ok":     ok,
        "orphan": orphan,
    }

    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()

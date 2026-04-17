#!/usr/bin/env python3
"""
site-build-sidebar.py — generate DocSidebar.astro from site-manifest.glitch.

Reads site-manifest.glitch directly and writes site/src/components/DocSidebar.astro.
This is a pure deterministic transform — no LLM involved.

Usage:
    python3 scripts/site-build-sidebar.py

Exits 0 on success; non-zero on error.
"""
import sys
from pathlib import Path

MANIFEST_PATH = Path("site-manifest.glitch")
SIDEBAR_PATH  = Path("site/src/components/DocSidebar.astro")

# ── import the manifest parser from the same scripts dir ─────────────────────
import importlib.util
_spec = importlib.util.spec_from_file_location(
    "site_read_manifest",
    Path(__file__).parent / "site-read-manifest.py",
)
if _spec is None or _spec.loader is None:
    print("FAIL: cannot load site-read-manifest.py", file=sys.stderr)
    sys.exit(1)
_mod = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_mod)
parse_manifest = _mod.parse_manifest


def build_sidebar(manifest: dict) -> str:
    """Render DocSidebar.astro from parsed manifest data."""
    sections = manifest["sections"]

    lines = []
    lines.append("---")
    lines.append("const { currentSlug } = Astro.props;")
    lines.append("const sections = [")

    for i, section in enumerate(sections):
        comma_after = "," if i < len(sections) - 1 else ""
        label = section["label"].replace('"', '\\"')
        lines.append(f'  {{ label: "{label}", pages: [')
        pages = section["pages"]
        for j, page in enumerate(pages):
            comma_page = "," if j < len(pages) - 1 else ""
            slug  = page["slug"].replace('"', '\\"')
            title = page["title"].replace('"', '\\"')
            lines.append(f'    {{ slug: "{slug}", title: "{title}" }}{comma_page}')
        lines.append(f'  ]}}{comma_after}')

    lines.append("];")
    lines.append("---")
    lines.append('<nav class="doc-sidebar">')
    lines.append("  {sections.map(s => (")
    lines.append('    <div class="sidebar-section">')
    lines.append('      <div class="sidebar-label">{s.label}</div>')
    lines.append("      {s.pages.map(p => (")
    lines.append('        <a href={`/docs/${p.slug}`}')
    lines.append('           class:list={["sidebar-link", { active: currentSlug === p.slug }]}>')
    lines.append("          {p.title}")
    lines.append("        </a>")
    lines.append("      ))}")
    lines.append("    </div>")
    lines.append("  ))}")
    lines.append("</nav>")

    return "\n".join(lines) + "\n"


def main():
    if not MANIFEST_PATH.exists():
        print(f"ERROR: manifest not found: {MANIFEST_PATH}", file=sys.stderr)
        sys.exit(1)

    manifest = parse_manifest(MANIFEST_PATH)

    if not manifest["sections"]:
        print("ERROR: no sections found in manifest", file=sys.stderr)
        sys.exit(1)

    sidebar_content = build_sidebar(manifest)

    SIDEBAR_PATH.parent.mkdir(parents=True, exist_ok=True)
    SIDEBAR_PATH.write_text(sidebar_content, encoding="utf-8")

    # Count for reporting
    total_pages = sum(len(s["pages"]) for s in manifest["sections"])
    print(
        f"OK: wrote {SIDEBAR_PATH} "
        f"({len(manifest['sections'])} sections, {total_pages} pages)"
    )


if __name__ == "__main__":
    main()

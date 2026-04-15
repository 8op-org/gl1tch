#!/usr/bin/env bash
set -euo pipefail

SITE_DIR="$(cd "$(dirname "$0")" && pwd)"
GENERATED="$SITE_DIR/generated"
CONTENT_DOCS="$SITE_DIR/src/content/docs"
CONTENT_CHANGELOG="$SITE_DIR/src/content/changelog"

# ── Check generated output exists ────────────────
if [[ ! -f "$GENERATED/docs.json" ]]; then
  echo "ERROR: $GENERATED/docs.json not found. Run the site-update workflow first."
  exit 1
fi

if [[ ! -f "$GENERATED/build-report.md" ]]; then
  echo "ERROR: $GENERATED/build-report.md not found. Run the site-update workflow first."
  exit 1
fi

# ── Gate: check diff-review passed ───────────────
if grep -q "^PASS" "$GENERATED/build-report.md"; then
  VERDICT="PASS"
else
  VERDICT="FAIL"
fi
if [[ "$VERDICT" != "PASS" ]]; then
  echo "ERROR: Diff-review did not pass."
  echo "Verdict: $VERDICT"
  echo "See: $GENERATED/build-report.md"
  exit 1
fi
echo "diff-review: PASS"

# ── Split docs JSON into individual markdown files ──
rm -rf "$CONTENT_DOCS"/*.md
node -e "
// Extract JSON array from potentially noisy provider output
const raw = require('fs').readFileSync('$GENERATED/docs.json', 'utf8');
const start = raw.indexOf('[');
const end = raw.lastIndexOf(']');
if (start === -1 || end === -1) { console.error('ERROR: no JSON array found in docs.json'); process.exit(1); }
const docs = JSON.parse(raw.slice(start, end + 1));
for (const doc of docs) {
  const fm = [
    '---',
    'title: \"' + doc.title.replace(/\"/g, '\\\\\"') + '\"',
    'order: ' + doc.order,
    'description: \"' + doc.description.replace(/\"/g, '\\\\\"') + '\"',
    '---',
    '',
    doc.content
  ].join('\n');
  require('fs').writeFileSync('$CONTENT_DOCS/' + doc.slug + '.md', fm);
  console.log('  wrote: ' + doc.slug + '.md');
}
"
echo "docs: split into content files"

# ── Write changelog ──────────────────────────────
TODAY=$(date +%Y-%m-%d)
cat > "$CONTENT_CHANGELOG/$TODAY.md" << HEADER
---
title: "Update $TODAY"
date: "$TODAY"
---

$(cat "$GENERATED/changelog.md")
HEADER
echo "changelog: wrote $TODAY.md"

# ── Build ────────────────────────────────────────
cd "$SITE_DIR"
npx astro build
echo "astro: build succeeded"

echo ""
echo "Site built to $SITE_DIR/dist/"
echo "Preview: cd $SITE_DIR && npx astro preview"

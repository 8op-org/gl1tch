#!/usr/bin/env bash
# Gate: run Playwright page tests against built site
# Starts astro preview on :4322, runs tests, stops server
set -euo pipefail

cd site

# Build first if dist doesn't exist
if [ ! -d dist ]; then
  npx astro build 2>&1 | tail -3
fi

# Run tests (playwright config handles the dev server)
npx playwright test --reporter=line 2>&1
echo "PASS: all playwright tests passed"

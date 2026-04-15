#!/bin/bash
# Build the Svelte frontend and compile into the glitch binary via go:embed.
set -euo pipefail

cd "$(dirname "$0")/.."

echo ">> Building frontend..."
cd gui
npm ci --silent 2>/dev/null || npm install --silent
npm run build
cd ..

echo ">> Copying dist to internal/gui/dist..."
rm -rf internal/gui/dist
cp -r gui/dist internal/gui/dist

echo ">> Building glitch binary..."
go build -o glitch .

echo ">> Done: ./glitch ($(du -h glitch | cut -f1) with embedded GUI)"

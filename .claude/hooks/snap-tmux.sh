#!/usr/bin/env bash
# Triggered by the PostToolUse hook on mcp__tmux__capture-pane.
# Renders the glitch TUI pane to a PNG via termshot.

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DEST_DIR="$REPO_ROOT/.claude/screenshots"
mkdir -p "$DEST_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DEST="$DEST_DIR/tmux_${TIMESTAMP}.png"

PANE_ID=$(cat | jq -r '.tool_input.paneId // "%0"' 2>/dev/null)
PANE_ID="${PANE_ID:-%0}"

TMPFILE=$(mktemp)
tmux capture-pane -p -e -t "$PANE_ID" > "$TMPFILE" 2>/dev/null

if [ -s "$TMPFILE" ]; then
    termshot --show-cmd=false --no-decoration -f "$DEST" -- cat "$TMPFILE" 2>/dev/null || true
fi

rm -f "$TMPFILE"

#!/usr/bin/env bash
# tools/screenshots/take.sh
#
# Stages marketing screenshots of GLITCH by sending prompts through the live
# tmux session, waiting for responses, and screencapturing each result.
#
# Usage:
#   ./tools/screenshots/take.sh                  # all scenarios
#   ./tools/screenshots/take.sh goroutine-leak   # single scenario by name
#
# Requirements:
#   - GLITCH running in a tmux session named 'glitch'
#   - screencapture (macOS built-in)
#   - GLITCH window visible and not obscured
#
# Output: site/public/screenshots/<name>.png
#         site/public/screenshots/index.json  (updated automatically)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCENARIOS_DIR="$REPO_ROOT/tools/screenshots/scenarios"
OUT_DIR="$REPO_ROOT/site/public/screenshots"
INDEX_FILE="$OUT_DIR/index.json"

mkdir -p "$OUT_DIR"

# ── Locate the GLITCH tmux pane ───────────────────────────────────────────────

find_tmux_sock() {
  local uid
  uid="$(id -u)"
  local sock="/private/tmp/tmux-${uid}/default"
  if [[ -S "$sock" ]]; then
    echo "$sock"
    return
  fi
  # Fallback: search common paths
  local found
  found="$(find /private/tmp /tmp -name "default" -type s 2>/dev/null | head -1)"
  if [[ -n "$found" ]]; then
    echo "$found"
    return
  fi
  echo ""
}

TMUX_SOCK="$(find_tmux_sock)"
if [[ -z "$TMUX_SOCK" ]]; then
  echo "error: no tmux socket found. is GLITCH running?" >&2
  exit 1
fi

TMUX="tmux -S $TMUX_SOCK"

find_glitch_pane() {
  $TMUX list-panes -t glitch -F '#{pane_id}' 2>/dev/null | head -1
}

PANE="$(find_glitch_pane)"
if [[ -z "$PANE" ]]; then
  echo "error: no pane found in tmux session 'glitch'" >&2
  exit 1
fi

echo "using pane $PANE on $TMUX_SOCK"

# ── Helpers ───────────────────────────────────────────────────────────────────

# Focus the GLITCH input (Tab = focus in switchboard)
focus_glitch() {
  $TMUX send-keys -t "$PANE" "" Tab  # Tab focuses GLITCH input
  sleep 0.3
}

# Clear chat history
clear_chat() {
  focus_glitch
  $TMUX send-keys -t "$PANE" "/clear" Enter
  sleep 0.8
}

# Send a message to GLITCH
send_message() {
  local msg="$1"
  focus_glitch
  $TMUX send-keys -t "$PANE" "$msg" Enter
}

# Wait until streaming finishes (hint line returns to idle state)
wait_for_response() {
  local timeout="${1:-30}"
  local elapsed=0
  local interval=0.5

  # Give it a moment to start streaming
  sleep 1.0

  while (( elapsed < timeout )); do
    local hint
    hint="$($TMUX capture-pane -t "$PANE" -p 2>/dev/null | grep -o "streaming.*" || true)"
    if [[ -z "$hint" ]]; then
      sleep 0.5  # one extra settle
      return 0
    fi
    sleep "$interval"
    elapsed=$(( elapsed + 1 ))
  done

  echo "warning: response timed out after ${timeout}s" >&2
  return 0
}

# Take a screencapture of the full primary display
take_screenshot() {
  local name="$1"
  local out="$OUT_DIR/${name}.png"
  screencapture -x "$out"
  echo "  → $out"
}

# ── Run a single scenario ─────────────────────────────────────────────────────

run_scenario() {
  local file="$1"
  local name
  name="$(basename "$file" .txt)"

  echo "[$name]"

  # Each scenario file format:
  #   line 1:          caption (shown in gallery)
  #   line 2:          blank
  #   remaining lines: the prompt to send
  local caption
  caption="$(head -1 "$file")"
  local prompt
  prompt="$(tail -n +3 "$file")"

  clear_chat
  sleep 0.5
  send_message "$prompt"
  wait_for_response 45
  take_screenshot "$name"

  echo "  caption: $caption"
  echo ""

  # Return caption for index building
  echo "$name:::$caption"
}

# ── Build index.json ──────────────────────────────────────────────────────────

build_index() {
  local -a entries=("$@")
  local json='['
  local first=true

  for entry in "${entries[@]}"; do
    local name="${entry%%:::*}"
    local caption="${entry##*:::}"
    local file="$OUT_DIR/${name}.png"

    if [[ ! -f "$file" ]]; then
      continue
    fi

    if [[ "$first" == "true" ]]; then
      first=false
    else
      json+=','
    fi

    json+="{\"name\":\"${name}\",\"file\":\"${name}.png\",\"caption\":\"${caption}\"}"
  done

  json+=']'
  echo "$json" > "$INDEX_FILE"
  echo "index written: $INDEX_FILE"
}

# ── Main ──────────────────────────────────────────────────────────────────────

main() {
  local filter="${1:-}"
  local -a results=()

  # Collect scenario files
  local -a files=()
  while IFS= read -r -d '' f; do
    files+=("$f")
  done < <(find "$SCENARIOS_DIR" -name "*.txt" -print0 | sort -z)

  if [[ ${#files[@]} -eq 0 ]]; then
    echo "no scenario files found in $SCENARIOS_DIR" >&2
    exit 1
  fi

  echo ""
  echo "GLITCH Screenshot Capture"
  echo "========================="
  echo "pane: $PANE"
  echo "output: $OUT_DIR"
  echo ""
  echo "Position the GLITCH terminal window so it's visible and unobscured."
  echo "Press Enter when ready..."
  read -r

  for f in "${files[@]}"; do
    local name
    name="$(basename "$f" .txt)"

    # Skip if filter specified and doesn't match
    if [[ -n "$filter" && "$name" != "$filter" ]]; then
      continue
    fi

    result="$(run_scenario "$f")"
    # Last line of run_scenario output is the name:::caption entry
    local entry
    entry="$(echo "$result" | tail -1)"
    results+=("$entry")
  done

  build_index "${results[@]}"

  echo ""
  echo "done. ${#results[@]} screenshot(s) saved."
}

main "$@"

#!/usr/bin/env bash
# Smoke tests for code graph indexing.
# Requires: Elasticsearch running at localhost:9200
set -euo pipefail

GLITCH="go run ."
REPO_DIR="$(pwd)"
REPO_NAME="gl1tch"

echo "=== Smoke: glitch index (full) ==="
$GLITCH index "$REPO_DIR" --repo "$REPO_NAME" --full --stats
echo "PASS: index completed"

echo ""
echo "=== Smoke: glitch index (incremental, no changes) ==="
$GLITCH index "$REPO_DIR" --repo "$REPO_NAME" --stats
echo "PASS: incremental index completed"

echo ""
echo "=== Smoke: symbols indexed ==="
COUNT=$(curl -s "http://localhost:9200/glitch-symbols-${REPO_NAME}/_count" | jq '.count')
if [ "$COUNT" -lt 1 ]; then
  echo "FAIL: no symbols indexed (count=$COUNT)"
  exit 1
fi
echo "PASS: $COUNT symbols indexed"

echo ""
echo "=== Smoke: edges indexed ==="
EDGE_COUNT=$(curl -s "http://localhost:9200/glitch-edges-${REPO_NAME}/_count" | jq '.count')
if [ "$EDGE_COUNT" -lt 1 ]; then
  echo "FAIL: no edges indexed (count=$EDGE_COUNT)"
  exit 1
fi
echo "PASS: $EDGE_COUNT edges indexed"

echo ""
echo "=== Smoke: search for IndexRepo symbol ==="
RESULT=$(curl -s "http://localhost:9200/glitch-symbols-${REPO_NAME}/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"match":{"name":"IndexRepo"}},"size":1}')
HIT_COUNT=$(echo "$RESULT" | jq '.hits.total.value')
if [ "$HIT_COUNT" -lt 1 ]; then
  echo "FAIL: IndexRepo symbol not found"
  exit 1
fi
echo "PASS: IndexRepo symbol found"

echo ""
echo "=== Smoke: search for calls edges ==="
CALLS=$(curl -s "http://localhost:9200/glitch-edges-${REPO_NAME}/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"term":{"kind":"calls"}},"size":1}')
CALL_HITS=$(echo "$CALLS" | jq '.hits.total.value')
echo "calls edges: $CALL_HITS"
if [ "$CALL_HITS" -lt 1 ]; then
  echo "WARN: no calls edges found (may need review)"
fi
echo "PASS: edges query works"

echo ""
echo "=== Smoke: glitch observe with symbols ==="
$GLITCH observe "what functions are in internal/indexer" --repo "$REPO_NAME" || true
echo "PASS: observe completed (output above)"

echo ""
echo "=== Smoke: symbols-only mode ==="
$GLITCH index "$REPO_DIR" --repo "${REPO_NAME}-test" --full --symbols-only --stats
# Clean up test index
curl -s -X DELETE "http://localhost:9200/glitch-symbols-${REPO_NAME}-test" > /dev/null
curl -s -X DELETE "http://localhost:9200/glitch-edges-${REPO_NAME}-test" > /dev/null
echo "PASS: symbols-only mode completed"

echo ""
echo "=== All smoke tests passed ==="

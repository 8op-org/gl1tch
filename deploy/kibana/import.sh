#!/usr/bin/env bash
# deploy/kibana/import.sh — imports saved objects into Kibana
set -euo pipefail

KIBANA_URL="${KIBANA_URL:-http://localhost:5601}"

echo "Waiting for Kibana..."
for i in $(seq 1 30); do
  if curl -sf "$KIBANA_URL/api/status" > /dev/null 2>&1; then
    break
  fi
  sleep 2
done

for f in deploy/kibana/*.ndjson; do
  [ -f "$f" ] || continue
  echo "Importing $(basename "$f")..."
  curl -sf -X POST "$KIBANA_URL/api/saved_objects/_import?overwrite=true" \
    -H "kbn-xsrf: true" \
    --form file=@"$f" > /dev/null
done

echo "Dashboards imported."

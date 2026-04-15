#!/usr/bin/env python3
"""Split site/generated/docs.json into individual content markdown files."""
import json, os

DOCS_JSON = "site/generated/docs.json"
CONTENT_DIR = "site/src/content/docs"

with open(DOCS_JSON) as f:
    raw = f.read()

start = raw.index("[")
end = raw.rindex("]")
docs = json.loads(raw[start : end + 1])

os.makedirs(CONTENT_DIR, exist_ok=True)

# Clear old files
for f in os.listdir(CONTENT_DIR):
    if f.endswith(".md"):
        os.remove(os.path.join(CONTENT_DIR, f))

for doc in docs:
    title = doc["title"].replace('"', '\\"')
    desc = doc.get("description", "").replace('"', '\\"')
    fm = f'---\ntitle: "{title}"\norder: {doc["order"]}\ndescription: "{desc}"\n---\n\n{doc["content"]}'
    path = os.path.join(CONTENT_DIR, f'{doc["slug"]}.md')
    with open(path, "w") as f:
        f.write(fm)
    print(f"  wrote: {doc['slug']}.md")

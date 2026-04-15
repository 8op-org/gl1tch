#!/usr/bin/env python3
"""Gate 2: Every stub in docs/site/ has a matching entry in docs.json."""
import os, json, glob, sys

stubs = {
    os.path.splitext(os.path.basename(p))[0]
    for p in glob.glob("docs/site/*.md")
}

with open("site/generated/docs.json") as f:
    raw = f.read()
start = raw.index("[")
end = raw.rindex("]")
docs = json.loads(raw[start : end + 1])
slugs = {d["slug"] for d in docs}

missing = stubs - slugs
extra = slugs - stubs

if missing or extra:
    print("FAIL")
    for m in missing:
        print(f"  missing doc for stub: {m}")
    for e in extra:
        print(f"  extra doc without stub: {e}")
    sys.exit(1)
else:
    print(f"PASS: all {len(stubs)} stubs covered")

---
title: "DSL Reference"
order: 2
description: "Extended forms for data pipelines, conditionals, and Elasticsearch integration."
---

This page covers forms beyond the basics in [Workflow Syntax](/docs/workflow-syntax). If you haven't read that first, start there.

## Threading macro

Pipe data through a chain of transforms:

```glitch
(def context
  (-> (read-file "data.json")
      (filter (contains "error"))
      (join "\n")))
```

## Collection forms

```glitch
(step "errors-only"
  (filter (contains "ERROR") :from "logs"))

(step "totals"
  (reduce "+" :from "counts"))
```

## Conditionals

Run steps only when a predicate passes:

```glitch
(when "test -f results.json"
  (step "upload"
    (run "curl -X POST -d @results.json https://api.example.com/results")))

(when-not "git diff --quiet"
  (step "commit"
    (run "git add -A && git commit -m 'auto-commit'")))
```

## Elasticsearch forms

Native ES integration — no shell curl needed:

```glitch
(step "find-errors"
  (search :index "logs-*" :query {"match": {"level": "error"}} :size 50))

(step "store-result"
  (index :index "results" :doc "~(step analysis)"))

(step "cleanup"
  (delete :index "temp-*" :query {"match_all": {}}))
```

## Embedding

Generate vector embeddings via your configured provider:

```glitch
(step "vectorize"
  (embed :input "~(step summary)" :provider "ollama" :model "nomic-embed-text"))
```

## Data transforms

Reshape JSON between steps:

```glitch
(step "flatten-results"
  (flatten "nested-data"))

(step "pick-fields"
  (pick :key "title" :key "status" :from "raw-data"))

(step "add-metadata"
  (assoc :key "reviewed" :status "true" :from "record"))
```

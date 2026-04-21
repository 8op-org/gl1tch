---
title: "Code Intelligence"
description: "How gl1tch understands your codebase beyond raw text search, enabling structured navigation and graph-aware queries."
order: 8
---

`glitch` can build a semantic graph of your codebase, understanding the relationships between symbols like functions, classes, and variables. This powers rich code navigation for you and allows `glitch observe` to answer complex questions about your code's structure and behavior.

## Indexing your Codebase

To enable code intelligence, you first need to index your repository. Run `glitch index` from your repository's root:

```bash
glitch index
```

This command extracts symbols (functions, methods, types, etc.) and their relationships (calls, imports, inheritance) from your code using advanced AST parsing. The extracted data is stored in a per-repository search index.

If your code changes, running `glitch index` again will quickly update the index, only reprocessing files that have been modified.

You can also force a full re-index:

```bash
glitch index --full
```

Specify which languages to index:

```bash
glitch index --languages go,python,javascript
```

Or only index symbols and edges, skipping full content chunks for text search:

```bash
glitch index --symbols-only
```

Once indexed, you can view basic statistics:

```bash
glitch index --stats
```

## Understanding the Symbol Graph

`glitch` builds a graph of your code with the following kinds of information:

### Symbols

These are the fundamental building blocks of your code. Each symbol has:

*   **File:** Where the symbol is defined.
*   **Kind:** `function`, `method`, `type`, `interface`, `class`, `field`, `const`, `var`, `import`, `export`.
*   **Name:** The symbol's identifier.
*   **Signature:** The full declaration line(s) (e.g., function signature, class definition).
*   **Language:** The programming language detected.
*   **Line Range:** The start and end lines of the symbol.
*   **Docstring:** The leading comment block, if present.

### Relationships (Edges)

Symbols are connected by the following relationships:

*   `contains`: A parent symbol contains a child symbol (e.g., a class contains methods).
*   `imports`: A symbol imports another symbol.
*   `exports`: A symbol exports another symbol.
*   `extends`: A class or interface extends another.
*   `implements`: A class implements an interface.
*   `calls`: A function or method calls another.

## Graph-Aware AI with `glitch observe`

The `glitch observe` command leverages your code graph to answer questions that require an understanding of code structure and relationships.

For example, to find out what calls a specific function:

```bash
glitch observe "what calls `IndexRepo`" --repo my-project
```

You can control how deeply `glitch observe` traverses the graph with the `--depth` flag:

```bash
# Finds direct callers of `IndexRepo`
glitch observe "what calls `IndexRepo`" --repo my-project --depth 1

# Finds callers of callers (2 levels deep)
glitch observe "what calls `IndexRepo`" --repo my-project --depth 2
```

When `glitch observe` detects a graph-shaped question, it performs a Breadth-First Search (BFS) traversal of your codebase's symbol graph, collecting relevant symbols and relationships up to the specified `--depth`. This structured information, along with source code snippets, is then provided to the LLM to synthesize a grounded answer.

For questions that don't involve code relationships, `glitch observe` behaves as usual, using text search over code chunks.

## New Research Tools for Structured Code Navigation

When using `(llm ...)` steps with tools enabled, `glitch` provides new capabilities for querying your code graph:

### `search_symbols`

Allows you to query for symbols by name, kind, file, or language. It returns structured data about the symbols, such as their name, kind, file path, line range, and signature.

```glitch
(llm
  :prompt "Find all Go functions related to indexing in the `internal/indexer` directory."
  :tools ["search_symbols"])
```

### `search_edges`

Helps you discover relationships between symbols. You can ask "what calls X", "what does Y call", "what imports Z", or "what implements W".

```glitch
(llm
  :prompt "Show me all symbols that `IndexRepo` calls directly."
  :tools ["search_edges"])
```

### `symbol_context`

Given a symbol's identifier, this tool returns the symbol's full details, its related edges (callers, callees, imports, etc.), and the exact source code snippet where it's defined.

```glitch
(llm
  :prompt "Give me the full context for the `LanguageExtractor` type."
  :tools ["symbol_context"])
```
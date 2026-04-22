(ns glitch.mcp.tools)

(def tool-definitions
  [{"name" "glitch_run"
    "description" "Execute a glitch workflow file and return its output. Use this to run automation pipelines defined in .glitch/workflows/. Pass input text and/or key-value parameters. Returns the workflow's stdout on success, or the error message on failure."
    "inputSchema"
    {"type" "object"
     "properties"
     {"file"  {"type" "string" "description" "Path to the workflow file"}
      "input" {"type" "string" "description" "Input text to pass to the workflow"}
      "set"   {"type" "object" "description" "Key-value pairs to set as workflow parameters"}}
     "required" ["file"]}}

   {"name" "glitch_eval"
    "description" "Evaluate a Clojure expression with the full glitch DSL loaded. Use this to programmatically compose and execute workflow steps, query state, or build pipelines dynamically. Available functions: llm, sh, ref, input, params, param, search, save, read-file, call-workflow, json-extract, validate, validate-schema, gate, consensus, composite-score, search-symbols, search-edges, symbol-context, trace, grounded?"
    "inputSchema"
    {"type" "object"
     "properties"
     {"expression" {"type" "string" "description" "Clojure expression to evaluate"}}
     "required" ["expression"]}}

   {"name" "glitch_check"
    "description" "Validate a glitch workflow file for syntax errors without executing it. Returns 'ok' if valid, or a description of the syntax error found."
    "inputSchema"
    {"type" "object"
     "properties"
     {"file" {"type" "string" "description" "Path to the workflow file to check"}}
     "required" ["file"]}}

   {"name" "glitch_search_symbols"
    "description" "Search the code intelligence index for symbol definitions (functions, methods, classes, types, structs, interfaces, traits, enums). Supports wildcard matching with *. Use this instead of grep when you need structured symbol metadata across a repository."
    "inputSchema"
    {"type" "object"
     "properties"
     {"name"     {"type" "string" "description" "Symbol name (wildcard with * OK)"}
      "kind"     {"type" "string" "description" "Symbol kind: function, method, class, type, interface, trait, struct, enum"}
      "language" {"type" "string" "description" "Language filter: go, python, javascript, rust, java, c"}
      "file"     {"type" "string" "description" "File path pattern"}
      "repo"     {"type" "string" "description" "Repo name (default: cwd basename)"}
      "limit"    {"type" "integer" "description" "Max results (default: 20)"}}
     "required" ["name"]}}

   {"name" "glitch_search_edges"
    "description" "Query code relationships in the intelligence index: calls, imports, contains, extends, implements, references. Supports depth traversal for multi-hop queries (e.g. what calls the functions that call X). Use this to understand how code connects."
    "inputSchema"
    {"type" "object"
     "properties"
     {"source" {"type" "string" "description" "Source symbol name or ID"}
      "target" {"type" "string" "description" "Target symbol name or ID"}
      "kind"   {"type" "string" "description" "Edge kind: calls, imports, contains, extends, implements, references"}
      "depth"  {"type" "integer" "description" "BFS traversal depth (default: 1)"}
      "repo"   {"type" "string" "description" "Repo name"}
      "limit"  {"type" "integer" "description" "Max results (default: 50)"}}}}

   {"name" "glitch_symbol_context"
    "description" "Get a complete picture of a symbol: its definition plus all relationships (callers, callees, parent, children, implementors). Use this when you need to understand a symbol's role in the codebase in one call rather than multiple search_edges queries."
    "inputSchema"
    {"type" "object"
     "properties"
     {"name" {"type" "string" "description" "Symbol name"}
      "repo" {"type" "string" "description" "Repo name"}}
     "required" ["name"]}}])

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
     "required" ["name"]}}

   {"name" "glitch_list_workflows"
    "description" "List available glitch workflows in .glitch/workflows/ with their filenames and descriptions. Use this to discover what automation is available before running a workflow."
    "inputSchema"
    {"type" "object"
     "properties"
     {"path" {"type" "string" "description" "Directory to scan (default: .glitch/workflows/)"}}}}

   {"name" "glitch_recall"
    "description" "Search for workflows by what they do, not by filename. Returns matching workflows with descriptions and paths. Use this to find workflows before running them with glitch_run."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query" {"type" "string" "description" "Natural language description of what you're looking for (e.g. 'summarize ES errors', 'fetch github issues')"}}
     "required" ["query"]}}

   {"name" "glitch_advise"
    "description" "Get a recommendation for which glitch primitives or workflows to use for a given task. Returns a structured recommendation with approach type, relevant primitives, reasoning, and a concrete example. Use this when you're unsure whether glitch can help with the current task."
    "inputSchema"
    {"type" "object"
     "properties"
     {"task"    {"type" "string" "description" "Natural language description of the task"}
      "context" {"type" "string" "description" "Optional additional context (repo, files, domain)"}}
     "required" ["task"]}}

   {"name" "glitch_search"
    "description" "Search file contents using ripgrep. Supports regex patterns, glob filtering, multiline matching, and context lines. Use this for fast text search across a repository."
    "inputSchema"
    {"type" "object"
     "properties"
     {"pattern"   {"type" "string" "description" "Regex pattern to search for"}
      "path"      {"type" "string" "description" "File or directory to search (default: current directory)"}
      "glob"      {"type" "string" "description" "Glob pattern to filter files (e.g. \"*.go\", \"*.{ts,tsx}\")"}
      "multiline" {"type" "boolean" "description" "Enable multiline matching where . matches newlines (default: false)"}
      "context"   {"type" "integer" "description" "Number of context lines before and after each match"}
      "max_count" {"type" "integer" "description" "Maximum matches per file (default: 200)"}}
     "required" ["pattern"]}}

   {"name" "glitch_read_file"
    "description" "Read the contents of a file. Returns up to 200 lines by default. Use offset and limit to read specific ranges of large files."
    "inputSchema"
    {"type" "object"
     "properties"
     {"path"   {"type" "string" "description" "Absolute or relative path to the file"}
      "offset" {"type" "integer" "description" "Line number to start reading from (1-based, default: 1)"}
      "limit"  {"type" "integer" "description" "Maximum number of lines to return (default: 200)"}}
     "required" ["path"]}}])

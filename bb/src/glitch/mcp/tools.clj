(ns glitch.mcp.tools)

(def tool-definitions
  [{"name" "glitch_search"
    "description" "Search code using ripgrep. Returns structured results with file paths, line numbers, and matched content."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query"      {"type" "string" "description" "Search pattern (regex by default)"}
      "path"       {"type" "string" "description" "Directory to search in"}
      "glob"       {"type" "string" "description" "File filter glob (e.g. *.clj, *.{ts,tsx})"}
      "fixed"      {"type" "boolean" "description" "Literal string mode, no regex"}
      "multiline"  {"type" "boolean" "description" "Match across newlines"}
      "pcre2"      {"type" "boolean" "description" "Enable PCRE2 (lookaround, backreferences)"}
      "context"    {"type" "integer" "description" "Lines of context around matches"}
      "limit"      {"type" "integer" "description" "Max matches per file"}
      "smart_case" {"type" "boolean" "description" "Case-insensitive unless uppercase present (default true)"}}
     "required" ["query" "path"]}}

   {"name" "glitch_symbols"
    "description" "Search for symbol definitions (functions, types, classes) using language-aware ripgrep patterns."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query"    {"type" "string" "description" "Symbol name to search for"}
      "path"     {"type" "string" "description" "Directory to search in"}
      "language" {"type" "string" "description" "Language hint: clojure, go, python, javascript, typescript, rust"}}
     "required" ["query" "path"]}}

   {"name" "glitch_run"
    "description" "Execute a glitch workflow file."
    "inputSchema"
    {"type" "object"
     "properties"
     {"file"  {"type" "string" "description" "Path to the workflow file"}
      "input" {"type" "string" "description" "Input text to pass to the workflow"}
      "set"   {"type" "object" "description" "Key-value pairs to set as workflow parameters"}}
     "required" ["file"]}}

   {"name" "glitch_eval"
    "description" "Evaluate a Clojure expression via SCI and return the result."
    "inputSchema"
    {"type" "object"
     "properties"
     {"expression" {"type" "string" "description" "Clojure expression to evaluate"}}
     "required" ["expression"]}}

   {"name" "glitch_check"
    "description" "Check a workflow file for syntax errors."
    "inputSchema"
    {"type" "object"
     "properties"
     {"file" {"type" "string" "description" "Path to the workflow file to check"}}
     "required" ["file"]}}

   {"name" "glitch_read_file"
    "description" "Read a file and return its first 200 lines."
    "inputSchema"
    {"type" "object"
     "properties"
     {"path" {"type" "string" "description" "Path to the file to read"}}
     "required" ["path"]}}

   {"name" "glitch_search_symbols"
    "description" "Search indexed code symbols by name, kind, language, or file path"
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
    "description" "Search code relationships (calls, imports, contains, extends, implements, references) with optional depth traversal"
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
    "description" "Get a symbol's definition and all its relationships (callers, callees, parent, children, implementors)"
    "inputSchema"
    {"type" "object"
     "properties"
     {"name" {"type" "string" "description" "Symbol name"}
      "repo" {"type" "string" "description" "Repo name"}}
     "required" ["name"]}}])

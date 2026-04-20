(ns glitch.mcp.tools)

(def tool-definitions
  [{"name" "glitch_search"
    "description" "Hybrid semantic + keyword code search across indexed repositories."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query" {"type" "string" "description" "Search query text"}
      "repo" {"type" "string" "description" "Repository path to search within"}
      "limit" {"type" "integer" "description" "Maximum number of results to return"}}
     "required" ["query"]}}

   {"name" "glitch_index"
    "description" "Index or reindex a repository for code search."
    "inputSchema"
    {"type" "object"
     "properties"
     {"repo" {"type" "string" "description" "Path to the repository to index"}
      "reindex" {"type" "boolean" "description" "Force full reindex if true"}}
     "required" ["repo"]}}

   {"name" "glitch_run"
    "description" "Execute a glitch workflow by name."
    "inputSchema"
    {"type" "object"
     "properties"
     {"workflow" {"type" "string" "description" "Name of the workflow to run"}
      "input" {"type" "string" "description" "Input text to pass to the workflow"}
      "set" {"type" "object" "description" "Key-value pairs to set as workflow parameters"}}
     "required" ["workflow"]}}

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

   {"name" "glitch_grep"
    "description" "Regex search in code files using grep."
    "inputSchema"
    {"type" "object"
     "properties"
     {"pattern" {"type" "string" "description" "Regex pattern to search for"}
      "path" {"type" "string" "description" "Directory or file path to search in"}
      "glob" {"type" "string" "description" "File glob pattern to filter files"}}
     "required" ["pattern"]}}

   {"name" "glitch_symbols"
    "description" "Search symbol names (functions, types, definitions) in indexed code."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query" {"type" "string" "description" "Symbol name or pattern to search for"}
      "repo" {"type" "string" "description" "Repository path to search within"}}
     "required" ["query"]}}

   {"name" "glitch_read_file"
    "description" "Read a file and return its first 200 lines."
    "inputSchema"
    {"type" "object"
     "properties"
     {"path" {"type" "string" "description" "Path to the file to read"}}
     "required" ["path"]}}])

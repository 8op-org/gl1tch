# MCP tool definitions for the glitch server.
# Each tool follows the MCP tool schema: name, description, inputSchema.
# String keys throughout — these serialize directly to JSON.

(def tool-definitions
  @[
    # 1. glitch_search — hybrid semantic + keyword code search
    @{"name" "glitch_search"
      "description" "Hybrid semantic + keyword code search across indexed repositories."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"query" @{"type" "string" "description" "Search query text"}
          "repo" @{"type" "string" "description" "Repository path to search within"}
          "limit" @{"type" "integer" "description" "Maximum number of results to return"}}
        "required" @["query"]}}

    # 2. glitch_index — index or reindex a repo
    @{"name" "glitch_index"
      "description" "Index or reindex a repository for code search."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"repo" @{"type" "string" "description" "Path to the repository to index"}
          "reindex" @{"type" "boolean" "description" "Force full reindex if true"}}
        "required" @["repo"]}}

    # 3. glitch_run — execute a workflow by name
    @{"name" "glitch_run"
      "description" "Execute a glitch workflow by name."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"workflow" @{"type" "string" "description" "Name of the workflow to run"}
          "input" @{"type" "string" "description" "Input text to pass to the workflow"}
          "set" @{"type" "object" "description" "Key-value pairs to set as workflow parameters"}}
        "required" @["workflow"]}}

    # 4. glitch_eval — evaluate a Janet expression
    @{"name" "glitch_eval"
      "description" "Evaluate a single Janet expression and return the result."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"expression" @{"type" "string" "description" "Janet expression to evaluate"}}
        "required" @["expression"]}}

    # 5. glitch_check — lint a workflow file
    @{"name" "glitch_check"
      "description" "Lint a workflow file by parsing it and reporting errors."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"file" @{"type" "string" "description" "Path to the workflow file to check"}}
        "required" @["file"]}}

    # 6. glitch_grep — regex search in code
    @{"name" "glitch_grep"
      "description" "Regex search in code files using grep."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"pattern" @{"type" "string" "description" "Regex pattern to search for"}
          "path" @{"type" "string" "description" "Directory or file path to search in"}
          "glob" @{"type" "string" "description" "File glob pattern to filter files (e.g. *.janet)"}}
        "required" @["pattern"]}}

    # 7. glitch_symbols — search symbol names
    @{"name" "glitch_symbols"
      "description" "Search symbol names (functions, types, definitions) in indexed code."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"query" @{"type" "string" "description" "Symbol name or pattern to search for"}
          "repo" @{"type" "string" "description" "Repository path to search within"}}
        "required" @["query"]}}

    # 8. glitch_read_file — read a file (first 200 lines)
    @{"name" "glitch_read_file"
      "description" "Read a file and return its first 200 lines."
      "inputSchema"
      @{"type" "object"
        "properties"
        @{"path" @{"type" "string" "description" "Path to the file to read"}}
        "required" @["path"]}}
  ])

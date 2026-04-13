package research

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/8op-org/gl1tch/internal/esearch"
)

// Tool describes a callable tool the LLM can invoke during research.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Params      string `json:"params"`
}

// ToolResult holds the output of a tool execution.
type ToolResult struct {
	Tool   string `json:"tool"`
	Output string `json:"output"`
	Err    string `json:"error,omitempty"`
}

// ToolSet holds available tools for the LLM to call during research.
type ToolSet struct {
	repoPath string
	es       *esearch.Client
}

// NewToolSet returns a ToolSet rooted at repoPath with an optional ES client.
func NewToolSet(repoPath string, es *esearch.Client) *ToolSet {
	return &ToolSet{repoPath: repoPath, es: es}
}

// Definitions returns the 8 tool definitions available for tool-use.
func (ts *ToolSet) Definitions() []Tool {
	return []Tool{
		{
			Name:        "grep_code",
			Description: "Search code for a regex pattern",
			Params:      "pattern (required), path (optional), glob (optional)",
		},
		{
			Name:        "read_file",
			Description: "Read the first 200 lines of a file",
			Params:      "path (required)",
		},
		{
			Name:        "git_log",
			Description: "Show recent git commits",
			Params:      "query (optional, --grep), path (optional), limit (optional, default 20)",
		},
		{
			Name:        "git_diff",
			Description: "Show git diff stats between refs",
			Params:      "ref1 (optional, default HEAD~10), ref2 (optional), path (optional)",
		},
		{
			Name:        "search_es",
			Description: "Search Elasticsearch with a multi_match query",
			Params:      "query (required), index (optional)",
		},
		{
			Name:        "list_files",
			Description: "List files in a directory tree",
			Params:      "path (optional), depth (optional, default 3)",
		},
		{
			Name:        "fetch_issue",
			Description: "Fetch a GitHub issue via gh CLI",
			Params:      "repo (required), number (required)",
		},
		{
			Name:        "fetch_pr",
			Description: "Fetch a GitHub pull request via gh CLI",
			Params:      "repo (required), number (required)",
		},
	}
}

// ValidTool returns true if name matches a known tool.
func (ts *ToolSet) ValidTool(name string) bool {
	for _, t := range ts.Definitions() {
		if t.Name == name {
			return true
		}
	}
	return false
}

// Execute dispatches a tool call by name with the given params.
func (ts *ToolSet) Execute(ctx context.Context, name string, params map[string]string) ToolResult {
	switch name {
	case "grep_code":
		return ts.grepCode(ctx, params)
	case "read_file":
		return ts.readFile(ctx, params)
	case "git_log":
		return ts.gitLog(ctx, params)
	case "git_diff":
		return ts.gitDiff(ctx, params)
	case "search_es":
		return ts.searchES(ctx, params)
	case "list_files":
		return ts.listFiles(ctx, params)
	case "fetch_issue":
		return ts.fetchIssue(ctx, params)
	case "fetch_pr":
		return ts.fetchPR(ctx, params)
	default:
		return ToolResult{Tool: name, Err: fmt.Sprintf("unknown tool: %s", name)}
	}
}

func (ts *ToolSet) grepCode(ctx context.Context, params map[string]string) ToolResult {
	pattern := params["pattern"]
	if pattern == "" {
		return ToolResult{Tool: "grep_code", Err: "missing required param: pattern"}
	}

	searchPath := ts.repoPath
	if p, ok := params["path"]; ok && p != "" {
		searchPath = p
	}

	args := []string{"-rn", "--max-count=5"}
	if g, ok := params["glob"]; ok && g != "" {
		args = append(args, "--include="+g)
	}
	args = append(args, pattern, searchPath)

	out, err := exec.CommandContext(ctx, "grep", args...).CombinedOutput()
	if err != nil {
		// grep returns exit 1 for no matches — that's not an error
		if len(out) == 0 {
			return ToolResult{Tool: "grep_code", Output: "(no matches)"}
		}
	}
	return ToolResult{Tool: "grep_code", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) readFile(ctx context.Context, params map[string]string) ToolResult {
	p := params["path"]
	if p == "" {
		return ToolResult{Tool: "read_file", Err: "missing required param: path"}
	}

	filePath := p
	// If it's not an absolute path, treat as relative to repo
	if !strings.HasPrefix(p, "/") {
		filePath = ts.repoPath + "/" + p
	}

	out, err := exec.CommandContext(ctx, "head", "-n", "200", filePath).CombinedOutput()
	if err != nil {
		return ToolResult{Tool: "read_file", Err: fmt.Sprintf("read_file: %s", string(out))}
	}
	return ToolResult{Tool: "read_file", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) gitLog(ctx context.Context, params map[string]string) ToolResult {
	limit := params["limit"]
	if limit == "" {
		limit = "20"
	}

	args := []string{"-C", ts.repoPath, "log", "--oneline", "-n", limit}
	if q, ok := params["query"]; ok && q != "" {
		args = append(args, "--grep="+q)
	}
	if p, ok := params["path"]; ok && p != "" {
		args = append(args, "--", p)
	}

	out, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return ToolResult{Tool: "git_log", Err: fmt.Sprintf("git_log: %s", string(out))}
	}
	return ToolResult{Tool: "git_log", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) gitDiff(ctx context.Context, params map[string]string) ToolResult {
	ref1 := params["ref1"]
	if ref1 == "" {
		ref1 = "HEAD~10"
	}

	args := []string{"-C", ts.repoPath, "diff", ref1}
	if ref2, ok := params["ref2"]; ok && ref2 != "" {
		args = append(args, ref2)
	}
	args = append(args, "--stat")
	if p, ok := params["path"]; ok && p != "" {
		args = append(args, "--", p)
	}

	out, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return ToolResult{Tool: "git_diff", Err: fmt.Sprintf("git_diff: %s", string(out))}
	}
	return ToolResult{Tool: "git_diff", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) searchES(ctx context.Context, params map[string]string) ToolResult {
	if ts.es == nil {
		return ToolResult{Tool: "search_es", Output: "elasticsearch not available"}
	}

	query := params["query"]
	if query == "" {
		return ToolResult{Tool: "search_es", Err: "missing required param: query"}
	}

	index := params["index"]
	if index == "" {
		index = "*"
	}

	esQuery := map[string]interface{}{
		"size": 10,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": query,
			},
		},
	}
	raw, err := json.Marshal(esQuery)
	if err != nil {
		return ToolResult{Tool: "search_es", Err: fmt.Sprintf("search_es: marshal: %s", err)}
	}

	resp, err := ts.es.Search(ctx, []string{index}, raw)
	if err != nil {
		return ToolResult{Tool: "search_es", Err: fmt.Sprintf("search_es: %s", err)}
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	return ToolResult{Tool: "search_es", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) listFiles(ctx context.Context, params map[string]string) ToolResult {
	searchPath := ts.repoPath
	if p, ok := params["path"]; ok && p != "" {
		searchPath = p
	}

	depth := params["depth"]
	if depth == "" {
		depth = "3"
	}

	args := []string{
		searchPath,
		"-maxdepth", depth,
		"-type", "f",
		"-not", "-path", "*/.git/*",
		"-not", "-path", "*/node_modules/*",
		"-not", "-path", "*/vendor/*",
	}

	out, err := exec.CommandContext(ctx, "find", args...).CombinedOutput()
	if err != nil {
		return ToolResult{Tool: "list_files", Err: fmt.Sprintf("list_files: %s", string(out))}
	}
	return ToolResult{Tool: "list_files", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) fetchIssue(ctx context.Context, params map[string]string) ToolResult {
	repo := params["repo"]
	number := params["number"]
	if repo == "" || number == "" {
		return ToolResult{Tool: "fetch_issue", Err: "missing required params: repo, number"}
	}

	out, err := exec.CommandContext(ctx, "gh", "issue", "view", number,
		"--repo", repo,
		"--json", "number,title,body,comments,labels",
	).CombinedOutput()
	if err != nil {
		return ToolResult{Tool: "fetch_issue", Err: fmt.Sprintf("fetch_issue: %s", string(out))}
	}
	return ToolResult{Tool: "fetch_issue", Output: truncateOutput(string(out), 8000)}
}

func (ts *ToolSet) fetchPR(ctx context.Context, params map[string]string) ToolResult {
	repo := params["repo"]
	number := params["number"]
	if repo == "" || number == "" {
		return ToolResult{Tool: "fetch_pr", Err: "missing required params: repo, number"}
	}

	out, err := exec.CommandContext(ctx, "gh", "pr", "view", number,
		"--repo", repo,
		"--json", "number,title,body,files,additions,deletions,state,reviews",
	).CombinedOutput()
	if err != nil {
		return ToolResult{Tool: "fetch_pr", Err: fmt.Sprintf("fetch_pr: %s", string(out))}
	}
	return ToolResult{Tool: "fetch_pr", Output: truncateOutput(string(out), 8000)}
}

// truncateOutput truncates s at max bytes, appending a suffix if truncated.
func truncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "... (truncated)"
}

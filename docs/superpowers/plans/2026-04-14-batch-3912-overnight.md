# Overnight Batch #3912 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate all 7 sub-issue deliverables for elastic/observability-robots#3912 using local LLMs and Claude, with full telemetry in Elasticsearch and Kibana dashboards.

**Architecture:** Go code changes to support provider model passthrough and enriched telemetry, 14 sexpr workflows (7 issues x 2 variants), a batch shell script, and Vega-based Kibana dashboards.

**Tech Stack:** Go, Ollama (qwen3-coder:30b, qwen3.5:35b-a3b, qwen3:8b), Claude CLI, Elasticsearch 8.17, Kibana 8.17

---

### Task 1: Provider model passthrough

**Files:**
- Modify: `internal/provider/provider.go:68-111`
- Modify: `internal/pipeline/runner.go:119`
- Modify: `internal/research/toolloop.go` (if it calls RunProvider)

- [ ] **Step 1: Update RunProvider signature to accept model**

In `internal/provider/provider.go`, change `RunProvider` and `RenderCommand` to accept model:

```go
func (r *ProviderRegistry) RenderCommand(name string, data map[string]string) (string, error) {
	p, ok := r.providers[name]
	if !ok {
		available := make([]string, 0, len(r.providers))
		for k := range r.providers {
			available = append(available, k)
		}
		sort.Strings(available)
		return "", fmt.Errorf("provider %q not found; available: %s", name, strings.Join(available, ", "))
	}

	tmpl, err := template.New("cmd").Parse(p.Command)
	if err != nil {
		return "", fmt.Errorf("bad template for provider %q: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render provider %q: %w", name, err)
	}
	return buf.String(), nil
}

func (r *ProviderRegistry) RunProvider(name, model, prompt string) (string, error) {
	p, ok := r.providers[name]
	if !ok {
		avail := make([]string, 0)
		for n := range r.providers {
			avail = append(avail, n)
		}
		return "", fmt.Errorf("provider %q not found (available: %s)", name, strings.Join(avail, ", "))
	}

	data := map[string]string{"prompt": prompt, "model": model}

	if strings.Contains(p.Command, "{{.prompt}}") || strings.Contains(p.Command, "{{.model}}") {
		rendered, err := r.RenderCommand(name, data)
		if err != nil {
			return "", err
		}
		return RunShell(rendered)
	}
	return RunShellWithStdin(p.Command, prompt)
}
```

- [ ] **Step 2: Update all callers of RunProvider**

In `internal/pipeline/runner.go:119`, change:
```go
out, err = reg.RunProvider(prov, rendered)
```
to:
```go
out, err = reg.RunProvider(prov, model, rendered)
```

In `internal/research/toolloop.go`, grep for `RunProvider` — the tiered runner calls `callProvider` which calls `RunProvider` internally. Check `internal/provider/tiers.go:99`:
```go
raw, err := tr.reg.RunProvider(name, prompt)
```
Change to:
```go
raw, err := tr.reg.RunProvider(name, model, prompt)
```

- [ ] **Step 3: Update claude.yaml provider**

Write `~/.config/glitch/providers/claude.yaml`:
```yaml
name: claude
command: claude -p --model {{.model}} --output-format text
```

- [ ] **Step 4: Build and verify**

Run: `cd ~/Projects/gl1tch && go build ./...`
Expected: clean compile

- [ ] **Step 5: Commit**

```bash
git add internal/provider/provider.go internal/provider/tiers.go internal/pipeline/runner.go
git commit -m "feat: pass model to external providers via template"
```

---

### Task 2: Telemetry schema enhancements

**Files:**
- Modify: `internal/esearch/telemetry.go:42-55`
- Modify: `internal/esearch/mappings.go:63-81`

- [ ] **Step 1: Add fields to LLMCallDoc**

In `internal/esearch/telemetry.go`, update `LLMCallDoc`:

```go
type LLMCallDoc struct {
	RunID            string  `json:"run_id"`
	Step             string  `json:"step"`
	Tier             int     `json:"tier"`
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	TokensIn         int64   `json:"tokens_in"`
	TokensOut        int64   `json:"tokens_out"`
	TokensTotal      int64   `json:"tokens_total"`
	CostUSD          float64 `json:"cost_usd"`
	LatencyMS        int64   `json:"latency_ms"`
	Escalated        bool    `json:"escalated"`
	EscalationReason string  `json:"escalation_reason"`
	WorkflowName     string  `json:"workflow_name"`
	Issue            string  `json:"issue"`
	ComparisonGroup  string  `json:"comparison_group"`
	Timestamp        string  `json:"timestamp"`
}
```

- [ ] **Step 2: Add WorkflowRunDoc and IndexWorkflowRun**

Add to `internal/esearch/telemetry.go`:

```go
type WorkflowRunDoc struct {
	RunID           string  `json:"run_id"`
	WorkflowName    string  `json:"workflow_name"`
	Issue           string  `json:"issue"`
	ComparisonGroup string  `json:"comparison_group"`
	TotalSteps      int     `json:"total_steps"`
	LLMSteps        int     `json:"llm_steps"`
	TotalTokensIn   int64   `json:"total_tokens_in"`
	TotalTokensOut  int64   `json:"total_tokens_out"`
	TotalCostUSD    float64 `json:"total_cost_usd"`
	TotalLatencyMS  int64   `json:"total_latency_ms"`
	ReviewPass      bool    `json:"review_pass"`
	Timestamp       string  `json:"timestamp"`
}

func (t *Telemetry) IndexWorkflowRun(ctx context.Context, doc WorkflowRunDoc) error {
	if t == nil {
		return nil
	}
	return t.indexDoc(ctx, IndexWorkflowRuns, doc.RunID, doc)
}
```

- [ ] **Step 3: Update mappings**

In `internal/esearch/mappings.go`, update `LLMCallsMapping` to add:
```json
"tokens_total":       { "type": "long" },
"workflow_name":      { "type": "keyword" },
"issue":              { "type": "keyword" },
"comparison_group":   { "type": "keyword" }
```

Add new constant and mapping:
```go
const IndexWorkflowRuns = "glitch-workflow-runs"

const WorkflowRunsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":           { "type": "keyword" },
      "workflow_name":    { "type": "keyword" },
      "issue":            { "type": "keyword" },
      "comparison_group": { "type": "keyword" },
      "total_steps":      { "type": "integer" },
      "llm_steps":        { "type": "integer" },
      "total_tokens_in":  { "type": "long" },
      "total_tokens_out": { "type": "long" },
      "total_cost_usd":   { "type": "float" },
      "total_latency_ms": { "type": "long" },
      "review_pass":      { "type": "boolean" },
      "timestamp":        { "type": "date" }
    }
  }
}`
```

Update `AllIndices()` to include `IndexWorkflowRuns: WorkflowRunsMapping`.

- [ ] **Step 4: Build and verify**

Run: `cd ~/Projects/gl1tch && go build ./...`

- [ ] **Step 5: Commit**

```bash
git add internal/esearch/telemetry.go internal/esearch/mappings.go
git commit -m "feat: enriched telemetry — workflow_name, issue, comparison_group, workflow runs index"
```

---

### Task 3: Pipeline runner — enriched telemetry + workflow run summary

**Files:**
- Modify: `internal/pipeline/runner.go`
- Modify: `cmd/workflow.go`

- [ ] **Step 1: Extend RunOpts and update runner**

In `internal/pipeline/runner.go`, update `RunOpts`:

```go
type RunOpts struct {
	Telemetry       *esearch.Telemetry
	Issue           string
	ComparisonGroup string
}
```

Update the `Run` function to:
1. Parse issue/comparison_group from workflow name if not in opts
2. Populate new fields in LLMCallDoc
3. Track totals and index WorkflowRunDoc after completion

Replace the entire `Run` function with the version that adds name parsing, enriched telemetry fields, and end-of-run summary doc. Key additions:

After `runID := esearch.NewRunID()`, add name parsing:
```go
issue := ""
compGroup := ""
if len(opts) > 0 {
    tel = opts[0].Telemetry
    issue = opts[0].Issue
    compGroup = opts[0].ComparisonGroup
}
if issue == "" || compGroup == "" {
    // Parse from workflow name: "3918-wrapper-curl-local" → issue="3918", group="local"
    wname := w.Name
    if strings.HasSuffix(wname, "-local") {
        compGroup = "local"
        wname = strings.TrimSuffix(wname, "-local")
    } else if strings.HasSuffix(wname, "-claude") {
        compGroup = "claude"
        wname = strings.TrimSuffix(wname, "-claude")
    }
    // Leading digits = issue number
    for i, c := range wname {
        if c < '0' || c > '9' {
            if i > 0 {
                issue = wname[:i]
            }
            break
        }
    }
}
```

Add accumulators before the step loop:
```go
var totalTokensIn, totalTokensOut int64
var totalCostUSD float64
var totalLatencyMS int64
var llmSteps int
var lastLLMOutput string
```

In each LLM call telemetry block, populate the new fields and accumulate:
```go
tokensTotal := int64(result.TokensIn) + int64(result.TokensOut)
totalTokensIn += int64(result.TokensIn)
totalTokensOut += int64(result.TokensOut)
totalLatencyMS += result.Latency.Milliseconds()
llmSteps++
lastLLMOutput = result.Response

tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
    RunID:           runID,
    Step:            fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
    Tier:            0,
    Provider:        "ollama",
    Model:           model,
    TokensIn:        int64(result.TokensIn),
    TokensOut:       int64(result.TokensOut),
    TokensTotal:     tokensTotal,
    CostUSD:         0,
    LatencyMS:       result.Latency.Milliseconds(),
    WorkflowName:    w.Name,
    Issue:           issue,
    ComparisonGroup: compGroup,
    Timestamp:       time.Now().UTC().Format(time.RFC3339),
})
```

Same pattern for the non-ollama provider path (with EstimateTokens for token counts).

After the step loop, before returning, index the workflow run summary:
```go
if tel != nil {
    reviewPass := strings.Contains(strings.ToUpper(lastLLMOutput), "OVERALL: PASS")
    tel.IndexWorkflowRun(context.Background(), esearch.WorkflowRunDoc{
        RunID:           runID,
        WorkflowName:    w.Name,
        Issue:           issue,
        ComparisonGroup: compGroup,
        TotalSteps:      len(w.Steps),
        LLMSteps:        llmSteps,
        TotalTokensIn:   totalTokensIn,
        TotalTokensOut:  totalTokensOut,
        TotalCostUSD:    totalCostUSD,
        TotalLatencyMS:  totalLatencyMS,
        ReviewPass:      reviewPass,
        Timestamp:       time.Now().UTC().Format(time.RFC3339),
    })
}
```

- [ ] **Step 2: Build and verify**

Run: `cd ~/Projects/gl1tch && go build ./... && go test ./internal/pipeline/`

- [ ] **Step 3: Commit**

```bash
git add internal/pipeline/runner.go
git commit -m "feat: pipeline runner emits enriched telemetry + workflow run summary"
```

---

### Task 4: Recursive workflow loading

**Files:**
- Modify: `internal/pipeline/types.go:64-89`

- [ ] **Step 1: Make LoadDir walk subdirectories**

Replace `LoadDir` in `internal/pipeline/types.go`:

```go
func LoadDir(dir string) (map[string]*Workflow, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	workflows := make(map[string]*Workflow)
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := filepath.Ext(d.Name())
		if ext != ".yaml" && ext != ".yml" && ext != ".glitch" {
			return nil
		}
		w, err := LoadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			return nil
		}
		workflows[w.Name] = w
		return nil
	})
	return workflows, nil
}
```

- [ ] **Step 2: Build and test**

Run: `cd ~/Projects/gl1tch && go build ./... && go test ./internal/pipeline/`

- [ ] **Step 3: Commit**

```bash
git add internal/pipeline/types.go
git commit -m "feat: LoadDir walks subdirectories for workflow discovery"
```

---

### Task 5: Wipe ES indices

- [ ] **Step 1: Delete old indices**

```bash
curl -sf -X DELETE http://localhost:9200/glitch-llm-calls
curl -sf -X DELETE http://localhost:9200/glitch-workflow-runs
curl -sf -X DELETE http://localhost:9200/glitch-tool-calls
curl -sf -X DELETE http://localhost:9200/glitch-research-runs
```

- [ ] **Step 2: Rebuild glitch and trigger index creation**

```bash
cd ~/Projects/gl1tch && go build -o /tmp/glitch-batch .
/tmp/glitch-batch workflow list >/dev/null 2>&1
```

Indices will be created on first telemetry write.

---

### Task 6: Write sexpr workflows

**Files:**
- Create: `~/Projects/stokagent/workflows/batch-3912/3916-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3916-claude.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3918-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3918-claude.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3919-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3919-claude.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3920-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3920-claude.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3921-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3921-claude.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3922-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3922-claude.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3923-local.glitch`
- Create: `~/Projects/stokagent/workflows/batch-3912/3923-claude.glitch`

Each workflow follows the shapes from the spec. Shell steps gather repo data, LLM steps analyze/write/review. Claude variants swap provider+model defs.

- [ ] **Step 1: Create directory and write all 14 workflow files**

See spec for workflow structure. Each use-case workflow (3918-3922) follows shape 2. 3916 follows shape 1. 3923 follows shape 3.

- [ ] **Step 2: Verify workflows load**

```bash
/tmp/glitch-batch workflow list | grep 3916
```

Expected: both `3916-intro-local` and `3916-intro-claude` appear.

- [ ] **Step 3: Commit to stokagent**

```bash
cd ~/Projects/stokagent
git add workflows/batch-3912/
git commit -m "feat: batch-3912 sexpr workflows — local + claude variants for all 7 sub-issues"
```

---

### Task 7: Batch runner script

**Files:**
- Create: `~/Projects/stokagent/scripts/batch-3912.sh`

- [ ] **Step 1: Write batch script**

See spec for script content. Runs all 14 workflows in dependency order with logging.

- [ ] **Step 2: Make executable and commit**

```bash
chmod +x ~/Projects/stokagent/scripts/batch-3912.sh
cd ~/Projects/stokagent && git add scripts/batch-3912.sh
git commit -m "feat: overnight batch runner for #3912"
```

---

### Task 8: Test run — verify telemetry with #3916 pair

- [ ] **Step 1: Run 3916-local**

```bash
cd ~/Projects/gl1tch && /tmp/glitch-batch workflow run 3916-intro-local
```

- [ ] **Step 2: Run 3916-claude**

```bash
/tmp/glitch-batch workflow run 3916-intro-claude
```

- [ ] **Step 3: Verify telemetry in ES**

```bash
curl -s 'http://localhost:9200/glitch-llm-calls/_search?size=10' \
  -H 'Content-Type: application/json' \
  -d '{"query":{"term":{"issue":"3916"}}}' | python3 -m json.tool | head -40
```

Verify: `workflow_name`, `issue`, `comparison_group`, `tokens_total` fields are populated.

```bash
curl -s 'http://localhost:9200/glitch-workflow-runs/_search?size=10' \
  -H 'Content-Type: application/json' | python3 -m json.tool | head -40
```

Verify: two workflow run docs with `review_pass` field.

- [ ] **Step 4: Verify output files**

```bash
ls -la results/3916/ results/3916-claude/ 2>/dev/null
```

---

### Task 9: Kibana dashboards (Vega)

**Files:**
- Dashboard and Vega visualizations created via Kibana API

- [ ] **Step 1: Create data views**

Create data views for `glitch-llm-calls`, `glitch-workflow-runs` via API.

- [ ] **Step 2: Create Vega visualizations and dashboard**

Use Kibana saved objects API with Vega specs for each panel. Vega specs are JSON — they survive API creation.

- [ ] **Step 3: Verify dashboard loads**

Open `http://localhost:5601/app/dashboards#/view/glitch-llm-dashboard` and verify panels render with data.

---

### Task 10: Run full overnight batch

- [ ] **Step 1: Execute batch**

```bash
nohup ~/Projects/stokagent/scripts/batch-3912.sh > ~/Projects/gl1tch/results/batch-3912.log 2>&1 &
```

- [ ] **Step 2: Verify completion**

Check log and ES for 14 workflow run docs.

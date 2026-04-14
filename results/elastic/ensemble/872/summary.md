I now have thorough evidence. Here's the research summary:

---

## Research Summary: User Error Tracking in elastic/ensemble

### Root Cause Analysis

The issue is that **workflow step failures surface minimal context to users**. The error reporting pipeline has the plumbing for structured errors but three critical gaps prevent actionable output:

#### Gap 1: No Error Tracking ID
There is **no unique error identifier** anywhere in the codebase. When a step fails, `__safe_exec()` at `worker/job.py:128-145` creates a `JobStepStatus` with a `data` dict containing at most `cause` and `help` keys — but no correlation ID that could tie the failure to server-side logs or support tickets.

#### Gap 2: Generic Exception Fallback Loses Context
At `worker/job.py:141-143`, any non-`StepExecutionError` exception is caught and reduced to:
```python
JobStepStatus(state="failed", data={"cause": str(err)})
```
This loses the stack trace, error class, and any structured context. The user sees only the stringified exception.

#### Gap 3: Renderer Has No Guidance or Links
The `JobRenderer` at `client/renderers.py:15-74` displays error data as a key/value table (line 45-71) but:
- Never shows troubleshooting guidance or links
- Never shows a reporting channel
- Only displays `step_status.data` if it exists — no fallback message for empty data

### Affected Files

| File | Lines | Role |
|------|-------|------|
| `src/ensemble/core/exceptions.py` | 6-27 | `ExplainedError` base with `cause`/`help` — **no error ID field** |
| `src/ensemble/core/exceptions.py` | 29-38 | `StepExecutionError` — carries `data: dict` but **no ID** |
| `src/ensemble/worker/job.py` | 128-145 | `__safe_exec()` — catches exceptions, builds `JobStepStatus` |
| `src/ensemble/worker/executors/base.py` | 285-290 | Converts `ExplainedError` → `StepExecutionError` with `cause`/`help` |
| `src/ensemble/client/renderers.py` | 30-43 | Step state display (skull emoji + "failed") |
| `src/ensemble/client/renderers.py` | 45-71 | Error data rendering — **no guidance/links/error ID** |

### What PR #747 Does
PR #747 ("feat: realtime logging to UI terminal") adds real-time log streaming to the frontend. It is **tangentially related** — it improves visibility but does not address structured error reporting, tracking IDs, or user guidance.

### Concrete Next Steps

1. **Add an error tracking ID** — Generate a short unique ID (e.g., UUID prefix or hash) in `__safe_exec()` at `worker/job.py:131-143` and include it in `JobStepStatus.data["error_id"]`. Log the same ID server-side so support can correlate.

2. **Extend `ExplainedError`** — Add optional `troubleshooting_url` and `report_channel` fields to `ExplainedError` at `core/exceptions.py:6-27`. Each specialized error subclass (lines 41-430) can then provide domain-specific links.

3. **Enhance the renderer** — Update `JobRenderer.render()` at `client/renderers.py:45-71` to:
   - Always display `error_id` when present
   - Render `help` as actionable guidance text
   - Show `troubleshooting_url` as a clickable link
   - Show `report_channel` as a Slack deep link
   - Provide a fallback message when `data` is empty (e.g., "An unexpected error occurred. Report this error ID to #support.")

4. **Enrich the generic fallback** — At `worker/job.py:141-143`, capture the exception class name and include it in data alongside the new error ID, so even unhandled exceptions have traceable identifiers.
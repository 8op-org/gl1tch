## Context

`orcai-plugins` has one plugin (Ollama) and no automated e2e validation. `opencode run` is already present on the system (`v1.3.0`) and already knows about Ollama models via `opencode models` — so the opencode plugin is primarily a thin adapter that bridges orcai's stdin/stdout calling convention to `opencode run`'s positional-arg interface. The jq sidecar follows the same pattern: wrap a system CLI as a zero-Go-code sidecar descriptor. The e2e test harness ties all pieces together.

## Goals / Non-Goals

**Goals:**
- `orcai-opencode` binary: wrap `opencode run --model <model> <prompt>` to be callable from orcai pipeline steps.
- `jq` sidecar YAML: expose system `jq` as a pipeline step without any Go code.
- Shell-based e2e test harness in `tests/` with a `make test-e2e` target; each test is a small `.sh` script with `pass`/`fail` assertions.
- New example pipelines: `jq-transform.pipeline.yaml` (HTTP → jq) and `opencode-local.pipeline.yaml` (opencode with ollama backend).
- Prereq-checking helper in the test harness that skips tests gracefully when a tool is absent.

**Non-Goals:**
- Making `opencode` interactive or streaming within orcai — `opencode run` in non-interactive mode is sufficient.
- CI/CD automation (GitHub Actions) for the test suite — that is a follow-up.
- Managing opencode sessions or history from within orcai.
- Installing `jq` or `opencode` — user is responsible.

## Decisions

### Decision: opencode plugin passes prompt as CLI positional arg, not stdin

**Chosen**: The `orcai-opencode` binary reads stdin as the prompt, then shells out to `opencode run "$prompt"`. The prompt becomes a positional argument to `opencode run`, not stdin.

**Rationale**: `opencode run [message..]` takes the message as positional args. Unlike the Ollama daemon which exposes an HTTP API and naturally consumes a JSON body, `opencode` is a CLI-first tool. Passing the prompt as a positional arg is the only non-interactive interface.

**Alternative considered**: Running `opencode` with stdin piped in. Rejected — `opencode run` doesn't read stdin as the message; passing it as positional args is the documented interface.

---

### Decision: jq sidecar is pure YAML — no Go binary; CliAdapter passes all vars as ORCAI_* env vars

**Chosen**: Updated `CliAdapter.Execute` in orcai core to set every entry in `vars` as `ORCAI_<KEY>=<value>` in the subprocess environment (inheriting the current environment and overlaying). The jq plugin is then a pure YAML sidecar with `command: sh` and `args: ["-c", "jq \"${ORCAI_FILTER:-.}\""]` — no compiled binary at all. Pipeline steps set `vars: { filter: ".name" }` which becomes `ORCAI_FILTER=.name` and the shell command expands it.

**Rationale**: Eliminates a Go binary entirely; demonstrates the power of the sidecar model for wrapping any system CLI. Also fixes the spec/impl gap: the `cli-adapter-sidecar` spec says vars should be passed as `ORCAI_*` env vars, but the prior implementation only handled the `model` key as a `--model` flag. Now all plugins benefit — `ORCAI_OLLAMA_URL` works for the Ollama plugin without any binary changes.

**Alternative originally considered**: Go `orcai-jq` shim. Replaced after user feedback; the pure YAML approach is simpler and more composable.

---

### Decision: e2e tests are shell scripts, not Go test files

**Chosen**: Each test is a `tests/<name>.sh` script that runs `orcai pipeline run` and checks output with `grep`, `jq`, or `test` assertions. A root `tests/run_all.sh` orchestrates them and reports pass/fail.

**Rationale**: Shell scripts are the most natural way to test CLI tools end-to-end. They don't require pulling in a testing framework, they run exactly as a user would, and they are easy to read and extend. Go-based integration tests would add compile overhead and obscure the "just run it" nature of e2e tests.

**Alternative considered**: Go `TestMain`-based integration tests. Rejected — overly heavyweight for testing CLI pipelines; the purpose is to validate the pipeline YAML files and plugin binaries, not Go internals.

---

### Decision: Test harness skips gracefully on missing prerequisites

**Chosen**: Each test script checks for its required tools at the top (using `command -v`) and exits with code `77` (a conventional "skip" exit code, also used by autoconf). The root runner treats exit 77 as SKIP, not FAIL.

**Rationale**: Allows the harness to run in environments where Ollama or opencode aren't available without failing the suite. Developers running only unit tests still get a green run.

## Risks / Trade-offs

- **`opencode run` prompt escaping** → Shell quoting a multi-line prompt passed as a positional arg requires care. The binary uses `exec.Command` (not `sh -c`), so quoting is handled by the Go runtime. No shell injection risk. Mitigation: use `exec.Command("opencode", "run", "--model", model, "--", prompt)` with explicit `--` separator.
- **`opencode run` may open a TUI** → In some configurations, `opencode run` without `--format` may try to render a TUI. Mitigation: always pass `--format default` or `--format json` to force non-interactive output mode.
- **jq filter injection** → If a pipeline step passes an untrusted `filter` value, it becomes a `jq` expression. Mitigation: document that `filter` values are trusted pipeline-author input, not user input; no sandboxing at this layer.
- **Test flakiness from LLM output** → e2e tests that check LLM completions can't assert exact text. Mitigation: tests assert non-empty output and exit 0; they do NOT assert specific content.

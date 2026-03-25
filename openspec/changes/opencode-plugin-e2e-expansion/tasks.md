## 1. OpenCode Plugin — Binary

- [x] 1.1 Create `plugins/opencode/` with `go.mod` (module `github.com/adam-stokes/orcai-plugins/plugins/opencode`)
- [x] 1.2 Create `plugins/opencode/main.go`: read stdin as prompt, resolve model from `--model` flag then `ORCAI_MODEL` env var, error if neither set
- [x] 1.3 Invoke `opencode run --model <model> --format default -- <prompt>` via `exec.Command`; pipe stdout to binary stdout; propagate exit code
- [x] 1.4 Handle empty stdin: exit non-zero with descriptive error
- [x] 1.5 Write unit tests: empty stdin, missing model, `--model` flag used, `ORCAI_MODEL` fallback, exit code propagation via mock exec
- [x] 1.6 Verify `go test ./...` passes in `plugins/opencode/`

## 2. OpenCode Plugin — Sidecar & Makefile

- [x] 2.1 Create `plugins/opencode/opencode.yaml` declaring `name: opencode`, description, `command: orcai-opencode`; add comment documenting `vars.model` (e.g. `ollama/llama3.2`)
- [x] 2.2 Create `plugins/opencode/Makefile` with `build`, `install` (to `~/.local/bin` if `/usr/local/bin` unwritable), and `test` targets
- [x] 2.3 Run `make install` in `plugins/opencode/` and verify `which orcai-opencode` and `~/.config/orcai/wrappers/opencode.yaml` exist

## 3. orcai Core — CliAdapter env var support

- [x] 3.1 Update `CliAdapter.Execute` in orcai to pass all `vars` as `ORCAI_<KEY>=<value>` env vars on the subprocess (inheriting current env)
- [x] 3.2 Add tests `TestCliAdapter_Execute_VarsAsEnv` and `TestCliAdapter_Execute_FilterViaEnv` to verify env var passing and shell-based jq pattern
- [x] 3.3 Verify all existing plugin tests still pass

## 4. jq Plugin — Pure YAML Sidecar (no Go binary)

- [x] 4.1 Create `plugins/jq/jq.yaml`: `command: sh`, `args: ["-c", "jq \"${ORCAI_FILTER:-.}\""]`; document `vars.filter` → `ORCAI_FILTER`
- [x] 4.2 Create `plugins/jq/Makefile` with `install` (copies YAML only) and `test` targets
- [x] 4.3 Run `make install` in `plugins/jq/` and verify `~/.config/orcai/wrappers/jq.yaml` exists

## 5. Example Pipelines

- [ ] 5.1 Create `examples/opencode-local.pipeline.yaml`: single `opencode` step, `model: ollama/llama3.2`, static prompt, prerequisite comment block
- [ ] 5.2 Create `examples/jq-transform.pipeline.yaml`: `builtin.http_get` step fetching `https://httpbin.org/json`, `jq` step with `filter: .slideshow.title`, `builtin.assert` step verifying result contains text
- [x] 5.3 Manually run `orcai pipeline run examples/opencode-local.pipeline.yaml` and verify exit 0 and non-empty output
- [x] 5.4 Manually run `orcai pipeline run examples/jq-transform.pipeline.yaml` and verify exit 0 and extracted value in output

## 6. E2E Test Harness

- [x] 6.1 Create `tests/run_all.sh`: iterate `tests/test-*.sh`, run each, print PASS/SKIP/FAIL, exit 1 if any FAIL
- [x] 6.2 Create `tests/test-ollama-simple.sh`: check prereqs (`orcai`, `orcai-ollama`, `ollama`), run `llama3.2-prompt.pipeline.yaml`, assert exit 0 and non-empty output
- [x] 6.3 Create `tests/test-jq-transform.sh`: check prereqs (`orcai`, `jq`), run `jq-transform.pipeline.yaml`, assert exit 0 and non-empty output
- [x] 6.4 Create `tests/test-opencode-local.sh`: check prereqs (`orcai`, `orcai-opencode`, `opencode`, `ollama`), run `opencode-local.pipeline.yaml`, assert exit 0 and non-empty output
- [x] 6.5 Create `tests/test-builtin-assert.sh`: create an inline passing assert pipeline using `builtin.http_get` + `builtin.assert contains:Sample`, run it, assert exit 0
- [x] 6.6 Make all test scripts executable (`chmod +x`)

## 7. Root Makefile & README Updates

- [x] 7.1 Add `test-e2e` target to root `Makefile` (create `Makefile` if absent) that runs `bash tests/run_all.sh` with a prerequisites reminder
- [x] 7.2 Update root `README.md`: add opencode and jq to the plugins table; add an "End-to-End Tests" section describing `make test-e2e` and prerequisites
- [x] 7.3 Commit and push all changes to `adam-stokes/orcai-plugins`

## 8. End-to-End Validation

- [x] 8.1 Run `make test-e2e` from repo root — verify PASS for all available tests, SKIP for any missing tools
- [x] 8.2 Verify `openspec status --change opencode-plugin-e2e-expansion` shows all tasks complete

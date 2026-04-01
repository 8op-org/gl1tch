PLUGINS_DIR := ../glitch-plugins

.PHONY: build run plugins debug debug-tmux

build:
	go build -o glitch .

run: build
	./glitch

# Build and install all core provider plugins from ../glitch-plugins.
# Run this during development whenever plugin code changes.
plugins:
	$(MAKE) -C $(PLUGINS_DIR) install

debug:
	dlv debug . -- server

debug-tmux:
	tmux new-session -d -s glitch-debug 2>/dev/null || true
	tmux send-keys -t glitch-debug "dlv debug . -- server" Enter

.PHONY: test-integration test-integration-astro

# Run all integration tests (requires ollama running with at least llama3.2).
# Override model with: GLITCH_SMOKE_MODEL=llama3.2:1b make test-integration
test-integration:
	go test -tags=integration -timeout=15m -v ./internal/pipeline/...

# Run only the Astro pipeline generation + build test.
# Requires: ollama (llama3.2 or GLITCH_SMOKE_MODEL), bun, shell sidecar.
test-integration-astro:
	go test -tags=integration -timeout=15m -v -run TestAstroPipelineGenAndRun ./internal/pipeline/...

# Run only the docs audit test (reads real code + docs, checks for MISSING: lines).
test-integration-docs-audit:
	go test -tags=integration -timeout=5m -v -run TestDocsAudit ./internal/pipeline/...

# Run the full docs-improve end-to-end test (isolated temp repo, real commit).
# This is the full cron loop test: audit → pick → rewrite → polish → write → commit.
# Requires: ollama + claude API key.
test-integration-docs-improve:
	go test -tags=integration -timeout=15m -v -run TestDocsImprove ./internal/pipeline/...

# Run only the pick step test (fast — just validates structured output format).
test-integration-docs-pick:
	go test -tags=integration -timeout=5m -v -run TestDocsImprove_PickProducesStructuredOutput ./internal/pipeline/...

# Run the docs-improve pipeline manually against the real repo right now.
# Use this to trigger a single improvement cycle on demand.
run-docs-improve:
	glitch pipeline run docs-improve

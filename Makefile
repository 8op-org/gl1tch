BINARY := orcai

.PHONY: build run test clean debug-build debug debug-connect debug-tmux

all: build run

run: build
	-tmux kill-session -t orcai 2>/dev/null
	rm -f ~/.config/orcai/layout.yaml ~/.config/orcai/keybindings.yaml
	bin/$(BINARY)

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

clean:
	rm -f bin/$(BINARY) bin/$(BINARY)-debug

debug-build:
	go build -gcflags="all=-N -l" -o bin/$(BINARY)-debug .

debug: debug-build
	@echo "Delve listening on :2345 — connect with: make debug-connect"
	dlv exec --headless --listen=:2345 --api-version=2 ./bin/$(BINARY)-debug

debug-connect:
	dlv connect :2345

debug-tmux: debug-build
	@bash $(shell pwd)/scripts/debug-tmux.sh

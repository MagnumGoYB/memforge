GO ?= go
MEMFORGE_CACHE_DIR ?= $(CURDIR)/.cache/memforge
GOCACHE ?= $(MEMFORGE_CACHE_DIR)/go-build
GOMODCACHE ?= $(MEMFORGE_CACHE_DIR)/go-mod
COMMIT_MSG_FILE ?= .git/COMMIT_EDITMSG
COMMIT_RANGE ?=

.PHONY: setup check test test-packages test-harness vet build run validate validate-pr-body commitlint commitlint-range build-plugin-binaries package-plugins smoke-plugin-runtime

setup:
	git config core.hooksPath .githooks

check:
	@test -z "$$(gofmt -l $$(git ls-files '*.go'))"
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) vet ./...

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./...

test-packages:
	@test -n "$(PKGS)" || (echo 'PKGS is required, for example: make test-packages PKGS="./internal/index ./internal/compiler"' >&2; exit 2)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test $(PKGS)

test-harness:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./harness

vet:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) vet ./...

build:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build ./cmd/memforge

run:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) run ./cmd/memforge $(ARGS)

validate-pr-body:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) run ./tools/validate-pr-body

commitlint:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) run ./tools/commitlint --edit "$(COMMIT_MSG_FILE)"

commitlint-range:
	@test -n "$(COMMIT_RANGE)" || (echo 'COMMIT_RANGE is required, for example: make commitlint-range COMMIT_RANGE="origin/main..HEAD"' >&2; exit 2)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) run ./tools/commitlint --range "$(COMMIT_RANGE)"

build-plugin-binaries:
	mkdir -p dist/bin
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o dist/bin/memforge-darwin-arm64 ./cmd/memforge
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o dist/bin/memforge-darwin-amd64 ./cmd/memforge
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o dist/bin/memforge-linux-arm64 ./cmd/memforge
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o dist/bin/memforge-linux-amd64 ./cmd/memforge
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o dist/bin/memforge-windows-arm64.exe ./cmd/memforge
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o dist/bin/memforge-windows-amd64.exe ./cmd/memforge

package-plugins: build-plugin-binaries
	./tools/package_plugins.sh

smoke-plugin-runtime:
	smoke_root="$(MEMFORGE_CACHE_DIR)/smoke-plugin/claude-code/memforge"; \
	platform="$$($(GO) env GOOS)-$$($(GO) env GOARCH)"; \
	rm -rf "$$smoke_root"; \
	mkdir -p "$$smoke_root/bin/$$platform"; \
	cp plugins/claude-code/memforge/bin/memforge-mcp-launcher.js "$$smoke_root/bin/memforge-mcp-launcher.js"; \
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -trimpath -o "$$smoke_root/bin/$$platform/memforge" ./cmd/memforge; \
	printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | MEMFORGE_HOME=$$(mktemp -d) MEMFORGE_PLUGIN_ROOT="$$smoke_root" node "$$smoke_root/bin/memforge-mcp-launcher.js"

validate: check test build

# Barikoi Go SDK — developer commands (the equivalent of the TypeScript
# repo's package.json scripts; see CONTRIBUTING.md).
#
# Everything runs inside Docker (golang:1.22) when the local machine has no
# Go toolchain, so contributors only need Docker + make. With Go installed,
# the targets use the local toolchain directly.

GO ?= go
DOCKER_GO_IMAGE := golang:1.22

# Packages whose coverage is measured by `test`/`cover` (the generated
# client and the example program are excluded, mirroring the TypeScript
# repo's coverage config).
COVERPKGS := github.com/barikoi/barikoiapis-golang,github.com/barikoi/barikoiapis-golang/client

# Use the Docker-based toolchain when `go` is not on PATH. BARIKOI_API_KEY
# is forwarded so integration tests and the example work unchanged.
GO_IN_DOCKER := docker run --rm -v $(CURDIR):/app -w /app -e BARIKOI_API_KEY $(DOCKER_GO_IMAGE)

ifeq ($(shell command -v $(GO) 2>/dev/null),)
	GO := $(GO_IN_DOCKER) go
	GOFMT := $(GO_IN_DOCKER) gofmt
else
	GOFMT := gofmt
endif

# Export variables from .env if present (does not override the environment).
define load_env
	test -f .env && set -a && . ./.env && set +a || true
endef

.PHONY: help build test test-integration cover fmt lint check codegen example tidy

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Compile all packages
	$(GO) build ./...

test: ## Run unit tests (httptest stubs; no API key or network needed)
	$(GO) test -race -coverpkg=$(COVERPKGS) -covermode=atomic -cover ./test/ ./client/

test-integration: ## Run live API tests; BARIKOI_API_KEY from env or .env
	@$(load_env); \
	$(GO) test -race -v -run Integration ./test/

cover: ## Unit tests with a coverage report (coverage.out)
	$(GO) test -race -coverpkg=$(COVERPKGS) -covermode=atomic -coverprofile=coverage.out ./test/ ./client/
	$(GO) tool cover -func=coverage.out | tail -1

fmt: ## Format all Go sources
	$(GOFMT) -w .

lint: ## gofmt check + go vet
	@files=$$($(GOFMT) -l .); if [ -n "$$files" ]; then echo "unformatted files:"; echo "$$files"; exit 1; fi
	$(GO) vet ./...

check: lint test ## Everything CI runs

codegen: ## Regenerate gen/ from openapi/barikoi-api-spec.yaml
	./scripts/codegen.sh

example: ## Run examples/basic against the live API (BARIKOI_API_KEY from env or .env)
	@$(load_env); \
	$(GO) run ./examples/basic

tidy: ## go mod tidy
	$(GO) mod tidy

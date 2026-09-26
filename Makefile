# liki monorepo — skill client + engine/counsel MCP services
# Quality gate: make gate == the branch's actual CI contract.

.DEFAULT_GOAL := help

export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache

.PHONY: help build build-skill build-engine-mcp build-engine-rpc clean fmt fmt-check lint-engine lint-python \
        go-version-check lint lint-md check test test-contracts test-engine test-counsel test-engine-mcp \
        full-data verify golden gate hooks version build-archive

build-skill build-archive: ## Build the installable skill archive
	scripts/build-archive.sh

build-engine-mcp: ## Build the engine MCP server
	@cd engine && go build -o ../bin/engine-mcp ./cmd/engine-mcp/

build-engine-rpc: ## Build the transition JSON-RPC server used by liki-web
	@cd engine && go build -o ../bin/engine-rpc ./cmd/engine-rpc/

build: build-skill build-engine-mcp build-engine-rpc ## Build skill archive and both engine runtimes

fmt: ## Format Go sources
	@cd engine && gofmt -w .

fmt-check: ## Reject unformatted Go sources
	@cd engine && test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

go-version-check: ## Require the security-supported Go toolchain
	scripts/check-go-version.sh

lint-md: ## Markdown structure and style check
	scripts/lint-md.sh

lint-engine: go-version-check ## Engine static analysis
	@cd engine && golangci-lint run ./... && go vet ./...

lint-python: ## Python static analysis
	scripts/lint-python.sh

lint: lint-md lint-engine lint-python ## Full lint entry point

check: lint-md fmt-check go-version-check ## Static contracts: markdown, format, vet, schema, docs
	python3 -m pytest tests/test_skill_markdown_structure.py tests/test_skill_frontmatter.py -q
	@cd engine && go vet ./...
	python3 scripts/check_schema.py
	python3 scripts/check_docs.py

test-contracts: ## Root skill / rules / documentation contract suite
	python3 -m pytest tests/ -q

test-engine: go-version-check ## Engine unit and integration suites
	@cd engine && go test -count=1 ./...
	@cd engine && go test -tags integration -count=1 -timeout 60s \
		./internal/agent/ ./internal/http/ ./internal/engine/bazi/

test-counsel: ## Counsel MCP integration suite (bootstraps .venv and local engine)
	scripts/test-counsel.sh

test-engine-mcp: ## Engine MCP root/domain protocol smoke
	scripts/test-engine-mcp.sh

full-data: ## 160-question assertion coverage and zero-hit check
	scripts/test-full-data.sh

test: test-contracts test-engine test-counsel test-engine-mcp ## All deterministic tests

verify: test-engine-mcp test-counsel ## Current-architecture MCP integration verification

golden: ## Deterministic multi-source golden suites
	@cd engine && go test -count=1 -run \
		'Golden|External|Matrix|Pattern' ./internal/engine/...

gate: lint-engine lint-python check test full-data build-archive ## Full push gate: lint, static, tests, data coverage, archive

hooks: ## Install repository git hooks
	git config core.hooksPath .githooks

version: ## Bump all runtime CalVer surfaces together
	@python3 scripts/bump_version.py

clean: ## Remove generated root binaries
	rm -f bin/engine-mcp bin/engine-rpc

help: ## List primary targets
	@grep -h '^[a-z][a-z0-9_.-]*:.*##' $(firstword $(MAKEFILE_LIST)) | \
		awk -F':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

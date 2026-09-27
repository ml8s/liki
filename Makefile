# liki monorepo — skill client + engine/counsel MCP services
# Quality gate: make gate == the branch's actual CI contract.
#
# 统一语义（四仓一致）：
#   make lint   静态检查聚合
#   make check  纯静态契约检查（lint+格式+schema+docs+专家一致性；不跑测试、不写盘）
#   make test   本仓全部测试（单元/集成；不含 e2e、不含 build）
#   make gate   check + test + 发布产物（full-data/archive/deployment 工件）＝推送门槛
#   make build  本仓产物（skill 归档 + engine 二进制；镜像由 CI release 构建）
#   make image  本地镜像（调试用，非发布路径）
#   make e2e    系统 E2E 聚合（liki-deploy 编排）

.DEFAULT_GOAL := help

export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache

.PHONY: help build build-skill build-engine-mcp build-engine-rpc clean fmt fmt-check lint-engine lint-python \
        go-version-check lint lint-md check test test-contracts test-engine test-counsel test-engine-mcp \
        full-data golden gate hooks version build-archive build-deployment \
        build-web-skill-bundle agents-validate image image-engine image-counsel image-assembly

build-skill build-archive: ## Build the installable skill archive
	scripts/build-archive.sh

build-web-skill-bundle: build-archive ## Build the liki-web skill bundle (release asset)
	scripts/build-web-skill-bundle.sh

DEPLOYMENT_PROFILES ?= experts single

build-deployment: ## Generate AgentDeployment artifacts for all profiles
	@set -eu; for p in $(DEPLOYMENT_PROFILES); do \
		python3 scripts/generate_deployment.py --profile "$$p" --out dist/agents; \
	done
	@set -eu; for p in $(DEPLOYMENT_PROFILES); do \
		python3 scripts/check_deployment_schema.py --profile "$$p"; \
	done

LIKI_AGENTS_VALIDATOR ?= liki-agents

agents-validate: build-deployment ## Validate generated artifacts with the liki-agents runtime
	@set -eu; for p in $(DEPLOYMENT_PROFILES); do \
		$(LIKI_AGENTS_VALIDATOR) validate -deployment "dist/agents/$$p/deployment.json"; \
	done

build-engine-mcp: ## Build the engine MCP server
	@cd engine && go build -o ../bin/engine-mcp ./cmd/engine-mcp/

build-engine-rpc: ## Build the transition JSON-RPC server used by liki-web
	@cd engine && go build -o ../bin/engine-rpc ./cmd/engine-rpc/

build: build-skill build-engine-mcp build-engine-rpc ## Build local artifacts (skill archive + engine binaries). 镜像由 CI release 构建，非本地。

# ── 镜像（本地调试用，非发布路径）──
# 生产镜像由 CI release 构建并推送 GHCR；以下目标仅用于本地构建/调试。

IMAGE_REGISTRY ?= ghcr.io/ml8s
IMAGE_TAG ?= dev

image-engine: ## Build liki-engine image locally (debug; production uses CI release)
	docker build -f engine/dev/Dockerfile -t $(IMAGE_REGISTRY)/liki-engine:$(IMAGE_TAG) engine

image-counsel: ## Build liki-counsel image locally (debug; production uses CI release)
	docker build -f counsel/Dockerfile -t $(IMAGE_REGISTRY)/liki-counsel:$(IMAGE_TAG) .

image-assembly: build-deployment ## Build assembly images locally (multi + single; pulls liki-agents base from GHCR)
	@rm -rf assembly/dist && cp -r dist/agents assembly/dist
	@docker build -f assembly/Dockerfile --build-arg PROFILE=experts -t $(IMAGE_REGISTRY)/liki-multi-expert:$(IMAGE_TAG) assembly
	@docker build -f assembly/Dockerfile --build-arg PROFILE=single -t $(IMAGE_REGISTRY)/liki-single-expert:$(IMAGE_TAG) assembly
	@rm -rf assembly/dist

image: image-engine image-counsel ## Build all liki-owned runtime images locally (debug)

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

check: lint-md lint-python fmt-check go-version-check ## 纯静态契约检查（lint + Python lint + 格式 + schema + docs + 专家方法论一致性 + 脚本可执行位）。不跑测试、不生成工件。
	python3 scripts/check_schema.py
	python3 scripts/check_docs.py
	python3 scripts/check_workflow.py
	bash scripts/check-expert-methodology.sh
	bash scripts/check-exec-bits.sh

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

golden: ## Deterministic multi-source golden suites
	@cd engine && go test -count=1 -run \
		'Golden|External|Matrix|Pattern' ./internal/engine/...

gate: lint-engine lint-python check test full-data build-archive build-deployment ## Full push gate: lint, static, tests, data coverage, archive, deployment 工件

hooks: ## Install repository git hooks
	git config core.hooksPath .githooks

version: ## Bump all runtime CalVer surfaces together
	@python3 scripts/bump_version.py

clean: ## Remove generated root binaries
	rm -f bin/engine-mcp bin/engine-rpc

help: ## List primary targets
	@grep -h '^[a-z][a-z0-9_.-]*:.*##' $(firstword $(MAKEFILE_LIST)) | \
		awk -F':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

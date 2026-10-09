# liki monorepo — skill client + engine/counsel MCP services
# Quality gate: make gate 是"本地可执行的推送门槛"（lint + 静态契约 + 测试 + 数据覆盖 + 发布工件）。
#   CI 另外强制执行：漏洞扫描（govulncheck/pip-audit/npm）、镜像构建与冒烟、cosign 签名与 attestation、-race 与覆盖率门槛
#   （这些依赖 registry 凭据/签名密钥，不适合本地）。`make vuln` 可在本地手动跑漏洞扫描。
#
# 统一语义（四仓一致）：
#   make lint   静态检查聚合
#   make check  纯静态契约检查（lint+格式+schema+docs+专家一致性；不跑测试、不写盘）
#   make test   本仓全部测试（单元/集成；不含 e2e、不含 build）
#   make gate   check + test + 发布产物（full-data/archive/deployment 工件）＝推送门槛
#   make build  本仓产物（skill 归档 + engine 二进制；镜像由 CI release 构建）
#   make image  本地镜像（调试用，非发布路径）
#   make e2e    系统 E2E（由 liki-deploy 编排，不在本仓提供 target）

.DEFAULT_GOAL := help

export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache

.PHONY: help build build-skill build-engine-mcp build-engine-rpc clean fmt fmt-check lint-engine lint-python \
        go-version-check lint vuln lint-md check test test-contracts test-engine test-counsel test-engine-mcp \
        test-functional verify full-data golden gate hooks version build-archive build-deployment \
        check-actions build-web-skill-bundle build-release-manifest check-release-manifest \
        agents-validate image image-engine image-counsel image-assembly

build-skill build-archive: ## Build the installable skill archive
	scripts/build-archive.sh

build-web-skill-bundle: build-archive ## Build the liki-web skill bundle (release asset)
	scripts/build-web-skill-bundle.sh

build-release-manifest: build-web-skill-bundle build-deployment ## Build the unified liki release manifest
	python3 scripts/build_release_manifest.py

check-release-manifest: ## Verify an existing unified release manifest without writing
	python3 scripts/build_release_manifest.py --check

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
# 装配镜像的 base（本地先 make image 出 liki-agents:dev；CI 用 release pin 覆盖）
BASE_IMAGE ?= $(IMAGE_REGISTRY)/liki-agents:$(IMAGE_TAG)

image-engine: ## Build liki-engine image locally (debug; production uses CI release)
	docker build -f engine/dev/Dockerfile -t $(IMAGE_REGISTRY)/liki-engine:$(IMAGE_TAG) engine

image-counsel: ## Build liki-counsel image locally (debug; production uses CI release)
	docker build -f counsel/Dockerfile -t $(IMAGE_REGISTRY)/liki-counsel:$(IMAGE_TAG) .

image-assembly: build-deployment ## Build assembly images locally (multi + single; pulls liki-agents base from GHCR)
	@rm -rf assembly/dist && mkdir -p assembly/dist/agents && cp -r dist/agents/. assembly/dist/agents/
	@cp -r skills assembly/dist/skills && cp -r expert-packs assembly/dist/expert-packs
	@docker build -f assembly/Dockerfile --build-arg BASE_IMAGE=$(BASE_IMAGE) --build-arg PROFILE=experts -t $(IMAGE_REGISTRY)/liki-multi-expert:$(IMAGE_TAG) assembly
	@docker build -f assembly/Dockerfile --build-arg BASE_IMAGE=$(BASE_IMAGE) --build-arg PROFILE=single -t $(IMAGE_REGISTRY)/liki-single-expert:$(IMAGE_TAG) assembly
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

vuln: ## Vulnerability scan (govulncheck + pip-audit); CI also runs these as required jobs
	@if command -v govulncheck >/dev/null 2>&1; then (cd engine && govulncheck ./...); else echo "govulncheck 未安装，跳过 engine 扫描（CI 会跑）"; fi
	@if command -v pip-audit >/dev/null 2>&1; then pip-audit -r counsel/requirements.lock --require-hashes --disable-pip; else echo "pip-audit 未安装，跳过 Python 扫描（CI 会跑）"; fi

check: lint-md lint-python fmt-check go-version-check check-actions ## 纯静态契约检查（lint + Python lint + 格式 + schema + docs + 专家方法论一致性 + 脚本可执行位）。不跑测试、不生成工件。
	python3 scripts/check_schema.py
	python3 scripts/check_docs.py
	python3 scripts/check_workflow.py
	bash scripts/check-exec-bits.sh
	python3 scripts/sync_expert_packs.py --check

sync-expert-packs: ## 从根 skill 生成 expert-packs 方法论卡（唯一生成入口）
	python3 scripts/sync_expert_packs.py

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

test-functional: ## Rule engine functional/behavior contract suite
	python3 -m pytest tests/test_functional.py -q

verify: test-engine-mcp test-counsel ## Engine/counsel MCP integration verification

full-data: ## 160-question assertion coverage and zero-hit check
	scripts/test-full-data.sh

test: test-contracts test-engine test-counsel test-engine-mcp ## All deterministic tests

golden: ## Deterministic multi-source golden suites
	@cd engine && go test -count=1 -run \
		'Golden|External|Matrix|Pattern' ./internal/engine/...

gate: lint-engine check test full-data build-release-manifest ## Full push gate: lint, static, tests, data coverage, unified release artifacts

check-actions: ## Reject mutable reusable workflow/action references
	python3 scripts/check_actions_pinned.py

hooks: ## Install repository git hooks
	git config core.hooksPath .githooks

version: ## Bump all runtime CalVer surfaces together
	@python3 scripts/bump_version.py

clean: ## Remove generated root binaries
	rm -f bin/engine-mcp bin/engine-rpc

help: ## List primary targets
	@grep -h '^[a-z][a-z0-9_.-]*:.*##' $(firstword $(MAKEFILE_LIST)) | \
		awk -F':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

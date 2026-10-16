# liki monorepo — skill client + engine/counsel MCP services
# Quality gate: make gate 是"本地可执行的源码门槛"（lint + 静态检查 + 测试）。
#   CI 另外强制执行：漏洞扫描（govulncheck/pip-audit/npm）、-race 与覆盖率门槛。
#   发版镜像由 `make candidate` 本地构建；promotion 只做同 digest 转发和签名。
#
# 统一语义（四仓一致）：
#   make lint   静态检查聚合
#   make check  纯静态检查（lint+格式+数据结构+docs+专家一致性；不跑测试、不写盘）
#   make test   本仓全部测试（单元/集成；不含 e2e、不含 build）
#   make gate   check + test ＝推送门槛
#   make release-gate gate + 160题 + runtime validate + 发布工件 ＝发版准备
#   make candidate RELEASE_VERSION=... 本地构建、冒烟并记录 digest ＝发版准备
#   make build  本仓产物（skill 归档 + engine 二进制；镜像仅本地候选构建）
#   make image  本地开发镜像（local tag；candidate 由 make candidate 发布）
#   make e2e    系统 E2E（由 liki-deploy 编排，不在本仓提供 target）

RELEASE_VERSION ?=
export RELEASE_VERSION

.DEFAULT_GOAL := help

export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache
NODE_BIN := $(HOME)/.nvm/versions/node/v$(shell cat .nvmrc)/bin
PYTHON_BIN := $(HOME)/app/python3.12.14/bin
export PATH := $(PYTHON_BIN):$(NODE_BIN):$(HOME)/app/go/bin:$(HOME)/app/go-tools/bin:$(HOME)/go/bin:$(PATH)
export GOTOOLCHAIN ?= go1.26.9

.PHONY: help build build-skill build-engine-mcp clean fmt fmt-check lint-engine lint-python smoke \
        go-version-check runtime-version-check lint vuln lint-md check test test-content test-engine test-engine-coverage test-counsel test-engine-mcp \
        full-data golden gate release-gate hooks version build-archive build-deployment \
        workbuddy-package \
        check-actions build-web-skill-bundle build-release-manifest check-release-manifest candidate \
        runtime-validate image image-engine image-counsel image-assembly

build-skill build-archive: ## Build the installable skill archive
	scripts/build-archive.sh

build-web-skill-bundle: build-archive ## Build the liki-web skill bundle (release asset)
	scripts/build-web-skill-bundle.sh

build-release-manifest: build-web-skill-bundle build-deployment ## Build the unified liki release manifest
	python3 scripts/build_release_manifest.py

check-release-manifest: ## Verify an existing unified release manifest without writing
	python3 scripts/build_release_manifest.py --check

DEPLOYMENT_PROFILES ?= experts

build-deployment: ## Generate AgentDeployment artifacts for all profiles
	@set -eu; for p in $(DEPLOYMENT_PROFILES); do \
		python3 scripts/generate_deployment.py --profile "$$p" --out dist/agents; \
	done

LIKI_AGENT_RUNTIME_VERSION ?= $(shell python3 scripts/agent_runtime_pin.py --version)
LIKI_AGENT_RUNTIME_IMAGE ?= $(shell python3 scripts/agent_runtime_pin.py --ref)
# 覆盖为本地二进制时，跳过容器内 pinned 运行时（仅用于离线开发）
LIKI_AGENT_RUNTIME_VALIDATOR ?=

runtime-validate: build-deployment ## Validate generated artifacts with the pinned liki-agent-runtime runtime (override LIKI_AGENT_RUNTIME_VALIDATOR=<bin>)
	@set -eu; \
	if [ -n "$(LIKI_AGENT_RUNTIME_VALIDATOR)" ]; then \
		for p in $(DEPLOYMENT_PROFILES); do \
			$(LIKI_AGENT_RUNTIME_VALIDATOR) validate -deployment "dist/agents/$$p/deployment.json"; \
		done; \
	else \
		for p in $(DEPLOYMENT_PROFILES); do \
			docker run --rm -v "$$PWD/dist/agents:/deployment:ro" \
				-e LIKI_AGENT_RUNTIME_DEPLOYMENT_FILE=/deployment/$$p/deployment.json \
				--entrypoint /liki-agent-runtime $(LIKI_AGENT_RUNTIME_IMAGE) \
				validate -deployment /deployment/$$p/deployment.json; \
		done; \
	fi

build-engine-mcp: ## Build the engine MCP server
	@cd engine && go build -o ../bin/engine-mcp ./cmd/engine-mcp/

build: build-skill build-engine-mcp ## Build local artifacts (skill archive + engine binary). 镜像仅由本地 candidate 构建。

workbuddy-package: ## 打包 WorkBuddy 连接器提交包到 dist/workbuddy/（上传用）
	scripts/package-workbuddy-connectors.sh

# ── 镜像（本地候选，非发布路径）──
# 生产镜像由本地 `make candidate` 构建并推送 GHCR；CI promotion 只转发已验证 digest。

IMAGE_REGISTRY ?= ghcr.io/ml8s
IMAGE_TAG ?= local
RUNTIME_VERSION ?= $(shell cat skills/liki/VERSION.txt)
# 装配镜像的 base（本地候选可用 liki-agent-runtime:local；CI 用 release pin 覆盖）
BASE_IMAGE ?= $(IMAGE_REGISTRY)/liki-agent-runtime:$(IMAGE_TAG)

image-engine: ## Build the local engine development image
	docker build -f engine/dev/Dockerfile -t $(IMAGE_REGISTRY)/liki-engine:$(IMAGE_TAG) engine

image-counsel: ## Build the local counsel development image
	docker build -f counsel/Dockerfile -t $(IMAGE_REGISTRY)/liki-counsel:$(IMAGE_TAG) .

image-assembly: build-deployment ## Build the production experts assembly image (pulls liki-agent-runtime base from GHCR)
	@rm -rf assembly/dist && mkdir -p assembly/dist/agents && cp -r dist/agents/. assembly/dist/agents/
	@cp -r skills assembly/dist/skills && cp -r expert-packs assembly/dist/expert-packs
	@docker build -f assembly/Dockerfile --build-arg BASE_IMAGE=$(BASE_IMAGE) --build-arg VERSION=$(RUNTIME_VERSION) --build-arg PROFILE=experts -t $(IMAGE_REGISTRY)/liki-experts:$(IMAGE_TAG) assembly
	@rm -rf assembly/dist

image: image-engine image-counsel ## Build all liki-owned local candidate images

candidate: ## Build, smoke, and record candidate images in the local registry (requires RELEASE_VERSION)
	scripts/build-candidate.sh

candidate-publish: ## Publish a system-tested local candidate to GHCR
	scripts/publish-candidate.sh

smoke: image ## Smoke engine RPC/MCP and counsel MCP images
	IMAGE_TAG=$(IMAGE_TAG) scripts/smoke-image.sh

fmt: ## Format Go sources
	@cd engine && gofmt -w .

fmt-check: ## Reject unformatted Go sources
	@cd engine && test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

go-version-check: ## Require the security-supported Go toolchain
	scripts/check-go-version.sh

runtime-version-check: ## 本地 node/python 必须等于 .nvmrc/.python-version（与 CI 同源）
	scripts/check-runtime-versions.sh

lint-md: ## Markdown structure and style check
	scripts/lint-md.sh

lint-engine: go-version-check ## Engine static analysis (pinned golangci-lint)
	scripts/lint-engine.sh

lint-python: ## Python static analysis
	scripts/lint-python.sh

lint: lint-md lint-engine lint-python ## Full lint entry point

vuln: ## Vulnerability scan (govulncheck + pip-audit); CI also runs these as required jobs
	@if command -v govulncheck >/dev/null 2>&1; then (cd engine && govulncheck ./...); else echo "govulncheck 未安装，跳过 engine 扫描（CI 会跑）"; fi
	@if command -v pip-audit >/dev/null 2>&1; then pip-audit -r counsel/requirements.lock --require-hashes --disable-pip; else echo "pip-audit 未安装，跳过 Python 扫描（CI 会跑）"; fi

check: lint-md lint-python fmt-check go-version-check runtime-version-check check-actions ## 纯静态检查（lint + Python lint + 格式 + 数据结构 + docs + 专家方法论一致性 + 脚本可执行位）。不跑测试、不生成工件。
	python3 scripts/check_factor_data.py
	python3 scripts/check_docs.py
	python3 scripts/check_doc_paths.py
	python3 scripts/check_workflow.py
	python3 scripts/check_mcp_schema.py --check
	python3 scripts/check_expert_docs.py
	python3 scripts/agent_runtime_pin.py --check
	bash scripts/check-exec-bits.sh
	python3 scripts/sync_expert_packs.py --check
	python3 scripts/check_workbuddy_connectors.py

sync-expert-packs: ## 从根 skill 生成 expert-packs 方法论卡（唯一生成入口）
	python3 scripts/sync_expert_packs.py

test-content: ## Root skill / rules / documentation suite
	python3 -m pytest tests/ -q

test-engine: go-version-check ## Engine unit and integration suites
	@cd engine && go test -count=1 ./...
	@cd engine && go test -tags integration -count=1 -timeout 60s \
		./internal/mcpcontract/ ./internal/http/ ./internal/engine/bazi/

test-engine-coverage: go-version-check ## Engine unit coverage floor (shared with CI)
	scripts/test-engine-coverage.sh

test-counsel: ## Counsel MCP integration suite (bootstraps .venv and local engine)
	scripts/test-counsel.sh

test-engine-mcp: ## Engine MCP root/domain protocol smoke
	scripts/test-engine-mcp.sh

full-data: ## 160-question assertion coverage and zero-hit check
	scripts/test-full-data.sh

test: test-content test-engine test-counsel test-engine-mcp ## All deterministic tests
test-python-coverage: ## Run root Python domain suite with coverage
	coverage erase
	coverage run --source=counsel/app -m pytest tests -q
	coverage report --fail-under=65
	coverage xml
	coverage html


golden: ## Deterministic multi-source golden suites
	@cd engine && go test -count=1 -run \
		'Golden|External|Matrix|Pattern' ./internal/engine/...

gate: lint-engine check test ## Full source push gate: lint, static checks, and tests

release-gate: gate full-data runtime-validate build-release-manifest ## Release preparation gate: source gate plus release artifacts

check-actions: ## Reject mutable reusable workflow/action references
	python3 scripts/check_actions_pinned.py

hooks: ## Install repository git hooks
	git config core.hooksPath .githooks

version: ## Bump all runtime CalVer surfaces together
	@python3 scripts/bump_version.py

clean: ## Remove generated root binaries
	rm -f bin/engine-mcp

help: ## List primary targets
	@grep -h '^[a-z][a-z0-9_.-]*:.*##' $(firstword $(MAKEFILE_LIST)) | \
		awk -F':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

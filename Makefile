# liki monorepo — skill client + engine/counsel MCP services
# Quality gate: make gate 是"本地可执行的源码门槛"（lint + 静态检查 + 测试）。
#   CI 另外强制执行：漏洞扫描（govulncheck/pip-audit/npm）、-race 与覆盖率门槛。
#   镜像只在本地构建，tag = git sha；CI 不构建镜像。
#
# 统一语义（四仓一致）：
#   make lint   静态检查聚合
#   make check  静态检查（lint+格式+数据结构+docs+专家一致性；临时构建分发包后校验）
#   make test   本仓全部测试（单元/集成；不含 e2e、不含 build）
#   make gate   check + test ＝推送门槛
#   make eval-mingli160 手动 skill 校准/评估报告；不属于测试或发布门禁
#   make build  本仓产物（skill 归档 + engine 二进制）
#   make image  本地镜像（tag 由 IMAGE_TAG 指定）
#   make e2e    系统 E2E（由 liki-deploy 编排，不在本仓提供 target）

.DEFAULT_GOAL := help

export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache
NODE_BIN := $(HOME)/.nvm/versions/node/v$(shell cat .nvmrc)/bin
PYTHON_BIN := $(CURDIR)/counsel/.venv/bin
export PATH := $(PYTHON_BIN):$(NODE_BIN):$(HOME)/sdk/go1.26.11/bin:$(HOME)/app/go-tools/bin:$(HOME)/go/bin:$(PATH)
export GOTOOLCHAIN ?= go1.26.11
# Stale host GOROOT would force the 1.26.10 standard library onto the 1.26.11 toolchain.
unexport GOROOT

.PHONY: help build build-skill build-engine-mcp clean fmt fmt-check lint-engine lint-python smoke \
        go-version-check runtime-version-check lint security-engine security-python python-ci-deps lint-md check test test-content test-engine \
        test-engine-package test-engine-coverage test-engine-golden test-counsel test-engine-mcp verify-city-data \
        test-python-coverage \
        eval-mingli160 golden gate hooks version build-archive build-deployment \
        workbuddy-package skillhub-package distribution-package \
        workbuddy-avatars \
        brand-marks \
        sync-domain-language check-domain-language check-actions check-content \
        runtime-validate image image-engine image-counsel image-liki-mcp image-assembly \
        smoke smoke-engine smoke-counsel smoke-liki-mcp

build-skill build-archive: ## Build the installable skill archive
	scripts/build-archive.sh

DEPLOYMENT_PROFILES ?= liki-expert

build-deployment: ## Generate AgentDeployment artifacts for all profiles
	@rm -rf dist/liki-expert && mkdir -p dist
	@set -eu; for p in $(DEPLOYMENT_PROFILES); do \
		python3 scripts/generate_deployment.py --profile "$$p" --out dist; \
	done

AGENT_RUNTIME_IMAGE ?=
# 覆盖为本地二进制时，跳过容器内 pinned 运行时（仅用于离线开发）
AGENT_RUNTIME_VALIDATOR ?=

runtime-validate: build-deployment ## Validate deployment artifacts (AGENT_RUNTIME_IMAGE=<ref> or AGENT_RUNTIME_VALIDATOR=<bin>)
	@set -eu; \
	if [ -z "$(AGENT_RUNTIME_VALIDATOR)" ] && [ -z "$(AGENT_RUNTIME_IMAGE)" ]; then \
		echo "set AGENT_RUNTIME_IMAGE=agent-runtime-local/agent-runtime:<sha> (or AGENT_RUNTIME_VALIDATOR)" >&2; exit 2; \
	fi; \
	if [ -n "$(AGENT_RUNTIME_VALIDATOR)" ]; then \
		for p in $(DEPLOYMENT_PROFILES); do \
			$(AGENT_RUNTIME_VALIDATOR) validate -deployment "dist/$$p/deployment.json"; \
		done; \
	else \
		for p in $(DEPLOYMENT_PROFILES); do \
			docker run --rm -v "$$PWD/dist:/deployment:ro" \
				-e AGENT_RUNTIME_DEPLOYMENT_FILE=/deployment/$$p/deployment.json \
				--entrypoint /agent-runtime $(AGENT_RUNTIME_IMAGE) \
				validate -deployment /deployment/$$p/deployment.json; \
		done; \
	fi

build-engine-mcp: ## Build the engine MCP server
	@cd engine && go build -o ../bin/engine-mcp ./cmd/engine-mcp/

build: build-skill build-engine-mcp ## Build local artifacts (skill archive + engine binary).

workbuddy-package: ## Build free and pro WorkBuddy packages
	python3 scripts/build_channel_package.py --channel workbuddy --variant free
	python3 scripts/build_channel_package.py --channel workbuddy --variant pro

skillhub-package: ## Build free and pro SkillHub packages
	python3 scripts/build_channel_package.py --channel skillhub --variant free
	python3 scripts/build_channel_package.py --channel skillhub --variant pro

distribution-package: workbuddy-package skillhub-package ## Build all channel packages

workbuddy-avatars: ## 从包内 avatars/expert.svg 重新栅格化专家头像（512×512 PNG）
	scripts/build-workbuddy-avatars.sh

brand-marks: ## 用阿里妈妈刀隶体重建品牌图形（1 张专家头像矢量源 + 2 个连接器图标），并栅格化专家头像
	python3 scripts/gen_brand_marks.py
	scripts/build-workbuddy-avatars.sh

# ── 镜像（本地构建）──
# 本地 make image 只打 IMAGE_TAG；部署由 liki-deploy 按 sha 组装。

IMAGE_REGISTRY ?= liki-local
IMAGE_TAG ?= local
RUNTIME_VERSION ?= $(shell cat skills/liki/VERSION.txt)
# 装配镜像的 base：由调用方传入本地 runtime 镜像引用。
BASE_IMAGE ?=

image-engine: ## Build the local engine development image
	docker build -f engine/dev/Dockerfile --build-arg VERSION=$(RUNTIME_VERSION) -t $(IMAGE_REGISTRY)/liki-engine:$(IMAGE_TAG) engine

image-counsel: ## Build the local counsel development image
	docker build -f counsel/Dockerfile --build-arg VERSION=$(RUNTIME_VERSION) -t $(IMAGE_REGISTRY)/liki-counsel:$(IMAGE_TAG) .

image-liki-mcp: ## Build the LIKI product MCP gateway image
	docker build -f liki-mcp/Dockerfile -t $(IMAGE_REGISTRY)/liki-mcp:$(IMAGE_TAG) liki-mcp

image-assembly: build-deployment ## Build the aggregate expert image (BASE_IMAGE=<local runtime ref> required)
	@test -n "$(BASE_IMAGE)" || { echo "set BASE_IMAGE=agent-runtime-local/agent-runtime:<sha>" >&2; exit 2; }
	@rm -rf assembly/dist && mkdir -p assembly/dist/deployment && cp -r "dist/$(DEPLOYMENT_PROFILES)" "assembly/dist/deployment/$(DEPLOYMENT_PROFILES)"
	@cp -r skills assembly/dist/skills
	@docker build -f assembly/Dockerfile --build-arg BASE_IMAGE=$(BASE_IMAGE) --build-arg VERSION=$(RUNTIME_VERSION) -t $(IMAGE_REGISTRY)/liki-expert:$(IMAGE_TAG) assembly
	@rm -rf assembly/dist

image: image-engine image-counsel image-liki-mcp ## Build all liki-owned local images

smoke: image ## Smoke ENGINE, COUNSEL, and LIKI-MCP images separately
	IMAGE_TAG=$(IMAGE_TAG) scripts/smoke-image.sh

smoke-engine: image-engine ## Smoke the ENGINE image
	IMAGE_TAG=$(IMAGE_TAG) scripts/smoke-image.sh engine

smoke-counsel: image-counsel ## Smoke the COUNSEL image
	IMAGE_TAG=$(IMAGE_TAG) scripts/smoke-image.sh counsel

smoke-liki-mcp: image-liki-mcp ## Smoke the LIKI-MCP image
	IMAGE_TAG=$(IMAGE_TAG) scripts/smoke-image.sh liki-mcp

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

security-engine: ## Run the pinned engine vulnerability scanner
	go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
	"$(shell go env GOPATH)/bin/govulncheck" ./engine/...

security-python: ## Run the pinned Python dependency audit
	python3 -m pip install pip-audit==2.10.0
	pip-audit -r counsel/requirements.lock --require-hashes --disable-pip

python-ci-deps: ## Install the single pinned Python dependency set for check/test
	uv venv --allow-existing counsel/.venv --python "$$(cat .python-version)"
	uv pip install --python counsel/.venv/bin/python -r scripts/requirements-ci.txt

check: lint-md lint-python fmt-check go-version-check runtime-version-check check-actions check-content ## 静态检查（lint + Python lint + 格式 + 数据结构 + docs + 领域语言；dist 中临时构建分发包后校验）。不跑测试。
	python3 scripts/check_factor_data.py
	python3 scripts/check_docs.py
	python3 scripts/check_doc_paths.py
	python3 scripts/sync_domain_language.py --check
	python3 scripts/check_domain_language.py
	python3 scripts/check_workflow.py
	python3 scripts/check_mcp_schema.py --check
	bash scripts/check-exec-bits.sh


check-content: distribution-package ## Build ephemeral packages and verify every channel deterministically.
	@python3 scripts/check_expert_docs.py
	python3 scripts/build_channel_package.py --channel workbuddy --variant free --check
	python3 scripts/build_channel_package.py --channel workbuddy --variant pro --check
	python3 scripts/build_channel_package.py --channel skillhub --variant free --check
	python3 scripts/build_channel_package.py --channel skillhub --variant pro --check

sync-domain-language: ## 从 tools/domain-language.json 生成 docs/DOMAIN_LANGUAGE.md
	python3 scripts/sync_domain_language.py

P ?= ./...
BASE ?= baseline
OLD ?= baseline
NEW ?= current
BENCH_COUNT ?= 5
BENCHTIME ?= 1s

test-content: test-python-coverage ## Root Python contract suite and Counsel integration coverage

test-engine-package: go-version-check ## Run one engine package, e.g. P=./internal/engine/bazi
	@cd engine && go test -count=1 $(P)

test-engine: go-version-check ## Engine unit and integration suites
	@cd engine && go test -count=1 ./...

bench-engine: go-version-check ## Run focused engine and MCP boundary benchmarks
	@cd engine && go test -run '^$$' -bench 'Benchmark(MCP|ComputeChart)' \
		-benchmem ./internal/mcpcontract ./internal/engine/bazi \
		./internal/engine/qimen ./internal/engine/ziwei

bench-engine-save: go-version-check ## Save a reproducible benchmark baseline under .benchmarks
	@mkdir -p .benchmarks
	@cd engine && go test -run '^$$' -bench 'Benchmark(MCP|ComputeChart)' \
		-count=$(BENCH_COUNT) -benchtime=$(BENCHTIME) -benchmem \
		./internal/mcpcontract ./internal/engine/bazi ./internal/engine/qimen \
		./internal/engine/ziwei > ../.benchmarks/$(BASE).bench
	@echo "[bench-engine-save] wrote .benchmarks/$(BASE).bench"

bench-engine-compare: ## Compare two saved baselines with benchstat (or diff if unavailable)
	@if command -v benchstat >/dev/null 2>&1; then \
		benchstat .benchmarks/$(OLD).bench .benchmarks/$(NEW).bench; \
	else \
		echo "benchstat is not installed; showing unified diff instead"; \
		diff -u .benchmarks/$(OLD).bench .benchmarks/$(NEW).bench; \
	fi

bench-engine-1cpu: go-version-check ## Run focused benchmarks with Go limited to one logical CPU
	@cd engine && GOMAXPROCS=1 go test -run '^$$' -bench 'Benchmark(MCP|ComputeChart)' \
		-benchmem ./internal/mcpcontract ./internal/engine/bazi \
		./internal/engine/qimen ./internal/engine/ziwei

test-engine-coverage: go-version-check ## Engine unit coverage floor (shared with CI)
	scripts/test-engine-coverage.sh

verify-city-data: ## Verify embedded city data can be rebuilt byte-for-byte
	python3 scripts/build_city_data.py --verify

test-counsel: ## Counsel MCP integration suite (bootstraps .venv and local engine)
	scripts/test-counsel.sh

test-engine-mcp: ## Engine MCP root/domain protocol smoke
	scripts/test-engine-mcp.sh

eval-mingli160: ## Manual Mingli160 skill calibration/evaluation report (not a gate)
	scripts/eval-mingli160.sh

test: test-content test-engine-coverage test-engine-mcp ## All deterministic tests
test-python-coverage: ## Run root Python domain suite with coverage
	scripts/test-python-coverage.sh


test-engine-golden: ## Focused engine golden/oracle/matrix subset
	@cd engine && go test -count=1 -run \
		'Golden|External|Matrix|Pattern' ./internal/engine/...

golden: test-engine-golden ## Backward-compatible alias for the focused golden suite

gate: lint-engine python-ci-deps check verify-city-data build-archive build-deployment test ## Full source push gate: same checks and tests as CI


check-actions: ## Reject mutable reusable workflow/action references
	python3 scripts/check_actions_pinned.py

hooks: ## Install repository git hooks
	git config core.hooksPath .githooks

version: ## Bump all runtime CalVer surfaces together
	@python3 scripts/bump_version.py

python-env: ## Provision the exact CI Python (.python-version) via uv + counsel venv
	uv python install $$(cat .python-version)
	uv venv --allow-existing counsel/.venv --python $$(cat .python-version)
	uv pip install --python counsel/.venv/bin/python -r counsel/requirements.lock --require-hashes
	uv pip install --python counsel/.venv/bin/python -r counsel/requirements-dev.txt

clean: ## Remove generated root binaries
	rm -f bin/engine-mcp

help: ## List primary targets
	@grep -h '^[a-z][a-z0-9_.-]*:.*##' $(firstword $(MAKEFILE_LIST)) | \
		awk -F':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

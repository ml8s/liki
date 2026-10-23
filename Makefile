# liki monorepo — skill client + engine/counsel MCP services
# Quality gate: make gate 是"本地可执行的源码门槛"（lint + 静态检查 + 测试）。
#   CI 另外强制执行：漏洞扫描（govulncheck/pip-audit/npm）、-race 与覆盖率门槛。
#   镜像只在本地构建，tag = git sha；CI 不构建镜像。
#
# 统一语义（四仓一致）：
#   make lint   静态检查聚合
#   make check  纯静态检查（lint+格式+数据结构+docs+专家一致性；不跑测试、不写盘）
#   make test   本仓全部测试（单元/集成；不含 e2e、不含 build）
#   make gate   check + test ＝推送门槛
#   make eval-mingli160 手动 skill 校准/评估报告；不属于测试或发布门禁
#   make build  本仓产物（skill 归档 + engine 二进制）
#   make image  本地镜像（tag 由 IMAGE_TAG 指定）
#   make e2e    系统 E2E（由 liki-deploy 编排，不在本仓提供 target）

RELEASE_VERSION ?=
export RELEASE_VERSION

.DEFAULT_GOAL := help

export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache
NODE_BIN := $(HOME)/.nvm/versions/node/v$(shell cat .nvmrc)/bin
PYTHON_BIN := $(CURDIR)/counsel/.venv/bin
export PATH := $(PYTHON_BIN):$(NODE_BIN):$(HOME)/app/go1.26.10/bin:$(HOME)/app/go-tools/bin:$(HOME)/go/bin:$(PATH)
export GOTOOLCHAIN ?= go1.26.10

.PHONY: help build build-skill build-engine-mcp clean fmt fmt-check lint-engine lint-python smoke \
        go-version-check runtime-version-check lint vuln lint-md check test test-content test-engine \
        test-engine-package test-engine-coverage test-engine-golden test-counsel test-engine-mcp \
        test-python-coverage \
        eval-mingli160 golden gate hooks version build-archive build-deployment \
        workbuddy-package \
        workbuddy-avatars \
        brand-marks \
        sync-expert-packs sync-domain-language check-domain-language check-actions check-content \
        build-web-skill-bundle \
        runtime-validate image image-engine image-counsel image-assembly

build-skill build-archive: ## Build the installable skill archive
	scripts/build-archive.sh

build-web-skill-bundle: build-archive ## Build the liki-web skill bundle (local artifact)
	scripts/build-web-skill-bundle.sh

DEPLOYMENT_PROFILES ?= natal divination fengshui naming

build-deployment: ## Generate AgentDeployment artifacts for all profiles
	@set -eu; for p in $(DEPLOYMENT_PROFILES); do \
		python3 scripts/generate_deployment.py --profile "$$p" --out dist/agents; \
	done

LIKI_AGENT_RUNTIME_IMAGE ?=
# 覆盖为本地二进制时，跳过容器内 pinned 运行时（仅用于离线开发）
LIKI_AGENT_RUNTIME_VALIDATOR ?=

runtime-validate: build-deployment ## Validate deployment artifacts (LIKI_AGENT_RUNTIME_IMAGE=<ref> or LIKI_AGENT_RUNTIME_VALIDATOR=<bin>)
	@set -eu; \
	if [ -z "$(LIKI_AGENT_RUNTIME_VALIDATOR)" ] && [ -z "$(LIKI_AGENT_RUNTIME_IMAGE)" ]; then \
		echo "set LIKI_AGENT_RUNTIME_IMAGE=liki-local/liki-agent-runtime:<sha> (or LIKI_AGENT_RUNTIME_VALIDATOR)" >&2; exit 2; \
	fi; \
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

build: build-skill build-engine-mcp ## Build local artifacts (skill archive + engine binary).

workbuddy-package: ## 打包全部 WorkBuddy 提交包（连接器 / 技能 / 专家）到 dist/workbuddy/ 并校验
	scripts/package-workbuddy.sh

workbuddy-avatars: ## 从包内 avatars/expert.svg 重新栅格化专家头像（512×512 PNG）
	scripts/build-workbuddy-avatars.sh

brand-marks: ## 用阿里妈妈刀隶体重建品牌图形（4 张头像矢量源 + 2 个连接器图标），并栅格化头像
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

image-assembly: build-deployment ## Build the experts assembly image (BASE_IMAGE=<local runtime ref> required)
	@test -n "$(BASE_IMAGE)" || { echo "set BASE_IMAGE=liki-local/liki-agent-runtime:<sha>" >&2; exit 2; }
	@rm -rf assembly/dist && mkdir -p assembly/dist/agents && cp -r dist/agents/. assembly/dist/agents/
	@cp -r skills assembly/dist/skills && cp -r expert-packs assembly/dist/expert-packs
	@docker build -f assembly/Dockerfile --build-arg BASE_IMAGE=$(BASE_IMAGE) --build-arg VERSION=$(RUNTIME_VERSION) -t $(IMAGE_REGISTRY)/liki-experts:$(IMAGE_TAG) assembly
	@rm -rf assembly/dist

image: image-engine image-counsel ## Build all liki-owned local images

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

check: lint-md lint-python fmt-check go-version-check runtime-version-check check-actions check-content ## 纯静态检查（lint + Python lint + 格式 + 数据结构 + docs + 领域语言；skill/packaging 校验按内容指纹执行）。不跑测试、不提交工件。
	python3 scripts/check_factor_data.py
	python3 scripts/check_docs.py
	python3 scripts/check_doc_paths.py
	python3 scripts/sync_domain_language.py --check
	python3 scripts/check_domain_language.py
	python3 scripts/check_workflow.py
	python3 scripts/check_mcp_schema.py --check
	bash scripts/check-exec-bits.sh


CONTENT_PATHS := skills/liki expert-packs workbuddy agents tools/mcp-tool-catalog.json \
	scripts/sync_expert_packs.py scripts/check_expert_docs.py \
	scripts/check_workbuddy_connectors.py scripts/check_workbuddy_zips.py
CONTENT_FILES := $(shell find $(CONTENT_PATHS) -type f 2>/dev/null | LC_ALL=C sort)
CONTENT_FINGERPRINT := $(shell printf '%s\n' $(CONTENT_FILES) | xargs sha256sum 2>/dev/null | sha256sum | cut -d' ' -f1)
CONTENT_STAMP := .cache/content-check.$(CONTENT_FINGERPRINT)

$(CONTENT_STAMP):
	@mkdir -p $(@D)
	@echo "[check-content] content fingerprint $(CONTENT_FINGERPRINT); running skill and package checks"
	@python3 scripts/check_expert_docs.py
	python3 scripts/sync_expert_packs.py --check
	python3 scripts/check_workbuddy_connectors.py
	python3 scripts/check_workbuddy_zips.py
	@touch $@

check-content: $(CONTENT_STAMP) ## Run skill and WorkBuddy checks only when related content changes.
	@echo "[check-content] content fingerprint unchanged; skipped"

sync-expert-packs: ## 从根 skill 生成 expert-packs 方法论卡（唯一生成入口）
	python3 scripts/sync_expert_packs.py

sync-domain-language: ## 从 tools/domain-language.json 生成 docs/DOMAIN_LANGUAGE.md
	python3 scripts/sync_domain_language.py

PYTEST_TARGETS ?= tests/
P ?= ./...

test-content: ## Root skill / rules / documentation suite
	python3 -m pytest $(PYTEST_TARGETS) -q

test-engine-package: go-version-check ## Run one engine package, e.g. P=./internal/engine/bazi
	@cd engine && go test -count=1 $(P)

test-engine: go-version-check ## Engine unit and integration suites
	@cd engine && go test -count=1 ./...

test-engine-coverage: go-version-check ## Engine unit coverage floor (shared with CI)
	scripts/test-engine-coverage.sh

test-counsel: ## Counsel MCP integration suite (bootstraps .venv and local engine)
	scripts/test-counsel.sh

test-engine-mcp: ## Engine MCP root/domain protocol smoke
	scripts/test-engine-mcp.sh

eval-mingli160: ## Manual Mingli160 skill calibration/evaluation report (not a gate)
	scripts/eval-mingli160.sh

test: test-content test-engine test-counsel test-engine-mcp ## All deterministic tests
test-python-coverage: ## Run root Python domain suite with coverage
	coverage erase
	coverage run --source=counsel/app -m pytest tests -q
	coverage report --fail-under=40
	coverage xml
	coverage html


test-engine-golden: ## Focused engine golden/oracle/matrix subset
	@cd engine && go test -count=1 -run \
		'Golden|External|Matrix|Pattern' ./internal/engine/...

golden: test-engine-golden ## Backward-compatible alias for the focused golden suite

gate: lint-engine check test ## Full source push gate: lint, static checks, and tests


check-actions: ## Reject mutable reusable workflow/action references
	python3 scripts/check_actions_pinned.py

hooks: ## Install repository git hooks
	git config core.hooksPath .githooks

version: ## Bump all runtime CalVer surfaces together
	@python3 scripts/bump_version.py

python-env: ## Provision the exact CI Python (.python-version) via uv + counsel venv
	uv python install $$(cat .python-version)
	rm -rf counsel/.venv
	uv venv counsel/.venv --python $$(cat .python-version)
	uv pip install --python counsel/.venv/bin/python -r counsel/requirements.lock --require-hashes
	uv pip install --python counsel/.venv/bin/python -r counsel/requirements-dev.txt

clean: ## Remove generated root binaries
	rm -f bin/engine-mcp

help: ## List primary targets
	@grep -h '^[a-z][a-z0-9_.-]*:.*##' $(firstword $(MAKEFILE_LIST)) | \
		awk -F':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

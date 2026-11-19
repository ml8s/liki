PYTHON ?= $(if $(wildcard .venv/bin/python),.venv/bin/python,python3)
IMAGE_REGISTRY ?= liki-local
IMAGE_TAG ?= local
AGENTGO_IMAGE ?=
EXPERT_IMAGE ?= $(IMAGE_REGISTRY)/liki-expert:$(IMAGE_TAG)
NODE_BIN := $(HOME)/.nvm/versions/node/v$(shell cat .nvmrc)/bin
NPM_FROM_PATH := $(shell command -v npm 2>/dev/null)
NPM_FALLBACK := $(NODE_BIN)/npm
NPM ?= $(if $(NPM_FROM_PATH),$(NPM_FROM_PATH),$(NPM_FALLBACK))
NPM_BIN_DIR := $(if $(NPM_FROM_PATH),$(dir $(NPM_FROM_PATH)),$(NODE_BIN))
export PATH := $(NPM_BIN_DIR):$(PATH)

.PHONY: check test gate docs-lint build-expert-deployment build-expert-image test-expert-image workbuddy-package skillhub-package distribution-package distribution-check setup-hooks

check: docs-lint distribution-check ## Static repository, documentation, and package checks
	$(PYTHON) tests/test_repository_boundaries.py

docs-lint: ## Markdown style checks
	$(NPM) run lint:docs

test: ## Skill and distribution contract tests
	$(PYTHON) tests/test_distribution_contracts.py
	$(PYTHON) -m pytest tests -q

build-expert-deployment: ## Generate the AgentDeployment artifact for the Liki expert
	@rm -rf dist/liki-expert && mkdir -p dist
	$(PYTHON) scripts/generate_deployment.py --out dist

build-expert-image: build-expert-deployment ## Assemble the Liki expert image from an AgentGo base image
	@test -n "$(AGENTGO_IMAGE)" || { echo "AGENTGO_IMAGE is required" >&2; exit 1; }
	@rm -rf assembly/dist && mkdir -p assembly/dist/deployment
	@cp -r dist/liki-expert assembly/dist/deployment/liki-expert
	@cp -r skills assembly/dist/skills
	docker build -f assembly/Dockerfile \
		--build-arg BASE_IMAGE="$(AGENTGO_IMAGE)" \
		--build-arg COMMIT="$$(git rev-parse HEAD)" \
		--build-arg BUILD_TIME="$$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
		-t "$(EXPERT_IMAGE)" assembly
	@rm -rf assembly/dist

test-expert-image: build-expert-image ## Validate the expert deployment and smoke the assembled image
	$(PYTHON) scripts/test_expert_image.py --image "$(EXPERT_IMAGE)"

workbuddy-package: ## Build free and pro WorkBuddy packages
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant free
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant pro

skillhub-package: ## Build free and pro SkillHub packages
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant free
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant pro

distribution-package: workbuddy-package skillhub-package ## Build all channel packages

distribution-check: distribution-package ## Verify every channel package deterministically
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant free --check
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant pro --check
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant free --check
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant pro --check

gate: check test build-expert-deployment ## Full source gate: boundaries, docs, packages, and expert deployment

setup-hooks: ## Configure Git to run the full gate before push
	git config core.hooksPath .githooks
	test -x .githooks/pre-push
	@echo "✓ Git hooks installed; pre-push now runs make gate"

PYTHON ?= /usr/bin/python3
NODE_BIN := $(HOME)/.nvm/versions/node/v$(shell cat .nvmrc)/bin
NPM_FROM_PATH := $(shell command -v npm 2>/dev/null)
NPM_FALLBACK := $(NODE_BIN)/npm
NPM ?= $(if $(NPM_FROM_PATH),$(NPM_FROM_PATH),$(NPM_FALLBACK))
NPM_BIN_DIR := $(if $(NPM_FROM_PATH),$(dir $(NPM_FROM_PATH)),$(NODE_BIN))
export PATH := $(NPM_BIN_DIR):$(PATH)

.PHONY: check test gate docs-lint workbuddy-package skillhub-package distribution-package distribution-check

check: docs-lint distribution-check ## Static repository, documentation, and package checks
	$(PYTHON) tests/test_repository_boundaries.py

docs-lint: ## Markdown style checks
	$(NPM) run lint:docs

test: ## Skill and distribution contract tests
	$(PYTHON) tests/test_distribution_contracts.py
	$(PYTHON) -m pytest tests -q

workbuddy-package: ## Build free and pro WorkBuddy packages
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant free
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant pro

skillhub-package: ## Build free and pro SkillHub packages
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant free
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant pro

distribution-package: workbuddy-package skillhub-package ## Build all channel packages

distribution-check: distribution-package ## Verify every channel package deterministically
	$(PYTHON) scripts/check_expert_docs.py
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant free --check
	$(PYTHON) scripts/build_channel_package.py --channel workbuddy --variant pro --check
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant free --check
	$(PYTHON) scripts/build_channel_package.py --channel skillhub --variant pro --check

gate: check test ## Full source gate: boundaries, docs, and distribution contracts

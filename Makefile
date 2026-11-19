PYTHON ?= python3
NODE_BIN := $(HOME)/.nvm/versions/node/v$(shell cat .nvmrc)/bin
NPM_FROM_PATH := $(shell command -v npm 2>/dev/null)
NPM_FALLBACK := $(NODE_BIN)/npm
NPM ?= $(if $(NPM_FROM_PATH),$(NPM_FROM_PATH),$(NPM_FALLBACK))
NPM_BIN_DIR := $(if $(NPM_FROM_PATH),$(dir $(NPM_FROM_PATH)),$(NODE_BIN))
export PATH := $(NPM_BIN_DIR):$(PATH)

.PHONY: check test gate docs-lint

check: docs-lint ## Static repository and documentation checks
	$(PYTHON) tests/test_repository_boundaries.py

docs-lint: ## Markdown style checks
	$(NPM) run lint:docs

test: ## Skill distribution contract tests
	$(PYTHON) tests/test_distribution_contracts.py

gate: check test ## Full source gate: boundaries, docs, and distribution contracts

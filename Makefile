PYTHON ?= python3

.PHONY: check test gate

check: ## Static repository boundary checks
	$(PYTHON) tests/test_repository_boundaries.py

test: ## Skill distribution contract tests
	$(PYTHON) tests/test_distribution_contracts.py

gate: check test ## Full source gate: boundaries and distribution contracts

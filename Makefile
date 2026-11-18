PYTHON ?= python3

.PHONY: check

check: ## Validate the skill distribution contract
	$(PYTHON) tests/test_distribution_contracts.py
	$(PYTHON) tests/test_repository_boundaries.py

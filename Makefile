# liki-skills — 客户端（skill 分发 + WorkBuddy 材料）

.DEFAULT_GOAL := help

.PHONY: help build test

build: ## 打包 skill archive（分发包变薄：MCP 指令 + Aipay）
	scripts/build-archive.sh

test: ## skill 结构契约测试（无需引擎）
	python3 -m pytest tests/test_unified_skill_structure.py tests/test_ready_to_use_contracts.py tests/test_root_docs_contract.py

help: ## 列出 target
	@echo "make build   打包 skill archive"
	@echo "make test    结构契约测试"
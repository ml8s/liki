# liki monorepo — skill 客户端 + 服务端（engine + analysis）
# 同库：skill（MCP 指令）+ engine（Go 排盘）+ analysis（Python 判断层 MCP）

.DEFAULT_GOAL := help

export PATH := $(HOME)/go/bin:$(HOME)/app/go/bin:$(PATH)
export GOCACHE ?= /tmp/gocache

.PHONY: help build build-skill build-engine build-mcp test test-engine test-analysis test-contracts verify

build-skill: ## 打包 skill archive（分发包变薄：MCP 指令 + Aipay）
	scripts/build-archive.sh

build-engine: ## 编译引擎
	@cd engine && go build -o ../bin/liki-engine ./cmd/liki/

build-mcp: ## 编译引擎 MCP server
	@cd engine && go build -o ../bin/liki-mcp ./cmd/liki-mcp/

build: build-skill build-engine ## 构建全部

test-engine: ## 引擎 Go 测试
	@cd engine && go test -count=1 ./...

test-analysis: ## analysis 判断层测试（需本地引擎）
	@cd analysis && .venv/bin/python -m pytest tests/

test-contracts: ## 引擎契约测试（qimen/qiming/divination/engine-snapshot）
	@python3 -m pytest tests/test_qimen_tools.py tests/test_qiming_data.py tests/test_divination_architecture.py tests/test_divination_contracts.py tests/test_divination_snapshot_ask.py tests/test_engine_snapshot_sync.py

test: ## skill 结构契约测试 + 引擎契约
	python3 -m pytest tests/test_unified_skill_structure.py tests/test_ready_to_use_contracts.py tests/test_root_docs_contract.py
	@$(MAKE) --no-print-directory test-contracts

verify: ## 起本地引擎 + analysis 测试
	@bash -c 'cd engine && go build -o /tmp/liki-engine ./cmd/liki/ && trap "fuser -k 18082/tcp 2>/dev/null || true" EXIT; setsid /tmp/liki-engine -addr 127.0.0.1:18082 >/tmp/liki-engine.log 2>&1 < /dev/null & sleep 1; cd ../analysis && LIKI_RPC_URL=http://127.0.0.1:18082/jsonrpc .venv/bin/python -m pytest tests/'

help: ## 列出 target
	@echo "make build-skill  打包 skill archive（MCP 指令 + Aipay）"
	@echo "make build-engine  编译引擎"
	@echo "make build-mcp     编译引擎 MCP server"
	@echo "make test-engine   引擎 Go 测试"
	@echo "make test-analysis analysis 测试（需引擎）"
	@echo "make test-contracts 引擎契约测试"
	@echo "make verify        起本地引擎 + analysis 测试"
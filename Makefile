# liki monorepo — skill 客户端 + 服务端（engine + analysis）
# 同库：skill（MCP 指令）+ engine（Go 排盘）+ analysis（Python 判断层 MCP）
# 质量门禁：make check（静态）→ make test（全部测试）→ make gate（check+test，pre-push）

.DEFAULT_GOAL := help

export PATH := $(HOME)/go/bin:$(HOME)/app/go/bin:$(PATH)
export GOCACHE ?= /tmp/gocache
ENGINE_MCP_PORT ?= 18081

.PHONY: help build build-skill build-engine build-mcp \
        check test test-contracts test-engine test-analysis gate

build-skill: ## 打包 skill archive（MCP 指令 + Aipay）
	scripts/build-archive.sh

build-engine: ## 编译引擎
	@cd engine && go build -o ../bin/liki-engine ./cmd/liki/

build-mcp: ## 编译引擎 MCP server
	@cd engine && go build -o ../bin/liki-mcp ./cmd/liki-mcp/

build: build-skill build-engine ## 构建全部

# ── Check：所有静态检查（格式 / 风格 / 结构 / 契约 / 数据质量）───────────────

check: ## 静态检查：markdownlint + skill md 结构 + go vet + 断言 schema + 文档契约
	npx markdownlint-cli2 "skills/**/*.md"
	python3 -m pytest tests/test_skill_markdown_structure.py tests/test_skill_frontmatter.py -q
	@cd engine && go vet ./...
	python3 scripts/check_schema.py
	python3 scripts/check_docs.py

# ── Test：起本地引擎 + 全部测试（skill + engine + analysis 集成）────────────

test-contracts: ## skill 结构契约测试（tests/ 全量）
	python3 -m pytest tests/ -q

test-engine: ## 引擎 Go 测试
	@cd engine && go test -count=1 ./...

test-analysis: ## 起本地引擎 + analysis 测试
	@bash -c 'cd engine && go build -o /tmp/liki-mcp ./cmd/liki-mcp/ && fuser -k $(ENGINE_MCP_PORT)/tcp 2>/dev/null || true; setsid /tmp/liki-mcp -addr 127.0.0.1:$(ENGINE_MCP_PORT) >/tmp/liki-mcp.log 2>&1 < /dev/null & for i in 1 2 3 4 5 6 7 8 9 10; do curl -fsS -m 2 http://127.0.0.1:$(ENGINE_MCP_PORT)/health >/dev/null 2>&1 && break; sleep 1; done; cd ../analysis && LIKI_MCP_URL=http://127.0.0.1:$(ENGINE_MCP_PORT)/mcp .venv/bin/python -m pytest tests/ -q'

test: test-contracts test-engine test-analysis ## 全部测试（skill + engine + analysis）

# ── Gate：check + test（全面测试，pre-push 门槛）───────────────────────────

gate: check test ## 全面测试：check + test（推送前门槛）

help: ## 列出 target
	@echo "make check          静态检查（markdownlint + go vet + schema + docs）"
	@echo "make test-contracts skill 结构契约测试（tests/ 全量）"
	@echo "make test-engine     引擎 Go 测试"
	@echo "make test-analysis   analysis 测试（需引擎 18081）"
	@echo "make test            起本地引擎 + 全部测试（skill + engine + analysis）"
	@echo "make gate            全面测试（check + test，pre-push）"
	@echo "make build-skill     打包 skill archive"
	@echo "make build-engine    编译引擎"
	@echo "make build-mcp       编译引擎 MCP server"
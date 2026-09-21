# liki monorepo Makefile — skills（Python）+ engine（Go）双栈
#
# 只需要记住 2 个命令：
#   make check          # 所有静态检查（格式 + lint + schema + docs）
#   make test           # 所有测试（pytest + Go 引擎全量）
#   make gate           # CI 静态检查 + test + verify → 通过 = 可以推送
#
# 本地 check 自动修复格式；CI 用 `make ci` 只验证不修改。

.DEFAULT_GOAL := help

# ── 工具路径 ──
VERSION_FILES := skills/liki/VERSION.txt engine/cmd/liki/VERSION
VERSION_CONTRACTS := skills/liki/natal/tools/natal_projection_contract.json skills/liki/divination/tools/qimen_projection_contract.json
VERSION_MANIFESTS := skills/liki/natal/tools/skill-tools.json skills/liki/divination/tools/skill-tools.json
MDL := $(shell ls $(HOME)/.npm/_npx/*/node_modules/.bin/markdownlint-cli2 2>/dev/null | tail -1)

export PATH := $(HOME)/go/bin:$(HOME)/app/go/bin:$(PATH)
export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache

# ════════════════════════════════════════════════════════════
#  开发者入口
# ════════════════════════════════════════════════════════════

.PHONY: help check test gate verify ci fmt lint fmt-md lint-md lint-go \
	check-data test-go build build-skill build-engine golden benchmark \
	version hooks _mdl _golden-all

check: fmt-md lint-md lint-go check-data ## 所有静态检查（格式 + lint + schema + docs）
	@echo ""
	@echo "✓ check 通过"

test: ## 所有测试（pytest + Go 引擎全量 unit + integration + RPC 冒烟）
	python3 -m pytest tests/ -q --ignore=tests/test_integration.py
	cd engine && scripts/ci-engine.sh

gate: ci test verify ## 推送前门槛（静态验证 + test + verify；不修改文件）
	@echo ""
	@echo "✓ gate 通过（静态验证 + test + verify）"

ci: lint-md lint-go check-data ## CI 静态检查（不修复格式）
	@echo ""
	@echo "✓ ci 通过"

verify: ## 端到端：起本地引擎 → integration + 160 题全量数据检查
	@bash -c '. scripts/local-engine.sh; ensure_local_engine; trap stop_local_engine EXIT; LIKI_RPC_URL="$$LOCAL_RPC" python3 -m pytest tests/test_integration.py -q'
	@bash -c '. scripts/local-engine.sh; ensure_local_engine; trap stop_local_engine EXIT; LIKI_RPC_URL="$$LOCAL_RPC" python3 scripts/eval_hybrid.py'

# ════════════════════════════════════════════════════════════
#  静态检查原语
# ════════════════════════════════════════════════════════════

fmt-md: ## 修复 markdown 格式（--fix）
	@$(MAKE) --no-print-directory _mdl ARGS="--fix"

lint-md: ## 检查 markdown 格式（不修改文件）
	@$(MAKE) --no-print-directory _mdl

lint-go: ## Go 静态检查（golangci-lint + go vet）
	cd engine && golangci-lint run ./... && go vet ./...

check-data: ## 断语表质量 + 文档引用（check_schema + check_docs）
	@bash tests/check.sh

fmt: fmt-md ## 修复全部格式

lint: lint-md lint-go ## 全部静态检查（不修改）

# 内部：markdownlint 分发
_mdl:
	@set -e; \
	if [ -n "$(MDL)" ]; then \
		PATH="$(HOME)/.nvm/versions/node/v22.19.0/bin:$(PATH)" $(MDL) $(ARGS); \
	else \
		npx --yes markdownlint-cli2@0.17.2 $(ARGS); \
	fi

# ════════════════════════════════════════════════════════════
#  构建
# ════════════════════════════════════════════════════════════

build-skill: ## 打包 skill archive
	scripts/build-archive.sh

build-engine: ## 编译引擎二进制
	@cd engine && go build -o ../bin/liki-engine ./cmd/liki/

build: build-skill build-engine ## 构建全部

# ════════════════════════════════════════════════════════════
#  专项
# ════════════════════════════════════════════════════════════

golden: ## golden 全量
	@$(MAKE) --no-print-directory _golden-all

benchmark: ## 160 题命理准确率评测
	bash tests/benchmark/mingli160/run.sh

hooks: ## 安装 git hooks
	git config core.hooksPath .githooks

version: ## 写入今日 CalVer
	@BASE=$$(TZ=Asia/Shanghai date +%Y.%m.%d); SERIAL=0; FOUND=0; \
	for F in $(VERSION_FILES); do \
		V=$$(cat "$$F"); \
		case "$$V" in \
			"$$BASE".*) S=$${V##*.}; [ "$$S" -gt "$$SERIAL" ] && SERIAL=$$S; FOUND=1 ;; \
		esac; \
	done; \
	if [ "$$FOUND" -eq 1 ]; then SERIAL=$$((SERIAL + 1)); else SERIAL=0; fi; \
	VERSION="$$BASE.$$SERIAL"; \
	for F in $(VERSION_FILES); do echo "$$VERSION" > "$$F"; done; \
	for F in $(VERSION_CONTRACTS); do \
		sed -i 's/"version": "[^"]*"/"version": "'"$$VERSION"'"/' "$$F"; \
	done; \
	for F in $(VERSION_MANIFESTS); do \
		python3 -c 'import json,sys; p=sys.argv[1]; d=json.load(open(p,encoding="utf-8")); d.setdefault("info",{})["version"]=sys.argv[2]; json.dump(d,open(p,"w",encoding="utf-8"),ensure_ascii=False,indent=2); open(p,"a",encoding="utf-8").write("\n")' "$$F" "$$VERSION"; \
	done; \
	echo "✅ 版本 → $$VERSION"

# ════════════════════════════════════════════════════════════
#  help
# ════════════════════════════════════════════════════════════

help: ## 列出所有 target
	@echo ""
	@echo "\033[1m静态检查:\033[0m"
	@echo "  make check          所有静态检查（格式 + lint + schema + docs）"
	@echo "  make ci             CI 静态检查（不修复格式）"
	@echo ""
	@echo "\033[1m测试:\033[0m"
	@echo "  make test           所有测试（pytest + Go 引擎全量）"
	@echo "  make verify         端到端（需本地引擎）"
	@echo ""
	@echo "\033[1m推送:\033[0m"
	@echo "  make gate           推送前门槛（静态验证 + test + verify）"
	@echo ""
	@echo "\033[1m构建:\033[0m"
	@echo "  make build          skill archive + engine binary"
	@echo ""
	@echo "\033[1m格式:\033[0m"
	@echo "  make fmt            修复全部格式"
	@echo "  make lint           检查全部格式（不修改）"
	@echo ""
	@echo "\033[1m专项:\033[0m"
	@echo "  make golden         golden 全量"
	@echo "  make benchmark      160 题评测"
	@echo ""
	@echo "\033[1m其他:\033[0m"
	@echo "  make version        bump CalVer"
	@echo "  make hooks          安装 git hooks"
	@echo ""
	@grep -h '^[a-z].*##' $(firstword $(MAKEFILE_LIST)) | grep -v '^#' | awk -F':.*## ' '{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

# ════════════════════════════════════════════════════════════
#  内部实现
# ════════════════════════════════════════════════════════════

test-go: ## Go 单元测试（race + short）
	cd engine && go test -race -count=1 -short ./...

_golden-all:
	@cd engine && go test -count=1 -run 'Golden|External|Matrix|Pattern' ./internal/engine/qimen ./internal/engine/ziwei
_golden-bazi:
	@cd engine && go test -count=1 -run 'TestBaziMultiOracleGolden_AllTermBoundaries' ./internal/engine/bazi
_golden-tianwen:
	@cd engine && go test -count=1 -run 'TestTianwenMultiOracleGolden_LunarMonthBoundaries' ./internal/engine/tianwen
_golden-liuyao:
	@cd engine && go test -count=1 -run 'TestCore64Golden_JingFangEightPalaces' ./internal/engine/liuyao
_golden-huangli:
	@cd engine && go test -count=1 -run 'TestQueryDateMatrixGolden_TwoOracleCalendarAndEventRules' ./internal/engine/huangli
_golden-bazhai:
	@cd engine && go test -count=1 -run 'TestMingGuaCycleGolden_AllYearsAndGenders' ./internal/engine/bazhai
_golden-xuankong:
	@cd engine && go test -count=1 -run 'TestFlyingStarMatrixGolden_AllYunAndOppositePairs' ./internal/engine/xuankong

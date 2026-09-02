# liki monorepo Makefile — skills（Python）+ engine（Go）双栈
#
# 版本策略：根 Makefile 统一写入 4 个 skill VERSION 与 engine VERSION

# ── 统一版本（skill 4 份 + engine 1 份，同步 bump）──
VERSION_FILES := skills/liki-bazi/VERSION skills/liki-divination/VERSION skills/liki-fengshui/VERSION skills/liki-naming/VERSION engine/cmd/liki/VERSION

version: ## 写入今日日期（CalVer）
	@BASE=$$(date +%Y.%m.%d); SERIAL=0; FOUND=0; \
	for F in $(VERSION_FILES); do \
		V=$$(cat "$$F"); \
		case "$$V" in \
			"$$BASE".*) S=$${V##*.}; [ "$$S" -gt "$$SERIAL" ] && SERIAL=$$S; FOUND=1 ;; \
		esac; \
	done; \
	if [ "$$FOUND" -eq 1 ]; then SERIAL=$$((SERIAL + 1)); else SERIAL=0; fi; \
	VERSION="$$BASE.$$SERIAL"; \
	for F in $(VERSION_FILES); do echo "$$VERSION" > "$$F"; done; \
	echo "✅ 版本 → $$VERSION"

# ── 构建 ──
build-archive: ## 打包 skill → dist/*.tar.gz + index.json
	scripts/build-archive.sh

build: build-archive ## 构建全部（当前 = skills archive + engine 二进制）
	@cd engine && go build -o ../bin/liki-engine ./cmd/liki/

# ── 验证 / 测试 ──
hooks: ## 安装 git hooks（贡献者克隆后执行一次；core.hooksPath 是本地配置不会随 clone 带上）
	git config core.hooksPath .githooks

check: ## Skills 改表后验证（schema 校验 + 数据检查）
	@bash tests/check.sh

test: ## Skills python 单测（规则引擎；integration 由服务已起阶段跑）
	python3 -m pytest tests/ -q --ignore=tests/test_integration.py

test-integration: ## Skill 全链路集成测试（本地起引擎 + LIKI_RPC_URL 连它；脱离生产）
	@bash -c '. scripts/local-engine.sh; ensure_local_engine; trap stop_local_engine EXIT; LIKI_RPC_URL="$$LOCAL_RPC" python3 -m pytest tests/test_integration.py -q'

# ── Engine 测试（全部在 engine/ 子目录，自含）──
test-engine: ## Engine 全量测试（lint + vet + unit race + integration + RPC 冒烟）
	cd engine && scripts/ci-engine.sh

test-all: test test-engine test-integration ## 全量（单项目：skills 单测 + engine 全量 + skill 全链路集成）

# 推送前 PATH 补充（golangci-lint / go）
export PATH := $(HOME)/go/bin:$(HOME)/app/go/bin:$(PATH)

pre-push: ## 推送前门槛测试（与 CI 对齐——绿了再推，~30s）
	@echo "=== [1/6] check_docs（文档契约 × 4 skill）==="
	@for s in liki-bazi liki-divination liki-fengshui liki-naming; do \
		python3 tests/check_docs.py "skills/$$s" || exit 1; \
	done
	@echo "=== [2/6] Python 单测 ==="
	python3 -m pytest tests/ --ignore=tests/test_integration.py -q --tb=short || exit 1
	@echo "=== [3/6] eval_hybrid 冒烟（前 3 题验证管线通）==="
	python3 -c "import tests.eval_hybrid" || exit 1
	@echo "=== [4/6] Go build + vet ==="
	cd engine && go build ./... && go vet ./... || exit 1
	@echo "=== [5/6] golangci-lint ==="
	cd engine && golangci-lint run ./... || exit 1
	@echo "=== [6/6] Go 单测（-short）==="
	cd engine && go test -short -count=1 ./... || exit 1
	@echo ""
	@echo "✓ 推送前门槛检查全部通过（CI 同集，绿了再推）"

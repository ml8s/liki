# liki monorepo Makefile — skills（Python）+ engine（Go）双栈
#
# 版本策略：统一单一时版号（5.0.0），skill 与 engine 同步
#   - 单一事实源：根 Makefile 统一 bump，同时写入 4 skill VERSION 与 engine VERSION
#   - 历史上 skill 走 4.x、engine 走 2.6.x；本次 big release 合并后统一为同一版本

# ── 统一版本（skill 4 份 + engine 1 份，同步 bump）──
VERSION_FILES := skills/liki-bazi/VERSION skills/liki-divination/VERSION skills/liki-fengshui/VERSION skills/liki-naming/VERSION engine/cmd/liki/VERSION

version: ## 写入今日日期（CalVer）
	@DATE=$$(date +%Y.%m.%d); for F in $(VERSION_FILES); do echo "$$DATE" > $$F; done; \
	echo "✅ 版本 → $$DATE（tag 按需：git tag -a $$DATE）"

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
test-engine: ## Engine 全量测试（lint + vet + unit race + integration + RPC 冒烟 74/74）
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

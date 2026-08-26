# liki monorepo Makefile — skills（Python）+ engine（Go）双栈
#
# 版本策略：统一单一时版号（5.0.0），skill 与 engine 同步
#   - 单一事实源：根 Makefile 统一 bump，同时写入 4 skill VERSION 与 engine VERSION
#   - 历史上 skill 走 4.x、engine 走 2.6.x；本次 big release 合并后统一为同一版本

# ── 统一版本（skill 4 份 + engine 1 份，同步 bump）──
VERSION_FILES := skills/liki-bazi/VERSION skills/liki-divination/VERSION skills/liki-fengshui/VERSION skills/liki-naming/VERSION engine/cmd/liki/VERSION

version-patch: ## Bump PATCH（统一，如 5.0.0 → 5.0.1）
	@for F in $(VERSION_FILES); do \
		V=$$(cat $$F); MAJOR=$${V%%.*}; REST=$${V#*.}; MINOR=$${REST%.*}; PATCH=$${REST#*.}; \
		echo "$$MAJOR.$$MINOR.$$((PATCH + 1))" > $$F; \
	done; \
	echo "✅ 统一版本 → $$(cat skills/liki-bazi/VERSION)"

version-minor: ## Bump MINOR（统一，如 5.0.0 → 5.1.0）
	@for F in $(VERSION_FILES); do \
		V=$$(cat $$F); MAJOR=$${V%%.*}; REST=$${V#*.}; MINOR=$${REST%.*}; \
		echo "$$MAJOR.$$((MINOR + 1)).0" > $$F; \
	done; \
	echo "✅ 统一版本 → $$(cat skills/liki-bazi/VERSION)"

version-major: ## Bump MAJOR（统一，如 5.0.0 → 6.0.0）
	@for F in $(VERSION_FILES); do \
		V=$$(cat $$F); MAJOR=$${V%%.*}; \
		echo "$$((MAJOR + 1)).0.0" > $$F; \
	done; \
	echo "✅ 统一版本 → $$(cat skills/liki-bazi/VERSION)（skill + engine 同步）"

# ── 构建 ──
build-archive: ## 打包 skill → dist/*.tar.gz + index.json
	scripts/build-archive.sh

build: build-archive ## 构建全部（当前 = skills archive + engine 二进制）
	@cd engine && go build -o ../bin/liki-engine ./cmd/liki/

# ── 验证 / 测试 ──
check: ## Skills 改表后验证（schema 校验 + 数据检查）
	@bash tests/check.sh

test: ## Skills python 单测（规则引擎；integration 由服务已起阶段跑）
	python3 -m pytest tests/ -q --ignore=tests/test_integration.py

test-integration: ## Skills 全链路集成测试（需引擎服务：LIKI_RPC_URL 指向引擎 /jsonrpc）
	python3 -m pytest tests/test_integration.py -q

# ── Engine 测试（全部在 engine/ 子目录，自含）──
test-engine: ## Engine 全量测试（lint + vet + unit race + integration + RPC 冒烟 74/74）
	cd engine && scripts/ci-engine.sh

test-all: test test-engine ## 全量（skills + engine）

pre-push: build-archive ## 推送前检查（重算指纹 + 校验）
	@for s in liki-bazi liki-divination liki-fengshui liki-naming; do \
		python3 tests/check_docs.py "skills/$s" || exit 1; \
	done
	@echo "✓ 推送前检查通过"

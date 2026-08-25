# liki monorepo Makefile — skills（Python）+ engine（Go）双栈
#
# 版本策略：两套子版本各自独立 bump
#   - skills（4 子 skill 同步定版）→ VERSION_FILES
#   - engine（Go 引擎）             → engine/cmd/liki/VERSION

# ── Skills 版本（四 skill 同步）──
VERSION_FILES := skills/liki-bazi/VERSION skills/liki-divination/VERSION skills/liki-fengshui/VERSION skills/liki-naming/VERSION

version-patch: ## Bump PATCH（工程级：四 skill 同步，如 4.0.0 → 4.0.1）
	@for F in $(VERSION_FILES); do \
		V=$$(cat $$F); MAJOR=$${V%%.*}; REST=$${V#*.}; MINOR=$${REST%.*}; PATCH=$${REST#*.}; \
		echo "$$MAJOR.$$MINOR.$$((PATCH + 1))" > $$F; \
	done; \
	echo "✅ 四 skill 版本 → $$(cat skills/liki-bazi/VERSION)"

version-minor: ## Bump MINOR（工程级：四 skill 同步，如 4.0.0 → 4.1.0）
	@for F in $(VERSION_FILES); do \
		V=$$(cat $$F); MAJOR=$${V%%.*}; REST=$${V#*.}; MINOR=$${REST%.*}; \
		echo "$$MAJOR.$$((MINOR + 1)).0" > $$F; \
	done; \
	echo "✅ 四 skill 版本 → $$(cat skills/liki-bazi/VERSION)"

version-major: ## Bump MAJOR（工程级：四 skill 同步，如 4.0.0 → 5.0.0）
	@for F in $(VERSION_FILES); do \
		V=$$(cat $$F); MAJOR=$${V%%.*}; \
		echo "$$((MAJOR + 1)).0.0" > $$F; \
	done; \
	echo "✅ 四 skill 版本 → $$(cat skills/liki-bazi/VERSION)"

# ── Engine 版本（独立，engine/cmd/liki/VERSION）──
version-engine-patch: ## Bump engine PATCH（如 2.6.0 → 2.6.1）
	@cd engine && make version-patch

version-engine-minor: ## Bump engine MINOR（如 2.6.0 → 2.7.0）
	@cd engine && make version-minor

version-engine-major: ## Bump engine MAJOR（如 2.6.0 → 3.0.0）
	@cd engine && make version-major

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

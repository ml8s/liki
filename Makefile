# 版本以工程为单位（四 skill 同步定版，子 skill 不独立定版本号）
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

build-archive: ## 打包 skill → dist/liki.tar.gz + index.json
	scripts/build-archive.sh

check: ## 改表后验证（schema 校验 + 数据检查）
	@bash tests/check.sh

# ── 测试 ──
test: ## 运行 python 单测（规则引擎：因子/断语/agent_cli 分派；integration 由服务已起阶段跑）
	python3 -m pytest tests/ -q --ignore=tests/test_integration.py

test-integration: ## 全链路集成测试（需引擎服务：LIKI_RPC_URL 指向引擎 /jsonrpc）
	python3 -m pytest tests/test_integration.py -q

test-all: test ## 全量（当前 = test；后续加 lint/archive 校验）

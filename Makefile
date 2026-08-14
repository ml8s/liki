VERSION_FILE := VERSION

version-patch: ## Bump PATCH (1.26.0 → 1.26.1)
	@V=$$(cat $(VERSION_FILE)); \
	MAJOR=$${V%%.*}; \
	REST=$${V#*.}; \
	MINOR=$${REST%.*}; \
	PATCH=$${REST#*.}; \
	echo "$$MAJOR.$$MINOR.$$((PATCH + 1))" > $(VERSION_FILE); \
	echo "✅ $$V → $$(cat $(VERSION_FILE))"

version-minor: ## Bump MINOR (1.26.0 → 1.27.0)
	@V=$$(cat $(VERSION_FILE)); \
	MAJOR=$${V%%.*}; \
	REST=$${V#*.}; \
	MINOR=$${REST%.*}; \
	echo "$$MAJOR.$$((MINOR + 1)).0" > $(VERSION_FILE); \
	echo "✅ $$V → $$(cat $(VERSION_FILE))"

version-major: ## Bump MAJOR (1.26.0 → 2.0.0)
	@V=$$(cat $(VERSION_FILE)); \
	MAJOR=$${V%%.*}; \
	echo "$$((MAJOR + 1)).0.0" > $(VERSION_FILE); \
	echo "✅ $$V → $$(cat $(VERSION_FILE))"

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

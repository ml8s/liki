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

smoke-test: ## API + 文件完整性测试
	@tests/smoke.sh

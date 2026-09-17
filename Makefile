# liki monorepo Makefile — skills（Python）+ engine（Go）双栈
#
# 版本策略：根 Makefile 统一写入 unified skill VERSION.txt 与 engine VERSION

# ── 统一版本（skill 1 份 + engine 1 份，同步 bump）──
VERSION_FILES := skills/liki/VERSION.txt engine/cmd/liki/VERSION
VERSION_CONTRACTS := skills/liki/bazi/tools/natal_projection_contract.json skills/liki/divination/tools/qimen_projection_contract.json
VERSION_MANIFESTS := skills/liki/bazi/tools/skill-tools.json skills/liki/divination/tools/skill-tools.json

version: ## 写入今日日期（CalVer）
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

test-functional: ## 规则引擎功能测试（因子/断语/场景/冲突）
	python3 -m pytest tests/test_functional.py -q

golden-bazi: ## 八字多源共识 golden（只读 checked-in fixture）
	@cd engine && go test -count=1 -run 'TestBaziMultiOracleGolden_AllTermBoundaries' ./internal/engine/bazi

golden-bazi-generate: ## 重新生成八字多源 golden（开发机需 lunar-python/sxtwl/bazi-calculator）
	@BAZI_CALCULATOR_DIR="$${BAZI_CALCULATOR_DIR:-/tmp/bazi-calculator}" \
		PYTHONPATH="$${PYTHONPATH:-}" python3 tests/golden/bazi/generate.py \
		--output engine/internal/engine/bazi/testdata/multi_oracle_golden.json

golden-tianwen: ## 天文历法两源共识 golden（只读 checked-in fixture）
	@cd engine && go test -count=1 -run 'TestTianwenMultiOracleGolden_LunarMonthBoundaries' ./internal/engine/tianwen

golden-tianwen-generate: ## 重新生成天文历法 golden（开发机需 lunar-python/sxtwl）
	@PYTHONPATH="$${PYTHONPATH:-}" python3 tests/golden/tianwen/generate.py

golden-liuyao: ## 六爻 64 卦八宫 golden
	@cd engine && go test -count=1 -run 'TestCore64Golden_JingFangEightPalaces' ./internal/engine/liuyao

golden-liuyao-generate: ## 重新生成六爻 64 卦领域 golden
	@python3 tests/golden/liuyao/generate.py

golden-huangli: ## 黄历两源日柱 + 建除黄黑道矩阵 golden
	@cd engine && go test -count=1 -run 'TestQueryDateMatrixGolden_TwoOracleCalendarAndEventRules' ./internal/engine/huangli

golden-huangli-generate: ## 重新生成黄历矩阵 golden（开发机需 lunar-python/sxtwl）
	@PYTHONPATH="$${PYTHONPATH:-}" python3 tests/golden/huangli/generate.py

golden-bazhai: ## 八宅命卦 1900-2099 全周期 golden
	@cd engine && go test -count=1 -run 'TestMingGuaCycleGolden_AllYearsAndGenders' ./internal/engine/bazhai

golden-bazhai-generate: ## 重新生成八宅命卦周期 golden
	@python3 tests/golden/bazhai/generate.py

golden-xuankong: ## 玄空九运十二山向飞星矩阵 golden
	@cd engine && go test -count=1 -run 'TestFlyingStarMatrixGolden_AllYunAndOppositePairs' ./internal/engine/xuankong

golden-xuankong-generate: ## 重新生成玄空飞星矩阵 golden
	@python3 tests/golden/xuankong/generate.py

golden-engine: ## 全部 engine 领域 golden（含八字/紫微既有锚点与奇门外部 golden）
	@make golden-bazi golden-tianwen golden-liuyao golden-huangli golden-bazhai golden-xuankong
	@cd engine && go test -count=1 -run 'Golden|External|Matrix|Pattern' ./internal/engine/qimen ./internal/engine/ziwei

test-integration: ## Skill 全链路集成测试（本地起引擎 + LIKI_RPC_URL 连它；脱离生产）
	@bash -c '. scripts/local-engine.sh; ensure_local_engine; trap stop_local_engine EXIT; LIKI_RPC_URL="$$LOCAL_RPC" python3 -m pytest tests/test_integration.py -q'

benchmark-mingli160: ## 160 题命理准确率基准（模型 + skill-up；不进入 pre-push）
	bash tests/benchmark/mingli160/run.sh

skillup-smoke-validate: ## 校验跨领域 skill-up 功能 smoke 配置（不调模型）
	bash tests/skillup/run.sh --validate

skillup-smoke: ## 跨领域 skill-up 功能 smoke（模型 + 本地 engine；不进入 pre-push）
	bash tests/skillup/run.sh

skillup-smoke-bazi: ## 八字功能 smoke
	bash tests/skillup/run.sh bazi

skillup-smoke-divination: ## 问卦功能 smoke（六爻 + 奇门）
	bash tests/skillup/run.sh divination

skillup-smoke-fengshui: ## 风水功能 smoke（八宅 + 玄空）
	bash tests/skillup/run.sh fengshui

skillup-smoke-naming: ## 起名功能 smoke
	bash tests/skillup/run.sh naming

# ── Engine 测试（全部在 engine/ 子目录，自含）──
test-engine: ## Engine 全量测试（lint + vet + unit race + integration + RPC 冒烟）
	cd engine && scripts/ci-engine.sh

test-all: test test-engine test-integration ## 全量（单项目：skills 单测 + engine 全量 + skill 全链路集成）

# 推送前 PATH 补充（golangci-lint / go）
export PATH := $(HOME)/go/bin:$(HOME)/app/go/bin:$(PATH)
export GOCACHE ?= /tmp/gocache
export GOLANGCI_LINT_CACHE ?= /tmp/golangci-lint-cache

pre-push: ## 推送前门槛测试（与 CI 对齐——绿了再推，~2min）
	@echo "=== [1/8] README / user guide lint ==="
	@make --no-print-directory lint-readme || exit 1
	@echo "=== [2/8] check_docs（unified Liki skill）==="
	@python3 tests/check_docs.py skills/liki || exit 1
	@echo "=== [3/8] Python 单测 ==="
	python3 -m pytest tests/ --ignore=tests/test_integration.py -q --tb=short || exit 1
	@echo "=== [4/8] eval_hybrid 冒烟（前 3 题验证管线通）==="
	python3 -c "import tests.eval_hybrid" || exit 1
	@echo "=== [5/8] Go build + vet ==="
	cd engine && go build ./... && go vet ./... || exit 1
	@echo "=== [6/8] golangci-lint ==="
	cd engine && golangci-lint run ./... || exit 1
	@echo "=== [7/8] Go 单测（-short）==="
	cd engine && go test -short -count=1 ./... || exit 1
	@echo "=== [8/8] 流年/断语全量数据检查（对齐 CI full check）==="
	@bash -c '. scripts/local-engine.sh; ensure_local_engine; trap stop_local_engine EXIT; LIKI_RPC_URL="$$LOCAL_RPC" python3 tests/eval_hybrid.py' || exit 1
	@echo ""
	@echo "✓ 推送前门槛检查全部通过（CI 同集，绿了再推）"

lint-readme: ## 检查 README 和用户指南 Markdown 结构
	npx --yes markdownlint-cli2@0.17.2 "README.md" "README.en.md" "docs/USER_GUIDE.md" "docs/USER_GUIDE.en.md"

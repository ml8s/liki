# 贡献指南

> English contributing guide: [CONTRIBUTING.md](./CONTRIBUTING.md)

感谢你考虑为liki.hk贡献！

## 提交 Issue

- 命理计算结果不准确 → 请附上出生时间、地点、期望结果和实际结果
- SKILL 行为异常 → 请说明触发了什么流程、期望什么、实际发生了什么
- 建议新技能 → 请描述用户场景，越具体越好

## 分支模型

`main` 是唯一的长期分支；不要使用 `dev`、`master` 或长期发布分支。维护者在 `main`
上做小而聚焦的提交。外部贡献者可以 fork 后创建短生命周期分支，PR 目标仍然是 `main`。

发布 = 本地按各仓 commit sha 构建镜像 → `liki-deploy` 跑全系统 e2e → 部署并 smoke；CI 不构建镜像。

## 提交变更

1. 同步最新 `main`：`git pull --ff-only`
2. 安装 git hooks（一次）：`make hooks`
3. 有用户可见行为变化时更新 `CHANGELOG.md`（README 统计数字有变时一并更新）
4. README / 用户指南改动需保持现有标题结构和双语一致性，并运行 `make check`
5. 提交时描述清楚改了什么、为什么；不相关的变更保持独立提交

## 代码规范

### 本地工具链

版本只在一处声明，CI 与本地读同一份，`make check` 校验本地一致：

- Go 1.26.11：`go.mod` + `GOTOOLCHAIN`（`make go-version-check`）
- Python 3.12.14：`.python-version`（`make runtime-version-check`）
- Node 24.14.1：`.nvmrc`（`make runtime-version-check`）
- golangci-lint v2.13.2：`deps/golangci-lint.version`（本地与 CI 共用，对账测试锁定）
- ruff 0.16.0：`lint-python.sh` 的 `RUFF_VERSION`

GitHub Action 版本一律 40 位 SHA pin；升级由 Dependabot 每周分组一个 PR 跟进。
合并前看 release notes——大版本常有 breaking（如 docker/login-action v4）。
Python 工具链用 [uv](https://docs.astral.sh/uv/) 固化：`make python-env` 一次性安装 `.python-version` 精确解释器并重建 counsel venv（锁定依赖 + 测试工具，与 CI 同源）。国内网络可 `export UV_DEFAULT_INDEX=https://pypi.tuna.tsinghua.edu.cn/simple` 加速。

### 语言和服务边界

- 根 `skills/liki/SKILL.md` 只做产品路由；每个领域用 `ENTRY.md` 进入。SKILL.md 以中文为主，术语保持原文。
- 领域 `ENTRY.md` 只做域内路由、工具契约与硬边界；`app` 卡平铺为 `references/<域>-app-<目标>.md`（`natal-app-outcome.md`…）并承载流程、交互闸口与输出模板，方法论卡平铺为 `references/<域>-<体系>-<卡>.md`（`natal-bazi-*.md`、`divination-qimen-*.md`…）；一份信息只在一处定义（`tests/test_unified_skill_structure.py` 钉住路由可达与引用形态）。
- **单一聚合专家**：`skills/liki/` 是唯一内容源；`distribution/workbuddy/liki-pro/` 只保存 WorkBuddy 渠道资产。`distribution/liki-expert/deployment.yaml` 与 `distribution/liki-expert/profile.json` 是唯一部署配置。领域路由由 `references/<域>-ENTRY.md` 承载，不需要子代理。
- **卡写作约定（`scripts/check_docs.py` + `tests/test_unified_skill_structure.py` 钉住）**：
  方法论卡（`references/*-app-*.md`、`references/<域>-<体系>-*.md`）不带 frontmatter——卡的身份由平铺命名（域 / app / 体系段）与 `<域>-ENTRY.md` 路由表表达，元数据写在卡内没有消费方；
  `## 输出模板` 必须给模板体（代码块 / 列表 / 表），示例放 `### 示例` 之下；
  一条规则只在一处定义——域级公共口径进该域 `ENTRY.md` / `TOOLS.md`，跨体系合参进 `app/`，被专家包自包含逼出的同源双份（如 natal / naming 的用神优先序）用测试钉成同字；
  卡内只引工具字段名与工具名，不写实现内部数据文件名（`data/*.json`）；
  卡间互引统一平铺全路径 `references/<平铺名>.md`——裸名 / 旧嵌套 / 跨域 / 根级共享卡 / `scripts/` 依赖在生成器 fail-closed。
- 专家身份的唯一手写源是 `skills/liki/SKILL.md`；WorkBuddy 包内入口由唯一身份、`distribution/workbuddy/liki-pro/channel.json` 和付费 overlay 生成，`make check` 漂移门守护且拒绝包内出现第二份身份副本。
- `engine-mcp` 只做确定性排盘/历法/风水计算；`counsel-mcp` 只做因子、断语、起名和问卦判断。
- 公开 `/engine` 与 `/counsel` 前缀由 edge 负责；两个服务内部只实现 `/mcp`。
- 不恢复旧 Python CLI 或 JSON-RPC 发现/调用路径；历史评测资产已归档，迁移前不得进入 gate。

### Lint 与安全扫描

- 引擎 lint 用 golangci-lint v2（配置 `engine/.golangci.yml`）。本地安装用官方二进制脚本，**不要 `go install`**（golangci-lint 与 Go 版本强耦合，官方明确不推荐该方式）：

  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

- Python lint 使用 Ruff 0.16.0：`make lint-python`。
- Go 漏洞扫描使用 `govulncheck`；counsel 哈希锁使用 `pip-audit`。两者在 CI 分别执行。
- 镜像运行时不用发布版本；构建身份是 Git short SHA。需要生成分发包时运行
  `make build-archive` 或 `make distribution-package`。

## 有意保留的重复（不要合并）

评审时下面这些看起来可以合并或去重，但每一项都是刻意的；动手之前先确认理由仍然成立。

| 重复 | 为什么保留 |
| --- | --- |
| `engine/**/zz_load_test.go`（12 个） | Go 不能跨包共享 `TestMain`；抽成公共包会把 `testing` 带进非测试构建依赖。9 个调自身 `Load()`，3 个调依赖包的 `Load()`（fengshui→luoshu、tianwen→ganzhi、http→tables）。 |
| 脚本序言 `set -eo pipefail` + `ROOT="$(cd "$(dirname "$0")/.." && pwd)"`（7 个脚本） | 每个脚本必须能独立运行（`make`、CI、手工直调）；抽公共 include 会引入 sourcing 依赖。 |
| 空文件 `tests/__init__.py` | 根 `tests/` 显式建包，pytest 导入名是 `tests.<module>`；`counsel/tests/` 无 `__init__.py`、导入名是裸 basename，两套测试的模块名不会互相覆盖。 |
| `skills/liki/SKILL.md` 与渠道包内 `skills/liki/SKILL.md` 正文同源 | 根 `SKILL.md` 是唯一手写身份源，渠道包内入口由构建器按平台格式渲染（WorkBuddy 另加方法论卡索引）；同源保证市场入口、ADK skill 与部署 instruction 不漂移（`make check` 漂移门）。 |
| 单文件目录 `tools/` | 目录里只有生成物 `mcp-tool-catalog.json`：保持 `scripts/check_mcp_schema.py --emit` 生成、`make check` 漂移检测，而不是合并目录或手改。 |

## 设计原则（为什么这样设计）

- **MCP 调用边界分层**：`engine-mcp` 只做确定性排盘/历法/风水计算，`counsel-mcp` 只做因子、断语、起名和问卦判断。edge 负责 `/engine` 与 `/counsel` 公开前缀；服务内只允许 `/mcp`。领域文档描述能力与变量绑定，不手写传输层细节。
- **历史事件只验证整体框架**：校准/结论验证回退的对象是「格局+用神+大运」的综合解读框架，不是单一用神选择——事件是框架的综合结果，无法反推单一变量（v1.23.0 教训）。落地处：`references/natal-app-mingshu.md` 历史事件校准节、`references/natal-bazi-calibration.md`。
- **用神只有扶抑 / 调候两派，必须聚合出唯一结论**：格局（格神）不是用神，只作结构证据参与聚合决策（`references/natal-bazi-yongshen.md`）；不并列列出让用户选——并列等于把专业判断推给用户（v1.16.0 教训，已落地 yongshen.md 聚合决策表）。
- **Skill 文档正向引导优先**：流程性指引写「做什么、怎么做」；域规则性禁令（档位防跳、筛字池细则、反重复红线等从真实错误沉淀的专业精度）保留原样，不属于负面清理对象。工具自省、隐私合规等样板只在各 surface 的入口文档存一份，域卡不重复。

## 相关文档

- [docs/SYSTEM.md](./docs/SYSTEM.md)：系统全况与本仓对齐——四仓分工、权威源、专家定义结构、MCP 连接器模型、生成/校验链、变更传播。
- [docs/MCP_CONTRACTS.md](./docs/MCP_CONTRACTS.md)：MCP 契约模型——typed wire DTO → JSON Schema → 运行时校验，engine / counsel 两侧模式。
- [distribution/workbuddy/README.md](./distribution/workbuddy/README.md)：WorkBuddy 技能 / 专家包结构、分发产物与校验流程。

## 发布版本模型

- 内部运行时不用发布版本，构建身份是 Git short SHA；镜像 tag 和 OCI revision
  同源。
- 不恢复手工版本文件、`make version` 或内部版本对账。
- WorkBuddy / SkillHub 平台若强制 semver，由 `scripts/version_projection.py`
  在打包时从 Git commit 时间派生；这是外部平台兼容层，不是内部 release contract。

## 推送前检查清单

**改方法名/函数名时**（全量搜索所有引用点）：

```bash
# 搜代码（含脚本）
grep -rn "旧方法名" --include="*.go" --include="*.sh" --include="*.py"
# 搜 skill 文档
grep -rn "旧方法名" skills/liki/references/*-app-*.md skills/liki/references/*-ENTRY.md skills/liki/SKILL.md
```

**改领域概念/字段名时**（唯一命名源和别名门禁）：

```bash
python3 scripts/sync_domain_language.py
python3 scripts/check_domain_language.py
```

不要在代码、契约、注释、测试或文档中发明同义词。若必须保留旧字段，先在
`tools/domain-language.json` 登记兼容 `scope`、理由和删除版本；checker 只允许
这些位置出现旧名。

**添加新 MCP 工具时**（Pydantic / typed registry 契约是唯一事实源，同步补契约与测试）：

```bash
# engine：补 handler、InputSchema/OutputSchema 与 tool schema 测试
grep -R "bazi_chart" engine/cmd/engine-mcp engine/internal/mcpcontract

# counsel：在 app/contracts 补 Pydantic 契约，再补业务实现与错误映射测试
grep -R "natal_compute_factors" counsel/app counsel/tests
```

**改 skill 文档后**（需要出包时）：

```bash
make build-archive
```

**推送前本地 CI**：

```bash
npm ci # once, for the locked local Markdown lint dependency
# 全量（engine + skills 单测 + 全链路集成，本地自动起引擎）+ schema/文档一致性
make gate
```

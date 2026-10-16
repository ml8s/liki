# 贡献指南

感谢你考虑为liki.hk贡献！

## 提交 Issue

- 命理计算结果不准确 → 请附上出生时间、地点、期望结果和实际结果
- SKILL 行为异常 → 请说明触发了什么流程、期望什么、实际发生了什么
- 建议新技能 → 请描述用户场景，越具体越好

## 分支模型

`main` 是唯一的长期分支；不要使用 `dev`、`master` 或长期发布分支。维护者在 `main`
上做小而聚焦的提交。外部贡献者可以 fork 后创建短生命周期分支，PR 目标仍然是 `main`。

发布永远从已通过 `candidate-test` 和 promotion 的 `main` commit 打 `v<CalVer>` tag。

## 提交变更

1. 同步最新 `main`：`git pull --ff-only`
2. 安装 git hooks（一次）：`make hooks`
3. 升版本用根 Makefile 统一写入当日日期和序号（`VERSION.txt`、engine/counsel VERSION、根与专家包 SKILL 元数据、插件版本、profiles、counsel 投影契约同步）：

   ```bash
   make version
   ```

4. 同步更新 `CHANGELOG.md`（README 统计数字有变时一并更新）
5. README / 用户指南改动需保持现有标题结构和双语一致性，并运行 `make check`
6. 提交时描述清楚改了什么、为什么；不相关的变更保持独立提交

## 代码规范

### 本地工具链

版本只在一处声明，CI 与本地读同一份，`make check` 校验本地一致：

- Go 1.26.9：`go.mod` + `GOTOOLCHAIN`（`make go-version-check`）
- Python 3.12.14：`.python-version`（`make runtime-version-check`）
- Node 24.14.1：`.nvmrc`（`make runtime-version-check`）
- golangci-lint v2.13.2：`deps/golangci-lint.version`（本地与 CI 共用，对账测试锁定）
- ruff 0.16.0：`lint-python.sh` 的 `RUFF_VERSION`

GitHub Action 版本一律 40 位 SHA pin；升级由 Dependabot 每周分组一个 PR 跟进。
合并前看 release notes——大版本常有 breaking（如 docker/login-action v4）。
可选：[uv](https://docs.astral.sh/uv/)，用于复现 counsel Python 依赖锁

### 语言和服务边界

- 根 `skills/liki/SKILL.md` 只做产品路由；每个领域用 `ENTRY.md` 进入。SKILL.md 以中文为主，术语保持原文。
- 领域 `ENTRY.md` 只做域内路由、工具契约与硬边界；`app` 卡按用户目标命名（`outcome.md`、`marriage.md`…）并承载流程、交互闸口与输出模板，方法论卡按体系命名（`bazi/`、`ziwei/`、`liuyao/`…）；一份信息只在一处定义（`tests/test_unified_skill_structure.py` 钉住路由可达与引用形态）。
- **专家 = 顶层领域（1:1）**：`expert-packs/liki-<domain>` 与根 `skills/liki/references/<domain>/` 同名（`<domain>` = `natal` / `divination` / `fengshui` / `naming`）；包内 `skills/<domain>/references/` 是该子树的整棵镜像，领域卡只写本体系知识、`app/` 承载场景与跨体系合参（跨域路径引用在生成器 fail-closed）。`agents/<domain>/agent.yaml`（含 router 的 `sub_agents`）与 `profiles/experts.json` 必须同步为同一套领域。
- **卡写作约定（`scripts/check_docs.py` + `tests/test_unified_skill_structure.py` 钉住）**：
  方法论卡（`references/*/app/*.md`、`references/*/domains/**`）不带 frontmatter——卡的身份由目录与 `ENTRY.md` 路由表表达，元数据写在卡内没有消费方；
  `## 输出模板` 必须给模板体（代码块 / 列表 / 表），示例放 `### 示例` 之下；
  一条规则只在一处定义——域级公共口径进该域 `ENTRY.md` / `TOOLS.md`，跨体系合参进 `app/`，被专家包自包含逼出的同源双份（如 natal / naming 的用神优先序）用测试钉成同字；
  卡内只引工具字段名与工具名，不写实现内部数据文件名（`data/*.json`）。
- 专家身份的唯一手写源是包根 `expert-packs/liki-<domain>/SKILL.md`（元数据 + 角色正文，内部部署与市场包共用；`plugin.json` 的 `agents`/`agentName` 指向它本身）；域 `skills/<domain>/SKILL.md`（含方法论卡索引）与 `plugin.json` 元数据由 `scripts/sync_expert_packs.py` 生成，`make check` 漂移门守护且拒绝包内出现第二份身份副本；无专家包的 router 身份文件是 `agents/router/router.md`。
- `engine-mcp` 只做确定性排盘/历法/风水计算；`counsel-mcp` 只做因子、断语、起名和问卦判断。
- 公开 `/engine` 与 `/counsel` 前缀由网关负责；两个服务内部只实现 `/mcp` 与 `/mcp/{domain}`。
- 不恢复旧 Python CLI 或 JSON-RPC 发现/调用路径；历史评测资产已归档，迁移前不得进入 gate。

### Lint 与安全扫描

- 引擎 lint 用 golangci-lint v2（配置 `engine/.golangci.yml`）。本地安装用官方二进制脚本，**不要 `go install`**（golangci-lint 与 Go 版本强耦合，官方明确不推荐该方式）：

  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

- Python lint 使用 Ruff 0.16.0：`make lint-python`。
- Go 漏洞扫描使用 `govulncheck`；counsel 哈希锁使用 `pip-audit`。两者在 CI 分别执行。
- 每次升版本必须同步更新：`VERSION.txt`（make 统一 bump）+ `CHANGELOG.md`；需要生成分发包时运行 `make build-archive`

## 有意保留的重复（不要合并）

评审时下面这些看起来可以合并或去重，但每一项都是刻意的；动手之前先确认理由仍然成立。

| 重复 | 为什么保留 |
| --- | --- |
| `engine/**/zz_load_test.go`（13 个，除 `package` 行外逐字节相同） | Go 不能跨包共享 `TestMain`；抽成公共包会把 `testing` 带进非测试构建依赖。每个测试包各自 `Load()` 一次数据表。 |
| 脚本序言 `set -eo pipefail` + `ROOT="$(cd "$(dirname "$0")/.." && pwd)"`（7 个脚本） | 每个脚本必须能独立运行（`make`、CI、手工直调）；抽公共 include 会引入 sourcing 依赖。 |
| 版本面 `skills/liki/VERSION.txt`、`engine/cmd/engine-mcp/VERSION`、`counsel/app/VERSION.txt` | 三个工件可独立分发，各自需要运行时版本面；同步由 `tests/test_version_surfaces.py` 钉住。 |
| 空文件 `tests/__init__.py` | 根 `tests/` 显式建包，pytest 导入名是 `tests.<module>`；`counsel/tests/` 无 `__init__.py`、导入名是裸 basename，两套测试的模块名不会互相覆盖。 |
| 镜像卡之间逐卡重复的边界 / 参数句（如 `divination/domains/qimen/*` 专占卡的公共排盘步、`naming/app/*` 卡的公共 `[必读]` 声明） | 每张卡随专家包独立分发，必须在包内自包含；合并会引入跨包引用或包内死链。 |
| `skills/liki/references/natal/domains/{bazi,ziwei}/calibration.md` 的考时六步 | 两张卡会被 `sync_expert_packs.py` 分别镜像进单侧专家包，包内只有一侧，卡片必须自包含；差异仅“公历 → 农历”与年号界两处，合并即产生跨侧死链。 |
| 包根 `SKILL.md` 与 `skills/*/SKILL.md` 正文同源 | 包根 `SKILL.md` 是唯一手写身份源，域 `SKILL.md` 由生成器按 ADK skill 格式渲染（另加自动方法论卡索引）；同源保证市场入口、ADK skill 与部署 instruction 三面不漂移（`make check` 漂移门）。 |
| 单文件目录 `tools/` | 目录里只有生成物 `mcp-tool-catalog.json`：保持 `scripts/check_mcp_schema.py --emit` 生成、`make check` 漂移检测，而不是合并目录或手改。 |

## 设计原则（为什么这样设计）

- **MCP 调用边界分层**：`engine-mcp` 只做确定性排盘/历法/风水计算，`counsel-mcp` 只做因子、断语、起名和问卦判断。网关负责 `/engine` 与 `/counsel` 公开前缀；服务内只允许 `/mcp` 与 `/mcp/{domain}`。领域文档描述能力与变量绑定，不手写传输层细节。
- **历史事件只验证整体框架**：校准/结论验证回退的对象是「格局+用神+大运」的综合解读框架，不是单一用神选择——事件是框架的综合结果，无法反推单一变量（v1.23.0 教训）。落地处：`app/mingshu.md` 历史事件校准节、`domains/bazi/calibration.md`。
- **三派用神必须聚合出唯一结论**：扶抑/调候/格局三派按决策表聚合（`domains/bazi/yongshen.md`），不并列列出让用户选——并列等于把专业判断推给用户（v1.16.0 教训，已落地 yongshen.md 聚合决策表）。

## 相关文档

- [docs/SYSTEM.md](./docs/SYSTEM.md)：系统全况与本仓对齐——四仓分工、权威源、专家定义结构、MCP 连接器模型、生成/校验链、变更传播。
- [docs/MCP_CONTRACTS.md](./docs/MCP_CONTRACTS.md)：MCP 契约模型——typed wire DTO → JSON Schema → 运行时校验，engine / counsel 两侧模式。

## 发布版本模型

- 日常开发 / 兼容性版本使用 CalVer，由 `make version` 统一 bump。
- 正式产品发行版使用日期版本 Git tag（如 `v2026.10.12.0`），与运行时 CalVer 同源。
- 不要恢复并行 SemVer 版本面；CalVer 是唯一版本标识。
- 完整规则见 [docs/RELEASE_MODEL.md](./docs/RELEASE_MODEL.md)。

## 推送前检查清单

**改方法名/函数名时**（全量搜索所有引用点）：

```bash
# 搜代码（含脚本）
grep -rn "旧方法名" --include="*.go" --include="*.sh" --include="*.py"
# 搜 skill 文档
grep -rn "旧方法名" skills/liki/references/*/app/*.md skills/liki/references/*/ENTRY.md skills/liki/SKILL.md
```

**添加新 MCP 工具时**（Pydantic / typed registry 契约是唯一事实源，同步补契约与测试）：

```bash
# engine：补 handler、InputSchema/OutputSchema、领域路由与 tool schema 测试
grep -R "bazi_chart" engine/cmd/engine-mcp engine/internal/mcpcontract

# counsel：在 app/contracts 补 Pydantic 契约，再补业务实现与错误映射测试
grep -R "compute_factors" counsel/app counsel/tests
```

**改 skill 文档后**（需要出包时）：

```bash
make build-archive
```

**推送前本地 CI**：

```bash
# 全量（engine + skills 单测 + 全链路集成，本地自动起引擎）+ schema/文档一致性
make gate
```

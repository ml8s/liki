# TODO — Liki skill 单一真相源收尾（交接）

> 状态：本轮「单一真相源」重构已在工作树完成并通过 `make check` / `make test`，**未提交、未 bump 版本、未记 CHANGELOG**。
> 本文件只记录待决策/待修项与证据，供后续接手人执行；完成后请删除本文件或归档。

## 0. 现行架构速查（改动前先读）

| 内容类型 | 唯一手写源 | 生成物 |
| --- | --- | --- |
| 方法论卡 | `skills/liki/references/` | `expert-packs/liki-<domain>/skills/<domain>/references/*.md` |
| 专家身份 | `expert-packs/liki-<domain>/SKILL.md` | 域 `skills/<domain>/SKILL.md`、`plugin.json` 的 `agents`/`agentName`/`description`/`profession`/`version` |
| Router 身份 | `agents/router/router.md` | 部署 profile 的 router instruction |
| 部署配置 | `agents/<name>/agent.yaml` | `dist/agents/**` |

- 生成入口：`python3 scripts/sync_expert_packs.py`（`make sync-expert-packs`）；漂移门：`--check`（已并入 `make check`）。
- 包内**只允许一份身份文件**（包根 `SKILL.md`）；出现 `agents/` 身份副本时生成器直接报错。
- 身份 frontmatter 契约（`IDENTITY_KEYS`，见 `scripts/sync_expert_packs.py`）：`name` / `description` / `description_en` / `display_name` / `display_name_en` / `profession` / `keywords` / `skills` / `max_turns`（`version`/`author` 由 bump 写入）。
- 注意：身份文件的 `keywords` / `skills` / `max_turns` 目前**只被校验、无消费方**（生成器与部署器都不读），属于「写而无用」的契约，见待办 6。
- 改身份文件后必须 `make sync-expert-packs`，再 `make check`（含双向漂移与版本面）。

## 1. 关键词过度承诺（P1）

**问题**：SEO/检索词宣称了方法论未覆盖的能力。

**证据**：`skills/liki/references/` 内检索 `公司|品牌|店铺|选址|办公室|名字测试` 命中 **0**；但下列三处均出现这些词：

- `skills/liki/SKILL.md` 的 `metadata.keywords`：含 `公司起名`、`品牌命名`、`名字测试`、`办公室风水`、`店铺选址`。
- `expert-packs/liki-naming/SKILL.md` 的 `keywords`：含 `公司起名`、`品牌命名`、`名字测试`。
- `expert-packs/liki-fengshui/SKILL.md` 的 `keywords`：含 `办公室风水`、`店铺选址`。
- `tests/test_skill_metadata_seo.py` 的 `KEYWORD_PHRASES` 是同一列表的**镜像副本**，必须与上面同改，否则测试会继续要求这些词。

**同类问题（手写的 plugin.json 市场文案）**，两处 quickPrompts 同样过度承诺：

- `expert-packs/liki-naming/.codebuddy-plugin/plugin.json`：`起个公司名/品牌名`。
- `expert-packs/liki-fengshui/.codebuddy-plugin/plugin.json`：`办公室/店铺选址看哪个方位好`。

quickPrompts / tags / displayDescription / defaultInitPrompt 不参与生成，改完无需 sync。

**动作（建议：删承诺，不新增方法论）**：要么补齐对应方法论卡，要么删除这些词。

1. 从上述 3 个身份/根文件中删除未覆盖词；liki 根 keywords 由 25 词降至 20 词。
2. 同步更新 `tests/test_skill_metadata_seo.py` 的 `KEYWORD_PHRASES`。
3. 替换两处 plugin.json 的过度承诺 quickPrompt 为已覆盖场景（naming 可用「给外国人起个中文名」或「帮我改个名字」；fengshui 可用命卦/方位适配类）。
4. `make sync-expert-packs && make check && make test`。

## 2. 紫微侧「合盘」承诺缺方法论（P1）

**证据**：`skills/liki/references/natal/domains/ziwei/` 内无 `hepan` 类卡（`rg 合盘|bond|双人` 命中 0）；工具 `ziwei_bond` 存在于 `tools/mcp-tool-catalog.json`。但以下位置宣称紫微合盘：

- `expert-packs/liki-ziwei/SKILL.md`：`description`、`keywords`（`紫微合盘`）、`## 核心能力` 第 5 条「合盘、考时」。
- `expert-packs/liki-ziwei/.codebuddy-plugin/plugin.json`：quickPrompt「紫微合盘，看两人是否适配」。
- `agents/router/router.md`：紫微路由行含「合盘、考时」。
- 对照：八字侧有 `skills/liki/references/natal/domains/bazi/hepan.md`，紫微侧无对应卡。

**动作（二选一，需先定）**：a) 补 `hepan` 类卡（对照 bazi/hepan.md 的「双人盘事实合参」结构，与 `ziwei_bond` 对齐）；b) 删除紫微合盘承诺（含 router 路由行、plugin prompt、身份 keywords/能力条目）。
若选 b，注意 `skills/liki/references/natal/ENTRY.md`「合盘（双人）」统一走 `references/natal/app/compatibility.md`，其中紫微侧仅作证据，需确认文案一致。

## 3. liki 根 skill 执行循环缺「委派」分支（P2）

**证据**：`agents/router/agent.yaml` 的 `tools.allow` 只有 `skilltoolset`（无领域工具），因此多专家部署下必须委派；该约束目前**只写在 `agents/router/router.md`**，根 skill 未表达。

**动作**：在 `skills/liki/SKILL.md` 的 `## 执行循环` 第 2 步补一句：当环境未暴露该领域工具时，委派给持有该域工具的专家，本 agent 只做编排与结论汇总（单 agent 客户端不受影响，行为不变）。

## 4. 跨域 app 卡未镜像进专家包（P2，需决策）

**证据**：`scripts/sync_expert_packs.py` 的 `MAPPINGS` 只镜像 `fengshui/app` 与 `naming/app`。`skills/liki/references/natal/app/*`、`skills/liki/references/divination/app/*` 是跨体系编排卡，同时引用 bazi+ziwei（如 `skills/liki/references/natal/app/marriage.md`）或 qimen+liuyao+huangli，单包镜像会产生跨包死链，触发生成器 fail-closed。

**动作（建议 a）**：

- a) 维持现状，并在 `expert-packs/README.md` 明确「跨域编排卡只随根 skill 分发，单领域包只带本域方法论卡与本域场景卡」——**推荐**：这些卡本质是合参编排，拆到单包即失真。
- b) 按侧拆分（`decision.md`/`date.md` → liki-qimen，`outcome.md` → liki-liuyao，`question.md` 跨域 → 仅根），需逐卡改写引用。
- c) 扩展生成器支持跨包链接（改动最大，需放宽死链校验）。

## 5. 运行时契约只随根 skill（P3）

Aipay（`skills/liki/SKILL.md` + `skills/liki/references/payment/AIPAY.md`）、Feedback（`skills/liki/references/ops/FEEDBACK.md` + `skills/liki/feedback.schema.json` + `skills/liki/scripts/feedback.py`）、输出契约只存在于根 skill；单独安装某专家包时没有这些面。

**动作**：明确产品意图——专家包是否独立售卖/独立运行。若需要，抽一份包内最小运行时子集并纳入生成/校验；若不需要，在 `expert-packs/README.md` 写明「专家包依赖 liki 运行时，不单独提供 Aipay/Feedback」。

## 6. 身份 frontmatter 的「写而无用」字段（P3）

`keywords` / `skills` / `max_turns` 被 `scripts/sync_expert_packs.py` 强制要求非空/整数，但没有任何生成物消费它们。

**动作**：要么接入消费方（如由 `keywords` 生成 `plugin.json` 的 `tags`/quickPrompts，由 `max_turns` 生成部署 agent 配置），要么把契约降级为「人类可读元数据」并在文档中说明，避免后续误以为改了就会生效。

## 7. 提交与发布（收尾）

- 重构（`refactor(skills)`)与发布流水线（`feat(ci)`）已分别提交；另一条发布线的 manifest/测试文件已并入流水线提交。
- 打下一个 CalVer tag 前仍需 bump `VERSION.txt`（当前 `2026.10.15.0`）并补 `CHANGELOG.md`。
- 推送门槛仍为 `make gate`。

## 验收命令

```make
make sync-expert-packs   # 身份/方法论改动后重新生成
make check               # 含 sync --check 双向漂移、md/ruff/docs/paths、身份契约
make test                # pytest + engine/counsel/engine-mcp
```

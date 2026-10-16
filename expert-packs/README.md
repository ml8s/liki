# Liki 专家包（CodeBuddy 兼容层）

> 生产系统只发布全功能 `liki-experts` 装配镜像，其部署配置来自仓库根目录 `agents/` 与
> `profiles/`。本目录同时承载**专家身份的唯一手写源**（包根 `<pack>/SKILL.md`：元数据 +
> 角色/指令正文，内部部署与市场包共用同一份）与 CodeBuddy/WorkBuddy 市场包兼容层。

## 包结构

**专家与顶层领域一一对应**（`liki-natal` / `liki-divination` / `liki-fengshui` / `liki-naming`，
与根 `skills/liki/references/<domain>/` 同名）。每个 `liki-<domain>` 包的路径来源固定如下：

| 路径 | 来源 |
| --- | --- |
| `SKILL.md` | **手写：专家身份唯一源**（frontmatter 元数据 + 角色/指令正文，市场入口，也是插件声明的 agent 定义文件） |
| `skills/<domain>/SKILL.md` | 生成，源为身份文件（正文 + 自动方法论卡索引；ADK skill 入口） |
| `skills/<domain>/references/**` | 生成，源为 `skills/liki/references/<domain>/` 整棵子树（ENTRY/TOOLS/app/domains，逐字节镜像 + 路径 rebase） |
| `.codebuddy-plugin/plugin.json` | `agents`/`agentName`/description/profession/version 由身份文件生成，其余字段手写 |
| `.mcp.json` | 手写连接器契约 |
| `avatars/expert.png` | 手写占位资源，见下 |
| `README.md` | 手写包说明 |

领域卡只写本体系知识、`app/` 场景卡承载用户目标与跨体系合参（一份信息只在一处定义）；
因此单领域包天然自包含，不做多对一聚合，也不会产生跨包死链（`sync_expert_packs.py` 对
跨域引用 fail-closed）。

包内**只有一处身份文件**（包根 `SKILL.md`）：`plugin.json` 的 `agents: ["./SKILL.md"]`、
`agentName` 都指向它本身，`make check` 会拒绝重新出现 `agents/` 之类的身份副本。

身份文件 frontmatter 契约（生成器与校验器共同消费）：`name`（插件/skill 名，= `agentName`）、
`description`（域 skill 描述）、`description_en`（`plugin.json.description` + 部署 agent
描述）、`display_name` / `display_name_en`（市场展示名）、`profession`（en/zh）、
`keywords`（→ `plugin.json.keywords` 市场检索词）、`skills`（skill 名列表 → `plugin.json.skills`
预加载路径 `./skills/<name>`）、`maxTurns`（WorkBuddy agent 回合上限，本文件即 agent 定义）、
`version` / `author`；正文即部署 instruction 骨架。

生成物不要手改：`make sync-expert-packs` 覆盖并清理孤儿文件，`make check` 通过
`python3 scripts/sync_expert_packs.py --check` 做双向漂移检测（含域 `SKILL.md` 与
`plugin.json` 元数据）。

## 头像

`avatars/expert.png` 是 512×512、16-bit RGBA 的项目自有占位图：不含任何第三方素材，
因此不附带额外许可义务。它不参与 `sync-expert-packs` 生成，也无需在包之间统一——
替换为正式漫画/插画风头像时，同样必须使用自有或已获授权的素材，并在包 `README.md`
记录来源。

## 校验

`python3 scripts/check_expert_docs.py`（`make check`）校验身份文件的 frontmatter 契约与
边界语义、手写连接器（`.mcp.json`）与市场描述的有效性，并禁止全部描述面硬编码 MCP
方法名；生成物的字节一致性与版本面由 `sync_expert_packs.py --check` 与
`tests/test_version_surfaces.py` 守护。

## 运行时边界

专家包依赖 liki 运行时：Aipay（根 `skills/liki/SKILL.md` +
`skills/liki/references/payment/AIPAY.md`）、Feedback（`skills/liki/references/ops/FEEDBACK.md` +
`skills/liki/feedback.schema.json` + `skills/liki/scripts/feedback.py`）与输出契约只随根
skill 提供，包内不重复携带。若将来要独立售卖 / 独立运行专家包，再评估抽一份包内最小运行时子集。

## 上架前待办

- `avatars/expert.png` 当前为项目自有占位图（纯色 + 文字）；上架前替换为正式漫画/插画风头像。
- 用官方 `expert-manager` 技能逐项校验后上架。

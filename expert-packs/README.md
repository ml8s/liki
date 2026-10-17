# Liki 专家包（CodeBuddy 兼容层）

> 生产系统只发布全功能 `liki-experts` 装配镜像，其部署配置来自仓库根目录 `agents/` 与
> `profiles/`。本目录同时承载**专家身份的唯一手写源**（包根 `<pack>/SKILL.md`：元数据 +
> 角色/指令正文，内部部署与市场包共用同一份）与 CodeBuddy/WorkBuddy 市场包兼容层。

## 包结构

**专家与顶层领域一一对应**（`liki-natal` / `liki-divination` / `liki-fengshui` / `liki-naming`，
与根 `skills/liki/references/<domain>/` 同名）。每个 `liki-<domain>` 包的路径来源固定如下：

| 路径 | 来源 |
| --- | --- |
| `SKILL.md` | **手写：专家身份唯一源**（frontmatter 元数据 + 角色/指令正文；包内其余入口都由它投影） |
| `skills/<domain>/SKILL.md` | 生成，源为身份文件（正文 + 自动方法论卡索引；ADK skill 入口） |
| `agents/<pack>.md` | 生成，源为身份文件（WorkBuddy agent md：平台字段名 `displayName`/`maxTurns` + 身份正文；`plugin.json` 的 `agents`/`agentName` 指向它） |
| `skills/<domain>/references/**` | 生成，源为 `skills/liki/references/<domain>/` 整棵子树（ENTRY/TOOLS/app/domains，逐字节镜像 + 路径 rebase） |
| `.codebuddy-plugin/plugin.json` | `agents`/`agentName`/description/profession/version/keywords/skills/dependencies 生成（含连接器依赖），其余字段手写 |
| `avatars/expert.png` | 手写占位资源，见下 |
| `README.md` | 手写包说明 |

领域卡只写本体系知识、`app/` 场景卡承载用户目标与跨体系合参（一份信息只在一处定义）；
因此单领域包天然自包含，不做多对一聚合，也不会产生跨包死链（`sync_expert_packs.py` 对
跨域引用 fail-closed）。

包内**只有一处身份文件**（包根 `SKILL.md`，手写）；平台入口 `agents/<pack>.md` 由它投影
（平台字段名与文件名约定见该文件 frontmatter），`agents/` 只允许这一个文件——多出来的即身份副本
回流，`sync_expert_packs.py --check` 与 `check_expert_docs.py` 都会拦。

身份文件 frontmatter 契约（生成器与校验器共同消费）：`name`（插件/skill 名，= `agentName`）、
`description`（域 skill 描述）、`description_en`（`plugin.json.description` + 部署 agent
描述）、`display_name` / `display_name_en`（市场展示名）、`profession`（en/zh）、
`keywords`（→ `plugin.json.keywords` 市场检索词）、`skills`（skill 名列表 → `plugin.json.skills`
预加载路径 `./skills/<name>`）、`maxTurns`（WorkBuddy agent 回合上限，本文件即 agent 定义）、
`version` / `author`；正文即部署 instruction 骨架。连接器依赖不在身份文件里，它的单一源是
`workbuddy/connector-ids.json`（→ `plugin.json.dependencies.connectors`）。

生成物不要手改：`make sync-expert-packs` 覆盖并清理孤儿文件，`make check` 通过
`python3 scripts/sync_expert_packs.py --check` 做双向漂移检测（含域 `SKILL.md` 与
`plugin.json` 元数据）。

## 头像

每个包的头像由同目录矢量源 `avatars/expert.svg` 生成（`make workbuddy-avatars` → 512×512
PNG，8-bit RGBA，单张 ≤500KB，符合 WorkBuddy 头像规范）。四张是同一套设计：深蓝圆角底 +
金色田字格 + 一个本域汉字——**命**（本命）、**卦**（占卜）、**宅**（风水）、**名**（起名）。
设计常量（底色、田字格几何、字高、字面表）集中在 `scripts/gen_brand_marks.py`
（连接器图标用的是同一脚本、同一母题），`make brand-marks` 据此重建四份矢量源并栅格化；
换字、改配色只改该脚本顶部。

汉字取自阿里妈妈刀隶体（`Alimama DaoLiTi`，阿里巴巴免费商用字体；本仓不附带字体文件）：
生成时用 `inkscape -T` 把字转成轮廓路径，源里只留路径——不出现 `<text>`、`font-family`、
`dominant-baseline`。市场宿主的渲染字体不可控，靠系统字体取字会静默换字形或错位，
因此 `make workbuddy-avatars` 与 `make check`（`scripts/check_expert_docs.py`）都会拒绝
带文本依赖的源。田字格与配色为项目自绘、无第三方素材；字形轮廓的字体来源记录在此，
矢量源不参与 `sync-expert-packs` 生成。若换成人物插画，同样必须使用自有或已获授权的素材，
并在包 `README.md` 记录来源。

## 校验

`python3 scripts/check_expert_docs.py`（`make check`）校验身份文件的 frontmatter 契约与
边界语义、`plugin.json.dependencies` 的连接器依赖（且包内不得带 `.mcp.json`）与市场描述
的有效性，并禁止全部描述面硬编码 MCP 方法名；生成物的字节一致性与版本面由
`sync_expert_packs.py --check` 与 `tests/test_version_surfaces.py` 守护。

## 运行时边界

专家包依赖 liki 运行时：Aipay（根 `skills/liki/SKILL.md` +
`skills/liki/references/payment/AIPAY.md`）、Feedback（`skills/liki/references/ops/FEEDBACK.md` +
`skills/liki/feedback.schema.json` + `skills/liki/scripts/feedback.py`）与输出契约只随根
skill 提供，包内不重复携带。若将来要独立售卖 / 独立运行专家包，再评估抽一份包内最小运行时子集。

## 端点归属

专家**不自带 MCP**，只声明对已上架连接器的依赖：

```json
"dependencies": { "connectors": ["oc_9af58cd723529afc", "oc_f31cd1614c314762"] }
```

两个 ID 的单一源是 `workbuddy/connector-ids.json`（`source` = `liki-engine` / `liki-counsel`），
由 `sync_expert_packs.py` 投影进四个 `plugin.json`；`check_expert_docs.py` 比对投影值，并禁止包内
出现 `.mcp.json`——平台在 `plugin.json` 未声明 `dependencies.mcpServers` 时会自动读取插件根
`.mcp.json` 作为兜底，两个来源并存会让同一次召唤出现两套工具面与两组引导卡片。

选择连接器聚合端点而不是自带域级端点，是为了让一次连接在四个专家间复用（引导卡片 2 张而非
12 张）、让连接器成为工具面的唯一来源（专家包不再重复声明端点），引导卡片也用连接器自己的图标。
代价是专家看到的工具面是全领域：领域边界由身份文件正文的「硬边界」承担（如本命专家：六爻、
奇门、黄历择日、起名、风水属其他专家），工具名本身也带领域前缀。

子域端点（`engine/mcp/{bazi,ziwei,liuyao,qimen,huangli,fengshui}`、
`counsel/mcp/{bazi,ziwei,liuyao,qimen,naming}`，见 `engine/cmd/engine-mcp/server.go` 的
`mcpDomains`）保留给自建分域部署，不参与 WorkBuddy 上架。

## 上架前待办

- 用官方 `expert-manager` 技能逐项校验后用 `make workbuddy-package` 出的 zip 上架。

## 版本面

平台按 semver 解析制品版本（实测 4 段 CalVer 被判「格式不合法」），所以包内版本面（包根
`SKILL.md`、域 `SKILL.md`、`.codebuddy-plugin/plugin.json`）统一写发行 CalVer 的平台投影：
`2026.10.16.2` → `2026.1016.2`，规则定义在 `scripts/version_projection.py`，单一源仍是
`skills/liki/VERSION.txt`，由 `make version` 更新、`tests/test_version_surfaces.py` 守护。

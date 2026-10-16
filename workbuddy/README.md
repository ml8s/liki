# WorkBuddy 上架交付物

本目录是 WorkBuddy 开放平台的上架交付根。本仓提交三类上架物：

| WorkBuddy 上架物 | 本仓目录 | 交付形态 |
| --- | --- | --- |
| 连接器 | `workbuddy/<source>-connector/` | `connector-meta.json` + `mcp.json`（或 `cli.json`）+ `icon.svg` |
| 技能 | `skills/liki/` | `SKILL.md` + `references/`（领域方法论与工具契约的唯一真相源） |
| 专家 | `expert-packs/liki-natal/`、`expert-packs/liki-divination/`、`expert-packs/liki-fengshui/`、`expert-packs/liki-naming/` | `.codebuddy-plugin/plugin.json` + `avatars/` + 身份文件 + `skills/` |

## 连接器

一个连接器只能声明一个 MCP Server；`connector-meta.json` 的 `source` 全局唯一、只用小写字母/数字/连字符。

| 目录 | source | MCP Server | 能力 |
| --- | --- | --- | --- |
| `workbuddy/liki-engine-connector/` | `liki-engine` | `https://liki.hk/engine/mcp` | 排盘与历法计算：八字、紫微、六爻、奇门、黄历、风水 |
| `workbuddy/liki-counsel-connector/` | `liki-counsel` | `https://liki.hk/counsel/mcp` | 规则判断：因子快照、本命断语、应期断语、合盘、起名 |

两个 Server 都是免鉴权远程服务，因此不写 `auth_mode`，`mcp.json` 中不出现任何凭证。

两个连接器的 `version` 与技能/专家同源：写发行版本 `skills/liki/VERSION.txt` 的平台投影
（`scripts/version_projection.py`，`2026.10.16.2` → `2026.1016.2`），由 `make version` 统一更新，
`check_workbuddy_connectors.py` 比对投影值，禁止手写。

连接器只承担「接入」：AI 侧的使用说明与领域方法论由独立技能 `skills/liki/` 承担（工具契约见各域的 `TOOLS.md`），连接器包内不复制第二份，同一事实只维护一处。平台侧同样自洽——连接器规范把
MCP 方案的 `skills/` 列为可选（「MCP 已提供标准工具描述时可选」），而本仓工具层是契约强制的：
`engine/internal/mcpcontract` 测试要求每个工具的描述与输出字段描述非空，`check_mcp_schema.py`
要求每个工具、必需输出字段与枚举字段的描述非空，全量契约列在 `tools/mcp-tool-catalog.json`；
工具链顺序也由契约决定（判断工具吃排盘工具的输出）。这里不再放第二份工具说明，避免同一事实
两处维护。

连接器图标与专家头像同族（同一母题、同一脚本）：深色圆角底 + 金色田字格 + 本域汉字——
排盘引擎取「盘」、规则判断取「断」，与专家头像的 命/卦/宅/名 成一系列。平台图标规范对背景
只写「背景 建议透明」，这里选统一与耐看：深色圆角牌自带底色，亮暗主题下都立得住，金字始终
有对比度（透明底配深色字在深色主题下会整个消失）。

图标由 `scripts/gen_brand_marks.py` 生成、`make brand-marks` 重建，按规范直接交 SVG
（PNG/JPG 建议 64×64）；源里只有路径，不含 `<text>` 与字体依赖，`make check` 会拒绝
带文本字形的图标。

## 技能

`skills/liki/` 按技能规范直接打包提交：`{skill-name}/SKILL.md` + `references/`。

frontmatter 已按平台必填补齐：顶层 `description` / `description_zh` / `description_en` / `version` / `author`（`version` 写 `skills/liki/VERSION.txt` 的平台投影：平台按 semver 解析，4 段 CalVer 会被判非法；`metadata` 只留 ADK 侧字段）。仍待实探的两点：包根附加文件（`.mcp.json` / `VERSION.txt` / `feedback.schema.json`）平台是否认，以及文档的 `@references/xxx.md` 引用写法是否为必需。

## 提交

官网只收 zip，三类提交物统一由 `make workbuddy-package` 打包到 `dist/workbuddy/`（`dist/` 不入库），
zip 顶层就是提交单元目录本身：

| zip | 顶层目录 | 内容 |
| --- | --- | --- |
| `liki-engine-connector.zip`、`liki-counsel-connector.zip` | `<source>-connector/` | `connector-meta.json` + `mcp.json` + `icon.svg` |
| `liki.zip` | `liki/` | `SKILL.md` + `references/`、`scripts/` 等子资源 |
| `liki-natal.zip`、`liki-divination.zip`、`liki-fengshui.zip`、`liki-naming.zip` | `<pack>/` | `.codebuddy-plugin/` + `agents/` + `avatars/` + `skills/` + `.mcp.json` + `README.md` |

打包脚本自带逐字节校验（`scripts/check_workbuddy_zips.py`：与源码目录比对、顶层目录名与
`plugin.json.name` 一致、专家包不带包根身份手写源 `SKILL.md`）。

1. 登录 <https://open.workbuddy.cn/console>，完成入驻与资质认证。
2. 在控制台创建对应业务类型（连接器 / 技能 / 专家），提交上表里的 zip，交 WorkBuddy 团队审核；
   审核通过后进入市场，后续更新重新提交，通常 10～15 分钟同步生效。

## 校验

```sh
python3 scripts/check_workbuddy_connectors.py   # 连接器源码目录
python3 scripts/check_workbuddy_zips.py         # dist/workbuddy/ 下的提交 zip
```

两条都在 `make check` 里。前者校验元信息必填字段与描述字数、`source` 唯一性与命名、单一 MCP
Server、远程地址 HTTPS、`minWorkbuddyVersion` 与所用字段的最低版本匹配、示例条数、图标存在
（字形须落成路径）、包内无真实凭证；后者把 zip 与源码逐字节比对并校验平台解析要求的目录约定
（干净 checkout 没打包时自动跳过）。

## 上架前核对

- 在控制台确认 `liki-engine`、`liki-counsel` 两个 `source` 的占用情况；历史用过的 `source` 有 `liki-engine`、`engine-pro`、`engine-pro-bazi`、`engine-pro-ziwei`。
- `minWorkbuddyVersion` 为 `4.24.0`：用到 `name_zh`/`name_en` 与 `examples_zh`/`examples_en` 这两个 4.24.0 字段。
- 版本面已定：技能/专家/连接器三类制品统一写发行 CalVer 的平台投影（`scripts/version_projection.py`）。
- agent 定义路径与字段形态、端点归属、按域头像已按平台文档落地（见 `expert-packs/README.md`）。

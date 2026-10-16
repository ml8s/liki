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

连接器只承担「接入」：AI 侧的使用说明与领域方法论由独立技能 `skills/liki/` 承担（工具契约见各域的 `TOOLS.md`），连接器包内不复制第二份，同一事实只维护一处。

## 技能

`skills/liki/` 按技能规范直接打包提交：`{skill-name}/SKILL.md` + `references/`。

提交前需补齐平台必填的 frontmatter 字段：`description_zh`、`description_en`、`author`，以及顶层 `version`（现为 `metadata.version`）。补进真相源、还是由打包步骤生成一份提交视图，待定——顶层 `version` 与 `metadata.version` 并存即形成双源，需二选一。

## 提交

1. 登录 <https://open.workbuddy.cn/console>，完成入驻与资质认证。
2. `make workbuddy-package` 产出 `dist/workbuddy/<source>-connector.zip`（zip 顶层就是连接器目录，内含 `connector-meta.json`、`mcp.json`、`icon.svg`）。
3. 在控制台创建连接器并提交该 zip，交 WorkBuddy 团队审核；审核通过后进入连接器市场，后续更新重新提交，通常 10～15 分钟同步生效。

## 校验

```sh
python3 scripts/check_workbuddy_connectors.py
```

已包含在 `make check` 中，校验内容：元信息必填字段与描述字数、`source` 唯一性与命名、单一 MCP Server、远程地址 HTTPS、`minWorkbuddyVersion` 与所用字段的最低版本匹配、示例条数、图标存在、包内无真实凭证。

## 上架前核对

- 在控制台确认 `liki-engine`、`liki-counsel` 两个 `source` 的占用情况；历史用过的 `source` 有 `liki-engine`、`engine-pro`、`engine-pro-bazi`、`engine-pro-ziwei`。
- `minWorkbuddyVersion` 为 `4.24.0`：用到 `name_zh`/`name_en` 与 `examples_zh`/`examples_en` 这两个 4.24.0 字段。
- 专家包的待补项（agent 定义路径与字段形态、`version` 语义化、端点归属、按域头像）见 `expert-packs/README.md`「上架前待办」。

# WorkBuddy 上架交付物

本目录是 WorkBuddy 上架交付根。Liki 只上架一个专家：`expert-packs/liki`。它通过两个连接器访问同一套 MCP 能力。

| 上架物 | 目录 | 交付物 |
| --- | --- | --- |
| 连接器 | `workbuddy/liki-engine-connector/`、`workbuddy/liki-counsel-connector/` | `connector-meta.json` + `mcp.json` + `icon.svg` |
| 技能 | `skills/liki/` | `SKILL.md` + `references/` + `scripts/` 等运行资源 |
| 专家 | `expert-packs/liki/` | `.codebuddy-plugin/plugin.json` + `agents/` + `avatars/` + `skills/` + `README.md` |

“领域”只是专家内部路由：natal、divination、naming、fengshui 是 `skills/liki/references/` 里的入口卡，不是独立专家包。

## 连接器

| 目录 | source | MCP Server | 能力 |
| --- | --- | --- | --- |
| `workbuddy/liki-engine-connector/` | `liki-engine` | `https://liki.hk/engine/mcp` | 确定性排盘、历法计算与事实输出 |
| `workbuddy/liki-counsel-connector/` | `liki-counsel` | `https://liki.hk/counsel/mcp` | 断语、因子、应期、合盘与起名评估 |

两个连接器都是免鉴权远程 MCP，因此不声明 `auth_mode`，`mcp.json` 内不含凭证。

平台分配的连接器 ID 保存在 `workbuddy/connector-ids.json`，由 `scripts/sync_expert_packs.py` 投影到 `expert-packs/liki/.codebuddy-plugin/plugin.json` 的 `dependencies.connectors`。专家包自身不带 `.mcp.json`，连接器 ID 与平台 `source` 分开维护。

连接器版本由 `skills/liki/VERSION.txt` 经 `scripts/version_projection.py` 投影；`make version` 统一更新，`scripts/check_workbuddy_connectors.py` 校验。

连接器只负责接入。使用规则和方法论在 `skills/liki/` 维护；MCP 工具描述由 engine / counsel 契约测试和 `tools/mcp-tool-catalog.json` 守护。

## 技能与专家

`skills/liki/SKILL.md` 是根技能入口，负责路由、MCP 使用规则、输出契约和运行边界。

`expert-packs/liki/SKILL.md` 是 WorkBuddy 专家身份的唯一手写源。`make sync-expert-packs` 由此生成：

- `expert-packs/liki/skills/liki/`：根技能方法论卡的全量镜像；
- `expert-packs/liki/agents/liki.md`：WorkBuddy agent 入口；
- `expert-packs/liki/.codebuddy-plugin/plugin.json` 中的可投影字段。

`plugin.json` 显示为一个专家 `liki`（Liki 命理师），依赖两个连接器，内部按用户意图路由到领域入口卡。

品牌资源由 `make brand-marks` 生成：连接器图标分别是「盘」「断」，专家头像是「命」。矢量源不使用 `<text>` 或字体依赖。

## 提交与校验

```sh
make workbuddy-package
python3 scripts/check_workbuddy_connectors.py
```

`make workbuddy-package` 生成：

| zip | 顶层目录 |
| --- | --- |
| `liki-engine-connector.zip`、`liki-counsel-connector.zip` | `<source>-connector/` |
| `liki.zip` | `liki/` |
| `liki-expert.zip` | `liki/`；专家包根身份 `SKILL.md` 只用于内部部署和同步，平台 zip 排除该文件 |

提交前核对：

- WorkBuddy 控制台中两个 connector ID 已填入 `workbuddy/connector-ids.json`。
- `minWorkbuddyVersion` 满足平台字段要求。
- 三个制品版本都来自 `skills/liki/VERSION.txt` 的平台投影。
- `make check` 通过。

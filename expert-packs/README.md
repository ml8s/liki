# Liki 专家包（CodeBuddy 兼容层）

> 生产系统的唯一 AgentDeployment source of truth 是仓库根目录 `agents/` 与 `profiles/`，
> 运行时只发布全功能 `liki-experts` 装配镜像。本目录仅保留给 CodeBuddy 旧市场包兼容，
> 不作为生产专家定义或发布分身。

## 包结构

每个 `liki-<domain>` 包的路径来源固定如下：

| 路径 | 来源 |
| --- | --- |
| `SKILL.md` | 生成 |
| `skills/<domain>/SKILL.md`、`skills/<domain>/references/*.md` | 生成，源为 `skills/liki/references/` |
| `agents/<domain>-expert.md` | 生成，源为 `agents/personas/<domain>.md` |
| `.codebuddy-plugin/plugin.json` | 生成（版本、persona 元数据、connectors） |
| `.mcp.json` | 手写连接器契约 |
| `avatars/expert.png` | 手写占位资源，见下 |
| `README.md` | 手写包说明 |

生成物不要手改：`make sync-expert-packs` 覆盖并清理孤儿文件，`make check` 通过
`python3 scripts/sync_expert_packs.py --check` 做双向漂移检测。

## 头像

`avatars/expert.png` 是 512×512、16-bit RGBA 的项目自有占位图：不含任何第三方素材，
因此不附带额外许可义务。它不参与 `sync-expert-packs` 生成，也无需在包之间统一——
替换为正式漫画/插画风头像时，同样必须使用自有或已获授权的素材，并在包 `README.md`
记录来源。

## 校验

`python3 scripts/check_expert_docs.py`（`make check`）校验 persona、包内 SKILL 卡、
`plugin.json` 元数据、`.mcp.json` 连接器描述与运行时版本一致，并禁止这些描述硬编码 MCP 方法名。

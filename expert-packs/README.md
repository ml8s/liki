# Liki WorkBuddy 专家包

Liki 只有一个 WorkBuddy 专家：`liki`（Liki 命理师）。它覆盖八字、紫微、六爻、奇门、黄历择日、风水与起名；领域入口卡用于专家内路由，不表示多个专家。

## 单一权威源

- 根技能方法论：`skills/liki/references/`
- WorkBuddy 专家身份：`expert-packs/liki/SKILL.md`
- 部署身份：`agents/liki/liki.md`
- 部署配置：`agents/liki/agent.yaml`
- 发布 profile：`profiles/liki.json`

本包是 WorkBuddy 提交兼容层；生产部署使用 `agents/`、`profiles/liki.json` 生成的 `deployment.json`，并烘进 `liki-expert` 装配镜像。

## 包结构

| 路径 | 来源 |
| --- | --- |
| `SKILL.md` | 手写专家身份唯一源 |
| `expert-packs/liki/agents/liki.md` | `make sync-expert-packs` 从身份源投影 |
| `skills/liki/SKILL.md` | 从身份源投影的 ADK skill 入口 |
| `skills/liki/references/*.md` | 根技能 `skills/liki/references/*.md` 的全量镜像 |
| `.codebuddy-plugin/plugin.json` | 身份源与 `workbuddy/connector-ids.json` 投影 |
| `avatars/expert.svg` / `expert.png` | 品牌资源 |

`plugin.json` 依赖两个 WorkBuddy 连接器，包内没有 `.mcp.json`。

## 生成与校验

修改根技能后运行：

```sh
make sync-expert-packs
make check
```

`make check` 会执行：

- `scripts/sync_expert_packs.py --check`：检测方法论卡、身份投影和 plugin 元数据漂移；
- `scripts/check_expert_docs.py`：校验身份 frontmatter、连接器依赖与描述契约；
- 版本面测试：技能、专家、连接器版本同源。

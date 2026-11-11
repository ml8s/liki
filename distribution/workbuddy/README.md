# Liki WorkBuddy pro 渠道包

Liki 的 WorkBuddy pro 渠道包是 `liki-pro`（Liki 命理师）。它覆盖八字、紫微、六爻、奇门、黄历择日、风水与起名；领域入口卡用于专家内路由，不表示多个专家。

## 单一权威源

- 专家身份与公共 instruction：`skills/liki/SKILL.md`
- 根技能方法论：`skills/liki/references/`
- WorkBuddy 渠道配置：`distribution/workbuddy/liki-pro/channel.json`
- 付费行为 overlay：`distribution/workbuddy/liki-pro/overlays/payment.md`
- 部署配置：`distribution/liki-expert/deployment.yaml`
- 内部宿主协议：`distribution/liki-expert/host-adapter.md`
- 发布 profile：`distribution/liki-expert/profile.json`

本目录是 WorkBuddy pro 渠道适配层。zip 包由 `make build-channel-package` 在构建时从唯一源组装；生产部署使用 `distribution/liki-expert/` 配置生成的 `deployment.json`，并烘进 `liki-expert` 装配镜像。

## 包结构

| 路径 | 来源 |
| --- | --- |
| `skills/liki/SKILL.md` | 构建时从唯一身份 + 付费 overlay 投影的 ADK skill 入口 |
| `skills/liki/references/*.md` | 构建时从根技能 `skills/liki/references/*.md` 复制 |
| `.mcp.json` | 指向 liki-mcp 付费网关（`https://mcp.tokflux.com/mcp`） |
| `.codebuddy-plugin/plugin.json` | 渠道配置投影 |
| `avatars/expert.svg` / `expert.png` | 品牌资源 |

包自带 `.mcp.json` 指向付费网关；`plugin.json` 不声明 WorkBuddy 连接器依赖。

## 生成与校验

修改根技能后运行：

```sh
make build-channel-package
make check
```

`make check` 会执行：

- `scripts/build_channel_package.py --check`：检测方法论卡、身份投影和 plugin 元数据漂移；
- `scripts/check_expert_docs.py`：校验唯一身份、渠道配置、付费 overlay 与描述契约；
- 版本面测试：技能、专家、连接器版本同源。

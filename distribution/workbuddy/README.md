# Liki WorkBuddy 渠道包

Liki 的 WorkBuddy 渠道包分为 `liki-free`（灵机 / Liki 命理师）与 `liki-pro`（Liki 命理师（专业版））。两者都是专家包，覆盖八字、紫微、六爻、奇门、黄历择日、风水与起名；领域入口卡用于专家内路由，不表示多个专家。

## 单一权威源

- 专家身份与公共 instruction：`skills/liki/SKILL.md`
- 根技能方法论：`skills/liki/references/`
- WorkBuddy 渠道配置：`distribution/workbuddy/liki-free/channel.json` / `liki-pro/channel.json`
- WorkBuddy 市场元数据：`distribution/workbuddy/liki-free/skill-meta.json` / `liki-pro/skill-meta.json`
- 付费行为 overlay：`distribution/workbuddy/liki-pro/overlays/payment.md`
- 部署配置：`distribution/liki-expert/deployment.yaml`
- 内部宿主协议：`distribution/liki-expert/host-adapter.md`
- 发布 profile：`distribution/liki-expert/profile.json`

本目录是 WorkBuddy 渠道适配层（free / pro）。zip 包由 `make distribution-package` 在构建时从唯一源组装；生产部署使用 `distribution/liki-expert/` 配置生成的 `deployment.json`，并烘进 `liki-expert` 装配镜像。

## 包结构

| 路径 | 来源 |
| --- | --- |
| `skills/liki/SKILL.md` | 构建时从唯一身份投影；pro 追加付费 overlay |
| `skills/liki/references/*.md` | 构建时从根技能 `skills/liki/references/*.md` 复制 |
| `.mcp.json` | free 指向 engine + counsel；pro 指向 liki-mcp 付费网关 |
| `skill-meta.json` | WorkBuddy 差异字段：展示名、Agent 显示名、职称与 keywords |
| `.codebuddy-plugin/plugin.json` | 渠道静态配置；displayName / profession / quickPrompts 由构建器从 skill-meta 投影 |
| `avatars/expert.svg` / `expert.png` | 品牌资源；PNG 为 WorkBuddy 要求的 512×512 投影 |

free 的 `.mcp.json` 指向 engine + counsel，pro 指向付费网关；`plugin.json` 不声明 WorkBuddy 连接器依赖。

## 生成与校验

修改根技能后运行：

```sh
make distribution-package
make check-content
```

`make check-content` 会执行：

- `scripts/build_channel_package.py --check`：检测方法论卡、身份投影和 plugin 元数据漂移；
- `scripts/check_expert_docs.py`：校验唯一身份、渠道配置、付费 overlay 与描述契约；
- 版本面测试：技能、专家、连接器版本同源。

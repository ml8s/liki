# Liki 系统全况与本仓对齐

本文描述 Liki 单专家系统，以及本仓（领域与制品仓）的职责边界。

## 全局架构

```text
浏览器 → Caddy → liki-web (BFF)
                   └─AG-UI→ liki-agent-runtime（装配镜像）
                              ├─MCP→ engine-mcp（免费，直连）
                              └─MCP→ counsel-mcp（免费，直连）

外部 Skill → Caddy → liki-mcp（付费网关，SkillPay）
                         ├─→ engine-mcp
                         └─→ counsel-mcp
```

Liki 只有一个生产专家身份：`liki`；其 Expert assembly 产物是 `liki-expert`。`natal`、`divination`、`naming`、`fengshui` 是专家内部的领域路由和方法论组织，不是独立 Agent 或产品域。

| 仓 | 角色 | 产物 |
|---|---|---|
| liki | 领域、MCP 与制品仓 | engine / counsel MCP、专家身份、deployment 工件、装配镜像、根 skill |
| liki-agent-runtime | 通用运行时仓 | 纯运行时镜像、AgentDeployment 校验 |
| liki-web | 产品 BFF 仓 | BFF、前端和公共站点 |
| liki-deploy | 发布编排仓 | compose、env、Caddyfile 和部署流程 |

## 本仓制品

1. **专家身份**：`skills/liki/SKILL.md` 是唯一身份和公共 instruction 源；`distribution/liki-expert/deployment.yaml` 是部署配置；`distribution/liki-expert/host-adapter.md` 只追加内部宿主协议；`distribution/liki-expert/profile.json` 是唯一发布 profile。
2. **WorkBuddy 渠道配置**：`distribution/workbuddy/liki-pro/channel.json` 只承载平台投影字段；`distribution/workbuddy/liki-pro/overlays/payment.md` 只追加付费行为；`skills/liki/references/` 是方法论卡唯一源。
3. **MCP 服务**：`engine/` 与 `counsel/` 分别构建 `liki-engine` 和 `liki-counsel`。
4. **装配镜像**：`scripts/generate_deployment.py` 生成 `deployment.json` 和 instruction，装配镜像包含 deployment 与 `/skills`。
5. **分发产物**：`liki-free` / `liki-pro`；WorkBuddy 和 SkillHub 使用同一套方法论生成渠道包。

## 权威源与生成

| 事实 | 唯一源 |
|---|---|
| 方法论与路由卡 | `skills/liki/references/` |
| 专家身份与公共 instruction | `skills/liki/SKILL.md` |
| WorkBuddy 渠道投影字段 | `distribution/workbuddy/liki-pro/channel.json` |
| 部署 profile | `distribution/liki-expert/profile.json` |
| 分发变体 | `distribution/variants/free.json`, `distribution/variants/pro.json` |

`make build-channel-package` 不再生成第二份方法论源；它在构建 WorkBuddy 包时从唯一身份、唯一方法论卡、渠道配置和付费 overlay 组装包内容。`make check` 对漂移 fail-closed。

## 生成链

```text
skills/liki/SKILL.md ────────────┐
skills/liki/references/ ─────────┤
distribution/workbuddy/liki-pro/channel.json ──┼→ build_channel_package → dist/workbuddy/liki-pro.zip
distribution/workbuddy/liki-pro/overlays/ ─────┤
skills/liki/references/ ───────────────────────┘

skills/liki/SKILL.md ────────────┐
distribution/liki-expert/host-adapter.md ─────┤
distribution/liki-expert/deployment.yaml ──────────┼→ generate_deployment → deployment.json
distribution/liki-expert/profile.json ──────────────┘         ↓
liki-agent-runtime → liki-expert
```

装配镜像包含单专家 profile 和公共 skill 树：

```text
/deployment/liki-expert/deployment.json
/skills/liki/
```

## 变更传播

| 变更 | 后续 |
|---|---|
| 根方法论卡 / 唯一身份 | `make distribution-package` 构建渠道包；重建装配镜像 |
| 工具白名单 / 内部宿主协议 | 重新生成 deployment；重建并重发 `liki-expert` |
| engine / counsel 工具契约 | 对齐 `distribution/liki-expert/deployment.yaml`；重建并重发 |
| deployment contract | 先升级 runtime 校验，再更新生成器并重新 validate |

## CI 与发布

`ci.yml` 执行 lint、结构契约、文档检查和测试。镜像按 commit SHA 本地构建；`liki-deploy` 负责装配测试、发布和 smoke。

## 相关文档

- [Domain naming matrix](DOMAIN_NAMING.md)
- [Domain language matrix](DOMAIN_LANGUAGE.md)
- [MCP contract design](MCP_CONTRACT_DESIGN.md)
- [Digest policy](DIGEST_POLICY.md)

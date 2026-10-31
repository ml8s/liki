# Liki 系统全况与本仓对齐

本文描述 Liki 单专家系统，以及本仓（领域与制品仓）的职责边界。

## 全局架构

```text
浏览器 → Caddy → liki-web (BFF)
                   └─AG-UI→ liki-agent-runtime（装配镜像）
                              └─MCP→ engine-mcp / counsel-mcp
```

Liki 只有一个生产专家：`liki`。`natal`、`divination`、`naming`、`fengshui` 是专家内部的领域路由和方法论组织，不是独立 Agent 或产品域。

| 仓 | 角色 | 产物 |
|---|---|---|
| liki | 领域、MCP 与制品仓 | engine / counsel MCP、专家身份、deployment 工件、装配镜像、根 skill |
| liki-agent-runtime | 通用运行时仓 | 纯运行时镜像、AgentDeployment 校验 |
| liki-web | 产品 BFF 仓 | BFF、前端和公共站点 |
| liki-deploy | 发布编排仓 | compose、env、Caddyfile 和部署流程 |

## 本仓制品

1. **专家定义**：`agents/liki/liki.md` 是部署身份；`agents/liki/agent.yaml` 是部署配置；`profiles/liki.json` 是唯一发布 profile。
2. **WorkBuddy 身份**：`expert-packs/liki/SKILL.md` 是 WorkBuddy 专家身份唯一手写源；`skills/liki/references/` 是方法论卡唯一源。
3. **MCP 服务**：`engine/` 与 `counsel/` 分别构建 `liki-engine` 和 `liki-counsel`。
4. **装配镜像**：`scripts/generate_deployment.py` 生成 `deployment.json`，装配镜像只包含 `deployment/liki/deployment.json`。
5. **WorkBuddy 交付物**：两个连接器、一个根技能、一个专家包；见 `workbuddy/README.md`。

## 权威源与生成

| 事实 | 唯一源 |
|---|---|
| 方法论与路由卡 | `skills/liki/references/` |
| 生产 agent 身份 | `agents/liki/liki.md` |
| WorkBuddy 专家身份 | `expert-packs/liki/SKILL.md` |
| 部署 profile | `profiles/liki.json` |
| 连接器 ID | `workbuddy/connector-ids.json` |

`make sync-expert-packs` 把根技能方法论卡镜像到 WorkBuddy 包，并从 WorkBuddy 身份源投影包内入口和 plugin 元数据。`make check` 对漂移 fail-closed。

## 生成链

```text
skills/liki/references/ ─────────┐
expert-packs/liki/SKILL.md ──────┼→ sync_expert_packs → expert-packs/liki
agents/liki/liki.md ─────────────┤
agents/liki/agent.yaml ──────────┼→ generate_deployment → deployment.json
profiles/liki.json ──────────────┘         ↓
liki-agent-runtime → liki-expert
```

装配镜像只包含单专家 profile：

```text
/deployment/liki/deployment.json
```

运行时进程加载该 profile。一个产品、一个专家、一个部署单元。

## 变更传播

| 变更 | 后续 |
|---|---|
| 根方法论卡 | `make sync-expert-packs` 更新 WorkBuddy 包；重建装配镜像 |
| 专家身份 / 工具白名单 | 重新生成 deployment；重建并重发 `liki-expert` |
| engine / counsel 工具契约 | 对齐 `agents/liki/agent.yaml`；重建并重发 |
| deployment contract | 先升级 runtime 校验，再更新生成器并重新 validate |

## CI 与发布

`ci.yml` 执行 lint、结构契约、文档检查和测试。镜像按 commit SHA 本地构建；`liki-deploy` 负责装配测试、发布和 smoke。

## 相关文档

- [Domain naming matrix](DOMAIN_NAMING.md)
- [Domain language matrix](DOMAIN_LANGUAGE.md)
- [MCP contract design](MCP_CONTRACT_DESIGN.md)

# Liki 系统全况与本仓对齐

> 本文档描述 Liki 多 Agent 系统的全局架构，以及本仓（内容/领域仓）在其中
> 的角色与对齐点。让本仓负责人知道：我产什么、谁消费、契约在哪、变更如何传播。

## 一、系统全局架构

Liki 是多 Agent 命理专家系统，四仓分工：

```text
浏览器 → Caddy → liki-web (BFF)
                   └─AG-UI→ liki-agents (装配镜像: 运行时+组合工件) ──MCP──> engine-mcp / counsel-mcp
                              router → 全功能领域专家委派，审计留痕
```

| 仓 | 角色 | 产物 |
|---|---|---|
| **liki（本仓）** | 内容/领域仓 | 专家内容、engine/counsel MCP、**deployment 工件**、**装配镜像**、web skill bundle |
| liki-agents | 通用运行时仓 | 纯运行时镜像、AgentDeployment schema 契约 |
| liki-web | 产品 BFF 仓 | liki-web 镜像（BFF + 前端） |
| liki-deploy | 发布编排仓 | 纯清单：compose + env + Caddyfile，只 pin 版本 |

## 二、本仓角色（liki 内容仓）

本仓是系统**内容与确定性的源头**，产出四类制品：

1. **专家定义**：`agents/<name>/`（唯一 AgentDeployment source of truth；`skills/liki-*` 中的专家包仅是外部客户端兼容资产）
2. **Engine / Counsel MCP**：`engine/`、`counsel/` 源码 → GHCR 镜像 `liki-engine`、`liki-counsel`
3. **AgentDeployment 工件**：由 `agents/` 发布结构 + `scripts/generate_deployment.py` 生成 → 烘焙进装配镜像
4. **Web skill bundle**：`scripts/build-web-skill-bundle.sh` → release asset，供 liki-web 消费

## 三、关键对齐点（负责人必读）

### 3.0 方法论权威源（最高优先规则）

**方法论卡的唯一权威源是根 `skills/liki`（各域 `domains/` 下）**。独立专家包
（`liki-bazi` 等）是根 skill 的对外发布形态，其方法论卡由
`scripts/sync-expert-methodology.sh` 从根同步，**不要手改专家包的方法论卡**。

- **优化 skill 只改根 liki**，再运行同步脚本 + 校验。
- `make check` 会跑 `scripts/check-expert-methodology.sh` 校验根与专家包一致。
- 手工双份会漂移：改专家包不被根感知，改根不同步专家包都会破坏一致性校验。

### 3.1 专家定义结构（`agents/<name>/`）

每个专家一个目录：

```text
agents/router/agent.yaml        # name/description/mode/sub_agents/tools.allow
agents/router/instruction.md    # 指令骨架（方法论卡构建时自动追加）
agents/bazi/agent.yaml
agents/bazi/instruction.md
agents/bazi/output-schema.json  # 可选：结构化输出
```

- `agent.yaml` 字段与 **ADK** 对齐（mode: chat/task/single_turn；sub_agents；tools.allow）
- **`tools.allow` 是权威工具白名单**（替代旧的 env 白名单），每笔 engine/counsel 工具调用必须允许
- `instruction.md` 构建时自动合并对应领域的方法论卡（除 SKILL.md），要求**自包含**（<2MB）

### 3.1.1 MCP 连接器模型（客户端分域）

客户端（CodeBuddy 等）通过 `.mcp.json` 连接**分域端点**，每个专家只见自己领域工具。engine 分域 = 一个术数或共享 aux；counsel 分域 = 一个判断域。

| 专家 | engine 连接器 | counsel 连接器 |
|---|---|---|
| bazi | `engine-bazi` → `/engine/mcp/bazi` | `counsel-bazi` → `/counsel/mcp/bazi` |
| ziwei | `engine-ziwei` → `/engine/mcp/ziwei` | `counsel-ziwei` → `/counsel/mcp/ziwei` |
| liuyao | `engine-liuyao` → `/engine/mcp/liuyao` | `counsel-liuyao` → `/counsel/mcp/liuyao` |
| qimen | `engine-qimen` → `/engine/mcp/qimen` **+** `engine-huangli` → `/engine/mcp/huangli` | `counsel-qimen` → `/counsel/mcp/qimen` |
| fengshui | `engine-fengshui` → `/engine/mcp/fengshui` | — |
| naming | `engine-bazi` → `/engine/mcp/bazi`（取用神） | `counsel-naming` → `/counsel/mcp/naming` |

要点：

- **黄历是 engine 独立 `huangli` 子域**（`huangli.*` 前缀），不是 aux——aux 是 `time./tianwen./city.`（`/engine/mcp/aux`）。
- **奇门连双 engine 连接器**（`engine-qimen` + `engine-huangli`）：分域端点下 `engine-qimen` 不含黄历，择日额外连 `engine-huangli`。
- **命名连 `engine-bazi`**：起名用神取字依赖八字排盘，故复用 bazi 排盘连接器（排盘能力，非专家协作）。
- 运行时（装配镜像）agent 连**全量 `/mcp`**（compose `LIKI_MCP_ENGINE_URL`），工具可见性由 `tools.allow` 白名单控制；分域连接器仅用于客户端安装独立 skill 时的隔离。

### 3.2 生成与校验链（`make build-deployment`）

```text
profiles/experts.json ─┐
agents/*/agent.yaml ─────┴→ scripts/generate_deployment.py → dist/agents/experts/
                                      ↓
contracts/agent-definition.version → scripts/check_deployment_schema.py（schema 校验）
```

- **schema 契约由 liki-agents 拥有**：本仓通过 `contracts/agent-definition.version`（version+digest）pin 具体 schema，生成器校验对齐，避免漂移
- `make build-deployment` 生成唯一全功能 experts 工件并校验，已接入 `make check` 和 CI
- 本仓 `contracts/agent-definition.schema.json` 是从 liki-agents 复制的 schema 快照（CI 用其 digest 校验）

### 3.3 装配镜像（liki-experts）

`assembly/Dockerfile`：

```dockerfile
ARG BASE_IMAGE=ghcr.io/ml8s/liki-agents:<base>   # liki-agents release 显式 pin
FROM ${BASE_IMAGE}
COPY dist/agents/${PROFILE}/ /deployment/        # 组合工件烘焙进镜像
ENV LIKI_AGENTS_DEPLOYMENT_FILE=/deployment/deployment.json
ENV LIKI_AGENTS_DEPLOYMENT_DIGEST=<digest>       # CI 计算，生产启动校验
```

- **组合在 liki**：专家拓扑/指令/白名单在 CI 生成并烘焙，不在运行时装配
- 发布时序：**先 liki-agents release，后 liki release**（装配镜像 FROM 依赖）
- 镜像不含密钥/源码：端点/令牌由部署环境注入

### 3.4 Web skill bundle（liki-web 消费）

`make build-web-skill-bundle` 产出 `dist/liki-web-skill-bundle.tar.gz`（skills/liki 树 + webapp + liki.tar.gz + index.json），release 时上传为 asset。liki-web 的 `prepare-skills` action 下载解包，不再 checkout 本仓源码。

## 四、变更如何传播（负责人须知）

| 本仓变更 | 影响 | 传播路径 |
|---|---|---|
| 改专家 instruction/方法论卡 | 装配镜像内容变 | 重新生成工件 → 重发装配镜像 → deploy 换 tag |
| 改 `tools.allow` 白名单 | 专家可用工具变 | 同上 |
| 新增专家 | profile/工件/装配镜像 | 加 `agents/<name>/` → profile 加 agent → 重发 |
| 改 engine/counsel 工具语义 | 白名单需对齐 | 更新 agents/*/agent.yaml → 重发 |
| schema 契约升级 | 生成器/工件格式 | 更新 `contracts/agent-definition.version` → 重新生成校验 |

## 五、本仓 CI 发布链（release 触发）

- `check`：lint + schema + docs + **build-deployment**（工件生成校验）
- `docker-publish`：push `liki-engine`
- `publish-experts`：生成唯一全功能工件 → 计算 digest → 构建并 push `liki-experts`
- `publish-web-skill-bundle`：上传 bundle 到 release asset

## 六、相关文档

- `liki-agents/docs/ARCHITECTURE.md`：运行时与 AgentDeployment 契约详解
- `liki-deploy/docs/ARCHITECTURE.md`：系统落地计划与发布时序
- `liki-deploy/docs/DEPLOYMENT.md`：生产发布流程

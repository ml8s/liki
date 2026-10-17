# Liki 系统全况与本仓对齐

> 本文档描述 Liki 多 Agent 系统的全局架构，以及本仓（内容/领域仓）在其中
> 的角色与对齐点。让本仓负责人知道：我产什么、谁消费、边界在哪、变更如何传播。

## 一、系统全局架构

Liki 是多 Agent 命理专家系统，四仓分工：

```text
浏览器 → Caddy → liki-web (BFF)
                   └─AG-UI→ liki-agent-runtime (装配镜像: 运行时+组合工件) ──MCP──> engine-mcp / counsel-mcp
                              router → 全功能领域专家委派，审计留痕
```

| 仓 | 角色 | 产物 |
|---|---|---|
| **liki（本仓）** | 内容/领域仓 | 专家内容、engine/counsel MCP、**deployment 工件**、**装配镜像**、web skill bundle |
| liki-agent-runtime | 通用运行时仓 | 纯运行时镜像、AgentDeployment 运行时校验 |
| liki-web | 产品 BFF 仓 | liki-web 镜像（BFF + 前端） |
| liki-deploy | 发布编排仓 | 纯清单：compose + env + Caddyfile，只 pin 版本 |

## 二、本仓角色（liki 内容仓）

本仓是系统**内容与确定性的源头**，产出五类制品：

1. **专家定义**：`agents/<name>/agent.yaml`（部署配置）+ 身份文件 `expert-packs/liki-<domain>/SKILL.md`（角色/指令骨架唯一手写源，内部部署与市场包共用）；`expert-packs/` 同时是 CodeBuddy/WorkBuddy 市场包兼容层（上架 + multi-expert 镜像）
2. **Engine / Counsel MCP**：`engine/`、`counsel/` 源码 → GHCR 镜像 `liki-engine`、`liki-counsel`
3. **AgentDeployment 工件**：由身份文件 + `agents/` 发布结构 + `profiles/` 经 `scripts/generate_deployment.py` 生成 → 烘焙进装配镜像
4. **Web skill bundle**：`scripts/build-web-skill-bundle.sh` → release asset，供 liki-web 消费
5. **WorkBuddy 上架交付物**：`workbuddy/` 下的连接器提交包（`connector-meta.json` + `mcp.json` + `icon.svg`；一个连接器只声明一个 MCP Server，免鉴权故不写 `auth_mode`），技能与专家包分别是 `skills/liki/` 与 `expert-packs/liki-*`；四类上架物与提交流程见 `workbuddy/README.md`，提交前检查由 `python3 scripts/check_workbuddy_connectors.py` 守卫

## 三、关键对齐点（负责人必读）

### 3.0 权威源（最高优先规则）

**方法论卡的唯一权威源是根 `skills/liki/references/`（各域 `domains/` 下）；专家身份的唯一手写源是包根 `expert-packs/liki-<domain>/SKILL.md`**。
独立专家包（`expert-packs/liki-*`）是根 skill 的对外发布形态，其方法论卡与包内 `SKILL.md` /
`plugin.json` 元数据由 `make sync-expert-packs` 生成并由 CI 漂移检测守护，
**不要手改专家包的方法论卡与生成物**。

- **优化 skill 只改根 liki，改人设只改身份文件**，再运行 `make sync-expert-packs` + 校验。
- `make check` 会跑 `python3 scripts/sync_expert_packs.py --check` 校验权威源与专家包一致。
- 手工双份会漂移：改生成物不被源感知，改源不同步专家包都会破坏一致性校验。

### 3.1 专家定义结构

每个专家由两层组成：**身份文件**（角色骨架，唯一手写源）＋ **agent.yaml**（部署配置）。

```text
expert-packs/liki-natal/SKILL.md    # 身份单一源：frontmatter 元数据 + 角色/能力/边界正文（市场入口）
agents/natal/agent.yaml             # 部署配置：name/mode/sub_agents/tools.allow/skills.root
agents/natal/output-schema.json     # 可选：结构化输出
agents/router/router.md             # router 例外：无专家包，身份文件就在其部署目录
```

**专家与顶层领域一一对应**（`natal` / `divination` / `fengshui` / `naming`，与根
`skills/liki/references/<domain>/` 同名）：包内 `skills/<domain>/references/` 是根
`references/<domain>/` 整棵子树（ENTRY/TOOLS/app/domains）的镜像，领域卡只写本体系知识、
`app/` 承载场景与跨体系合参，因此不做多对一聚合、也不产生跨包死链。

- `agent.yaml` 字段与 **ADK** 对齐（mode: chat/task/single_turn；sub_agents；tools.allow）
- **`tools.allow` 是权威工具白名单**（替代旧的 env 白名单），每笔 engine/counsel 工具调用必须允许
- 身份文件 frontmatter 映射到各消费方：`description_en` → 部署 agent 描述 + `plugin.json.description`；
  `description` → 域 `skills/<domain>/SKILL.md` 描述；`display_name(_en)` → 域 skill `metadata`；
  `profession` → `plugin.json.profession`；`keywords` → `plugin.json.keywords`（市场检索词）；
  `skills`（skill 名列表）→ `plugin.json.skills`（预加载路径 `./skills/<name>`）；`maxTurns` /
  `displayName` 是 WorkBuddy agent md 字段（`plugin.json.agents` 指向生成物
  `agents/<pack>.md`，它由身份文件投影、平台字段名用 camelCase）；`version` 单一源为
  `skills/liki/VERSION.txt`（agent.yaml 不再携带）
- 身份正文＝部署 instruction 骨架，同时是包内域 `SKILL.md` 的正文：两处同源渲染，包内不再有
  第二份身份副本，**杜绝漂移**（`expert-packs/README.md`：包结构来源表）
- 域 `SKILL.md` 的方法论卡索引由生成器按镜像卡自动生成，禁止手维护卡名清单
- 方法论卡不入身份文件：由 ADK skilltoolset 按需 `load_skill` 渐进加载（避免每轮 prompt 膨胀）

### 3.1.1 MCP 连接器模型（专家依赖已上架连接器）

WorkBuddy 专家不声明自带 MCP：四个 `plugin.json` 只有 `dependencies.connectors`（两个聚合连接器的
平台 ID，单一源 `workbuddy/connector-ids.json`），包内不带 `.mcp.json`。用户在召唤专家前按连接器
自身的引导卡片完成一次连接，四个专家共用；引导卡片的名称、描述、图标由连接器提供。

| 专家 | 依赖的连接器 | 工具面 |
|---|---|---|
| natal | 排盘引擎 `liki-engine` + 判断层 `liki-counsel` | 全领域（领域边界由身份正文「硬边界」承担） |
| divination | 同上 | 同上 |
| fengshui | 同上 | 同上 |
| naming | 同上 | 同上 |

域级端点（`/engine/mcp/{bazi,ziwei,liuyao,qimen,huangli,fengshui}`、`/counsel/mcp/{bazi,ziwei,liuyao,qimen,naming}`）
仍由服务暴露，供自建分域部署使用，不参与 WorkBuddy 上架。

要点：

- **黄历是 engine 独立 `huangli` 子域**（`huangli.*` 前缀），不是 aux——aux 是 `time./tianwen./city.`（`/engine/mcp/aux`）；聚合连接器下择日工具随排盘引擎一并可见，占卜专家因此不必额外连第三个服务。
- 专家工具面变宽后，**领域边界由身份正文「硬边界」给出**（如本命专家：六爻、奇门、黄历择日、起名、风水属其他专家），工具名本身带领域前缀。
- 运行时（装配镜像）agent 连**全量 `/mcp`**（compose `LIKI_MCP_ENGINE_URL`），工具可见性由 `tools.allow` 白名单控制（`agent.yaml`）；WorkBuddy 侧没有白名单机制，靠上面的「硬边界」。

### 3.2 生成与校验链（`make build-deployment`）

```text
expert-packs/liki-*/SKILL.md ───────────┐   （域专家身份文件）
agents/router/router.md ────────────────┤   （router 身份文件）
agents/*/agent.yaml ────────────────────┼→ scripts/generate_deployment.py → dist/agents/experts/
profiles/experts.json ──────────────────┘                ↓
deps/liki-agent-runtime.base → pinned liki-agent-runtime runtime（validate）
```

- **运行时 base 由本仓 pin**：`deps/liki-agent-runtime.base`（version+immutable digest）锁定 liki-agent-runtime 发布版本；测试（`make runtime-validate`）与部署（CI 装配 `FROM`）共用该 pin，**不与 liki 版本同名对齐**（各仓独立发版）
- `make build-deployment` 只生成工件；`make runtime-validate` 用 pinned liki-agent-runtime runtime 执行真实校验
- `make build-release-manifest` 只打包当前 `dist` 工件，不隐式运行测试
- `make gate` 与 CI `release-integrity` 都先执行 `runtime-validate`，再构建 release manifest；发布工件不能绕过 pinned runtime validate

`assembly/Dockerfile` 按 `PROFILE` 参数构建两套装配镜像：

```dockerfile
ARG BASE_IMAGE=ghcr.io/ml8s/liki-agent-runtime:<version>@<digest>  # 来源 deps/liki-agent-runtime.base
FROM ${BASE_IMAGE}
COPY dist/agents/${PROFILE}/ /deployment/        # 组合工件烘焙进镜像
ENV LIKI_AGENT_RUNTIME_DEPLOYMENT_FILE=/deployment/deployment.json
```

- `PROFILE=experts` → `liki-experts`（全功能，含全部子专家）
- AgentDeployment digest 只作为审计 provenance；部署完整性由 final assembly image digest 承担
- **组合在 liki**：专家拓扑/指令/白名单在 CI 生成并烘焙，不在运行时装配
- **依赖模型**：liki 通过 `deps/liki-agent-runtime.base` 锁定 liki-agent-runtime 版本；测试与部署共用该 pin。**liki-agent-runtime 未发新版则 pin 不变**（两者独立发版，不要求版本同名）；采用新 liki-agent-runtime 时先更新 pin（`python3 scripts/agent_runtime_pin.py --check` 在 `make check` 中校验）
- 镜像不含密钥/源码：端点/令牌由部署环境注入

### 3.4 Web skill bundle（liki-web 消费）

`make build-web-skill-bundle` 产出 `dist/liki-web-skill-bundle.tar.gz`（skills/liki 树 + liki.tar.gz + index.json），release 时上传为 asset。liki-web 的 `prepare-skills` action 下载解包，不再 checkout 本仓源码。

## 四、变更如何传播（负责人须知）

| 本仓变更 | 影响 | 传播路径 |
|---|---|---|
| 改身份文件（人设/指令骨架）或方法论卡 | 装配镜像内容变 | 重新生成工件 → 重发装配镜像 → 更新 signed digest lock |
| 改 `tools.allow` 白名单 | 专家可用工具变 | 同上 |
| 新增专家 | profile/工件/装配镜像 | 加 `agents/<name>/` → profile 加 agent → 重发 |
| 改 engine/counsel 工具语义 | 白名单需对齐 | 更新 agents/*/agent.yaml → 重发 |
| 工件格式变化 | 生成器/装配镜像 | 先发布 liki-agent-runtime runtime → 更新 base pin → 重新生成并 validate |

## 五、本仓 CI 发布链（release 触发）

- `check`：lint + 数据结构 + docs（不生成工件）
- `test`：仓库/引擎/counsel 测试（不做跨仓 runtime validate）
- `gate`：追加 160 题覆盖、pinned runtime validate、发布工件与镜像 smoke
- `docker-publish`：push `liki-engine`
- `publish-web-skill-bundle`：上传 bundle 到 release asset

## 六、相关文档

- `liki-agent-runtime/docs/ARCHITECTURE.md`：运行时与 AgentDeployment 工件详解
- `liki-deploy/docs/ARCHITECTURE.md`：系统架构与发布时序
- `liki-deploy/docs/OPERATIONS.md`：生产发布流程

- [Domain naming matrix](DOMAIN_NAMING.md): canonical repository/image/service names.

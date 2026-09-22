# WorkBuddy 接入计划：独立 MCP + Skill/Connector/Expert + RPC 退役

> 状态：已确认关键决策，待实施（2026-09-21）
> 目标：让 Liki 完整支持 WorkBuddy 开放平台体系（Skill / Connector / Expert），引擎对外协议收敛为标准的独立 MCP Server，私有 JSON-RPC 具备平滑退役路径。

## 已确认决策（2026-09-21）

| 编号 | 决策 | 结论 |
| --- | --- | --- |
| D1 | MCP 库 | 官方 `modelcontextprotocol/go-sdk` |
| D2 | Skill Python 工具层 | **先双轨后迁移**：保留 npx Skill 走 `/jsonrpc`，另发 WorkBuddy 配套 Skill；后续再把 Python 工具层改为 MCP client，最后退役 RPC |
| D3 | Expert | **一次做完（本期）**：本期范围 = 阶段一（独立 MCP）+ 阶段三（WorkBuddy 上架 Skill + Connector）+ 阶段四（专家上架） |

本期明确不做：专家团、Buddy 应用、RPC 退役（阶段五，后续触发）。

## 1. 背景与目标

WorkBuddy 开放平台首期开放五类生态能力：Buddy 应用、专家、专家团、连接器、外部应用接入。其中与 Liki 直接相关的组合形态是：

- **Skill**：教 AI「怎么做」——Liki 已经是 `SKILL.md` 形态；
- **连接器（Connector）**：让 AI 接触外部系统——Liki 的 Go 引擎是典型的外部计算系统，应以**独立 MCP Server** 暴露，官方推荐「MCP + Skill」方案；
- **专家（Expert）**：以专业角色被召唤——可在连接器之上包装「命理师」人设，声明依赖连接器 / Skill。

目标：

1. 提供**标准、完全独立**的 MCP Server，与现有私有 JSON-RPC（`/jsonrpc`）解耦；
2. 以 WorkBuddy「连接器（MCP + Skill）」形式上架；
3. 保留现有 `npx skills` 分发的 Skill 体验，同时具备 **RPC 退役路径**；
4. 包装 Expert「命理师」，让 Liki 以专业角色出现在专家中心（本期）。

## 2. 现状盘点

### 2.1 引擎（engine/，Go 1.26.2，module `liki-engine`）

- HTTP JSON-RPC 2.0 服务，唯一端点 `POST /jsonrpc`（`engine/internal/http/rpc.go`）；
- 方法注册内核：`RPCRegistry`（`engine/internal/agent/rpc_registry.go`），每个方法 = `Name / Description / Params(schema) / Result(schema) / Handler`；
- 共约 30 个方法，8 个域：`bazi.*`、`ziwei.*`、`bazhai.*`、`liuyao.*`、`qimen.*`、`xuankong.*`、`huangli.days`、`qiming.*`，另有 `tianwen.time` / `city.coords` / `time.now` / `rpc.discover`；
- 已有 OpenRPC 1.4.1 文档生成（`openrpc.go`），参数校验用 `santhosh-tekuri/jsonschema`；
- 启动入口：`engine/cmd/liki/main.go`，监听 `:8080`，含 `/jsonrpc`、`/health`、`/version`、限流、CORS、BodyLimit、Recover 中间件。

### 2.2 Skill（skills/liki/）

- 入口 `SKILL.md`，调用链 `SKILL.md → ENTRY.md → app 卡 → Python 工具或固定 RPC`；
- `natal/`、`divination/` 走 Python 工具层：`skills/liki/natal/tools/paipan.py`、`skills/liki/divination/tools/divination_rpc.py`，通过 `LIKI_RPC_URL`（默认 `https://liki.hk/jsonrpc`）POST JSON-RPC，并用 `rpc.discover` 做版本/方法校验；
- `naming/`、`fengshui/` 无 Python 工具层，直接发固定 JSON-RPC 报文。

### 2.3 与 WorkBuddy 的差距

| WorkBuddy 能力 | Liki 现状 | 差距 |
| --- | --- | --- |
| Skill | 已有完整 SKILL.md 结构 | 需按 WorkBuddy Skill 上架规范核对字段/格式；另出连接器配套 Skill |
| 连接器（MCP + Skill） | 无 | 需独立 MCP Server + `connector-meta.json` / `mcp.json` + 配套调用 Skill |
| 专家 | 无 | 需 `plugin.json` + `agents/*.md` 命理师人设 + avatars，声明依赖连接器（本期做） |
| 专家团 / Buddy 应用 | 无 | 非本期目标（Buddy 应用仅企业认证） |

## 3. 目标架构

```text
WorkBuddy 用户
  ├─ Skill 指令层：SKILL.md（路由 / 安全 / 流程 / 反馈）
  ├─ Connector：MCP client ──HTTPS──▶ https://liki.hk/mcp（独立 MCP Server）
  └─ Expert「命理师」：plugin.json + agents/*.md，声明依赖 Liki Connector + 预加载配套 Skill
                      │
      liki-mcp（新独立二进制，Go）
          │   import engine/internal（RPCRegistry + 领域引擎，零复制）
          └── Streamable HTTP（推荐，WorkBuddy 要求） + stdio（本地调试）
```

关键点：**独立 MCP Server 复用 `engine/internal` 内核，但不依赖 `/jsonrpc` 传输层**。两个入口共享同一个 `RPCRegistry` 与领域引擎，`/jsonrpc` 可随时下线而不影响 MCP 用户。

## 4. 独立 MCP Server 设计

### 4.1 形态

- 新二进制 `engine/cmd/liki-mcp/`（与 `engine/cmd/liki/` 并列），独立监听端口，默认 `:8081`；
- 仅 `import` 现有 `engine/internal/agent`、`engine/internal/engine/*`，不复制业务逻辑；
- 端点为标准 MCP：
  - `Streamable HTTP`（WorkBuddy 连接器硬要求）挂 `/mcp`；
  - `stdio` 便于本地调试与 `mcp-spec-check`；
  - 复用引擎的中间件（限流、CORS、BodyLimit、Recover、SecurityHeaders），新增会话无关约束。
- 版本与身份：`Implementation{Name: "liki-mcp", Version: <BuildTime>}`，与引擎 `VERSION` 契约保持一致。

### 4.2 Go MCP 库选型（已调研）

| 库 | 维护 | 传输 | 备注 |
| --- | --- | --- | --- |
| `modelcontextprotocol/go-sdk`（推荐） | 官方 + Google | stdio / command；HTTP 用 `NewStreamableHTTPHandler` / `NewSSEHandler` | 支持到 2026-07-28 spec，泛型 `AddTool` + struct tag 生成 schema，自带 jsonschema 校验 |
| `mark3labs/mcp-go` | 社区（8.8k★） | stdio / SSE / Streamable HTTP / 进程内 | 实现 2025-11-25 spec，上手快；缺 SEP-2575（stateless），`mcp-spec-check` 最新项过不了 |

**已确认 D1**：采用官方 `go-sdk`。实施时验证 go 1.26.2 兼容性；如遇阻塞，回退 `mcp-go` 并在此更新。

### 4.3 工具映射

原则：`RPCRegistry` 的每个 method → 一个 MCP tool，**逐个方法映射**（schema 精确，AI 选工具更准），不做全量网关 tool（兜底可另议）。

- tool name = method name（如 `bazi.chart`）；
- description = 现有 `RPCMethod.Description`（已面向 AI agent 撰写）；
- input schema = 现有 `Params` JSON Schema：
  - `go-sdk` 泛型 `AddTool[In, Out]` 用 Go struct tag 推断 schema；现有 schema 是 JSON 字符串，需转换——两个选择：
    - 低层 `Server.AddTool(&mcp.Tool{InputSchema: jsonschema...})`，把现有 JSON Schema 解析进 `jsonschema.Schema`（**推荐，零改写 handler 参数契约**）；
    - 或为每个方法写输入 struct（改动大，不推荐一期做）；
- handler：`reg.Execute(ctx, method, params)`，结果 → `&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(result)}}}`；
- 错误映射：`agent.RPCError` → `IsError: true` 的结果（不要泄露内部细节）；协议级错误用 `jsonrpc.Error`。

### 4.4 `rpc.discover` 的替代

MCP 原生 `tools/list` 已覆盖；不再暴露 `rpc.discover`。版本/能力自检迁移到 `initialize` 的 `serverInfo` 与 `tools/list`。

### 4.5 Resources / Prompts（可选）

- 可将 `FAQ.md`、领域知识摘要注册为 resources，本期可不做；
- Prompts（快捷提示词）与 Expert 的 `quickPrompts` 对应，做 Expert 时再引入。

## 5. Skill 侧改造（Connector 配套）

WorkBuddy 连接器 = MCP Server + 配套 Skill（教 AI 如何正确调用工具，补充前置条件、参数选择、高风险确认与错误恢复）。

现有 `SKILL.md` 的分发与调用现状：

- `npx skills add ml8s/liki` 分发的 Skill 直接 POST `/jsonrpc`（Python 工具层）；
- WorkBuddy 上架时，Skill 内容定位为「如何调用连接器工具完成命理任务」。

**已确认 D2（双轨，先不退役）**：保留现有 Skill 与 `/jsonrpc` 不动，另发 WorkBuddy 版「连接器配套 Skill」，两者并存。`/jsonrpc` 继续为 npx 客户端服务。RPC 退役触发条件：全部消费者（含 Python 工具层）迁到 MCP 并度过观测期后，见第 7 节阶段 5。

## 6. WorkBuddy 上架三件套

### 6.1 连接器（Connector）

WorkBuddy 要求（来自官方文档要点，实施时逐项核对）：

- 方案二选一：MCP + Skill 或 CLI + Skill，一个连接器只能选一种 → **选 MCP + Skill**；
- 一个连接器只对应一个 MCP Server；
- 硬要求：HTTPS、SSE 或 `streamableHttp`、工具 30 秒内响应；
- 文件：`connector-meta.json`（注册/展示）、`mcp.json`（连接配置，Streamable HTTP）、配套 Skill；
- 凭证不入包、最小权限。Liki 引擎无状态，无需 OAuth 场景，但需确认 WorkBuddy 对无鉴权 MCP 的接入约定。

### 6.2 Skill

- 按 WorkBuddy Skill 规范核对现有 `SKILL.md`（目录、入口、字段格式、标签、快捷提示词等）；
- 内容改为「调用连接器工具」，保留现有领域流程、安全、反馈契约。

### 6.3 专家「命理师」（本期）

依据 WorkBuddy 官方专家规范（`/docs/expert`），专家是一个插件包：

```text
workbuddy/liki-expert/
├── .codebuddy-plugin/
│   └── plugin.json
├── avatars/
│   └── expert.png          # 512×512 PNG ≤500KB，统一漫画/插画风格命理师头像
├── agents/
│   └── liki-master.md      # Agent 定义：frontmatter + 系统提示词
└── README.md
```

#### 6.3.1 plugin.json 要点（字段按官方规范逐项填写）

| 字段 | 值（草案） | 说明 |
| --- | --- | --- |
| name | `liki-master` | 小写+连字符，专家唯一标识 |
| expertType | `agent` | 单人专家；不做 team |
| version | SemVer | 与发行版版本契约对应（D6 确认映射） |
| description | 英文简短描述 | |
| author | {name, email} | 仓库作者信息 |
| agents | `["./agents/liki-master.md"]` | |
| agentName | `liki-master` | 与 MD 文件名一致 |
| displayName | {en, zh} | zh 如「Liki 命理师」 |
| profession | {en, zh} | zh 如「命理顾问」 |
| displayDescription | {en, zh} | **zh 硬约束 40-50 字**，突出核心能力（八字/紫微/六爻/奇门/风水/起名 + 可回溯依据） |
| avatar | `avatars/expert.png` | |
| categoryId | `12-IndustryConsultant`（**已确认 D8**） | 15 个行业分类选一 |
| defaultInitPrompt | {en, zh} | 必须与 quickPrompts 第一条一致 |
| plugin | `liki-master` | 与 name 一致 |
| tags | {en, zh}[] | **固定 3 个**，如 八字/风水/起名（zh），Bazi/Feng Shui/Naming（en） |
| quickPrompts | {en, zh}[] | **固定 3 个**，第一条 = defaultInitPrompt |
| skills | 预加载 WorkBuddy 配套 Skill（**已确认 D9**） | 人设 + 工具 + 流程三层齐备 |
| dependencies | 见 6.3.3 | 声明连接器依赖 |

#### 6.3.2 Agent 定义文件（agents/liki-master.md）

- frontmatter：`name`、`description`（英文，AI 判断激活用）、`displayName{en,zh}`、`profession{en,zh}`、`maxTurns`（默认 50）；
- 正文 = 系统提示词，结构：人设 → 核心能力（四类领域）→ 工作流程（信息收集 → 调连接器工具 → 规则断语 → 条件性结论）→ 输出规范（保留因子/出处、可回溯）→ 边界（不构成医疗/法律/投资建议）；
- **不加 tools**：工具权限由系统分配 + 连接器提供（官方规定开发者不可自加工具）。

#### 6.3.3 MCP / 连接器依赖声明

**已确认 D10 = 方式 A**：`dependencies.connectors: ["liki"]`——依赖已上架的 Liki Connector，用户召唤专家前先引导连接，引导卡片由连接器自身提供；专家包不重复声明端点。上架顺序：先 Connector，后 Expert。

（备选方式 B：插件根 `.mcp.json` 自带 MCP，`url: https://liki.hk/mcp`，`type: http`，`x-workbuddy.auth.type: "none"`；本期不用，仅记录。）

#### 6.3.4 头像

- 512×512 PNG，≤500KB，漫画/插画风格、专业自然的命理师形象；
- 生成方式待定（AI 绘图或设计稿），放入 `avatars/expert.png`。

#### 6.3.5 校验

- 官方 `expert-manager` 技能用于流程与字段校验（引用 `plugin-json-spec.md` / `agent-md-spec.md` / `avatar-spec.md`）；
- 提交审核前逐项核对：字段必填、tags=3、quickPrompts=3、defaultInitPrompt 与 quickPrompts[0] 一致、displayDescription 40-50 字、expertType=agent。

## 7. RPC 退役路径（阶段化，后续触发）

1. 独立 MCP 上线：`/jsonrpc` 与 `/mcp` 并存，现有 npx 客户端不受影响；
2. npx Skill 迁移（D2 双轨验证通过后）：Python 工具层改 MCP client，npx 客户端全部走 MCP；
3. 观测期（至少一个发布周期）：确认 MCP 侧覆盖所有消费者后，移除 `rpc.go` 的 `/jsonrpc` handler；`RPCRegistry` 与 OpenRPC 文档保留为内核（`rpc.discover` 仅服务内部 / 测试）；
4. 文档与版本契约同步更新（`RELEASE_MODEL.md`、`SKILL_PACKAGE.md`、README 调用链描述）。

退役是阶段五的产出，不在本期执行；本期只保证退役路径成立（`/jsonrpc` 与 `/mcp` 相互解耦）。

## 8. 测试与验收

- 现有 `golden` / `functional` / `integration` / 160 题基准不动（内核零改动）；
- 新增 MCP 集成测试：
  - `initialize` / `tools/list` / `tools/call` 全量方法走查（参数、错误、结果包封）；
  - 与 `/jsonrpc` 同参数输入输出一致性对照测试（回归保障）；
  - `mcp-spec-check` 合规扫描；
- 连接器联调：WorkBuddy 测试环境（个人认证 → 上架 Skill + Connector → 测试调用）；
- `make gate` 纳入新测试。

## 9. 分阶段实施

| 阶段 | 内容 | 交付物 | 验收 | 归属 |
| --- | --- | --- | --- | --- |
| 一：独立 MCP | 引入 go-sdk；`engine/cmd/liki-mcp`（Streamable HTTP + stdio）；全量工具映射；MCP 集成测试 + spec check | 可独立运行的 `liki-mcp` | `mcp-spec-check` 通过；30 方法调用全部通过 | **本期** |
| 二：Skill 过渡 | 后续将 Python 工具层迁移到 MCP client；版本契约更新 | npx Skill 走 MCP（双轨验证后） | 现有 skill-up smoke / 集成测试通过 | 后续（D2 双轨生效后） |
| 三：WorkBuddy 上架 | 核对 Skill 规范；`connector-meta.json` + `mcp.json`；入驻认证；测试环境联调上线 | Skill + Connector 上架 | WorkBuddy 内自然语言调用成功 | **本期** |
| 四：专家 | `plugin.json` + `agents/liki-master.md` + avatars + 依赖声明 + 校验 | 专家「命理师」上架 | 专家中心可召唤，连接引导正常，调用成功 | **本期** |
| 五：RPC 退役 | 全部消费者迁 MCP 后，删 `/jsonrpc` handler，更新文档契约 | 引擎仅暴露 MCP | `make gate` 全绿；旧 URL 返回 404 | 后续（依赖阶段二完成） |

## 10. 风险与开放决策点

**已确认**：D1 官方 go-sdk；D2 双轨后迁移；D3 专家本期一次做完；D8 专家分类 `12-IndustryConsultant`；D9 专家预加载配套 Skill；D10 依赖方式 A（依赖已上架连接器）。

剩余开放项（实施时逐项确认）：

- **D4 HTTPS 与端点**：`liki.hk` 已具备；`/mcp` 路径与反向代理配置待定（连接器 mcp.json 指向）。
- **D5 鉴权**：WorkBuddy 对无鉴权公共 MCP 的接入约定需在联调时确认（连接器 mcp.json 据此定）。
- **D6 版本协议**：MCP `serverInfo.version` 与 CalVer/SemVer 的映射（`VERSION.txt` 契约）。
- **D7 限流/配额**：MCP 端点是否沿用现有 `NewRateLimiter` 限流策略，需在联调确认 WorkBuddy 端调用频率预期。

## 11. 本期待办落地清单（阶段一 + 三 + 四）

### 阶段一：独立 MCP Server（已完成，2026-09-21）

- [x] `engine/cmd/liki-mcp/main.go` 骨架：监听端口（默认 `:8081`）、版本注入（复用 VERSION embed）、复用中间件（限流 / CORS / BodyLimit / Recover / SecurityHeaders）
- [x] 引入 `github.com/modelcontextprotocol/go-sdk`（v1.8.0，go 1.26.4 兼容验证通过）
- [x] 工具映射：遍历 `RPCRegistry.Names()`，每个 method → `mcp.Tool`（低层 `Server.AddTool`，`InputSchema` 直接复用现有 Params JSON Schema，零改写参数契约）
- [x] handler 实现：`reg.Execute` → 剥离 RPC `_product/data` 信封 → `CallToolResult` 纯数据文本；业务错误（`RPCError`/参数校验失败）→ `IsError: true` 结果，协议错误用 `jsonrpc.Error`
- [x] 传输：Streamable HTTP 挂 `/mcp`（+ `/health` `/version`）；`-stdio` flag 本地调试
- [x] MCP 集成测试（`engine/cmd/liki-mcp/mcp_test.go`）：initialize / tools/list / tools/call / 无效参数 IsError / 未知工具 / 与 `/jsonrpc` 输出一致性对照 / unwrap 单元
- [x] 官方 conformance 验证：核心生命周期全过；业务工具集场景基线化（`conformance-baseline.yml`），`Baseline check passed`
- [x] `Makefile`：`build-mcp` / `test-mcp`（含 conformance）目标；Dockerfile 同镜像双二进制，compose 加 `mcp` 服务（`8083:8081`）

**验证结果**：`go test -race -count=1 -short ./...` 全绿；`golangci-lint` / `go vet` 0 issues；`mcp-spec-check` 报告 UNKNOWN（协商到 2025-11-25 stateful，符合 WorkBuddy 当前要求的 Streamable HTTP；2026-07-28 stateless 属生态级后续升级）。

### 阶段三：WorkBuddy 上架（Connector + Skill）

- [x] 对照 WorkBuddy Skill 规范核对现有 `SKILL.md`（字段、目录、标签、快捷提示词）——现有 Skill 为 SkillHub/npx 分发，维持不动；WorkBuddy 侧用连接器配套 Skill
- [x] 编写连接器配套 Skill `workbuddy/liki-connector/skills/liki-usage/SKILL.md`（工具用法手册：参数/调用链/示例/错误恢复/安全边界）
- [x] `connector-meta.json`（source=`liki`、type=mcp、examples 中英各 4 条、`minWorkbuddyVersion: 4.24.0`）+ `mcp.json`（streamableHttp，`https://liki.hk/mcp`，timeout 30000）+ `icon.svg`（临时「Liki」文字标，**待替换为正式品牌标识**）
- [ ] 个人认证入驻 → 测试环境联调（D5 鉴权、D7 限流确认）→ 上架 Skill + Connector（需人工/平台操作）
- [ ] 更新 `CHANGELOG.md` 与文档契约（README 调用链、`RELEASE_MODEL.md` 若涉及）

**校验**：connector-meta/mcp.json/plugin.json JSON 合法；`examples` 各 4 条；单 MCP Server；`auth_mode` 省略（无需认证，符合文档）；配套 Skill frontmatter 必填项齐全。

**提交前检查（32 项核对，31 通过）**：source kebab-case、单 server + HTTPS、无凭证硬编码、Skill 覆盖 32 工具、专家字段全部合规（tags/quickPrompts/字数/categoryId/dependencies/skills/homepage）。唯一待办：`avatars/expert.png` 头像。

### 阶段四：专家「命理师」上架

- [x] 新建 `workbuddy/liki-expert/`：`plugin.json` + `agents/liki-master.md`（人设/流程/边界）+ `skills/liki-usage/SKILL.md`（预加载）+ `README.md`
- [x] `plugin.json`：categoryId=`12-IndustryConsultant`、skills 预加载配套 Skill、dependencies.connectors=["liki"]（D8/D9/D10 已定）
- [ ] 头像 `avatars/expert.png`（512×512 PNG ≤500KB，漫画/插画风格）——**待生成**（当前缺）
- [x] 官方规范校验：tags=3、quickPrompts=3、quickPrompts[0]=defaultInitPrompt、displayDescription.zh=45 字（40-50）、plugin=name、agentName 匹配——脚本校验全过
- [ ] 用 `expert-manager` 技能逐项复核 + 本地上架测试（`--plugin-dir` 或 WorkBuddy 测试环境）
- [ ] 上架顺序：先 Connector，后 Expert（需人工/平台操作）

### 后续（不在本期）

- [ ] 阶段二：Python 工具层迁 MCP client（D2 双轨生效）
- [ ] 阶段五：RPC 退役（删 `/jsonrpc`，`RPCRegistry` 保留为内核）

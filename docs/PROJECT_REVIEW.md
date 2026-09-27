# Liki 本地项目评审

## 评审结论

Liki 当前的工程成熟度较高。仓库已经形成清晰的「确定性排盘引擎 + 规则判断层 + Agent 部署工件 + 兼容 Skill 包」分层结构，契约测试、黄金数据、MCP 集成、Docker 冒烟、依赖扫描和发布流水线都比较完整。总体适合按「修复少量发布阻断项后进入 release」处理。

本次快照评审给出综合评分 **8.2 / 10**：

| 维度 | 评价 |
| --- | --- |
| 架构边界 | 优。engine / counsel / agents / skills 责任清楚，工具目录和 Agent 白名单有契约约束。 |
| 测试与数据质量 | 优。根契约、Go 引擎、counsel 集成、golden/oracle 和 160 题覆盖形成多层防线。 |
| 安全与供应链 | 良。基础镜像 digest、Actions SHA pin、依赖哈希和漏洞扫描较完善，但仍有发布 action 未 pin 和公网工具信任边界问题。 |
| 发布与运维 | 良。版本同步、镜像 tag、健康检查、readiness 和发布物摘要完整；`engine-rpc` 过渡面和部分发布链一致性需要收敛。 |
| 产品交付一致性 | 中偏良。新增领域专家包后，根 Skill 路由与 root archive 打包范围出现明显不一致。 |

## 评审范围与证据

- 评审时间：2026-09-26 23:58，Asia/Shanghai。
- 分支 / 基线：`dev` @ `83f3f2f`（完整 SHA：`83f3f2f1cc4fe699f22ef0c01eeb262e5555534b`）。
- 工作区：非干净状态，包含 deployment、experts image、web skill bundle、领域专家包等未提交改动；本评审核对的是当前工作区快照，不只是最新 commit。
- 代码规模：约 23,853 行 Go 非测试代码、32,021 行 Go 测试代码；约 8,283 行 counsel 应用 Python 代码；根契约与数据测试约 11,643 行 Python。仓库跟踪文件 873 个。

### 本地验证结果

| 检查 | 结果 |
| --- | --- |
| `python3 -m pytest tests -q` | 通过：1086 passed，29 subtests passed。 |
| `scripts/lint-python.sh` | 失败：`scripts/check_deployment_schema.py` 有 2 个 Ruff 错误。 |
| `scripts/lint-md.sh` | 通过：139 个 Markdown 文件，0 issue。 |
| Go 格式 / short tests / integration tests，使用 pinned `golang:1.26.6-alpine` | 通过：全部 engine 包通过，包括 agent、HTTP、bazi integration。 |
| `go vet ./...`，使用 pinned Go 镜像 | 通过。 |
| counsel tests 对本地产出的当前 engine 镜像 | 通过：65 passed。 |
| engine MCP smoke | 通过：root 27 tools；bazi / huangli / fengshui / aux 分域工具面正常。 |
| 160 题规则覆盖 | 通过：160 题，零命中题 0；不同断语域按题目结构有不同覆盖，属预期分层覆盖。 |
| `make build-deployment` + deployment schema 校验 | 通过：router + 6 个领域专家工件生成并校验成功。 |
| `npm audit` | 通过：0 vulnerabilities。 |
| engine Docker 镜像构建 | 通过。 |

未把本地 `make gate` 作为整体通过结论：当前工作区已因 Python lint 失败而无法满足完整推送门槛。本机 shell 缺少 Go 1.26.6，Go 检查改用仓库 CI 同版本 pinned Docker 镜像完成。

## 架构评价

### 系统边界

当前架构的核心优点是「计算」和「解释」没有混在一起：

1. `engine/`：Go 实现 MCP 排盘、历法、天文、城市、八字、紫微、六爻、奇门、黄历和风水计算。工具目录统计为 27 个 engine tools。
2. `counsel/`：Python MCP 负责因子快照、断语、起名、六爻 / 奇门追问。工具目录统计为 12 个 counsel tools。工具边界有 manifest JSON Schema 校验，factor snapshot 带 digest、side、chart digest 和 context provenance。
3. `agents/` + `profiles/`：AgentDeployment 的 source of truth。当前专家拓扑为 router、bazi、ziwei、liuyao、qimen、fengshui、naming 共 7 个节点。`tools.allow` 与 `contracts/mcp-tool-catalog.json` 做交叉校验。
4. `skills/`：root skill 和六类领域专家包承担外部客户端兼容入口；构建时方法论文档会合入 AgentDeployment instruction。
5. `agents-image/`：只把 generated deployment 烘焙进 pinned liki-agents 基础镜像，构建上下文通过 `.dockerignore` 限制到工件和 Dockerfile，边界干净。

这个设计对命理系统尤其合适：排盘和规则命中可回归、可回溯，LLM 只负责流程组织与语言表达，不能凭空生成盘面或评分。

### 契约与版本治理

项目在版本和契约治理上明显高于一般 Skill 项目：

- 运行时使用统一 CalVer：`skills/liki/VERSION.txt`、counsel、engine MCP、engine RPC 当前均为 `2026.09.26.0`。
- AgentDeployment schema 由外部 liki-agents 仓拥有，本仓用 `contracts/agent-definition.version` 同时 pin version 和 SHA-256 digest，CI 还会从 OCI artifact 拉取后校验 digest。
- MCP 工具目录 `contracts/mcp-tool-catalog.json` 有独立 schema、runtime version 校验，并会在 release 时作为 OCI artifact 发布。
- GitHub Release 会发布 engine、counsel、experts 镜像、MCP tool catalog 和 web skill bundle；镜像记录 OCI digest。

### 测试质量

测试策略是当前项目最大的强项：

- 根测试不只测函数，还测试 SKILL 结构、frontmatter、文档引用、manifest、MCP tool catalog、deployment generation 等产品契约。
- engine 有大量 golden、external anchor、property、fuzz、domain oracle、integration tests，Go 测试代码量超过业务代码量。
- counsel tests 会启动本地 engine，覆盖 MCP 端到端路径，而不是只 mock 内部函数。
- 160 题数据检查能保证题目不会整体落入零断语覆盖，但它不是模型级准确率评测；仓库也明确把 model-backed evaluation 资产与当前 gate 分离。

## 主要发现

### P1：当前工作区不能通过完整质量门槛

`scripts/lint-python.sh` 当前失败，会阻塞 `make lint-python`、`make gate` 和 CI `python-lint`：

```text
scripts/check_deployment_schema.py:19:8 F401 os imported but unused
scripts/check_deployment_schema.py:95:29 W292 no newline at end of file
```

这是机械修复项：删除未用 import，补文件末尾换行。但因为 `make gate` 明确定义为推送门槛，当前状态不应直接进入 release。

### P1：root Skill 路由到未打包的新专家包

当前 `skills/liki/SKILL.md` 已把六爻、奇门、起名、风水路由到：

- `liki-liuyao`
- `liki-qimen`
- `liki-naming`
- `liki-fengshui`

这四个包已经出现在 `skills/` 中，并且 deployment generator 会读取它们的方法论卡；但 `scripts/build-archive.sh` 目前仍只把 root skill 和 `liki-bazi` / `liki-ziwei` 方法论卡打进 `dist/liki.tar.gz`。实际检查结果：

```text
liki-liuyao   missing
liki-qimen    missing
liki-naming   missing
liki-fengshui missing
```

这会造成安装根 archive 的客户端看到新路由，却拿不到对应专家包和指令。建议二选一：

1. 若四个新包是 root distribution 的正式组成，扩展 `build-archive.sh` 与打包校验，把四个包的方法论和必要入口纳入 archive；
2. 若它们只是独立兼容包，root archive 中的路由应继续指向现有 `divination` / `naming` / `fengshui` ENTRY，或显式说明需要额外安装对应包。

同时应新增 release archive 的路由闭合测试：解析 root SKILL 中引用的专家包，断言 archive listing 中存在对应资源。

### P1：发布上传 action 仍是浮动 tag

绝大多数 GitHub Actions 已使用完整 commit SHA pin，这是很好的供应链实践。但 release 上传步骤仍是：

```yaml
uses: softprops/action-gh-release@v2
```

建议改成与仓库其他 Actions 一致的 exact commit SHA，并在注释中标注真实版本。否则最后一个低权限敏感步骤仍受上游 tag 移动影响。

### P2：外部专家包连接的是全量 MCP 面，实际授权弱于 AgentDeployment

`agents/*/agent.yaml` 中的 `tools.allow` 是强约束，能实现专家最小权限。但新增的 `liki-liuyao`、`liki-qimen`、`liki-naming`、`liki-fengshui` 等外部兼容包的 `.mcp.json` 大多连接 root `https://liki.hk/engine/mcp` 和 `https://liki.hk/counsel/mcp`。这类客户端能发现全量工具，Skill 文档中的边界只是提示，不是服务端授权。

建议：

- 能使用单一分域端点的包改连 `/engine/mcp/{domain}`、`/counsel/mcp/{domain}`；
- 跨域场景使用按专家签发的受限 token，或说明该兼容包的信任模型弱于装配 Agent；
- 在文档中明确「外部 Skill 包边界 = 提示，装配 Agent 边界 = 授权」。

### P2：factor provenance 是完整性证据，不是公网真实性凭证

counsel 对 `factors` 做了较好的失败闭合校验：

- digest 能发现快照被裁剪或修改；
- `_provenance.side` 能防止 bazi / ziwei 混用；
- chart digest 能发现 period query 使用了另一张盘。

但 SHA-256 使用的是无密钥哈希，MCP 客户端理论上可以自行构造 `factors` 并重算 digest。也就是说它证明「factor、context、chart 相互一致」，不能证明「factor 一定来自本服务刚排出的盘」。公网 root MCP 上这是可被滥用的信任边界。

建议把 root aggregate 的信任模型写进 `docs/RUNTIME.md`；如需强保证，可让 `natal_query` / `period_query` 接收 chart 并服务端重算因子，或引入服务端签名的短时效 snapshot token。后者需要谨慎处理隐私与无存储原则。

### P2：`engine-rpc` 过渡运行时仍在最小镜像中

`engine/dev/Dockerfile` 同时打入 `engine-mcp` 和 `engine-rpc`，默认 entrypoint 还是 `engine-rpc`。`docs/RUNTIME.md` 已明确这是过渡状态并列出移除门槛，处理方式正确。但在当前 MCP 主架构下，它仍扩大镜像面、增加配置混淆，也可能延长旧 `/jsonrpc` 生命周期。

建议给 `engine-rpc` 移除设置明确目标版本或时间窗，并在指标中跟踪：

- `/jsonrpc` 当前流量；
- 依赖 `engine-rpc` 的部署数量；
- liki-web MCP 迁移完成度。

达到门槛后删除二进制、Dockerfile 构建段、README / runtime 文档和 smoke job。

### P2：架构文档略落后于当前专家拓扑

`docs/SYSTEM.md` 的总体模型仍把问卜描述为单个 `divination` 专家，而当前 AgentDeployment 已经拆成 `liuyao` 和 `qimen`，专家总数变为 6 个领域专家 + router。测试和 deployment 已对齐，但系统文档、README 架构图和发布物矩阵应同步，否则新维护者容易把 `skills/liki-*`、`skills/liki/divination` 与 `agents/liuyao` / `agents/qimen` 的关系混淆。

### P3：内部 archive 的可复现性不完整

`scripts/build-web-skill-bundle.sh` 对外层 bundle 使用了 sorted tar、固定 mtime、固定 owner 和 `gzip -n`，方向正确。但内层 `dist/liki.tar.gz` 由 `scripts/build-archive.sh` 用普通 `tar czf` 生成，没有 normalize 排序、mtime、owner 和 gzip header。因此外层的「可复现」仍依赖内层 tar 的字节稳定性。

建议把 `build-archive.sh` 也改成 reproducible tar，或在 bundle manifest 中明确只有外层容器可复现。

### P3：counsel 主文件和路径注入略有耦合

`counsel/app/counsel_mcp.py` 现在承担工具 schema 适配、natal、divination、naming、认证、限流、body limit 和 Starlette 装配，约 677 行。功能正确，但继续增加领域时会降低可读性。建议按以下方向拆分：

- MCP schema tool adapter；
- natal / divination / naming tool factories；
- HTTP middleware；
- app factory / routes。

同时 `sys.path.insert` 是为了让进程内工具模块 import legacy 目录，短期可行，但长期建议把 `natal/tools`、`naming`、`divination/tools` 改成显式 package 或适配层，避免模块名遮蔽和隐式导入路径。

### P3：水平扩容时限流语义需要说明

engine 和 counsel 的限流器都有容量上限和淘汰逻辑，能避免伪造 key 导致内存放大。但它们是进程内状态。多副本部署时，实际阈值约为「单副本阈值 × 副本数」。这对当前无状态 API 可接受，但应在 runtime 文档中明确；如果未来有计费或强防滥用需求，应迁移到共享限流层。

## 安全与隐私评价

### 已做较好的部分

- Docker 基础镜像和 uv 工具均使用 digest pin。
- engine 最终镜像基于 `scratch`，只含证书、用户和两个二进制；counsel 使用 non-root UID `65534`。
- Actions 几乎全部 SHA pin；Node、Go、Python 依赖均有审计入口。
- engine / counsel 支持 constant-time Bearer token、body limit、CORS、trusted proxy hops、限流与公开健康探针分离。
- 默认不信任 `X-Forwarded-For`，避免伪造限流键。
- 服务无数据库，出生数据定位为当前会话上下文；反馈契约明确禁止对话原文、出生数据、完整命盘和 RPC 全文。
- 未知城市外部地理编码可由 `LIKI_EXTERNAL_GEOCODING=off` 关闭，并且 README 已披露 Nominatim 外发风险。

### 需要收敛的部分

- `softprops/action-gh-release@v2` 未 SHA pin。
- 公网 root counsel 的 factor snapshot 只是一致性证据，不是真实性凭证。
- 外部专家包当前可发现全量工具面，弱于 AgentDeployment 的白名单模型。
- release bundle 与镜像分属不同 release job，虽然同 commit 源头一致，但 deploy 仓应优先使用 image digest 和 bundle SHA-256，不应只依赖 tag。

## 建议路线

### 合入当前变更前必须处理

1. 修复 `scripts/check_deployment_schema.py` 的两个 Ruff 错误。
2. 解决 root Skill 新路由与 `dist/liki.tar.gz` 不一致的问题，并加打包闭合测试。
3. SHA pin `softprops/action-gh-release`。
4. 更新 `docs/SYSTEM.md` 和 README 架构描述，明确 router + 6 专家、root skill / 兼容包 / AgentDeployment 的关系。
5. 本地或 CI 重新执行 `make gate`，把通过结果作为 release 前置条件。

### 下一个迭代

1. 给 root MCP factor snapshot 写清信任边界，必要时引入服务端重算或签名快照。
2. 收敛外部专家包的 MCP 端点与工具暴露面。
3. 给 `engine-rpc` 设定明确下线日期，并增加流量迁移指标。
4. 发布物增加 SBOM 和签名 / provenance；部署文档要求 deploy 仓 pin image digest。
5. 将 counsel app factory 按领域与 middleware 拆分，减少主文件复杂度。

### 长期改进

1. 将核心规则数据继续外置，并保持 schema version、golden coverage 和经典出处审计。
2. 建立跨仓 release matrix：liki-agents、engine、counsel、experts、web bundle、tool catalog、liki-web、liki-deploy 的兼容组合一目了然。
3. 为安全边界增加专项测试：未授权 token、超大 body、伪造 proxy header、错误 domain snapshot、跨盘 digest、root aggregate 滥用路径。
4. 统一本地开发 bootstrap，避免贡献者因缺少 Go 1.26.6 直接卡在 `make check`；可提供 devcontainer 或 `make tools`。

## 总体判断

Liki 已经不是简单 prompt / document skill，而是一个有确定内核、契约治理和多目标发布的工程系统。当前最大风险不是算法架构，而是最新专家拓扑扩张后「源、文档、deployment、distribution archive」四者短期漂移。先把 lint、archive 路由闭合、action pin 和文档同步修掉，这个工作区就具备比较扎实的 release 条件。

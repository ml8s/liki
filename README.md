<h1 align="center">Liki</h1>

<p align="center">
  专业命理 Skill<br>
  懂命理，用 Liki。<br>
  排盘由 Go 引擎计算，判断由规则表解释，依据可回溯。<br>
  八字 · 紫微 · 六爻 · 奇门 · 黄历择日 · 风水 · 起名
</p>

<p align="center">
  <a href="./README.en.md"><img alt="English" src="https://img.shields.io/badge/English-4a9e6b?style=flat-square"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/ml8s/liki/actions/workflows/ci.yml/badge.svg"></a>
  <a href="./LICENSE"><img alt="license" src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat-square"></a>
  <a href="https://liki.hk"><img alt="website" src="https://img.shields.io/badge/liki.hk-6d5acf?style=flat-square"></a>
</p>

## 安装

Liki 通过标准 MCP 提供能力，先配置 MCP 再安装 skill：

1. **配置 MCP**（二选一）：
   - **自动**：客户端支持插件 MCP 声明时，skill 自带 `.mcp.json`，启用即自动连接。
   - **手动**：在客户端添加两个 MCP server：
     - `liki-analysis`（判断层）：`https://liki.hk/analysis/mcp`
     - `liki-engine`（排盘/起名/风水）：`https://liki.hk/mcp`
2. **安装 skill**：

   ```bash
   npx skills add ml8s/liki
   ```

安装后，在支持 Agent Skills 的 AI 客户端中直接提问即可。Skill 启动时会检查版本；提示更新后重新执行：

```bash
npx skills add ml8s/liki -y
```

## 快速开始

| 目标 | 直接这样问 |
| --- | --- |
| 命理 | `算八字，1990-05-20 12:00 北京出生，男` |
| 起名 | `宝宝起名，2024-06-10 广州出生，男，姓陈` |
| 问卦 | `这个项目能成吗？什么时候有结果？` |
| 择日 | `下个月哪天适合搬家？` |
| 风水 | `我家风水怎么样？` |

第一次使用命理时，建议提供出生日期、尽量精确的时间、出生城市和性别。信息不全时，Skill 会追问或进入考时流程。

## 能力概览

| 领域 | 覆盖能力 |
| --- | --- |
| 命理 | 八字、紫微、大运流年、性格、婚姻、事业、财运、健康、学业、六亲、合盘 |
| 起名 | 新生儿起名、成人改名、外国人中文名、自选名字评估 |
| 问卦 | 六爻成败与应期、奇门方向与时机、黄历择日 |
| 风水 | 八宅命卦与门主灶、玄空飞星、流年风水 |

## 可信度与边界

- 排盘由 Go 天文历算引擎完成，支持真太阳时、经纬度时区和节气计算。
- 断语来自 819 条规则真值表，命中结果保留因子依据和经典出处。
- 160 道命理师大赛真题用于独立评测；答案与评测过程隔离。
- 出生信息只保留在当前会话；Skill 不索要真实姓名，不在对话之外存储数据。
- 命理结论是传统文化视角的条件性解读，不构成医疗、法律、投资或重大人生决策建议。

## 常见问题

### 不知道出生时辰怎么办？

Skill 会追问，或使用真实事件进入考时流程。证据不足时明确标注，不会默认时辰。

### 需要联网吗？

默认需要访问 JSON-RPC 引擎。高级用户可以自建 engine，并用 `LIKI_RPC_URL` 指向本地服务。

### 出生数据会被存储吗？

不会。出生数据只在当前会话上下文中使用；不写入本地档案，也不提交到反馈。

### 怎么更新？

按提示重新执行 `npx skills add ml8s/liki -y`。版本校验失败时不降级调用旧 RPC。

## 文档

| 文档 | 用途 |
| --- | --- |
| [用户指南](./docs/USER_GUIDE.md) | 完整使用说明、领域流程、FAQ 与输出边界 |
| [Skill 包结构](./docs/SKILL_PACKAGE.md) | 统一 Skill 的目录、入口和打包契约 |
| [README 规范](./docs/README_STYLE.md) | 中英文 README 的结构、标题和排版契约 |
| [八字 / 紫微模型](./docs/BAZI_MODEL.md) | 排盘、因子、断语和查询边界 |
| [问卦模型](./docs/DIVINATION_MODEL.md) | 六爻、奇门、黄历的 snapshot 与 answer 契约 |
| [风水模型](./docs/FENGSHUI_MODEL.md) | 八宅、玄空、流年和冲突裁决 |
| [起名模型](./docs/NAMING_MODEL.md) | 用神策略、字池、候选名和校验边界 |
| [反馈模型](./docs/FEEDBACK_MODEL.md) | Agent feedback 的隐私和契约 |
| [版本与发布](./docs/RELEASE_MODEL.md) | CalVer 运行时版本和 SemVer 发行版 |

## 开发者

### 开发环境

```bash
make hooks         # 安装 git hooks
make check         # 所有静态检查（格式 + lint + schema + docs）
make gate          # 本地推送前门槛（lint + check + test，~3min）
make build-archive # 打包 unified Liki skill
```

### 架构

```text
skills/liki/
├── SKILL.md              # 唯一 skill 入口：路由、安全、反馈
├── VERSION.txt           # 唯一分发版本
├── FAQ.md                # 运行失败与恢复契约
├── natal/                # 八紫双盘 / 本命：八字 + 紫微 + 合参
├── divination/           # 六爻 + 奇门 + 黄历：ENTRY / TOOLS / app / domains / tools
├── fengshui/             # 八宅 + 玄空：ENTRY / RPC / app / domains
└── naming/               # 起名：ENTRY / RPC / app / domains
```

仓库根的 `engine/`、`tests/` 和 `scripts/` 分别承载引擎、评测和构建脚本；可安装包只来自 `skills/liki`。调用链固定为：`SKILL.md` → `ENTRY.md` → app 卡 → Python 工具或固定 RPC。natal / divination 通过 Python 工具层编排 RPC、snapshot、因子和断语；naming / fengshui 没有本地 Python 工具层，只使用固定 JSON-RPC 报文。

### 引擎镜像

引擎镜像随 GitHub Release 自动发布：`docker pull ghcr.io/ml8s/liki-engine:latest`。源码构建使用 `engine/deploy/docker-compose.yml`。

### 领域契约

| 契约 | 用途 |
| --- | --- |
| [natal TOOLS](./skills/liki/natal/TOOLS.md) | 五个本命分析工具的完整 stdin 报文 |
| [divination TOOLS](./skills/liki/divination/TOOLS.md) | 六爻、奇门、黄历工具报文 |
| [naming ENTRY](./skills/liki/naming/ENTRY.md) | 起名：用神取用 + 五行选字（engine-pro MCP 工具）|
| [fengshui ENTRY](./skills/liki/fengshui/ENTRY.md) | 风水：八宅、玄空与流年（engine MCP 工具）|

### 测试与发布

```bash
make test           # 所有测试（pytest + Go 引擎全量）
make verify        # 端到端集成测试
make golden # golden 全量
```

分层详见 [Release model](./docs/RELEASE_MODEL.md)：`lint → check → test → verify → gate`。

正式发布使用 SemVer tag，运行兼容版本使用 CalVer。规则见 [Release model](./docs/RELEASE_MODEL.md)。

### 设计原则

- 分层单一职责：根入口、领域入口、App 卡、领域知识和工具层不互相替代。
- 单一事实源：工具契约来自 `skill-tools.json`，因子和断语来自 CSV 长表。
- 双体系显式合参：八字和紫微分侧计算，冲突分层列证。
- 评测驱动：golden、functional、integration、skill-up smoke 和 160 题基准分层运行。

## 贡献

阅读 [CONTRIBUTING.md](./CONTRIBUTING.md)，提交 PR 前同步更新 `CHANGELOG.md` 和版本契约。版本历史见 [CHANGELOG.md](./CHANGELOG.md)。

> 懂命理，用 Liki。

## 许可与声明

MIT。命理结论为传统文化视角，仅供参考；不构成医疗诊断、法律建议、金融投资预测或重大人生决策。

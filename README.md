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
     - `counsel`（判断层）：`https://liki.hk/counsel/mcp`
     - `engine`（排盘/风水）：`https://liki.hk/mcp/engine`
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

默认需要访问 engine / counsel MCP 服务（liki.hk）。高级用户可以自建服务，并用环境变量（`LIKI_MCP_URL` 指向 engine，`LIKI_COUNSEL_SERVICE_DOMAIN` 指向 counsel）指向本地。

### 出生数据会被存储吗？

不会。出生数据只在当前会话上下文中使用；不写入本地档案，也不提交到反馈。

### 怎么更新？

按提示重新执行 `npx skills add ml8s/liki -y`。版本校验失败时不降级调用旧 RPC。

## 文档

| 文档 | 用途 |
| --- | --- |
| [用户指南](./docs/USER_GUIDE.md) | 完整使用说明、领域流程、FAQ 与输出边界 |
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

Liki 采用「排盘（计算）与判断（规则）正交化」的两层架构，通过标准 MCP 提供能力：

#### 领域模型

- **engine（Go）** —— 确定性计算层：天文历算、排盘、历法、黄历、字库。
  输出结构化盘面/卦象/事实（chart / pan / snapshot），不做命理判断。
- **counsel（Python）** —— 判断层：八字/紫微判断、六爻/奇门算卦、起名评估。
  消费 engine 的盘面事实，用规则表（真值表 + 引擎规则表）产出断语与候选项，依据可回溯。
- **黄历** —— 纯 engine（历法 + 建除事项适配），不走 counsel。

#### 服务端点

| 层 | MCP 端点 | 领域 |
| --- | --- | --- |
| engine | `/mcp/engine/{bazi,ziwei,liuyao,qimen,huangli,...}` | 排盘 / 历法 / 黄历 |
| counsel | `/counsel/mcp/{bazi,ziwei,liuyao,qimen,naming}` | 判断 / 算卦 / 起名 |

#### 调用链

`SKILL.md` 路由 → engine 排盘 → counsel 判断 → 断语（`assertion_id` + 经典依据，可回溯）

#### 代码结构

- `engine/` —— Go 引擎（排盘/历法/黄历），分域 MCP 服务
- `counsel/` —— Python 判断层（多域 MCP server）
- `skills/liki/` —— skill 能力文档（路由 / 边界 / 领域知识，工具经 `tools/list` 自举）
- `tests/` —— 契约与集成测试；`scripts/` —— 构建与评测脚本

### 引擎镜像

引擎镜像随 GitHub Release 自动发布：`docker pull ghcr.io/ml8s/liki-engine:latest`。本地源码构建使用 `engine/dev/docker-compose.yml`。

### 领域契约

| 契约 | 用途 |
| --- | --- |
| [natal TOOLS](./skills/liki/natal/TOOLS.md) | 八字/紫微本命与应期判断的编排与契约（工具经 `tools/list` 自举）|
| [divination TOOLS](./skills/liki/divination/TOOLS.md) | 六爻、奇门、黄历的编排与契约 |
| [naming ENTRY](./skills/liki/naming/ENTRY.md) | 起名：用神取用 + 五行选字（counsel）|
| [fengshui ENTRY](./skills/liki/fengshui/ENTRY.md) | 风水：八宅、玄空与流年（engine）|

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
- 单一事实源：工具 schema 来自工具清单（skill/counsel-tools.json，经 `tools/list` 自举），因子和断语来自 CSV 长表与引擎规则表。
- 双体系显式合参：八字和紫微分侧计算，冲突分层列证。
- 评测驱动：golden、functional、integration、skill-up smoke 和 160 题基准分层运行。

## 贡献

阅读 [CONTRIBUTING.md](./CONTRIBUTING.md)，提交 PR 前同步更新 `CHANGELOG.md` 和版本契约。版本历史见 [CHANGELOG.md](./CHANGELOG.md)。

> 懂命理，用 Liki。

## 许可与声明

MIT。命理结论为传统文化视角，仅供参考；不构成医疗诊断、法律建议、金融投资预测或重大人生决策。

<h1 align="center">
  <img src="assets/logo.svg" alt="" width="96" align="absmiddle">
  Liki Skill
</h1>

<p align="center">
  <strong>专业命理 Skill。</strong><br>
  八字 · 紫微 · 六爻 · 奇门 · 黄历择日 · 风水 · 起名
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/ml8s/liki/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://modelcontextprotocol.io/specification/2025-06-18/basic/transports"><img alt="MCP" src="https://img.shields.io/badge/MCP-compatible-0891B2?style=flat-square"></a>
  <a href="./LICENSE"><img alt="License" src="https://img.shields.io/github/license/ml8s/liki?style=flat-square"></a>
  <a href="https://liki.hk"><img alt="Website" src="https://img.shields.io/badge/site-liki.hk-6D5ACF?style=flat-square"></a>
  <a href="./README.md"><img alt="English" src="https://img.shields.io/badge/docs-English-4A9E6B?style=flat-square"></a>
</p>

Liki 命理 — 专业命理分析。Liki Skill 面向支持 MCP 的助手，由专业计算引擎驱动，提供可验证的排盘与 AI 辅助解读。

## Liki Skill 是什么

Liki Skill 是 Liki 命理面向 AI 助手的分发形态，覆盖八字排盘、紫微斗数、六爻占卜、奇门遁甲、黄历择日、风水、起名七大领域。

它不是"给大模型套个提示词"。排盘由天文历算引擎计算，断语从带经典出处的规则真值表里命中，时辰不清先考时再下结论——同一个问题，任何时候问，都得到同一个有依据的答案。

传输协议是 [MCP Streamable HTTP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports)。

## 30 秒装上

```bash
npx skills add ml8s/liki
```

> **端点更新：** 免费 Liki 使用 `https://mcp.tokflux.com/engine/mcp` 和 `https://mcp.tokflux.com/counsel/mcp`。付费渠道变体使用 `https://mcp.tokflux.com/mcp`。旧 `liki.hk/engine/mcp` 和 `liki.hk/counsel/mcp` 已永久退役，不做跳转。工具不可用时执行：
>
> ```bash
> npx skills add ml8s/liki -y
> ```

通过 `npx` 安装需要本机有 Node.js；Skill 运行时不依赖本地 Node。装完即用，Skill 通过托管的免费 engine/counsel MCP 面工作，不需要配置任何 API key；付费渠道变体使用产品 MCP 端点；不想碰命令行，可以直接用 [liki.hk](https://liki.hk) 网页版排盘。提示更新时执行：

```bash
npx skills add ml8s/liki -y
```

支持 Claude Code、Codex、Cursor、WorkBuddy 以及一切支持 MCP 的宿主环境。

## 能问什么

| 想问什么 | 直接这样说 |
| --- | --- |
| 八字命盘 | `算八字，1990-05-20 12:00 北京出生，男` |
| 事业时机 | `我明年适合换工作吗？八字是 1988-07-02 成都出生` |
| 婚姻合盘 | `合婚：男 1990-05-20 12:00 北京，女 1992-03-15 14:30 广州` |
| 六爻问卦 | `这个项目能成吗？什么时候有结果？` |
| 黄历择日 | `下个月哪天适合搬家？` |
| 风水 | `我家房子坐北朝南，2008 年建的，风水怎么样？` |
| 宝宝起名 | `宝宝起名，2024-06-10 广州出生，男，姓陈` |

命理类问题建议提供出生日期、尽量精确的时辰、出生城市和性别。信息不全也没关系——Skill 会追问，或进入考时流程，不会默默替你猜。

## 直接问大模型算命，问题在哪

同一张盘问三遍，大模型可能给你三个答案。命理恰恰"差之毫厘，谬以千里"：节气边界错一天，年柱就换，后面的断语全部跟着错。

|  | 直接问大模型 | Liki Skill |
| --- | --- | --- |
| 排盘 | 模型心算，节气/真太阳时易错，结果可能不同 | 天文历算引擎，同样输入永远同样盘面 |
| 断语 | 自由发挥，没有出处 | 829 条规则真值表，每条带经典原文 |
| 时辰 | 钟表时间直接用，边界处静默二选一 | 候选集 + 原因 → 逐盘校准，证据不足明说 |
| 输出 | 纯文本，难以复用 | 结构化 JSON + SHA-256 摘要，可接入产品 |
| 验证 | 无法验证 | 开源，全链路可审计 |

## 准确性是怎么做出来的

### 排盘引擎：四方对算出来的

`192` 个节气边界案例（2024–2027 年，每年 24 节气 × 2 个锚点）与 **lunar-python、sxtwl、bazi-calculator** 三个开源实现对算，策略只有一条：`all_oracles_must_agree`——任何一个不一致都不允许通过。紫微实现与 iztro 兼容数据源对齐。引擎与经典不符的地方，按《穷通宝鉴》等原文校正，固化为防回归锚点。

### 断语：从真值表里命中，不是临场发挥

断语不是模型生成的——**829 条断语、893 个条件组、1355 条件行**，底层是 511 个本命因子和 111 个流年因子。命中即直断，不命中明说；每条断语自带依据，可以一路点回《三命通会》《滴天髓》《黄帝内经》的原文。比如这条真实数据：

| 断语 | 依据 |
| --- | --- |
| 五行缺木：肝/胆/筋骨/目 偏弱 | 木主肝（《黄帝内经》"东方生风，风生木，木生酸，酸生肝"） |

### 考时：钟表时间不等于命理时间

医院钟表记的是标准时，不是出生那一刻的太阳——经度不同，真太阳时可能差出半个时辰；节气交界差几分钟年柱就换；23 点后的晚子时，日柱归属最容易错。所以考时是显式流程：

```text
你的时间 + 出生地 → 归一化解析 → 候选集（每个候选带原因）
→ 逐候选完整排盘 → 对照 2-3 件你记得的真实大事及年份
→ 校准结论 + 证据覆盖说明；证据不足？明说"考时证据不足"，绝不默认
```

在此之上，160 道命理师大赛真题（Mingli160）用于规则链覆盖与零命中检查，作为手动校准/评估集。

## 安全与边界

- 出生信息只保留在当前会话：不写档案、不要真实姓名、不进反馈。
- 城市查询只用内置 WGS-84 地理表，不做外部地理编码请求。
- 敏感问题走关怀优先的安全策略。
- 结论是传统文化视角下的条件性解读，不构成医疗、法律、投资或重大人生决策建议。

## 常见问题

### AI 算命到底准不准？

直接问大模型，答案是"临场生成"的；Liki Skill 的答案是"算"出来的——排盘来自与三个开源实现对算过的天文引擎，断语来自带经典出处的真值表，可复现、可检验。

### 不知道出生时辰怎么办？

Skill 会追问，或用 2-3 件你记得的真实大事（带年份）进入考时流程；证据不足会明说"考时证据不足"，绝不默认时辰。宝宝和青少年的盘直接按提供的时辰排；起名走期望五行策略，不需要考时。

### 收费吗？

Skill 本体免费安装使用（`liki-free`）。另有付费的 `liki-pro` 分发变体，走 liki-mcp 产品端点，面向 SkillHub / WorkBuddy 渠道，支持 SkillPay / 402 支付流程。

### 我的生辰八字会被保存吗？

不会。出生信息只在当前会话上下文中使用，不写入本地档案，也不通过反馈提交。

### 和 liki.hk 网页版是什么关系？

同一套引擎与规则体系。liki.hk 面向网页用户，Liki Skill 把同样的能力带进你的 AI 助手。计算库在 [ml8s/mingligo](https://github.com/ml8s/mingligo)，产品 MCP 在 [ml8s/liki-mcp](https://github.com/ml8s/liki-mcp)，本仓库承载可安装 Skill、渠道分发配置与 expert 装配定义。

## 文档与链接

- 官网：[liki.hk](https://liki.hk)
- 计算库：[ml8s/mingligo](https://github.com/ml8s/mingligo)
- 产品 MCP：[ml8s/liki-mcp](https://github.com/ml8s/liki-mcp)

## 参与贡献

欢迎提交 Issue 和 Pull Request。Skill 文案要保持事实可查，每条规则都注明经典出处。提交前运行：

```bash
python3 -m venv .venv
.venv/bin/python -m pip install -r requirements-dev.txt
make setup-hooks
make gate
```

`make setup-hooks` 会配置当前 checkout：每次 `git push` 前先运行 `make gate`。

## 许可证

本项目基于 [MIT License](LICENSE) 发布。

> 懂命理，用 Liki。

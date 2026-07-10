<p align="center">
  <img alt="灵机 Liki" src="https://img.shields.io/badge/灵机-Liki-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>灵机 liki — 人人都是命理师</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-官网-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

## 目录

- [简介](#简介)
- [为什么选灵机](#为什么选灵机)
- [快速开始](#快速开始)
- [安装](#安装)
- [Skills 详情](#skills-详情)
- [命理引擎](#命理引擎)
- [架构](#架构)
- [版本](#版本)
- [相关链接](#相关链接)
- [免责声明](#免责声明)

## 简介

**灵机（liki）** 是全家桶命理玄学 Skill，`npx skills add ml8s/liki` 一条命令装完即用。四个子 skill 覆盖命理咨询的主要场景：

- **[命理](./mingli/)** — 八字紫微综合，排盘、用神、合会冲刑、大运流年
- **[起名](./qiming/)** — 八字用神 + 五格三才，全维度起名
- **[问卦](./wengua/)** — 智能路由奇门遁甲 / 六爻 / 黄历择日
- **[风水](./fengshui/)** — 八宅命卦 + 玄空飞星



## 为什么选灵机

- **AI 原生** — 不是简单的 prompt，而是完整的 Agent 工作流：参数收集 → API 排盘 → LLM 解读 → 结构化报告
- **计算精准** — 后端基于天文历算，支持真太阳时、夏令时校正、经纬度时区换算
- **版本管理** — 每个 skill 启动时自动检查更新，持续迭代
- **注重隐私** — 不在对话外存储出生信息，纯会话内计算
- **开箱即用** — 一条命令安装，直接在 AI 助手中调用

## 快速开始

```bash
npx skills add ml8s/liki     # 安装全家桶
```

然后在 AI 助手中直接对话：

> 帮我排个八字，1990年5月20日中午12点，北京出生，男

Agent 会自动收集参数、调 `tianwen.time` 校正真太阳时、排盘、取用神，输出结构化命理报告。

## 安装

```bash
# 从 GitHub 安装（仅全家桶）
npx skills add ml8s/liki

# 或从官网安装
npx skills add https://liki.hk/skills
```

## Skills 详情

### 命理 `liki-mingli`

八字紫微综合命理分析，先排八字（排盘、用神、合会冲刑），再排紫微（十二宫、星曜、四化），LLM 综合两套体系解读。覆盖大运流年全时间维度。

| 功能 | Method |
|------|--------|
| 排盘 | `bazi.chart` |
| 用神 | `bazi.yongshen` |
| 合会冲刑 | `bazi.hehui` |
| 合盘 | `bazi.bond` |
| 紫微排盘 | `ziwei.chart` |
| 紫微大限/流年 | `ziwei.daxian` / `ziwei.liunian` |
| 流年/流月/流日/流时 | `bazi.liunian` / `liuyue` / `liuri` / `liushi` |

### 起名 `liki-qiming`

AI 起名顾问，结合八字用神与五格三才，提供全维度起名服务。

| 功能 | Method |
|------|--------|
| 命理取字 | `qiming.pick` |
| 五格笔画 | `qiming.wuge` |
| 组合生成 | `qiming.build` |
| 终检评估 | `qiming.check` |

**流程**：收集生辰 → 排八字 → 命理决策（用+用 / 用+喜）→ `pick` 取字 → `wuge` 五格 → `build` 组合 → 筛选 8 名 → `check` 终检 → 报告

### 问卦 `liki-wengua`

AI 占卜择日师，LLM 智能路由选择最合适的方法：

| 问题类型 | 方法 | 场景 |
|----------|------|------|
| 择日（哪天好） | 黄历 | 嫁娶、开业、搬家 |
| 吉凶结果 | 六爻 | 成败、得失、寻物 |
| 时机策略 | 奇门遁甲 | 进退、方向、局势 |

| 功能 | Method |
|------|--------|
| 奇门排盘 | `qimen.pan` |
| 六爻起卦 / 装卦 | `liuyao.qigua` / `liuyao.chart` |
| 黄历日览 / 月览 | `huangli.date` / `huangli.month` |
| 八字合参择日 | `huangli.bond.date` / `bond.month` |

### 风水 `liki-fengshui`

AI 风水分析师，兼顾八宅与玄空两大流派。八宅命卦看人宅匹配，玄空飞星断时空吉凶，三元九运辅助趋势判断。

| 功能 | Method |
|------|--------|
| 八宅命卦 / 宅盘 | `bazhai.minggua` / `bazhai.chart` |
| 玄空三元九运 / 飞星盘 | `xuankong.sanyuan` / `xuankong.chart` |

## 命理引擎

所有 skill 共享同一套 JSON-RPC 2.0 命理引擎，按领域分组：

| 领域 | Method | 说明 |
|------|--------|------|
| 天文 | `tianwen.time` | 真太阳时 + 公历 + 农历 |
| 天文 | `time.now` | 服务端 UTC/CST 时间 |
| 天文 | `city` | 城市经纬度查询 |
| 八字 | `bazi.chart` | 排盘（四柱、十神、大运） |
| 八字 | `bazi.yongshen` | 三派用神（扶抑 / 调候 / 格局） |
| 八字 | `bazi.hehui` | 合会冲刑 |
| 八字 | `bazi.chart_extra` | 补充信息（胎元/命宫/身宫等） |
| 八字 | `bazi.bond` | 合盘 |
| 八字 | `bazi.liunian` / `liuyue` / `liuri` / `liushi` | 流年 / 流月 / 流日 / 流时 |
| 八字 | `bazi.xiaoyun` / `xiaoxian` | 小运 / 小限 |
| 紫微 | `ziwei.chart` | 排盘（十二宫、星曜、四化） |
| 紫微 | `ziwei.daxian` | 大限 |
| 紫微 | `ziwei.liunian` / `liuyue` / `liuri` | 流年 / 流月 / 流日 |
| 紫微 | `ziwei.bond` | 合盘 |
| 起名 | `qiming.pick` | 五行字库 |
| 起名 | `qiming.wuge` | 五格笔画配对 |
| 起名 | `qiming.build` | 组合生成 |
| 起名 | `qiming.check` | 五格 + 三才 + 音韵 + 五行终检 |
| 占卜 | `qimen.pan` | 奇门遁甲排盘（时家 / 日家 / 月家 / 年家） |
| 占卜 | `liuyao.qigua` / `chart` | 六爻起卦 / 装卦 |
| 占卜 | `huangli.date` / `month` | 黄历日览 / 月览 |
| 占卜 | `huangli.bond.date` / `bond.month` | 八字合参择日 / 择月 |
| 风水 | `bazhai.minggua` / `chart` | 八宅命卦 / 宅盘 |
| 风水 | `xuankong.sanyuan` / `chart` | 玄空三元九运 / 飞星盘 |

## 架构

```
用户提问
  │
  ▼
┌──────────────────────────────────────┐
│ Skill Agent (SKILL.md)               │
│                                      │
│ 参数收集 → 会话路由 → 调用引擎       │
│     │                              │
│     ▼                              │
│ rpc.discover ──→ JSON-RPC 2.0 ──→  │
│                  liki.hk            │
│                  ├─ tianwen.time    │
│                  ├─ bazi.chart      │
│                  ├─ bazi.yongshen   │
│                  ├─ ziwei.chart     │
│                  ├─ qiming.*        │
│                  ├─ qimen.pan       │
│                  ├─ liuyao.*        │
│                  ├─ huangli.*       │
│                  └─ bazhai/xuankong │
└──────────────────────────────────────┘
  │
  ▼
LLM 解读 → 结构化报告
```

每个 Skill 是一个自包含的 Agent 定义文件（`SKILL.md`），包含完整的会话流程和 API 调用方法。公共规则（行为边界、输出规范、错误处理）统一放在根 `SKILL.md` 中。引擎方法发现通过 `rpc.discover`。

## 版本

当前稳定版 **1.9.0**。详见 [CHANGELOG.md](./CHANGELOG.md)。

- 所有 skill 统一版本管理，启动时自动检查更新
- 遵循语义化版本，兼容性变化在 minor 版本体现

## 相关链接

- 官网：[liki.hk](https://liki.hk)
- AI Agent 入口：[llms.txt](https://liki.hk/llms.txt)
- GitHub：[ml8s/liki](https://github.com/ml8s/liki)

## 免责声明

灵机（Liki）提供的命理分析基于中国传统文化视角，仅供研究与参考。命理结论**不构成**医疗诊断、法律建议、金融投资预测或人生重大决策的依据。请理性看待，保持自主积极的人生态度。

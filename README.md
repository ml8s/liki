<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)">
    <img alt="灵机 Liki" src="https://img.shields.io/badge/灵机-Liki-oklch(0.58_0.18_266)?style=for-the-badge&logo=openai&logoColor=white&labelColor=oklch(0.25_0.05_266)">
  </picture>
</p>

<p align="center">
  <strong>AI 命理顾问</strong> &mdash; 八字 · 紫微 · 起名 · 占卜择日 · 风水
</p>

<p align="center">
  <a href="#安装"><img src="https://img.shields.io/badge/npx-skills_add-liki?style=flat&logo=npm&logoColor=white&labelColor=oklch(0.25_0.05_266)&color=oklch(0.58_0.18_266)"></a>
  <a href="./CHANGELOG.md"><img src="https://img.shields.io/badge/version-1.7.1-oklch(0.55_0.12_145)?style=flat&logo=semver&logoColor=white&labelColor=oklch(0.25_0.05_145)"></a>
  <a href="#license"><img src="https://img.shields.io/badge/license-MIT-oklch(0.45_0.08_200)?style=flat"></a>
</p>

---

## 简介

**灵机（Liki）** 是一个 AI 驱动的中国传统命理技能包合集，面向 Claude Code 等 AI 编程助手。每个技能都是一个独立的 AI Agent，内置完整会话流程，通过 JSON-RPC 2.0 后端提供精准的命理计算，再由 LLM 进行人性化解读。

五个技能覆盖了命理咨询的主要场景：

- **[起名](./naming/)** &mdash; 八字用神 + 五格三才，全维度起名
- **[八字](./bazi/)** &mdash; 排盘、合盘、大运、流年流月流日
- **[紫微](./ziwei/)** &mdash; 十二宫星曜四化，大限流年
- **[问事](./ask/)** &mdash; 智能路由奇门遁甲 / 六爻 / 黄历择日
- **[风水](./fengshui/)** &mdash; 八宅命卦 + 玄空飞星

## 为什么选灵机

- **AI 原生** &mdash; 不是简单的 prompt，而是完整的 Agent 工作流：参数收集 → API 排盘 → LLM 解读 → 结构化报告
- **计算精准** &mdash; 后端基于天文历算，支持真太阳时、夏令时校正、经纬度时区换算
- **版本管理** &mdash; 每个技能启动时自动检查更新，持续迭代
- **注重隐私** &mdash; 不在对话外存储出生信息，纯会话内计算
- **开箱即用** &mdash; 安装后直接在 Claude Code 中调用，零配置

## 安装

```bash
# 从 npm 安装
npx skills add ml8s/liki            # 全家桶
npx skills add ml8s/liki/naming     # 起名
npx skills add ml8s/liki/bazi       # 八字
npx skills add ml8s/liki/ziwei      # 紫微
npx skills add ml8s/liki/ask        # 问事
npx skills add ml8s/liki/fengshui   # 风水

# 或从官网安装
npx skills add https://liki.hk/skills           # 全家桶
npx skills add https://liki.hk/skills/naming    # 起名
npx skills add https://liki.hk/skills/bazi      # 八字
npx skills add https://liki.hk/skills/ziwei     # 紫微
npx skills add https://liki.hk/skills/ask       # 问事
npx skills add https://liki.hk/skills/fengshui  # 风水
```

## 技能详情

### 起名 `liki-naming`

AI 起名顾问，结合八字用神与五格三才，提供全维度起名服务。

| 场景 | 说明 |
|------|------|
| 宝宝起名 | 结合生辰八字，五行补益 + 音韵字形 |
| 成人改名 | 参考原名，优化五行搭配 |
| 公司起名 | 行业五行 + 品牌音韵 |
| 品牌命名 | 定位风格，批量候选 |

**流程**：收集生辰 → 排八字取用神 → `pick` 五行字库 → `wuge` 五格筛选 → `build` 组合 → `check` 终检 → 报告

### 八字 `liki-bazi`

AI 八字命理分析，基于《滴天髓》十天干体性解读，覆盖全时间维度。

| 功能 | Method |
|------|--------|
| 排盘 | `bazi.chart` |
| 合盘 | `bazi.bond` |
| 大运 | 排盘内包含 |
| 流年/流月/流日/流时 | `bazi.liunian` 等 |
| 小运/小限 | `bazi.xiaoyun` / `bazi.xiaoxian` |

### 紫微 `liki-ziwei`

AI 紫微斗数分析师，十二宫星曜四化全量排盘。先排八字获取四柱，再排紫微盘，提供命盘全局解读 + 大限流年推断。

### 问事 `liki-ask`

AI 占卜择日师，LLM 智能路由选择最合适的方法：

| 问题类型 | 方法 | 场景 |
|----------|------|------|
| 择日（哪天好） | 黄历 | 嫁娶、开业、搬家 |
| 吉凶结果 | 六爻 | 成败、得失、寻物 |
| 时机策略 | 奇门遁甲 | 进退、方向、局势 |

支持黄历八字合参择日，结合命主喜忌筛选最佳日期。

### 风水 `liki-fengshui`

AI 风水分析师，兼顾八宅与玄空两大流派。八宅命卦看人宅匹配，玄空飞星断时空吉凶，三元九运辅助趋势判断。

## 架构

```
┌────────────┐     ┌──────────────┐     ┌────────────┐
│ 用户/LLM   │ ──▶ │ Skill Agent  │ ──▶ │ JSON-RPC   │
│            │ ◀── │ (SKILL.md)   │ ◀── │ liki.hk     │
└────────────┘     └──────────────┘     └────────────┘
```

每个 Skill 是一个自包含的 Agent 定义文件（`SKILL.md`），包含完整的会话流程、参数收集规范、API 调用方法和行为边界。后端提供 JSON-RPC 2.0 接口，方法发现通过 `rpc.discover`。

## 版本

当前稳定版 **1.7.1**。详见 [CHANGELOG.md](./CHANGELOG.md)。

- 每个技能独立版本管理，启动时自动检查更新
- 遵循语义化版本，兼容性变化在 minor 版本体现

## 免责声明

灵机（Liki）提供的命理分析基于中国传统文化视角，仅供研究与参考。命理结论**不构成**医疗诊断、法律建议、金融投资预测或人生重大决策的依据。请理性看待，保持自主积极的人生态度。

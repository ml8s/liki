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

## 简介

**灵机（liki）** 是全家桶命理玄学 Skill，`npx skills add ml8s/liki` 一条命令装完即用。四个子 skill 覆盖命理咨询的主要场景：

- **[命理](./mingli/)** — 八字紫微综合，排盘、用神、合会冲刑、大运流年
- **[起名](./qiming/)** — 八字用神 + 五格三才，全维度起名
- **[问卦](./wengua/)** — 智能路由奇门遁甲 / 六爻 / 黄历择日
- **[风水](./fengshui/)** — 八宅命卦 + 玄空飞星

## 为什么选灵机

- **AI 原生** — 不是简单的 prompt，而是完整的 Agent 工作流：参数收集 → API 排盘 → LLM 解读 → 结构化报告
- **计算精准** — 后端基于天文历算，支持真太阳时、夏令时校正、经纬度时区换算
- **版本管理** — 启动时自动检查更新，持续迭代
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
# 从 GitHub 安装
npx skills add ml8s/liki

# 或从官网安装
npx skills add https://liki.hk/skills
```

## Skills 详情

### 命理

八字紫微综合命理分析，先排八字（排盘、用神、合会冲刑），再排紫微（十二宫、星曜、四化），LLM 综合两套体系解读。覆盖大运流年全时间维度。

### 起名

结合八字用神与五格三才，提供全维度起名服务。流程：收集生辰 → 排八字取用神 → 命理决策 → 取字组合 → 五格评估 → 筛选 8 名 → 终检报告。

### 问卦

LLM 智能路由，根据问题自动选择最合适的方法：
- **择日**（哪天好）→ 黄历
- **问吉凶**（成败/得失/寻物）→ 六爻
- **问策略**（该不该做/往哪走）→ 奇门遁甲

### 风水

兼顾八宅与玄空两大流派。八宅命卦看人宅匹配，玄空飞星断时空吉凶，三元九运辅助趋势判断。

## 版本

当前稳定版 **1.9.0**。详见 [CHANGELOG.md](./CHANGELOG.md)。

## 相关链接

- 官网：[liki.hk](https://liki.hk)
- GitHub：[ml8s/liki](https://github.com/ml8s/liki)

## 免责声明

灵机（Liki）提供的命理分析基于中国传统文化视角，仅供研究与参考。命理结论**不构成**医疗诊断、法律建议、金融投资预测或人生重大决策的依据。请理性看待，保持自主积极的人生态度。

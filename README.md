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

**灵机（liki）** 是命理玄学 Skill 合集，`npx skills add ml8s/liki` 一条命令装完即用。包含四个子 skill：

- **[命理](./mingli/)** — 八字紫微综合，排盘、用神、合会冲刑、大运流年
- **[起名](./qiming/)** — 八字用神 + 五格三才，全维度起名
- **[问卦](./wengua/)** — 智能路由奇门遁甲 / 六爻 / 黄历择日
- **[风水](./fengshui/)** — 八宅命卦 + 玄空飞星

## 你能用它做什么

| 你想 | 用它 |
|------|------|
| 算八字、看紫微、查大运流年 | **命理** |
| 给宝宝或公司起名、改名 | **起名** |
| 选吉日、问吉凶、寻物、问方向 | **问卦** |
| 看住宅或办公室风水 | **风水** |

## 特性

- **Agent 工作流** — 参数收集 → API 排盘 → LLM 解读 → 结构化报告
- **天文历算后端** — 集成真太阳时、夏令时校正、经纬度时区换算
- **自动更新** — 启动时检查版本
- **隐私安全** — 出生信息发送至 API 实时运算，服务端不持久存储
- **即装即用** — 一条命令安装，安装后可直接调用

## 快速开始

```bash
npx skills add ml8s/liki     # 安装全家桶
```

然后在 AI 助手中直接对话：

> 帮我排个八字，1990年5月20日中午12点，北京出生，男

Agent 收集参数后，调用 `tianwen.time` 校正真太阳时，再排盘、取用神，输出结构化命理报告。

## 安装

```bash
# 从 GitHub 安装
npx skills add ml8s/liki

# 或从官网安装
npx skills add https://liki.hk/skills
```

## Skills 详情

### 命理

八字紫微综合命理分析，覆盖排盘、用神、合会冲刑、合盘、大运流年全维度。

> 算八字，1990-05-20 12:00 北京出生，男

### 起名

结合八字用神与五格三才，全维度起名服务。

> 宝宝起名，2024-03-15 10:00 广州出生，男，姓王

### 问卦

LLM 智能路由，根据问题自动选择最合适的方法：
- **择日**（哪天好）→ 黄历
- **问吉凶**（成败/得失/寻物）→ 六爻
- **问策略**（该不该做/往哪走）→ 奇门遁甲

> 明天适合搬家吗 ｜ 丢了东西能找到吗

### 风水

兼顾八宅与玄空两大流派。八宅命卦看人宅匹配，玄空飞星断时空吉凶，三元九运辅助趋势判断。

> 我家门朝东，这个户型风水怎么样

## 版本

当前稳定版 **1.10.0**。详见 [CHANGELOG.md](./CHANGELOG.md)。

## 相关链接

- 官网：[liki.hk](https://liki.hk)
- GitHub：[ml8s/liki](https://github.com/ml8s/liki)

## 免责声明

灵机（Liki）提供的命理分析基于中国传统文化视角，仅供研究与参考。命理结论**不构成**医疗诊断、法律建议、金融投资预测或人生重大决策的依据。请理性看待，保持自主积极的人生态度。

---

有问题或建议？[提 Issue](https://github.com/ml8s/liki/issues)

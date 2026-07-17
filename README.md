<p align="center">
  <img alt="liki.hk" src="https://img.shields.io/badge/liki.hk-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>liki.hk — AI 命理 Skills</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-官网-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="./README.en.md"><img src="https://img.shields.io/badge/English-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

## 简介

**liki.hk** 是 AI 命理 Skills，`npx skills add ml8s/liki` 一条命令装完即用。覆盖四大领域：

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

- **精准 API 排盘** — 集成天文历算后端，真太阳时/夏令时/经纬度时区校正
- **全领域覆盖** — 八字紫微起名问卦风水，一套 Skills 覆盖
- **智能路由** — 一句话自动分发到对应领域，无需指定 skill
- **隐私安全** — 出生信息发送至 API 实时运算，服务端不持久存储

## 快速开始

```bash
npx skills add ml8s/liki     # 安装整套
npx skills add https://liki.hk/   # 或从官网
```

然后在 AI 助手中直接对话：

## Skills 详情

### 命理

八字紫微综合命理分析，覆盖排盘、用神、格局判定、合会冲刑、合盘、大运流年、综合建议与历史事件校准。

> 算八字，1990-05-20 12:00 北京出生，男
> 看我今年的流年运势
> 帮我合一下八字，1984-06-15 北京，男 和 1986-09-20 上海，女
> 排个紫微斗数盘，1990-05-20 北京出生，男

### 起名

结合八字用神与五格三才的起名服务。

> 宝宝起名，2024-03-15 10:00 广州出生，男，姓王
> 帮我改个名字，1995-08-10 深圳出生，女，姓李
> 评测一下'张伟'这个名字

### 问卦

LLM 自动判断，根据问题选择对应方法：
- **择日**（哪天好）→ 黄历
- **问吉凶**（成败/得失/寻物）→ 六爻
- **问策略**（该不该做/往哪走）→ 奇门遁甲

> 明天适合搬家吗
> 丢了东西能找到吗
> 最近事业运怎么样
> 今天出门往哪个方向好

### 风水

兼顾八宅与玄空两大流派。八宅命卦看人宅匹配，玄空飞星断时空吉凶，三元九运辅助趋势判断。

> 我家门朝东，这个户型风水怎么样
> 这房子坐北朝南，风水如何
> 办公室在8楼，座位靠窗，要注意什么

## 版本

当前稳定版 **1.22.0**。详见 [CHANGELOG.md](./CHANGELOG.md)。

## 参考

本项目的设计参考了以下开源项目：

- [weizeW/mingli-skills](https://github.com/weizeW/mingli-skills) — 四维交叉验证框架
- [jinchenma94/bazi-skill](https://github.com/jinchenma94/bazi-skill) — 九本经典结构化摘要、历史事件校准
- [dzcmemory-web/bazi-ziwei-skill](https://github.com/dzcmemory-web/bazi-ziwei-skill) — 八字+紫微综合印证模式
- [yanouyuan-bit/bazi-roundtable](https://github.com/yanouyuan-bit/bazi-roundtable) — 多流派互审与结论强度标注
- [shizhilya/yuan](https://github.com/shizhilya/yuan) — 结论先行输出设计
- [ai-freer/fortune-skill](https://github.com/ai-freer/fortune-skill) — 三层计算架构
- [hhszzzz/taibu](https://github.com/hhszzzz/taibu) — Agent 友好设计
- [SylarLong/iztro](https://github.com/SylarLong/iztro) — 紫微斗数排盘引擎
- [2021291696/high-confidence-mingli-skill](https://github.com/2021291696/high-confidence-mingli-skill) — 置信度体系、人格画像推断
- [DestinyLinker/MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) — 命理评测参照

## 相关链接

- 官网：[liki.hk](https://liki.hk)
- GitHub：[ml8s/liki](https://github.com/ml8s/liki)

## 免责声明

liki.hk 提供的命理分析基于中国传统文化视角，仅供研究与参考。命理结论**不构成**医疗诊断、法律建议、金融投资预测或人生重大决策的依据。请理性看待，保持自主积极的人生态度。

---

有问题或建议？[提 Issue](https://github.com/ml8s/liki/issues)

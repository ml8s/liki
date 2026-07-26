<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-命理师的__Skill-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — 命理师的 Skill</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-在线演示-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="./README.en.md"><img src="https://img.shields.io/badge/English-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

**Liki** 是命理师的 Skill——引擎计算体系、三派方法论库、AI 辅助解读、可定制工具链。命理师用来执业，爱好者用来学习。`npx skills add ml8s/liki` 装完即用，在 AI 助手中完成八字、紫微、起名、六爻、奇门、择日、风水等全面命理分析。

覆盖 **8 个独立领域**，内置 **9 份方法论文档**，八字紫微**双体系交叉验证**，报告流程经 **generate → review → revise** 三阶段审查。

## 领域

| 领域 | 说明 |
|------|------|
| 八字 | 排盘、用神、格局、合会冲刑、合盘、大运流年 |
| 紫微 | 十二宫、四化、三方四正、特殊格局、大限流年 |
| 起名 | 八字用神 + 五格三才。支持外国人英文定中文姓 |
| 六爻 | 起卦→装卦→断卦 |
| 奇门 | 排盘、断事、择吉 |
| 黄历 | 择日查询 |
| 八宅 | 命卦配门主灶 |
| 玄空 | 飞星盘、三元九运、流年飞星 |

## 快速开始

```bash
npx skills add ml8s/liki
```

然后在 AI 助手中直接对话：

> 算八字，1990-05-20 12:00 北京出生，男
> 排个紫微盘，1988-03-15 上海出生，女
> 宝宝起名，2024-06-10 广州出生，男，姓陈
> 明天适合搬家吗

也可生成综合命书报告：

> 帮我出一份命书

AI 助手会完成八字+紫微全流程分析，输出综合论断 + 八字报告 + 紫微报告。

## 架构

```
Liki（本仓库）     → Skill（流程定义 + 方法论文档）
liki-engine        → 天文历算 API（[开源计算引擎](https://github.com/ml8s/liki-engine)）
liki.hk            → 基于 engine + skill 构建的在线演示
```

## 项目结构

```
bazi/       → 八字：SKILL + 报告格式 + knowledge 方法论文档
ziwei/      → 紫微：SKILL + 报告格式 + knowledge 方法论文档
naming/     → 起名：SKILL + 报告格式
liuyao/     → 六爻：SKILL + 报告格式
qimen/      → 奇门：SKILL + 报告格式
huangli/    → 黄历：SKILL + 报告格式
bazhai/     → 八宅：SKILL + 报告格式
xuankong/   → 玄空：SKILL + 报告格式
reports/    → 综合报告（命书 + 合盘，generate → review → revise）
```

## 特性

- **考时校准** — 出生时间不确定时，用已发生的人生事件反向验证时辰，成人排盘后强制考时
- **格局优先用神聚合** — 三家用神（格局/调候/扶抑）按优先级聚合：格局定框架、调候修正、扶抑微调，输出为空时跳过
- **天文历算** — 真太阳时/夏令时/经纬度时区校正，排盘不靠 LLM 心算
- **双系交叉验证** — 八字和紫微各自独立排盘，结论互相印证
- **9 份方法论文档** — yongshen/geju/tiaohou/wangshuai/hehui/gongwei/dayun/shishen/hepan
- **三阶段报告流程** — generate → review → revise，质量可控
- **语义版本管理** — 语义版本号 + CHANGELOG + VERSION 自检更新

## 版本

当前稳定版 **1.32.0**。详见 [CHANGELOG.md](./CHANGELOG.md)。

## 参考

设计参考了以下开源项目：

- [weizeW/mingli-skills](https://github.com/weizeW/mingli-skills) — 四维交叉验证框架，提示词工程化方法论参考
- [jinchenma94/bazi-skill](https://github.com/jinchenma94/bazi-skill) — 经典摘要参考文件设计、历史事件校准机制
- [dzcmemory-web/bazi-ziwei-skill](https://github.com/dzcmemory-web/bazi-ziwei-skill) — 八字+紫微综合印证模式
- [shizhilya/yuan](https://github.com/shizhilya/yuan) — 结论先行输出设计
- [hhszzzz/taibu](https://github.com/hhszzzz/taibu) — MCP 全栈命理，Agent 友好设计参考
- [SylarLong/iztro](https://github.com/SylarLong/iztro) — 紫微斗数排盘引擎
- [ai-freer/fortune-skill](https://github.com/ai-freer/fortune-skill) — 三层计算架构
- [yanouyuan-bit/bazi-roundtable](https://github.com/yanouyuan-bit/bazi-roundtable) — 多流派互审与结论强度标注理念
- [DestinyLinker/MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) — 命理评测参照
- [2021291696/high-confidence-mingli-skill](https://github.com/2021291696/high-confidence-mingli-skill) — 置信度体系、人格画像推断

## 协议

MIT。命理结论为传统文化视角，仅供参考，不构成专业建议。

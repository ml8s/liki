---
name: app-study
description: 学业判断 — 学历层次、学习能力、考试运
依赖域: bazi,ziwei
---

# 学业分析

## 依赖的领域知识

[必读] - bazi: tools/bazi/xueye.csv（xue_201-209 档位）
[必读] - bazi: domains/bazi/fangfa/shishen.md「十神组合」

[必读] - ziwei: domains/ziwei/fangfa/yingqi.md「学业紫微应期」
## 用户问法 → 领域信号（翻译表）

| 用户问题 | 对应领域信号 | 调用的决策表 |
|---------|-------------|-------------|
| 能读到什么学历 | 印星状态（查 tools/bazi/xueye.csv） | domains/bazi/fangfa/xueye.md「印星状态（查 tools/bazi/xueye.csv）」|
| 学习能力如何 | 印星+食伤组合 | domains/bazi/fangfa/shishen.md「十神」+ domains/bazi/fangfa/xueye.md |
| 考试运如何 | 大运流年对印星生扶/克制 | tools/bazi/xueye.csv|
| 适合学什么方向 | 食伤配印/官印相生/财坏印 | domains/bazi/fangfa/shishen.md「十神组合」|

## 📖 流程卡

⟳ 执行主干 Phase 1-2（时辰判定 + 排盘快照），此处不重复

第1步：印星状态验证（查 tools/bazi/xueye.csv 因子）
  → 调用 domains/bazi/fangfa/xueye.md「印星状态（查 tools/bazi/xueye.csv）」
  输出：□ 三关通过____关 印星状态____

第2步：大运辅助判断
  → 调用 tools/bazi/xueye.csv表
  输出：□ 当前大运____ 对印星____

第3步：学历等级定档
  → 调用 tools/bazi/xueye.csv（xue_201-209 档位）
  输出：□ 学历档____ 依据____

第4步：紫微验证【强制——断「具体细节」时必做，禁止只凭八字粗断】
  → 调用 domains/ziwei/fangfa/yingqi.md「学业」表
  → 调用 domains/ziwei/fangfa/liunian.md → 检查流年文昌/文曲是否被化科引动
  → 调用 domains/ziwei/fangfa/liunian.md → 流年命宫是否落父母宫/命宫/福德宫（考试有利宫位）
  输出：□ 文昌文曲____ 化科____ 流年命宫____ 与八字（一致/相反）____


📖 搜索 domains/bazi/SKILL.md → 获取排盘数据
## 📖 输出模板

### 长格式

```
【学业分析】
印星状态：{透干/藏支/不现}，{得令/失令}，{有根/无根}，印星状态{得令/被财克/有根}
大运影响：当前{印/官/食伤/财}运，{有利/不利}学业
学历倾向：{硕士以上/大学/专科/中学/其他}
学习特点：{食伤配印/官印相生/食伤无印} → {描述}
```

### 短格式

```
学业：印星{状态}，倾向{学历档}。当前大运{有利/不利}。
```

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 成年人问学业（非在校） | 学历定档按原局印星终生有效；大运影响按当前运判断深造/进修可能 |
| 原局无印星 | 查食伤（技术学习能力）+ 官杀（压力中学业） |

- **信号冲突 → 查 SKILL.md 执行主干（Phase 6 紫微交叉裁决）**：本卡与紫微/八字互斥时，回主干按「一票否决类/一般冲突」裁决，禁止本卡内自决。

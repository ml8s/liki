---
name: app-hepan
description: 合盘分析 — 感情匹配、婚姻合婚、关系评估
依赖域: bazi,ziwei
---

# 合盘分析

## 依赖的领域知识（预留跨域）

- bazi: domains/bazi/fangfa/hepan.md「合盘评估」
- bazi: domains/bazi/fangfa/gongwei.md「宫位论」
- bazi: domains/bazi/duanyu/shishen.md「十神组合」
- ziwei: domains/ziwei/SKILL.md「紫微合盘」（ziwei.bond）

## 用户问法 → 领域信号（翻译表）

| 用户问题 | 对应领域信号 | 调用的决策表 |
|---------|-------------|-------------|
| 两人合不合适 | 日干关系+夫妻宫互动+五行互补 | domains/bazi/fangfa/hepan.md |
| 能不能结婚 | 致命问题排除（年冲/日支冲/空亡）| domains/bazi/fangfa/hepan.md |
| 关系能持续多久 | 大运同步度 | domains/bazi/fangfa/hepan.md + domains/bazi/fangfa/dayun.md |
| 合盘紫微验证 | 夫妻宫/命宫互入 | ziwei.bond |

## 📖 流程卡

第0步：排盘（前置，双方）
  → 调用 bazi.chart(solar_time_a, gender_a) 排甲方八字
  → 调用 bazi.chart(solar_time_b, gender_b) 排乙方八字
  → 双方各调 bazi.fullchart(chart) 取全量
  → （紫微验证用）双方各调 ziwei.chart + ziwei.fullchart
  校验：双方四柱齐全，缺一方则补问

第1步：致命问题排除
  → 调用 domains/bazi/fangfa/hepan.md「综合评级」
  输出：□ 年冲？____ 日支冲？____ 配偶星空亡？____ 日干纯克？____

第2步：匹配评估
  输出：□ 日干关系____ 夫妻宫互动____ 五行互补度____ 十神传递____

第3步：可持续性
  → 调用 domains/bazi/fangfa/dayun.md（对比双方大运同步度）
  输出：□ 大运同步度____

第4步：紫微验证（交叉验证）
  双方都排紫微盘（ziwei.chart）→ 调用 ziwei.bond(a, b)
  → 看夫妻宫互动、命宫互入、禄马/四化入、五行生克
  输出：□ 紫微夫妻宫____ 命宫互入____ 生克____

第5步：综合评级
  输出：□ 良配/可配/慎配/不利 ____

📖 搜索 domains/bazi/fangfa/hepan.md → 读取合盘评估

## 📖 输出模板

### 一、双方命盘概况
各自日主、身强身弱、用神、关键特征。

### 二、五行互补
双方五行配置对比，哪方补哪方所缺，是否有过强五行重叠。

### 三、日主与十神关系
日主之间生克关系，十神互涉（对方对自己表现为何种十神），对相处模式的影响。

### 四、地支互动
合、冲、刑、害关系汇总。合多则相契相容，冲多则摩擦变动，逐条解读。

### 五、综合评价
有利因素与需注意的方面。性格互补或冲突的关键点，相处建议。

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 只提供了一方的八字 | 提示需要双方出生信息才能合盘 |
| 问合盘但只说"帮我看看我俩" | 提示提供双方出生信息 |

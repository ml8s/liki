---
name: app-study
description: 学业判断 — 学历层次、学习能力、考试运
依赖域: bazi,ziwei
---

# 学业分析

## 依赖的领域知识

- bazi: domains/bazi/duanyu/xueye.md「学历等级决策表」
- bazi: domains/bazi/duanyu/shishen.md「十神组合」

- ziwei: domains/ziwei/duanyu/yingqi.md「学业紫微应期」
## 用户问法 → 领域信号（翻译表）

| 用户问题 | 对应领域信号 | 调用的决策表 |
|---------|-------------|-------------|
| 能读到什么学历 | 印星三关状态 | domains/bazi/duanyu/xueye.md「印星三关」|
| 学习能力如何 | 印星+食伤组合 | domains/bazi/duanyu/shishen.md「十神」+ domains/bazi/duanyu/xueye.md |
| 考试运如何 | 大运流年对印星生扶/克制 | domains/bazi/duanyu/xueye.md「大运辅助」|
| 适合学什么方向 | 食伤配印/官印相生/财坏印 | domains/bazi/duanyu/shishen.md「十神组合」|

## 📖 流程卡

第0步：排盘（前置）
  → 调用 bazi.chart(solar_time, gender) 排八字 → 四柱/大运
  → 调用 bazi.fullchart(chart) 取全量（十神/藏干/神煞/空亡）
  → 调用 ziwei.chart(lunar, gender) 排紫微 → 十二宫
  → 调用 ziwei.fullchart(chart) 取全量（杂曜/长生/小限）
  校验：四柱齐全、十二宫齐全，缺失则补问出生时间

第1步：印星有效性三关验证
  → 调用 domains/bazi/duanyu/xueye.md「印星三关」
  输出：□ 三关通过____关 印星状态____

第2步：大运辅助判断
  → 换运年检测：当前年龄____ 是否在换运年±1年内？____ 若是→该年事件强度+1级
  → 调用 domains/bazi/duanyu/xueye.md「大运辅助」表
  输出：□ 当前大运____ 对印星____

第3步：学历等级定档
  → 调用 domains/bazi/duanyu/xueye.md「学历等级决策表」
  输出：□ 学历档____ 依据____

第4步：紫微验证【强制——断「具体细节」时必做，禁止只凭八字粗断】
  → 调用 domains/ziwei/duanyu/yingqi.md「学业」表
  → 调用 domains/ziwei/duanyu/liunian.md → 检查流年文昌/文曲是否被化科引动
  → 调用 domains/ziwei/duanyu/liunian.md → 流年命宫是否落父母宫/命宫/福德宫（考试有利宫位）
  输出：□ 文昌文曲____ 化科____ 流年命宫____ 与八字（一致/相反）____


📖 搜索 domains/bazi/SKILL.md → 获取排盘数据
#### 紫微→八字翻译

| 紫微信号 | 翻译结论 |
|---------|---------|
## 📖 输出模板

### 长格式

```
【学业分析】
印星状态：{透干/藏支/不现}，{得令/失令}，{有根/无根}，三关通过{0-3}关
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

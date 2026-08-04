---
name: app-marriage
description: 婚姻分析 — 何时结婚、婚姻质量、感情走向
依赖域: bazi,ziwei
---

# 婚姻分析

## 依赖的领域知识

- bazi: domains/bazi/duanyu/shishen.md「女命婚姻——官杀混杂判断」
- bazi: domains/bazi/fangfa/gongwei.md「宫位论」
- bazi: domains/bazi/fangfa/dayun.md「应期决策表」
- bazi: domains/bazi/fangfa/liuqin.md「六亲——配偶」

- ziwei: domains/ziwei/duanyu/yingqi.md「婚姻紫微应期」
## 用户问法 → 领域信号（翻译表）

| 用户问题 | 对应领域信号 | 调用的决策表 |
|---------|-------------|-------------|
| 什么时候结婚 | 官星/财星状态 + 大运窗口 + 流年引动 | domains/bazi/duanyu/shishen.md + domains/bazi/fangfa/dayun.md |
| 婚姻是否顺利 | 夫妻宫冲刑 + 官杀混杂取清 | domains/bazi/duanyu/shishen.md + domains/bazi/fangfa/gongwei.md |
| 对方是什么样的人 | 配偶星状态（正财/正官） + 十神组合 | domains/bazi/fangfa/liuqin.md + domains/bazi/duanyu/shishen.md |
| 会不会离婚 | 夫妻宫冲破 + 官星被合/被克 | domains/bazi/duanyu/shishen.md + domains/bazi/fangfa/gongwei.md |
| 什么时候遇到对象 | 大运流年引动配偶星 | domains/bazi/fangfa/dayun.md + domains/bazi/duanyu/shishen.md |

## 📖 流程卡

第0步：排盘（前置）
  → 调用 bazi.chart(solar_time, gender) 排八字 → 四柱/大运
  → 调用 bazi.fullchart(chart) 取全量（十神/藏干/神煞/空亡）
  → 调用 ziwei.chart(lunar, gender) 排紫微 → 十二宫
  → 调用 ziwei.fullchart(chart) 取全量（杂曜/长生/小限）
  校验：四柱齐全、十二宫齐全，缺失则补问出生时间

第1步：确定性别，定位配偶星（男看财星、女看官杀）
  → 调用 domains/bazi/duanyu/shishen.md「女命婚姻——官杀混杂判断」填表
  输出：□ 配偶星____ 官杀几位____ 有取清？____

第2步：夫妻宫（日支）检查
  → 调用 domains/bazi/fangfa/gongwei.md「日柱」
  输出：□ 日支冲刑____ 日支被合？____
  □ 合化类型（绊/化/动）____ 合化出什么五行____
  □ 合化结果：化出{五行} → 是{用神/忌神}
    化用神→吉应（感情顺/配偶得力），可判窗口年
    化忌神→凶应（感情纠葛/配偶拖累），大运引动亦不以吉论

第3步：大运流年婚姻窗口
  → 换运年检测：当前年龄____ 是否在换运年±1年内？____ 若是→该年事件强度+1级，优先于此运其他流年判断
  → 调用 domains/bazi/fangfa/dayun.md「应期决策表」筛选候选年份
  → **婚变流年（经典正断）**：女命正官（夫星）受克之年 = 婚变窗口——官坐绝地/被伤官冲克/被比劫争合之年优先。男命看正财（妻星）受克同理
  输出：□ 窗口年份____ 引动类型____

第4步：紫微验证【强制——断「具体细节」（何年结婚/离婚/配偶情况）时必做，禁止只凭八字粗断】
  → 调用 domains/ziwei/duanyu/yingqi.md「婚姻」表
  输出：□ 夫妻宫星曜____ 四化____ 桃花星____ 与八字（一致/相反）____

第5步：综合结论
  → 整合前四步输出
  输出：□ 婚姻质量____ 建议窗口年____


📖 搜索 domains/bazi/SKILL.md → 获取排盘数据
#### 紫微→八字翻译

| 紫微信号 | 翻译结论 |
|---------|---------|
## 📖 输出模板

### 长格式（专项分析时）

```
【婚姻分析】
配偶星状态：{官/财星透干/藏支/不现}，{清/浊}
夫妻宫：{冲/合/刑/害/无}，{吉利/需要磨合}
大运窗口：当前大运 {生扶/克制} 配偶星
建议窗口年：{年份}（{引动方式}）
风险提示：{如果有的话}
```

### 短格式（命盘报告引用）

```
婚姻：配偶星{状态}，夫妻宫{状态}。{1-2句总结}
```

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 用户未婚但问离婚 | 先问是否已有稳定对象，有则分析当前关系，无则分析命局倾向 |
| 男命问感情但原局无财星 | 食伤为财源，查食伤状态（食伤生财为隐性妻星） |
| 女命问婚姻但原局无官杀 | 查财星（财生官杀为隐性夫星），或大运引动 |
| 已离婚问再婚 | 查七杀是否清透，大运有无正官/正财出现 |
| 用户只提供了一个人的八字但想看合盘 | 提示需要双方出生信息，引导到 app/compatibility.md |

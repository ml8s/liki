---
name: app-career
description: 事业分析 — 职业方向、事业起伏、成就层次
依赖域: bazi,ziwei
---

# 事业分析

## 依赖的领域知识

- bazi: domains/bazi/duanyu/shiye.md「决策表」
- bazi: domains/bazi/fangfa/gongwei.md「宫位论」
- bazi: domains/bazi/fangfa/dayun.md「应期决策表」

- ziwei: domains/ziwei/duanyu/yingqi.md「事业紫微应期」
## 用户问法 → 领域信号（翻译表）

| 用户问题 | 对应领域信号 | 调用的决策表 |
|---------|-------------|-------------|
| 适合做什么行业 | 十神组合（官印/食伤/财） | domains/bazi/duanyu/shiye.md「职业类型」|
| 事业什么时候好 | 官杀/财运引动之年 | domains/bazi/duanyu/shiye.md + domains/bazi/fangfa/dayun.md |
| 适合创业还是打工 | 食伤生财 vs 官印相生 | domains/bazi/duanyu/shiye.md「十神方向」|
| 能到什么层次 | 官财透干定层次 | domains/bazi/duanyu/shiye.md「透干表」|

## 📖 流程卡

第0步：排盘（前置）
  → 调用 bazi.chart(solar_time, gender) 排八字 → 四柱/大运
  → 调用 bazi.fullchart(chart) 取全量（十神/藏干/神煞/空亡）
  → 调用 ziwei.chart(lunar, gender) 排紫微 → 十二宫
  → 调用 ziwei.fullchart(chart) 取全量（杂曜/长生/小限）
  校验：四柱齐全、十二宫齐全，缺失则补问出生时间

第1步：官财透干定层次
  → 调用 domains/bazi/duanyu/shiye.md「官财透干定层次」
  输出：□ 透干组合____ 事业档____

第2步：十神组合定方向
  → 调用 domains/bazi/duanyu/shiye.md「十神组合定方向」
  输出：□ 组合____ 职业方向____

第3步：大运起伏
  → 换运年检测：当前年龄____ 是否在换运年±1年内？____ 若是→该年事件强度+1级
  → 调用 domains/bazi/duanyu/shiye.md「大运影响」+ domains/bazi/fangfa/dayun.md「应期表」
  输出：□ 当前大运____ 窗口年____

第4步：紫微验证【强制——断「具体细节」时必做，禁止只凭八字粗断】
  → 调用 domains/ziwei/duanyu/yingqi.md「事业」表
  输出：□ 官禄宫星曜____ 化权/化禄____ 与八字（一致/相反）____


📖 搜索 domains/bazi/SKILL.md → 获取排盘数据

**紫微翻译**：详见 `domains/ziwei/duanyu/yingqi.md`「事业」行（官禄化权→上升/化忌→受阻/紫微天府→管理/太阴文曲女命→主内）

## 📖 输出模板

### 长格式

```
【事业分析】
事业层次：{管理+赚钱/有职无财/有钱无权/普通职员}
职业方向：{体制内/创业/技术/管理/经商}
当前大运：{官杀/财/印/食伤/比劫}运 → {上升/求财/稳定/变动/竞争}期
建议窗口：{年份}（{引动方式}）
```

### 短格式

```
事业：{职业方向}取向，当前{大运类型}运。{年份}有{上升/变动}机会。
```

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 官杀财星均不透 | 查地支藏干，或食伤/印星定方向 |
| 用户问换工作 | 结合大运流年引动判断 |

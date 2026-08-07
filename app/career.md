---
name: app-career
description: 事业分析 — 职业方向、事业起伏、成就层次
依赖域: bazi,ziwei
---

# 事业分析

## 依赖的领域知识

[必读] - bazi: domains/bazi/duanyu/shiye.md「决策表」
[必读] - bazi: domains/bazi/fangfa/gongwei.md「宫位论」
[必读] - bazi: domains/bazi/fangfa/dayun.md「应期决策表」

[必读] - ziwei: domains/ziwei/duanyu/yingqi.md「事业紫微应期」
## 用户问法 → 领域信号（翻译表）

| 用户问题 | 对应领域信号 | 调用的决策表 |
|---------|-------------|-------------|
| 适合做什么行业 | 十神组合（官印/食伤/财） | domains/bazi/duanyu/shiye.md「职业类型」|
| 事业什么时候好 | 官杀/财运引动之年 | domains/bazi/duanyu/shiye.md + domains/bazi/fangfa/dayun.md |
| 适合创业还是打工 | 食伤生财 vs 官印相生 | domains/bazi/duanyu/shiye.md「十神方向」|
| 能到什么层次 | 官财透干定层次 | domains/bazi/duanyu/shiye.md「透干表」|

## 📖 流程卡

⟳ 执行主干 Phase 1-2（时辰判定 + 排盘快照），此处不重复

第1步：官财透干定层次
  → 调用 domains/bazi/duanyu/shiye.md「官财透干定层次」
  输出：□ 透干组合____ 事业档____

第2步：十神组合定方向
  → 调用 domains/bazi/duanyu/shiye.md「十神组合定方向」
  输出：□ 组合____ 职业方向____

第3步：大运起伏
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

- **信号冲突 → 查 SKILL.md 执行主干（Phase 6 紫微交叉裁决）**：本卡与紫微/八字互斥时，回主干按「一票否决类/一般冲突」裁决，禁止本卡内自决。

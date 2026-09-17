---
name: app-wealth
description: 财运分析 — 财源类型、收入层次、风险提示
依赖域: bazi,ziwei
---

# 财运分析


> 工具报文：只使用 `bazi/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
## 依赖的领域知识

[必读] - bazi: bazi/domains/bazi/wealth.md「决策表」
[必读] - bazi: bazi/domains/bazi/shishen.md「十神」

[必读] - ziwei: bazi/domains/ziwei/yingqi.md「财运紫微应期」
## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
|---|---|---|---|
| 1 | 财星 | 分别发送 `TOOLS.md §3.1` 的 `query.十神`、`query.用神` | 正 / 偏财、透藏、是否为用 |
| 2 | 胜财能力 | 发送 `TOOLS.md §3.1 query.旺衰`；读取 `bazi/domains/bazi/wealth.md` | 能担 / 不能担 / 中和 |
| 3 | 风险与运势 | 分别发送 `TOOLS.md §3.1` 的 `query.大运`、`query.十神` | 当前运与比劫夺财风险 |
| 4 | 具体细节或流年 | 分别发送 `TOOLS.md §3.1` 的 `query.财帛`、`query.田宅`；用户问具体年份、应期或流年时发送 `§4.1 yearly_range.wealth` | 紫微财库与流年财信号 |

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 原局无财星 | 查食伤（食伤生财为隐性财源）|
| 财多身弱 | 标注为"财多身弱反为贫"，需大运帮身方可担财 |

## 输出模板

| 输出 | 内容 |
|---|---|
| 结论 | 财源类型、胜财能力与财富层次 |
| 依据 | 财星透藏、是否为用、身强弱、比劫风险 |
| 运势 | 当前大运与流年财信号 |
| 建议 | 得财方向与风险控制 |

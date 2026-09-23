---
name: app-wealth
description: 财运分析 — 财源类型、收入层次、风险提示
依赖域: bazi,ziwei
---

# 财运分析

> 工具报文：只使用 `natal/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] - bazi: natal/domains/bazi/wealth.md「决策表」
[必读] - bazi: natal/domains/bazi/shishen.md「十神」

[必读] - ziwei: natal/domains/ziwei/yingqi.md「财运紫微应期」

> **本卡的 `analyze_natal` / `analyze_periods` 调用均传 `topics=["wealth"]`，仅保留财运域断语。**

## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 财星 | 分别发送 `TOOLS.md §2` 的 `analyze_natal`、`analyze_natal` | 正 / 偏财、透藏、是否为用 |
| 2 | 胜财能力 | 发送 `TOOLS.md §2 analyze_natal`；读取 `natal/domains/bazi/wealth.md` | 能担 / 不能担 / 中和 |
| 3 | 风险与运势 | 分别发送 `TOOLS.md §2` 的 `analyze_periods(time_scope.type="decade")`、`analyze_natal` | 当前运与比劫夺财风险 |
| 4 | 具体细节或流年 | 分别发送 `TOOLS.md §2` 的 `analyze_natal(topics=["wealth"])`、`analyze_natal(topics=["property"])`；用户问具体年份、应期或流年时发送 `§4.1 analyze_periods` | 紫微财库与流年财信号 |

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 原局无财星 | 查食伤（食伤生财为隐性财源）|
| 财多身弱 | 标注为"财多身弱反为贫"，需大运帮身方可担财 |

## 输出模板

### 示例

```text
结论：财运中等偏上，偏财有根，比劫不旺。
依据：偏财得地透干（cai_105）；比劫不旺，夺财压力小。
运势：当前大运食伤运生财（2025-2034），2027 丁未流年财星得令。
建议：适合中长期投资；2026 丙午年比劫现，短期不宜大额投入。
```

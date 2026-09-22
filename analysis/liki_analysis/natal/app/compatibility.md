---
name: app-compatibility
description: 合盘分析 — 感情匹配、婚姻合婚、关系评估
依赖域: bazi,ziwei
---

# 合盘分析

> 工具报文：只使用 `natal/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] - bazi: natal/domains/bazi/hepan.md「双人盘事实合参」
[必读] - bazi: natal/domains/bazi/shishen.md「十神组合」
[必读] - 合盘工具：`compare_birth_charts(chart_ref_a, chart_ref_b)`（一次调用返回八字+紫微原始对照）

> **本卡的 `analyze_natal` / `analyze_periods` 调用均传 `topics=["marriage"]`，仅保留婚姻域断语。**

## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 双方出生信息 | 分别 `create_birth_chart` | 双方`chart_ref` |
| 2 | 原始对照 | `compare_birth_charts(chart_ref_a, chart_ref_b)` | 双方日主、夫妻宫、配偶星、干支关系与紫微宫位事实 |
| 3 | 证据整理 | 读取 `natal/domains/bazi/hepan.md` | 优势证据、摩擦证据、证据缺口 |
| 4 | 输出 | 按模板列共同点、差异与相处建议 | 参考结论 + 依据链 |

## 边界条件

| 异常场景 | 处理方式 |
| --------- | --------- |
| 只提供了一方的八字 | 提示需要双方出生信息才能合盘 |
| 问合盘但只说"帮我看看我俩" | 提示提供双方出生信息 |
| 缺一方时辰或缺性别 | 明示对应对象无法完整评估 |
| 单一冲、合、神煞 | 只作证据，不下关系成败结论 |

## 输出模板

不输出缘分分数或“良配 / 慎配”标签。

### 示例

```text
结论：日主互补，合盘中等偏吉。
八字合婚：甲方戊土，乙方壬水，戊土克壬水；甲方配偶星乙木透干得地。
紫微合参：甲方夫妻宫太阳（旺），乙方夫妻宫天同（庙）；双方宫位吉星并见。
建议：互补大于冲突；日主克泄关系在磨合期需关注（hun_303/hun_305）。
```

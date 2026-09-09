---
name: app-outcome
description: 问一件事能不能成 / 结果如何 — 默认六爻
依赖域: liuyao
---

# 问结果

## 路由规则

- 用户核心目标是判断“能不能成 / 会不会发生 / 结果如何”。
- 默认进入六爻；不自动加排奇门。
- 高风险问题先安全分流，不排盘。

## 流程

| 步骤 | 动作 | 产物 |
|---|---|---|
| 1 | 调 `divination_route(category=event_outcome)` 或由用户指定六爻 | route |
| 2 | 按 `app/liuyao-chart.md` 起卦并生成 Reading | casting + reading |
| 3 | 解释 snapshot / topic / timing / conditions | 倾向与依据 |
| 4 | report validate → audit → session create | 可追问结论 |

## 输出

```text
结论：一句话倾向与成败条件。
用神：用神、世应、旺衰。
关键动变：支持 / 阻碍。
应期：条件候选，不写确定日期。
建议：一个现实动作或核查条件。
```

---
name: app-decision
description: 问该不该做 / 方向 / 策略 / 时机 — 默认奇门
依赖域: qimen
---

# 问决策与方向

## 路由规则

- 用户核心目标是“该不该做”“往哪走”“怎么推进”“现在时机如何”。
- 默认进入奇门；不自动加排六爻。
- 医疗、法律、重大财务和人身安全问题先分流。

## 流程

| 步骤 | 动作 | 产物 |
|---|---|---|
| 1 | 调 `divination_route(category=strategy_decision / direction_space / timing_action)` 或由用户指定奇门 | route |
| 2 | 按 `app/qimen-chart.md` 调 `qimen_read` | chart + snapshot |
| 3 | 解释用神落宫、门星神、生克、空亡、马星、应期候选 | 方向与时机 |
| 4 | report validate；如有专占断语一并引用 | 审计后结论 |
| 5 | `qimen_session create` | 可追问原局 |

## 输出

```text
结论：一句话方向 / 时机 / 决策倾向。
用神：符号、落宫、旺衰、空亡马星。
态势：支持与阻碍来源。
时机：相关应期候选，不写确定日期。
建议：一个现实动作或核查条件。
```

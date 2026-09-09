---
name: app-decision
description: 问该不该做 / 方向 / 策略 / 时机 — 默认奇门
依赖域: qimen
---

# 问决策与方向

## 路由规则

- 用户核心目标是策略、方向或行动时机。
- 默认进入奇门；不自动加排六爻。
- 医疗、法律、重大财务和人身安全问题先分流。

## 流程

| 步骤 | 动作 |
|---|---|
| 1 | 按 `app/question.md` 判断为策略 / 方向 / 时机 |
| 2 | 按 `app/qimen-chart.md` 调 `qimen_snapshot` |
| 3 | 调 `qimen_ask(snapshot, message)` |
| 4 | 按 `app/qimen-chart.md` 输出 |

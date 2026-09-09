---
name: app-outcome
description: 问结果 / 能不能成 / 何时有结果 — 默认六爻
依赖域: liuyao
---

# 问结果

## 路由规则

- 用户核心目标是事件结果或事件应期。
- 默认进入六爻；不自动加排奇门。
- 高风险事项先安全分流。

## 流程

| 步骤 | 动作 |
|---|---|
| 1 | 按 `app/question.md` 判断为事件结果 |
| 2 | 按 `app/liuyao-chart.md` 调 `liuyao_snapshot` |
| 3 | 调 `liuyao_ask(snapshot, message)` |
| 4 | 按 `app/liuyao-chart.md` 输出 |

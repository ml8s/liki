---
name: app-date
description: 问哪天适合做事 — 默认黄历择日
依赖域: huangli
---

# 问日期

## 路由规则

- 用户核心目标是选择日期，而非判断事件成败。
- 默认进入黄历；不自动加排六爻或奇门。
- 高风险事项先安全分流。

## 流程

| 步骤 | 动作 | 产物 |
|---|---|---|
| 1 | 调 `divination_route(category=date_selection)` 或由用户指定黄历 | route |
| 2 | 调 `huangli_days`，传入 `question / event / start_date / end_date` | 候选日 |
| 3 | 读取 `domains/huangli/jiri.md`、`yiji.md` | 推荐与排除依据 |
| 4 | 输出推荐日、排除日和边界 | 择日结果 |

## 输出

```text
结论：推荐 1-3 个日期。
依据：建除、黄黑道、宜忌、冲煞、干支禁忌。
排除：明显不利日期及原因。
边界：未做八字合参时明确说明。
```

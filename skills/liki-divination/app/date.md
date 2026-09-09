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

| 步骤 | 动作 |
|---|---|
| 1 | 按 `app/question.md` 判断为择日 |
| 2 | 调 `huangli_days`，传入 `question / event / start_date / end_date` |
| 3 | 读取 `domains/huangli/jiri.md`、`yiji.md` |
| 4 | 输出推荐日、排除日和边界 |

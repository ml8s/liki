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
| 2 | 调 `huangli_days`，传入 `question / event / start_date / end_date / days`；事项适配由 engine 判定 |
| 3 | 读取 `domains/huangli/jiri.md`、`yiji.md` |
| 4 | 输出推荐日、排除日和边界 |

## 事项枚举

| 用户事项 | event |
|---|---|
| 嫁娶 / 婚礼 | `wedding` |
| 领证 / 订婚 | `engage` |
| 开业 / 开市 | `opening` |
| 签约 / 签合同 | `sign` |
| 搬家 / 入宅 | `move` |
| 出行 | `travel` |
| 动土 / 修造施工 | `build` / `renovation` |
| 考试 | `exam` |
| 就医 / 治病 | `medical` |
| 祭祀 | `sacrifice` |
| 扫除 / 清洁 | `cleaning` |
| 安床 | `bed_install` |
| 纳财 / 收账 | `income` |
| 丧葬 / 埋葬 | `funeral` |

表外事项不自行改写为相近 event；先向用户确认目标，再选择上表或仅输出通用黄历事实。

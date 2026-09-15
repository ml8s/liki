---
name: app-outcome
description: 问结果 / 能不能成 / 何时有结果 — 默认六爻
依赖域: liuyao
---

# 问结果


> 工具报文：只使用 `divination/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
## 路由规则

- 用户核心目标是事件结果或事件应期。
- 默认进入六爻；不自动加排奇门。
- 高风险事项正常走六爻流程；输出前附 `safety_advisory` 和专业建议。

## 流程

| 步骤 | 动作 |
|---|---|
| 1 | 按 `divination/app/question.md` 判断为事件结果 |
| 2 | 按 `divination/app/liuyao-snapshot.md` 发送 `TOOLS.md §1 liuyao_snapshot` 报文 |
| 3 | 发送 `TOOLS.md §2 liuyao_ask` 报文，`snapshot` 原样绑定 `$LIUYAO_SNAPSHOT` |
| 4 | 按 `divination/app/liuyao-snapshot.md` 输出 |

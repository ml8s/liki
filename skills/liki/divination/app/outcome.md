---
name: app-outcome
description: 问结果 / 能不能成 / 何时有结果 — 默认六爻
依赖域: liuyao
---

# 问结果

> 工具报文：只使用 `divination/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 路由规则

- 用户核心目标是事件结果或事件应期。
- 默认进入六爻；不自动加排奇门。
- 高风险事项正常走六爻流程；输出前附 `safety_advisory`。

## 流程

| 步骤 | 动作 |
| --- | --- |
| 1 | 按 `divination/app/question.md` 判断为事件结果 |
| 2 | 按 `divination/app/liuyao-snapshot.md` 发送 `TOOLS.md §1 六爻起卦` 报文 |
| 3 | 发送 `TOOLS.md §2 六爻追问` 报文，`snapshot` 原样绑定 `$LIUYAO_SNAPSHOT` |
| 4 | 按 `divination/app/liuyao-snapshot.md` 输出 |

## 边界条件

| 场景 | 处理 |
| --- | --- |
| 目标混合或缺失 | 先按 `question.md` 确认单一目标 |
| 高风险事项 | 正常执行，附 `safety_advisory` |
| 追问 | 复用同一 snapshot；新事件必须新建 snapshot |

## 输出模板

```text
结论：可成 / 难成 / 有条件成 / 时机未明。
依据：用神、世应、动爻和关键旺衰。
应期：只列 answer 引用的候选。
边界：传统六爻视角，不承诺现实结果。
```

### 示例

```text
结论：项目合作有利的倾向偏高，但资金到位时间偏晚。
用神：妻财（利益）伏藏，官鬼持世。
动爻：初爻子孙动，克制官鬼；应期有变数。
应期：answer 引用乙未日（约 7 天后），资金到账须再确认。
建议：签署前确认付款条款；7 天后再看进展。
```

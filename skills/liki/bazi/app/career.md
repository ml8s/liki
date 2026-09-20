---
name: app-career
description: 事业分析 — 职业方向、事业起伏、成就层次
依赖域: bazi,ziwei
---

# 事业分析

> 工具报文：只使用 `bazi/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] - bazi: bazi/domains/bazi/career.md「决策表」
[必读] - bazi: bazi/domains/bazi/dayun.md「应期决策表」

[必读] - ziwei: bazi/domains/ziwei/yingqi.md「事业紫微应期」

> **本卡所有 query / yearly_range 调用必须传 `domains=["事业"]`**，只保留 事业 域断语，排除跨域噪声。

## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 事业层次 | 分别发送 `TOOLS.md §3.1` 的 `query.十神`、`query.格局`、`query.用神`、`query.旺衰` | 透干组合、事业档、身强弱 |
| 2 | 职业方向 | 分别发送 `TOOLS.md §3.1` 的 `query.十神`、`query.神煞` | 十神取象与职业类型 |
| 3 | 事业起伏 | 发送 `TOOLS.md §3.1 query.大运` + 读取 `bazi/domains/bazi/dayun.md` | 当前运与窗口年 |
| 4 | 具体细节或流年 | 分别发送 `TOOLS.md §3.1` 的 `query.官禄`、`query.迁移`；用户问具体年份、应期或流年时发送 `§4.1 yearly_range.career` | 紫微事业信号与流年事件 |

## 边界条件

| 异常场景 | 处理方式 |
| --------- | --------- |
| 官杀财星均不透 | 查地支藏干，或食伤/印星定方向 |
| 用户问换工作 | 结合大运流年引动判断（发送 `TOOLS.md §4.1 yearly_range.career` 查流年应期） |

## 输出模板

### 示例

```text
结论：事业中上，职业方向技术管理。
依据：官杀透干有根（shi_103），身强能任官杀；食伤生财为才华出口。
运势：当前大运食伤运（2025-2034），食伤生财，事业窗口 2026-2028。
建议：走技术管理方向；流年丁未（2027）食伤生财旺，突破概率高。
```

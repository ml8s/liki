---
name: app-decision
description: 问该不该做 / 方向 / 策略 / 时机 — 默认奇门
依赖域: qimen
---

# 问决策与方向

> 工具报文：只使用 `divination/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 路由规则

- 用户核心目标是策略、方向或行动时机。
- 默认进入奇门；不自动加排六爻。
- 医疗、法律、重大财务和人身安全问题正常执行奇门流程；输出前附 `safety_advisory`。

## 流程

| 步骤 | 动作 |
| --- | --- |
| 1 | 按 `divination/app/question.md` 判断为策略 / 方向 / 时机 |
| 2 | 按 `divination/app/qimen-snapshot.md` 发送 `TOOLS.md §3 奇门排盘` 报文 |
| 3 | 发送 `TOOLS.md §4 奇门追问` 报文，`snapshot` 原样绑定 `$QIMEN_SNAPSHOT` |
| 4 | 按 `divination/app/qimen-snapshot.md` 输出 |

## 边界条件

| 场景 | 处理 |
| --- | --- |
| 缺地点 / 时间 | 按 `qimen-snapshot.md` 补城市、经度或使用服务端当前时间 |
| 高风险事项 | 正常执行，附 `safety_advisory` |
| 长期命局问题 | 转八字域，不用奇门替代终身命局 |

## 输出模板

```text
结论：可行 / 需调整 / 暂缓 / 时机未明。
依据：用神落宫、门星神和关键生克。
时机：answer 引用的应期候选。
边界：传统奇门视角，不承诺现实结果。
```

### 示例

```text
结论：不宜立即推进，等待下旬时机更稳。
用神：开门落震宫（囚），天芮（病星）同宫。
态势：开门囚为通道不畅；天芮为内部隐患。
时机：answer 引用下旬甲申日，开门转旺。
建议：本周内不签约；下旬重新评估。
```

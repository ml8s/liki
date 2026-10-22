# Liki 本命八字紫微 — 八字命盘 / 紫微斗数 / 合参合盘 / 考时

Liki 本命域：**八字侧与紫微侧都在本域内自举**——各自排盘、取因子、出断语，再由本域完成合参、合盘、考时与综合命书编排。工具、参数与字段以 MCP 自省和 schema 为准。

## 路由

执行侧按用户点名的体系选择：点名八字只跑八字侧，点名紫微只跑紫微侧，未点名时两侧都跑并合参（同向时综合，冲突时并列两侧证据）。

| 用户场景 | 入口卡 | 历史验证聚焦 |
| --- | --- | --- |
| 无明确问题 / 要看命书 | `references/natal-app-mingshu.md` | 全盘大事 |
| 婚姻与感情 | `references/natal-app-marriage.md` | 结婚 / 婚变年 |
| 事业方向与起伏 | `references/natal-app-career.md` | 升职 / 离职 / 创业年 |
| 财源与风险 | `references/natal-app-wealth.md` | 得财 / 破财年 |
| 健康倾向 | `references/natal-app-health.md` | 大病 / 手术年 |
| 学业与考试 | `references/natal-app-study.md` | 升学 / 毕业 / 中断年 |
| 性格与体型 | `references/natal-app-personality.md` | 状态题可免验证 |
| 父母六亲与子女 | `references/natal-app-family.md` | 父母灾 / 子女出生年 |
| 合盘（双人）| `references/natal-app-compatibility.md` | 婚运窗口 |
| 考时（时辰存疑）| `references/natal-app-calibration.md` | — |

场景卡只写分体系信息需求（合参所需）、合参要求与输出；信息如何取得对应用层透明（由执行拓扑决定），本域完成合参与命书编排。

## 硬边界

- 断语以工具输出为准（规则真值表）；方法论卡只用于解释依据，不自行编造。
- 跨体系合参与断语取舍见对应场景卡的「双盘合参」节。
- 部分失败降级：单体系信息不可得时，降级为单体系结论并明示缺口与置信影响；不得以单体系冒充合参。

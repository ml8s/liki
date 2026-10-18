# Liki 本命八字紫微 — 八字命盘 / 紫微斗数 / 合参合盘 / 考时

Liki 本命域：**八字侧与紫微侧都在本域内自举**——各自排盘、取因子、出断语，再由本域完成合参、合盘、考时与综合命书编排。

## 工具契约

agent 经 `engine` MCP（排盘）与 `counsel` MCP（判断）使用两侧能力；工具与参数由 MCP 自省和 schema 提供，链路与输出契约见 `references/TOOLS.md`。

## 路由

执行侧按用户点名的体系选择：点名八字只跑八字侧，点名紫微只跑紫微侧，未点名时两侧都跑并合参（同向时综合，冲突时并列两侧证据）。

| 用户场景 | 入口卡 | 历史验证聚焦 |
| --- | --- | --- |
| 无明确问题 / 要看命书 | `references/app/mingshu.md` | 全盘大事 |
| 婚姻与感情 | `references/app/marriage.md` | 结婚 / 婚变年 |
| 事业方向与起伏 | `references/app/career.md` | 升职 / 离职 / 创业年 |
| 财源与风险 | `references/app/wealth.md` | 得财 / 破财年 |
| 健康倾向 | `references/app/health.md` | 大病 / 手术年 |
| 学业与考试 | `references/app/study.md` | 升学 / 毕业 / 中断年 |
| 性格与体型 | `references/app/personality.md` | 状态题可免验证 |
| 父母六亲与子女 | `references/app/family.md` | 父母灾 / 子女出生年 |
| 合盘（双人）| `references/app/compatibility.md` | 婚运窗口 |
| 考时（时辰存疑）| 八字侧 `references/domains/bazi/calibration.md` 编排，紫微侧 `references/domains/ziwei/calibration.md` 交叉验证 | — |

场景卡只写场景查询、裁决与输出；工具链与断语字段见 `references/TOOLS.md`，全局硬边界见 `SKILL.md` 的硬边界一节。

## 硬边界

- 断语以工具输出为准（规则真值表）；方法论卡只用于解释依据，不自行编造。
- 跨体系合参与断语取舍按 `references/TOOLS.md`「合参与裁决」执行；各场景的取值见对应场景卡的「双盘合参」节。

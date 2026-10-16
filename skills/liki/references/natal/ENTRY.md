# Liki 本命八字紫微 — 八字命盘 / 紫微斗数 / 合参合盘 / 考时

Liki 本命域：**八字侧与紫微侧都在本域内自举**——各自排盘、取因子、出断语，再由本域完成合参、合盘、考时与综合命书编排。

## 工具契约

agent 经 `engine` MCP（排盘）与 `counsel` MCP（判断）使用两侧工具；工具、链路与输出契约见 `references/natal/TOOLS.md`。

## 路由

执行侧按用户点名的体系选择：点名八字只跑八字侧，点名紫微只跑紫微侧，未点名时两侧都跑并合参（同向时综合，冲突时并列两侧证据）。

| 用户场景 | 入口卡 |
| --- | --- |
| 主题分析（婚姻、事业、财运、健康、学业、性格、六亲、流年）| `references/natal/app/README.md` 中的对应主题卡 |
| 快速扫描 / 综合命书 | `references/natal/app/mingshu.md`、`references/natal/app/mingshu-full.md` |
| 合盘（双人）| `references/natal/app/compatibility.md` |
| 考时（时辰存疑）| 八字侧 `references/natal/domains/bazi/calibration.md` 编排，紫微侧 `references/natal/domains/ziwei/calibration.md` 交叉验证 |

## 硬边界

- 断语以工具输出为准（规则真值表）；方法论卡只用于解释依据，不自行编造。

# Natal 领域入口

Liki 本命域统一处理八字、紫微、合参、大运/大限、流年、合盘和定盘。本域 agent 不直接调用 RPC。进入本域后先读取本入口，再按用户场景读取一张 App 卡和当前任务需要的 domain 卡；保持上下文聚焦。

工具契约：`natal/TOOLS.md`（工具参数由 `liki-analysis` MCP 连接器 schema 提供）。

- 调用：通过 `liki-analysis` MCP 连接器使用 `create_birth_chart` 等工具（参数见 TOOLS.md）
- 响应：成功读 `data`；失败读 `error.code` / `error.message`
- 工具层启动时校验 engine 版本和必需 MCP 能力；不满足即 fail closed

## 领域工具

| 场景 | 工具 | 产物 |
| --- | --- | --- |
| 出生信息 | `create_birth_chart` | 不可变 `chart_ref` 与盘摘要 |
| 本命问题 | `analyze_natal` | 按 topic 收敛后的本命断语 |
| 大运/大限/流年 | `analyze_periods` | 按 topic 与时间层收敛后的断语 |
| 双人关系 | `compare_birth_charts` | 八字与紫微合盘结果 |
| 时辰考订 | `calibrate_birth_time` | 候选盘的人生事件信号 |

城市解析、真太阳时校正和引擎编排都在工具层内部。用户给具体时刻时收集出生城市或经度；用户只给时辰时按既定时辰排盘；工具提示时辰临界时先向用户复核，必要时进入定盘流程。

## 流程

| 步骤 | 条件 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 有出生信息 | `create_birth_chart` | `chart_ref` |
| 2 | 有明确人生问题 | 将中文问题映射为受控 topic | `topics` 参数 |
| 3 | 问本命 | `analyze_natal` | 本命断语 |
| 4 | 问应期、流年或限运 | `analyze_periods` | 时间层断语 |
| 5 | 双人关系 | 分别建盘后 `compare_birth_charts` | 合盘结果 |

同一会话复用 `chart_ref`；多领域问题先处理主场景，再用次 topic 佐证。信号冲突读取 `natal/domains/bazi/caijue.md` 裁决；有真实事件时用 3-5 段已发生时段验证。

## 边界

- 城市解析失败时按 `LOCATION_NOT_RESOLVED` 追问附近较大城市或经度。
- 已定时辰采用 `solar_time_correction=off`，避免二次校正。
- 时辰临界信号只触发复核；临界本身不构成吉凶结论。
- 流年区间含端点，单次最多 120 年。

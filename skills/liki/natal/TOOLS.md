# Natal 能力与编排（专家执行）

本域为编排层，实际能力由对应专家执行（排盘与判断分离）：

- **八字**：`liki-bazi`（排盘 + 本命判断 + 应期判断 + 合盘 + 考时）
- **紫微**：`liki-ziwei`（历法换算 + 排盘 + 本命判断 + 应期判断 + 合盘 + 考时）

agent 应路由到对应专家执行；工具经 `tools/list` 自举（schema 自描述），按能力域调用，不依赖具体工具名。成功响应读 `data`；失败响应读 `error.code` 和 `error.message`。

## 标准流程

1. **排盘**：出生信息 → 本命盘（八字：四柱/大运；紫微：公历 → 农历 → 十二宫/大限）。
2. **取因子**：本命盘 → 因子快照（含防篡改摘要）。
3. **本命判断**：因子快照 + 问题域 → 本命断语。
4. **应期判断**：因子快照 + 时间层 + 问题域 → 大运/大限/流年断语。

## 跨域编排

| 场景 | 交给 | 说明 |
| --- | --- | --- |
| 八字看命/流年 | `liki-bazi` | 排盘 + 本命/应期判断 |
| 紫微看命/流年 | `liki-ziwei` | 历法换算 + 排盘 + 本命/应期判断 |
| 合盘（双人）| 双方专家 | 双人合盘 |
| 考时（时辰存疑）| 双方专家 | 按专家 `calibration.md` 用应期判断编排 |
| 合参/综合命书 | 双方专家 | 两侧断语同向综合（不臆造）|

## 输出契约（断语公共字段）

`assertion_id`、`side`、`topic`、`method`、`time_scope`、`event_type`、`event`、`conclusion`、`source`、`evidence`。`side` 是 `bazi` / `ziwei`。

Topic 是受控人生问题闭集：`adversity`、`appearance`、`career`、`chart_structure`、`children`、`family`、`health`、`marriage`、`mental`、`origin`、`personality`、`property`、`relocation`、`social`、`study`、`wealth`。

时间层由 `time_scope` 的 `type` 表达：`current_year`、`year`、`year_range`、`current_decade`、`decade`。年份为**年号**（八字公历立春界 / 紫微农历春节界，年号级一致），不做日期换算。年份区间含端点，最多 120 年。
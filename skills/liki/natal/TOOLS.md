# Natal 工具契约（专家执行）

本域为编排层，实际工具由对应专家执行（排盘与判断分离）：

- **八字**：`liki-bazi` → `engine-bazi` 排盘 `bazi_chart`（出生信息 → `chart`），`judgment-bazi` 判断 `compute_factors` / `natal_query` / `period_query`
- **紫微**：`liki-ziwei` → `engine-aux` 历法 `tianwen_time`（公历 → 农历 `lunar`），`engine-ziwei` 排盘 `ziwei_chart(lunar, gender)`，`judgment-ziwei` 判断

agent 应路由到对应专家执行；详细工具流程见各专家 `SKILL.md`。成功响应读 `data`；失败响应读 `error.code` 和 `error.message`。

## 标准流程（分域）

1. **排盘（engine）**：`bazi_chart`（八字）或 `tianwen_time` + `ziwei_chart`（紫微）→ `chart`
2. **取因子（judgment）**：`compute_factors(chart)` → `factors` + `factors_digest` + `context`
3. **本命断语**：`natal_query(factors, topics, context)` → 本命断语
4. **应期断语**：`period_query(factors, time_scope, topics, chart)` → 大运/大限/流年断语

## 跨域编排

| 场景 | 交给 | 说明 |
| --- | --- | --- |
| 八字看命/流年 | `liki-bazi` | engine-bazi 排盘 + judgment-bazi 判断 |
| 紫微看命/流年 | `liki-ziwei` | engine-aux + engine-ziwei + judgment-ziwei |
| 合盘（双人）| 双方专家 | engine `bazi_bond` / `ziwei_bond` |
| 考时（时辰存疑）| 双方专家 | 按专家 `calibration.md` 用 `period_query` 编排 |
| 合参/综合命书 | 双方专家 | 两侧断语同向综合（不臆造）|

## 响应契约

| 工具 | `data` 契约 |
| --- | --- |
| `bazi_chart` / `ziwei_chart` | `{nian/yue/ri/shi 或宫位, da_yun/daxian, gender, ...}`（engine 排盘）|
| `compute_factors` | `{factors, factors_digest, context}` |
| `natal_query` | `{assertions}` |
| `period_query` | `{chart, query, periods}` |

断语公共字段：`assertion_id`、`side`、`topic`、`method`、`time_scope`、`event_type`、`event`、`conclusion`、`source`、`evidence`。`side` 是 `bazi` / `ziwei`。

Topic 是受控人生问题闭集：`adversity`、`appearance`、`career`、`chart_structure`、`children`、`family`、`health`、`marriage`、`mental`、`origin`、`personality`、`property`、`relocation`、`social`、`study`、`wealth`。

时间层由 `time_scope` 的 `type` 表达：`current_year`、`year`、`year_range`、`current_decade`、`decade`。年份为**年号**（八字公历立春界 / 紫微农历春节界，年号级一致），不做日期换算。年份区间含端点，最多 120 年。
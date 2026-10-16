# Natal 工具与判断

- **八字侧**（`side` = `bazi`）：排盘 + 本命判断 + 应期判断（含合盘、考时）。
- **紫微侧**（`side` = `ziwei`）：历法换算 + 排盘 + 本命判断 + 应期判断（含合盘、考时）。

用 `side` 区分两侧；卡内动作的 `八字侧` / `紫微侧` 即以对应 `side` 调用，两侧并列时各自调用后再合参。

卡内未限定 topic 的判断调用取该卡领域；结论只保留该域断语，其他域信号仅用于解释机制。

## 标准流程

1. **排盘**：出生信息 → 本命盘（八字：四柱/大运；紫微：公历 → 农历 → 十二宫/大限）。
2. **取因子**：本命盘 → 因子快照（含防篡改摘要）。
3. **本命判断**：因子快照 + 问题域 + `side` → 本命断语。
4. **应期判断**：因子快照 + 时间层 + 问题域 + `side` → 大运/大限/流年断语。

## 输出契约（断语公共字段）

`assertion_id`、`side`、`topic`、`method`、`time_scope`、`event_type`、`event`、`conclusion`、`source`、`evidence`。`side` 是 `bazi` / `ziwei`。

Topic 是受控人生问题闭集（`键`＝中文标签，卡内 `本命判断（标签）` 即取该 topic）：`marriage`＝婚姻、`family`＝家庭、`children`＝子女、`career`＝事业、`wealth`＝财运、`health`＝健康、`mental`＝精神、`study`＝学业、`personality`＝性格、`appearance`＝外貌、`origin`＝出身、`relocation`＝迁移、`social`＝人际、`property`＝房产、`adversity`＝风险、`chart_structure`＝盘面结构。

时间层由 `time_scope` 的 `type` 表达：`current_year`、`year`、`year_range`、`current_decade`、`decade`；卡内 `应期判断（十年期）` 取十年期层，无限定则按问题域取年/十年层。年份为**年号**（八字公历立春界 / 紫微农历春节界，年号级一致），不做日期换算。年份区间含端点，最多 120 年。

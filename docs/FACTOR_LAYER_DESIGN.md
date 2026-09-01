# Liki 因子分层方案

> 范畴：`liki-bazi`（唯一使用真值表/因子层的 skill；divination/fengshui/naming 为纯 Markdown LLM 驱动，无此层）。

## 一、当前分层

```
pan（engine 排盘输出，领域事实，只读）
  │
  ▼ factors（唯一 Python 求值层）
       · 算子(_op/_liu_op/_atomic)从 pan 直读，签名无中间上下文参数
       · _base_ctx_from_pan(chart)：聚合 shishen/wuxing；FactorContext 只读挂载，不写回 pan
       · 真值表复合求值 factors.csv
       · _domain_facts_from_pan：透传稳定领域事实
  │
  ▼ snap = 完整领域快照（孑合字段表，bool/int/str/结构）{八字, 紫微, context}
       · 本命快照按当前调用生命周期生成；NatalContext 支持多年显式复用
  │
  ▼ 断语长表（assertions.csv + assertion_conditions.csv，等值匹配 snap）→ 断语结论
```

即：`pan → factors → snap → 断语长表`。无独立提取层、无中间 fac 参数——factors 为单一求值层，assertions 为断语长表层。

## 二、缓存

- **ctx 层**：`_base_ctx_from_pan` 聚合的 shishen/wuxing 挂在 `FactorContext`；一次求值中算子被多次触发只聚合一次。
- **snap 层**：本命快照按调用方编排复用；不做全局 pan 内容缓存；用户会话中反复查询不同断语域（婚姻/事业/财运…）时，同一内容 pan 只算一次完整 snap，后续域直接匹配。
- 两层缓存都**不写入公共 pan**：调用方不会看到实现细节，也不存在把缓存误存成命理数据的问题。
- `current_year != 0`（大运域等）的快照不缓存，按年重算（量小）。

## 三、snap 内容

- **断语因子**：factors.csv 真值表求值出的本命/流年因子（八字/紫微各自）。
- **领域事实透传**（`_domain_facts_from_pan`）：
  - 八字侧：四柱纳音/藏干/旬空/自合/魁罡/自合名、三元、旬空、三奇贵人、拱夹、纳音生克、大运。
  - 紫微侧：宫位（含星曜/杂曜/博士/岁前/小限/紫微/长生）、局数、命主、身主、命宫、身宫、空宫、年干、时支、紫微星位。
  - context：性别、公历出生、农历出生、当前年份。
- 领域事实属稳定命理概念（入口入 snap），供断语与未来扩展直接引用。

## 四、主要函数

- `paipan.full_paipan(...)` → 完整 pan。
- `factors.prepare_natal_context(pan)` → 多年流年复用的本命求值上下文与八字快照（不写回 pan）。
- `factors.evaluate_snap_from_pan(pan, current_year=0)` → snap（含领域透传，本命快照单次求值内由 FactorContext/NatalContext 复用）。
- `factors.evaluate_liunian_snap_from_pan(pan, liunian_pan, year, natal_context=None)` → 流年快照。
- `factors.evaluate_factors(gender, chart, shushi, current_year)` / `evaluate_liunian_factors(...)` → 双盘因子快照（算子从 pan 直读）。
- `duanyu.query(rule, pan)` / `yearly_range(pan, start, end, rules)` → 断语（先校验完整 pan；
  `yearly_range` 单次最多 120 年并复用本命上下文）。

## 五、验证

- 全量 Python 单测：**239 passed, 1 skipped**。
- `check.sh`（check_schema/check_docs/版本一致）全绿。
- 多域查询按当前调用生成 snap；多年 yearly_range/calibrate 通过 NatalContext 复用本命上下文。

## 六、可后续推进

- **枚举拆列归并**（字段表化，低优先）：`夫妻宫状态`(5列→1字段)；`含[patterns]`(42列→格局字段)；`宫含`(247列)不建议（多属布尔归属判断）。
- `_wuxing_from_pan`（3行）可内联进 `_base_ctx_from_pan`（收益极微）。

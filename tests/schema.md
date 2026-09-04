# 断语库 schema（常量表 + 因子真值表 + 断语真值表）

## 架构与数据流

```text
paipan.py
  full_paipan / liunian / city_coords / bond
  → pan（引擎排盘结果；返回与入口均经 pan_schema 校验）

pan_schema.py
  validate_natal_pan(pan)
  → 拒绝快照、裁剪盘、手工半截盘

operators_natal.py
  _op / _base_ctx_from_pan
operators_liunian.py
  _liu_op / target helpers

factors.py
  prepare_natal_context(pan)
  evaluate_factors(gender, pan, shushi, current_year)
  evaluate_liunian_factors(gender, pan, liunian_data, zw_liunian_data, year, shushi)
  evaluate_snap_from_pan(pan, current_year)
  evaluate_liunian_snap_from_pan(pan, liunian_pan, year, natal_context)
  → snap（八字快照 + 紫微快照 + context）

factor_tables.py
  factors.csv / factors_liunian.csv 长表
  → OR group + AND term + direct 表达式

domain_snapshot.py
  project_domain_facts(pan)
  → reserved 稳定领域事实（契约由 domain_snapshot_contract.json 保护）

yearly_eval.py
  resolve_rules / yearly_snapshot / query_year_rules

assertion_store.py
  assertion long-table index

duanyu.py
  query / query_yearly / yearly_range
  → {八字: [...], 紫微: [...], 合参: [...]}
```

`calibrate.py` 是独立考时工具，编排 `paipan → factors → duanyu`。候选 `correct=true` 必须提供 longitude；`correct=false` 表示已明确时辰，longitude 可省略。
`detail=true` 时聚合输出各流年域的机械 evidence；`detail=false` 只保留精简断语。

域契约：`query` 只接受本命域；`query_yearly` / `yearly_range` 只接受流年域。`yingqi` 是流年域，必须通过 `yearly_range` 查询。
`query(year=...)` 只允许 `大运 / 大限` 限运域；省略 year 时由服务端当前时间推导。
限运域 query 结果附带 `current_year / current_year_source`；显式传入 year 时 source 为 `specified`。

## 硬约束

1. **规则只由表定义**：代码做机械解析与求值，命理成员关系在 `constants.json`，因子组合在 CSV。
2. **单一常量来源**：基础闭集、十神大类、六亲角色、五行生克、干支关系与紫微星组都在 `constants.json`。
3. **真值表驱动**：`factors.csv` 生成本命快照，`factors_liunian.csv` 生成流年快照。
4. **双术数分开 + 显式合参**：八字表用八字因子，紫微表用紫微因子；八字×紫微同一条件支撑的结论必须写入 `side=common` 表，并在双盘合并快照上匹配。`性别` 是排盘上下文，不计入因子。
5. **流年 target 显式**：因子名和算子参数必须显式包含目标，如 `流年配偶星透`、`流年克[财星]`；禁止隐藏 target。
6. **结论层分离**：程序只输出断语，综合解释由 LLM 完成。
7. **真值表禁止空定义**：断语行必须有约束；因子行必须有直通表达式或条件列。
8. **标量值域闭集**：标量因子的断语约束值必须来自 `constants.json` 对应闭集。
9. **计算错误必须暴露**：因子求值、必需断语表读取和 `time.now` 失败不得降级为 0 / 空表 / 本地时间。
10. **完整 pan 契约**：`query` / `yearly_range` 只接受 `full_paipan` 完整返回的 pan；快照、裁剪盘和手工半截盘必须显式报错。
11. **批量跨度上限**：`yearly_range` 单次起止年含端点跨度最多 120 年。
12. **领域模型不可按当前消费裁剪**：`factors*.csv` 与 domain snapshot reserved facts 是稳定命理领域模型；当前断语未消费不是删除依据。

## 常量表分层

| 层 | 键 | 说明 |
|---|---|---|
| 基础闭集 | 五行 / 天干 / 地支 / 十神 / 十二长生 / 旺衰状态 / 紫微主星 / 紫微煞星 / 紫微六吉星 / 紫微文星 | 原子词汇 |
| 十神大类 | 官杀 / 印星 / 财星 / 食伤 / 比劫 | 对十个原子十神做完整、不交叉 partition |
| 六亲角色 | 配偶星 / 子女星 / 父星 / 母星 / 日主 | 角色引用十神大类或原子十神，按性别解析 |
| 事件宫位 | 配偶星 / 父星 / 母星 / 子女星 / 官杀 / 财星 / 日主 | 配偶、官杀、财、日主为日支；父宫、母宫为年支；子女宫为时支 |
| 关系表 | 天干五合、地支六合、三合、三会、六冲、六害、三刑、旬空 | 稳定关系闭集 |
| 算子语义配置 | 十神大类日主关系、十神旺弱规则、夫妻宫关系优先级、官杀取清、用忌映射、格局十神、算子柱位、紫微四化条件、紫微亮度分组、紫微星组别名、紫微星曜特殊值、紫微宫位特殊条件 | operator 只做机械查表求值，命理成员与优先级不写入代码 |
| 流年机械配置 | 事件宫位、事件宫位默认、四柱、四柱序号、干支来源、关系取合类型、关系取冲类型、三合半合、天干、地支、旬空起点、流年宫名后缀 | 流年目标解析、来源解析、柱位、成合 / 取冲方式与旬空顺序由表驱动 |
| 流年年界 | 八字干支年、紫微农历年、具体日期使用说明 | `yearly_range.year_basis` 的领域语义由常量表提供 |
| 结构闭集 | 性别闭集、四柱、大限段数 | pan 结构校验与考时入参复用同一常量来源 |
| 命理侧 | bazi / ziwei / common 代码、八字 / 紫微 / 合参标签、快照与断言侧范围 | 断言索引、快照生成、流年聚合与考时聚合复用同一输出契约 |

## 因子真值表

| 列 | 含义 |
|---|---|
| 字段 | 含义 |
|---|---|
| factor_id | 快照键 |
| shushi | bazi / ziwei / common |
| group_id | 同因子内 OR 分组 |
| term_index | 同 group 内 AND 序号 |
| kind | direct / condition / factor_ref |
| expression | 算子表达式或因子引用 |
| expected | 期望值 |
| basis | 命理依据，不参与求值 |

条件列中的 `引用本命[X]` 是跨层引用：读取本命八字快照因子 X，不是当前层原子事实。

参数使用规则：

- 具体十神参数 → 提取原子因子。
- 十神大类或六亲角色参数 → 类级复合因子。
- 流年 target 参数必须显式。
- 流年三合/三会约束必须三方齐备成局；两支半合不按完整合会因子命中。
- `原语直通[...,任意]` 可返回字符串标量。

## 断语长表

断语元数据：

```csv
assertion_id,rule,side,领域,事件类型,时间层,事件,结论,依据,经典依据
```

断语条件：

```csv
assertion_id,condition_group_id,factor,expected
```

规则：

- `assertion_id` 全局唯一。
- `side ∈ {bazi, ziwei, common}`；`rule` 是命理域。
- `领域`、`事件类型` 是受控闭集；`时间层` 必须与本命/大限/流年 rule 一致。
- 同一 `assertion_id` 的同一 `condition_group_id` 内多条 condition 是 AND。
- 同一 `assertion_id` 的不同 `condition_group_id` 之间是 OR。
- `expected` 运行时按整数优先解析，失败保留字符串。
- loader 名称格式为 `{side}_{rule}`，例如 `bazi_格局`。

匹配规则：任一条件组内全部约束与快照值精确相等即命中；不同条件组为 OR。

## 校验

`python3 tests/check_schema.py` 校验：

1. 约束键存在于本命因子、流年因子或排盘上下文；
2. 流年表不引用不可达本命因子，必须先写入 `本命X` 流年引用因子；
3. 单侧表无跨术数死条件，合参必须显式使用 `side=common`；
4. 无空条件组、重复 ID、同组重复因子或互斥条件；
5. 标量约束列具有字符串直通；
6. 生产表不得含评测 case、迭代阶段、旧内部路径等过程残留。

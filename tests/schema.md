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
  → {八字: [...], 紫微: [...]}
```

`calibrate.py` 是独立考时工具，编排 `paipan → factors → duanyu`。

域契约：`query` 只接受本命域；`query_yearly` / `yearly_range` 只接受流年域。`yingqi` 是流年域，必须通过 `yearly_range` 查询。

## 硬约束

1. **规则只由表定义**：代码做机械解析与求值，命理成员关系在 `constants.json`，因子组合在 CSV。
2. **单一常量来源**：基础闭集、十神大类、六亲角色、五行生克、干支关系与紫微星组都在 `constants.json`。
3. **真值表驱动**：`factors.csv` 生成本命快照，`factors_liunian.csv` 生成流年快照。
4. **双术数分开**：八字表用八字因子，紫微表用紫微因子；`性别` 是排盘上下文，不计入因子。
5. **流年 target 显式**：因子名和算子参数必须显式包含目标，如 `流年配偶星透`、`流年克[财星]`；禁止隐藏 target。
6. **结论层分离**：程序只输出断语，综合解释由 LLM 完成。
7. **真值表禁止空定义**：断语行必须有约束；因子行必须有直通表达式或条件列。
8. **标量值域闭集**：标量因子的断语约束值必须来自 `constants.json` 对应闭集。
9. **计算错误必须暴露**：因子求值、必需断语表读取和 `time.now` 失败不得降级为 0 / 空表 / 本地时间。
10. **完整 pan 契约**：`query` / `yearly_range` 只接受 `full_paipan` 完整返回的 pan；快照、裁剪盘和手工半截盘必须显式报错。
11. **批量跨度上限**：`yearly_range` 单次起止年含端点跨度最多 120 年。

## 常量表分层

| 层 | 键 | 说明 |
|---|---|---|
| 基础闭集 | 五行 / 天干 / 地支 / 十神 / 十二长生 / 旺衰状态 / 紫微主星 / 紫微煞星 / 紫微六吉星 / 紫微文星 | 原子词汇 |
| 十神大类 | 官杀 / 印星 / 财星 / 食伤 / 比劫 | 对十个原子十神做完整、不交叉 partition |
| 六亲角色 | 配偶星 / 子女星 / 父星 / 母星 / 日主 | 角色引用十神大类或原子十神，按性别解析 |
| 事件宫位 | 配偶星 / 父星 / 母星 / 子女星 / 官杀 / 财星 / 日主 | 配偶、官杀、财、日主为日支；父宫、母宫为年支；子女宫为时支 |
| 关系表 | 天干五合、地支六合、三合、三会、六冲、六害、三刑、旬空 | 稳定关系闭集 |

## 因子真值表

| 列 | 含义 |
|---|---|
| 字段 | 含义 |
|---|---|
| factor_id | 快照键 |
| shushi | bazi / ziwei |
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
assertion_id,rule,side,事件,结论,依据,经典原文
```

断语条件：

```csv
assertion_id,factor,expected
```

规则：

- `assertion_id` 全局唯一。
- `side ∈ {bazi, ziwei}`；`rule` 是命理域。
- 同一 `assertion_id` 的多条 condition 是 AND。
- `expected` 运行时按整数优先解析，失败保留字符串。
- loader 名称格式为 `{side}_{rule}`，例如 `bazi_格局`。

匹配规则：约束值与快照值精确相等，全部约束同时成立才命中。

## 校验

`python3 tests/check_schema.py` 校验：

1. 约束键存在于本命因子、流年因子或排盘上下文；
2. 流年表不引用不可达本命因子；
3. 无跨术数死条件；
4. 无死列、重复 ID、重复因子定义；
5. 标量约束列具有字符串直通；
6. `引用本命[X]` 的 X 是本命因子。

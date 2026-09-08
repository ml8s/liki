# 因子与断言模型契约

因子清单的唯一事实源是长表：

- `skills/liki-bazi/tools/factors/factors.csv`
- `skills/liki-bazi/tools/factors/factors_liunian.csv`

本文只记录分层、统计与不可变契约，不复制因子行。CSV 是因子清单唯一事实源，文档、测试或快照不得再维护第二份逐因子清单。

## 模块分层

```text
paipan.py           排盘 RPC 适配
factor_context.py   只读求值上下文
factor_tables.py    长表加载与 OR/AND 分组
operators_natal.py  本命机械算子
operators_liunian.py 流年机械算子
factors.py          因子门面：snap 编排
domain_snapshot.py  稳定领域事实只读投影
assertion_store.py  断言长表索引
duanyu.py           断言门面
```

数据流为：

```text
pan → factors → snap → assertions
```

## 统计口径

| 口径 | 数量 | 事实源 |
|---|---:|---|
| 本命因子 | 459 | `factors.csv` |
| 本命八字因子 | 179 | `factors.csv` |
| 本命紫微因子 | 277 | `factors.csv` |
| 本命定义组 | 496 | `factors.csv` |
| 本命数据行 | 559 | `factors.csv` |
| 本命直通原子 | 46 | `factors.csv` |
| 本命提取原子 | 295 | `factors.csv` |
| 本命复合因子 | 118 | `factors.csv` |
| 流年因子 | 101 | `factors_liunian.csv` |
| 流年八字因子 | 69 | `factors_liunian.csv` |
| 流年紫微因子 | 32 | `factors_liunian.csv` |
| 流年定义组 | 105 | `factors_liunian.csv` |
| 流年数据行 | 101 | `factors_liunian.csv` |
| 流年直通原子 | 4 | `factors_liunian.csv` |
| 流年提取原子 | 59 | `factors_liunian.csv` |
| 流年复合因子 | 38 | `factors_liunian.csv` |

统计由 `tests/test_factor_model.py` 与表数据同步校验，不需要手工维护第二份清单。

## 查询契约

`calibrate.py` 是独立考时工具，编排 `paipan → factors → duanyu`。候选 `correct=true` 必须提供 longitude；`correct=false` 表示已明确时辰，longitude 可省略。`detail=true` 输出机械 evidence；`detail=false` 只保留断语。

- `duanyu.query` 只接受本命域；`query_yearly` / `yearly_range` 只接受流年域，`yingqi` 必须通过流年查询。
- `query(year=...)` 只允许 `大运 / 大限` 限运域；省略 year 时由服务端当前时间推导。
- 限运域结果附带 `current_year / current_year_source`；显式传 year 时 source 为 `specified`。
- `query` / `yearly_range` 只接受 `full_paipan` 完整返回的 pan，拒绝快照、裁剪盘和手工半截盘。
- `yearly_range` 单次起止年含端点跨度最多 120 年。

## 长表结构

两张因子表均使用同一字段：

| 字段 | 含义 |
|---|---|
| `factor_id` | 因子名，也是快照键 |
| `shushi` | `bazi` / `ziwei` / `common` |
| `group_id` | 同因子内 OR 分组 |
| `term_index` | 同组内 AND 序号 |
| `kind` | `direct` / `condition` / `factor_ref` |
| `expression` | 算子表达式或被引用因子 |
| `expected` | 期望值；`direct` 行省略 |
| `basis` | 命理依据，不参与求值 |

求值规则：

- 同一 `factor_id + group_id` 内多行 AND。
- 同一 `factor_id` 的不同 `group_id` 之间 OR。
- `factor_ref` 只能引用存在且无环的因子。
- 流年引用本命因子必须显式使用 `引用本命[...]`。
- `factor_names` 只裁剪本次输出投影，不裁剪因子表领域模型。

参数使用规则：

- 具体十神参数提取原子因子；十神大类或六亲角色参数提取类级复合因子。
- 流年 target 参数必须显式。
- 流年三合 / 三会约束必须三方齐备成局；两支半合不按完整合会因子命中。
- `原语直通[...,任意]` 可返回字符串标量。

## 常量与闭集

十神、五行、干支、十二长生、紫微星曜、宫位、神煞、关系表与命理侧闭集均以 `skills/liki-bazi/tools/constants.json` 为唯一事实源。代码只做机械查表、解析与求值，不内置命理结论。

| 层 | 内容 | 说明 |
|---|---|---|
| 基础闭集 | 五行、天干、地支、十神、十二长生、旺衰状态、紫微主星、煞星、六吉星、文星 | 原子词汇 |
| 十神大类 | 官杀、印星、财星、食伤、比劫 | 十个原子十神的完整不交叉 partition |
| 六亲角色 | 配偶星、子女星、父星、母星、日主 | 引用十神大类或原子十神，按性别解析 |
| 事件宫位 | 配偶星、父星、母星、子女星、官杀、财星、日主 | 对应日支、年支或时支 |
| 关系表 | 天干五合、地支六合、三合、三会、六冲、六害、三刑、旬空 | 稳定关系闭集 |
| 算子语义 | 旺弱规则、宫位关系、用忌映射、格局十神、紫微四化与亮度分组等 | operator 只做机械查表 |
| 流年机械 | 事件宫位、干支来源、关系类型、三合半合、旬空起点、流年宫名 | 流年 target 与求值由表驱动 |
| 流年年界 | 八字干支年、紫微农历年 | `yearly_range.year_basis` 领域语义 |
| 结构闭集 | 性别、四柱、大限段数 | pan 校验与考时入参复用 |
| 命理侧 | bazi / ziwei / common 与输出标签 | 快照、断言与考时聚合复用 |

## 断言长表

断语元数据字段：

```csv
assertion_id,rule,side,领域,事件类型,时间层,事件,结论,依据,经典依据
```

断言条件字段：

```csv
assertion_id,condition_group_id,factor,expected
```

- `assertion_id` 全局唯一；`side ∈ {bazi, ziwei, common}`。
- `领域`、`事件类型` 是受控闭集；`时间层` 必须与本命、大限或流年 rule 一致。
- 同一 `assertion_id + condition_group_id` 内多行 AND；不同 `condition_group_id` 之间 OR。
- `expected` 按整数优先解析，失败保留字符串。
- loader 名称格式为 `{side}_{rule}`，例如 `bazi_格局`。
- 跨术数条件必须写入 `side=common`，并在双盘合并快照上匹配。

## 稳定领域事实

`domain_snapshot.py` 从完整 `pan` 只读投影稳定命理事实，入口和目标字段由 `domain_snapshot_contract.json` 锁定：

- 八字：纳音、藏干、旬空、自合、魁罡、三元、三奇、拱夹、大运等。
- 紫微：宫位与星曜、局数、命主、身主、命身宫、空宫、大限等。
- 上下文：性别、公历出生、农历出生、当前年份。

这些领域事实与因子表同属稳定领域模型；当前断语是否消费不作为删除依据。

## 排盘上下文

| 上下文 | 值域 | 消费方式 |
|---|---|---|
| 性别 | `male` / `female` | 排盘上下文，不算因子；断言匹配时并入对应视图 |
| 当前年份 | 公历年 | 仅本命「大运」域查询当前大运时使用 |
| 公历出生 | 字符串 | `full_paipan` 出生事实透传 |
| 农历出生 | 字符串 | `full_paipan` 出生事实透传 |

## 缓存与复用

- `FactorContext` 只缓存一次求值内的基础聚合，不写回公共 `pan`。
- `NatalContext` 只复用本命基础聚合和本命快照；流年盘、流年快照与公共 `pan` 保持只读。
- 本命快照按调用生命周期生成，不做全局 pan 内容缓存。

## 硬约束

1. 命理规则只由 `constants.json`、因子长表与断言长表定义；代码只做机械解析、查表与求值。
2. 因子行必须有直通表达式或条件；断言行必须有条件组。
3. 标量因子的断语约束值必须来自对应常量闭集。
4. 因子求值、断言表读取和 `time.now` 失败不得降级为 0、空表或本地时间。
5. 生产表不得包含评测 case、迭代阶段或旧内部路径等过程残留。

`python3 tests/check_schema.py` 校验约束键、流年可达性、单侧表边界、条件组完整性、标量闭集与生产纯度。

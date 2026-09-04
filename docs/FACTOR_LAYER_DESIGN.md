# Liki 因子分层方案

> 范畴：`liki-bazi`（唯一使用真值表/因子层的 skill；divination/fengshui 不使用此层，naming 走 qiming RPC）。

## 一、当前分层

```
pan（engine 排盘输出，领域事实，只读）
  │
  ▼ factors（唯一 Python 求值层）
       · 算子(_op/_liu_op/_atomic)从 pan 直读，签名无中间上下文参数
       · _base_ctx_from_pan(chart)：聚合 shishen/wuxing；FactorContext 只读挂载，不写回 pan
       · 真值表复合求值 factors.csv
       · project_domain_facts：透传稳定领域事实
  │
  ▼ snap = 完整领域快照（结构化字段表，bool/int/str/结构）{八字, 紫微, context}
       · 本命快照按当前调用生命周期生成；NatalContext 支持多年显式复用
  │
  ▼ 断语长表（assertions.csv + assertion_conditions.csv，等值匹配 snap）→ 断语结论
```

即：`pan → factors → snap → 断语长表`。factors 是唯一 Python 求值层，assertions 是断语长表层。

## 二、缓存

- **ctx 层**：`_base_ctx_from_pan` 聚合的 shishen/wuxing 挂在 `FactorContext`；一次求值中算子被多次触发只聚合一次。
- **snap 层**：本命快照按调用方编排复用；不做全局 pan 内容缓存；同一会话中重复查询不同断语域时复用已生成的完整 snap。
- 两层缓存都**不写入公共 pan**：调用方不会看到实现细节，也不存在把缓存误存成命理数据的问题。
- `current_year != 0`（大运域等）的快照不缓存，按年重算（量小）。

## 三、snap 内容

- **断语因子**：factors.csv 真值表求值出的本命/流年因子（八字/紫微各自）。
- **领域事实透传**（`project_domain_facts`）：
  - 八字侧：四柱纳音/藏干/旬空/自合/魁罡/自合名、三元、旬空、三奇贵人、拱夹、纳音生克、大运。
  - 紫微侧：宫位（含星曜/杂曜/博士/岁前/小限/紫微/长生）、局数、命主、身主、命宫、身宫、空宫、年干、时支、紫微星位、大限。
  - context：性别、公历出生、农历出生、当前年份。
- 领域事实属稳定命理概念（入口入 snap），供断语与未来扩展直接引用。
- 因子表与领域事实同属稳定领域模型；当前断语未消费不作为裁剪依据。
- 领域事实的四柱成员、源字段、目标事实名与顶层引用全部由
   `domain_snapshot_contract.json` 定义；`domain_snapshot.py` 只做只读投影。

## 四、主要函数

- `paipan.full_paipan(...)` → 完整 pan。
- `factors.prepare_natal_context(pan)` → 多年流年复用的本命求值上下文与八字快照（不写回 pan）。
- `factors.evaluate_snap_from_pan(pan, current_year=0, sides=None, factor_names=None)` → snap（含领域透传；query 按单侧域与断言因子闭包求精简集，默认双侧全量）。
- `factors.evaluate_liunian_snap_from_pan(pan, liunian_pan, year, natal_context=None)` → 流年快照。
- 流年快照支持 factor_names 裁剪；yearly_range 根据展开后的断言域计算流年因子闭包，并同步裁剪复用的本命引用因子。
- `factor_names` 只控制本次计算投影；`factors*.csv` 仍完整保留命理领域模型。
- `factors.evaluate_factors(gender, chart, shushi, current_year, factor_names=None)` / `evaluate_liunian_factors(...)` → 双盘因子快照（算子从 pan 直读）。
- `duanyu.query(rule, pan, year=None)` / `yearly_range(pan, start, end, rules)` → 断语（先校验完整 pan；
  `year` 仅限大运/大限域；`yearly_range.rules` 必填、单次最多 120 年并复用本命上下文）。
- 断言输出分为八字、紫微、合参三层；common 断言在双盘合并快照上匹配。
- 命理侧代码、输出标签、快照侧与断言侧范围由 `constants.json → 命理侧` 统一定义；
  断言索引、快照生成、流年聚合与考时聚合复用同一契约。

## 五、验证

- 全量 Python 单测：**295 passed**。
- `check.sh`（check_schema/check_docs/版本一致）全绿。
- 多域查询按当前调用生成 snap；多年 yearly_range/calibrate 通过 NatalContext 复用本命上下文。

## 六、P2 表契约

- 因子长表按文件路径做进程内只读缓存；缓存 key 由加载层私有管理。
- 一个因子只归属 `bazi`、`ziwei` 或 `common`；单侧因子集合不相交，common 因子可消费双盘机械事实。
- `factor_ref` 目标必须存在；因子引用图必须无环。
- `direct` 行的表达式是唯一值来源，必须省略 `expected`。
- 真值表对同一原子表达式只求值一次；本命/流年算子注册表不相交。
- 算子中的命理成员、优先级、角色映射、格局映射、宫位特殊条件与干支来源统一来自
   `constants.json`；代码只做机械解析、查表与求值，不内置命理结论。
- `test_domain_semantic_contracts.py` 静态检查 operator 源码，防止十神、五行、干支、
   星曜与关系成员等闭集回流代码。
- `NatalContext` 只复用本命基础聚合和本命快照；流年盘、snap 和公共 pan 保持只读。

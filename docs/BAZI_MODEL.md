# 八字领域模型（八紫双盘）

`liki-bazi` 的领域名 `bazi` 是八紫双盘同参的稳定领域包名：它覆盖八字四柱、紫微斗数，以及两条体系之间的 `common` 合参边界。狭义“八字”只在 `shushi=bazi`、RPC 名称或具体柱盘上下文中表示四柱子系统。

本文记录该领域的稳定对象、分层边界和求值契约。因子与断言是本命断语的条件机制，不是领域模型的顶层名字。

## 1. 领域对象

| 对象 | 含义 | 边界 |
|---|---|---|
| Pan | 一次排盘得到的只读领域上下文，包含八字、紫微或双盘事实 | 只接受 `full_paipan` 的完整返回；拒绝裁剪盘、手工半截盘和旧快照 |
| Side | 命理侧别：`bazi` / `ziwei` / `common` | 八字与紫微各自求值；跨体系断言只能在双盘合并上下文命中 |
| AtomicFact | engine 输出的确定性命理事实 | Python 只读取，不复算十神、五行、关系、宫位和亮度规则 |
| Factor | 可被断言引用的稳定条件因子 | 因子清单由长表唯一定义，OR / AND 分组与引用关系可校验 |
| Assertion | 一条带条件组和经典依据的断语 | 不直接内嵌命理推导；只能引用可达因子 |
| TimeLayer | 本命、大运 / 大限、流年等时间层 | 跨层引用必须显式；限运域必须携带目标年份上下文 |

## 2. 模块分层

```text
paipan.py           排盘 RPC 适配
factor_context.py   只读求值上下文
factor_tables.py    长表加载与 OR/AND 分组
operators_natal.py  本命机械算子
operators_liunian.py 流年机械算子
factors.py          因子门面：snap 编排
natal_projection.py  稳定本命事实只读投影
assertion_store.py  断言长表索引
duanyu.py           断言门面
```

数据流为：

```text
pan → factors → snap → assertions
```

## 3. Engine 原子事实

命理结论先由 engine 计算为原子事实，Python 算子只读取和组合，不复算：

| 字段 | 含义 |
|---|---|
| `full.nian/yue/ri/shi.shen_sha` | engine 神煞事实；勾绞、元辰按年干阴阳 × 性别定向，学堂 / 词馆按年命纳音同气正位 |
| `full.shen_sha_school` | 天乙、桃花、驿马、华盖、将星、劫煞、灾煞采用年参照与日参照并集；不隐含单一流派 |
| `full.yong_shen.tiao_hou` | 《穷通宝鉴》主用 / 辅用表候选；`primary / secondary` 记录表内天干在四柱的透干、藏支与所在柱，并投影其宫支参与的合会冲刑；不机械反推忌神，原文额外条件与扶抑 / 格局上下文另核 |
| `full.yong_shen.fu_yi.basis` | 扶抑强弱表输入：日主根型、月令旺相、印比数量、日主藏干根，以及日主 / 印比 / 日支 / 日主根相关的关系投影；动态再权衡仍须整体合参 |
| `full.yong_shen.ge_ju.structure` | 月令格神、克格神与生格神三类原子状态：透干柱、藏干根、季节旺相与组合强弱；同时投影与三类五行相关的天干五合和地支合会冲刑 |
| `full.yong_shen.ge_ju` | 月令藏干透干格局候选、格神、十神与来源；`structure` 是候选证据，不输出成格 / 败格 / 救应终判，不得冒充完整子平结论 |
| `full.yong_shen.*.relation_facts[].targets` | 关系投影的角色闭集；同一地支多藏干可命中多个角色，必须全部保留，不得按第一个藏干截断 |
| `full.lu_roots` | 透干十神得十干禄 |
| `full.gan_he` | 全部两两天干五合事实，含 `adjacent / separated / remote` 柱距与 `contested` 争合候选标记 |
| `full.san_he / full.san_hui` | 完整三合局与三会方，含成员支与所在柱；结构投影只消费这些 engine 定位事实 |
| `full.relation_groups` | 去重后的完整关系组：紧邻天干五合进 `gan_he`，隔位 / 远隔进 `gan_he_candidate`；地支六合 / 三合 / 三会 / 六冲 / 六害 / 三刑按全组或成对规则 |
| `full.ten_god_states` | 十神透干 / 藏支、数量、通根、得令与组合旺弱 |
| `full.element_states` | 五行季节旺弱、组合旺弱、生克方向与克者旺弱 |
| `full.da_yun.steps[].rooted / root_refs` | 大运干通根事实与坐支本气 / 原局藏干证据 |
| `full.atomic_facts.day_master_element` | 日主五行 |
| `full.atomic_facts.month_longevity` | 日主在月支的十二长生 |
| `full.atomic_facts.year_stem_ten_god / month_main_ten_god / hour_stem_ten_god` | 年干、月令本气、时干十神 |
| `full.atomic_facts.pattern_god_transparent` | 月令格局的格神是否透干 |
| `full.atomic_facts.pillar_punishments` | 年 / 月 / 日 / 时柱是否参与命局相刑 |
| `full.atomic_facts.officer_killing_cleaned` | 官杀混杂后七杀被合 / 冲取清 |
| `full.atomic_facts.wealth_tomb_present` | 日干财星墓库现于四柱 |
| `full.atomic_facts.wealth_star_in_tomb` | 财星与墓库同柱 |
| `full.atomic_facts.spouse_palace_state` | 夫妻宫冲 / 合 / 刑 / 害 / 静 |
| `full.atomic_facts.day_branch_type` | 日支桃花 / 驿马 / 墓库 |
| `full.atomic_facts.year_officer_killing` | 年柱官杀攻身 |
| `full.da_yun` | 按节气顺逆与晚子时口径计算起运；23:00–24:00 属子时 |
| `ziwei.palace_facts` | 紫微宫位星曜、星组、四化、亮度与主星数量原子事实 |
| `ziwei.patterns` | 紫微格局候选；`金灿光辉` 限太阳独坐命宫午宫，`日月反背` 只取命宫三方四正文言明的日戌月辰 / 日子月午结构，不得泛化为任意双落陷 |

流年层同样由 `bazi.liunian` 输出原子事实：

| 字段 | 含义 |
|---|---|
| `liunian.shen_sha` | 流年动态神煞与值年神煞；丧门取太岁前二位、吊客取后二位 |
| `atomic_facts.controls_elements` | 流年干或支所克五行 |
| `atomic_facts.controls_targets` | 日主 / 财官印食等语义目标是否受流年干支克 |
| `atomic_facts.unfavorable_gan / unfavorable_branch` | 流年干 / 支为扶抑忌神 |
| `atomic_facts.wealth_breaks_seal` | 印星流年被本支克 |
| `atomic_facts.combinations` | 三合 / 三会 / 三刑 / 半合，是否包含流年支 |
| `atomic_facts.year_branch_relations` | 流年支与原局各支的合 / 冲 / 刑 / 害关系 |
| `atomic_facts.year_branch_controlled_by` | 原局旺相五行所克流年支 |
| `atomic_facts.year_gan_controls_day_gan / day_gan_controls_year_gan` | 流年干与日干方向性相克 |
| `atomic_facts.gan_combines` | 流年干与日干五合 |
| `atomic_facts.dayun_gan_combines` | 大运干与流年干五合 |
| `atomic_facts.dayun_zhi_clashes_year` | 大运支与流年支六冲 |
| `atomic_facts.day_void_branches` | 日柱旬空支 |
| `atomic_facts.year_equals_day_pillar / dayun_equals_year_pillar` | 流年与日柱 / 大运干支伏吟 |
| `atomic_facts.year_gan_equals_natal_year_gan` | 流年干伏吟年柱 |

这些字段的命理口径由 engine 单测与 `tests/fixtures/domain_oracle/` 锁定。Python 新增因子时不得重新实现上述推导。

紫微 `宫含` 算子只对 `ziwei.palace_facts` 做 palace / kind / target / star exact match；Python 不再遍历宫位、推导四化落宫、解释亮度分组或计算主星数量。
紫微盘显式输出 `school`：当前闰月口径为 iztro 兼容的“前十五日本月、后十五日次月”。这不是《紫微斗数全书》闰月按下月口径；两派不得混写。
本命宫名使用 engine 闭集：`命宫、兄弟、夫妻、子女、财帛、疾厄、迁移、仆役、官禄、田宅、福德、父母`；除命宫外不追加“宫”字。`任意` 只表示跨全部本命宫匹配，不是宫名。
流年紫微同样消费 engine 宫名闭集；`流曜入宫` 与 `流年宫化` 不做带“宫”字后的显示别名适配。

`ten_god_states.transparent / hidden` 只描述该具体十神；`rooted / timely` 按其五行判定。`ten_god_states.strength` 的口径是：得令，或该具体十神透干且五行通根为 `strong`；失令且该十神不透、五行无根为 `weak`；其余为 `neutral`。同五行的另一十神透干，不会把本十神错误升级为透干有根。`element_states.season_strength` 只表达月令旺相休囚死；`element_states.strength` 是五行聚合态；`controller_strength` 表达受克目标的克者是否旺相。Python 只读取这些结论，不再维护五行生克、天干五行、得令状态或十神旺弱规则表。

## 4. 统计口径

| 口径 | 数量 | 事实源 |
|---|---:|---|
| 本命因子 | 475 | `factors.csv` |
| 本命八字因子 | 198 | `factors.csv` |
| 本命紫微因子 | 277 | `factors.csv` |
| 本命定义组 | 515 | `factors.csv` |
| 本命数据行 | 586 | `factors.csv` |
| 本命直通原子 | 50 | `factors.csv` |
| 本命提取原子 | 310 | `factors.csv` |
| 本命复合因子 | 115 | `factors.csv` |
| 流年因子 | 101 | `factors_liunian.csv` |
| 流年八字因子 | 69 | `factors_liunian.csv` |
| 流年紫微因子 | 32 | `factors_liunian.csv` |
| 流年定义组 | 105 | `factors_liunian.csv` |
| 流年数据行 | 101 | `factors_liunian.csv` |
| 流年直通原子 | 4 | `factors_liunian.csv` |
| 流年提取原子 | 58 | `factors_liunian.csv` |
| 流年复合因子 | 39 | `factors_liunian.csv` |

口径说明：直通因子仅含 direct 表达式；提取因子是单条件组且不引用其他因子；复合因子含多条件组或 factor_ref。

统计由 `tests/test_bazi_model.py` 与表数据同步校验，不需要手工维护第二份清单。

## 5. 查询契约

`pan_digest` 是 canonical SHA-256 完整性摘要，用于发现误改或手工拼装；当前不承担服务端防伪造签名职责。
`pan_schema` 要求 `ziwei.gong_wei` 按 engine 12 宫闭集完整输出，`palace_facts` 非空且 palace 名必须落在同一闭集内。
`bazi.fullchart` 只接受 engine `bazi.chart` 产生的 canonical lean chart：四柱必须构成合法六十甲子，且每柱 `na_yin` 存在并与干支一致；缺失或错配直接报错，不把裁剪盘扩展成空纳音。

`calibrate.py` 是独立考时工具，编排 `paipan → factors → duanyu`。候选 `correct=true` 必须提供 longitude；`correct=false` 表示已明确时辰，longitude 可省略。`detail=true` 输出机械 evidence；`detail=false` 只保留断语。场景领域过滤只作用于断语，不删除机械 evidence。

考时事件的 `rule` 若是场景别名，同样应用 `场景领域过滤`；例如 `yearly_study` 只保留学业断语，避免用婚姻或财运信号校时。

- `duanyu.query` 只接受本命域；`query_yearly` / `yearly_range` 只接受流年域，`yingqi` 必须通过流年查询。
- `duanyu.query(rule=用神)` 除断语外返回 `yong_shen_context`，直接投影 engine 的 `yong_shen / element_states / ten_god_states`；Python 不重算三派、不推导最终喜忌。
- `query(year=...)` 只允许 `大运 / 大限` 限运域；省略 year 时由服务端当前时间推导。
- 限运域结果附带 `current_year / current_year_source`；显式传 year 时 source 为 `specified`。
- `query` / `yearly_range` 支持可选 `domains` 过滤器；有效领域来自所选 rule 展开后的断语表。未知领域 fail closed，过滤结果不携带 snapshot evidence。
- `八字专属域 / 紫微专属域` 只限制对应 bazi / ziwei 断言表；若该域还有 common 断言，`query` 必须同时生成双盘快照，否则跨术数条件会变成死规则。
- 场景别名可在 `constants.json` 的 `场景领域过滤` 中声明主领域。未显式传 `domains` 时，纯场景查询应用默认领域过滤。
- `query` / `yearly_range` 只接受 `full_paipan` 完整返回的 pan，拒绝快照、裁剪盘和手工半截盘。
- `yearly_range` 单次起止年含端点跨度最多 120 年。

`full_paipan` 在真太阳时或既定时辰距时辰交界 ≤30 分钟时，返回可选 `calibration_hint`。该提示只表达“接近交界，建议用人生大事校准”，不修改四柱，也不构成吉凶结论。

## 6. 因子长表

因子清单的唯一事实源是长表：

- `skills/liki-bazi/tools/factors/factors.csv`
- `skills/liki-bazi/tools/factors/factors_liunian.csv`

本文只记录分层、统计与不可变契约，不复制因子行。CSV 是因子清单唯一事实源，文档、测试或快照不得再维护第二份逐因子清单。

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

## 7. 常量与闭集

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

## 8. 断言长表

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
- 命中结果附带 `trace`：命中的 condition group、每个因子 expected / actual。`brief` 不输出 trace；`detail=true` 或直接 `query` 保留完整解释链。
- common 断语只表达八字与紫微同向证据；无 common 命中时，LLM 只能并列解释两侧结果，不得伪装成已合参。

## 9. 稳定领域事实

`natal_projection.py` 从完整 `pan` 只读投影稳定命理事实，入口和目标字段由 `natal_projection_contract.json` 锁定：

- 八字：纳音、藏干、旬空、自合、魁罡、三元、三奇、拱夹、大运等。
- 紫微：宫位与星曜、局数、命主、身主、命身宫、空宫、大限等。
- 上下文：性别、公历出生、农历出生、当前年份。

这些领域事实与因子表同属稳定领域模型；当前断语是否消费不作为删除依据。

## 10. 排盘上下文

| 上下文 | 值域 | 消费方式 |
|---|---|---|
| 性别 | `male` / `female` | 排盘上下文，不算因子；断言匹配时并入对应视图 |
| 当前年份 | 公历年 | 仅本命「大运」域查询当前大运时使用 |
| 公历出生 | 字符串 | `full_paipan` 出生事实透传 |
| 农历出生 | 字符串 | `full_paipan` 出生事实透传 |

## 11. 缓存与复用

- `FactorContext` 只缓存一次求值内的基础聚合，不写回公共 `pan`。
- `NatalContext` 只复用本命基础聚合和本命快照；流年盘、流年快照与公共 `pan` 保持只读。
- 本命快照按调用生命周期生成，不做全局 pan 内容缓存。

## 12. 硬约束

1. 命理规则只由 `constants.json`、因子长表与断言长表定义；代码只做机械解析、查表与求值。
2. 因子行必须有直通表达式或条件；断言行必须有条件组。
3. 标量因子的断语约束值必须来自对应常量闭集。
4. 因子求值、断言表读取和 `time.now` 失败不得降级为 0、空表或本地时间。
5. 生产表不得包含评测 case、迭代阶段或旧内部路径等过程残留。

`python3 tests/check_schema.py` 校验约束键、流年可达性、单侧表边界、条件组完整性、标量闭集与生产纯度。

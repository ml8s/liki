# Natal 领域模型（八紫双盘）

`skills/liki/natal` 是八紫双盘同参的应用领域包；`natal` 表示本命分析场景，覆盖八字四柱、紫微斗数，以及两条体系之间的 `common` 合参边界。狭义“八字”只在 `shushi=bazi`、RPC 名称或具体柱盘上下文中表示四柱子系统。

本文记录该领域的稳定对象、分层边界和求值契约。因子与断言是本命断语的条件机制，不是领域模型的顶层名字。

## 1. 领域对象

| 对象 | 含义 | 边界 |
|---|---|---|
| BirthChart | 一次排盘得到的不可变资源，包含八字、紫微和双盘事实 | 公共层只暴露 `chart_ref`；token 解码后校验 digest，拒绝裁剪盘、手工半截盘和旧快照 |
| Pan | BirthChart 解码后的只读领域上下文 | 仅存在于工具层内部，不再作为 LLM 参数 |
| Topic | 用户人生问题闭集 | 由 `topic_routes.json` 声明；路由到命理规则和断语领域 |
| Side | 命理侧别：`bazi` / `ziwei` / `common` | 八字与紫微各自求值；跨体系断言只能在双盘合并上下文命中 |
| AtomicFact | engine 输出的确定性命理事实 | Python 只读取，不复算十神、五行、关系、宫位和亮度规则 |
| Factor | 可被断言引用的稳定条件因子 | 因子清单由长表唯一定义，OR / AND 分组与引用关系可校验 |
| Assertion | 一条带条件组和经典依据的断语 | 不直接内嵌命理推导；只能引用可达因子 |
| TimeLayer | 本命、大运 / 大限、流年等时间层 | 跨层引用必须显式；限运域必须携带目标年份上下文 |

## 2. 模块分层

```text
paipan.py           排盘 RPC 适配与城市解析
chart_token.py      不可变 BirthChart 资源编解码
analytics.py        公共分析编排与 TopicRouter
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
BirthInput → BirthChart → chart_ref
chart_ref → Pan → factors → snap → assertions
Topic + TimeScope → TopicRouter → assertion domain filter
```

## 3. Engine 原子事实

命理结论先由 engine 计算为原子事实，Python 算子只读取和组合，不复算：

| 字段 | 含义 |
|---|---|
| `full.nian/yue/ri/shi.shen_sha` | engine 神煞事实；勾绞、元辰按年干阴阳 × 性别定向，学堂 / 词馆按年命纳音同气正位 |
| `full.nian/yue/ri/shi.day_master_trend` | 日主在该柱地支上的十二长生；`self_sitting` 是该柱天干坐支，两者不同 |
| `full.nian/yue/ri/shi.xun / xun_kong` | 该柱自身所在旬与旬空二支；`is_void` 仍固定表示该柱地支值日柱旬空 |
| `full.shen_sha_school` | 天乙、桃花、驿马、华盖、将星、劫煞、灾煞采用年参照与日参照并集；不隐含单一流派 |
| `full.tiao_hou` | 《穷通宝鉴》主用 / 辅用表候选；`primary_wuxing / secondary_wuxing` 表达表内主辅五行，`primary / secondary` 记录对应天干在四柱的透干、藏支与所在柱，并投影其宫支参与的合会冲刑；`secondary_condition` 原样保留“水多用戊”等条件，不得当无条件喜神；不机械反推忌神，扶抑 / 格局上下文另核 |
| `full.fu_yi.basis` | 扶抑强弱表输入：日主根型、月令旺相、印比数量、日主藏干根，以及日主 / 印比 / 日支 / 日主根相关的关系投影；动态再权衡仍须整体合参 |
| `full.ge_ju.structure` | 月令格神、克格神与生格神三类原子状态：透干柱、藏干根、季节旺相与组合强弱；同时投影与三类五行相关的天干五合和地支合会冲刑 |
| `full.ge_ju` | 月令藏干透干格局候选、格神、十神、顺逆用与来源；不输出 `yong / xi / ji`，不把格神冒充最终用神；`structure` 是候选证据，不输出成格 / 败格 / 救应终判 |
| `full.fu_yi.basis.relation_facts[].targets` / `full.ge_ju.structure.relation_facts[].targets` / `full.tiao_hou.primary.relation_facts[].targets` | 关系投影的角色闭集；同一地支多藏干可命中多个角色，必须全部保留，不得按第一个藏干截断 |
| `full.lu_roots` | 透干十神得十干禄 |
| `full.gan_he` | 全部两两天干五合事实，含 `adjacent / separated / remote` 柱距与 `contested` 争合候选标记 |
| `full.gan_chong` | 甲庚、乙辛、丙壬、丁癸四组天干相冲；柱距只作强度证据，紧邻进 `gan_chong`，隔位 / 远隔进 `gan_chong_candidate` |
| `full.san_he / full.san_hui` | 完整三合局与三会方，含成员支与所在柱；结构投影只消费这些 engine 定位事实 |
| `full.san_he_partial` | 半合证据；三合缺一支时必须包含旺支，缺旺支的生墓二支是拱合候选，不标半合 |
| `pair relation facts` | 两支交叉（流年对原局、合盘对柱）只允许半合 / 拱候选；完整三合 / 三会由全盘 `ComputeHeHui` 判定 |
| `full.tai_xi` | 日柱胎息：日干五合配、日支六合配 |
| `full.gong_jia` | 只输出经典八拱：三合首墓二支拱旺支、三会首尾二支拱中支；任意隔一支不冒充拱局 |
| `full.day_xun / full.day_xun_kong` | 日柱所在旬与旬空；这是全盘 `is_void` 的判定基准 |
| `full.relation_groups` | 去重后的稳定关系组：紧邻天干五合 / 相冲进 `gan_he` / `gan_chong`，隔位 / 远隔进对应候选组；地支六合 / 三合 / 半合 / 三会 / 六冲 / 六害 / 三刑 / 六破 / 暗合按全组或成对规则 |
| `bond` | 双盘原始事实层：八字侧输出日主、夫妻宫、四柱交叉、十神视角、实盘五行 / 扶抑互见、纳音、性别配偶星与神煞出现事实；紫微侧只并排命宫 / 夫妻宫 / 子女宫主星。字段用 `a / b` 或 `a_to_b / b_to_a` 明示方向；不输出合婚评级 |
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

这些字段的命理口径由 engine 单测、`tests/fixtures/domain_oracle/` 与 `tests/golden/bazi/` 的三源共识 fixture 锁定。Python 新增因子时不得重新实现上述推导。

紫微 `宫含` 算子只对 `ziwei.palace_facts` 做 palace / kind / target / star exact match；Python 不再遍历宫位、推导四化落宫、解释亮度分组或计算主星数量。
紫微盘显式输出 `school`：当前闰月口径为 iztro v2.6.1 兼容的“前十五日本月、后十五日次月”。这不是《紫微斗数全书》闰月按下月口径；两派不得混写。主星亮度表同步 iztro v2.6.1 对太阳、太阴、七杀居酉的修正。
`ziwei.liuyue` 输入必须已经是完整农历日期（`year/month/day/leap`）；engine 只消费该领域输入。真实闰月由 `resolved_period.calendar_period` 表达，实际流月由 `resolved_period.flow_month` 表达：闰月 1–15 日取本闰月，16 日至月末取下一农历月。Python 不得判断 `day > 15`，也不得把闰月提前改写成普通月。
`ziwei.liuyue` 不接受公历、裸农历月份或 `half / fixLeap` 修正参数；同一农历年的六月与闰六月是两个真实月份。农历月不存在、日期超过真实月末或缺少 `day / leap` 时必须 fail closed。
`ziwei.liuri` / `ziwei.liushi` 同样必须提供完整农历日期（`year/month/day/leap`），用于定位真实干支纪日；不得用裸农历序数表达闰月。流盘 `xing_yao` 数组只保证确定性展示顺序：禄、羊、陀、魁、钺、马、鸾、喜、昌、曲；该顺序不是吉凶排序，也不改变星曜落宫与作用。
本命宫名使用 engine 闭集：`命宫、兄弟、夫妻、子女、财帛、疾厄、迁移、仆役、官禄、田宅、福德、父母`；除命宫外不追加“宫”字。因子 DSL 通配符是代码级语法 token `FACTOR_WILDCARD`，当前值为 `任意`；它只表示跨闭集匹配，不是宫名、星曜名或刑组名。
流年紫微同样消费 engine 宫名闭集；`流曜入宫` 与 `流年宫化` 不做带“宫”字后的显示别名适配。`流年宫化` 的三维 key 为 `[宫位,星曜,四化]`，与 `宫含` 同构；`宫位`/`星曜` 取通配符表示该维通配，星曜维度必须显式声明、不得省略。
`三刑` 的二维 key 为 `[来源,刑组]`：`来源` 决定参与判定的地支，`刑组` 决定是哪一组刑，两者都必须显式声明；来源不得使用通配符。刑组闭集取自 `constants.json` 的「三刑」（地支 → 同组其余地支），成员按集合比较，因此与本命侧 `关系[liu_xing,组名]` 的组名书写次序（`丑戌未` / `丑未戌`）无关。刑组字符串必须恰好枚举全部成员，不得重复、缺字或使用通配符；不自洽时 fail closed。`三刑` 因子与断语条件逐组对齐。

跨层因子值契约为 `FactorValue`：只接受 `0 / 1` 或领域字符串；空字符串表示字符串型因子当前不可用。不接受 `null`、boolean、数组或对象。断语 `trace.factors` 中的 `expected / actual` 必须使用同一 `FactorValue` 契约。
`analytics.analyze_natal` 与 `analytics.analyze_periods` 是公共边界；两者最终调用内部断语门面。公共断语边界必须校验参与断语的因子都在 snapshot 中；缺失因子 fail closed，不得静默按 `0` 参与匹配。

`ten_god_states.transparent / hidden` 只描述该具体十神；`rooted / timely` 按其五行判定。`ten_god_states.strength` 的口径是：得令，或该具体十神透干且五行通根为 `strong`；失令且该十神不透、五行无根为 `weak`；其余为 `neutral`。同五行的另一十神透干，不会把本十神错误升级为透干有根。`element_states.season_strength` 只表达月令旺相休囚死；`element_states.strength` 是五行聚合态；`controller_strength` 表达受克目标的克者是否旺相。Python 只读取这些结论，不再维护五行生克、天干五行、得令状态或十神旺弱规则表。

## 4. 统计口径

| 口径 | 数量 | 事实源 |
|---|---:|---|
| 本命因子 | 510 | `factors.csv` |
| 本命八字因子 | 219 | `factors.csv` |
| 本命紫微因子 | 291 | `factors.csv` |
| 本命定义组 | 565 | `factors.csv` |
| 本命数据行 | 678 | `factors.csv` |
| 本命直通原子 | 50 | `factors.csv` |
| 本命提取原子 | 322 | `factors.csv` |
| 本命复合因子 | 138 | `factors.csv` |
| 流年因子 | 111 | `factors_liunian.csv` |
| 流年八字因子 | 79 | `factors_liunian.csv` |
| 流年紫微因子 | 32 | `factors_liunian.csv` |
| 流年定义组 | 114 | `factors_liunian.csv` |
| 流年数据行 | 121 | `factors_liunian.csv` |
| 流年直通原子 | 4 | `factors_liunian.csv` |
| 流年提取原子 | 64 | `factors_liunian.csv` |
| 流年复合因子 | 43 | `factors_liunian.csv` |

口径说明：直通因子仅含 direct 表达式；提取因子是单条件组且不引用其他因子；复合因子含多条件组或 factor_ref。

统计由 `tests/test_bazi_model.py` 与表数据同步校验，不需要手工维护第二份清单。

## 5. 查询契约

`pan_digest` 是 canonical SHA-256 完整性摘要，用于发现误改或手工拼装；当前不承担服务端防伪造签名职责。
`pan_schema` 要求 `ziwei.gong_wei` 按 engine 12 宫闭集完整输出，`palace_facts` 非空且 palace 名必须落在同一闭集内。
`bazi.fullchart` 只接受 engine `bazi.chart` 产生的 canonical lean chart：四柱必须构成合法六十甲子，且每柱 `na_yin` 存在并与干支一致；缺失或错配直接报错，不把裁剪盘扩展成空纳音。

`calibrate.py` 是独立考时工具，编排 `paipan → factors → duanyu`。候选 `correct=true` 必须提供 longitude；`correct=false` 表示已明确时辰，longitude 可省略。`detail=true` 输出全量断语（结论 + 依据 + 经典依据 + trace）；`detail=false` 只保留精简字段。场景领域过滤只作用于断语。

考时事件的 `rule` 若是场景别名，同样应用 `场景领域过滤`；例如 `yearly_study` 只保留学业断语，避免用婚姻或财运信号校时。

- 公共工具闭集是 `create_birth_chart`、`analyze_natal`、`analyze_periods`、`compare_birth_charts`、`calibrate_birth_time`；旧 `full_paipan / query / yearly_range / bond / calibrate / city_coords` 不再是 LLM 公共面。
- `duanyu.query` 降级为内部本命断语门面；`query_yearly` / `yearly_range` 降级为内部流年求值门面，`yingqi` 仍只能走流年求值。
- `duanyu.query(rule=用神)` 除断语外返回 `fu_yi / tiao_hou / ge_ju / element_states / ten_god_states`；Python 不重算三源、不推导最终喜忌。
- `TopicRouter` 从 `topic_routes.json` 读取 `topic → natal_rules / annual_rules / decade_rules / domains`；LLM 只传受控 topic，不选择断语表。
- `analyze_natal` 输出 `assertion_id / side / method / topic / time_scope / event / conclusion / evidence`；内部中文标签只保留在 `source` 和展示内容中。
- `analyze_periods` 把流年、流年区间和大运/大限统一为 `time_scope`；年份区间含端点，最多 120 年。
- 所有分析只接受 `create_birth_chart` 返回的 `chart_ref`；token 解码后还原完整 pan 并校验摘要，拒绝快照、裁剪盘和手工半截盘。
- 限运结果附带 `current_year / current_year_source`；显式锚点年 source 为 `specified`，当前限运 source 来自 engine。

`create_birth_chart` 在真太阳时或既定时辰距时辰交界 ≤30 分钟时，返回可选 `calibration_hint`。该提示只表达“接近交界，建议用人生大事校准”，不修改四柱，也不构成吉凶结论。

除无符号的 `minutes_to_boundary` 外，提示同时给出可核验的机械字段：`boundary_offset_minutes`（有符号，负数=早于交界）、`current_shichen` / `alternate_shichen`（当前时辰与跨过最近交界后的时辰，含 `name` / `branch` / `span`）与 `direction`（`later` / `earlier`）。这些字段直接回答“往哪边偏会翻”，调用方不再自行二次推断。时辰名由 `constants.json` 的「地支」与既有交界表推导，不写死命理成员。**换日口径不在本层表达**——提示只标注两小时窗口，`23:00` 前后同属一个窗口，不区分早晚子时。

## 6. 因子长表

因子清单的唯一事实源是长表：

- `skills/liki/natal/tools/factors/factors.csv`
- `skills/liki/natal/tools/factors/factors_liunian.csv`

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
- `原语直通[...,<FACTOR_WILDCARD>]` 可返回字符串标量。

## 7. 常量与闭集

十神、五行、干支、十二长生、紫微星曜、宫位、神煞、关系表与命理侧闭集均以 `skills/liki/natal/tools/constants.json` 为唯一事实源。代码只做机械查表、解析与求值，不内置命理结论。

| 层 | 内容 | 说明 |
|---|---|---|
| 基础闭集 | 五行、天干、地支、十神、十二长生、旺衰状态、紫微主星、煞星、六吉星、文星 | 原子词汇 |
| 十神大类 | 官杀、印星、财星、食伤、比劫 | 十个原子十神的完整不交叉 partition |
| 六亲角色 | 配偶星、子女星、父星、母星、日主 | 引用十神大类或原子十神，按性别解析 |
| 事件宫位 | 配偶星、父星、母星、子女星、官杀、财星、日主 | 对应日支、年支或时支 |
| 关系表 | 天干五合、地支六合、三合、三会、六冲、六害、六破、暗合、三刑、旬空 | 稳定关系闭集 |
| 算子语义 | 旺弱规则、宫位关系、用忌映射、格局十神、紫微四化与亮度分组等 | operator 只做机械查表；DSL token 在 `factor_tokens.py` |
| 流年机械 | 事件宫位、干支来源、关系类型、三合半合、旬空起点、流年宫名 | 流年 target 与求值由表驱动 |
| 流年年界 | 八字干支年、紫微农历年 | `analyze_periods.year_basis` 领域语义 |
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
| 公历出生 | 字符串 | `create_birth_chart` 出生事实透传 |
| 农历出生 | 字符串 | `create_birth_chart` 出生事实透传 |

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
6. 断言 DNF 的每个条件组必须通过静态可达性审计；性别等封闭上下文和直通标量的不同取值互斥，因子引用先展开成原子条件再检查冲突。

`python3 tests/check_schema.py` 校验约束键、流年可达性、单侧表边界、条件组完整性、标量闭集与生产纯度；`tests/test_mingli_assertion_reachability.py` 展开因子 DNF 并拒绝永久不可达条件组。

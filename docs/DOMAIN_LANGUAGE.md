# Liki canonical domain language

This document is generated from [`tools/domain-language.json`](../tools/domain-language.json).
Do not edit the tables by hand. One domain concept has one ID and one canonical
projection per surface. A legacy name is accepted only while it is declared in
that concept's compatibility list.

## Concepts

| Concept ID | Status | Definition | Go | Python | JSON | 中文 | Aliases |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `instant` | canonical | 内部天文学使用的绝对时间点；不是公开 MCP wire 字段。 | `time.Time` | `datetime.datetime` | `(internal)` | 内部精确时刻 | 禁止：`timestamp`, `time_point`, `utc_time` |
| `standard_time` | canonical | 当地时间标准时：精确时刻 + 标准偏移，不含 DST；排盘/流年工具的公开时间输入。 | `StandardTime` | `str` | `standard_time` | 标准时间 | 禁止：`legal_time`, `standard_local_time` |
| `zone` | canonical | IANA 时区规则，用于把精确时刻映射到当地壁钟时间；偏移量是派生结果，不是时区身份。 | `*time.Location` | `zoneinfo.ZoneInfo` | `timezone` | IANA 时区 | 禁止：`time_zone`, `iana_timezone` |
| `apparent_solar_time` | canonical | 由标准时间、经度修正和均时差推导出的真太阳时；用于日界和时辰等真太阳尺度规则。 | `tianwen.ApparentSolarFromStandard` | `projection function` | `apparent_solar_projection` | 真太阳时 | 禁止：`SolarTime`, `solar_time`<br>兼容：`solar_time`（v2026.12.01.0）, `true_solar_time`（v2026.12.01.0） |
| `gregorian_date` | planned | 公历日历中的日期，只表达日历日期，不表达时间尺度或时区。 | `CivilDate` | `CivilDate` | `date` | 公历日期 | — |
| `lunar_date` | canonical | 农历合历日期，可包含闰月标记，但不包含时辰。 | `LunarDate` | `LunarDate` | `lunar` | 农历日期 | — |
| `shichen` | canonical | 基于真太阳时尺度的十二时辰。 | `ganzhi.Zhi` | `ganzhi.Zhi` | `shichen` | 时辰 | — |
| `solar_term` | canonical | 太阳视黄经确定的二十四节气边界。 | `SolarTerm` | `SolarTerm` | `solar_term` | 节气 | — |
| `solar_term_state` | canonical | 某一时刻所在节气窗口的状态，包含上一个、当前和下一个节气边界。 | `tianwen.SolarTermState` | `SolarTermState` | `solar_term_state` | 节气状态 | 禁止：`solar_term_context` |
| `chart` | canonical | 确定性排盘计算结果；输入是排盘时间上下文与领域参数，输出可被判断层消费。 | `Chart` | `Chart` | `*_chart` | 排盘 | 禁止：`paipan_result`, `pan_result` |
| `chart_input` | planned | 一次排盘所需的显式输入，不包含隐式当前时间或地点默认值。 | `ChartInput` | `ChartInput` | `chart_input` | 排盘输入 | — |
| `factor` | canonical | 由排盘事实按规则表推导的判断因子。 | `Factor` | `Factor` | `factor_snapshot` | 因子 | — |
| `factor_snapshot` | planned | 一次判断所需的因子集合、来源和完整性元数据的不可变快照。 | `FactorSnapshot` | `FactorSnapshot` | `factors` | 因子快照 | — |
| `assertion` | canonical | 规则表在因子快照上命中的结构化断语，保留规则、证据和出处。 | `Assertion` | `Assertion` | `assertion` | 断语 | — |
| `rule` | canonical | 表驱动的条件与结论单元，拥有稳定规则 ID 和版本。 | `Rule` | `Rule` | `rule` | 规则 | — |
| `evidence` | canonical | 支撑结论的命中的结构化事实和规则片段。 | `Evidence` | `Evidence` | `evidence` | 依据链 | — |
| `provenance` | canonical | 事实、规则和经典出处的可追溯来源。 | `Provenance` | `Provenance` | `provenance` | 来源追溯 | — |
| `digest` | canonical | 对规范化领域对象计算的抗碰撞摘要，用于追问与重放校验。 | `Digest` | `digest` | `*_digest` | 摘要 | — |
| `snapshot` | canonical | 可追问的完整领域输入快照，绑定内容摘要，不因后续对话静默变化。 | `Snapshot` | `Snapshot` | `*_snapshot` | 快照 | — |
| `answer` | canonical | 基于既有快照回答追问得到的结构化结果。 | `Answer` | `Answer` | `answer` | 追问回答 | — |
| `yongshen` | canonical | 命局裁决得到的用神结论及其来源、证据和置信度。 | `YongShen` | `YongShen` | `selected_yongshen` | 用神 | 禁止：`use_god`<br>兼容：`explicit_yong_shen`（v2026.12.01.0）, `yong_shen_candidates`（v2026.12.01.0） |
| `geju` | canonical | 以月令格神为核心的子平格局候选与结构。 | `GeJu` | `GeJu` | `ge_ju` | 格局 | — |
| `tiaohou` | canonical | 调候用神候选及其月令条件。 | `TiaoHou` | `TiaoHou` | `tiao_hou` | 调候 | — |
| `fuyi` | canonical | 扶抑用神候选及其强弱事实。 | `FuYi` | `FuYi` | `fu_yi` | 扶抑 | — |
| `dayun` | canonical | 八字大运区间及其起运、干支和顺逆信息。 | `DaYun` | `DaYun` | `dayun` | 大运 | — |
| `liunian` | canonical | 流年干支及其与本命的组合事实。 | `LiuNian` | `LiuNian` | `liunian` | 流年 | — |
| `yingqi` | canonical | 事件可能兑现的时间窗口或时间点。 | `YingQi` | `YingQi` | `ying_qi` | 应期 | — |

## Rules

1. The `id` is immutable and is never reused for a different concept.
2. Code, contracts, comments, tests, and documentation must use the canonical projection for their surface.
3. A compatibility name must be scoped, must state its reason, and must declare a removal version.
4. New code may not introduce a forbidden alias or an unregistered synonym.
5. Runtime and deployment names remain governed by [DOMAIN_NAMING.md](./DOMAIN_NAMING.md).

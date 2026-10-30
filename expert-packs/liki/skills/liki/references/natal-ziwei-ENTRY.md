# 紫微层入口 — 任务路由与执行流

紫微侧自举入口：任务路由、执行流与断语契约。能力、参数与结果以运行时自省描述和契约为准。

## 任务路由

| 信息需求（应用词） | 领域卡 | 侧查询 |
| --- | --- | --- |
| 事业本命 | `references/natal-ziwei-gong12.md`、`references/natal-ziwei-zhuxing.md` | 本命判断（事业） |
| 事业应期 | `references/natal-ziwei-liunian.md`、`references/natal-ziwei-yingqi.md` | 应期判断 |
| 婚姻本命 | `references/natal-ziwei-gong12.md`、`references/natal-ziwei-zhuxing.md` | 本命判断（婚姻） |
| 婚姻应期 | `references/natal-ziwei-yingqi.md` | 应期判断 |
| 财运本命 | `references/natal-ziwei-gong12.md` | 本命判断（财运） |
| 房产 | `references/natal-ziwei-gong12.md` | 本命判断（房产） |
| 健康 / 精神 | `references/natal-ziwei-gong12.md`、`references/natal-ziwei-yingqi.md` | 本命判断（健康 / 精神） |
| 学业 | `references/natal-ziwei-yingqi.md`、`references/natal-ziwei-gong12.md` | 本命判断（学业） |
| 性格 | `references/natal-ziwei-gexing.md`、`references/natal-ziwei-zhuxing.md` | 本命判断（性格） |
| 六亲 | `references/natal-ziwei-gong12.md` | 本命判断（家庭） |
| 相貌 | `references/natal-ziwei-xiangmao.md` | 本命判断（外貌） |
| 合盘 | `references/natal-ziwei-gong12.md` | 合盘 |
| 考时校准 | `references/natal-ziwei-calibration.md` | 校准交叉验证 |
| 基础（排盘 / 格局 / 流年） | `references/natal-ziwei-calibration.md`、`references/natal-ziwei-geju.md`、`references/natal-ziwei-liunian.md` | 排盘 / 格局 / 流年 |

## 输入约定

- 执行本层前须具备：原始出生信息（公历日期、Gregorian time或时段、出生地、性别；时辰存疑时附候补时辰）、所需信息（主题 × 维度）、约束（用户点名体系等）。
- 跨轮：每次执行都从原始出生信息与已确认事实（如已定时辰、校时结论）出发，不依赖先前轮次的隐含上下文。

## 执行流

1. 历法与排盘：出生信息 → 取时间事实（运行时自举时间候选归一能力）；历史时制提示或多候选须先向用户确认，不静默取一 → 农历 / 时辰 → 十二宫与大限。
2. 因子：因子快照（本侧不含用神裁决，星曜与宫位提供独立证据）。
3. 本命判断：因子快照 + 问题域 → 本侧断语。
4. 应期判断：因子快照 + 时间层 + 问题域 → 大限 / 流年 / 流月 / 流日断语。
5. 完成时以 `finish_task` 提交：盘面要点、断语、依据摘要、不确定项。

## 断语契约

- 断语公共字段（`side` / `topic` / `time_scope` 等）以 counsel 输出 schema 为准（自省读取）；`side` 是 `bazi` / `ziwei`。
- topic 标签（counsel topic_routes 单一源）：`marriage`＝婚姻、`family`＝家庭、`children`＝子女、`career`＝事业、`wealth`＝财运、`health`＝健康、`mental`＝精神、`study`＝学业、`personality`＝性格、`appearance`＝外貌、`origin`＝出身、`relocation`＝迁移、`social`＝人际、`property`＝房产、`adversity`＝风险、`chart_structure`＝盘面结构。
- 时间层 `time_scope`：`current_year` / `year` / `year_range` / `current_decade` / `decade`；年份为年号（紫微农历春节界），不做日期换算；区间含端点。
- 直接采纳判断输出为最终判断；判断输出与规则卡冲突 → 以判断输出为准并说明冲突。
- 排盘失败 → 提示重试，不降级查表。
- 重结论（无子、重大灾变、残疾等）须宫、星、限运、流年独立闭环；现实已发生事实可作验证，不作命理反推。
- 问凶时凶信号定调、吉信号不否决；问吉时凶象不排除该年；吉凶并存按域列证。
- 语义直指优先；另解同样有断语支撑时并列说明证据；主断语优先于辅象，同源重复信号不累加。
- 禁止用其他问题的结论 + 命理修辞互斥推算。
- 本侧断语只用本侧因子；跨体系合参由场景卡编排，不在本侧代裁。

## 硬边界

- 只做紫微侧；八字属八字专家，六爻、奇门、黄历择日、起名、风水属其他专家。
- 星曜与宫位提供独立证据，不替代八字用神裁决。

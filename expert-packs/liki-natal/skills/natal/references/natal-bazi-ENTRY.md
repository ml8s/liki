# 八字层入口 — 任务路由与执行流

八字侧自举入口：任务路由、执行流与断语契约。能力、参数与结果以运行时自省描述和契约为准。

## 任务路由

| 信息需求（应用词） | 领域卡 | 侧查询 |
| --- | --- | --- |
| 事业本命 | `references/natal-bazi-career.md`、`references/natal-bazi-wangshuai.md`、`references/natal-bazi-shishen.md` | 本命判断（事业） |
| 事业应期 | `references/natal-bazi-dayun.md` | 应期判断 |
| 婚姻本命 | `references/natal-bazi-shishen.md`、`references/natal-bazi-gongwei.md` | 本命判断（婚姻） |
| 婚姻应期 | `references/natal-bazi-dayun.md` | 应期判断 |
| 财运本命 | `references/natal-bazi-wealth.md`、`references/natal-bazi-shishen.md` | 本命判断（财运） |
| 财运应期 | `references/natal-bazi-dayun.md` | 应期判断 |
| 健康 / 精神 | `references/natal-bazi-wuxing-health.md`、`references/natal-bazi-hehui.md`、`references/natal-bazi-dayun.md` | 本命判断（健康 / 精神） |
| 学业 | `references/natal-bazi-study.md`、`references/natal-bazi-dayun.md` | 本命判断（学业） |
| 性格 | `references/natal-bazi-shishen.md`、`references/natal-bazi-wangshuai.md` | 本命判断（性格） |
| 六亲 | `references/natal-bazi-family.md`、`references/natal-bazi-gongwei.md`、`references/natal-bazi-hehui.md` | 本命判断（家庭） |
| 合盘 | `references/natal-bazi-hepan.md` | 合盘 |
| 考时校准 | `references/natal-bazi-calibration.md` | 校准编排 |
| 基础（排盘 / 用神 / 结构） | `references/natal-bazi-yongshen.md`、`references/natal-bazi-geju.md`、`references/natal-bazi-tiaohou.md` | 排盘 / 因子 / 用神裁决 |

## 输入约定

- 执行本层前须具备：原始出生信息（公历日期、Gregorian time或时段、出生地、性别；时辰存疑时附候补时辰）、所需信息（主题 × 维度）、约束（用户点名体系等）。
- 跨轮：每次执行都从原始出生信息与已确认事实（如已定时辰、校时结论）出发，不依赖先前轮次的隐含上下文。

## 执行流

1. 校准与排盘：出生信息 → 取时间事实；返回历史时制提示（如夏令时）或时间落在边界产生的多候选时，候选即本流程候选项，须先向用户确认，不静默取一 → 排四柱。
2. 因子与用神：因子快照内含唯一用神裁决，直接消费同一份，不自行裁决、不改口径。
3. 本命判断：因子快照 + 问题域 → 本侧断语。
4. 应期判断：因子快照 + 时间层 + 问题域 → 大运 / 流年 / 流月断语。
5. 完成时以 `finish_task` 提交：盘面要点、裁决与断语、依据摘要、不确定项。

## 断语契约

- 断语公共字段（`side` / `topic` / `time_scope` 等）以 counsel 输出 schema 为准（自省读取）；`side` 是 `bazi` / `ziwei`。
- topic 标签（counsel topic_routes 单一源）：`marriage`＝婚姻、`family`＝家庭、`children`＝子女、`career`＝事业、`wealth`＝财运、`health`＝健康、`mental`＝精神、`study`＝学业、`personality`＝性格、`appearance`＝外貌、`origin`＝出身、`relocation`＝迁移、`social`＝人际、`property`＝房产、`adversity`＝风险、`chart_structure`＝盘面结构。
- 时间层 `time_scope`：`current_year` / `year` / `year_range` / `current_decade` / `decade`；年份为年号（八字公历立春界），不做日期换算；区间含端点。
- 直接采纳判断输出为最终判断；判断输出与规则卡冲突 → 以判断输出为准并说明冲突。
- 排盘失败 → 提示重试，不降级查表。
- 重结论（无子、重大灾变、残疾等）须宫、星、限运、流年独立闭环；现实已发生事实可作验证，不作命理反推。
- 问凶时凶信号定调、吉信号不否决；问吉时凶象不排除该年；吉凶并存按域列证。
- 语义直指优先；另解同样有断语支撑时并列说明证据；主断语优先于辅象，同源重复信号不累加。
- 禁止用其他问题的结论 + 命理修辞互斥推算。
- 本侧断语只用本侧因子；跨体系合参由场景卡编排，不在本侧代裁。

## 硬边界

- 只做八字侧；紫微属紫微专家，六爻、奇门、黄历择日、起名、风水属其他专家。
- 单侧断语不得混用异侧因子。

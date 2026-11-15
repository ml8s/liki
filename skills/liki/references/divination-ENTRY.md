# Liki 六爻奇门占卜择日 — 六爻算卦 / 奇门遁甲 / 黄历择日

> 覆盖六爻结果与应期、奇门策略与方向、黄历择日。工具、参数与字段以 MCP 自省和 契约 为准。

## 路由

每个占问（六爻、奇门、黄历择日）先经 `references/divination-app-question.md` 锁定目标并 ⛔ 确认方法，确认后按下表进入方法卡；用户明确指定六爻、奇门或黄历时，通过安全校验后优先采用该指定。

| 用户目标 | 方法 | 入口卡 |
| --- | --- | --- |
| 事件结果 / 能不能成 / 何时有结果 | 六爻 | `references/divination-app-outcome.md` |
| 策略 / 进退 / 方向 / 行动时机 | 奇门 | `references/divination-app-decision.md` |
| 哪天适合做事 | 黄历 | `references/divination-app-date.md` |
| 长期命局 | — | 转本命域（natal），不临时起卦 |

## 流程产物与边界

- 六爻流水线固定为：六爻起卦方法 创建唯一 `event` → 六爻排盘方法 排盘 → 六爻解读方法 出断语。`question` 与 `origin` / `focus` 由调用方（LLM）持有；event 只携带 `casting` 与 `casting_context.yong_shen`，不含问题、起卦来源或焦点。
- 六爻排盘方法 与 六爻解读方法 都只消费同一 `event`，返回的 chart 与 reading 都不再包含 event 副本；同一事追问基于已返回的断语解释，不得重新起卦。只有新事件、新时间或新决策才是新问测。
- App 带入的 `consultation_context` 若含 `event`（或旧记录的 `casting` + `casting_context`），先归一为同一 event；不得重新起卦。
- 事项与显式用神二选一：普通问事按 `focus.kind=matter` 选 14 类事项；`relationship` 必须带 `perspective=male/female`；同性、非二元或用户明确要求时才传 `focus.kind=yong_shen`。用神表见 `references/divination-liuyao-yongshen.md`。
- 最终结论由持有 `question` 的 LLM 结合 六爻解读方法 的断语解读；不要把问题文本放进 event 或 ask 参数。
- 黄历择日返回连续日期的适配事实；事项给定时按建除事项规则输出适配推荐，只输出日期适配事实，不预测现实结果。问事必填，事项取值与参数以 契约 为准。
- 奇门专占（失物 / 捕盗 / 走失 / 捕亡）共用以下口径：排盘取默认时家转盘基础盘，不传事项或用神；断语读取 Reading 中的专占断言，只解释工具命中的表内候选；候选并列给出，不合并成单一必然结论；方向与内外由八宫方向表、阴阳遁表投影；中宫无方向 / 内外，用神落中宫时不给对应候选。

## 硬边界

- 高风险主题（医疗、自伤、怀孕、重大人身或财务风险）由共享安全规则识别为 `安全提示`；高风险时先列支持资源再解读，工具继续返回正常盘面或择日结果。

## 输出契约

- 先给一句话判断，再列用神、盘面、动爻或择日依据等关键因子。
- 应期只解释 Reading 引用且与所问对象相关的候选。
- 结论只基于工具返回的 Reading、evidence 和领域事实，不把条件性倾向写成确定结果。
- 竞技 / 赛事类只给态势倾向与条件候选，不预测具体比分或胜负数值。

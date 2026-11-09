---
name: liki
license: MIT
description: 懂命理，用 Liki。专业命理 Skill：八字排盘、紫微斗数、大运流年、六爻占卜、奇门遁甲、黄历择日、家居风水、起名；覆盖婚姻、事业、财运、健康、学业、六亲、合盘与择时。规则引擎判断，依据可回溯；传统文化视角，不构成专业建议。
metadata:
  slug: liki
  display_name: Liki 专业命理 Skill
  display_name_en: Liki Professional Chinese Metaphysics Skill
  description_zh: 专业命理 Skill：八字算命、紫微斗数、六爻占卜、奇门遁甲、黄历择日、风水布局、起名、取名与流年运势分析。
  description_en: Chinese metaphysics skill covering BaZi (Four Pillars), Zi Wei Dou Shu, I Ching divination, Qi Men Dun Jia, almanac and Feng Shui charts, plus naming — every chart is computed by deterministic engines and every reading comes from rule-based judgments with traceable evidence.
  version: 2026.1109.0
  author: Liki
  agent_created: 'true'
  keywords: 八字合婚,婚姻分析,感情走向,事业分析,职业方向,财运分析,投资时机,学业分析,考试运,健康分析,五行体质,怀孕生育时机,新生儿起名,成人改名,八字起名,结婚吉日,开业吉日,搬家吉日,八宅风水,玄空飞星
---

# Liki — 专业命理 Skill

> **懂命理，用 Liki。**

## MCP 前置

Liki 通过标准 MCP 提供能力，两个免费端点：

| MCP | 端点 |
| --- | --- |
| ENGINE | `https://liki.hk/engine/mcp` |
| COUNSEL | `https://liki.hk/counsel/mcp` |

> 启动确认连接、比较本地 / 远程版本文件；远程 10 秒不可达时标注“版本未校验”后继续，托管模式跳过。工具、版本、digest或契约校验失败读 `references/FAQ.md`。

## MCP 自省规则

1. 先通过 MCP 自省发现当前连接暴露的领域功能。
2. 只调用自省返回的工具；参数以工具 schema 为准，返回事实以实际结果为准。
3. 工具描述描述业务用途；schema 与文档冲突时，以 MCP 自省返回和实际工具结果为准。

## sync_ui 提案

`sync_ui` 是前端确认卡触发器，不直接写入档案。仅当全部满足时自动调用一次：

1. 档案 `is_calibrated=false`。
2. 当前会话暴露 `sync_ui` 工具。
3. 考时已有唯一 canonical `standard_time`；或用户明确确认“就按当前/某候选时间”。
4. 用户没有拒绝该结论，也没有要求先不改档案。

payload：

```json
{
  "kind": "birth_time_calibration",
  "standard_time": "<RFC3339 canonical standard_time>",
  "summary": "<一句话用户可读结论>",
  "evidence": ["<短证据>", "<短证据>"]
}
```

规则：

- `standard_time` 必须来自引擎候选、用户明确确认，或交叉校准后的唯一结论。
- `summary` 只写一句用户可读结论；`evidence` 最多 5 条，每条只写短证据。
- `summary` 和 `evidence` 仅用于前端展示，不得包含出生数据、内部状态或档案写入承诺。
- 证据不足、多候选未定、用户拒绝、工具未暴露时不调用。
- 同一结论不要重复触发；用户说“同步/确认这个时间”或提供了新证据时可再次调用。
- 调用成功只表示提案已送到前端；必须让用户点击前端确认，档案才可能更新。

## 领域路由

通过 MCP 工具完成命理任务，按用户意图选择领域入口：

| 用户意图 | 领域入口 |
| --- | --- |
| 看命、婚姻、事业、财运、健康、学业、性格、六亲、合盘、大运、流年（本命场景） | `references/natal-ENTRY.md` |
| 八字、四柱、用神、十神、大运（八字侧单侧） | `references/natal-bazi-ENTRY.md` |
| 紫微、斗数、十二宫、星曜、大限（紫微侧单侧） | `references/natal-ziwei-ENTRY.md` |
| 六爻、问卦、占卜、算卦、应期、奇门、奇门遁甲、策略、方位、择时、黄历、择日 | `references/divination-ENTRY.md` |
| 起名、改名、宝宝起名、名字评估 | `references/naming-ENTRY.md` |
| 风水、八宅、玄空、流年风水 | `references/fengshui-ENTRY.md` |

意图不清时，先用一个问题确认主目标。例如：「你想看的是八字命盘分析，还是给某个具体事情算卦？」

多域需求先处理主领域；其他领域仅作显式佐证，并标注证据来源。

## 执行循环

1. 按上表进入领域 ENTRY，缺参时按 ⛔ / 💬 规则确认。
2. 通过 MCP 获取排盘、判断或事实：成功读 结构化结果 并把命盘、快照和 digest 原样传给下游工具；失败读 工具错误 与错误文本；幂等工具修正参数后重试，起卦等事件创建工具不自动重试。
3. 读取对应 app / domains 方法论卡解释工具结果。
4. 按“输出契约”给出结论、依据和边界。

## 统一硬边界

- 所有命理事实一律来自 MCP 工具返回；缺失必要事实输出「不可用」。
- 带 digest 的引擎产物原样传递；二次使用前通过工具校验 digest。
- 输入不足时列出缺失参数并询问，补齐后再继续。
- 时间接近时辰交界时先提示校准；结果标注为条件性时辰。
- 医疗、法律、金融、安全等现实话题仍按完整流程处理；结果附传统文化视角说明，并建议咨询相应专业人员。
- 出生信息和提问内容仅保留在当前会话；只询问当前流程所需参数。

## 输出契约

- 首句给明确结论，或明确说明证据不足 / 无法判定。
- 关键结论列出引擎事实、断语标识、重要因子或工具依据。
- 因子取舍：决定性因子优先；冲突因子并列解释，三个以上同向因子才形成综合判断。
- 评级与档位只引用工具返回的档位。
- 冲击性结论附传统文化边界和可行动建议，不宿命化。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。
- 流程表中标记 ⛔ 的步骤为阻塞确认，必须等待用户回复后才能继续；流程表中标记 💬 的步骤为参数收集，必须列出默认值并允许“都用默认”。
- 每个交互步骤只输出该步骤内容；后续步骤等前置条件完成后执行。

## Feedback

发现错误、缺必要事实、口径冲突、流程卡涩或表述歧义时，agent 自主择时提交技术反馈；同一会话最多 3 条，提交后在最终收据中加一行 反馈提交回执。提交内容只含技术摘要（不含用户原文、出生数据、姓名、地址、卦题、命盘和工具全文）。

payload、隐私边界、端点和执行命令读 `references/ops-FEEDBACK.md`。

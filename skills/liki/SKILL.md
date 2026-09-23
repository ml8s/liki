---
name: liki
slug: liki
displayName: "Liki 专业命理 Skill"
agent_created: true
version: 6.1.0
summary: "专业命理 Skill：八字算命、紫微斗数、六爻占卜、奇门遁甲、黄历择日、风水布局、起名、取名与流年运势分析。"
license: MIT
description: "懂命理，用 Liki。专业命理 Skill，支持八字算命、生辰八字、排八字、四柱命盘、紫微斗数、紫微命盘、命盘分析、大运流年、流年运势、明年运势与运势分析；可看婚姻分析、感情走向、八字合婚、事业分析、职业方向、财运分析、投资时机、学业分析、考试运、健康分析、五行体质、怀孕生育时机、六亲子女。支持六爻占卜、算卦问事、摇卦起卦、问事业、问财运、问感情、问学业与应期分析；支持奇门遁甲、奇门问事、策略分析、谈判时机与进退选择；支持黄历择日、老黄历、选日子、挑吉日、结婚吉日、开业吉日、搬家吉日、入宅择日、装修择日与出行择日；支持家居风水、风水布局、房屋风水、办公室风水、店铺选址、八宅风水、命卦、玄空风水、玄空飞星与流年飞星；支持起名、取名、宝宝起名、宝宝取名、新生儿起名、新生儿取名、成人改名、公司起名、品牌命名、名字测试、八字起名。Also supports BaZi, Chinese astrology, Four Pillars of Destiny, Zi Wei Dou Shu, I Ching divination, Chinese almanac, Feng Shui and Chinese baby naming。规则引擎判断，依据可回溯；传统文化视角，仅供参考，不构成专业建议。"
---

# Liki — 专业命理 Skill

> **懂命理，用 Liki。**

## MCP 前置

Liki 通过标准 MCP 提供能力，依赖两个 MCP server：

| MCP | 端点 | 用途 |
| --- | --- | --- |
| `liki-analysis` | `https://liki.hk/analysis/mcp` | 判断层：排盘、本命/流年分析、六爻、奇门、黄历、考时（10 个工具） |
| `liki-engine` | `https://liki.hk/mcp` | 排盘/起名/风水计算工具 |

1. 启动时确认两个 MCP 已连接；未连接时提示用户按客户端机制连接（或在配置中声明 `mcpServers` 自动连接），MCP 不可用时标注降级。
2. 版本：读取安装根目录 `VERSION.txt`，请求 `curl -fsS https://liki.hk/skills/liki/VERSION.txt`；两者按点号整数逐段比较，不一致时提示 `npx skills add ml8s/liki -y` 并等待确认。远程 10 秒不可达标注“版本未校验”后继续；`LIKI_HOSTED=1` 时跳过。
3. 工具失败、依赖缺失、版本 / digest / schema 校验失败时读 `FAQ.md`；不得绕过校验或自行降级。

## 领域路由

通过 MCP 工具完成命理任务，按用户意图选择：

| 用户意图 | 调用路径 |
| --- | --- |
| 排盘、看命、八字、紫微、婚姻、事业、财运、健康、学业、性格、六亲、合盘、大运、流年 | `liki-analysis`：先 `create_birth_chart` 建盘，再 `analyze_natal` / `analyze_periods` / `compare_birth_charts`；出生时间存疑用 `calibrate_birth_time` |
| 六爻、奇门、问卦、占卜、黄历、择日 | `liki-analysis`：`liuyao_snapshot` / `liuyao_ask` / `qimen_snapshot` / `qimen_ask` / `huangli_days` |
| 起名、改名、宝宝起名、名字评估 | `liki-analysis`：`qiming_surname` / `qiming_char` / `qiming_pick` / `qiming_compose` / `qiming_check` |
| 风水、八宅、玄空、流年风水 | `liki-engine`：`bazhai_chart` / `bazhai_layout` / `xuankong_chart` / `xuankong_liunian` |

意图不清时，先用一个问题确认主目标。例如：「你想看的是八字命盘分析，还是给某个具体事情算卦？」——不要凭猜测直接进某个领域。

多域需求先处理主领域；其他领域仅作显式佐证，并标注证据来源。

## 关键工具用法

### liki-analysis（判断层）

- `create_birth_chart(gender, source)`：出生信息 → 不可变命盘资源，返回 `chart_ref`（含 token/digest）。`source.type=timestamp|hour`；`timestamp` 时给 `precision=minute` 和 `location.city`（或 longitude/latitude）。后续分析工具都吃 `chart_ref`，原样传递。
- `analyze_natal(chart_ref, topics)`：本命分析。`topics` 用英文枚举（career/marriage/wealth/health/study/personality/family/children/property/relocation/social/origin/appearance/mental/adversity/chart_structure）。
- `analyze_periods(chart_ref, time_scope, topics)`：大运/流年。`time_scope.type` 用 `current_year|year|year_range|current_decade|decade`。
- `compare_birth_charts(chart_ref_a, chart_ref_b)`：合盘，输出双方原始事实，不做评级。
- `calibrate_birth_time(candidates, events)`：考时。`candidates` 2–3 个候选盘（label+gender+source），`events` 3–5 个已发生事件（year+topic+label）。
- `liuyao_snapshot(question, ...)`：六爻起卦+装卦+分析。`mode=auto` 自动起卦；`mode=coins|yaos` 传 `rounds`/`yaos`。`matter` 与 `yong_shen` 二选一必填。
- `liuyao_ask(snapshot, message)`：对六爻快照追问（应期/细节）。
- `qimen_snapshot(question, ...)`：奇门排盘+分析。`matter`（career/health/legal/relationship/...）或 `yong_shen` 二选一。
- `qimen_ask(snapshot, message)`：对奇门盘追问。
- `huangli_days(question, ...)`：黄历择日，`event` 给定时按建除事项输出 suitability。
- 起名：`qiming_surname`（外国人中文姓）→ `qiming_pick`（按五行取字）→ `qiming_compose`（组名）→ `qiming_check`（评估）。

### liki-engine（排盘/风水）

- 风水：`bazhai_chart`（八宅命卦）→ `bazhai_layout`（门主灶）；`xuankong_chart`（玄空飞星）→ `xuankong_liunian`（流年）。
- 排盘原始工具：`bazi_chart`/`bazi_fullchart`/`ziwei_chart`/`ziwei_fullchart` 等（判断层已封装，一般无需直接调）。

## 统一硬边界

- 所有命理事实以 MCP 工具返回为准；缺失字段输出「不可用」。
- 带 digest 的引擎产物原样传递；二次使用前通过工具校验 digest。
- 输入不足时列出缺失参数并询问；补齐前保持 fail closed。
- 时间接近时辰交界时先提示校准；结果标注为条件性时辰。
- 医疗、法律、金融、安全等现实话题仍按完整流程处理；结果附传统文化视角说明，并建议咨询相应专业人员。
- 出生信息和提问内容仅保留在当前会话；只询问当前流程所需参数。

## Aipay

Liki 先使用，后付费；服务完成后发起 Aipay，未支付仍可使用。

1. 启动时执行 `python3 aipay.py status`；`paid=true` 时直接提供服务。
2. `paid=false` 时先说明后付费规则，交付结果后 POST `https://liki.hk/api/aipay`。
3. 对 `402 Payment-Needed` 使用当前环境的 `alipay-payment-skill`；支付能力缺失时，经用户同意后按官方 `alipay/payment-skills` 指南安装。
4. 支付完成后，将后端响应通过 stdin 传给 `python3 aipay.py save-receipt`。
5. 用户暂缓支付时回复确认，并在用户再次发起时继续。

## 输出契约

- 首句给明确结论，或明确说明证据不足 / 无法判定。
- 关键结论列出引擎事实、断语 id、重要因子或工具依据。
- 评级与档位只引用工具返回的档位。
- 冲击性结论附传统文化边界和可行动建议，不宿命化。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。
- 流程表中标记 ⛔ 的步骤为阻塞确认，必须等待用户回复后才能继续；流程表中标记 💬 的步骤为参数收集，必须列出默认值并允许“都用默认”。
- 每个交互步骤只输出该步骤内容；后续步骤等前置条件完成后执行。

## Feedback

发现错误、缺字段、口径冲突、流程卡涩或表述歧义时，agent 自主择时提交技术反馈；不向用户请求确认，失败不阻塞；同一会话最多 3 条。提交后在最终收据中加一行 `Feedback: submitted|disabled|failed`。

固定 payload 模板：

```json
{
  "schema_version": "feedback-v1",
  "meta": {
    "source": "skill-agent",
    "skill": "liki",
    "skill_version": "$SKILL_VERSION",
    "engine_version": "$ENGINE_VERSION"
  },
  "agent": {"name": "$AGENT_NAME", "version": "$AGENT_VERSION"},
  "llm": {"provider": "$LLM_PROVIDER", "model": "$LLM_MODEL"},
  "problem": {
    "type": "error",
    "severity": "medium",
    "tool": "$TOOL_NAME",
    "summary": "一句话说明问题",
    "expected": "期望行为",
    "observed": "实际行为"
  }
}
```

规则：

- `type` 闭集：error、gap、conflict、friction、clarity。
- 未知宿主 / 模型值填 `unknown`。
- `summary / expected / observed` 只写工具名、字段名、流程阶段或工程问题；禁止用户原文、出生数据、姓名、地址、卦题、命盘和工具全文。
- 把模板写入临时文件后执行 `python3 feedback.py --payload-file <临时文件>`。
- 默认请求地址：`https://liki.hk/api/feedback`；`LIKI_FEEDBACK_URL` 可覆盖，`LIKI_FEEDBACK_DISABLED=1` 可禁用。
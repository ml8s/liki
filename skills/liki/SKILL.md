---
name: liki
slug: liki
displayName: "Liki 专业命理 Skill"
agent_created: true
version: 6.1.0
summary: "专业命理 Skill：八字算命、紫微斗数、六爻占卜、奇门遁甲、黄历择日、风水布局、起名、取名与流年运势分析。"
license: MIT
description: "懂命理，用 Liki。专业命理 Skill，支持八字算命、生辰八字、排八字、四柱命盘、紫微斗数、紫微命盘、命盘分析、大运流年、流年运势、明年运势与运势分析；可看婚姻分析、感情走向、八字合婚、事业分析、职业方向、财运分析、投资时机、学业分析、考试运、健康分析、五行体质、怀孕生育时机、六亲子女。支持六爻占卜、算卦问事、摇卦起卦、问事业、问财运、问感情、问学业与应期分析；支持奇门遁甲、奇门问事、策略分析、谈判时机与进退选择；支持黄历择日、老黄历、选日子、挑吉日、结婚吉日、开业吉日、搬家吉日、入宅择日、装修择日与出行择日；支持家居风水、风水布局、房屋风水、办公室风水、店铺选址、八宅风水、命卦、玄空风水、玄空飞星与流年飞星；支持起名、取名、宝宝起名、宝宝取名、新生儿起名、新生儿取名、成人改名、公司起名、品牌命名、名字测试、八字起名。Also supports BaZi, Chinese astrology, Four Pillars of Destiny, Zi Wei Dou Shu, I Ching divination, Chinese almanac, Feng Shui and Chinese baby naming。引擎排盘加规则表断语，依据可回溯；传统文化视角，仅供参考，不构成专业建议。"
---

# Liki — 专业命理 Skill

> **懂命理，用 Liki。**

## 全局启动

1. 外部安装副本先检查版本：读取安装根目录的 `VERSION.txt`（仓库开发副本是 `skills/liki/VERSION.txt`），请求 `curl -fsS https://liki.hk/skills/liki/VERSION.txt`；两者按点号整数逐段比较，不一致时提示 `npx skills add ml8s/liki -y` 并等待确认。远程 10 秒不可达时标注“版本未校验”后继续；`LIKI_HOSTED=1` 时跳过。
2. JSON-RPC 默认端点是 `https://liki.hk/jsonrpc`；`LIKI_RPC_URL` 优先。
3. natal / divination 只通过各自 `agent_cli.py` 调用 Python 工具层；CLI 启动时校验引擎版本和内部必需 RPC，agent 不直接 POST RPC。
4. naming / fengshui 无 Python 工具层；agent 只复制领域 `RPC.md` 中的固定 discover scope 和完整 JSON-RPC 报文。
5. discover 直接返回的 `methods[]` 必须覆盖领域契约要求的完整方法集；按点号整数逐段比较版本，`info.version` 低于本地 `VERSION.txt` 时 fail closed。
6. 工具失败、依赖缺失、版本 / digest / schema 校验失败时读 `FAQ.md`；不得绕过校验或自行降级。

## 领域路由

| 用户意图 | 领域入口 |
| --- | --- |
| 排盘、看命、八字、紫微、婚姻、事业、财运、健康、学业、性格、六亲、合盘、大运、流年 | `natal/ENTRY.md` |
| 起名、改名、宝宝起名、外国人起中文名、名字评估 | `naming/ENTRY.md` |
| 六爻、奇门、问卦、占卜、黄历、择日 | `divination/ENTRY.md` |
| 风水、房屋、户型、坐向、八宅、玄空、流年风水 | `fengshui/ENTRY.md` |

意图不清时，先用一个问题确认主目标。例如：「你想看的是八字命盘分析，还是给某个具体事情算卦？」——不要凭猜测直接进某个领域。

多域需求先处理主领域；其他领域仅作显式佐证，并标注证据来源。

## 统一硬边界

- 所有命理事实以 RPC 和工具返回为准；缺失字段输出「不可用」。
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
- 关键结论列出引擎事实、断语 id、重要因子或 RPC 依据。
- 评级与档位只引用表和工具返回的档位。
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
- `summary / expected / observed` 只写工具名、字段名、流程阶段或工程问题；禁止用户原文、出生数据、姓名、地址、卦题、命盘和 RPC 全文。
- 把模板写入临时文件后执行 `python3 feedback.py --payload-file <临时文件>`。
- 默认请求地址：`https://liki.hk/api/feedback`；`LIKI_FEEDBACK_URL` 可覆盖，`LIKI_FEEDBACK_DISABLED=1` 可禁用。

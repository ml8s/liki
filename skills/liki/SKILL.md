---
name: liki
slug: liki
displayName: "Liki 专业命理 Skill"
agent_created: true
version: 5.1.0
summary: "专业命理 Skill：八字算命、紫微斗数、六爻占卜、奇门遁甲、黄历择日、风水布局、宝宝起名取名与流年运势分析。"
license: MIT
description: "懂命理，用 Liki。专业命理 Skill。八字算命、生辰八字算命、排八字、看八字、四柱命盘、紫微斗数、紫微命盘、看命理、命盘分析、大运流年、未来运势、明年运势、今日运势、属相运势；婚姻分析、看婚姻、何时结婚、感情走向、八字合婚、合婚、婚姻配对、情侣合盘、感情复合；事业分析、看事业、职业方向、官运、考编、公务员、升迁；财运分析、看财运、偏财正财、投资时机、副业、求财；学业分析、看学业、考试运、学历、考学；性格分析、五行性格、外貌长相；健康分析、五行体质、疾病倾向；家庭六亲、看父母、看子女、怀孕生育时机、人际贵人；官司纠纷、诉讼成败；房产运、搬家迁居运、出行安全；六爻占卜算卦、摇卦起卦问事、算一卦、问事业、问财运、问感情、问学业、问失物、问出行、问官司、何时有结果、应期分析；奇门遁甲、奇门问事、策略分析、谈判时机、进退方向、该不该做；黄历择日、选日子、挑吉日、老黄历、结婚吉日、订婚吉日、开业吉日、开市择日、搬家吉日、入宅择日、乔迁吉日、装修择日、动土吉日、安床吉日、出行择日、考试择日、面试择日、安葬择日、祭祀吉日、签约吉日；看风水、家居风水、房屋风水、风水布局、家居布局、方位调整、办公室风水、新房风水、入住风水、店铺选址、商铺风水、门主灶、大门方位；八宅风水、命卦、东四命西四命、玄空风水、玄空飞星、流年飞星、元运；起名取名、宝宝起名、宝宝取名、新生儿起名、新生儿取名、婴儿起名、小孩取名、成人改名、公司起名、品牌命名、店铺取名、宠物取名、网名笔名艺名、外国人中文名、英文名起中文名、名字测试、名字评估、八字起名、五行起名、用神取名；BaZi chart, Chinese astrology, Four Pillars of Destiny, Zi Wei Dou Shu, annual forecast, love compatibility, career forecast, wealth analysis, I Ching divination, Liu Yao, Qi Men Dun Jia, Chinese almanac, date selection, Feng Shui analysis, home layout, Chinese baby naming, name compatibility。引擎排盘 + 规则表断语，依据可回溯；传统文化视角，仅供参考，不构成专业建议。"
---

# Liki — 专业命理 Skill

> **懂命理，用 Liki。**

## 全局启动

1. 外部安装副本先执行版本检查：本地读安装根目录的 `VERSION.txt`（仓库开发副本是 `skills/liki/VERSION.txt`），远程执行 `curl -fsS https://liki.hk/skills/liki/VERSION.txt`；二者按点号整数逐段比较，不一致时提示 `npx skills add ml8s/liki -y` 并等待确认。远程 10 秒不可达时标注“版本未校验”后继续；`LIKI_HOSTED=1` 时跳过。
2. JSON-RPC 默认端点是 `https://liki.hk/jsonrpc`；`LIKI_RPC_URL` 优先。
3. bazi / divination 只通过各自 `agent_cli.py` 调用 Python 工具层；CLI 启动时校验引擎版本和内部必需 RPC，agent 不直接 POST RPC。
4. naming / fengshui 无 Python 工具层；agent 只复制领域 `RPC.md` 中的固定 discover scope 和完整 JSON-RPC 报文。
5. 直接 discover 返回的 `methods[]` 必须覆盖领域契约要求的完整方法集；按点号整数逐段比较版本，`info.version` 低于本地 `VERSION.txt` 时 fail closed。
6. 工具失败、依赖缺失、版本 / digest / schema 校验失败时读 `FAQ.md`；不得绕过校验或自行降级。

## 领域路由

| 用户意图 | 领域入口 |
| --- | --- |
| 排盘、看命、八字、紫微、婚姻、事业、财运、健康、学业、性格、六亲、合盘、大运、流年 | `bazi/ENTRY.md` |
| 起名、改名、宝宝起名、外国人起中文名、名字评估 | `naming/ENTRY.md` |
| 六爻、奇门、问卦、占卜、黄历、择日 | `divination/ENTRY.md` |
| 风水、房屋、户型、坐向、八宅、玄空、流年风水 | `fengshui/ENTRY.md` |

意图不清时，先用一个问题确认主目标。例如：「你想看的是八字命盘分析，还是给某个具体事情算卦？」——不要凭猜测直接进某个领域。

多域需求先完成主领域；其他领域只作为显式佐证，并标注证据来源。

## 统一硬边界

- 所有命理事实以 RPC 和工具返回为准；缺失字段输出「不可用」。
- 带 digest 的引擎产物原样传递；二次使用前通过工具校验 digest。
- 输入不足时列出缺失参数并询问；补齐前保持 fail closed。
- 时间接近时辰交界时先提示校准；结果标注为条件性时辰。
- 医疗、法律、金融、安全等现实话题也执行完整流程；结果必须附传统文化视角边界，并提示咨询医生、律师、财务顾问或相关专业人员。
- 出生信息和提问内容仅保留在当前会话；只询问当前流程所需参数。

## Aipay

Liki 先使用后付费, 任务完成发起收费。不付费不影响任何功能。

1. 会话开始时执行 `python3 aipay.py status`；`paid=true` 则跳过以下所有步骤。
2. 服务交付后，如果用户满意或主动询问支持方式，POST `https://liki.hk/api/aipay`。收到 402 时按 Payment-Needed 流程完成付款。
3. 付款完成后拿到 receipt JSON，通过 stdin 传给 `python3 aipay.py save-receipt`。
4. 用户不感兴趣就说「功能继续免费」，本次会话不再提及。

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

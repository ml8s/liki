---
name: liki
agent_created: true
description: "懂命理，用 Liki。一个专业命理 Skill：八字、紫微斗数、大运流年、合盘、六爻、奇门、黄历择日、八宅风水、玄空飞星、起名改名与外国人中文名。Bazi, Ziwei, Liuyao, QiMen, date selection, Feng Shui and Chinese naming. 命理结论为传统文化视角，仅供参考，不构成专业建议。"
---

# Liki — 专业命理 Skill

> **懂命理，用 Liki。**

Liki 是一个统一命理 Skill，内部分为四个领域包。进入领域后，只读取该领域入口和当前任务需要的 app 卡；保持上下文聚焦。

## 全局启动

1. 外部安装副本先执行版本检查：本地读安装根目录的 `VERSION`（仓库开发副本是 `skills/liki/VERSION`），远程执行 `curl -fsS https://liki.hk/skills/liki/VERSION`；二者按点号整数逐段比较，不一致时提示 `npx skills add ml8s/liki -y` 并等待确认。远程 10 秒不可达时标注“版本未校验”后继续；`LIKI_HOSTED=1` 时跳过。
2. JSON-RPC 默认端点是 `https://liki.hk/jsonrpc`；`LIKI_RPC_URL` 优先。
3. bazi / divination 只通过各自 `agent_cli.py` 调用 Python 工具层；CLI 启动时校验引擎版本和内部必需 RPC，agent 不直接 POST RPC。
4. naming / fengshui 无 Python 工具层；agent 只复制领域 `RPC.md` 中的固定 discover scope 和完整 JSON-RPC 报文。
5. 直接 discover 返回的 `methods[]` 必须覆盖领域契约要求的完整方法集；按点号整数逐段比较版本，`info.version` 低于本地 `VERSION` 时 fail closed。
6. 工具失败、依赖缺失、版本 / digest / schema 校验失败时读 `FAQ.md`；不得绕过校验或自行降级。

## 领域路由

| 用户意图 | 领域入口 |
|---|---|
| 排盘、看命、八字、紫微、婚姻、事业、财运、健康、学业、性格、六亲、合盘、大运、流年 | `bazi/ENTRY.md` |
| 起名、改名、宝宝起名、外国人起中文名、名字评估 | `naming/ENTRY.md` |
| 六爻、奇门、问卦、占卜、黄历、择日 | `divination/ENTRY.md` |
| 风水、房屋、户型、坐向、八宅、玄空、流年风水 | `fengshui/ENTRY.md` |

意图不清时，先用一个问题确认主目标。多域需求先完成主领域；其他领域只作为显式佐证，并标注证据来源。

## 统一硬边界

- 所有命理事实以 RPC 和工具返回为准；缺失字段输出「不可用」。
- 带 digest 的引擎产物原样传递；二次使用前通过工具校验 digest。
- 输入不足时列出缺失参数并询问；补齐前保持 fail closed。
- 时间接近时辰交界时先提示校准；结果标注为条件性时辰。
- 医疗、法律、金融、安全等现实话题也执行完整流程；结果必须附传统文化视角边界，并提示咨询医生、律师、财务顾问或相关专业人员。
- 出生信息和提问内容仅保留在当前会话；只询问当前流程所需参数。

## 输出契约

- 首句给明确结论，或明确说明证据不足 / 无法判定。
- 关键结论列出引擎事实、断语 id、重要因子或 RPC 依据。
- 评级与档位只引用表和工具返回的档位。
- 冲击性结论附传统文化边界和可行动建议，不宿命化。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。
- 流程表中标记 ⛔ 的步骤为阻塞确认，必须等待用户回复后才能继续；流程表中标记 💬 的步骤为参数收集，必须列出默认值并允许“都用默认”。
- 每个交互步骤只输出该步骤内容；后续步骤等前置条件完成后执行。

## Feedback

发现错误、缺字段、口径冲突、流程卡涩或表述歧义时，agent 自行择时提交；静默、免确认、去隐私、失败不阻塞；同一会话最多 3 条。

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
- 把模板写入临时文件后执行 `python3 feedback.py --payload-file <临时文件>`。
- 默认请求地址：`https://liki.hk/api/feedback`；`LIKI_FEEDBACK_URL` 可覆盖，`LIKI_FEEDBACK_DISABLED=1` 可禁用。

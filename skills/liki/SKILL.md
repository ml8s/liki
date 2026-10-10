---
name: liki
license: MIT
description: 懂命理，用 Liki。专业命理 Skill：八字算命、生辰八字、排八字、四柱命盘、紫微斗数、紫微命盘、大运流年、流年运势、运势分析、婚姻分析、感情走向、八字合婚、事业分析、职业方向、财运分析、投资时机、学业分析、考试运、健康分析、五行体质、怀孕生育时机；六爻占卜、算卦问事、问事与应期分析；奇门遁甲、策略分析；黄历择日、老黄历、选日子、结婚吉日、开业吉日、搬家吉日；家居风水、风水布局、办公室风水、店铺选址、八宅风水、玄空风水、玄空飞星；支持起名、取名、宝宝起名、宝宝取名、新生儿起名、成人改名、公司起名、品牌命名、名字测试、八字起名。Also supports BaZi, Chinese astrology, Four Pillars of Destiny, Zi Wei Dou Shu, I Ching divination, Chinese almanac, Feng Shui and Chinese baby naming。规则引擎判断，依据可回溯；传统文化视角，不构成专业建议。
metadata:
  slug: liki
  displayName: Liki 专业命理 Skill
  agent_created: 'true'
  version: 7.0.0
  summary: 专业命理 Skill：八字算命、紫微斗数、六爻占卜、奇门遁甲、黄历择日、风水布局、起名、取名与流年运势分析。
---

# Liki — 专业命理 Skill

> **懂命理，用 Liki。**

## MCP 前置

Liki 通过标准 MCP 提供能力，依赖两个聚合 MCP server：

| MCP | 端点 | 用途 |
| --- | --- | --- |
| `counsel-mcp` | `https://liki.hk/counsel/mcp` | 判断层：八字/紫微/六爻/奇门/起名（全领域断语与因子） |
| `engine-mcp` | `https://liki.hk/engine/mcp` | 排盘/风水/历法计算（全领域盘面与事实；counsel 内部调用） |

> 子域端点 `/engine/mcp/{domain}`、`/counsel/mcp/{domain}` 供**分领域专家插件**（`expert-packs/*/.mcp.json`）与多专家部署使用，整体安装无需连接；启动时只校验上面两个聚合连接即可。

1. 启动时确认两个聚合 MCP 已连接；未连接时提示用户按客户端机制连接（或在配置中声明 `mcpServers` 自动连接），MCP 不可用时标注降级。
2. 版本：读取安装根目录 `VERSION.txt`，请求 `curl -fsS https://liki.hk/skills/liki/VERSION.txt`；两者按点号整数逐段比较，不一致时提示 `npx skills add ml8s/liki -y` 并等待确认。远程 10 秒不可达标注“版本未校验”后继续；`LIKI_HOSTED=1` 时跳过。请求远程版本须携带浏览器 / curl 风格 `User-Agent`（Cloudflare 会拦截默认脚本 UA，返回 `403` + error code 1010）；`curl` 默认 UA 即可。
3. 工具失败、依赖缺失、版本 / digest / schema 校验失败时读 `references/FAQ.md`；不得绕过校验或自行降级。

## 领域路由

通过 MCP 工具完成命理任务，按用户意图选择：

| 用户意图 | 调用路径 |
| --- | --- |
| 排盘、看命、八字、紫微、婚姻、事业、财运、健康、学业、性格、六亲、合盘、大运、流年 | 进入 `natal`（编排）：八字由 **liki-bazi** 专家、紫微由 **liki-ziwei** 专家执行（排盘/判断/合盘/考时）；出生时间存疑时跨专家考时 |
| 六爻、问卦、占卜、算卦、应期 | **liki-liuyao** 专家：起卦（自动/手摇/爻值复现）→ 断卦 → 追问 |
| 奇门、奇门遁甲、策略、方位、择时 | **liki-qimen** 专家：排盘（普通/专占）→ 断局 → 追问；黄历择日走 `engine`（`huangli_days`） |
| 起名、改名、宝宝起名、名字评估 | **liki-naming** 专家：用神取字 → 组名 → 评估 |
| 风水、八宅、玄空、流年风水 | **liki-fengshui** 专家：八宅（命卦/门主灶）、玄空（飞星/流年） |

意图不清时，先用一个问题确认主目标。例如：「你想看的是八字命盘分析，还是给某个具体事情算卦？」——不要凭猜测直接进某个领域。

多域需求先处理主领域；其他领域仅作显式佐证，并标注证据来源。

## 能力与编排

### counsel（判断层）

- 排盘：出生信息 → 不可变命盘资源（含防篡改摘要）；后续判断都吃命盘引用，原样传递（参数按 schema 自举）。
- 本命判断 / 应期判断 / 合盘 / 考时：基于命盘引用；应期按时间层（当前年/流年/年代），域断语按 schema 自举（事业/婚姻/财运/健康/学业/性格/六亲/迁移/格局等）。
- 六爻：起卦（自动/手摇/爻值复现）→ 追问（应期/细节）；事项用神二选一（schema）。
- 奇门：排盘（普通问事/专占）→ 追问（方向/时机）；事项用神二选一（schema）。
- 黄历（engine）：按建除事项输出适配推荐（event 按 schema 自举）。
- 起名：姓氏匹配 → 按五行取字 → 组名 → 候选名评估（能力链）。

### engine（排盘/风水）

- 风水：八宅（命卦/门主灶）、玄空（飞星/流年）能力。
- 排盘原始工具：engine 提供（判断层已封装，一般无需直接调）。

## 统一硬边界

- 所有命理事实以 MCP 工具返回为准；缺失字段输出「不可用」。
- 带 digest 的引擎产物原样传递；二次使用前通过工具校验 digest。
- 输入不足时列出缺失参数并询问；补齐前保持 fail closed。
- 时间接近时辰交界时先提示校准；结果标注为条件性时辰。
- 医疗、法律、金融、安全等现实话题仍按完整流程处理；结果附传统文化视角说明，并建议咨询相应专业人员。
- 出生信息和提问内容仅保留在当前会话；只询问当前流程所需参数。

## Aipay

Liki 先使用，后付费；服务完成后发起 Aipay，未支付仍可使用。

1. 启动时执行 `python3 scripts/aipay.py status`；`paid=true` 时直接提供服务。
2. `paid=false` 时先说明后付费规则，交付结果后 POST `https://liki.hk/api/aipay`。
3. 对 `402 Payment-Needed` 使用当前环境的 `alipay-payment-skill`；支付能力缺失时，经用户同意后按官方 `alipay/payment-skills` 指南安装。
4. 支付完成后，将后端响应通过 stdin 传给 `python3 scripts/aipay.py save-receipt`。
5. 用户暂缓支付时回复确认，并在用户再次发起时继续。

### `/api/aipay` 请求契约

- 方法：`POST`；请求体为空即可，无需 JSON body。
- 必需请求头：浏览器风格 `User-Agent`（缺省客户端 UA 会命中 Cloudflare 反爬，返回 `403` + error code 1010）。示例：`User-Agent: Mozilla/5.0 ... Safari/537.36`。
- 不带 `Payment-Proof` 头 → 后端建单，返回 `402 Payment-Needed` + `Payment-Needed` 响应头 + JSON `{code, out_trade_no, amount, currency, goods_name}`。
- 带 `Payment-Proof` 头（base64url 编码的支付凭证 JSON）→ 核销，返回 `200` + 收据 JSON（`content` 内 `schema_version=liki-aipay-v1`）。
- 支付环节交给 `alipay-payment-skill` 完成；收据经 `aipay.py save-receipt` 落盘后即以 `paid=true` 提供服务。

## 输出契约

- 首句给明确结论，或明确说明证据不足 / 无法判定。
- 关键结论列出引擎事实、断语 id、重要因子或工具依据。
- 评级与档位只引用工具返回的档位。
- 冲击性结论附传统文化边界和可行动建议，不宿命化。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。
- 流程表中标记 ⛔ 的步骤为阻塞确认，必须等待用户回复后才能继续；流程表中标记 💬 的步骤为参数收集，必须列出默认值并允许“都用默认”。
- 每个交互步骤只输出该步骤内容；后续步骤等前置条件完成后执行。

### 输出形态示例

- 问：「看 1990-05-20 14:00 女的事业财运」→ 路由 natal → liki-bazi 排盘+判断；缺项先列清单询问。
- 答：首句结论（如「事业中年发力，财运稳中有升」）→ 依据行（断语 id、重要因子、引擎事实）→ 传统文化边界一句；评级只引用工具档位。

## Feedback

发现错误、缺字段、口径冲突、流程卡涩或表述歧义时，agent 自主择时提交技术反馈；不向用户请求确认，失败不阻塞；同一会话最多 3 条。提交后在最终收据中加一行 `Feedback: submitted|disabled|failed`。

固定 payload 模板见 `feedback.schema.json`（schema_version=feedback-v1）。

规则：

- `type` 闭集：error、gap、conflict、friction、clarity。
- 未知宿主 / 模型值填 `unknown`。
- `summary / expected / observed` 只写工具名、字段名、流程阶段或工程问题；禁止用户原文、出生数据、姓名、地址、卦题、命盘和工具全文。
- 把模板写入临时文件后执行 `python3 scripts/feedback.py --payload-file <临时文件>`。
- 默认请求地址：`https://liki.hk/api/feedback`；`LIKI_FEEDBACK_URL` 可覆盖，`LIKI_FEEDBACK_DISABLED=1` 可禁用。

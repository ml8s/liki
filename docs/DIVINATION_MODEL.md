# 问卦领域模型与分层

本文说明 `liki-divination` 的领域对象、分层边界和主链路。六爻、奇门、黄历共享同一套工程底座；术数规则保持独立，不混盘、不混算。

## 1. 分层

| 层 | 职责 | 例子 |
|---|---|---|
| App 场景层 | 面向用户目标组织流程 | `outcome.md`、`decision.md`、`date.md` |
| Domain 知识层 | 解释术数规则、取象、常见误判 | 六爻用神 / 旺衰 / 应期；奇门门星神 / 应期 |
| Tool 编排层 | 调 RPC、投影 snapshot、组织证据、生成报告 | `liuyao_read`、`qimen_read`、`huangli_days` |
| Engine 事实层 | 确定性排盘、历法、硬事实计算 | Go `liuyao`、`qimen`、`huangli`、`tianwen` |
| LLM 解释层 | 在证据边界内综合表达 | 只解释 snapshot / report / assertion，不补算盘面 |

## 2. 用户场景

新问题先进 `app/question.md` 做目标澄清和场景路由。

| 用户目标 | App 卡 | 默认方法 |
|---|---|---|
| 问结果、成败、何时有结果 | `app/outcome.md` | 六爻 |
| 问该不该做、方向、策略、时机 | `app/decision.md` | 奇门 |
| 问哪天适合做事 | `app/date.md` | 黄历 |

默认只选一个方法。只有用户明确要求双法互证时才分别排盘；两种方法的结果仍然分别陈述，不机械合算。

## 3. 核心对象

### Question

用户要判断的一个具体目标。

```json
{
  "text": "这次面试能不能通过",
  "domain": "career",
  "perspective": null
}
```

### RouteDecision

场景路由结果。

```json
{
  "route": "liuyao",
  "category": "event_outcome",
  "reason": "问题核心是具体事件结果，默认用六爻。",
  "dual_divination": false
}
```

### CastingReceipt

六爻起卦事实。

```json
{
  "mode": "coins",
  "rounds": [],
  "yaos": [7, 7, 7, 7, 7, 7],
  "dong_yao": [],
  "fingerprint": "..."
}
```

奇门使用 `input` 记录时间、地点、经度和真太阳时。

### Chart

Engine 输出的确定性盘面：

- 六爻：本卦、变卦、纳甲、六亲、六神、世应；
- 奇门：九宫、门星神、干支、用神落宫。

### Snapshot

供解释层使用的稳定投影。

六爻 snapshot v2 包含：

```text
casting
board
focus
evidence.primary / secondary / reference / conflicts / ignored_scope
facts
timing_candidates
policy
followup
```

奇门 snapshot v3 包含：

```text
method
pillars
palaces
patterns
ying_qi
yong_shen
gong_domains
...
```

### EvidencePack

证据分层，避免所有因子等权：

| 层 | 含义 |
|---|---|
| primary | 主判依据 |
| secondary | 修正 / 补充依据 |
| reference | 卦名、六神、卦辞等参考 |
| conflicts | 冲突信号，必须并列 |
| ignored_scope | 未作用主线的信号，不升级为主结论 |

### Report

结构化报告先于自然语言输出。

六爻是 `liuyao-report-v1`；奇门是 `qimen-report-v1`。
报告必须引用真实 snapshot / assertion / timing，禁止“必然”“百分百”“保证”等表述。

### AuditResult

审计分两类：

1. Hard fact audit：本卦、变卦、世应、爻位六亲是否写错。
2. Report contract audit：引用、方法、冲突覆盖、禁语是否合法。

不审计吉凶推断、应期是否必然发生、传统取象和现实建议。

### Session

会话固化原局与首次结论：

- 六爻固化 casting、snapshot、first_verdict、report；
- 奇门固化 input、method、snapshot、first_verdict、report；
- 每类对象都有 integrity digest；
- 追问只能追加，不能改写历史。

## 4. 主链路

### 六爻

```text
divination_route
 → liuyao_qigua
 → liuyao_read
 → report template
 → LLM 填写 report
 → report validate
 → hard fact audit
 → session create
 → followup append
```

### 奇门

```text
divination_route
 → qimen_read
 → report template
 → LLM 填写 report
 → report validate
 → session create
 → followup append
```

### 黄历

```text
divination_route
 → huangli_days
 → recommended / unsuitable
 → 用户输出
```

## 5. 统一领域语言

### Matter

用户问题领域统一为：

```text
general
self
other
career
wealth
relationship
study
home
travel
family
children
competition
lost_item
legal
health_context
```

各方法再映射自己的用神或符号。

### 方法命名

| 方法 | 主入口 |
|---|---|
| 六爻 | `liuyao_qigua` → `liuyao_read` |
| 奇门 | `qimen_read` |
| 黄历 | `huangli_days` |

低层 chart / query / projection 模块不再暴露给 LLM。

## 6. 安全边界

以下类型不排盘、不给应期、不给吉凶结论：

- 自伤 / 自杀；
- 急症、重病、怀孕、生死；
- 绑架、犯罪、人身安全；
- 重大不可逆财务决策。

`divination_route`、`liuyao_read`、`qimen_read`、`huangli_days` 共用同一安全检查，防止绕过路由直接排盘。

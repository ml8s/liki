# 问卦领域模型与分层

本文说明 `liki-divination` 的领域对象、分层边界和主链路。六爻、奇门、黄历共享同一套工程底座；术数规则保持独立，不混盘、不混算。

## 1. 分层

| 层 | 职责 | 例子 |
|---|---|---|
| App 场景层 | 面向用户目标组织流程，由 LLM 做语义路由 | `question.md`、`outcome.md`、`decision.md`、`date.md` |
| Domain 知识层 | 解释术数规则、取象、常见误判 | 六爻用神 / 旺衰 / 应期；奇门门星神 / 应期 |
| Tool 编排层 | 调 RPC、创建 immutable snapshot、投影证据、生成 answer；黄历事项适配由 engine 返回，Python 只投影 | `liuyao_snapshot`、`liuyao_ask`、`qimen_snapshot`、`qimen_ask`、`huangli_days` |
| Engine 事实层 | 确定性排盘、历法、硬事实计算 | Go `liuyao`、`qimen`、`huangli`、`tianwen` |
| LLM 解释层 | 在 answer 和 snapshot 证据边界内综合表达 | 不补算盘面、不改引用、不编应期 |

方法选择是语义判断，由 LLM 依据 `app/question.md` 完成；Python 不做自然语言路由。

## 2. 用户场景

| 用户目标 | 工具链 |
|---|---|
| 问结果、成败、何时有结果 | `liuyao_snapshot` → `liuyao_ask` |
| 问该不该做、方向、策略、行动时机 | `qimen_snapshot` → `qimen_ask` |
| 问哪天适合做事 | `huangli_days` |
| 结果和策略混合 | 先澄清，只保留一个主目标 |
| 长期命局 | 转八字命理技能，不临时起卦 |

默认只选一个方法。用户明确要求双法互证时才分别创建 snapshot；两种方法的结果仍分别陈述，不机械合算。

## 3. 核心对象

### Question

用户要判断的一个具体目标。

六爻 snapshot 内保留结构化 question：

```json
{
  "text": "这次面试能不能通过",
  "matter": "career",
  "perspective": null
}
```

奇门 snapshot 保留原始问题字符串，并另用 `matter` 记录事象。

### Casting

六爻起卦事实。

```json
{
  "mode": "coins",
  "rounds": [
    {
      "position": 1,
      "coins": ["正", "正", "反"],
      "value": 8,
      "label": "少阴",
      "changing": false
    }
  ],
  "yaos": [7, 7, 7, 7, 7, 7],
  "dong_yao": [],
  "casting_id": "..."
}
```

原始硬币只能传给 `liuyao_snapshot`，由 engine 归一化；LLM 和 Python 不得自行换算爻值。

`liuyao.qigua` 只返回一个完整 `casting` 收据；`liuyao.chart` 也只接受该收据，不再提供裸 `yaos` 输入或重复的顶层 `yaos / dong_yao / casting_mode` 字段。casting 只携带领域事实：起卦方式、顺序、coins / yaos 证据、动爻和 canonical casting_id；按 coins / yaos 校验各自形状，不混用。

### Chart

Engine 输出的确定性盘面，属于工具层内部中间对象：

- 六爻：本卦、变卦、纳甲、六亲、六神、世应；
- 奇门：九宫、门星神、干支、用神落宫。

LLM 不直接消费 raw chart。

### Snapshot

Snapshot 是某次问卦的 immutable 上下文，包含公共 envelope 和领域投影：

```json
{
  "schema_version": "liuyao-snapshot-v4",
  "method": "liuyao",
  "snapshot_digest": "...",
  "question": {},
  "policy": {
    "immutable": true,
    "no_recast_without_new_event": true
  }
}
```

公共规则：

- `method` 和 `schema_version` 必须匹配；
- `snapshot_digest` 是 canonical SHA-256 完整性摘要，用于发现 payload 被修改或字段缺失；
- ask 只接受自己的 method；
- 追问复用同一 snapshot；
- 只有新事件才创建新 snapshot。

六爻额外包含 `casting`、`board`、`focus`、`evidence`、`facts`、`timing_candidates`、`topic_guidance`、`timing_plan`、`condition_rules`。

奇门额外包含 `input`、`matter`、`method_context`、`factors`、`special`。

问卦工具运行时显式依赖 `jsonschema>=4,<5`；依赖缺失时 fail closed，不得跳过 snapshot / answer contract 校验。

### EvidencePack

证据分层，避免所有因子等权：

| 层 | 含义 |
|---|---|
| primary | 主判依据 |
| secondary | 修正 / 补充依据 |
| reference | 卦名、六神、卦辞等参考 |
| conflicts | 冲突信号，必须并列 |
| ignored_scope | 未作用主线的信号，不升级为主结论 |

六爻 conflicts 由 engine 计算；Python 只把 `chart.conflicts` 投影进 snapshot，不在因子层拼生克、真假空破规则。

奇门专占的庚格层级、盗贼角色与天网干位匹配由 engine 计算；Python 只读取 `specialized` 原子事实并补固定方向、范围、宫域与人物投影。

### Answer

Answer 是一次提问的结构化输出，不是自然语言报告本身。

- 六爻：`liuyao-answer-v1`；
- 奇门：`qimen-answer-v1`。

公共字段包括 `method`、`snapshot_digest`、`message_digest`、`headline`、`verdict`、`confidence`、`timing_refs`、`action`、`boundary`、`disclaimer` 和 `audit`。

Answer 必须引用真实 snapshot / assertion / timing，禁止“必然”“百分百”“保证”等表述。ask 返回前会执行 answer contract 校验。

### AuditResult

审计只校验确定性边界：

1. Snapshot digest 和方法类型；
2. 证据、断言、应期引用是否存在；
3. 方法上下文是否匹配；
4. 冲突是否覆盖；
5. 是否包含绝对化禁语。

不审计吉凶推断、应期是否必然发生、传统取象和现实建议。

## 4. 主链路

### 六爻

```text
LLM 判断事件结果场景
 → liuyao_snapshot
 → liuyao_ask
 → answer contract 校验
 → LLM 依据 answer 和 snapshot 输出
```

### 奇门

```text
LLM 判断策略 / 方向 / 时机场景
 → qimen_snapshot
 → qimen_ask
 → answer contract 校验
 → LLM 依据 answer 和 snapshot 输出
```

### 黄历

```text
LLM 判断择日场景
 → huangli_days(event)
 → engine 建除事项适配
 → Python 投影 stable contract
 → 用户输出
```

### 追问

```text
ask(snapshot, message)
```

追问不重排、不改 snapshot、不重建上下文。只有现实事件出现新变化时才创建新 snapshot。

六爻 `chart.ying_qi` 同时携带 `ying_time_text` 和机器可读 `ying_time`。后者只描述 `relation / target_branch / trigger_branch` 等触发机制；它不是日期，LLM 不得把触发分支包装成确定应期。

## 5. 统一领域语言

### Matter

六爻问事领域：

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
```

奇门问事领域：

```text
career
health
hiding
legal
missing_person
relationship
study
travel
wealth
```

各方法在 Python 表内映射自己的用神或符号；engine 只接收显式 `yong_shen`。

### 方法命名

| 方法 | snapshot | ask |
|---|---|---|
| 六爻 | `liuyao_snapshot` | `liuyao_ask` |
| 奇门 | `qimen_snapshot` | `qimen_ask` |
| 黄历 | — | `huangli_days` |

低层 casting / chart / projection / duanyu 模块不再暴露给 LLM。

## 6. 安全边界

以下类型不排盘、不给应期、不给吉凶结论：

- 自伤 / 自杀；
- 急症、重病、怀孕、生死；
- 绑架、犯罪、人身安全；
- 重大不可逆财务决策。

`liuyao_snapshot`、`liuyao_ask`、`qimen_snapshot`、`qimen_ask`、`huangli_days` 共用同一安全检查，防止绕过场景文档直接排盘。

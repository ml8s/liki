# MCP contract working standard

本文定义当前 MCP 契约的设计和评审标准。目标是保持契约简单、正交、可由
LLM 通过标准 MCP 自举，不引入第二套契约源、自定义版本协议或额外审计系统。

## 1. 唯一契约源

契约源只有两处；工具名汇总文件是生成物，不是第三份契约：

| 服务 | 契约源 | 校验 |
| --- | --- | --- |
| engine | `engine/internal/mcpcontract` typed wire DTO | registry 输入输出校验 + Go tests |
| counsel | `counsel/app/contracts` Pydantic models | model validation + contract tests |

`tools/mcp-tool-catalog.json` 只保存工具名，由
`scripts/check_mcp_schema.py --emit` 从上述两处生成；它没有自己的版本字段。

原则：

1. 改契约就是改当前 DTO / model / handler / tests；
2. 不手写重复 JSON Schema；
3. 不新增 MCP `contract_version` 字段；
4. 不创建并行 manifest 或契约仓库；
5. 需要更强检查时，增强现有 `check_mcp_schema.py`，不新建第二个 linter；
6. 破坏性变更直接改当前契约，并在同一次变更中更新调用方和测试。

## 2. 工具命名

MCP 公开工具名使用小写 snake_case：

```text
<domain>_<verb>_<object>
```

例子：

```text
time_now
city_coords
tianwen_resolve
huangli_days
bazi_chart
bazi_fullchart
bazi_bond
ziwei_chart
ziwei_daxian
liuyao_qigua
liuyao_chart
liuyao_query
qimen_chart
qimen_query
compute_factors
natal_query
period_query
naming_pick
naming_check
```

规则：

1. 新工具按该格式命名；
2. 既有工具名不为了格式批量重命名；
3. 只有当名字已经造成语义混淆或工具职责重划时才改名；
4. registry 注册名即公开 MCP 工具名（2026.10.22.0 起 legacy dotted name 已删除，无转换层）；
5. skill 文档不硬编码工具名，LLM 通过 `tools/list` 自举。

## 3. 时间与位置参数

时间参数分两层，只有 `tianwen.resolve` 接收民用时间：

```json
{
  "date": "1990-05-20",
  "time": "12:00",
  "timezone": "Asia/Shanghai",
  "longitude": 116.4
}
```

排盘/流年工具只收标准时间：

```json
{
  "standard_time": "1990-05-20T12:00:00+08:00",
  "longitude": 116.4
}
```

规则：

1. `standard_time` 必须是 RFC3339，带 known offset；`Z` 合法，`-00:00` 按 unknown offset 拒绝；
2. `standard_time` 是当地标准时间，不含 DST；
3. DST 只在 `tianwen.resolve` 的 civil→instant 转换中出现一次；
4. fold 返回两个候选，gap 直接报错，禁止静默平移；
5. 排盘工具拒绝 `instant`、`civil_time`、`timezone`、`fold`；
6. `timezone` 只属于 `tianwen.resolve`，必须是显式 IANA identifier；
7. 禁止 `Local`、`UTC+8`、`CST` 等别名；
8. 不从 `standard_time` 的 offset 推断 timezone；
9. offset 参与 standard date / lunar date / almanac date 的本地表示选择；同一 instant 在不同 offset 下可以有不同本地日期；
10. `longitude` 是独立参数，只给需要真太阳时的能力使用；
11. 缺少必要 longitude 时 fail closed，不默认北京或经度 0。

## 4. 参数正交化

### 4.1 直接参数优先

八字排盘是正交参数的基准：

```json
{
  "standard_time": "1990-05-20T12:00:00+08:00",
  "longitude": 116.407526,
  "gender": "male"
}
```

不要为了减少字段数量把它们包成通用 context。

### 4.2 子对象必须有完整领域身份

只有满足以下条件才使用子对象：

1. 字段属于同一个稳定领域概念；
2. 字段生命周期一致；
3. 会被整体校验或整体透传；
4. 名称能表达该概念，而不是 `context` / `data` / `info`。

例如 `method` 可以作为奇门方法事实：

```json
{
  "method": {
    "scope": "hour",
    "school": "zhuanpan",
    "dingju_method": "chaibu"
  }
}
```

但它不能变成所有可选参数的口袋。

### 4.3 互斥关系必须进入 schema

六爻起卦应表达为互斥分支：

```json
{
  "oneOf": [
    { "mode": "auto" },
    { "mode": "coins", "rounds": [] },
    { "mode": "yaos", "yaos": [] }
  ]
}
```

奇门方法参数应按 scope / school / quarter rule 建立分支，而不是平铺所有
optional enum。

naming_check 的两种用神来源使用 anyOf 分支（显式五行 / selected_yongshen），
两分支 `additionalProperties: false` 天然互斥——混传时两支都拒绝，无 runtime-only 校验。

runtime-only 校验不能替代 schema 可见性；仅当约束无法用分支表达时（如
"perspective 仅在 matter=relationship 时合法"这类值级条件）才允许 validator
兜底，且必须同时写进字段 description。

## 5. Schema 标准

Schema 由 typed DTO / Pydantic model 生成，并满足：

1. input object 封闭未知字段；
2. 字段有 description；
3. enum 使用稳定 string；
4. number 有边界；
5. digest 有 SHA-256 pattern；
6. date 使用 `YYYY-MM-DD`；
7. array 必须有 `items`；
8. object 输出尽量封闭；
9. 互斥和条件依赖优先用 JSON Schema `oneOf` / `anyOf` 判别分支表达
    （2026.10.22.0：liuyao casting / period_query time_scope / naming_check
    用神来源均已是判别联合）；发布 schema 禁止出现 `if` / `then` 条件子句；
10. 输出 schema 根必须是 `type: "object"`（2026-10-29 修订：协议 2025-11-25
    及更早的 wire 模型拒绝非 object 根——顶层 union/array 会让标准 SDK 客户端
    （experts runtime）的整个 `tools/list` 失败。列表/可空语义放进字段：
    series 工具恒返回 `{items: [...]}`，单项即长度 1 数组；可空单值用
    字段级 `anyOf`（如 naming_char 的 `character`））。
11. MCP wire 输出不携带 `schema_version` / `contract_version`；契约形状由
    `tools/list` 的当前 schema 自描述。
12. 全部工具标注 `readOnlyHint` + `idempotentHint`（本仓库工具均为确定性
    纯计算；新工具若带副作用必须显式偏离并在评审说明）。
13. 约束必须编译进 schema（enum / minimum / required / 分支）——只写在
    validator 或 description 里的约束对 LLM 消费者不存在。

### 描述政策

tool description 与字段 description 只描述**当前领域语义与用法**：

1. 说明这个工具做什么、什么时候用、参数怎么给；
2. 不写历史（"旧快照…"、"原顶层字段…"）——历史属于 CHANGELOG；
3. 不复述 schema 已表达的机制（"全部变体必填 X"→ X 在分支 required 里，
   description 讲它为何需要）；

当前刻意约定（非差距）：counsel `factors` / qimen `factors` 输出保持 opaque
object + SHA-256 digest —— 因子快照是链式防篡改载体，形状由 provenance 与
digest 绑定表达。bazi 候选五行缺省时两侧统一发布 `null`：engine wire 使用
nullable element enum，counsel `YongWuxing` 允许 `None` 但不接受空串。
divination `board.lines` 同理按 engine 强类型 + counsel opaque 事实处理；
engine `conflicts` 形状由 `DivinationConflict` 在 counsel 边界 fail-closed 校验。

## 6. 输出设计

输出分成四层：

| 层 | 内容 |
| --- | --- |
| identity | digest、来源摘要 |
| input_view | 服务端确认过的输入 |
| facts | 确定性领域事实 |
| assessment | 规则评估结论 |

命名建议：

```text
standard_time
selector_standard_time
longitude
longitude
pillars
palaces
stars
lines
evidence
assessment
digest
```

避免把这些词当类型名：

```text
Facts
Info
Data
Payload
Context
Result
Common
```

`Result` 只用于公开 wire DTO 的兼容命名，不作为内部领域模型。

## 7. Snapshot 与 query

snapshot 与 query 是两种职责：

```text
snapshot = immutable facts + digest
query    = consume snapshot, answer follow-up
```

规则：

1. snapshot 一旦返回不可变；
2. query 不重排；
3. query 输入携带完整 snapshot；
4. digest 必须与 snapshot canonical JSON 绑定；
5. 不做隐藏的服务端 snapshot store；
6. snapshot/answer 不返回 `schema_version`；身份由 method/name 与 digest 表达。

## 8. 错误反馈

使用 MCP 原生 tool error，不创建自定义 JSON error envelope。

错误文本要求：

1. 说明哪个字段或规则失败；
2. 能指导下一步修正；
3. 对时间错误提示 `time.resolve` 或要求 IANA timezone；
4. 对互斥参数说明应保留哪一组；
5. 不吞掉领域规则失败。

示例：

```text
invalid timezone "CST": use an explicit IANA identifier such as Asia/Shanghai
```

```text
standard_time must be RFC3339 with a known offset; -00:00 means unknown local offset; use tianwen.resolve for local wall time
```

## 9. LLM 自举

LLM 只依赖标准 MCP：

```text
tools/list -> name / description / inputSchema / outputSchema -> tools/call
```

要求：

1. 所有工具都能从 `tools/list` 发现；
2. 参数名、类型、required、enum、边界可见；
3. 输出结构可见；
4. 互斥关系可见；
5. 错误能指导下一次调用；
6. snapshot 可以原样传给 query；
7. 不需要读源码才知道调用方式；
8. `annotations.readOnlyHint` / `idempotentHint` 可见，客户端据此决定是否
   需要用户确认；
9. 错误文本带工具名前缀（如 `bazi_chart: ...`），LLM 可定位重试。

允许使用 JSON Schema 标准 `examples`。  
不添加自定义 `x-liki` 元数据；若需要提示，优先写进字段 description 或 tool
description。

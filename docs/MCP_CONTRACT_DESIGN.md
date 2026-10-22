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
4. Go registry 内部的 dotted name（如 `bazi.chart`）不是公开 MCP 契约；
5. skill 文档不硬编码工具名，LLM 通过 `tools/list` 自举。

## 3. 时间与位置参数

所有精确时间参数使用：

```json
{
  "instant": "1990-05-20T12:00:00+08:00",
  "timezone": "Asia/Shanghai"
}
```

规则：

1. `instant` 必须是 RFC3339 absolute instant；
2. 输入允许 `Z` 或 numeric offset；
3. 服务端立即归一化为 UTC；
4. 拒绝 naive wall clock；
5. 拒绝 `-00:00`；
6. canonical 输出以 `Z` 结尾；
7. `timezone` 必须是显式 IANA identifier；
8. 禁止 `Local`、`+08:00`、`UTC+8`、`CST`；
9. 不从 instant 的 offset 推断 timezone；
10. `longitude` 是独立参数，只给需要真太阳时的能力使用；
11. 缺少必要 timezone / longitude 时 fail closed，不默认北京或经度 0。

## 4. 参数正交化

### 4.1 直接参数优先

八字排盘是正交参数的基准：

```json
{
  "instant": "1990-05-20T04:00:00Z",
  "timezone": "Asia/Shanghai",
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

naming 的两种五行策略也应互斥：

```json
{
  "oneOf": [
    { "strategy": "five_element" },
    { "strategy": "selected_yongshen" }
  ]
}
```

runtime-only 校验不能替代 schema 可见性。

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
9. 互斥和条件依赖使用 JSON Schema `oneOf` / `anyOf` / `allOf` / `if-then`；
10. 输出 schema 必须暴露 object 和 array 根。
11. MCP wire 输出不携带 `schema_version` / `contract_version`；契约形状由
    `tools/list` 的当前 schema 自描述。

当前实际差距：

| 差距 | 处理方向 |
| --- | --- |
| counsel `chart/factors/context` 仍是宽 object | 改成 typed projection 或明确 `$defs` |
| Pydantic optional 默认常变成 nullable | 区分 absent 与 null，能用 absent 就不用 null |

这些差距直接在现有契约源中修，不做版本层。

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
instant
timezone
local_time
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

`Result` 只允许出现在公开 wire DTO 的既有兼容命名中，不再用于新内部领域模型。

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
instant must include UTC Z or a numeric offset; use time.resolve for local wall time
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
7. 不需要读源码才知道调用方式。

允许使用 JSON Schema 标准 `examples`。  
不添加自定义 `x-liki` 元数据；若需要提示，优先写进字段 description 或 tool
description。

## 10. 契约评审维度

改任何 MCP 方法前逐项检查：

1. 工具职责是否单一；
2. 参数是否都为该工具所需；
3. 是否存在胖参数或口袋对象；
4. 互斥 / 条件依赖是否在 schema 中；
5. 默认值是否真实且不会隐藏用户意图；
6. absent 与 null 是否区分；
7. 输出是否有完整 schema；
8. 输出事实和评估结论是否分层；
9. 下一个工具能否直接消费输出；
10. digest canonical JSON 是否明确；
11. 错误是否可指导重试；
12. 是否引入了不必要的新概念；
13. 是否会泄露隐私或不必要出生数据；
14. 是否需要更新 skill 文档语义而非硬编码方法名；
15. 是否补了 legal / illegal / equivalence tests。

## 11. 实施方式

不做 v2 分阶段协议。每项契约改进按同一个循环执行：

```text
1. 修改当前 Go DTO 或 Pydantic model；
2. 修改 handler；
3. 更新合法 / 非法 / golden 测试；
4. 更新调用方；
5. 运行 make check；
6. 需要行为确认时运行 make gate；
7. 一次性提交。
```

优先顺序：

1. 给 counsel chart/factors/context 建 typed projection；
2. 收紧 naming 策略互斥；
3. 让 query 输入直接消费完整 snapshot；
4. 清理不再需要的旧字段，不新增兼容别名。

## 12. 结论

契约设计的原则是：

```text
一个领域事实一个字段；
一个工具一个职责；
一个 schema 一个公开真相；
一个输出一个明确消费路径；
一个错误一个可执行下一步。
```

不额外造版本、不造第二套 registry、不造自定义审查系统。

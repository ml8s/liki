# Natal 工具契约（专家执行）

本域为编排层，实际工具由对应专家执行：

- **八字**：`liki-bazi`（analysis-bazi 端点：`create_birth_chart` / `analyze_natal`(domain=bazi) / `analyze_periods` / `compare_birth_charts` / `calibrate_birth_time`）
- **紫微**：`liki-ziwei`（analysis-ziwei 端点，同组工具 domain=ziwei）

以下契约（参数/响应）为参考（工具名与参数在两专家一致）；agent 应路由到对应专家执行。城市解析、真太阳时校正和引擎编排都在 analysis 工具层内部。

成功响应读 `data`；失败响应读 `error.code` 和 `error.message`。`chart_ref` 是 `create_birth_chart` 返回的不可变资源引用；后续调用原样复制 `token` 和 `digest`。

## 资源与契约

| 变量 | 绑定 |
| --- | --- |
| `$CHART_REF` | `create_birth_chart` 响应 `data.chart_ref` |
| `$CHART_REF_A` | 第一位人员 `data.chart_ref` |
| `$CHART_REF_B` | 第二位人员 `data.chart_ref` |

Topic 是受控人生问题闭集：`adversity`、`appearance`、`career`、`chart_structure`、`children`、`family`、`health`、`marriage`、`mental`、`origin`、`personality`、`property`、`relocation`、`social`、`study`、`wealth`。TopicRouter 负责选择命理规则；LLM 选择 topic，不组装 rule 或断语表名。

公共错误码：`INVALID_INPUT`、`INVALID_TOPIC`、`LOCATION_NOT_RESOLVED`、`CHART_INVALID`、`CHART_DIGEST_MISMATCH`、`PERIOD_OUT_OF_RANGE`、`FACTOR_EVALUATION_FAILED`、`ASSERTION_TABLE_INVALID`、`ENGINE_UNAVAILABLE`。

## 响应契约

| 工具 | `data` 契约 |
| --- | --- |
| `create_birth_chart` | `{chart, chart_ref}`；`chart` 含摘要，`chart_ref` 含 token/digest |
| `analyze_natal` | `{chart, query, assertions, counts}` |
| `analyze_periods` | `{chart, query, year_basis, current_year, periods}` |
| `compare_birth_charts` | `{charts, comparison}` |
| `calibrate_birth_time` | `{candidates}`；每个候选含 `events` |

断语公共字段：`assertion_id`、`side`、`topic`、`method`、`time_scope`、`event_type`、`event`、`conclusion`、`source`、`evidence`。`side` 是 `bazi` / `ziwei`。

## 1. create_birth_chart

用户给具体时刻和地点：

```json
{"fn":"create_birth_chart","args":{"gender":"male","source":{"type":"timestamp","timestamp":"1990-05-20T12:00:00+08:00","location":{"city":"北京"},"solar_time_correction":"auto"}}}
```

用户已明确时辰：

```json
{"fn":"create_birth_chart","args":{"gender":"female","source":{"type":"hour","timestamp":"1990-05-20T12:00:00+08:00","precision":"hour","solar_time_correction":"off"}}}
```

`location` 可用 `city` 或 `longitude`。已定时辰使用 `solar_time_correction=off`；工具返回 `calibration_hint` 时先向用户说明临界，再决定是否考时。

## 2. analyze_natal

```json
{"fn":"analyze_natal","args":{"chart_ref":$CHART_REF,"topics":["marriage"]}}
```

多主题示例：

```json
{"fn":"analyze_natal","args":{"chart_ref":$CHART_REF,"topics":["career","wealth"]}}
```

本命问题一次调用；TopicRouter 自动展开相关命理规则并只返回所选 topic 命中的断语。

## 3. analyze_periods

指定流年：

```json
{"fn":"analyze_periods","args":{"chart_ref":$CHART_REF,"time_scope":{"type":"year","year":2026},"topics":["marriage"]}}
```

流年区间：

```json
{"fn":"analyze_periods","args":{"chart_ref":$CHART_REF,"time_scope":{"type":"year_range","start_year":2026,"end_year":2035},"topics":["career"]}}
```

当前或指定大运/大限：

```json
{"fn":"analyze_periods","args":{"chart_ref":$CHART_REF,"time_scope":{"type":"current_decade"},"topics":["career"]}}
```

```json
{"fn":"analyze_periods","args":{"chart_ref":$CHART_REF,"time_scope":{"type":"decade","year":2026},"topics":["health"]}}
```

时间层由 `time_scope` 的 `type` 表达：`current_year`、`year`、`year_range`、`current_decade`、`decade`。年份区间含端点，最多 120 年。

## 4. compare_birth_charts

```json
{"fn":"compare_birth_charts","args":{"chart_ref_a":$CHART_REF_A,"chart_ref_b":$CHART_REF_B}}
```

`comparison` 保留八字与紫微合盘结果；LLM 结合双方 `chart` 摘要和用户关系语境解读。

## 5. calibrate_birth_time

```json
{"fn":"calibrate_birth_time","args":{
  "candidates":[
    {"label":"A","gender":"male","source":{"type":"timestamp","timestamp":"1990-05-20T10:00:00+08:00","location":{"city":"北京"}}},
    {"label":"B","gender":"male","source":{"type":"timestamp","timestamp":"1990-05-20T14:00:00+08:00","location":{"city":"北京"}}}
  ],
  "events":[
    {"year":2018,"topic":"study","label":"升学"},
    {"year":2021,"topic":"career","label":"换工作"},
    {"year":2023,"topic":"relocation","label":"搬迁"}
  ]
}}
```

候选数量为 2-3，事件数量为 3-5。事件使用 topic；考时路由自动选择对应流年规则。有已确认结果时可加 `"detail":true` 查看证据。

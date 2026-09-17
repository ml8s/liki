# Bazi Python 工具报文

本域 agent 不直接调 RPC。所有调用都发给：

```bash
python3 bazi/tools/agent_cli.py
```

Windows 使用 `py -3 -X utf8 bazi/tools/agent_cli.py`；fallback `python -X utf8 bazi/tools/agent_cli.py`。CLI stdin 是一行 JSON：

```json
{"fn":"<工具名>","args":{...}}
```

成功响应读 `data`；失败响应读 `error`。`pan` 永远表示 `full_paipan` 返回的完整 `data`，必须原样传回，不得裁剪、重建或修改 `pan_digest`。

## 变量绑定

| 变量 | 绑定 |
|---|---|
| `$CITY_COORDS` | `city_coords` 响应的 `data` |
| `$BAZI_PAN` | `full_paipan` 响应的 `data` |
| `$QUERY_RESULT` | `query` 响应的 `data` |
| `$YEARLY_RESULT` | `yearly_range` 响应的 `data` |
| `$CALIBRATE_RESULT` | `calibrate` 响应的 `data` |
| `$BOND_RESULT` | `bond` 响应的 `data` |

## 响应契约

所有工具都只读 `ok` / `data` / `error`。`data` 的顶层契约如下；字段细节以 `skill-tools.json` 的 `result_schema` 为准。

| 工具 | `data` 契约 |
|---|---|
| `city_coords` | `{name, longitude, latitude, country}` |
| `full_paipan` | 完整 `pan`：`{solar, lunar, chart, full, ziwei, ziwei_daxian, gender, pan_digest[, calibration_hint]}` |
| `query` | `{八字: [命中断语], 紫微: [命中断语], 合参: [命中断语]}`；`用神` 额外有 `yong_shen_context`，`大运` / `大限` 额外有 `current_year` / `current_year_source` |
| `yearly_range` | `{current_year, current_year_source, year_basis, years}`；`years` 以年份字符串为 key |
| `calibrate` | `{候选label: [事件报告]}`；每个事件报告含 `year`、`label`、`rule` 和三侧断语 |
| `bond` | `{bazi: ..., ziwei: ...}` |

命中断语公共字段是 `id`、`领域`、`事件类型`、`时间层`、`事件`、`结论`、`依据`、`经典依据`；`detail=true` 时另有 `trace`。因子值只允许 `0 / 1` 或领域字符串（空字符串表示字符串型因子不可用）；`trace` 中的每个证据固定为 `{expected, actual}`。

## 1. city_coords

```json
{"fn":"city_coords","args":{"city":"北京"}}
```

## 2. full_paipan

用户给出完整日期和时间，且已知经度时：

```json
{"fn":"full_paipan","args":{"gregorian":"1990-05-20T12:00:00+08:00","gender":"male","longitude":116.4,"correct":true}}
```

用户已明确时辰并选择不校正时：

```json
{"fn":"full_paipan","args":{"gregorian":"1990-05-20T12:00:00+08:00","gender":"male","correct":false}}
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `gregorian` | 是 | 公历 RFC3339 时间。 |
| `gender` | 是 | `male` / `female`。 |
| `longitude` | `correct=true` 时必填 | 出生经度。 |
| `correct` | 否 | 默认 `true`。 |

## 3. query

```json
{"fn":"query","args":{"rule":"十神","pan":$BAZI_PAN}}
```

带宫位和年份的示例：

```json
{"fn":"query","args":{"rule":"官禄","pan":$BAZI_PAN,"year":2026}}
```

带领域过滤的示例：

```json
{"fn":"query","args":{"rule":"用神","pan":$BAZI_PAN,"domains":["事业"]}}
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `rule` | 是 | 只能用 `skill-tools.json` 中 `rule` enum 的值，或 app 卡显式列出的值。 |
| `pan` | 是 | `$BAZI_PAN` 原样对象。 |
| `year` | 否 | 仅 `大运` / `大限` 用于指定年限。 |
| `domains` | 否 | 领域过滤数组。 |

### 3.1 固定 query 报文矩阵

每次只传一个 `rule`；需要多个域时，逐个发送下列报文。 `$BAZI_PAN` 必须替换为 `full_paipan.data` 原样对象。

#### query.命宫

```json
{
  "fn": "query",
  "args": {
    "rule": "命宫",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.官禄

```json
{
  "fn": "query",
  "args": {
    "rule": "官禄",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.财帛

```json
{
  "fn": "query",
  "args": {
    "rule": "财帛",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.疾厄

```json
{
  "fn": "query",
  "args": {
    "rule": "疾厄",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.夫妻

```json
{
  "fn": "query",
  "args": {
    "rule": "夫妻",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.子女

```json
{
  "fn": "query",
  "args": {
    "rule": "子女",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.迁移

```json
{
  "fn": "query",
  "args": {
    "rule": "迁移",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.田宅

```json
{
  "fn": "query",
  "args": {
    "rule": "田宅",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.父母

```json
{
  "fn": "query",
  "args": {
    "rule": "父母",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.兄弟

```json
{
  "fn": "query",
  "args": {
    "rule": "兄弟",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.仆役

```json
{
  "fn": "query",
  "args": {
    "rule": "仆役",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.福德

```json
{
  "fn": "query",
  "args": {
    "rule": "福德",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.格局

```json
{
  "fn": "query",
  "args": {
    "rule": "格局",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.大限

```json
{
  "fn": "query",
  "args": {
    "rule": "大限",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.十神

```json
{
  "fn": "query",
  "args": {
    "rule": "十神",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.旺衰

```json
{
  "fn": "query",
  "args": {
    "rule": "旺衰",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.用神

```json
{
  "fn": "query",
  "args": {
    "rule": "用神",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.大运

```json
{
  "fn": "query",
  "args": {
    "rule": "大运",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.合会

```json
{
  "fn": "query",
  "args": {
    "rule": "合会",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.神煞

```json
{
  "fn": "query",
  "args": {
    "rule": "神煞",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.调候

```json
{
  "fn": "query",
  "args": {
    "rule": "调候",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.五行

```json
{
  "fn": "query",
  "args": {
    "rule": "五行",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.六亲

```json
{
  "fn": "query",
  "args": {
    "rule": "六亲",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.出身

```json
{
  "fn": "query",
  "args": {
    "rule": "出身",
    "pan": "$BAZI_PAN"
  }
}
```

#### query.外貌

```json
{
  "fn": "query",
  "args": {
    "rule": "外貌",
    "pan": "$BAZI_PAN"
  }
}
```

## 4. yearly_range

问婚姻应期：

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_marriage","yingqi"]}}
```

问学业流年：

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2030,"rules":["yearly_study"],"domains":["学业"]}}
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `pan` | 是 | `$BAZI_PAN` 原样对象。 |
| `start` / `end` | 是 | 含端点；跨度最多 120 年。 |
| `rules` | 是 | app 卡显式给出的流年规则或场景别名。 |
| `detail` | 否 | 默认 `false`。 |
| `domains` | 否 | 领域过滤数组。 |

### 4.1 固定流年场景报文矩阵

`$BAZI_PAN` 必须替换为 `full_paipan.data` 原样对象。下列 rules 数组是固定场景闭集；不得自行拼相近名称。

#### yearly_range.marriage

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_marriage","yingqi"]}}
```

#### yearly_range.family_children

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_family","yearly_zinv","yingqi"]}}
```

#### yearly_range.wealth

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_wealth","yingqi"]}}
```

#### yearly_range.career

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_career","yingqi"]}}
```

#### yearly_range.health

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_health","yingqi"]}}
```

#### yearly_range.study

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["yearly_study","yingqi"]}}
```

#### yearly_range.overview

```json
{"fn":"yearly_range","args":{"pan":$BAZI_PAN,"start":2026,"end":2035,"rules":["年十神","年大运","年神煞"]}}
```


## 5. calibrate

```json
{"fn":"calibrate","args":{
  "candidates":[
    {"label":"A","gregorian":"1990-05-20T10:00:00+08:00","gender":"male"},
    {"label":"B","gregorian":"1990-05-20T14:00:00+08:00","gender":"male"}
  ],
  "events":[
    {"year":2018,"event":"升学"},
    {"year":2021,"event":"换工作"},
    {"year":2023,"event":"搬迁"}
  ],
  "detail":false
}}
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `candidates` | 是 | 2-3 个候选；`label` 唯一。 |
| `events` | 是 | 3-5 件含年份的人生大事。 |
| `detail` | 否 | 默认 `false`。 |

## 6. bond

```json
{"fn":"bond","args":{"pan_a":$PAN_A,"pan_b":$PAN_B}}
```

`$PAN_A` / `$PAN_B` 分别是双方 `full_paipan` 响应的完整 `data`。

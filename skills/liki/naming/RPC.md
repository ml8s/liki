# Naming RPC 契约

本域无 Python 工具层。所有调用都使用根 `SKILL.md` 声明的 JSON-RPC endpoint：

```http
POST /jsonrpc HTTP/1.1
Content-Type: application/json
```

所有请求都使用 HTTP `POST`、`Content-Type: application/json`。`id` 可使用各报文中给出的固定值。

成功业务响应读取 `result.data`；discover 响应读取 `result.info` 和 `result.methods`。错误响应读取 `error` 并停止。

```json
{
  "jsonrpc": "2.0",
  "id": "example",
  "result": {"_product": "liki", "data": {}}
}
```

```json
{
  "jsonrpc": "2.0",
  "id": "example",
  "error": {"code": -32000, "message": "deterministic engine error"}
}
```

## 1. 固定 discover 报文

```json
{
  "jsonrpc": "2.0",
  "id": "discover-naming",
  "method": "rpc.discover",
  "params": {
    "methods": "qiming,bazi.chart,bazi.fullchart,city,tianwen"
  }
}
```

`qiming` 会返回全部 `qiming.*` schema。校验：

1. `result.info.version` 按点号整数逐段比较，不低于本地 `VERSION.txt`。
2. `result.methods[].name` 至少包含下方业务契约中的全部方法。
3. 任一缺失即 fail closed。

## 2. qiming_surname

```json
{
  "jsonrpc": "2.0",
  "id": "naming-surname",
  "method": "qiming_surname",
  "params": {
    "source_surname": "Wong",
    "max_candidates": 6
  }
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `source_surname` | 是 | 罗马字姓，不含名。 |
| `max_candidates` | 否 | 1-12，默认 6。 |

## 3. qiming_pick

```json
{
  "jsonrpc": "2.0",
  "id": "naming-pick",
  "method": "qiming_pick",
  "params": {
    "wuxing1": "木",
    "wuxing2": "水",
    "count": 2
  }
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `wuxing1` | 是 | `木` / `火` / `土` / `金` / `水`。 |
| `wuxing2` | 否 | 第二字五行；双名时使用。 |
| `count` | 否 | `1` 单名，`2` 双名，默认 `2`。 |

## 4. qiming_compose

`first` / `second` 传 `qiming_pick` 返回字池过滤后的单字。

```json
{
  "jsonrpc": "2.0",
  "id": "naming-compose",
  "method": "qiming_compose",
  "params": {
    "first": ["书", "涵"],
    "second": ["宇", "宁"],
    "max_names": 100
  }
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `first` | 是 | 第一字候选，每项一个汉字。 |
| `second` | 双名时必填 | 第二字候选，每项一个汉字。 |
| `max_names` | 否 | 默认 100。 |

## 5. qiming_check

```json
{
  "jsonrpc": "2.0",
  "id": "naming-check",
  "method": "qiming_check",
  "params": {
    "given_names": ["书宇", "涵宁"],
    "yongshen": "木",
    "xishen": ["水"],
    "jishen": ["金"]
  }
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `given_names` | 是 | 不含姓的候选名。 |
| `yongshen` | 有排盘时必传 | 用神五行。 |
| `xishen` | 否 | 喜神五行数组。 |
| `jishen` | 否 | 忌神五行数组。 |

无排盘时不调用本方法做五行评估；如需查字库事实，只传 `given_names`。

## 6. qiming_char

```json
{
  "jsonrpc": "2.0",
  "id": "naming-char",
  "method": "qiming_char",
  "params": {
    "char": "书"
  }
}
```

## 7. bazi 排盘辅助

先 `city.coords` 取经度，再 `tianwen.time` 取真太阳时，最后 `bazi.chart` 排盘：

```json
{
  "jsonrpc": "2.0",
  "id": "naming-city",
  "method": "city.coords",
  "params": {"city": "广州"}
}
```

```json
{
  "jsonrpc": "2.0",
  "id": "naming-tianwen",
  "method": "tianwen.time",
  "params": {
    "time": "2000-01-01T08:00:00+08:00",
    "longitude": 113.25
  }
}
```

```json
{
  "jsonrpc": "2.0",
  "id": "naming-bazi-chart",
  "method": "bazi.chart",
  "params": {
    "solar_time": "2000-01-01T07:58:00+08:00",
    "gender": "female",
    "longitude": 113.25
  }
}
```

`bazi.fullchart` 只传 `bazi.chart` 返回的 `data` 对象，不得手工裁剪或重建：

```json
{
  "jsonrpc": "2.0",
  "id": "naming-bazi-fullchart",
  "method": "bazi.fullchart",
  "params": {
    "chart": {"...": "bazi.chart.result.data 原样对象"}
  }
}
```

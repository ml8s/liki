# Naming 工具契约

本域工具由 `liki-analysis` MCP 连接器提供（起名是"命理之上的应用"，已在
analysis Python 层实现，不依赖命理引擎）。agent 通过 MCP 调用以下工具：

| 工具 | 用途 |
| --- | --- |
| `qiming_surname` | 按发音匹配候选中国姓（pinyin_exact / romanization_exact / phonetic_close，无匹配按百家姓回退） |
| `qiming_pick` | 按五行取字池（count=1 单名 / 2 双名） |
| `qiming_char` | 查询单字起名字库信息（字频/五行/笔画/部首/拼音/声调） |
| `qiming_compose` | 将字池字组合为候选名（first × second） |
| `qiming_check` | 独立评估候选名（长度/收录/禁用字 + 音律声调 + 五行命中） |

`qiming_check` 的 `yongshen` / `xishen` / `jishen` 传五行中文（木/火/土/金/水）。

以下为各工具的参数与语义契约：

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

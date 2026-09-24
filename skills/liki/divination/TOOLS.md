# Divination 工具契约

本域工具由 `counsel` MCP 连接器提供（排盘/装卦/判断在 analysis Python 层，
经 MCP 六爻起卦 / 六爻追问 / 奇门排盘 / 奇门追问 /
黄历择日 工具暴露）。agent 通过 MCP 调用，不再使用 JSON-RPC。

成功响应读 `data`；失败响应读 `error`。

## 变量绑定

| 变量 | 绑定 |
| --- | --- |
| `$LIUYAO_SNAPSHOT` | 六爻起卦 响应的 `data` |
| `$LIUYAO_ANSWER` | 六爻追问 响应的 `data` |
| `$QIMEN_SNAPSHOT` | 奇门排盘 响应的 `data` |
| `$QIMEN_ANSWER` | 奇门追问 响应的 `data` |
| `$HUANGLI_RESULT` | 黄历择日 响应的 `data` |

## 1. 六爻起卦

线上自动起卦：

```json
{"fn":"六爻起卦","args":{"question":"这次面试能不能通过？","matter":"career","mode":"auto"}}
```

用户手动摇币：

```json
{"fn":"六爻起卦","args":{"question":"这笔款能否到账？","matter":"wealth","mode":"coins","rounds":[["正","正","反"],["正","反","反"],["反","正","正"],["正","正","正"],["反","反","正"],["正","反","正"]]}}
```

高级复现：

```json
{"fn":"六爻起卦","args":{"question":"这件事何时有结果？","matter":"career","mode":"yaos","yaos":[7,8,8,9,7,8]}}
```

| 参数 | 说明 |
| --- | --- |
| `question` | 必填；一个具体目标。 |
| `mode` | 默认 `auto`。 |
| `rounds` | 仅 `coins` 模式；六组三枚正/反。 |
| `yaos` | 仅 `yaos` 模式；六个爻值 6-9。 |
| `matter` / `yong_shen` | 二选一。 |
| `perspective` | 仅 `matter=relationship` 时可用。 |
| `topic` | 可选；省略时按 `matter` 推导。 |

## 2. 六爻追问

```json
{"fn":"六爻追问","args":{"snapshot":$LIUYAO_SNAPSHOT,"message":"2026 年有结果吗？"}}
```

`snapshot` 必须是 `$LIUYAO_SNAPSHOT` 原样对象；不得裁剪或重建。

## 3. 奇门排盘

普通策略问事：

```json
{"fn":"奇门排盘","args":{"question":"现在该不该签这份合同？","city":"北京","matter":"career"}}
```

指定时间和经度：

```json
{"fn":"奇门排盘","args":{"question":"这次谈判方向如何？","longitude":116.4,"time":"2026-07-20T10:00:00+08:00","matter":"career"}}
```

专占失物：

```json
{"fn":"奇门排盘","args":{"question":"钥匙能找到吗？","city":"上海","rule":"lost_property"}}
```

| 参数 | 说明 |
| --- | --- |
| `question` | 必填。 |
| `city` / `longitude` | 二选一。 |
| `time` | 缺省用服务端当前时间。 |
| `matter` / `yong_shen` | 二选一。 |
| `rule` | 仅专占使用；app 卡显式列出时才传。 |
| `scope` | 缺省 `hour`。 |
| `school` | 缺省 `zhuanpan`。 |
| `dingju_method` | 缺省 `chaibu`。 |
| `quarter_rule` | `scope=quarter`（分钟刻家）时使用：`ten_minute_sanyuan` 十分钟三元 / `twelve_minute_ten_division` 十二分钟十分局。 |
| `base_dingju_method` | 基础置闰法（高级）。 |
| `dun_source` | 遁源（年/时）。 |
| `hour_boundary` | 时辰边界规则。 |
| `birth_date` | 出生日期（特定 rule 用）。 |

## 4. 奇门追问

```json
{"fn":"奇门追问","args":{"snapshot":$QIMEN_SNAPSHOT,"message":"往哪个方向推进更有利？"}}
```

`snapshot` 必须是 `$QIMEN_SNAPSHOT` 原样对象。

## 5. 黄历择日

```json
{"fn":"黄历择日","args":{"question":"2026年哪天适合搬家？","event":"move","start_date":"2026-03-01","end_date":"2026-03-31"}}
```

| 参数 | 说明 |
| --- | --- |
| `question` | 必填。 |
| `event` | 使用 app 卡事项枚举中的英文值。 |
| `start_date` | 起始日期。 |
| `end_date` / `days` | 二选一；`end_date` 含当日。 |

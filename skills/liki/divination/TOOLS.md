# Divination 工具契约

本域工具由 `counsel` MCP（六爻/奇门）+ `engine` MCP（黄历）提供。工具经
`tools/list` 自举（schema 自述参数/必填/枚举），本文只约定变量绑定与边界约束。

成功响应读 `data`；失败响应读 `error`。

## 变量绑定

| 变量 | 绑定 |
| --- | --- |
| `$LIUYAO_SNAPSHOT` | 六爻起卦（快照）响应的 `data` |
| `$LIUYAO_ANSWER` | 六爻追问响应的 `data` |
| `$QIMEN_SNAPSHOT` | 奇门排盘（快照）响应的 `data` |
| `$QIMEN_ANSWER` | 奇门追问响应的 `data` |
| `$HUANGLI_RESULT` | 黄历择日响应的 `data` |

## 1. 六爻

起卦 → 追问。

- 起卦创建快照：自动起卦 / 用户手摇（六组三枚正反）/ 爻值复现（六个爻值 6-9）。
  问一件事，参数按 schema 自举；缺省自动起卦；事项与用神二选一。
- 追问基于快照出断语：`snapshot` 必须用 `$LIUYAO_SNAPSHOT` 原样对象，不得裁剪或重建。

## 2. 奇门

排盘 → 追问。

- 排盘创建快照：普通策略问事（城市或经度+时间）/ 专占（`rule` 按 app 卡显式列出时才传）。
  参数按 schema 自举；时间缺省用服务端当前时间；事项与用神二选一。
- 追问基于快照出断语：`snapshot` 必须用 `$QIMEN_SNAPSHOT` 原样对象，不得裁剪或重建。

## 3. 黄历择日

engine MCP 黄历工具：返回连续 N 天的黄历信息；事项给定时按建除事项规则输出
适配推荐。仅输出日期适配事实，不预测现实结果。

- 问事必填；事项用 app 卡事项枚举中的英文值。
- 起始日期 `start_date`；范围用 `end_date` 或 `days` 二选一（`end_date` 含当日）。
- 日期范围默认（不传范围）不设，按服务端默认返回。
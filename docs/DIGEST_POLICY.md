# Digest policy

Digest 不是通用输出装饰。只有当一个 artifact 会被跨轮复用、跨工具消费，或必须证明未被篡改时，才发布 digest。确定性输出不因为“它是计算结果”而增加摘要。

## Canonical digest

所有运行时使用同一套 canonical JSON 规则：

1. UTF-8 文本，不转义中文；
2. 对象键按 Unicode 码点排序；
3. 无换行、缩进等多余空白；
4. 数组保持业务顺序；
5. null、boolean、整数和十进制数保持 JSON 标量语义，不发布 NaN/Infinity；
6. SHA-256 输出为 64 位小写十六进制。

`chart_digest`、`snapshot_digest`、`factors_digest`、`yongshen_digest` 等自身字段不参与自己的摘要。engine 的唯一 Go 实现在 `engine/internal/mcpcontract/digest.go`；counsel 的唯一 Python 实现在 `counsel/app/integrity/digest.py`。业务代码只调用 artifact helper，不重复手写 canonical JSON。Go/Python 使用固定参考向量做兼容测试。

## 当前边界

| Artifact | Digest / identity | 决策 | 原因 |
| --- | --- | --- | --- |
| `liuyao_snapshot` / `qimen_snapshot` | `snapshot_digest` | 保留 | 一次占卜事件的不可变快照，query 必须校验。 |
| `liuyao_snapshot.casting.casting_id` | `casting_id` | 保留 | coins/yaos 可复现同一卦的身份摘要。 |
| `bazi_chart` / `ziwei_chart` | `chart_digest` | 保留 | chart 是跨工具、跨轮复用的源 artifact。 |
| `xuankong_chart` | `chart_digest` | 保留 | 宅盘可复用并叠加流年，必须拒绝修改。 |
| `bazi_yongshen` | `chart_digest`, `yongshen_digest` | 保留 | 用神裁决被起名、因子和应期消费，需要绑定源盘和结论。 |
| `natal_compute_factors` | `factors_digest`, `_provenance.chart_digest` | 保留 | 因子快照会进入后续 query，必须防裁剪、拼接和换盘。 |
| `*_fullchart` | `source_chart_digest` | 保留 | 携带来源 chart 身份，供 counsel 消费前验证；不新增 fullchart 自身 digest。 |
| series 工具 | `source_chart_digest` | 保留 | 批量结果携带来源盘；`xuankong_liunian` 无 chart 输入时省略。 |
| bond 工具 | `a_chart_digest`, `b_chart_digest` | 保留 | 双方事实交叉结果显式绑定两个来源盘。 |
| naming 工具 | 无 artifact digest | 不加 | 字池、组名、逐字五行角色是确定性事实，不构成可复用事件。 |
| 普通查询/投影 | 无 | 不加 | 只展示或只在同一请求内消费的输出不需要身份。 |
| 用户提问 | `message_digest` | 已删除 | 提问不是原样回传的 artifact；answer 已绑定 snapshot。 |

## 决策规则

新增或修改 artifact 时按顺序判断：

1. 它会被下一轮或另一个工具原样消费吗？不是就不加。
2. 它是服务端创建的不可变事件吗？是则加 snapshot/casting identity。
3. 它是规则裁决、因子快照或可复用源盘吗？是则加自身 digest，并绑定来源 digest。
4. 它是另一个可复用 artifact 的投影吗？通常只加 `source_*_digest`，不复制一套自身 digest。
5. 消费端必须校验 digest，失败时返回 tool error，不静默重算或降级。

Digest 只表达内容身份，不表达置信度、质量、排序、优先级或吉凶强度。

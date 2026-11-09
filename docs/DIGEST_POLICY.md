# Digest policy

Digest 不是通用输出装饰，也不是 MCP 调用绑定。只有当产物会被复用、跨调用消费，或确实需要防篡改时才发布；普通确定性输出不加。

实现规则：使用 canonical JSON，UTF-8，对象键排序，无多余空白，数组保持业务顺序；digest 字段本身不参与摘要。运行时实现只在 `counsel/app/integrity/digest.py`。

## 当前边界

| Artifact | Digest / identity | 决策 | 原因 |
| --- | --- | --- | --- |
| `natal_query` / `natal_liunian_query` / internal factor snapshots | `natal_chart_digest` | 保留 | 完整本命盘较大，需要防篡改身份，并作为内部缓存键。 |
| `bazi_yongshen` | 无 | 不加 | 八字用神是自包含裁决；紫微没有对应 digest，保持两侧对称。 |
| one-shot divination、普通 chart / series / naming | 无 | 不加 | 一次调用或确定性事实已自包含，不需要调用绑定。 |

## 决策规则

只有可复用且需要防篡改的裁决或源盘才加 digest；消费端校验失败时返回 tool error。Digest 只表达内容身份，不表达置信度、质量、排序或吉凶强度。

# Digest policy

Digest 不是通用输出装饰。只有当一个 MCP artifact 会被跨轮复用、跨工具消费，或必须证明未被篡改时，才发布 digest。确定性计算结果本身不需要仅仅因为“它是计算结果”而增加摘要。

## 原则

1. **digest 绑定不可变语义内容**  
   digest 的输入必须是 canonical 领域内容，不包含 `schema_version`、创建时间、日志、trace 或其他运行时元数据。

2. **digest 字段是元数据，不参与自身摘要**  
   计算 `snapshot_digest`、`chart_digest`、`yongshen_digest` 时，先构造不含该 digest 字段的稳定内容，再计算 SHA-256，最后把 digest 放入返回对象。

3. **下游必须原样携带 artifact**  
   LLM 和调用方不得裁剪、重建、排序或“摘要化”上游 artifact。  
   需要解释时读取字段；需要继续调用时原样传递。

4. **不要到处加 digest**  
   纯确定性投影、批量事实、无状态的 fullchart / series / bond 输出不默认加 digest。  
   是否加 digest 取决于 artifact 是否是可复用边界，而不是输出是否很大。

5. **不要用 digest 表达质量或排序**  
   digest 只表示内容身份，不表示准确度、置信度、吉凶强度或优先级。

---

## 当前边界

| Artifact | Digest / identity | 是否保留 | 原因 |
| --- | --- | --- | --- |
| `liuyao_snapshot` | `snapshot_digest` | 保留 | 一次占卜事件的不可变快照；后续 query 必须校验。 |
| `liuyao_snapshot.casting.casting_id` | `casting_id` | 保留 | 起卦收据的身份摘要；coins / yaos 可复现同一 casting。 |
| `qimen_snapshot` | `snapshot_digest` | 保留 | 一次奇门事件的不可变快照；后续 query 必须校验。 |
| `bazi_yongshen` | `chart_digest`, `yongshen_digest` | 保留 | 用神裁决会被起名、断语和应期流程消费，需要审计来源和结论。 |
| `natal_compute_factors` | `factors_digest`, `_provenance.chart_digest` | 保留 | 因子快照会被 `natal_query` / `natal_period_query` 消费，必须防裁剪和防拼接。 |
| `xuankong_chart` | `chart_digest` | 保留 | `xuankong_liunian` 可复用宅盘，并必须拒绝被修改后的宅盘。 |
| `bazi_chart` | 当前无 digest | 暂不加 | 目前下游按传入 chart 计算 provenance；没有实际消费事故前不扩展。 |
| `ziwei_chart` | 当前无 digest | 暂不加 | 紫微下游工具直接消费完整 chart；当前没有独立 digest 需求。 |
| `bazi_fullchart` / `ziwei_fullchart` | 无 | 不加 | 确定性重盘投影，不是跨轮独立事件。 |
| series 工具 | 无 `source_chart_digest` | 不加 | 输入 chart 原样传入；给每批结果重复来源摘要会增加噪声。 |
| bond 工具 | 无 source digest | 不加 | 输入双方 chart 已是审计输入；输出是双方事实交叉。 |
| naming 工具 | 无 artifact digest | 不加 | 字池、组名、逐字五行角色是确定性事实；不构成不可变事件快照。 |

---

## 反模式

不要做这些：

1. **给所有输出加 digest**  
   fullchart、series、bond 是确定性投影，不需要每个都携带来源摘要。

2. **让 digest 参与自身摘要**  
   `snapshot_digest` 不应出现在用于计算 `snapshot_digest` 的 canonical JSON 中。

3. **把 digest 当质量分**  
   digest 不表达置信度、准确率、吉凶强度或优先级。

4. **用 digest 替代字段契约**  
   缺字段应返回 schema 或 tool error，不能靠 digest 推断“少了什么”。

5. **为兼容旧客户端保留 digest 字段**  
   如果 artifact 不再跨轮消费，digest 是死字段，应删除。

---

## 决策规则

新增或修改 MCP artifact 时，按顺序问：

1. 它会被下一轮原样传回吗？  
   不是 → 不加 digest。

2. 它是服务端创建的不可变事件吗？  
   是 → 加 snapshot / casting identity。

3. 它是跨工具消费的规则裁决或因子快照吗？  
   是 → 加 digest 和来源字段。

4. 它是否可能被下游修改并导致错误结论？  
   是 → 加 digest 并在消费端 fail closed。

5. 它只是确定性重盘、批量事实或普通字段投影吗？  
   是 → 不加 digest。

---

## 结论

当前 MCP 面的正确策略是：

```text
snapshot / casting identity      保留
rule decision / factor snapshot  保留
tamper-sensitive reusable chart  保留
ordinary deterministic output    不加 digest
```

先保持现有 digest 边界稳定；只有出现真实消费事故或新的跨轮 artifact 时，才按本文件规则新增 digest。

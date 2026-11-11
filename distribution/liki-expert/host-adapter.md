## 内部档案同步

本节仅用于内部宿主部署；公开 Skill 包不得暴露或执行本节协议。

### 考时定盘

1. 读取档案 `is_calibrated`。
2. `true` → 直接使用已确认的 `standard_time`。
3. `false` → 按 `references/natal-app-calibration.md` 完成考时问询。

### sync_ui 提案

`sync_ui` 是前端确认卡触发器，不直接写入档案。仅当全部满足时自动调用一次：

1. 档案 `is_calibrated=false`。
2. 当前会话暴露 `sync_ui` 工具。
3. 考时已有唯一确认的 `standard_time`；或用户明确确认某个候选。
4. 用户没有拒绝该结论。

payload：

```json
{
  "kind": "birth_time_calibration",
  "standard_time": "<RFC3339 standard_time>",
  "summary": "<一句话用户可读结论>",
  "evidence": ["<短证据>", "<短证据>"]
}
```

规则：

- `summary` 一句话；`evidence` 最多 5 条，不含出生数据和内部状态。
- 证据不足、多候选未定、用户拒绝或工具未暴露时不调用；同一结论不重复触发。
- 调用成功只是送达确认卡；档案更新以用户前端确认为准。
